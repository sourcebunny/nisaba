#!/bin/bash

# Function to download files using curl
download_file() {
    echo "Downloading $1..."
    curl -L -o "$2" "$1"
}

# Step 1: Download llamafile if required
echo
echo "Nisaba Setup Script"
echo "https://github.com/sourcebunny/nisaba"
echo
echo "Do you want to download the llamafile binary required for the API endpoint? (y/n)"
read -r download_llama
case "$download_llama" in
    [Yy]|[Yy][Ee][Ss])
        download_file "https://github.com/Mozilla-Ocho/llamafile/releases/download/0.8.1/llamafile-0.8.1" "llamafile"
        ;;
    [Nn]|[Nn][Oo])
        echo
        echo "Skipping llamafile download."
        ;;
    *)
        echo
        echo "Invalid input, skipping llamafile download."
        ;;
esac

# Step 2: Configure required settings
echo
echo "Do you want to configure required settings? (y/n)"
read -r config_irc
if [[ "$config_irc" =~ ^[Yy]|[Yy][Ee][Ss]$ ]]; then
    mkdir -p config
    echo
    echo "Creating custom config/config.json"
    echo
    echo "For each option below, press 'Enter' to use the default value."
    read -p "Nickname (Nisaba): " nickname
    read -p "Server (irc.example.com): " server
    read -p "Port (6667): " port
    read -p "Use SSL (false): " use_ssl
    read -p "Validate SSL (false): " validate_ssl
    read -p "Allow Commands (true): " commands
    read -p "API URL (http://localhost:8080/v1/chat/completions): " api_url
    read -p "API Key (api_key_here): " api_key
    read -p "API Mode (chat): " api_mode
    read -p "Channel (#example): " channel
    read -p "Max Message Size (400): " message_size
    read -p "Delay Between Messages in Seconds (3): " delay

    cat << EOF > config/config.json
{
    "nickname": "${nickname:-Nisaba}",
    "server": "${server:-irc.example.com}",
    "port": "${port:-6667}",
    "use_ssl": ${use_ssl:-false},
    "validate_ssl": ${validate_ssl:-false},
    "commands": ${commands:-true},
    "api_url": "${api_url:-http://localhost:8080/v1/chat/completions}",
    "api_key": "${api_key:-api_key_here}",
    "api_mode": "${api_mode:-chat}",
    "channel": "${channel:-#example}",
    "message_size": ${message_size:-400},
    "delay": ${delay:-3}
}
EOF
else
    echo
    echo "Skipping step. The config.json file is required for Nisaba."
    echo
fi

# Step 3: API endpoint options presets
echo
echo "Do you want to choose a default options preset for the API endpoint? (y/n)"
read -r preset_choice
case "$preset_choice" in
    [Yy]|[Yy][Ee][Ss])
        echo
        echo "Choose preset: 1 for LLaMA Precise or 2 for Divine Intellect"
        read -r preset_option
        case "$preset_option" in
            1)
                cp config/options.precise.json.example config/options.json
                cp config/options.precise.json.example config/options.precise.json
                cp config/options.divine.json.example config/options.divine.json
                ;;
            2)
                cp config/options.divine.json.example config/options.json
                cp config/options.precise.json.example config/options.precise.json
                cp config/options.divine.json.example config/options.divine.json
                ;;
            *)
                echo "Invalid option, skipping step."
                ;;
        esac
        ;;
    [Nn]|[Nn][Oo])
        echo
        echo "Skipping step."
        ;;
    *)
        echo
        echo "Invalid input, skipping step."
        ;;
esac

# Step 4: Make binaries executable
echo
echo "Do you want to make nisaba.bin and llamafile executable? (y/n)"
read -r exec_bin
if [[ "$exec_bin" =~ ^[Yy]|[Yy][Ee][Ss]$ ]]; then
    chmod +x llamafile
    chmod +x nisaba*.bin
else
    echo
    echo "Skipping step. Both files will need to be made executable in order to run."
fi

# Step 5: Download model
echo
echo "Do you want to download an LLM model and save it as model.gguf? (y/n)"
read -r download_model
if [[ "$download_model" =~ ^[Yy]|[Yy][Ee][Ss]$ ]]; then
    echo
    echo "Select the LLM model to download:"
    echo
    echo "1. 1.1B (tiny) - 669 MB - TinyLlama-1.1B-Chat-v1.0-GGUF"
    echo "- Created by TinyLLaMA, GGUF Conversion by TheBloke"
    echo
    echo "2. 7B (small) - 4.37 GB - Mistral-7B-Instruct-v0.2-GGUF"
    echo "- Created by MistralAI, GGUF Conversion by TheBloke"
    echo
    echo "3. 10.7B (medium) - 6.46 GB - SOLAR-10.7B-Instruct-v1.0-GGUF"
    echo "- Created by Upstage, GGUF Conversion by TheBloke"
    echo
    echo "4. 34B (large) - 20.7 GB - Nous-Capybara-34B-GGUF"
    echo "- Created by NousResearch, GGUF Conversion by TheBloke"
    read -r model_size
    case "$model_size" in
        1)
            download_file "https://huggingface.co/TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF/resolve/main/tinyllama-1.1b-chat-v1.0.Q4_K_M.gguf" "model.gguf"
            ;;
        2)
            download_file "https://huggingface.co/TheBloke/Mistral-7B-Instruct-v0.2-GGUF/resolve/main/mistral-7b-instruct-v0.2.Q4_K_M.gguf" "model.gguf"
            ;;
        3)
            download_file "https://huggingface.co/TheBloke/SOLAR-10.7B-Instruct-v1.0-GGUF/resolve/main/solar-10.7b-instruct-v1.0.Q4_K_M.gguf" "model.gguf"
            ;;
        4)
            download_file "https://huggingface.co/TheBloke/Nous-Capybara-34B-GGUF/resolve/main/nous-capybara-34b.Q4_K_M.gguf" "model.gguf"
            ;;
        *)
            echo "Invalid model selection, skipping step."
            ;;
    esac
else
    echo
    echo "Skipping step."
fi

# Step 6. Provide instructions
echo
echo "Nisaba setup is complete!"
echo
echo "Usage: Run './llamafile' and then './nisaba-OS-ARCH.bin' using your shell."
echo
echo "Ensure that you use the proper llamafile args."
echo
echo "Example llamafile args for using only the CPU:"
echo "./llamafile -m model.gguf -ngl 0"
echo
echo "Example llamafile args for offloading layers to the GPU:"
echo "./llamafile -m model.gguf -ngl 999"
echo
