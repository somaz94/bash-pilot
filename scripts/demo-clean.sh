#!/usr/bin/env bash
# Re-exec under bash if invoked via zsh.
if [ -n "${ZSH_VERSION:-}" ]; then exec bash "$0" "$@"; fi
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=./_demo-common.sh
. "${SCRIPT_DIR}/_demo-common.sh"

echo "Cleaning up demo resources..."

# Guard against accidental rm -rf on an empty value.
rm -rf "${DEMO_DIR:?DEMO_DIR must be set}"

echo "Done! ${DEMO_DIR} removed."
