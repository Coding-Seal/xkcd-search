#!/usr/bin/env bash
# count_tests.sh — counts all Go test functions in the project
# Usage: ./scripts/count_tests.sh [--verbose]

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

VERBOSE=false
if [[ "${1:-}" == "--verbose" || "${1:-}" == "-v" ]]; then
    VERBOSE=true
fi

echo "=== Test Counter ==="
echo "Project: $ROOT"
echo ""

total=0
declare -A pkg_counts

while IFS= read -r file; do
    count=$(grep -c "^func Test" "$file" 2>/dev/null) || count=0
    count="${count//[^0-9]/}"   # strip any non-digits (newlines, etc.)
    count="${count:-0}"
    if [[ "$count" -gt 0 ]]; then
        pkg=$(dirname "$file" | sed "s|$ROOT/||")
        prev="${pkg_counts[$pkg]:-0}"
        pkg_counts["$pkg"]=$(( prev + count ))
        total=$(( total + count ))

        if $VERBOSE; then
            echo "  $file: $count"
        fi
    fi
done < <(find "$ROOT" -name "*_test.go" -not -path "*/vendor/*")

echo "--- Tests per package ---"
for pkg in $(echo "${!pkg_counts[@]}" | tr ' ' '\n' | sort); do
    printf "  %-60s %d\n" "$pkg" "${pkg_counts[$pkg]}"
done

echo ""
echo "=== TOTAL TEST FUNCTIONS: $total ==="

# Check against requirement
REQUIRED=150
if [[ "$total" -ge "$REQUIRED" ]]; then
    echo "✓ Meets minimum requirement of $REQUIRED (25 per member × 6)"
else
    echo "✗ Below minimum requirement of $REQUIRED — need $(( REQUIRED - total )) more"
fi

echo ""
echo "--- Coverage (last run) ---"
if [[ -f "$ROOT/build/cover.out" ]]; then
    go tool cover -func="$ROOT/build/cover.out" 2>/dev/null | tail -1
else
    echo "  No coverage report found. Run: make test"
fi
