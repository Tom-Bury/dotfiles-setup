#!/usr/bin/env zsh

# Copies the current commit SHA to the clipboard
function git_current_commit() {
	current_commit=$(git rev-parse HEAD)
	echo "Copied current commit SHA: $current_commit"
	echo $current_commit | pbcopy
}

alias gitcc='git_current_commit'