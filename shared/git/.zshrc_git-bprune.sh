#!/usr/bin/env zsh

function git_prune_branches() {
    git fetch --prune
	local excludePattern="\*\\|main\\|master\\|develop\\|development\\|prod\\|production"
    echo "About to delete the following branches:"
    git branch --merged | grep -v $excludePattern
    
    print -r -- "Confirm (y/N)? "
    read confirmation
    
    if [ "$confirmation" = "y" ]; then
        git branch --merged | grep -v $excludePattern | xargs -n 1 git branch -d
        echo "Branches deleted successfully!"
    else
        echo "Deletion cancelled."
    fi
}

alias gitbp='git_prune_branches'