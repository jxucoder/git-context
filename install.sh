#!/bin/bash
# ============================================================================
# git-ctx installer
# ============================================================================

set -euo pipefail

INSTALL_DIR="${HOME}/.local/bin"
SCRIPT_NAME="git-ctx"
REPO_URL="https://raw.githubusercontent.com/jiaruixu/git-context/main/git-ctx"

echo "Installing git-ctx..."

# Create install directory if needed
mkdir -p "$INSTALL_DIR"

# Copy or download the script
if [[ -f "git-ctx" ]]; then
    # Local install
    cp git-ctx "$INSTALL_DIR/$SCRIPT_NAME"
else
    # Download from repo
    curl -fsSL "$REPO_URL" > "$INSTALL_DIR/$SCRIPT_NAME"
fi

# Make executable
chmod +x "$INSTALL_DIR/$SCRIPT_NAME"

echo "Installed to: $INSTALL_DIR/$SCRIPT_NAME"

# Check if in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo "⚠️  $INSTALL_DIR is not in your PATH."
    echo ""
    echo "Add this to your ~/.bashrc or ~/.zshrc:"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo ""
fi

# Set up git alias
echo ""
echo "Setting up git alias..."
git config --global alias.ctx '!git-ctx'

echo ""
echo "✅ Done! You can now use:"
echo "   git ctx add --title \"My note\""
echo "   git ctx list"
echo "   git ctx task add \"My task\""
echo ""

