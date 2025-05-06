#!/usr/bin/env zsh

IGNORE_DIRS=(
    node_modules
    .git
    build
    dist
    .bundle
    Pods
    .gradle
    .jj
)

fd() {
    local dir
    local ignore_expr
    for ignore_dir in "${IGNORE_DIRS[@]}"; do
        ignore_expr="${ignore_expr} -not -path '*/${ignore_dir}/*'"
    done
    
    dir=$(eval find . -maxdepth 3 -type d $ignore_expr | fzf --prompt='CD into > ' \
                --color='prompt:#af5fff,header:#262626,gutter:-1,pointer:#af5fff' \
                --reverse \
                --pointer="⏺" \
                --query="$1"
        )
    if [[ -n $dir ]]; then
        cd "$dir"
    fi
}