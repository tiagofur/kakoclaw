#!/bin/bash
# Script to verify GitHub CLI setup for KakoClaw

echo "🔍 Checking GitHub CLI setup..."
echo

# Check if gh is installed
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI (gh) is not installed"
    echo "Install it with:"
    echo "  macOS:  brew install gh"
    echo "  Linux:  apt install gh (Debian/Ubuntu) or see https://cli.github.com/manual/installation"
    exit 1
fi

echo "✅ GitHub CLI is installed: $(gh --version | head -1)"
echo

# Check authentication status
echo "🔐 Checking authentication..."
if gh auth status &> /dev/null; then
    echo "✅ GitHub CLI is authenticated"
    gh auth status
else
    echo "❌ GitHub CLI is not authenticated"
    echo
    echo "To authenticate, run one of these:"
    echo
    echo "Option 1 - Interactive login (recommended):"
    echo "  gh auth login"
    echo
    echo "Option 2 - Use token from environment:"
    echo "  export GITHUB_TOKEN=ghp_your_token_here"
    echo "  gh auth status"
    echo
    echo "To get a token, visit: https://github.com/settings/tokens"
    exit 1
fi

echo
echo "🎉 Everything is set up correctly!"
echo "You can now use GitHub commands like:"
echo "  gh issue create --repo owner/repo --title 'Issue title' --body 'Description'"
echo "  gh pr list --repo owner/repo"
echo "  gh api repos/owner/repo/issues"
