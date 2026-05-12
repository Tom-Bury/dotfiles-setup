#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET="${1:-}"

if [[ "$TARGET" == "docker" ]]; then
  bash "$ROOT_DIR/docker/main.sh"
elif [[ "${OSTYPE:-}" == darwin* ]]; then
  zsh "$ROOT_DIR/osx/main.sh"
else
  zsh "$ROOT_DIR/linux/main.sh"
fi
