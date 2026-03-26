#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  ./deploy/deploy_ssh.sh user@server [/opt/moneyapp]

The script expects a filled .env file in the repo root.
It makes the server clone/fetch/reset the git repository,
uploads .env, and runs docker compose up --build -d remotely.
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
DEPLOY_REPO_URL="$(git -C "${REPO_ROOT}" remote get-url origin)"
DEPLOY_GIT_REF="$(git -C "${REPO_ROOT}" branch --show-current)"

for cmd in ssh scp git; do
  if ! command -v "${cmd}" >/dev/null 2>&1; then
    echo "missing required command: ${cmd}" >&2
    exit 1
  fi
done

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "missing ${ENV_FILE}. Copy .env.example to .env and fill it first." >&2
  exit 1
fi

if [[ -z "${DEPLOY_REPO_URL}" ]]; then
  echo "failed to detect origin remote URL" >&2
  exit 1
fi

if [[ -z "${DEPLOY_GIT_REF}" ]]; then
  echo "failed to detect current git branch" >&2
  exit 1
fi

echo "syncing git repo on ${SSH_TARGET}:${REMOTE_DIR}"
ssh "${SSH_TARGET}" \
  "REMOTE_DIR='${REMOTE_DIR}' DEPLOY_REPO_URL='${DEPLOY_REPO_URL}' DEPLOY_GIT_REF='${DEPLOY_GIT_REF}' bash -se" <<'EOF'
set -euo pipefail

if ! command -v git >/dev/null 2>&1; then
  echo "git is required on the server" >&2
  exit 1
fi

mkdir -p "${REMOTE_DIR}"

if [[ -d "${REMOTE_DIR}/.git" ]]; then
  echo "  updating existing git checkout"
  git -C "${REMOTE_DIR}" remote set-url origin "${DEPLOY_REPO_URL}"
else
  existing_entry=$(
    find "${REMOTE_DIR}" -mindepth 1 -maxdepth 1 \
      ! -name '.env' \
      ! -name '.env.*' \
      -print -quit
  )
  if [[ -n "${existing_entry}" ]]; then
    echo "target directory ${REMOTE_DIR} is not empty and is not a git repository" >&2
    echo "move it away or clean it manually before the first git-based deploy" >&2
    exit 1
  fi

  echo "  cloning repository"
  git clone --branch "${DEPLOY_GIT_REF}" --single-branch "${DEPLOY_REPO_URL}" "${REMOTE_DIR}"
fi

git -C "${REMOTE_DIR}" fetch --prune origin
git -C "${REMOTE_DIR}" checkout -B "${DEPLOY_GIT_REF}" "origin/${DEPLOY_GIT_REF}"
git -C "${REMOTE_DIR}" reset --hard "origin/${DEPLOY_GIT_REF}"
git -C "${REMOTE_DIR}" clean -fd
EOF

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

echo "building frontend/dist for nginx"
bash scripts/build_frontend_dist.sh /opt/moneyapp/frontend/dist

docker compose up --build -d
EOF

echo "deployment completed"
