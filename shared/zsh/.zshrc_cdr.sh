#!/usr/bin/env zsh

# searches for the Git root of the current directory and navigates into it
cdr() {
    cd "$(git rev-parse --show-toplevel)"
}