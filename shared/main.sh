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

  print_header "Creating folders 📂"
  safe_create_folder $HOME/personal/code
  safe_create_folder $HOME/work/code
  print_footer "Folders created"

  print_header "Setting up ZSH shell 🐚"
  source "$SCRIPT_DIR/zsh/setup-shell.sh"
  setup_shell $ROOT_DIR
  print_footer "Shell set up"

  if [ "$NODE" = true ]; then
    print_header "Setting up Node.js with NVM 📦"
    source "$SCRIPT_DIR/node/setup-nvm.sh"
    setup_nvm $ROOT_DIR
    print_footer "Node.js with NVM set up"
  fi

  if [ "$PYTHON" = true ]; then
    print_header "Setting up Python 🐍"
    source "$SCRIPT_DIR/python/setup-python.sh"
    setup_python
    print_footer "Python set up"
  fi
}

main $(dirname "$0")