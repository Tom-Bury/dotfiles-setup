#!/usr/bin/env zsh

setup_git() {
  local SCRIPT_DIR=$1
  source "$SCRIPT_DIR/../../utils.sh"

  local GIT_CONFIG="$HOME/.gitconfig"
  create_backup $GIT_CONFIG

  cp "$SCRIPT_DIR/.gitconfig" $GIT_CONFIG
  
  setup_extra_source_scripts "$SCRIPT_DIR"
}

setup_git $(dirname "$0")