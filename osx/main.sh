#!/usr/bin/env zsh

# Inspiration - General script
#
# - https://gist.github.com/codeinthehole/26b37efa67041e1307db
# - https://www.lotharschulz.info/2021/05/11/macos-setup-automation-with-homebrew/


main() {
  local SCRIPT_DIR=$1
  source "$SCRIPT_DIR/../utils.sh"

  local ROOT_DIR="$SCRIPT_DIR/.."

  sudo -v # Ask for the administrator password upfront
  
  rm -rf "$HOME/zshrc-scripts"
  
  zsh "$SCRIPT_DIR/../shared/main.sh"

  print_header "Setting up OSX preferences 🖥"
  source "$ROOT_DIR/osx/setup-osx-preferences.sh"
  require_password_on_sleep
  setup_typing_preferences
  setup_finder
  enable_snap_to_grid
  expand_save_and_print_dialogs
  setup_dock
  setup_mission_control
  setup_simulators

  killall Finder
  killall Dock
  print_footer "OSX preferences set up"

  print_header "Installing HomeBrew 🍺"
  source "$ROOT_DIR/osx/homebrew/setup-homebrew.sh"
  install_homebrew
  print_footer "HomeBrew installed"

  print_header "Downloading HomeBrew apps 📱"
  download_homebrew_apps $ROOT_DIR
  print_footer "HomeBrew apps downloaded"

  print_header "Syncing VSCode settings ⚙️"
  source "$ROOT_DIR/osx/vscode/setup-vscode.sh"
  sync_vscode_settings
  print_footer "VSCode settings synced"
}

main $(dirname "$0")
