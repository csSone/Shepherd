#!/usr/bin/env bash
set -euo pipefail
FILES=( AGENTS.md README.md VERSIONS.md web/README.md )
echo "Running diagnostics via lsp_diagnostics where available; otherwise using markdownlint as fallback."

FAILED=0
for f in "${FILES[@]}"; do
  if command -v oh-my-opencode-diagnostics >/dev/null 2>&1; then
    echo "Running lsp diagnostics for $f"
    oh-my-opencode-diagnostics --file "$f" --severity all || FAILED=$((FAILED+1))
  else
    echo "lsp diagnostics tool not available for $f; skipping."
    echo "Falling back to markdownlint"
    if command -v markdownlint >/dev/null 2>&1; then
      markdownlint "$f" || FAILED=$((FAILED+1))
    fi
  fi
done

if [ "$FAILED" -ne 0 ]; then
  echo "Some diagnostics failed. See above messages for details."
  exit 2
fi
echo "All diagnostics completed successfully."
