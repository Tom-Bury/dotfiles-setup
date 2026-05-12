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
  "yazi"
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

install_node_if_needed() {
  if command -v node >/dev/null 2>&1 && command -v npm >/dev/null 2>&1; then
    echo "node and npm are already installed"
    return
  fi

  print_header "Installing node and npm 🟩"
  run_sudo apt update

  if ! command -v node >/dev/null 2>&1; then
    run_sudo apt install -y nodejs
  fi

  if ! command -v npm >/dev/null 2>&1; then
    run_sudo apt install -y npm
  fi
  print_footer "node and npm installed"
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

sync_agents_to_root() {
  print_header "Syncing .agents to root 🤖"
  run_sudo mkdir -p /root/.agents
  run_sudo cp -a "$ROOT_DIR/shared/.agents/." /root/.agents/
  print_footer ".agents synced to /root/.agents"
}

sync_pi_to_root() {
  print_header "Syncing .pi to root 🥧"
  run_sudo mkdir -p /root/.pi/agent
  run_sudo cp -a "$ROOT_DIR/shared/.pi/agent/." /root/.pi/agent/
  print_footer ".pi synced to /root/.pi"
}

main() {
  install_apt_packages
  sync_agents_to_root
  sync_pi_to_root
  install_node_if_needed
  install_global_npm_packages
}

main
