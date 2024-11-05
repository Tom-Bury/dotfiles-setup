#!/usr/bin/env zsh

setup_nvm() {
  local ROOT_DIR=$1
  source "$ROOT_DIR/utils.sh"

  if test ! "$(command -v nvm)"; then
    # https://github.com/nvm-sh/nvm?tab=readme-ov-file#installing-and-updating
    curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh | bash
  fi

  setup_extra_source_scripts "$ROOT_DIR/shared/node"
}
