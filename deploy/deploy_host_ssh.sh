#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  ./deploy/deploy_host_ssh.sh user@server [/opt/moneyapp]

The script expects deploy/.env.host to exist locally.
It builds the frontend and Linux backend binary locally,
uploads them to the server, bootstraps PostgreSQL, applies
migrations, installs a systemd unit, and optionally installs
an nginx site config for host-level deployments without Docker.
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
ENV_FILE="${SCRIPT_DIR}/.env.host"
SYSTEMD_TEMPLATE="${SCRIPT_DIR}/systemd/moneyapp.service.template"
NGINX_TEMPLATE="${SCRIPT_DIR}/nginx/moneyapp.host.ssl.conf.template"

for cmd in ssh scp tar mktemp sed grep cut tail awk npm go bash; do
  if ! command -v "${cmd}" >/dev/null 2>&1; then
    echo "missing required command: ${cmd}" >&2
    exit 1
  fi
done

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "missing ${ENV_FILE}. Copy deploy/.env.host.example to deploy/.env.host and fill it first." >&2
  exit 1
fi

get_env() {
  local key="$1"
  local value

  value="$(grep -E "^${key}=" "${ENV_FILE}" | tail -n 1 | cut -d= -f2- || true)"
  printf '%s' "${value}"
}

get_env_or_default() {
  local key="$1"
  local fallback="$2"
  local value

  value="$(get_env "${key}")"
  if [[ -z "${value}" ]]; then
    printf '%s' "${fallback}"
    return
  fi

  printf '%s' "${value}"
}

quote_env_value() {
  local value="$1"
  value="${value//\'/\'\"\'\"\'}"
  printf "'%s'" "${value}"
}

require_env() {
  local key="$1"

  if [[ -z "$(get_env "${key}")" ]]; then
    echo "missing required key in ${ENV_FILE}: ${key}" >&2
    exit 1
  fi
}

contains_dsn_sensitive_chars() {
  local value="$1"
  printf '%s' "${value}" | grep -q '[/:@?&%=#]'
}

require_env APP_DOMAIN
require_env APP_PORT
require_env POSTGRES_DB
require_env POSTGRES_USER
require_env POSTGRES_PASSWORD
require_env AUTH_JWT_SECRET

APP_DOMAIN="$(get_env APP_DOMAIN)"
APP_PORT="$(get_env APP_PORT)"
APP_USER="$(get_env_or_default APP_USER moneyapp)"
APP_GROUP="$(get_env_or_default APP_GROUP "${APP_USER}")"
SYSTEMD_SERVICE_NAME="$(get_env_or_default SYSTEMD_SERVICE_NAME moneyapp)"
NGINX_SITE_NAME="$(get_env_or_default NGINX_SITE_NAME moneyapp.conf)"
TARGET_GOOS="$(get_env_or_default TARGET_GOOS linux)"
TARGET_GOARCH="$(get_env_or_default TARGET_GOARCH amd64)"
INSTALL_NGINX="$(get_env_or_default INSTALL_NGINX true)"
BOOTSTRAP_POSTGRES="$(get_env_or_default BOOTSTRAP_POSTGRES true)"
POSTGRES_HOST="$(get_env_or_default POSTGRES_HOST 127.0.0.1)"
POSTGRES_PORT="$(get_env_or_default POSTGRES_PORT 5432)"
POSTGRES_DB="$(get_env POSTGRES_DB)"
POSTGRES_USER="$(get_env POSTGRES_USER)"
POSTGRES_PASSWORD="$(get_env POSTGRES_PASSWORD)"
SSL_PRIMARY_DOMAIN="$(get_env_or_default SSL_PRIMARY_DOMAIN "${APP_DOMAIN%% *}")"
SSL_CERT_PATH="$(get_env_or_default SSL_CERT_PATH "/etc/letsencrypt/live/${SSL_PRIMARY_DOMAIN}/fullchain.pem")"
SSL_CERT_KEY_PATH="$(get_env_or_default SSL_CERT_KEY_PATH "/etc/letsencrypt/live/${SSL_PRIMARY_DOMAIN}/privkey.pem")"

