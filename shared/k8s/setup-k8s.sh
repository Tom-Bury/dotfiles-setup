#!/usr/bin/env zsh

setup_k8s() {
  local ROOT_DIR=$1
  source "$ROOT_DIR/utils.sh"

  setup_extra_source_scripts "$ROOT_DIR/shared/k8s"
}
