#!/bin/sh

# Run llamafile if model.gguf is present
if [ -f "/data/model.gguf" ]; then
  echo "Model file found. Running llamafile..."
  /app/llamafile -m /data/model.gguf -ngl 0 &
else
  echo "No model file found. Skipping llamafile execution..."
fi

# Run Nisaba
echo "Running Nisaba..."
/app/nisaba
