#!/bin/bash

# nlch Installation Script
# This script downloads and installs the latest release of nlch

set -e

# Configuration
REPO="kanishka-sahoo/nlch"
BINARY_NAME="nlch"
# Default install directory - will be adjusted based on OS
INSTALL_DIR="/usr/local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    local os
    local arch
    
    # Detect OS
    case "$(uname -s)" in
        Darwin)
            os="darwin"
            # On macOS, prefer /opt/homebrew/bin if it exists (Apple Silicon), otherwise /usr/local/bin
            if [[ "$(uname -m)" == "arm64" ]] && [[ -d "/opt/homebrew/bin" ]]; then
                INSTALL_DIR="/opt/homebrew/bin"
            else
                INSTALL_DIR="/usr/local/bin"
            fi
            ;;
        Linux)
            os="linux"
            ;;
        CYGWIN*|MINGW*|MSYS*)
            os="windows"
            ;;
        *)
            log_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)
            arch="amd64"
            ;;
        arm64|aarch64)
            arch="arm64"
            ;;
        *)
            log_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
    
    echo "${os}-${arch}"
}

# Get latest release info from GitHub API
get_latest_release() {
    local api_url="https://api.github.com/repos/${REPO}/releases/latest"
    local release_info
    
    log_info "Fetching latest release information..."
    
    if command -v curl >/dev/null 2>&1; then
        release_info=$(curl -s "$api_url")
    elif command -v wget >/dev/null 2>&1; then
        release_info=$(wget -qO- "$api_url")
    else
        log_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi
    
    if [ -z "$release_info" ]; then
        log_error "Failed to fetch release information"
        exit 1
    fi
    
    echo "$release_info"
}

# Extract download URL for the platform
get_download_url() {
    local release_info="$1"
    local platform="$2"
    local binary_suffix=""
    
    if [[ "$platform" == *"windows"* ]]; then
        binary_suffix=".exe"
    fi
    
    local asset_name="${BINARY_NAME}-${platform}${binary_suffix}"
    
    # Extract download URL using basic text processing (compatible with most systems)
    local download_url
    download_url=$(echo "$release_info" | grep -o "\"browser_download_url\":\s*\"[^\"]*${asset_name}\"" | cut -d'"' -f4)
    
    if [ -z "$download_url" ]; then
        log_error "No release asset found for platform: $platform"
        log_info "Available assets:"
        echo "$release_info" | grep -o "\"name\":\s*\"[^\"]*\"" | cut -d'"' -f4 | grep "$BINARY_NAME" || true
        exit 1
    fi
    
    echo "$download_url"
}

# Download and install binary
install_binary() {
    local download_url="$1"
    local platform="$2"
    local temp_dir
    local binary_suffix=""
    
    if [[ "$platform" == *"windows"* ]]; then
        binary_suffix=".exe"
    fi
    
    temp_dir=$(mktemp -d)
    local temp_file="${temp_dir}/${BINARY_NAME}${binary_suffix}"
    
    log_info "Downloading ${BINARY_NAME} from: $download_url"
    
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$temp_file" "$download_url"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$temp_file" "$download_url"
    else
        log_error "Neither curl nor wget is available"
        exit 1
    fi
    
    if [ ! -f "$temp_file" ]; then
        log_error "Failed to download binary"
        exit 1
    fi
    
    # Make binary executable
    chmod +x "$temp_file"
    
    # Create install directory if it doesn't exist
    if [ ! -d "$INSTALL_DIR" ]; then
        log_info "Creating install directory: $INSTALL_DIR"
        sudo mkdir -p "$INSTALL_DIR"
    fi
    
    # Install binary
    local install_path="${INSTALL_DIR}/${BINARY_NAME}"
    log_info "Installing to: $install_path"
    
    if [ -w "$INSTALL_DIR" ]; then
        cp "$temp_file" "$install_path"
    else
        sudo cp "$temp_file" "$install_path"
    fi
    
    # Clean up
    rm -rf "$temp_dir"
    
    log_success "${BINARY_NAME} installed successfully to $install_path"
}

# Verify installation
verify_installation() {
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local version
        version=$("$BINARY_NAME" --version 2>/dev/null || echo "unknown")
        log_success "Installation verified. Version: $version"
        log_info "You can now use '${BINARY_NAME}' from anywhere in your terminal"
        
        # Show update information
        log_info "Auto-update features:"
        log_info "  • Automatic update checks (once per day)"
        log_info "  • Run '${BINARY_NAME} --update' to update manually"
        log_info "  • Run '${BINARY_NAME} --check-update' to check for updates"
    else
        log_warning "Binary installed but not found in PATH"
        log_info "You may need to add $INSTALL_DIR to your PATH or restart your terminal"
        
        # macOS-specific PATH guidance
        if [[ "$(uname -s)" == "Darwin" ]]; then
            log_info "On macOS, add this line to your shell profile (~/.zshrc or ~/.bash_profile):"
        else
            log_info "To add to PATH, add this line to your shell profile:"
        fi
        log_info "  export PATH=\"\$PATH:$INSTALL_DIR\""
    fi
}

# Main installation process
main() {
    log_info "Starting nlch installation..."
    
    # Check if running as root (not recommended for the script itself)
    if [ "$EUID" -eq 0 ]; then
        log_warning "Running as root. The script will still work, but it's not recommended."
    fi
    
    # macOS-specific checks
    if [[ "$(uname -s)" == "Darwin" ]]; then
        # Check for Homebrew and suggest it if available
        if command -v brew >/dev/null 2>&1; then
            log_info "Homebrew detected. Note: You can also install via Homebrew in the future."
            log_info "  brew tap kanishka-sahoo/nlch && brew install nlch"
        fi
        
        # Check for Xcode Command Line Tools
        if ! xcode-select -p >/dev/null 2>&1; then
            log_warning "Xcode Command Line Tools not found. Some features may not work."
            log_info "Install with: xcode-select --install"
        fi
    fi
    
    # Detect platform
    local platform
    platform=$(detect_platform)
    log_info "Detected platform: $platform"
    
    # Get latest release
    local release_info
    release_info=$(get_latest_release)
    
    # Extract version
    local version
    version=$(echo "$release_info" | grep -o "\"tag_name\":\s*\"[^\"]*\"" | cut -d'"' -f4)
    log_info "Latest version: $version"
    
    # Get download URL
    local download_url
    download_url=$(get_download_url "$release_info" "$platform")
    
    # Install binary
    install_binary "$download_url" "$platform"
    
    # Verify installation
    verify_installation
    
    log_success "Installation complete!"
    log_info "Run '${BINARY_NAME} --help' to get started"
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "nlch Installation Script"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --version, -v  Show script version"
        echo ""
        echo "Environment Variables:"
        echo "  INSTALL_DIR    Installation directory (default: /usr/local/bin)"
        echo ""
        echo "This script downloads and installs the latest release of nlch"
        echo "from the GitHub repository: https://github.com/${REPO}"
        exit 0
        ;;
    --version|-v)
        echo "nlch installation script v1.0.0"
        exit 0
        ;;
esac

# Override install directory if specified
if [ -n "${INSTALL_DIR_OVERRIDE:-}" ]; then
    INSTALL_DIR="$INSTALL_DIR_OVERRIDE"
fi

# Run main installation
main "$@"
