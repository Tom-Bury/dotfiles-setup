#!/usr/bin/env zsh

# Shows the commit history of the given file in an brief format
function git_hist() {
	git log --format="%Cblue%<(12)%ad%Creset %Cred%<(20)%an%Creset %s" --date=short $1
}

alias gith='git_hist'