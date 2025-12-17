#!/bin/bash
# Script to update gh-pages branch with site/ directory contents
# This script creates a clean gh-pages branch with only the verification site files

set -e

SITE_DIR="site"

# Check if site directory exists
if [ ! -d "$SITE_DIR" ]; then
    echo "Error: $SITE_DIR directory not found"
    exit 1
fi

echo "Updating gh-pages branch with site/ contents..."

# Save current branch
CURRENT_BRANCH=$(git branch --show-current)

# Switch to gh-pages
git checkout gh-pages

# Remove all files except .git and .github
find . -maxdepth 1 ! -name '.git' ! -name '.github' ! -name '.' -exec rm -rf {} +

# Copy site files (excluding server.go which is for local testing only)
cp -r "$SITE_DIR"/* .
rm -f server.go

# Copy workflow from site/.github if it exists
if [ -d "$SITE_DIR/.github" ]; then
    cp -r "$SITE_DIR/.github"/* .github/
fi

# Stage and commit
git add -A
git commit -m "Update crawler verification site" || echo "No changes to commit"

echo ""
echo "Done! gh-pages branch updated."
echo ""
echo "To push:  git push origin gh-pages"
echo ""

# Switch back
git checkout "$CURRENT_BRANCH"
