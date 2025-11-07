#!/usr/bin/env zsh

# https://morsecodist.io/blog/tmac

tmac () {
    tmux has-session -t "$1" 2>/dev/null
    if [ $? != 0 ]; then
        if [ -f "$HOME/.tmux/$1" ]; then
            echo "Loading templated tmux session from $HOME/.tmux/$1"
            tmux source "$HOME/.tmux/$1"
        else
            echo "Creating new tmux session $1"
            tmux new-session -s "$1" -d
        fi

    else
        echo "Attaching to existing tmux session $1"
        tmux attach -t "$1"
    fi
}

_tmac_complete() {
    local word=${COMP_WORDS[COMP_CWORD]}
    local sessions=$(tmux list-sessions -F "#{session_name}" 2>/dev/null)
    COMPREPLY=( $(compgen -W "$sessions" -- "$word") )
}
complete -F _tmac_complete tmac