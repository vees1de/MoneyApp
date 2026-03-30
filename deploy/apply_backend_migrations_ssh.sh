#!/usr/bin/env bash

set -euo pipefail

readonly DEFAULT_SSH_TARGET="root@193.187.92.116"
readonly DEFAULT_REMOTE_DIR="/root/MoneyApp"

usage() {
  cat <<EOF
Usage:
  ./deploy/apply_backend_migrations_ssh.sh [user@server] [remote-dir]

Applies backend PostgreSQL migrations on the remote server via docker compose.

Default target: ${DEFAULT_SSH_TARGET}
Default remote dir: ${DEFAULT_REMOTE_DIR}
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

SSH_TARGET="${1:-${DEFAULT_SSH_TARGET}}"
REMOTE_DIR="${2:-${DEFAULT_REMOTE_DIR}}"

for cmd in ssh; do
  if ! command -v "${cmd}" >/dev/null 2>&1; then
    echo "missing required command: ${cmd}" >&2
    exit 1
  fi
done

echo ">>> applying backend migrations on ${SSH_TARGET}:${REMOTE_DIR}"
ssh "${SSH_TARGET}" \
  "REMOTE_DIR='${REMOTE_DIR}' bash -se" <<'REMOTE'
set -euo pipefail

cd "${REMOTE_DIR}"

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required on the server" >&2
  exit 1
fi

echo "  ensuring postgres is running"
docker compose up -d postgres

echo "  running migrate service"
if ! docker compose run --rm migrate; then
  echo "--- migrate logs ---" >&2
  docker compose logs --tail=120 migrate >&2 || true
  exit 1
fi

echo "  migrations applied successfully"
REMOTE

echo ">>> done"
