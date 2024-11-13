#!/usr/bin/env zsh

ROOT_DIR=$(dirname "$0")

if [[ "$OSTYPE" == "darwin"* ]]; then
    zsh "$ROOT_DIR/osx/main.sh"
else
    zsh "$ROOT_DIR/osx/main.sh"
fi
