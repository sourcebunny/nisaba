# nisaba

<img src="images/preview.png" width="800" />

Nisaba is an IRC bot designed to interact with users in a chat channel, using PrivateGPT for generating responses.

## Features

- Responds to messages directed at it by consulting PrivateGPT.
- Ignores messages from users listed in a blocklist.
- Splits long messages to adhere to IRC's message length limits.

## Requirements

- Go (Version 1.16 or later recommended)
- github.com/thoj/go-ircevent
- Fully configured PrivateGPT endpoint

## Setup

1. **Install Go**
- If you haven't already, follow the instructions on the official [Go website](https://golang.org/dl/).

2. **Install Dependencies**
- Get the IRC event package:

```
go get github.com/thoj/go-ircevent
```

3. **Configure the Bot**
- Create a `config.json` file with the bot's configuration. Example:

  ```json
  {
    "nickname": "Nisaba",
    "server": "irc.example.com",
    "port": "6667",
    "use_ssl": false,
    "validate_ssl": false,
    "api_url": "http://127.0.0.1/v1/completions",
    "api_key": "optional_key_here",
    "channel": "#example",
    "max_message_size": 400
  }
  ```
- Optionally, create a `blocklist.txt` file with usernames to ignore, one per line.

4. **Build the Bot**:
- Navigate to the project directory and run:

```
go build -o nisaba.o .
```

5. **Run the Bot**:
- When you run the binary, the bot will connect to the IRC server specified in the configuration.

```
./nisaba.o
```

## Usage

Simply interact with the bot in your configured IRC channel!

To get a response, prefix your message with the bot's name, e.g., "Nisaba, how are you?".

## Disclaimer
<details><summary>Software Disclaimer</summary>
  
The code in this repository is provided "as-is" without any warranty of any kind, either expressed or implied. It is intended for research and educational purposes only. The authors and contributors make no representations about the suitability of this software for any purpose and accept no liability for any consequences resulting directly or indirectly from the use of this software.


By using this software, you acknowledge and agree to assume all risks associated with its use, understanding that you are solely responsible for any damage to your computer system or loss of data that results from such activities. You also acknowledge that this software is not intended for use in production environments or for commercial purposes.
</details>
