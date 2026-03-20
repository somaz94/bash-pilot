#!/usr/bin/env bash
set -euo pipefail

DEMO_DIR="/tmp/bash-pilot-demo"

echo "Cleaning up demo resources..."

rm -rf "$DEMO_DIR"

echo "Done! ${DEMO_DIR} removed."
