#!/usr/bin/env zsh

function ollama_start() {
    pkill ollama
    ollama serve > /dev/null 2>&1 &;
    sleep 1
    ollama run llama3.1:8b ""
    ollama run starcoder2:3b ""
    ollama ps
}