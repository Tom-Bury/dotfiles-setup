#!/usr/bin/env zsh

gcpp() {
    projects=$(gcloud projects list)
    currProject=$(gcloud config get-value project)
    currProjectName=$(gcloud projects describe $currProject --format="value(name)")
    
    fzfPrompt="Select gcloud project (current: $currProjectName) > "

    selection=$(echo $projects | fzf --prompt=$fzfPrompt \
                --color='prompt:#af5fff,header:#262626,gutter:-1,pointer:#af5fff' \
                --reverse \
                --pointer="⏺" \
                --query="$1"
    )

    projectId=$(echo $selection| grep -oE '[0-9]+$')

    if [[ -n $selection ]]; then
        gcloud auth application-default set-quota-project $projectId --quiet
        gcloud config set project $projectId --quiet
    fi

    newProject=$(gcloud config get-value project)
    newProjectName=$(gcloud projects describe $newProject --format="value(name)")
    echo "\n\n✅ gcloud project switch done, now using: $newProjectName"
}

alias gcp='gcloud'
