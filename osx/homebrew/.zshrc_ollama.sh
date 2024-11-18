#!/usr/bin/env zsh

function ollama_start() {
    ollama serve > /dev/null 2>&1 &;

    sleep 5

    ollama run llama3.1:8b test
    ollama run starcoder2:3b test
}