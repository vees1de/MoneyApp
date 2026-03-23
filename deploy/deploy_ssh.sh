#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  ./deploy/deploy_ssh.sh user@server [/opt/moneyapp]

The script expects deploy/.env.prod to exist locally.
It syncs the repository to the server, runs docker compose,
applies SQL migrations, and installs the nginx site config.
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

SSH_TARGET="${1:-}"
REMOTE_DIR="${2:-/opt/moneyapp}"

if [[ -z "${SSH_TARGET}" ]]; then
  usage
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${SCRIPT_DIR}/.env.prod"
NGINX_TEMPLATE="${SCRIPT_DIR}/nginx/moneyapp.conf.template"

for cmd in ssh scp tar mktemp sed grep cut tail; do
  if ! command -v "${cmd}" >/dev/null 2>&1; then
    echo "missing required command: ${cmd}" >&2
    exit 1
  fi
done

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "missing ${ENV_FILE}. Copy deploy/.env.prod.example to deploy/.env.prod and fill it first." >&2
  exit 1
fi

get_env() {
  local key="$1"
  local value

  value="$(grep -E "^${key}=" "${ENV_FILE}" | tail -n 1 | cut -d= -f2- || true)"
  printf '%s' "${value}"
}

require_env() {
  local key="$1"

  if [[ -z "$(get_env "${key}")" ]]; then
    echo "missing required key in ${ENV_FILE}: ${key}" >&2
    exit 1
  fi
}

require_env APP_DOMAIN
require_env APP_PORT
require_env POSTGRES_DB
require_env POSTGRES_USER
require_env POSTGRES_PASSWORD
require_env REDIS_PASSWORD
require_env AUTH_JWT_SECRET

APP_DOMAIN="$(get_env APP_DOMAIN)"
APP_PORT="$(get_env APP_PORT)"

TMP_NGINX_CONF="$(mktemp)"
cleanup() {
  rm -f "${TMP_NGINX_CONF}"
}
trap cleanup EXIT

sed \
  -e "s|__SERVER_NAME__|${APP_DOMAIN}|g" \
  -e "s|__APP_PORT__|${APP_PORT}|g" \
  "${NGINX_TEMPLATE}" > "${TMP_NGINX_CONF}"

echo "syncing project to ${SSH_TARGET}:${REMOTE_DIR}"
ssh "${SSH_TARGET}" "mkdir -p '${REMOTE_DIR}'"
tar \
  --exclude='.git' \
  --exclude='frontend/node_modules' \
  --exclude='frontend/dist' \
  --exclude='backend/bin' \
  --exclude='deploy/.env.prod' \
  -czf - \
  -C "${REPO_ROOT}" . | ssh "${SSH_TARGET}" "tar -xzf - -C '${REMOTE_DIR}'"

echo "uploading deploy files"
scp "${ENV_FILE}" "${SSH_TARGET}:${REMOTE_DIR}/deploy/.env.prod"
scp "${TMP_NGINX_CONF}" "${SSH_TARGET}:${REMOTE_DIR}/deploy/nginx/moneyapp.conf"

echo "running remote deployment"
ssh "${SSH_TARGET}" "REMOTE_DIR='${REMOTE_DIR}' bash -se" <<'EOF'
set -euo pipefail

cd "${REMOTE_DIR}"

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required on the server" >&2
  exit 1
fi

if ! docker compose version >/dev/null 2>&1; then
  echo "docker compose plugin is required on the server" >&2
  exit 1
fi

if ! command -v nginx >/dev/null 2>&1; then
  echo "nginx is required on the server" >&2
  exit 1
fi

if [ "$(id -u)" -eq 0 ]; then
  SUDO=""
else
  SUDO="sudo"
fi

docker compose --env-file deploy/.env.prod -f deploy/docker-compose.prod.yml pull postgres redis kafka
docker compose --env-file deploy/.env.prod -f deploy/docker-compose.prod.yml build backend
docker compose --env-file deploy/.env.prod -f deploy/docker-compose.prod.yml up -d postgres redis kafka
docker compose --env-file deploy/.env.prod -f deploy/docker-compose.prod.yml run --rm migrate
docker compose --env-file deploy/.env.prod -f deploy/docker-compose.prod.yml up -d backend

$SUDO install -d /etc/nginx/sites-available /etc/nginx/sites-enabled
$SUDO install -m 644 deploy/nginx/moneyapp.conf /etc/nginx/sites-available/moneyapp.conf
$SUDO ln -sfn /etc/nginx/sites-available/moneyapp.conf /etc/nginx/sites-enabled/moneyapp.conf

if [ -e /etc/nginx/sites-enabled/default ]; then
  $SUDO rm -f /etc/nginx/sites-enabled/default
fi

$SUDO nginx -t
$SUDO systemctl enable --now nginx
$SUDO systemctl reload nginx
EOF

echo "deployment completed"
