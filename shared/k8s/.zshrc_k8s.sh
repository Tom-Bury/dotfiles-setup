#!/usr/bin/env zsh

k8sc() {
    contexts=$(kubectl config get-contexts --no-headers -o name)
    
    selection=$(echo $contexts | fzf --prompt='Select k8 context > ' \
                --color='prompt:#af5fff,header:#262626,gutter:-1,pointer:#af5fff' \
                --reverse \
                --pointer="⏺" \
                --query="$1"
    )

    if [[ -n $selection ]]; then
        kubectl config use-context $selection
    fi

    echo "kubectl config current-context: $(kubectl config current-context)"
}