# MD Documentation Diagnostics - 2026-02-20

Summary of last run:
- LSP-based diagnostics: unavailable due to missing Markdown LSP mapping in OH-MY-OPENCODE environment.
- Fallback diagnostics: executed via markdownlint with a substantial number of issues across AGENTS.md, README.md, VERSIONS.md, and web/README.md.

Key findings (highlights):
- AGENTS.md contains many structural and style issues (line length, missing blank lines around headings, fenced blocks without language, etc.).
- README.md contains numerous markdown style issues (headings spacing, lists, code fences, HTML usage, URLs, and table formatting).
- VERSIONS.md has table formatting issues and lists spacing problems; also line length violations.
- web/README.md has a fenced code block missing language spec on a section; some details blocks (HTML) exist in docs.

Recommended plan to complete the last task:
- Configure a Markdown LSP server in OH-MY-OPENCODE so lsp_diagnostics can run on .md files.
- After LSP mapping is established, run lsp_diagnostics for all modified docs and fix reported issues accordingly.
- Parallelly, apply a staged cleanup to the Markdown docs to meet common MD lint rules (blank lines around headings, proper fencing languages, and removing HTML blocks).
- Re-run build to ensure docs changes do not affect compilation or server startup.

Next steps (proposed commands):
- Install and configure Markdown LSP (example, markdown-language-server) and bind to .md files in OH-MY-OPENCODE.
- Run per-file diagnostics:
  - lsp_diagnostics --file /home/user/workspace/Shepherd/AGENTS.md --severity all
  - lsp_diagnostics --file /home/user/workspace/Shepherd/README.md --severity all
  - lsp_diagnostics --file /home/user/workspace/Shepherd/VERSIONS.md --severity all
  - lsp_diagnostics --file /home/user/workspace/Shepherd/web/README.md --severity all
- Build verification: make build

Notes:
- This report documents the challenges and a concrete path to fully satisfy the last remaining task in the TODO sequence.
- If you want, I can implement an automated patchset to fix the most critical Markdown issues in a staged manner in a follow-up commit.
