#!/usr/bin/env bash

set -euo pipefail

readonly DEFAULT_SSH_TARGET="root@193.187.92.116"
readonly DEFAULT_REMOTE_DIR="/root/MoneyApp"
readonly DEFAULT_FRONTEND_DIST_DIR="/root/MoneyApp/frontend/dist"

usage() {
  cat <<EOF
Usage:
  ./deploy/build_frontend_on_server.sh [user@server]

Builds frontend/dist directly on the server over SSH.
Default target: ${DEFAULT_SSH_TARGET}
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

SSH_TARGET="${1:-${DEFAULT_SSH_TARGET}}"

for cmd in ssh; do
  if ! command -v "${cmd}" >/dev/null 2>&1; then
    echo "missing required command: ${cmd}" >&2
    exit 1
  fi
done

echo "building frontend on ${SSH_TARGET}:${DEFAULT_REMOTE_DIR}"
ssh "${SSH_TARGET}" \
  "REMOTE_DIR='${DEFAULT_REMOTE_DIR}' FRONTEND_DIST_DIR='${DEFAULT_FRONTEND_DIST_DIR}' bash -se" <<'EOF'
set -euo pipefail

cd "${REMOTE_DIR}"

if ! command -v npm >/dev/null 2>&1; then
  echo "npm is required on the server" >&2
  exit 1
fi

if [[ ! -f "./scripts/build_frontend_dist.sh" ]]; then
  echo "missing scripts/build_frontend_dist.sh in ${REMOTE_DIR}" >&2
  exit 1
fi

bash ./scripts/build_frontend_dist.sh "${FRONTEND_DIST_DIR}"
EOF

echo "frontend build completed on server"
