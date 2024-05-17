#!/bin/sh

# Copy pre-push hook to .git/hooks/
cp ./.githooks/pre-push ./.git/hooks/pre-push

# Make the pre-push hook executable
chmod +x .git/hooks/pre-push

echo "Git push hooks installed successfully."
