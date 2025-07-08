# Natural Language Command Helper
`nlch` is a terminal program designed to help you be more productive. It uses context clues from the location it is invoked—such as the current directory, git information, the list of files and folders present—and uses it to generate a terminal command based on natural language input.

---

## Getting Started

### Prerequisites

- Go 1.20 or newer (https://golang.org/dl/)

### Build

```sh
go mod tidy         # Install dependencies
go build -o nlch    # Build the binary
```

### Run

```sh
./nlch "Describe your command here"
```
Or, for development:
```sh
go run main.go "Describe your command here"
```

### CLI Flags

- `--dry-run` — Show the command but do not execute it
- `--model` — Override the model to use
- `--provider` — Override the provider to use
- `--yes-im-sure` — Bypass confirmation for all commands, including dangerous ones
- `--verbose` — Show provider and model information before generating the command
- `--version` — Show version and exit

### Configuration

See below for configuration file details.

---


# Example usage

1. Default Usage:
```
$ nlch "List the files in decreasing order of size"
> Running command `ls -lS`...
> Confirm? [Y/n]: Y
total 8
-rw-r--r-- 1 user user 79 Jul  8 10:56 main.go
-rw-r--r-- 1 user user 49 Jul  8 10:43 go.mod
-rw-r--r-- 1 user user  0 Jul  8 10:56 README.md
```

2. Verbose Usage:
```
$ nlch --verbose "find all go files in this directory"
Provider: openrouter
Model: openai/gpt-4.1-nano
> Running command `find . -type f -name "*.go" -print0 | xargs -0 ls -l --color=auto`...
> Confirm? [Y/n]: Y
...
```

2. Dry Run usage
```
$ nlch --dry-run "Delete all files larger than 4GB"
> Running command `find . -type f -size +4G | rm`...
> This was a dry-run, thus no action was taken.
$ 
```

3. Use a different LLM/provider for this one time
```
$ nlch --model "gemini-2.5-flash-lite" --provider "gemini" "Check if Cloudflare is reachable"
> Running command `ping -c 4 1.1.1.1`...
> Confirm? [Y/n]: Y
PING 1.1.1.1 (1.1.1.1) 56(84) bytes of data.
64 bytes from 1.1.1.1: icmp_seq=1 ttl=52 time=16.7 ms
64 bytes from 1.1.1.1: icmp_seq=2 ttl=52 time=13.5 ms
64 bytes from 1.1.1.1: icmp_seq=3 ttl=52 time=14.1 ms
64 bytes from 1.1.1.1: icmp_seq=4 ttl=52 time=18.2 ms

--- 1.1.1.1 ping statistics ---
4 packets transmitted, 4 received, 0% packet loss, time 3004ms
rtt min/avg/max/mdev = 13.458/15.613/18.205/1.911 ms
```

## Note
Some commands, such as dd, rm, and few others have additional protections, requiring manual confirmation. To override this check, use the `--yes-im-aware-of-the-risk` flag/argument.

# Configuration
The program relies on config files to store your secrets and model providers. The configuration is stored in `~/.config/nlch/config.yaml`. Here is an example configuration:

```yaml
# The default provider to use for generating commands.
# This must match one of the keys under the 'providers' section.
default_provider: openrouter

# Configuration for different LLM providers.
providers:
    # Configuration for the OpenRouter provider.
    openrouter:
        # Your OpenRouter API key.
        key: "sk-or-v1-..."
        # The default model to use for this provider.
        # This can be overridden with the --model flag.
        default_model: "openai/gpt-4o"

    # Configuration for the Google Gemini provider.
    google-gemini:
        # Your Google Gemini API key.
        key: "your-gemini-api-key"
        # The default model to use for this provider.
        default_model: "gemini-1.5-flash"

    # Configuration for Ollama, needs custom url
    ollama:
        url: "https://your-ollama-url"
        default_model: "llama-3.1:8b"

    # we support these model providers:
    # openrouter, gemini, openai, anthropic. ollama
```

# Modularity
The project is highly modular, making it easy to add backends for additional model providers such as Vertex AI, DeepSeek, among others. Additionally, this project supports plugins for additional data to send as part of the context, special system prompts, among others.

---

## Extending nlch

### Adding a New Provider

1. Implement the `Provider` interface in a new file under `internal/provider/`.
2. Register your provider using `provider.Register()` in an init() function.
3. Add your provider's configuration to `~/.config/nlch/config.yaml`.

### Adding a New Plugin

1. Implement the `Plugin` interface in a new file under `internal/plugin/`.
2. Register your plugin using `plugin.Register()` in an init() function.
3. Plugins are automatically invoked during context gathering.

### Modifying Prompts

- Edit or extend prompt logic in `internal/prompt/builder.go`.

---

## Testing

To run tests (if present):

```sh
go test ./...
```

---
