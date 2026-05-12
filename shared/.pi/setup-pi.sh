#!/usr/bin/env zsh

setup_pi_config() {
  local ROOT_DIR=$1
  local DEST_HOME=${2:-$HOME}
  local SOURCE_DIR="$ROOT_DIR/shared/.pi/agent"
  local DEST_DIR="$DEST_HOME/.pi/agent"

  if [[ ! -d "$SOURCE_DIR" ]]; then
    echo "No shared pi config found at $SOURCE_DIR"
    return 0
  fi

  mkdir -p "$DEST_DIR"
  cp -a "$SOURCE_DIR/." "$DEST_DIR/"
}
