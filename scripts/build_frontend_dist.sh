#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
FRONTEND_DIR="${REPO_ROOT}/frontend"
OUTPUT_DIR="${1:-/opt/moneyapp/frontend/dist}"
ENV_FILE="${REPO_ROOT}/.env"

cleanup() {
  :
}
trap cleanup EXIT

if ! command -v npm >/dev/null 2>&1; then
  echo "npm is required" >&2
  exit 1
fi

if [[ ! -d "${FRONTEND_DIR}" ]]; then
  echo "missing frontend directory: ${FRONTEND_DIR}" >&2
  exit 1
fi

if [[ -f "${ENV_FILE}" ]]; then
  set -a
  # shellcheck disable=SC1090
  . "${ENV_FILE}"
  set +a
fi

if [[ ! -d "${FRONTEND_DIR}/node_modules" ]]; then
  echo "installing frontend dependencies"
  (
    cd "${FRONTEND_DIR}"
    npm ci
  )
fi

echo "building frontend in ${FRONTEND_DIR}"
(
  cd "${FRONTEND_DIR}"
  npm run build
)

if [[ ! -d "${FRONTEND_DIR}/dist" ]]; then
  echo "frontend build did not produce ${FRONTEND_DIR}/dist" >&2
  exit 1
fi

rm -rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"
cp -R "${FRONTEND_DIR}/dist/." "${OUTPUT_DIR}/"
chmod -R a+rX "${OUTPUT_DIR}"
