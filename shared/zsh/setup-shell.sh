#!/usr/bin/env zsh

SETUP_SHELL_SCRIPT_DIR="$(dirname "$0")"
source "$SETUP_SHELL_SCRIPT_DIR/../../utils.sh"

setup_shell() {
  local ROOT_DIR=$1
  source "$ROOT_DIR/utils.sh"
  install_oh_my_zsh
  setup_starship_prompt $ROOT_DIR
  setup_zshrc

  # Set up autocomplete and syntax highlighting plugins
  # Inspiration: https://gist.github.com/n1snt/454b879b8f0b7995740ae04c5fb5b7df
  setup_autosuggestions
  setup_syntax_highlighting
  setup_zsh_autocomplete

  setup_extra_source_scripts $ROOT_DIR

  source $HOME/.zshrc 
}

install_oh_my_zsh() {
  # Use Oh My ZSH mainly for plugins
  # https://github.com/ohmyzsh/ohmyzsh
  if [ ! -d "$HOME/.oh-my-zsh" ]; then
    sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended
  fi
}

setup_starship_prompt() {
  # https://starship.rs/
  # Installed through homebrew
  create_backup "$HOME/.config/starship.toml"
  cp "$ROOT_DIR/shared/starship/starship.toml" "$HOME/.config/starship.toml"
}

setup_zshrc() {
  # Add ZSH config
  create_backup "$HOME/.zshrc"
  cp "$ROOT_DIR/shared/zsh/.zshrc" "$HOME/.zshrc"
  source $HOME/.zshrc
}

setup_autosuggestions() {
  AUTOSUGGESTIONS_INSTALLATION_DIR=${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins/zsh-autosuggestions
  if [ ! -d $AUTOSUGGESTIONS_INSTALLATION_DIR ]; then
    git clone --depth 1 -- https://github.com/zsh-users/zsh-autosuggestions $AUTOSUGGESTIONS_INSTALLATION_DIR
  fi
}

setup_syntax_highlighting() {
  SYNTAX_HIGHLIGHTING_INSTALLATION_DIR=${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins/fast-syntax-highlighting
  if [ ! -d $SYNTAX_HIGHLIGHTING_INSTALLATION_DIR ]; then
    git clone --depth 1 -- https://github.com/zdharma-continuum/fast-syntax-highlighting.git $SYNTAX_HIGHLIGHTING_INSTALLATION_DIR
  fi
}

setup_zsh_autocomplete() {
  AUTOCOMPLETE_INSTALLATION_DIR=${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins/zsh-autocomplete
  if [ ! -d $AUTOCOMPLETE_INSTALLATION_DIR ]; then
    git clone --depth 1 -- https://github.com/marlonrichert/zsh-autocomplete.git $AUTOCOMPLETE_INSTALLATION_DIR
  fi

  # Autocompletion settings
  zstyle ':autocomplete:*' delay 0.1  # seconds (float)
  zstyle -e ':autocomplete:list-choices:*' list-lines 'reply=( $(( LINES / 3 )) )'
  zstyle ':autocomplete:history-incremental-search-backward:*' list-lines 8
  zstyle ':autocomplete:history-search-backward:*' list-lines 8
}
