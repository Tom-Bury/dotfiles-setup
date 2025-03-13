#!/usr/bin/env zsh

setup_gcloud() {
  local ROOT_DIR=$1
  source "$ROOT_DIR/utils.sh"

  setup_extra_source_scripts "$ROOT_DIR/shared/gcloud"
}
