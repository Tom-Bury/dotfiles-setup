#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

source "$ROOT_DIR/utils.sh"

run_sudo() {
  if [[ "${EUID:-$(id -u)}" -eq 0 ]]; then
    "$@"
  elif command -v sudo >/dev/null 2>&1; then
    sudo "$@"
  else
    echo "Error: root privileges required for: $*" >&2
    exit 1
  fi
}

APT_PACKAGES=(
  "ca-certificates"
  "curl"
  "git"
  "micro"
  "unzip",
  "file"
)

NPM_PACKAGES=(
  "@earendil-works/pi-coding-agent"
)

apt_package_available() {
  apt-cache show "$1" >/dev/null 2>&1
}

apt_package_installed() {
  dpkg-query -W -f='${Status}' "$1" 2>/dev/null | grep -q "install ok installed"
}

install_apt_packages() {
  print_header "Updating apt packages 📦"
  run_sudo apt update
  run_sudo apt upgrade -y
  print_footer "apt updated"

  print_header "Installing barebones tools 🧰"
  for package in "${APT_PACKAGES[@]}"; do
    if apt_package_installed "$package"; then
      echo "$package is already installed"
    elif apt_package_available "$package"; then
      echo "Installing $package"
      run_sudo apt install -y "$package"
    else
      echo "Skipping $package: not available in configured apt repositories"
    fi
  done
  print_footer "Barebones tools installed"
}

install_yazi_if_needed() {
  if command -v yazi >/dev/null 2>&1 && command -v ya >/dev/null 2>&1; then
    echo "yazi and ya are already installed"
    return
  fi

  local arch target tmpdir zip_path
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64)
      target="x86_64-unknown-linux-gnu"
      ;;
    aarch64|arm64)
      target="aarch64-unknown-linux-gnu"
      ;;
    *)
      echo "Error: unsupported architecture for yazi release: $arch" >&2
      exit 1
      ;;
  esac

  print_header "Installing yazi from GitHub release 📁"
  tmpdir="$(mktemp -d)"
  zip_path="$tmpdir/yazi.zip"

  curl -fL "https://github.com/sxyazi/yazi/releases/latest/download/yazi-${target}.zip" -o "$zip_path"
  unzip -q "$zip_path" -d "$tmpdir"
  run_sudo install -m 0755 "$tmpdir/yazi-${target}/yazi" /usr/local/bin/yazi
  run_sudo install -m 0755 "$tmpdir/yazi-${target}/ya" /usr/local/bin/ya
  rm -rf "$tmpdir"

  print_footer "yazi installed"
}

install_nvm() {
  unset NPM_CONFIG_PREFIX
  
  local bashrc="$HOME/.bashrc"
  local marker_begin="# >>> dotfiles-setup nvm >>>"
  local marker_end="# <<< dotfiles-setup nvm <<<"

  touch "$bashrc"

  if grep -Fq "$marker_begin" "$bashrc"; then
    echo "nvm already managed in $bashrc"
  else
    {
      printf '\n%s\n' "$marker_begin"
      printf "%s\n" "unset NPM_CONFIG_PREFIX"
      printf '%s\n' "$marker_end"
    } >> "$bashrc"
  fi


  curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.4/install.sh | bash

  source "$bashrc"

  nvm install --lts
  nvm use --lts
  nvm alias default node
}

install_global_npm_packages() {
  print_header "Installing global npm packages 🤖"
  for package in "${NPM_PACKAGES[@]}"; do
    if npm list -g --depth=0 "$package" >/dev/null 2>&1; then
      echo "$package is already installed globally"
    else
      echo "Installing $package globally"
      run_sudo npm install -g "$package"
    fi
  done
  print_footer "Global npm packages installed"
}

sync_agents() {
  print_header "Syncing .agents 🤖"
  run_sudo mkdir -p "$HOME/.agents"
  run_sudo cp -a "$ROOT_DIR/shared/.agents/." "$HOME/.agents/"
  run_sudo chown -R "$(id -un):$(id -gn)" "$HOME/.agents"
  print_footer ".agents synced to $HOME/.agents"
}

sync_pi() {
  print_header "Syncing .pi 🥧"
  run_sudo mkdir -p "$HOME/.pi/agent"
  run_sudo cp -a "$ROOT_DIR/shared/.pi/agent/." "$HOME/.pi/agent/"
  run_sudo chown -R "$(id -un):$(id -gn)" "$HOME/.pi"
  print_footer ".pi synced to $HOME/.pi"
}

sync_yazi() {
  print_header "Syncing yazi config 📁"
  run_sudo mkdir -p "$HOME/.config/yazi"
  run_sudo cp "$ROOT_DIR/shared/yazi/yazi.toml" "$HOME/.config/yazi/yazi.toml"
  print_footer "yazi config synced to $HOME/.config/yazi"
}

sync_yazi_shell_wrapper() {
  local bashrc="$HOME/.bashrc"
  local yazi_shell="$ROOT_DIR/shared/yazi/.zshrc_yazi.sh"
  local marker_begin="# >>> dotfiles-setup yazi shell wrapper >>>"
  local marker_end="# <<< dotfiles-setup yazi shell wrapper <<<"

  print_header "Syncing yazi shell wrapper 📁"
  touch "$bashrc"

  if grep -Fq "$marker_begin" "$bashrc"; then
    echo "yazi shell wrapper already managed in $bashrc"
  elif grep -Eq '^[[:space:]]*(function[[:space:]]+y[[:space:]]*(\(\))?|y[[:space:]]*\(\))[[:space:]]*\{' "$bashrc"; then
    echo "y() already exists in $bashrc; not adding duplicate"
  else
    {
      printf '\n%s\n' "$marker_begin"
      grep -v '^#!' "$yazi_shell"
      printf '%s\n' "$marker_end"
    } >> "$bashrc"
    echo "yazi shell wrapper added to $bashrc"
  fi

  print_footer "yazi shell wrapper synced to $bashrc"
}

main() {
  install_apt_packages
  install_yazi_if_needed
  sync_yazi
  sync_yazi_shell_wrapper
  sync_agents
  sync_pi
  install_nvm
  install_global_npm_packages

  source "$HOME/.bashrc"
}

main
