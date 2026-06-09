#!/usr/bin/env zsh

SETUP_SHELL_SCRIPT_DIR="$(dirname "$0")"
source "$SETUP_SHELL_SCRIPT_DIR/../../utils.sh"

setup_lazygit() {
  local ROOT_DIR=$1
  ln -s "$ROOT_DIR/shared/git/lazygit.yaml" "$HOME/.config/lazygit/config.yml"
}
