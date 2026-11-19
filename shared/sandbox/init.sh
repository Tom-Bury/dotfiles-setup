# Sourced by /etc/sandbox-persistent.sh when /mnt/dotfiles is mounted.
# Keep installed tools and writable state outside the read-only mount.

_dsb_link() {
  local source="$1"
  local target="$2"

  [[ -e "$source" ]] || return 0
  if [[ -L "$target" && "$(readlink "$target")" == "$source" ]]; then
    return 0
  fi

  # Never delete a directory from a shell startup hook.
  if [[ -d "$target" && ! -L "$target" ]]; then
    return 0
  fi
  mkdir -p "$(dirname "$target")" || return
  ln -sfnT "$source" "$target"
}

_dsb_init() {
  local config
  local pi_source=/mnt/dotfiles/shared/.pi/agent

  _dsb_link /mnt/dotfiles/shared/.agents/skills "$HOME/.agents/skills" || return

  # Repair the old directory link. It hides Pi's runtime and writable state.
  if [[ -L "$HOME/.pi/agent" && "$(readlink "$HOME/.pi/agent")" == "$pi_source" ]]; then
    rm "$HOME/.pi/agent" || return
  fi
  mkdir -p "$HOME/.pi/agent" || return

  # Share config through the read-only mount. Keep runtime state local.
  for config in AGENTS.md APPEND_SYSTEM.md settings.json keybindings.json; do
    _dsb_link "$pi_source/$config" "$HOME/.pi/agent/$config" || return
  done

  # Use the mounted config without changing root-owned template files.
  if [[ -d /mnt/dotfiles/shared/yazi ]]; then
    export YAZI_CONFIG_HOME=/mnt/dotfiles/shared/yazi
  fi

  case ":$PATH:" in
    *":$HOME/.local/bin:"*) ;;
    *) export PATH="$HOME/.local/bin:$PATH" ;;
  esac
}

_dsb_init
unset -f _dsb_init _dsb_link
