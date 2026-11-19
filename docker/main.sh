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
  "libatomic1"
  "micro"
  "neovim"
  "unzip"
  "file"
)

NPM_PACKAGES=(
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

install_fd() {
  print_header "Installing fd"
  if ! command -v fd >/dev/null 2>&1; then
    run_sudo apt install -y fd-find
    # Debian names the binary fdfind. Expose the standard name for other tools.
    run_sudo ln -sfn "$(command -v fdfind)" /usr/local/bin/fd
  fi
  fd --version
  print_footer "fd is ready"
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

install_go() {
  local version="1.27.4"
  local arch archive tmpdir installed_version

  if command -v go >/dev/null 2>&1; then
    installed_version="$(go version | awk '{print $3}')"
    if [[ "$installed_version" == "go$version" ]]; then
      echo "Go $version is already installed"
      return
    fi
  fi

  case "$(uname -m)" in
    x86_64|amd64)
      arch="amd64"
      ;;
    aarch64|arm64)
      arch="arm64"
      ;;
    *)
      echo "Error: unsupported architecture for Go: $(uname -m)" >&2
      exit 1
      ;;
  esac

  archive="go${version}.linux-${arch}.tar.gz"
  tmpdir="$(mktemp -d)"

  print_header "Installing Go $version 🐹"
  curl -fsSL "https://go.dev/dl/$archive" -o "$tmpdir/$archive"
  run_sudo rm -rf /usr/local/go
  run_sudo tar -C /usr/local -xzf "$tmpdir/$archive"
  run_sudo ln -sf /usr/local/go/bin/go /usr/local/bin/go
  run_sudo ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt
  rm -rf "$tmpdir"
  print_footer "Go $version installed"
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

  set +u
  source "$bashrc"

  nvm install --lts
  nvm use --lts
  nvm alias default node
  set -u
}

install_global_npm_packages() {
  print_header "Installing global npm packages 🤖"
  set +u
  nvm use --lts
  for package in "${NPM_PACKAGES[@]}"; do
    if npm list -g --depth=0 "$package" >/dev/null 2>&1; then
      echo "$package is already installed globally"
    else
      echo "Installing $package globally"
      npm install -g "$package"
    fi
  done
  set -u
  print_footer "Global npm packages installed"
}

sync_agents() {
  print_header "Syncing .agents 🤖"
  run_sudo mkdir -p "$HOME/.agents"
  run_sudo cp -a "$ROOT_DIR/shared/.agents/." "$HOME/.agents/"
  run_sudo chown -R "$(id -un):$(id -gn)" "$HOME/.agents"
  print_footer ".agents synced to $HOME/.agents"
}

managed_pi_installed() {
  local install_root="$HOME/.pi/agent/install"
  local launcher="$HOME/.local/bin/pi"

  [[ -x "$launcher" ]] \
    && [[ -f "$install_root/managed-install.json" ]] \
    && grep -Fq '"kind": "pi-managed-install"' "$install_root/managed-install.json"
}

sync_pi() {
  print_header "Syncing .pi 🥧"

  # The curl installer owns ~/.pi/agent/install and ~/.local/bin/pi. Re-running
  # it after migration can rediscover an old npm installation and prompt again.
  if managed_pi_installed; then
    echo "Pi managed installation already present"
  else
    curl -fsSL https://pi.dev/install.sh | sh
  fi

  run_sudo mkdir -p "$HOME/.pi/agent"
  run_sudo cp -a "$ROOT_DIR/shared/.pi/agent/." "$HOME/.pi/agent/"
  run_sudo chown -R "$(id -un):$(id -gn)" "$HOME/.pi"
  print_footer ".pi synced to $HOME/.pi"
}

sync_yazi() {
  print_header "Syncing yazi config 📁"
  run_sudo mkdir -p "$HOME/.config/yazi"
  run_sudo cp "$ROOT_DIR/shared/yazi/yazi.toml" "$HOME/.config/yazi/yazi.toml"
  run_sudo chown -R "$(id -un):$(id -gn)" "$HOME/.config/yazi"
  print_footer "yazi config synced to $HOME/.config/yazi"
}

sync_default_editor() {
  local bashrc="$HOME/.bashrc"
  local marker_begin="# >>> dotfiles-setup default editor >>>"
  local marker_end="# <<< dotfiles-setup default editor <<<"

  print_header "Setting default editor to nvim 📝"
  touch "$bashrc"

  if grep -Fq "$marker_begin" "$bashrc"; then
    echo "default editor already managed in $bashrc"
  else
    {
      printf '\n%s\n' "$marker_begin"
      printf '%s\n' 'export EDITOR="nvim"'
      printf '%s\n' 'export VISUAL="nvim"'
      printf '%s\n' 'export SUDO_EDITOR="nvim"'
      printf '%s\n' "$marker_end"
    } >> "$bashrc"
    echo "default editor added to $bashrc"
  fi

  print_footer "Default editor synced to $bashrc"
}

configure_pi_clipboard() {
  local persistent_file="/etc/sandbox-persistent.sh"
  local setting='export PI_CLIPBOARD_OSC52=1'

  # Send copies to the host terminal, even if the sandbox advertises Wayland.
  if ! run_sudo grep -Fxq "$setting" "$persistent_file" 2>/dev/null; then
    printf '\n%s\n' "$setting" | run_sudo tee -a "$persistent_file" >/dev/null
  fi
  export PI_CLIPBOARD_OSC52=1
}

install_sandbox_mount_init() {
  local persistent_file="/etc/sandbox-persistent.sh"
  local marker_begin="# >>> dotfiles-setup sandbox mounts >>>"
  local marker_end="# <<< dotfiles-setup sandbox mounts <<<"
  local tmpfile

  if run_sudo grep -Fq "$marker_begin" "$persistent_file" 2>/dev/null; then
    echo "sandbox mount init already registered"
    return
  fi

  tmpfile="$(mktemp)"
  {
    printf '\n%s\n' "$marker_begin"
    printf '%s\n' 'if [[ -r /mnt/dotfiles/shared/sandbox/init.sh ]]; then'
    printf '%s\n' '  source /mnt/dotfiles/shared/sandbox/init.sh'
    printf '%s\n' 'fi'
    printf '%s\n' "$marker_end"
  } > "$tmpfile"
  run_sudo tee -a "$persistent_file" < "$tmpfile" >/dev/null
  rm -f "$tmpfile"
  echo "sandbox mount init registered"
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
  install_fd
  install_yazi_if_needed
  install_go
  sync_yazi
  sync_yazi_shell_wrapper
  sync_default_editor
  install_sandbox_mount_init
  configure_pi_clipboard
  sync_agents
  install_nvm
  sync_pi
  install_global_npm_packages
}

main
