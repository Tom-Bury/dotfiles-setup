#!/usr/bin/env zsh

SETUP_SHELL_SCRIPT_DIR="$(dirname "$0")"
source "$SETUP_SHELL_SCRIPT_DIR/../../utils.sh"

setup_agent_skills() {
  local ROOT_DIR=$1
  source "$ROOT_DIR/utils.sh"

  for skill_dir in "$ROOT_DIR/shared/.agents/skills/"*/; do
    local skill_name=$(basename "$skill_dir")
    local dest_dir="$HOME/.agents/skills/$skill_name"
    safe_create_folder "$dest_dir"
    cp -r "$skill_dir"/* "$dest_dir/"
  done
}
