#!/bin/bash

# Release script for nlch
# This script helps create version branches for new releases

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

# Function to check if branch already exists
check_branch_exists() {
    local branch=$1
    if git branch -a | grep -q "origin/$branch\|$branch"; then
        log_error "Branch $branch already exists"
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
    local action=${2:-"create"}
    
    if [[ -z "$version" ]]; then
        echo "Usage: $0 <version> [create|finish]"
        echo ""
        echo "Examples:"
        echo "  $0 v1.0.0          # Create version branch v1.0.0"
        echo "  $0 v1.0.0 create   # Same as above"
        echo "  $0 v1.0.0 finish   # Merge v1.0.0 to main and create release"
        echo ""
        exit 1
    fi
    
    if [[ "$action" == "create" ]]; then
        create_version_branch "$version"
    elif [[ "$action" == "finish" ]]; then
        finish_version_release "$version"
    else
        log_error "Invalid action: $action. Use 'create' or 'finish'"
        exit 1
    fi
}

# Function to create a version branch
create_version_branch() {
    local version=$1
    
    log_info "Creating version branch for $version"
    
    # Validate inputs
    validate_version "$version"
    check_branch_exists "$version"
    check_clean_working_dir
    
    # Ensure we're on the main branch
    current_branch=$(git branch --show-current)
    if [[ "$current_branch" != "main" ]]; then
        log_warning "You're not on the main branch (current: $current_branch)"
        read -p "Switch to main branch? [y/N]: " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            git checkout main
        else
            log_info "Aborted"
            exit 0
        fi
    fi
    
    # Pull latest changes
    log_info "Pulling latest changes from main..."
    git pull origin main
    
    # Create version branch
    log_info "Creating version branch $version..."
    git checkout -b "$version"
    
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
    
    # Push version branch
    log_info "Pushing version branch..."
    git push origin "$version"
    
    log_success "Version branch $version created successfully!"
    log_info "You can now:"
    log_info "  1. Make additional commits to this branch if needed"
    log_info "  2. When ready to release, run: $0 $version finish"
    log_info "  3. Or merge the branch to main manually to trigger the release"
}

# Function to finish a version release
finish_version_release() {
    local version=$1
    
    log_info "Finishing release for $version"
    
    # Validate version
    validate_version "$version"
    check_clean_working_dir
    
    # Check if we're on the version branch
    current_branch=$(git branch --show-current)
    if [[ "$current_branch" != "$version" ]]; then
        log_info "Switching to version branch $version..."
        git checkout "$version"
    fi
    
    # Pull latest changes on version branch
    log_info "Pulling latest changes from version branch..."
    git pull origin "$version"
    
    # Switch to main and merge
    log_info "Switching to main and merging..."
    git checkout main
    git pull origin main
    git merge "$version" --no-ff -m "Merge branch '$version' for release

Release version $version"
    
    # Push main branch
    log_info "Pushing main branch..."
    git push origin main
    
    # Clean up version branch
    log_info "Cleaning up version branch..."
    git branch -d "$version"
    git push origin --delete "$version"
    
    log_success "Release $version process completed successfully!"
    log_info "GitHub Actions will now detect the version change and create the release automatically."
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
        echo "Usage: $0 <version> [create|finish]"
        echo ""
        echo "Examples:"
        echo "  $0 v1.0.0          # Create version branch v1.0.0"
        echo "  $0 v1.0.0 create   # Same as above"  
        echo "  $0 v1.0.0 finish   # Merge v1.0.0 to main and create release"
        echo ""
        echo "This script supports three workflows:"
        echo ""
        echo "Option 1 - Simple (all at once):"
        echo "  $0 v1.0.0 finish   # Creates branch, commits, merges, and triggers release"
        echo ""
        echo "Option 2 - Step by step:"
        echo "  $0 v1.0.0 create   # Creates branch with version bump"
        echo "  # Make additional commits..."
        echo "  $0 v1.0.0 finish   # Merges to main and triggers release"
        echo ""
        echo "Option 3 - Manual merge:"
        echo "  $0 v1.0.0 create   # Creates branch with version bump"
        echo "  # Make additional commits..."
        echo "  # Manually merge to main via GitHub UI or git merge"
        exit 0
        ;;
esac

# Run main function
main "$@"
