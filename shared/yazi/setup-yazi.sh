#!/usr/bin/env zsh

setup_yazi() {
  local ROOT_DIR=$1
  local yazi_config_dir="$HOME/.config/yazi"
  local yazi_config="$yazi_config_dir/yazi.toml"

  mkdir -p "$yazi_config_dir"
  create_backup "$yazi_config"
  cp "$ROOT_DIR/shared/yazi/yazi.toml" "$yazi_config"
}
