#!/usr/bin/env zsh

setup_ssh() {
    for key in ~/.ssh/id_*; do
        if [[ ! $key =~ \.pub$ ]]; then
            if [[ "$OSTYPE" == "darwin"* ]]; then
                ssh-add --apple-use-keychain $key
            else
                ssh-add $key
            fi
        fi
    done
}