#!/usr/bin/env bash
# check-error-leaks.sh
#
# Fails if any Go file contains patterns that leak internal err.Error() text
# to API clients. This is the safety net for the error-handling hardening
# work documented in docs/superpowers/specs/2026-06-18-api-error-handling-design.md.
#
# Allowed patterns (do NOT fail):
#   - Static string literals: http.Error(w, "msg", status)
#   - Calls wrapped in apierror.LogAndXxx or apierror.FromError
#   - Calls inside test files
#
# Disallowed patterns (FAIL):
#   - http.Error(w, X, Y) where X is err.Error() or contains err.Error()
#   - writeError/respondError/writeJSONError with err.Error() as message
#   - apierror.NewXxx(err.Error()) or with err.Error() in the message
#   - apierror.WriteError(..., apierror.NewXxx("ctx: "+err.Error()))

set -euo pipefail

ROOTS="${1:-internal cmd}"
EXIT=0

# Each pattern: ripgrep regex. -E for extended, -n for line numbers, -t go for Go files.
# We exclude test files and the apierror package itself (which legitimately uses err.Error() in tests).

check_pattern() {
    local label="$1"
    local pattern="$2"
    local matches
    # ripgrep matches; then filter out lines whose first non-whitespace is // or *
    matches=$(rg -n --type go \
        --glob '!internal/apierror/**' \
        --glob '!**/*_test.go' \
        --glob '!scripts/**' \
        --glob '!docs/**' \
        -e "$pattern" \
        $ROOTS 2>/dev/null \
        | awk -F: '
            {
                # Reconstruct: file, line, rest
                rest = $0
                sub(/^[^:]+:[^:]+:/, "", rest)
                # Strip leading whitespace
                sub(/^[ \t]+/, "", rest)
                # Skip comment lines
                if (rest ~ /^\/\// || rest ~ /^\*/) next
                print $0
            }' || true)
    if [ -n "$matches" ]; then
        echo "❌ LEAK PATTERN: $label"
        echo "$matches" | head -20
        echo ""
        EXIT=1
    else
        echo "✅ OK: no $label"
    fi
}

check_pattern "http.Error with err.Error" \
    'http\.Error\(\s*[^,]+,\s*[^,)]*err\.Error\(\)'

check_pattern "respondError with err.Error" \
    'respondError\(\s*[^,]+,\s*[^,]+,\s*[^,)]*err\.Error\(\)'

check_pattern "writeError with err.Error" \
    'writeError\(\s*[^,]+,\s*[^,]+,\s*[^,]+,\s*[^,)]*err\.Error\(\)'

check_pattern "writeJSONError with err.Error" \
    'writeJSONError\(\s*[^,]+,\s*[^,)]*err\.Error\(\)'

check_pattern "apierror.NewXxx with err.Error as message" \
    'apierror\.New[A-Z][a-zA-Z]*\([^)]*err\.Error\(\)[^)]*\)'

check_pattern "fmt.Sprintf error to http.Error" \
    'http\.Error\([^)]*fmt\.Sprintf[^)]*err\.Error\(\)'

if [ $EXIT -ne 0 ]; then
    echo ""
    echo "=========================================="
    echo "  ERROR LEAK DETECTED"
    echo "=========================================="
    echo "Internal error text must not be forwarded to API clients."
    echo "Use apierror.LogAndXxx, apierror.WriteError, or the per-package"
    echo "writeErrorFromErr helper instead. The ErrorNormalizerMiddleware"
    echo "is a safety net, not a fix - do not rely on it for new code."
    exit 1
fi

echo ""
echo "✅ All leak checks passed."
exit 0
