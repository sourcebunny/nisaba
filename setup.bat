@echo off
setlocal enabledelayedexpansion

echo Nisaba Setup Script
echo https://github.com/sourcebunny/nisaba
echo.

:: Step 1: Download llamafile if required
echo Do you want to download the llamafile binary required for the API endpoint? (y/n)
set /p download_llama=
if /I "!download_llama!"=="y" (
    echo.
    echo Downloading llamafile...
    powershell -Command "(New-Object Net.WebClient).DownloadFile('https://github.com/Mozilla-Ocho/llamafile/releases/download/0.7.1/llamafile-0.7.1', 'llamafile')"
    rename llamafile llamafile.exe
) else if /I "!download_llama!"=="n" (
    echo.
    echo Skipping llamafile download.
) else (
    echo.
    echo Invalid input, skipping llamafile download.
)
echo.

:: Step 2: Configure required settings
echo Do you want to configure required settings? (y/n)
set /p config_irc=
if /I "!config_irc!"=="y" (
    if not exist config mkdir config
    echo.
    echo Creating custom config/config.json
    echo.
    echo For each option below, press 'Enter' to use the default value.

    set /p nickname="Nickname (Nisaba): "
    if "!nickname!"=="" set nickname=Nisaba

    set /p server="Server (irc.example.com): "
    if "!server!"=="" set server=irc.example.com

    set /p port="Port (6667): "
    if "!port!"=="" set port=6667

    set /p use_ssl="Use SSL (true): "
    if "!use_ssl!"=="" set use_ssl=true

    set /p validate_ssl="Validate SSL (true): "
    if "!validate_ssl!"=="" set validate_ssl=true

    set /p commands="Commands (true): "
    if "!commands!"=="" set commands=true

    set /p api_url="API URL (http://localhost:8080/v1/chat/completions): "
    if "!api_url!"=="" set api_url=http://localhost:8080/v1/chat/completions

    set /p api_key="API Key (api_key_here): "
    if "!api_key!"=="" set api_key=api_key_here

    set /p api_mode="API Mode (chat): "
    if "!api_mode!"=="" set api_mode=chat

    set /p channel="Channel (#example): "
    if "!channel!"=="" set channel=#example

    set /p message_size="Max Message Size (400): "
    if "!message_size!"=="" set message_size=400

    (
        echo {
        echo     "nickname": "!nickname!",
        echo     "server": "!server!",
        echo     "port": "!port!",
        echo     "use_ssl": !use_ssl!,
        echo     "validate_ssl": !validate_ssl!,
        echo     "commands": !commands!,
        echo     "api_url": "!api_url!",
        echo     "api_key": "!api_key!",
        echo     "api_mode": "!api_mode!",
        echo     "channel": "!channel!",
        echo     "message_size": !message_size!
        echo }
    ) > config/config.json
) else (
    echo.
    echo Skipping step. The config.json file is required for Nisaba.
    echo.
)
echo.

:: Step 3: API endpoint options presets
echo Do you want to choose a default options preset for the API endpoint? (y/n)
set /p preset_choice=
if /I "!preset_choice!"=="y" (
    echo.
    echo Choose preset: 1 for LLaMA Precise or 2 for Divine Intellect
    set /p preset_option=
    if "!preset_option!"=="1" (
        copy "config\options.precise.json.example" "config\options.json"
        copy "config\options.precise.json.example" "config\options.precise.json"
        copy "config\options.divine.json.example" "config\options.divine.json"
    ) else if "!preset_option!"=="2" (
        copy "config\options.divine.json.example" "config\options.json"
        copy "config\options.precise.json.example" "config\options.precise.json"
        copy "config\options.divine.json.example" "config\options.divine.json"
    ) else (
        echo Invalid option.
    )
) else (
    echo.
    echo Skipping step.
)
echo.

:: Step 4: Download model
echo Do you want to download an LLM model and save it as model.gguf? (y/n)
set /p download_model=
if /I "!download_model!"=="y" (
    echo.
    echo Select the LLM model to download:
    echo.
    echo 1. 1.1B ^(tiny^) - 669 MB - TinyLlama-1.1B-Chat-v1.0-GGUF
    echo - Created by TinyLLaMA, GGUF Conversion by TheBloke
    echo.
    echo 2. 7B ^(small^) - 4.37 GB - Mistral-7B-Instruct-v0.2-GGUF
    echo - Created by MistralAI, GGUF Conversion by TheBloke
    echo.
    echo 3. 10.7B ^(medium^) - 6.46 GB - SOLAR-10.7B-Instruct-v1.0-GGUF
    echo - Created by Upstage, GGUF Conversion by TheBloke
    echo.
    echo 4. 34B ^(large^) - 20.7 GB - Nous-Capybara-34B-GGUF
    echo - Created by NousResearch, GGUF Conversion by TheBloke
    set /p model_size=
    if "!model_size!"=="1" (
        echo.
        echo Downloading model...
        powershell -Command "(New-Object Net.WebClient).DownloadFile('https://huggingface.co/TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF/resolve/main/tinyllama-1.1b-chat-v1.0.Q4_K_M.gguf', 'model.gguf')"
    ) else if "!model_size!"=="2" (
        echo.
        echo Downloading model...
        powershell -Command "(New-Object Net.WebClient).DownloadFile('https://huggingface.co/TheBloke/Mistral-7B-Instruct-v0.2-GGUF/resolve/main/mistral-7b-instruct-v0.2.Q4_K_M.gguf', 'model.gguf')"
    ) else if "!model_size!"=="3" (
        echo.
        echo Downloading model...
        powershell -Command "(New-Object Net.WebClient).DownloadFile('https://huggingface.co/TheBloke/SOLAR-10.7B-Instruct-v1.0-GGUF/resolve/main/solar-10.7b-instruct-v1.0.Q4_K_M.gguf', 'model.gguf')"
    ) else if "!model_size!"=="4" (
        echo.
        echo Downloading model...
        powershell -Command "(New-Object Net.WebClient).DownloadFile('https://huggingface.co/TheBloke/Nous-Capybara-34B-GGUF/resolve/main/nous-capybara-34b.Q4_K_M.gguf', 'model.gguf')"
    ) else (
        echo Invalid model selection, skipping step.
    )
) else (
    echo.
    echo Skipping step.
)
echo.

:: Step 5. Provide instructions
echo Nisaba setup is complete!
echo.
echo Usage: Run 'llamafile.exe' and then 'nisaba-windows-ARCH.exe' using your command prompt.
echo.
echo Ensure that you use the proper llamafile args.
echo.
echo Example llamafile args for using only the CPU:
echo llamafile.exe -m model.gguf -ngl 0
echo.
echo Example llamafile args for offloading layers to the GPU:
echo llamafile.exe -m model.gguf -ngl 999
echo.

endlocal
