# Sourced by /etc/sandbox-persistent.sh when a Docker Sandbox has the
# dotfiles repository mounted at /mnt/dotfiles. Keep this safe to run for
# every interactive shell and agent session.

_dsb_link() {
  local source="$1"
  local target="$2"

  [[ -d "$source" ]] || return 0
  if [[ -L "$target" && "$(readlink "$target")" == "$source" ]]; then
    return 0
  fi

  rm -rf "$target"
  mkdir -p "$(dirname "$target")"
  ln -s "$source" "$target"
}

# Add more _dsb_link calls here when sandbox configuration should come from
# the read-only mounted dotfiles repository rather than the base template.
_dsb_link /mnt/dotfiles/shared/.agents "$HOME/.agents"
_dsb_link /mnt/dotfiles/shared/.pi/agent "$HOME/.pi/agent"
_dsb_link /mnt/dotfiles/shared/yazi "$HOME/.config/yazi"

unset -f _dsb_link
