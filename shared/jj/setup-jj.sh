#!/usr/bin/env zsh

setup_jj() {
  local SCRIPT_DIR=$1
  source "$SCRIPT_DIR/../../utils.sh"

  local JJ_CONFIG

  if [[ "$OSTYPE" == "darwin"* ]]; then
      JJ_CONFIG="$HOME/Library/Application Support/jj/config.toml"
  else
      JJ_CONFIG="$XDG_CONFIG_HOME/jj/config.toml"
  fi


  create_backup $JJ_CONFIG

  cp "$SCRIPT_DIR/config.toml" $JJ_CONFIG
  
  setup_extra_source_scripts "$SCRIPT_DIR"
}

setup_jj $(dirname "$0")