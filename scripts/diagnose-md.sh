#!/usr/bin/env bash
set -euo pipefail
echo "Markdown linting for key docs..."
MD_FILES=( AGENTS.md README.md VERSIONS.md web/README.md )
echo "Files: ${MD_FILES[*]}"
npx -y markdownlint "${MD_FILES[@]}"
echo "MD lint completed."
