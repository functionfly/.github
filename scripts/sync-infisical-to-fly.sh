#!/usr/bin/env bash
# Sync secrets from Infisical to Fly.io (one-way: canonical store → runtime).
#
# Prerequisites:
#   - infisical CLI (https://infisical.com/docs/cli/overview)
#   - fly CLI, logged in: fly auth whoami
#   - INFISICAL_TOKEN (service token) or infisical login session
#
# Usage:
#   INFISICAL_ENV=prod FLY_APP=functionfly-control ./scripts/sync-infisical-to-fly.sh
#   STAGE=1 ./scripts/sync-infisical-to-fly.sh   # fly secrets import --stage (no immediate deploy)
#
# EU / self-hosted Infisical:
#   INFISICAL_API_URL=https://eu.infisical.com/api ./scripts/sync-infisical-to-fly.sh
#   # or: infisical --domain https://eu.infisical.com
#
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ALLOWLIST_FILE="${FLY_SECRETS_ALLOWLIST:-$ROOT/scripts/fly-secrets-allowlist.txt}"
INFISICAL_ENV="${INFISICAL_ENV:-prod}"
FLY_APP="${FLY_APP:-functionfly-control}"
STAGE="${STAGE:-}"

if [[ ! -f "$ALLOWLIST_FILE" ]]; then
  echo "Allowlist not found: $ALLOWLIST_FILE" >&2
  exit 1
fi

if ! command -v infisical >/dev/null 2>&1; then
  echo "infisical CLI not found. Install: https://infisical.com/docs/cli/overview" >&2
  exit 1
fi

FLY_BIN=fly
if ! command -v "$FLY_BIN" >/dev/null 2>&1; then
  FLY_BIN=flyctl
fi
if ! command -v "$FLY_BIN" >/dev/null 2>&1; then
  echo "Fly CLI not found (tried fly, flyctl). Install: curl -L https://fly.io/install.sh | sh" >&2
  exit 1
fi

if ! "$FLY_BIN" auth whoami >/dev/null 2>&1; then
  echo "Not authenticated to Fly. Run: fly auth login  (or set FLY_API_TOKEN for CI)" >&2
  exit 1
fi

# Build grep pattern from allowlist (only non-comment, non-empty lines)
mapfile -t _ALLOWED_KEYS < <(grep -v '^[[:space:]]*#' "$ALLOWLIST_FILE" | grep -v '^[[:space:]]*$' | sed 's/[[:space:]]//g')

if [[ ${#_ALLOWED_KEYS[@]} -eq 0 ]]; then
  echo "Allowlist is empty: $ALLOWLIST_FILE" >&2
  exit 1
fi

filter_dotenv() {
  while IFS= read -r line || [[ -n "${line:-}" ]]; do
    [[ -z "$line" ]] && continue
    [[ "$line" =~ ^[[:space:]]*# ]] && continue
    # KEY=... (first = separates key from value)
    if [[ ! "$line" =~ ^([A-Za-z_][A-Za-z0-9_]*)=(.*)$ ]]; then
      continue
    fi
    key="${BASH_REMATCH[1]}"
    case "$key" in
      VITE_*|INFISICAL_*)
        continue
        ;;
    esac
    local ok=0
    for allowed in "${_ALLOWED_KEYS[@]}"; do
      [[ "$key" == "$allowed" ]] && ok=1 && break
    done
    [[ "$ok" -eq 1 ]] || continue
    printf '%s\n' "$line"
  done
}

echo "→ Exporting Infisical env: $INFISICAL_ENV (project from .infisical.json or INFISICAL_PROJECT_ID)"
TMP_DOTENV="$(mktemp)"
FILTERED="$(mktemp)"
trap 'rm -f "$TMP_DOTENV" "$FILTERED"' EXIT

infisical export --env="$INFISICAL_ENV" --format=dotenv --silent --output-file="$TMP_DOTENV"

if [[ ! -s "$TMP_DOTENV" ]]; then
  echo "No secrets exported (empty file). Check INFISICAL_TOKEN, environment name, and project access." >&2
  exit 1
fi
filter_dotenv <"$TMP_DOTENV" >"$FILTERED"

LINE_COUNT=$(wc -l <"$FILTERED" | tr -d ' ')
if [[ "$LINE_COUNT" -eq 0 ]]; then
  echo "No secrets matched allowlist after filter. Keys in Infisical may use different names than $ALLOWLIST_FILE" >&2
  exit 1
fi

echo "→ Importing $LINE_COUNT secret(s) to Fly app: $FLY_APP"

STAGE_ARGS=()
if [[ -n "$STAGE" && "$STAGE" != "0" ]]; then
  STAGE_ARGS=(--stage)
  echo "  (staging only; machines pick up on next deploy/restart)"
fi

"$FLY_BIN" secrets import --app "$FLY_APP" "${STAGE_ARGS[@]}" <"$FILTERED"

echo "Done. Verify: $FLY_BIN secrets list --app $FLY_APP"
