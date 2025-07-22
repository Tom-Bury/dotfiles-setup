#!/usr/bin/env zsh

MODEL_PATH=$1

if [[ ! -f "$MODEL_PATH" ]];
then
    echo "❌ Model file not found: $MODEL_PATH"
    exit 1;
fi

FILE_NAME=$(basename "$MODEL_PATH")

if [[ ! "$FILE_NAME" =~ \.Modelfile$ ]];
then
    echo "❌ Provided model file $FILE_NAME is not a valid .Modelfile, format: <model_name>.Modelfile"
    exit 1;
fi


MODEL_NAME=$(basename "$MODEL_PATH" .Modelfile)
echo "Importing model $MODEL_NAME from $MODEL_PATH to Ollama 🦙"

ollama create $MODEL_NAME -f $MODEL_PATH