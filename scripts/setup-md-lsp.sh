#!/usr/bin/env bash
set -euo pipefail
echo "Setting up Markdown LSP server for OH-MY-OPENCODE..."

# 1) Install a Markdown language server (markdown-language-server)
if ! command -v markdown-language-server >/dev/null 2>&1; then
  echo "Installing markdown-language-server..."
  npm i -g vscode-langservers-extracted >/dev/null 2>&1 || true
  # Fallback to markdown-language-server if available via npm package name
  if ! command -v markdown-language-server >/dev/null 2>&1; then
    echo "markdown-language-server not found after install. Please install manually." >&2
    exit 1
  fi
else
  echo "markdown-language-server already installed."
fi

# 2) Create a simple LSP mapping config (example path)
CONFIG_DIR="$HOME/.oh-my-opencode/lsp"
mkdir -p "$CONFIG_DIR"
cat > "$CONFIG_DIR/markdown.yaml" << 'YAML'
extensions:
- .md
command: ["markdown-language-server", "--stdio"]
YAML
echo "Wrote MD LSP mapping to $CONFIG_DIR/markdown.yaml"

echo "Please ensure OH-MY-OPENCODE loads this LSP mapping."
echo "If needed, restart the OH-MY-OPENCODE service to apply changes."
