#!/usr/bin/env zsh

setup_go() {
  local ROOT_DIR=$1
  source "$ROOT_DIR/utils.sh"

  if test ! "$(command -v gvm)"; then
    # https://github.com/moovweb/gvm
    bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
  fi

  setup_extra_source_scripts "$ROOT_DIR/shared/go"
}
