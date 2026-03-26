#!/usr/bin/env bash

set -euo pipefail

readonly DEFAULT_REMOTE_DIR="/root/MoneyApp"
readonly DEFAULT_FRONTEND_DIST_DIR="/root/MoneyApp/frontend/dist"

usage() {
  cat <<'EOF'
Usage:
  ./deploy/build_frontend_ssh.sh user@server

Builds frontend/dist on the remote server via SSH.
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

SSH_TARGET="${1:-}"

if [[ -z "${SSH_TARGET}" ]]; then
  usage
  exit 1
fi

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

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required on the server" >&2
  exit 1
fi

if [[ ! -f "./scripts/build_frontend_dist.sh" ]]; then
  echo "missing scripts/build_frontend_dist.sh in ${REMOTE_DIR}" >&2
  exit 1
fi

bash ./scripts/build_frontend_dist.sh "${FRONTEND_DIST_DIR}"
EOF

echo "frontend build completed"
