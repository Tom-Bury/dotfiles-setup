#!/usr/bin/env zsh

# Copies the current commit SHA to the clipboard
function git_current_commit() {
	current_commit=$(git rev-parse HEAD)
	echo "Copied current commit SHA: $current_commit"
	echo $current_commit | pbcopy
}

alias gitcc='git_current_commit'

function git_last_tag() {
	tag=$(git describe --tags --abbrev=0)
	echo "Copied Git Tag: $tag"
	echo $tag | pbcopy
}

alias gitct='git_last_tag'

function git_create_push_tag() {
	git tag $1
	git push origin $1
}

alias gitpt='git_create_push_tag'