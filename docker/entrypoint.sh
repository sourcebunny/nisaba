#!/bin/sh

# Set default arguments
LLAMAFILE_ARGS="--nobrowser -ngl 0"

# Initialize model path
MODEL_FILE=""

# Check if the custom arguments file exists
if [ -f "/app/config/llamafile_args.txt" ]; then
  echo "Custom llamafile arguments file found. Reading arguments..."
  LLAMAFILE_ARGS=$(cat /app/config/llamafile_args.txt)
fi

# Check if model.gguf is present in the llamafile directory
if [ -f "/app/model.gguf" ]; then
  echo "Model file found in app directory."
  MODEL_FILE="/app/model.gguf"
# Else, check if model.gguf is present in the config directory
elif [ -f "/app/config/model.gguf" ]; then
  echo "Model file found in config directory."
  MODEL_FILE="/app/config/model.gguf"
fi

# Run llamafile if a model file was found
if [ -n "$MODEL_FILE" ]; then
  echo "Running llamafile with model file $MODEL_FILE and arguments: $LLAMAFILE_ARGS"
  /app/llamafile -m $MODEL_FILE $LLAMAFILE_ARGS &
else
  echo "No model file found. Skipping llamafile execution..."
fi

# Run Nisaba
echo "Running Nisaba..."
/app/nisaba
