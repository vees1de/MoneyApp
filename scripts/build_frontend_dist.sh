#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
OUTPUT_DIR="${1:-/opt/moneyapp/frontend/dist}"
IMAGE_TAG="${IMAGE_TAG:-moneyapp-frontend-dist:build}"
TMP_DIR="$(mktemp -d)"
CONTAINER_ID=""

cleanup() {
  if [[ -n "${CONTAINER_ID}" ]]; then
    docker rm -f "${CONTAINER_ID}" >/dev/null 2>&1 || true
  fi
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

mkdir -p "$(dirname "${OUTPUT_DIR}")"

docker build \
  --target frontend-dist \
  -f "${REPO_ROOT}/deploy/Dockerfile" \
  -t "${IMAGE_TAG}" \
  "${REPO_ROOT}" >/dev/null

CONTAINER_ID="$(docker create "${IMAGE_TAG}")"
docker cp "${CONTAINER_ID}:/." "${TMP_DIR}"

rm -rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"
cp -R "${TMP_DIR}/." "${OUTPUT_DIR}/"
chmod -R a+rX "${OUTPUT_DIR}"
