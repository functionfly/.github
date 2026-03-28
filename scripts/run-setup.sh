#!/usr/bin/env bash
# Run cmd/setup with Infisical (dev) when possible, else DB_* from .env / defaults.
# Use SKIP_INFISICAL=1 to force local env only (e.g. invalid token / 403).
set -euo pipefail

cd "$(dirname "$0")/.."

set -a
[ -f .env ] && . ./.env
set +a

run_plain() {
	DB_HOST="${DB_HOST:-localhost}" \
	DB_PORT="${DB_PORT:-5432}" \
	DB_USER="${DB_USER:-postgres}" \
	DB_PASSWORD="${DB_PASSWORD:-postgres}" \
	DB_NAME="${DB_NAME:-functionfly}" \
	DB_SSLMODE="${DB_SSLMODE:-disable}" \
	exec go run ./cmd/setup
}

if [ -n "${SKIP_INFISICAL:-}" ]; then
	run_plain
fi

if command -v infisical >/dev/null 2>&1 && [ -n "${INFISICAL_TOKEN:-}" ]; then
	if infisical run --env=dev -- go run ./cmd/setup; then
		exit 0
	fi
	echo "infisical run failed (e.g. 403). Falling back to DB_* from .env or defaults. To skip Infisical next time: SKIP_INFISICAL=1 make setup" >&2
fi

run_plain
