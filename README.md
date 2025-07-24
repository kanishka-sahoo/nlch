# Natural Language Command Help#### Build

```sh
go mod tidy         # Install dependencies
make build          # Build binaries for all supported OS/ARCH
make install        # Install the binary for your current OS

# To clean up built binaries
make clean
```

This project supports cross-platform binary generation for:
- Linux (amd64, arm64)
- Windows (amd64)
- macOS (amd64, arm64)

The `make install` command will copy the correct binary to a standard location for your OS:
- On Linux/macOS: `/usr/local/bin/nlch`
- On Windows: `%USERPROFILE%\bin\nlch.exe`

### Manual Installation

You can also manually download the latest release:

1. Go to the [Releases page](https://github.com/kanishka-sahoo/nlch/releases/latest)
2. Download the appropriate binary for your OS and architecture
3. Extract and place it in your PATH
4. Make it executable (Linux/macOS): `chmod +x nlch`inal program designed to help you be more productive. It uses context clues from the location it is invoked—such as the current directory, git information, the list of files and folders present—and uses it to generate a terminal command based on natural language input.

---

## Getting Started

### Quick Install (Recommended)

**Linux/macOS:**
```sh
curl -fsSL https://raw.githubusercontent.com/kanishka-sahoo/nlch/main/install.sh | bash
```

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/kanishka-sahoo/nlch/main/install.ps1 | iex
```

The installation scripts will:
- Automatically detect your OS and architecture
- Download the latest release from GitHub
- Install the binary to a standard location
- Verify the installation

### Build from Source

#### Prerequisites

- Go 1.20 or newer (https://golang.org/dl/)

### Build

```sh
go mod tidy         # Install dependencies
make build          # Build binaries for all supported OS/ARCH
make install        # Install the binary for your current OS

# To clean up built binaries
make clean
```

This project supports cross-platform binary generation for:
- Linux (amd64, arm64)
- Windows (amd64)
- macOS (amd64, arm64)

The `make install` command will copy the correct binary to a standard location for your OS:
- On Linux/macOS: `/usr/local/bin/nlch`
- On Windows: `%USERPROFILE%\bin\nlch.exe`

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
- `--update` — Check for and install updates
- `--check-update` — Check for updates without installing

### Configuration

After installation, you'll need to create a configuration file at `~/.config/nlch/nlch.yaml` (Linux/macOS) or `%APPDATA%\nlch\nlch.yaml` (Windows).

See below for configuration file details and examples.

### Auto-Updates

nlch includes built-in update functionality:

- **Automatic Check**: nlch automatically checks for updates once per day and notifies you if a new version is available
- **Manual Update**: Run `nlch --update` to check for and install updates immediately
- **Check Only**: Run `nlch --check-update` to check for updates without installing

The update system:
- Downloads the latest release from GitHub
- Automatically detects your OS and architecture
- Safely replaces the current binary
- Works on Linux, macOS, and Windows

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

## Development

### Creating a Release

For maintainers, creating a new release is automated:

```sh
# Create a new release (will trigger automated build and publish)
./release.sh v1.0.0

# Create a pre-release
./release.sh v1.0.0-beta1
```

The release script will:
1. Update version numbers in all relevant files
2. Build and test the binary
3. Create a git tag and push it
4. Trigger GitHub Actions to build binaries for all platforms
5. Automatically create a GitHub release with:
   - Pre-built binaries for all supported platforms
   - Checksums for verification
   - Auto-generated changelog
   - Installation instructions

### Release Workflow

The GitHub Actions workflow automatically:
- **Builds** binaries for Linux, macOS, and Windows (multiple architectures)
- **Creates** a GitHub release with detailed release notes
- **Uploads** all binaries and checksums
- **Updates** the Homebrew formula with new version and checksums
- **Tests** the installation scripts against the new release

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

## Release Process

This project uses an automated release system based on version branches:

### For Maintainers

**Option 1 - Quick Release:**
```sh
./new-version.sh v1.0.0          # Create version branch
# Make any final changes...
git push origin v1.0.0           # Push changes
# Create a merge/PR from v1.0.0 to main
# GitHub Actions automatically creates the release when merged
```

**Option 2 - Using the Release Script:**
```sh
./release.sh v1.0.0 create       # Create version branch with auto-version bump
# Make additional commits if needed...
./release.sh v1.0.0 finish       # Merge to main and trigger release
```

**Option 3 - Manual Process:**
```sh
git checkout main
git checkout -b v1.0.0
# Update version in main.go, nlch.rb, internal/update/update.go
git commit -m "Bump version to v1.0.0"
git push origin v1.0.0
# Merge to main via GitHub UI or command line
# Release is automatically created by GitHub Actions
```

### How It Works

1. **Version Branch**: Create a branch named with the semantic version (e.g., `v1.0.0`)
2. **Merge to Main**: When the version branch is merged to main, GitHub Actions detects it
3. **Automatic Release**: The workflow automatically:
   - Builds binaries for all platforms (Linux, macOS, Windows)
   - Creates a Git tag with the version
   - Creates a GitHub release with all binaries
   - Updates the Homebrew formula
   - Generates release notes with changelog

The release system supports both stable versions (`v1.0.0`) and pre-releases (`v1.0.0-beta1`).

---