DATABASE_DSN="$(get_env DATABASE_DSN)"
if [[ -z "${DATABASE_DSN}" ]]; then
  if contains_dsn_sensitive_chars "${POSTGRES_USER}" || contains_dsn_sensitive_chars "${POSTGRES_PASSWORD}"; then
    echo "DATABASE_DSN must be set explicitly when POSTGRES_USER or POSTGRES_PASSWORD contains URL-sensitive characters." >&2
    exit 1
  fi
  DATABASE_DSN="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable"
fi

TMP_DIR="$(mktemp -d)"
STAGE_DIR="${TMP_DIR}/stage"
BIN_DIR="${STAGE_DIR}/bin"
FRONTEND_DIR="${STAGE_DIR}/frontend"
MIGRATIONS_DIR="${STAGE_DIR}/backend/migrations"
RENDERED_DIR="${STAGE_DIR}/rendered"
mkdir -p "${BIN_DIR}" "${FRONTEND_DIR}" "${MIGRATIONS_DIR}" "${RENDERED_DIR}"

cleanup() {
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

echo "building frontend"
(
  cd "${REPO_ROOT}/frontend"
  npm ci
  npm run build
)

echo "building backend binary for ${TARGET_GOOS}/${TARGET_GOARCH}"
(
  cd "${REPO_ROOT}/backend"
  env CGO_ENABLED=0 GOOS="${TARGET_GOOS}" GOARCH="${TARGET_GOARCH}" go build -o "${BIN_DIR}/moneyapp" ./cmd/api
)

cp -R "${REPO_ROOT}/frontend/dist" "${FRONTEND_DIR}/dist"
cp "${REPO_ROOT}/backend/migrations/"*.sql "${MIGRATIONS_DIR}/"

FRONTEND_ROOT="${REMOTE_DIR}/frontend/dist"
ENV_TARGET="/etc/${SYSTEMD_SERVICE_NAME}.env"
SYSTEMD_TARGET="/etc/systemd/system/${SYSTEMD_SERVICE_NAME}.service"
NGINX_TARGET="/etc/nginx/sites-available/${NGINX_SITE_NAME}"

cat > "${RENDERED_DIR}/app.env" <<EOF
APP_NAME=$(quote_env_value "$(get_env_or_default APP_NAME moneyapp-backend)")
APP_ENV=$(quote_env_value "$(get_env_or_default APP_ENV production)")
HTTP_ADDR=$(quote_env_value "127.0.0.1:${APP_PORT}")
FRONTEND_DIST_DIR=$(quote_env_value "${FRONTEND_ROOT}")
POSTGRES_HOST=$(quote_env_value "${POSTGRES_HOST}")
POSTGRES_PORT=$(quote_env_value "${POSTGRES_PORT}")
POSTGRES_DB=$(quote_env_value "${POSTGRES_DB}")
POSTGRES_USER=$(quote_env_value "${POSTGRES_USER}")
POSTGRES_PASSWORD=$(quote_env_value "${POSTGRES_PASSWORD}")
DATABASE_DSN=$(quote_env_value "${DATABASE_DSN}")
DATABASE_MAX_OPEN_CONNS=$(quote_env_value "$(get_env_or_default DATABASE_MAX_OPEN_CONNS 20)")
DATABASE_MAX_IDLE_CONNS=$(quote_env_value "$(get_env_or_default DATABASE_MAX_IDLE_CONNS 10)")
DATABASE_CONN_MAX_LIFETIME=$(quote_env_value "$(get_env_or_default DATABASE_CONN_MAX_LIFETIME 30m)")
REDIS_ENABLED=$(quote_env_value "$(get_env_or_default REDIS_ENABLED false)")
REDIS_ADDR=$(quote_env_value "$(get_env_or_default REDIS_ADDR 127.0.0.1:6379)")
REDIS_PASSWORD=$(quote_env_value "$(get_env_or_default REDIS_PASSWORD "")")
REDIS_DB=$(quote_env_value "$(get_env_or_default REDIS_DB 0)")
REDIS_DASHBOARD_TTL=$(quote_env_value "$(get_env_or_default REDIS_DASHBOARD_TTL 30s)")
KAFKA_ENABLED=$(quote_env_value "$(get_env_or_default KAFKA_ENABLED false)")
KAFKA_BROKERS=$(quote_env_value "$(get_env_or_default KAFKA_BROKERS 127.0.0.1:9092)")
KAFKA_CLIENT_ID=$(quote_env_value "$(get_env_or_default KAFKA_CLIENT_ID moneyapp-backend)")
KAFKA_AUDIT_TOPIC=$(quote_env_value "$(get_env_or_default KAFKA_AUDIT_TOPIC moneyapp.audit)")
KAFKA_WRITE_TIMEOUT=$(quote_env_value "$(get_env_or_default KAFKA_WRITE_TIMEOUT 5s)")
AUTH_JWT_SECRET=$(quote_env_value "$(get_env AUTH_JWT_SECRET)")
AUTH_JWT_ISSUER=$(quote_env_value "$(get_env_or_default AUTH_JWT_ISSUER moneyapp)")
AUTH_ACCESS_TOKEN_TTL=$(quote_env_value "$(get_env_or_default AUTH_ACCESS_TOKEN_TTL 15m)")
AUTH_REFRESH_TOKEN_TTL=$(quote_env_value "$(get_env_or_default AUTH_REFRESH_TOKEN_TTL 720h)")
AUTH_ALLOW_INSECURE_DEV_AUTH=$(quote_env_value "$(get_env_or_default AUTH_ALLOW_INSECURE_DEV_AUTH false)")
DEFAULT_BASE_CURRENCY=$(quote_env_value "$(get_env_or_default DEFAULT_BASE_CURRENCY RUB)")
DEFAULT_TIMEZONE=$(quote_env_value "$(get_env_or_default DEFAULT_TIMEZONE Europe/Moscow)")
DEFAULT_WEEKLY_REVIEW_HOUR=$(quote_env_value "$(get_env_or_default DEFAULT_WEEKLY_REVIEW_HOUR 18)")
TELEGRAM_BOT_TOKEN=$(quote_env_value "$(get_env_or_default TELEGRAM_BOT_TOKEN "")")
YANDEX_CLIENT_ID=$(quote_env_value "$(get_env_or_default YANDEX_CLIENT_ID "")")
YANDEX_CLIENT_SECRET=$(quote_env_value "$(get_env_or_default YANDEX_CLIENT_SECRET "")")
YANDEX_REDIRECT_URL=$(quote_env_value "$(get_env_or_default YANDEX_REDIRECT_URL "")")
EOF

sed \
  -e "s|__APP_USER__|${APP_USER}|g" \
  -e "s|__APP_GROUP__|${APP_GROUP}|g" \
  -e "s|__REMOTE_DIR__|${REMOTE_DIR}|g" \
  -e "s|__ENV_FILE__|${ENV_TARGET}|g" \
  "${SYSTEMD_TEMPLATE}" > "${RENDERED_DIR}/moneyapp.service"

sed \
  -e "s|__SERVER_NAME__|${APP_DOMAIN}|g" \
  -e "s|__APP_PORT__|${APP_PORT}|g" \
  -e "s|__FRONTEND_ROOT__|${FRONTEND_ROOT}|g" \
  -e "s|__SSL_CERT_PATH__|${SSL_CERT_PATH}|g" \
  -e "s|__SSL_CERT_KEY_PATH__|${SSL_CERT_KEY_PATH}|g" \
  "${NGINX_TEMPLATE}" > "${RENDERED_DIR}/nginx.conf"

echo "uploading build artifacts to ${SSH_TARGET}:${REMOTE_DIR}"
ssh "${SSH_TARGET}" "mkdir -p '${REMOTE_DIR}'"
tar -czf - -C "${STAGE_DIR}" . | ssh "${SSH_TARGET}" "tar -xzf - -C '${REMOTE_DIR}'"

echo "running remote deployment"
ssh "${SSH_TARGET}" \
  "REMOTE_DIR='${REMOTE_DIR}' APP_USER='${APP_USER}' APP_GROUP='${APP_GROUP}' SYSTEMD_SERVICE_NAME='${SYSTEMD_SERVICE_NAME}' ENV_TARGET='${ENV_TARGET}' SYSTEMD_TARGET='${SYSTEMD_TARGET}' NGINX_TARGET='${NGINX_TARGET}' INSTALL_NGINX='${INSTALL_NGINX}' BOOTSTRAP_POSTGRES='${BOOTSTRAP_POSTGRES}' bash -se" <<'EOF'
set -euo pipefail

if [ "$(id -u)" -eq 0 ]; then
  SUDO=""
else
  SUDO="sudo"
fi

for cmd in install systemctl; do
  if ! command -v "${cmd}" >/dev/null 2>&1; then
    echo "missing required command on server: ${cmd}" >&2
    exit 1
  fi
done

if ! command -v psql >/dev/null 2>&1; then
  echo "psql is required on the server. Install PostgreSQL client/server first." >&2
  exit 1
fi

$SUDO install -d -m 755 "${REMOTE_DIR}" "${REMOTE_DIR}/bin" "${REMOTE_DIR}/frontend" "${REMOTE_DIR}/backend"

if ! getent group "${APP_GROUP}" >/dev/null 2>&1; then
  $SUDO groupadd --system "${APP_GROUP}"
fi

if ! id -u "${APP_USER}" >/dev/null 2>&1; then
  $SUDO useradd --system --home "${REMOTE_DIR}" --shell /usr/sbin/nologin --gid "${APP_GROUP}" "${APP_USER}"
fi

$SUDO install -d -o "${APP_USER}" -g "${APP_GROUP}" -m 755 "${REMOTE_DIR}" "${REMOTE_DIR}/bin" "${REMOTE_DIR}/frontend" "${REMOTE_DIR}/backend" "${REMOTE_DIR}/backend/migrations"
$SUDO chown -R "${APP_USER}:${APP_GROUP}" "${REMOTE_DIR}"
$SUDO chmod 755 "${REMOTE_DIR}/bin/moneyapp"

$SUDO install -m 600 "${REMOTE_DIR}/rendered/app.env" "${ENV_TARGET}"
$SUDO install -m 644 "${REMOTE_DIR}/rendered/moneyapp.service" "${SYSTEMD_TARGET}"

set -a
. "${ENV_TARGET}"
set +a

if [ "${BOOTSTRAP_POSTGRES}" = "true" ]; then
  $SUDO -u postgres psql -v ON_ERROR_STOP=1 \
    -v db_name="${POSTGRES_DB}" \
    -v db_user="${POSTGRES_USER}" \
    -v db_password="${POSTGRES_PASSWORD}" <<'SQL'
SELECT format('CREATE ROLE %I LOGIN PASSWORD %L', :'db_user', :'db_password')
WHERE NOT EXISTS (
  SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = :'db_user'
)\gexec

SELECT format('ALTER ROLE %I WITH LOGIN PASSWORD %L', :'db_user', :'db_password')
WHERE EXISTS (
  SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = :'db_user'
)\gexec

SELECT format('CREATE DATABASE %I OWNER %I', :'db_name', :'db_user')
WHERE NOT EXISTS (
  SELECT 1 FROM pg_database WHERE datname = :'db_name'
)\gexec
SQL
fi

for file in "${REMOTE_DIR}"/backend/migrations/*.sql; do
  if [ ! -f "${file}" ]; then
    continue
  fi

  PGPASSWORD="${POSTGRES_PASSWORD}" \
    psql \
      -v ON_ERROR_STOP=1 \
      -h "${POSTGRES_HOST}" \
      -p "${POSTGRES_PORT}" \
      -U "${POSTGRES_USER}" \
      -d "${POSTGRES_DB}" \
      -f "${file}"
done

$SUDO systemctl daemon-reload
$SUDO systemctl enable --now "${SYSTEMD_SERVICE_NAME}"
$SUDO systemctl restart "${SYSTEMD_SERVICE_NAME}"

if [ "${INSTALL_NGINX}" = "true" ]; then
  if ! command -v nginx >/dev/null 2>&1; then
    echo "nginx is required on the server when INSTALL_NGINX=true" >&2
    exit 1
  fi

  $SUDO install -d /etc/nginx/sites-available /etc/nginx/sites-enabled
  $SUDO install -m 644 "${REMOTE_DIR}/rendered/nginx.conf" "${NGINX_TARGET}"
  $SUDO ln -sfn "${NGINX_TARGET}" "/etc/nginx/sites-enabled/$(basename "${NGINX_TARGET}")"
  $SUDO nginx -t
  $SUDO systemctl enable --now nginx
  $SUDO systemctl reload nginx
fi
EOF

echo "deployment completed"
