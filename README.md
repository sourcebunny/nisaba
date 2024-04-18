# Nisaba

<img src="images/preview.png" width="800" />

Nisaba is an IRC bot written in Go, designed to interact with users in a chat channel, using llamafile for generating responses.

## Background

[Nisaba](https://en.wikipedia.org/wiki/Nisaba) is named after the Mesopotamian goddess of writing and grain.

This project began as a way to learn [Go](https://go.dev/learn/), aimed at creating a frontend for interacting with local OpenAI or similar endpoints.

Initially, the project used [PrivateGPT](https://github.com/zylon-ai/private-gpt) as its backend for generating responses.

As the project evolved, the need for more flexible API options led to a transition to [llamafile](https://github.com/Mozilla-Ocho/llamafile).

This switch was motivated by llamafile's [ease of use](https://justine.lol/oneliners/) and its API endpoint being [llama.cpp](https://github.com/ggerganov/llama.cpp) compatible.

## Features

- Responds to messages directed at it by consulting llamafile for generating responses.
- Configurable customization options for the bot, such as setting a custom bot name.
- Supports dynamic loading of different API options for response generation.
- Ignores messages from users listed in a block list.
- Splits long messages to adhere to IRC's message length limits.
- Allows commands through IRC, such as clearing message history or loading new options.

## To Do

<details>
<summary><strong>Planned Features</strong></summary>

These are some features that are currently planned for Nisaba.

- Default responses sent to users will be able to be customized.
  - For example: `I will think about that and be back with you shortly.`
  - Some default responses will be able to be toggled.
- Conversations will be able to be saved and loaded at a later time.
  - This will likely be in the form of `!save` and `!load` commands.
- Support for other chat platforms, such as Discord, will eventually be added.
  - For now, focus is being put on making things work properly within IRC.

</details>

<details>
<summary><strong>Known Issues</strong></summary>

These are issues, or shortcomings, that will be addressed in future releases.

Below each issue is a proposed solution that is currently being considered, or worked on, to address each issue.

- This project is in its early stages, some stability and performance issues are expected.
    - If you come across any issues, feel free to report them through GitHub issues.

</details>

## Requirements

<details>
<summary><strong>General</strong> (Automated Setup, Docker and Building)</summary>

These requirements apply to all setup methods.

- Linux, Mac, or Windows computer capable of running an LLM model for the AI backend.
- Fully configured llamafile API endpoint.
    - This is automatically downloaded and configured by the setup script.

</details>

<details>
<summary><strong>Docker</strong></summary>

The optional Docker container can be built to include all requirements.

- [Install Docker](https://docs.docker.com/engine/install/)

</details>

<details>
<summary><strong>Buliding</strong></summary>

To build the standalone Go binary, you will need the build requirements.

- [Install Go](https://go.dev/doc/install)
- Go Dependencies
    - [github.com/thoj/go-ircevent](https://github.com/thoj/go-ircevent)

</details>

## Setup

Nisaba can be run either as a standalone application or within a Docker container.

Each method requires a configured `config.json` file, and optionally `options.json`, located in the `./config` directory.
- These files can be created automatically by the `setup.sh` or `setup.bat` script, explained in the Automated Setup instructions.

Choose one of the setup methods below and follow the directions to configure Nisaba.

<details>
<summary><strong>Automated (Pre-Built) Setup</strong> - Simple setup using prepared scripts and binaries for Windows/Linux/Mac.</summary>

Follow these detailed steps to get Nisaba running quickly using the pre-built scripts included with the releases:

1. **Download the Pre-Built Binary Archive**
   - Visit the [Releases page](https://github.com/sourcebunny/nisaba/releases) on GitHub.
   - Download the appropriate archive for your operating system:
     - `nisaba-linux.tar.gz` for Linux
     - `nisaba-mac.tar.gz` for Mac
     - `nisaba-windows.zip` for Windows

2. **Prepare the Setup Script**
   - **For Linux or Mac**:
     - Extract the contents of the `.tar.gz` archive.
     - Open a terminal and navigate to the extracted directory.
     - Make the setup script executable:
       ```bash
       chmod +x setup.sh
       ```
   - **For Windows**:
     - Extract the contents of the `.zip` archive.
     - Open Command Prompt and navigate to the extracted directory.

3. **Run the Setup Script**
   - **For Linux or Mac**:
     - In your terminal, execute the script by running:
       ```bash
       ./setup.sh
       ```
   - **For Windows**:
     - In Command Prompt, execute the script by running:
       ```cmd
       setup.bat
       ```
   - Follow the on-screen prompts to configure your setup. The script will guide you through several steps:
        - **Download llamafile Binary**: The script will ask if you want to download the llamafile binary required for the API endpoint. Answer `y` for yes.
        - **Configure Requried Settings**: You will be prompted to configure required settings to create a config.json file. Answer `y` to proceed.
        - **Enter Configuration Details**: The script will then prompt you to enter various configuration details such as nickname, server, port, etc. Press 'Enter' to accept default values or enter your custom settings.
        - **Choose API Endpoint Options**: You'll have the option to select a default options preset for the API endpoint. Answer `y` and choose between provided presets like "LLaMA Precise" or "Divine Intellect".
        - **Make the Binaries Executable**: You will be prompted to make the binaries for Nisaba and llamafile executable. Answer `y` to proceed.
        - **Model Download**: Finally, the script will ask if you want to download a model and save it as `model.gguf`. Answer `y` and select the LLM model to download.

4. **Run Nisaba and Llamafile**
   - After configuration, start the services:
     - **For Linux**:
       - Run the llamafile binary first to start the endpoint:
         ```bash
         ./llamafile -m model.gguf -ngl 0
         ```
       - Then run the Nisaba binary:
         ```bash
         ./nisaba-linux-amd64.bin
         ```
     - **For Mac**:
       - Run the llamafile binary first to start the endpoint:
         ```bash
         ./llamafile -m model.gguf -ngl 0
         ```
       - Then run the Nisaba binary:
         ```bash
         ./nisaba-mac-amd64.bin
         ```
     - **For Windows**:
       - Run the llamafile binary first to start the endpoint:
         ```cmd
         .\llamafile.exe -m model.gguf -ngl 0
         ```
       - Then run the Nisaba binary:
         ```cmd
         .\nisaba-windows-amd64.exe
         ```

</details>

<details>
<summary><strong>Building Instructions and Setup</strong> - Instructions for manually building and running Nisaba from source.</summary>

1. **Install Go**
   - If you haven't already, follow the instructions on the official [Go website](https://golang.org/dl/).

2. **Install Dependencies**
   - Install the IRC event package:
     ```
     go get github.com/thoj/go-ircevent
     ```

3. **Configure the Bot**
   - Manually create a `config` directory in your project root and place your `config.json` file within this directory. Optionally, add an `options.json` for API parameters.
       - Use the `setup.sh` or `setup.bat` script to generate these files automatically.
   - Example `config.json` and `options.json` files are provided under `config/` for reference including popular API presets:
     - `config.json.example` to reference required settings file
     - `options.precise.json.example` for "LLaMA Precise"
     - `options.divine.json.example` for "Divine Intellect"
     - `options.json.example` to reference all available options
   - Rename the relevant example file to `options.json` if you wish to use it.

4. **Build the Bot**:
   - Navigate to the project directory and run:
     ```
     go build -o nisaba.bin .
     ```

5. **Run the Bot**:
   - Ensure that you have a llamafile API endpoint running.
   - Start the bot by running the binary:
     ```
     ./nisaba.bin
     ```

</details>

<details>
<summary><strong>Docker Setup</strong> - Guide for deploying Nisaba with Docker, including llamafile.</summary>

1. **Prepare Configurations**
   - Place `config.json`, `options.json` (if used), and `model.gguf` in a directory named `config` in the same directory as your `docker-compose.yml`.
   - Example `config.json` and `options.json` files are provided under `config/` for reference including popular API presets:
     - `config.json.example` to reference required settings file
     - `options.precise.json.example` for "LLaMA Precise"
     - `options.divine.json.example` for "Divine Intellect"
     - `options.json.example` to reference all available options

2. **Build and Run with Docker Compose**
   - Ensure the Docker Compose file is set to mount the `config` directory correctly:
     ```yaml
     version: '3.8'
     services:
       nisaba:
         build: .
         volumes:
           - ./model.gguf:/app/model.gguf
           - ./config:/app/config
     ```
   - Run the following command in the directory containing `docker-compose.yml`:
     ```
     docker-compose up --build
     ```

</details>

## Configuration

These configuration files can be placed in the `config/` directory, or the same directory as the Nisaba binary.

<details>
<summary><strong>Configuration Files</strong> - Overview of various configuration files used by Nisaba.</summary>

- **config.json**: Required main configuration for the IRC bot, specifying connection details and API settings.
- **options.json**: Optional parameters file designed to adjust llamafile's behavior, with settings like `temperature`, `top_k`, etc.
- **systemprompt.txt**: System prompt for Nisaba sent to the llamafile endpoint.
- **blocklist.txt**: Blocks specific IRC nicknames from interacting with Nisaba.
- **history.txt**: Stores message context dynamically; should not be edited manually.
- **llamafile_args.txt** (Docker only): Custom arguments to replace default llamafile settings under Docker.

</details>

## Usage

<details>
<summary><strong>Basic Interaction</strong> - How to interact with Nisaba.</summary>

To get a response from Nisaba, simply prefix your message with the bot's name, followed by your query.

For example: `Nisaba, how are you?`

After you send a message or command, Nisaba will use the API endpoint to generate a response, and then send that response back to you in the designated IRC channel.

</details>

<details>
<summary><strong>Using Commands</strong> - Commands available for controlling Nisaba.</summary>

Nisaba supports several commands that can be used to control the bot or modify its behavior dynamically.

These commands should be prefixed with the bot's name, optionally followed by a comma or colon, and the command:

- **!clear**: Clears the message history stored by the bot. Useful for resetting the context in "chat" mode.
  - `Nisaba, !clear`
- **!options [filename]**: Loads specific option settings from a file named `options.[filename].json` if present in the `config` directory. This allows you to dynamically change how Nisaba interacts with the llamafile API without restarting the bot.
  - `Nisaba, !options precise`
- **!system [message]**: Attaches a system prompt to the next message that Nisaba sends to the llamafile endpoint, affecting how responses are generated.
  - `Nisaba, !system You will respond using 100 words or less.`

</details>

## Credits

Special thanks go to the following projects and their contributors.

- **[Mozilla-Ocho/llamafile](https://github.com/Mozilla-Ocho/llamafile)**: 
  - A powerful LLM backend used by Nisaba under Docker to host the API endpoint.

- **[thoj/go-ircevent](https://github.com/thoj/go-ircevent)**:
  - An IRC event handling library in Go that enables Nisaba to connect and interact with IRC servers.

## Legal

<details><summary><strong>Third-Party Licenses</strong></summary>

This software includes components that are licensed under various open source licenses.

### go-ircevent

The `go-ircevent` library is used under the BSD 3-Clause License.

```
// Copyright (c) 2009 Thomas Jager. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
```

</details>

<details><summary><strong>Software Disclaimer</strong></summary>

The code in this repository is provided "as-is" without any warranty of any kind, either expressed or implied. It is intended for research and educational purposes only. The authors and contributors make no representations about the suitability of this software for any purpose and accept no liability for any consequences resulting directly or indirectly from the use of this software.


By using this software, you acknowledge and agree to assume all risks associated with its use, understanding that you are solely responsible for any damage to your computer system or loss of data that results from such activities. You also acknowledge that this software is not intended for use in production environments or for commercial purposes.
</details>
