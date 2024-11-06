###############################################
# ZSH setup
###############################################

export ZSH="$HOME/.oh-my-zsh"
ZSH_THEME=""

plugins=(zsh-autosuggestions fast-syntax-highlighting zsh-autocomplete)
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


###############################################
# Starship prompt https://starship.rs/
###############################################

eval "$(starship init zsh)"

###############################################
# Adjust PATH
###############################################

export PATH="$HOME/bin:$PATH"

###############################################
# Load extra scripts
###############################################

for file in $HOME/zshrc-scripts/.zshrc_*.sh; do
  source $file
done
