#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo "deploy/build_frontend_ssh.sh is deprecated." >&2
echo "use ./deploy/build_frontend_local_to_server.sh or ./deploy/build_frontend_on_server.sh" >&2
exec "${SCRIPT_DIR}/build_frontend_local_to_server.sh" "$@"
