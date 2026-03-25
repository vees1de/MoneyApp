#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  ./deploy/deploy_ssh.sh user@server [/opt/moneyapp]

The script expects a filled .env file in the repo root.
It syncs the repository to the server, uploads .env,
and runs docker compose up --build -d remotely.
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
ENV_FILE="${REPO_ROOT}/.env"

for cmd in ssh scp tar; do
  if ! command -v "${cmd}" >/dev/null 2>&1; then
    echo "missing required command: ${cmd}" >&2
    exit 1
  fi
done

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "missing ${ENV_FILE}. Copy .env.example to .env and fill it first." >&2
  exit 1
fi

echo "syncing project to ${SSH_TARGET}:${REMOTE_DIR}"
ssh "${SSH_TARGET}" "mkdir -p '${REMOTE_DIR}'"
tar \
  --exclude='.git' \
  --exclude='frontend/node_modules' \
  --exclude='frontend/dist' \
  --exclude='backend/bin' \
  --exclude='.env' \
  -czf - \
  -C "${REPO_ROOT}" . | ssh "${SSH_TARGET}" "tar -xzf - -C '${REMOTE_DIR}'"

echo "uploading .env"
scp "${ENV_FILE}" "${SSH_TARGET}:${REMOTE_DIR}/.env"

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

docker compose up --build -d
EOF

echo "deployment completed"
