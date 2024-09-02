
# SBOT Tool Instructions

The `sbot` tool is a command-line utility designed to work with ChatGPT. Below are the instructions on how to use the various commands available in `sbot`.

## Table of Contents
- [Basic Usage](#basic-usage)
- [Commands](#commands)
  - [Ask a Basic Unix Shell Query](#ask-a-basic-unix-shell-query)
  - [Run Last Command from Local History](#run-last-command-from-local-history)
  - [Show Local History](#show-local-history)
  - [Enable Debug Mode](#enable-debug-mode)
  - [Explain a Command](#explain-a-command)
  - [Ask a General GPT Question](#ask-a-general-gpt-question)
  - [Filter or Combine Query with Stdin](#filter-or-combine-query-with-stdin)
  - [Help](#help)
 
  
## Set Your API Token
Set your API token in the .env file 
```
mv .env.example .env && vim .env
```

If you have a *nix env variable

mv .env.example .env && sed -i "s/openai_token=/openai_token=$OPENAI_API_KEY/" .env

## Basic Usage

Once `sbot` is installed, you can run it from the command line. The general syntax for using `sbot` is:

```sh
sbot [options]
```

## Commands


### Ask a Basic Unix Shell Query

Ask a basic Unix shell query and get a command back.

**Usage**:
```sh
sbot -q "<your query>"
```

**Example**:
```sh
sbot -q "find all files in my current directory that are txt or json files"


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
### Explain a Command

Explain what a specific command does.

**Usage**:
```sh
sbot -e <command>
```

**Example**:
```sh
sbot -e "ls -l"
```

### Ask a General GPT Question

Ask a general question and get an answer from GPT.

**Usage**:
```sh
sbot -g "<your question>"
```

**Example**:
```sh
sbot -g "What is the capital of France?"
```

### Filter or Combine Query with Stdin

Filter or combine a query with stdin input.

**Usage**:
```sh
sbot -i "<your query>"
```

**Example**:
```sh
echo "what is a popular alternative to pet cat?" | sbot -i "what is the history of this animal?"
```



### Enable Debug Mode

Enable debug mode to get more detailed output for troubleshooting.

**Usage**:
```sh
sbot -d
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

Feel free to explore the various features and options available in `sbot` to make the most out of this tool for managing and deploying your projects.
