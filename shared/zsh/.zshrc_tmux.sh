#!/usr/bin/env zsh

# https://morsecodist.io/blog/tmac

tmac () {
    tmux has-session -t "$1" 2>/dev/null
    if [ $? != 0 ]; then
        tmux new-session -s "$1" -d
    fi
    tmux attach -t "$1"
}

_tmac_complete() {
    local word=${COMP_WORDS[COMP_CWORD]}
    local sessions=$(tmux list-sessions -F "#{session_name}" 2>/dev/null)
    COMPREPLY=( $(compgen -W "$sessions" -- "$word") )
}
complete -F _tmac_complete tmac