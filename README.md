
# SBOT Tool Instructions

`sbot` tool is a command-line utility that integrates with [OpenRouter](https://openrouter.ai) to deliver chat completions to your cli. `sbot` works by utilizing chat templates files stored in designated folders to configure chats with models available on OpenRouter.ai. There are four default templates in the chat_template folder, which you can use as templates to create your own. Below are instructions on how to use the various commands available in `sbot`.
## Table of Contents

- [Set Your API Key](#set-your-api-key)
- [Add an alias](#add-an-alias)
- [Basic Usage](#basic-usage)
- [Commands](#commands)
  - [Shell Query](#basic-query)
  - [Shortcut Query](#shortcut-query)
  - [PIPE](#pipe)
  - [Run Current Command](#run-current-command)
  - [Run Last Command from Local History](#run-last-command-from-local-history)
  - [Show Local History](#show-local-history)
  - [Enable Debug Mode](#enable-debug-mode)
  - [Help](#help)

## Install sbot
```
curl -sSL https://raw.githubusercontent.com/osamsam321/sbot/refs/heads/main/install.sh | sh
```

## Set Your API Key
Create a [OpenRouter key](https://openrouter.ai/settings/keys) and Create the .env file in the root of the project if it doesn't exist.
```
cd ~/.sbot
```

Set your API Key in the .env file and use a editor to add the key
```
vi .env
```

Optional. If you have a OPENROUTER_API_KEY env variable, you can run this command.

```
sed -i "s/OPENROUTER_API_KEY=/OPENROUTER_API_KEY=$OPENROUTER_API_KEY/" .env
```
## Add an alias

Add alias to a .bashrc, zshrc or any alias files e.g.

```
vi .bashrc
alias sbot="~/.sbot/bin/sbot"
source ~/.bashrc
```
## Basic Usage

Once `sbot` is installed, you can run it from the command line. The general syntax for using `sbot` is:

```sh
sbot [options]
```

## Commands

### Basic Query

Specifiy a chat template name to use and add your query. Your query will be added inside of your placeholder and content defined in the chat template.

**Usage**:
```sh
sbot -t <chat template> -q "<your query>"
```

**Example**:
```sh
sbot -t nix -q "find all files in my current directory that are txt or json files"
sbot -t explain-nix -q "ls -ltrah"

```
### Shortcut Query

You can run a query without a specified chat template name. Sbot will automatically use a chat template specified in the settings.json file. The nix prompt has the lowest by default

**Example**:
```sh
sbot -q "find all files in my current directory that are txt or json files"

```
### PIPE

You can also pipe content.

**Example**:
```sh
echo "how are you doing today?" | sbot -t general
```
You can still use the query option as a add on.

**Example**:
```sh
echo "list files" | sbot -t nix -q " that have the word cat in the filename."
```
### Run Current Command

Run the last command that exists in the local sbot history file.

**Usage**:
```sh
sbot -x
```

**Example**:
```sh
sbot -q "list files in sorted order" -t nix -x
```


### Run Last Command from Local History

Run the last command that exists in the local sbot history file.

**Usage**:
```sh
sbot -l
```

### Show Local History

Show the local history of commands executed with `sbot`.

**Usage**:
```sh
sbot -y
```

### Enable Debug Mode

Enable debug mode to get more detailed output for troubleshooting.

**Usage**:
```sh
sbot -d
```
### Additionals settings such shell type, a list of commands to block, or default chat templates can be specified in the settings.json file

Enable debug mode to get more detailed output for troubleshooting.

**Usage**:
```sh
vi ~/.sbot/settings.json
```
### Help

Show the help message for `sbot`.

**Usage**:
```sh
sbot -h
```

## Conclusion

This document provides a basic overview of how to use the `sbot` tool. For more detailed information on each command and its options, use the help command:

```sh
sbot -h
```
