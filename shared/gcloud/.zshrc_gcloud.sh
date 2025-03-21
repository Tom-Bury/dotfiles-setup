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


    if [[ -n $selection ]]; then
        newProject=$(echo $selection| grep -oE '[0-9]+$')
        newProjectId=$(gcloud projects describe $newProject --format="value(projectId)")
        newProjectName=$(gcloud projects describe $newProject --format="value(name)")
        
        gcloud auth application-default set-quota-project $newProjectId --quiet
        gcloud config set project $newProjectId --quiet
        
        echo "\n\n✅ gcloud project switch done, now using: $newProjectName"
    else
        echo "\n\n❌ gcloud project not changed"
    fi
}

alias gcp='gcloud'
