#!/usr/bin/env zsh

main() {
  local SCRIPT_DIR=$1
  local ROOT_DIR="$1/.."
  source "$SCRIPT_DIR/../utils.sh"

  zsh "$ROOT_DIR/setup-preferences.sh"

  set -a; source "$ROOT_DIR/.env"; set +a

  sudo -v # Ask for the administrator password upfront

  print_header "Setting up Git 🐙"
  zsh "$SCRIPT_DIR/git/setup-git.sh"
  print_footer "Git set up"
  
  print_header "Setting up JJ 🥋"
  zsh "$SCRIPT_DIR/jj/setup-jj.sh"
  print_footer "JJ set up"

  print_header "Creating folders 📂"
  safe_create_folder $HOME/personal/code
  safe_create_folder $HOME/work/code
  print_footer "Folders created"

  print_header "Setting up ZSH shell 🐚"
  source "$SCRIPT_DIR/zsh/setup-shell.sh"
  setup_shell $ROOT_DIR
  print_footer "Shell set up"

  if [ "$PYTHON" = true ]; then
    print_header "Setting up Python 🐍"
    source "$SCRIPT_DIR/python/setup-python.sh"
    setup_python
    print_footer "Python set up"
  fi

  if [ "$GO" = true ]; then
    print_header = "Setting up Go"
    source "$ROOT_DIR/shared/go/setup-go.sh"
    setup_go $ROOT_DIR
    print_footer "Go set up"
  fi

  print_header = "Setting up K8s"
  source "$ROOT_DIR/shared/k8s/setup-k8s.sh"
  setup_k8s $ROOT_DIR
  print_footer "K8s set up"
  
  print_header = "Setting up gcloud"
  source "$ROOT_DIR/shared/gcloud/setup-gcloud.sh"
  setup_gcloud $ROOT_DIR
  print_footer "gcloud set up"
}

main $(dirname "$0")