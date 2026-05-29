###############################################
# ZSH setup
###############################################

# should be in .zshenv
export XDG_CONFIG_HOME="$HOME/.config"
export XDG_CACHE_HOME="$HOME/.cache"
export XDG_DATA_HOME="$HOME/.local/share"
export XDG_STATE_HOME="$HOME/.local/state"


export ZSH="$HOME/.oh-my-zsh"
ZSH_THEME=""

# https://github.com/zsh-users/zsh-completions#manual-installation
fpath+=${ZSH_CUSTOM:-${ZSH:-~/.oh-my-zsh}/custom}/plugins/zsh-completions/src
autoload -U compinit && compinit

# In case the compinit gets too slow: https://gist.github.com/ctechols/ca1035271ad134841284

export NVM_LAZY_LOAD=true
export NVM_AUTO_USE=true

# fxf-tab does not interop with zsh-autocomplete
plugins=(fzf-tab fast-syntax-highlighting zsh-nvm zsh-autosuggestions)

source $ZSH/oh-my-zsh.sh

# Fix slowness of pastes with zsh-syntax-highlighting.zsh: https://gist.github.com/magicdude4eva/2d4748f8ef3e6bf7b1591964c201c1ab
pasteinit() {
  OLD_SELF_INSERT=${${(s.:.)widgets[self-insert]}[2,3]}
  zle -N self-insert url-quote-magic # I wonder if you'd need `.url-quote-magic`?
}

pastefinish() {
  zle -N self-insert $OLD_SELF_INSERT
}
zstyle :bracketed-paste-magic paste-init pasteinit
zstyle :bracketed-paste-magic paste-finish pastefinish
# Fix slowness of pastes

# mise
eval "$(mise activate zsh)"

###############################################
# Aliases
###############################################

alias cdp="cd $HOME/personal"
alias cdpc="cd $HOME/personal/code"

alias cdw="cd $HOME/work"
alias cdwc="cd $HOME/work/code"

alias cdd="cd $HOME/Desktop"

alias pbc='pbcopy'
alias pbc-branch='git rev-parse --abbrev-ref HEAD | pbcopy'
alias pbcd='pwd | pbcopy'

alias t1='tree -L 1'
alias t2='tree -L 2'
alias t3='tree -L 3'
alias t4='tree -L 4'

alias lg='lazygit'
alias ld='lazydocker'

# Starts a Docker Sandbox in the current directory, or if one already exists, opens it.
sbx_here() {
  local dir name matches count picked

  if ! command -v sbx >/dev/null 2>&1; then
    echo "sbx not found. Install sbx first." >&2
    return 1
  fi

  dir="$(pwd -P)"
  name="$(basename "$dir")"

  matches="$(
    sbx ls 2>/dev/null | awk -v dir="$dir" '
      NR > 1 && $NF == dir { print $1 }
    '
  )"

  count="$(printf '%s\n' "$matches" | sed '/^$/d' | wc -l | tr -d ' ')"

  case "$count" in
    0)
      echo "No sandbox for $dir. Creating: $name"
      sbx create shell . --name="$name"
      sbx run "$name"
      ;;
    1)
      sbx run "$matches"
      ;;
    *)
      if ! command -v fzf >/dev/null 2>&1; then
        echo "Multiple sandboxes found, but fzf not installed:" >&2
        printf '%s\n' "$matches" >&2
        return 1
      fi

      picked="$(printf '%s\n' "$matches" | fzf --prompt="Pick sandbox: ")"
      [ -n "$picked" ] || return 1

      sbx run "$picked"
      ;;
  esac
}
alias sb='sbx_here'

# https://stackoverflow.com/questions/19331497/set-environment-variables-from-file-of-key-value-pairs
# alias sourceenv="export $(grep -v '^#' .env | xargs)"
# alias unsourceenv="unset $(grep -v '^#' .env | sed -E 's/(.*)=.*/\1/' | xargs)"

###############################################
# Starship prompt https://starship.rs/
###############################################

eval "$(starship init zsh)"

###############################################
# Adjust PATH
###############################################

export PATH="$HOME/bin:$HOME/.docker/bin:$PATH"


###############################################
# Load extra scripts
###############################################

for file in $HOME/zshrc-scripts/.zshrc_*.sh; do
  sched +1 source $file
done

###############################################
# Git-Spice
###############################################

eval "$(git-spice shell completion zsh)"

###############################################
# FZF shell integration
###############################################
# https://github.com/junegunn/fzf#setting-up-shell-integration
  source <(fzf --zsh)