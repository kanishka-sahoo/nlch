#!/bin/bash

# Simple script to create a new version branch
# Usage: ./new-version.sh v1.0.0

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
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

VERSION=$1

if [[ -z "$VERSION" ]]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

# Validate version format
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
    echo "Invalid version format. Use: v1.0.0 or v1.0.0-beta1"
    exit 1
fi

# Check if branch already exists
if git branch -a | grep -q "$VERSION"; then
    echo "Branch $VERSION already exists"
    exit 1
fi

# Check working directory is clean
if [[ -n $(git status --porcelain) ]]; then
    log_warning "Working directory is not clean. Please commit or stash changes."
    git status --short
    read -p "Continue anyway? [y/N]: " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted"
        exit 0
    fi
fi

# Ensure we're on main
current_branch=$(git branch --show-current)
if [[ "$current_branch" != "main" ]]; then
    log_warning "You're not on the main branch (current: $current_branch)"
    read -p "Switch to main? [y/N]: " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git checkout main
    else
        echo "Aborted"
        exit 0
    fi
fi

# Pull latest
log_info "Pulling latest changes from main..."
git pull origin main

# Create version branch
log_info "Creating version branch $VERSION..."
git checkout -b "$VERSION"

log_success "Version branch $VERSION created!"
log_info "You can now:"
log_info "  1. Make commits to this branch"
log_info "  2. When ready, merge to main (manually or via GitHub PR)"
log_info "  3. GitHub Actions will automatically create the release"
