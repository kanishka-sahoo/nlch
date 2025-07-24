#!/bin/bash

# Release script for nlch
# This script helps create new releases with proper versioning

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Function to validate version format
validate_version() {
    local version=$1
    if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
        log_error "Invalid version format. Use: v1.0.0 or v1.0.0-beta1"
        exit 1
    fi
}

# Function to check if tag already exists
check_tag_exists() {
    local tag=$1
    if git tag -l | grep -q "^$tag$"; then
        log_error "Tag $tag already exists"
        exit 1
    fi
}

# Function to check working directory is clean
check_clean_working_dir() {
    if [[ -n $(git status --porcelain) ]]; then
        log_error "Working directory is not clean. Please commit or stash changes."
        git status --short
        exit 1
    fi
}

# Function to update version in files
update_version_in_files() {
    local version=$1
    local version_no_v=${version#v}
    
    log_info "Updating version in source files..."
    
    # Update main.go
    sed -i.bak "s/const version = \".*\"/const version = \"$version_no_v\"/" main.go
    rm -f main.go.bak
    
    # Update Homebrew formula
    sed -i.bak "s/version \".*\"/version \"$version_no_v\"/" nlch.rb
    rm -f nlch.rb.bak
    
    # Update update package
    sed -i.bak "s/var BuildVersion = \".*\"/var BuildVersion = \"$version_no_v\"/" internal/update/update.go
    rm -f internal/update/update.go.bak
    
    log_success "Version updated to $version in source files"
}

# Main function
main() {
    local version=$1
    
    if [[ -z "$version" ]]; then
        echo "Usage: $0 <version>"
        echo "Example: $0 v1.0.0"
        echo "         $0 v1.0.0-beta1"
        exit 1
    fi
    
    log_info "Starting release process for $version"
    
    # Validate inputs
    validate_version "$version"
    check_tag_exists "$version"
    check_clean_working_dir
    
    # Ensure we're on the main branch
    current_branch=$(git branch --show-current)
    if [[ "$current_branch" != "main" ]]; then
        log_warning "You're not on the main branch (current: $current_branch)"
        read -p "Continue anyway? [y/N]: " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Aborted"
            exit 0
        fi
    fi
    
    # Update version in files
    update_version_in_files "$version"
    
    # Build and test
    log_info "Building and testing..."
    go mod tidy
    go build -o nlch-test
    ./nlch-test --version
    rm -f nlch-test
    
    # Commit version changes
    log_info "Committing version changes..."
    git add main.go nlch.rb internal/update/update.go
    git commit -m "Bump version to $version"
    
    # Create and push tag
    log_info "Creating tag $version..."
    git tag -a "$version" -m "Release $version"
    
    log_info "Pushing changes and tag..."
    git push origin main
    git push origin "$version"
    
    log_success "Release $version created successfully!"
    log_info "GitHub Actions will now build and create the release automatically."
    log_info "You can monitor the progress at:"
    log_info "  https://github.com/kanishka-sahoo/nlch/actions"
    log_info ""
    log_info "The release will be available at:"
    log_info "  https://github.com/kanishka-sahoo/nlch/releases/tag/$version"
}

# Handle help
case "${1:-}" in
    --help|-h)
        echo "nlch Release Script"
        echo ""
        echo "Usage: $0 <version>"
        echo ""
        echo "Examples:"
        echo "  $0 v1.0.0        # Create stable release"
        echo "  $0 v1.0.0-beta1  # Create pre-release"
        echo ""
        echo "This script will:"
        echo "  1. Validate the version format"
        echo "  2. Check that the working directory is clean"
        echo "  3. Update version in source files"
        echo "  4. Build and test the binary"
        echo "  5. Commit the version changes"
        echo "  6. Create and push a git tag"
        echo "  7. Trigger GitHub Actions to build and release"
        exit 0
        ;;
esac

# Run main function
main "$@"
