#!/usr/bin/env zsh

main() {
  local SCRIPT_DIR=$1
  source "$SCRIPT_DIR/../utils.sh"

  print_header "Installing packages 📦"
  zsh "$SCRIPT_DIR/install-packages.sh"
  print_footer "Packages installed"

  zsh "$SCRIPT_DIR/../shared/main.sh"

  print_header "Setting up VSCode extensions ⚙️"
  zsh "$SCRIPT_DIR/setup-vscode-extensions.sh"
  print_footer "VSCode extensions set up"
}

main $(dirname "$0")