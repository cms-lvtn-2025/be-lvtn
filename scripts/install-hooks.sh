#!/bin/bash

echo "Installing git hooks..."

# Create hooks directory if it doesn't exist
mkdir -p .git/hooks

# Install pre-commit hook
if [ -f ".githooks/pre-commit" ]; then
    ln -sf ../../.githooks/pre-commit .git/hooks/pre-commit
    chmod +x .githooks/pre-commit
    chmod +x .git/hooks/pre-commit
    echo "✅ Pre-commit hook installed"
else
    echo "⚠️  .githooks/pre-commit not found"
fi

echo "Done! Git hooks are now active."
echo ""
echo "To test the hook, try:"
echo "  git add ."
echo "  git commit -m 'test commit'"
