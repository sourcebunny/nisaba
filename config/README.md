# Configuration

This directory contains example configuration files for Nisaba.

Below is an explanation of each file and the parameters they accept.

## `config.json`

This is the primary configuration file required to specify IRC connection details and API settings.

All options except for `server` and `channel` are optional in this file.

### Parameters
- **server** (string) (required): The IRC server Nisaba connects to. e.g., `"irc.example.com"`.
- **channel** (string) (required): The IRC channel Nisaba will join and operate within, e.g., `"#example"`.
- **port** (string): IRC server port, default is `"6667"`.
- **use_ssl** (boolean): Enables SSL connection to the IRC server, default is `false`.
- **validate_ssl** (boolean): Enables SSL certificate validation, default is `false`.
- **commands** (boolean): Flag to enable or disable command handling, default is `true`.
- **debug** (boolean): Flag to enable or disable debug output, default is `false`.
- **api_url** (string): URL of the API endpoint, default is `"http://localhost:8080/v1/chat/completions"`.
- **api_key** (string): Authentication key for the API if required, default is `"null"`.
- **api_mode** (string): Determines if the bot uses "chat" or "query" mode, default is `"chat"`.
  - The "chat" mode is intended to be used with the `/v1/chat/completions` API endpoint.
  - The "query" mode is intended to be used with the `/completion` API endpoint.
- **nickname** (string): The nickname that Nisaba will use in IRC, default is `"Nisaba"`.
- **message_size** (int): Maximum characters in each message sent by the bot, default is `400`.
- **delay** (int): Set the delay between messages in seconds, default is `3`.

## `options.json`

Optional parameters file designed to adjust llamafile's behavior in the request to its API.

### Parameters
- **temperature** (float): Default `0.8`
- **top_k** (integer): Default `40`
- **top_p** (float): Default `0.95`
- **min_p** (float): Default `0.05`
- **n_predict** (integer): Default `-1`
- **n_keep** (integer): Default `0`
- **tfs_z** (float): Default `1.0`
- **typical_p** (float): Default `1.0`
- **repeat_penalty** (float): Default `1.1`
- **repeat_last_n** (integer): Default `64`
- **presence_penalty** (float): Default `0.0`
- **frequency_penalty** (float): Default `0.0`
- **mirostat** (integer): Default `0`
- **mirostat_tau** (float): Default `5.0`
- **mirostat_eta** (float): Default `0.1`
- **seed** (integer): Default `-1`
- **n_probs** (integer): Default `0`
- **slot_id** (integer): Default `-1`
- **penalize_nl** (boolean): Default `true`
- **ignore_eos** (boolean): Default `false`
- **cache_prompt** (boolean): Default `false`

## `systemprompt.txt`

Contains the system prompt for Nisaba initially sent to the llamafile endpoint with the first message in "chat" mode.

The initial system message sets the conversational tone or instructions for the conversation with the assistant.

## `reminderprompt.txt`

Contains the system prompt for Nisaba sent to the llamafile endpoint with every new message in "chat" mode.

This system message reinforces the ongoing conversational tone or instructions.

## `blocklist.txt`

Blocks specific IRC nicknames from interacting with Nisaba. Add each username on a new line.

## `history.txt`

Stores message context dynamically to maintain conversation state across interactions.

It should not be edited manually as it may be modified during runtime.

## `llamafile_args.txt` (Docker only)

This file contains custom arguments to replace default llamafile settings when running under Docker.

It's useful for deploying Nisaba with specific performance configurations.
