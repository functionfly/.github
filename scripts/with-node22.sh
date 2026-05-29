#!/usr/bin/env bash
# Ensure Node.js >= 22.12 (required by Astro 6) before running a command.
set -euo pipefail

node_ok() {
  command -v node >/dev/null 2>&1 || return 1
  node -e 'const p=process.version.slice(1).split(".").map(Number); process.exit(p[0]>22||(p[0]===22&&p[1]>=12)?0:1)' 2>/dev/null
}

if ! node_ok; then
  if [ -s "${NVM_DIR:-$HOME/.nvm}/nvm.sh" ]; then
    # shellcheck disable=SC1091
    source "${NVM_DIR:-$HOME/.nvm}/nvm.sh"
    nvm use 22 >/dev/null 2>&1 || nvm use >/dev/null 2>&1 || true
  elif command -v fnm >/dev/null 2>&1; then
    eval "$(fnm env)"
    fnm use 22 >/dev/null 2>&1 || fnm use >/dev/null 2>&1 || true
  elif command -v mise >/dev/null 2>&1; then
    mise activate bash >/dev/null 2>&1 || true
  fi
fi

if ! node_ok; then
  echo "Error: Node.js >= 22.12.0 is required (Astro 6). Run: nvm install 22 && nvm use 22" >&2
  exit 1
fi

exec "$@"
