#!/usr/bin/env bash

set -euo pipefail

readonly DEFAULT_SSH_TARGET="root@193.187.92.116"
readonly DEFAULT_FRONTEND_DIST_DIR="/root/MoneyApp/frontend/dist"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
LOCAL_BUILD_DIR="$(mktemp -d)"

usage() {
  cat <<EOF
Usage:
  ./deploy/build_frontend_local_to_server.sh [user@server]

Builds frontend/dist on the local machine and uploads it to the server.
Default target: ${DEFAULT_SSH_TARGET}
EOF
}

cleanup() {
  rm -rf "${LOCAL_BUILD_DIR}"
}
trap cleanup EXIT

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

SSH_TARGET="${1:-${DEFAULT_SSH_TARGET}}"

for cmd in docker scp ssh; do
  if ! command -v "${cmd}" >/dev/null 2>&1; then
    echo "missing required command: ${cmd}" >&2
    exit 1
  fi
done

if [[ ! -f "${REPO_ROOT}/scripts/build_frontend_dist.sh" ]]; then
  echo "missing ${REPO_ROOT}/scripts/build_frontend_dist.sh" >&2
  exit 1
fi

echo "building frontend locally into ${LOCAL_BUILD_DIR}"
bash "${REPO_ROOT}/scripts/build_frontend_dist.sh" "${LOCAL_BUILD_DIR}"

echo "preparing ${SSH_TARGET}:${DEFAULT_FRONTEND_DIST_DIR}"
ssh "${SSH_TARGET}" \
  "FRONTEND_DIST_DIR='${DEFAULT_FRONTEND_DIST_DIR}' bash -se" <<'EOF'
set -euo pipefail

mkdir -p "${FRONTEND_DIST_DIR}"
find "${FRONTEND_DIST_DIR}" -mindepth 1 -maxdepth 1 -exec rm -rf -- {} +
EOF

echo "uploading frontend/dist to ${SSH_TARGET}:${DEFAULT_FRONTEND_DIST_DIR}"
scp -r "${LOCAL_BUILD_DIR}/." "${SSH_TARGET}:${DEFAULT_FRONTEND_DIST_DIR}/"

echo "frontend upload completed"
