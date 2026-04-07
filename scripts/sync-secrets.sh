#!/usr/bin/env bash
set -uo pipefail

DRY_RUN=false
TARGET="all"

for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=true ;;
    --vercel-only) TARGET="vercel" ;;
    --fly-only) TARGET="fly" ;;
    *) echo "Unknown arg: $arg"; exit 1 ;;
  esac
done

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
VERCEL="$HOME/.local/bin/vercel"
FLYCTL="$HOME/.fly/bin/flyctl"

parse_line() {
  local line="$1"
  line="${line#export }"
  KEY="${line%%=*}"
  VALUE="${line#*=}"
  VALUE="${VALUE%\"}"; VALUE="${VALUE#\"}"
  VALUE="${VALUE%\'}"; VALUE="${VALUE#\'}"
}

set_vercel_var() {
  local dir="$1" key="$2" value="$3" envs="$4"
  cd "$ROOT_DIR/$dir"

  if $DRY_RUN; then
    echo "  [dry-run] $key -> $envs"
    return 0
  fi

  # Remove from all envs first
  for env in production preview development; do
    "$VERCEL" env rm "$key" "$env" --yes 2>/dev/null || true
  done

  IFS=',' read -ra ENV_LIST <<< "$envs"
  for env in "${ENV_LIST[@]}"; do
    echo "$value" | "$VERCEL" env add "$key" "$env" --yes 2>&1 | grep -E "Added|Error" || true
  done
  return 0
}

sync_vercel() {
  echo "=== VERCEL ENVIRONMENT SYNC ==="
  local count=0

  # ── Root: .env (secrets only — skip infra/frontend vars) ──
  echo ""
  echo "--- Vercel: functionfly (.env secrets) ---"
  cd "$ROOT_DIR"
  while IFS= read -r line; do
    [ -z "$line" ] && continue; [[ "$line" == \#* ]] && continue
    parse_line "$line"
    [ -z "$KEY" ] && continue
    case "$KEY" in
      VITE_*|DB_*|REDIS_*|CACHE_*|ADVANCED_SECURITY_*|ARCHIVE_*|RATE_LIMIT_*|\
      VERIFICATION_*|TRUST_LEVEL_*|STORAGE_*|AWS_*|PORT|DEVELOPMENT|LOG_LEVEL|\
      EDGE_*|CORS_*|CONTENT_SECURITY_POLICY|HSTS_*|RAG_*|OPENROUTER_*|\
      DEFAULT_PROVIDER|DB_ENCRYPTION_*|DB_MASTER_KEY*|DB_SSL*|DB_MAX_*|\
      DB_CONNECTION_*) continue ;;
    esac
    set_vercel_var "." "$KEY" "$VALUE" "development,preview,production"
    count=$((count + 1))
  done < <(grep -v '^\s*#' .env 2>/dev/null | grep -v '^\s*$' || true)

  # ── Root: .env.local ──
  echo ""
  echo "--- Vercel: functionfly (.env.local) ---"
  while IFS= read -r line; do
    [ -z "$line" ] && continue; [[ "$line" == \#* ]] && continue
    parse_line "$line"
    [ -z "$KEY" ] && continue
    case "$KEY" in
      VITE_*|DB_*|REDIS_*|CACHE_*|ADVANCED_SECURITY_*|ARCHIVE_*|RATE_LIMIT_*|\
      VERIFICATION_*|TRUST_LEVEL_*|STORAGE_*|AWS_*|PORT|DEVELOPMENT|LOG_LEVEL|\
      EDGE_*|CORS_*|CONTENT_SECURITY_POLICY|HSTS_*|RAG_*|OPENROUTER_*|\
      DEFAULT_PROVIDER|DB_ENCRYPTION_*|DB_MASTER_KEY*|DB_SSL*|DB_MAX_*|\
      DB_CONNECTION_*) continue ;;
    esac
    set_vercel_var "." "$KEY" "$VALUE" "development,preview,production"
    count=$((count + 1))
  done < <(grep -v '^\s*#' .env.local 2>/dev/null | grep -v '^\s*$' || true)

  # ── Auth project ──
  echo ""
  echo "--- Vercel: auth (web/auth/.env.local) ---"
  while IFS= read -r line; do
    [ -z "$line" ] && continue; [[ "$line" == \#* ]] && continue
    parse_line "$line"
    [ -z "$KEY" ] && continue
    set_vercel_var "web/auth" "$KEY" "$VALUE" "preview,production"
    count=$((count + 1))
  done < <(grep -v '^\s*#' web/auth/.env.local 2>/dev/null | grep -v '^\s*$' || true)

  # ── Dashboard project ──
  echo ""
  echo "--- Vercel: dashboard (web/dashboard/.env.production) ---"
  while IFS= read -r line; do
    [ -z "$line" ] && continue; [[ "$line" == \#* ]] && continue
    parse_line "$line"
    [ -z "$KEY" ] && continue
    set_vercel_var "web/dashboard" "$KEY" "$VALUE" "production"
    count=$((count + 1))
  done < <(grep -v '^\s*#' web/dashboard/.env.production 2>/dev/null | grep -v '^\s*$' || true)

  echo ""
  echo "  Vercel sync done: $count vars"
}

sync_fly() {
  echo ""
  echo "=== FLY.IO SECRETS SYNC ==="

  local app="functionfly-control"
  local count=0
  local secrets_json="{}"

  cd "$ROOT_DIR"

  if [ -z "${FLY_API_TOKEN:-}" ]; then
    parse_line "$(grep '^FLY_API_TOKEN=' .env 2>/dev/null | head -1)"
    export FLY_API_TOKEN="$VALUE"
  fi

  while IFS= read -r line; do
    [ -z "$line" ] && continue; [[ "$line" == \#* ]] && continue
    parse_line "$line"
    [ -z "$KEY" ] && continue
    case "$KEY" in
      VITE_*|DB_HOST|DB_PORT|DB_SSLMODE|REDIS_ADDR|PORT|DEVELOPMENT|FLY_API_TOKEN*) continue ;;
    esac
    secrets_json=$(echo "$secrets_json" | jq --arg k "$KEY" --arg v "$VALUE" '. + {($k): $v}')
    count=$((count + 1))
  done < <(grep -v '^\s*#' .env 2>/dev/null | grep -v '^\s*$' || true)

  if $DRY_RUN; then
    echo "  [dry-run] Would set $count secrets on Fly app: $app"
    echo "$secrets_json" | jq -r 'keys[]' | sort
    return 0
  fi

  local tmpline
  tmpline=$(mktemp /tmp/fly-secrets-XXXXXX.env)
  echo "$secrets_json" | jq -r 'to_entries|map("\(.key)=\(.value)")|.[]' > "$tmpline"

  echo "  Importing $count secrets to Fly app: $app"
  cat "$tmpline" | "$FLYCTL" secrets import --app "$app" 2>&1 || {
    echo "  ERROR: flyctl secrets import failed."
    cat "$tmpline" | head -3
    echo "  ..."
    rm -f "$tmpline"
    return 1
  }

  rm -f "$tmpline"
  echo "  Fly.io sync done: $count secrets imported."
}

case "$TARGET" in
  all)    sync_vercel; sync_fly ;;
  vercel) sync_vercel ;;
  fly)    sync_fly ;;
esac

echo ""
echo "Secret sync complete."
