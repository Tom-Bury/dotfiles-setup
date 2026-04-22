#!/usr/bin/env zsh

SETUP_SHELL_SCRIPT_DIR="$(dirname "$0")"
source "$SETUP_SHELL_SCRIPT_DIR/../../utils.sh"

setup_lazygit() {
  local ROOT_DIR=$1
  source "$ROOT_DIR/utils.sh"

  TARGET="$HOME/Library/Application Support/lazygit/config.yml"
  create_backup "$TARGET"
  cp "$ROOT_DIR/shared/git/lazygit.yaml" "$TARGET"
}
