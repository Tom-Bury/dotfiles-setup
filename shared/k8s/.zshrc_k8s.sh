#!/usr/bin/env zsh

k8sc() {
    contexts=$(kubectl config get-contexts --no-headers -o name)
    
    selection=$(echo $contexts | fzf --prompt='Select K8s context > ' \
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

k8sp() {
    namespaces=$(kubectl -o name get namespaces)

    selection=$(echo $namespaces | fzf --prompt='Select K8s namespace > ' \
                --color='prompt:#af5fff,header:#262626,gutter:-1,pointer:#af5fff' \
                --reverse \
                --pointer="⏺" \
                --query="$1"
    )

    namespace="${selection##*/}"

    if [[ -n $namespace ]]; then
        kubectl get pods --namespace $namespace
    else
        echo "No namespace found in selection '$selection'"
    fi
}

alias k8s='kubectl'
