#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PACKAGES_FILE="$ROOT_DIR/scripts/coverage_packages.txt"
GO_CMD="${GO_CMD:-go}"
GOTOOLCHAIN_VALUE="${GOTOOLCHAIN:-go1.25.8}"
THRESHOLD="${COVERAGE_THRESHOLD:-30}"
PROFILE_PATH="${COVERAGE_OUT:-$ROOT_DIR/.artifacts/coverage/coverage.out}"
SUMMARY_PATH="${COVERAGE_SUMMARY:-$ROOT_DIR/.artifacts/coverage/summary.txt}"
HTML_PATH="${COVERAGE_HTML:-$ROOT_DIR/.artifacts/coverage/coverage.html}"
MODE="${1:-report}"

mkdir -p "$(dirname "$PROFILE_PATH")"
rm -f "$PROFILE_PATH" "$SUMMARY_PATH" "$HTML_PATH"
echo "mode: set" >"$PROFILE_PATH"

packages=()
while IFS= read -r line; do
  if [[ -n "$line" ]] && [[ ! "$line" =~ ^[[:space:]]*# ]]; then
    packages+=("$line")
  fi
done <"$PACKAGES_FILE"

if [[ ${#packages[@]} -eq 0 ]]; then
  echo "no coverage packages configured" >&2
  exit 1
fi

tmp_dir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT

cd "$ROOT_DIR"

for pkg in "${packages[@]}"; do
  safe_name="$(echo "$pkg" | tr '/.' '__')"
  profile="$tmp_dir/${safe_name}.out"

  if ! GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" "$GO_CMD" test -covermode=set -coverprofile="$profile" "$pkg" >/dev/null; then
    echo "coverage test failed for $pkg" >&2
    exit 1
  fi

  if [[ -f "$profile" ]] && [[ "$(wc -l <"$profile")" -gt 1 ]]; then
    tail -n +2 "$profile" >>"$PROFILE_PATH"
  fi
done

GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" "$GO_CMD" tool cover -func="$PROFILE_PATH" | tee "$SUMMARY_PATH"
GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" "$GO_CMD" tool cover -html="$PROFILE_PATH" -o "$HTML_PATH"

total_line="$(tail -n 1 "$SUMMARY_PATH")"
total_pct="$(echo "$total_line" | awk '{print $3}' | tr -d '%')"

if [[ -z "$total_pct" ]]; then
  echo "failed to parse total coverage" >&2
  exit 1
fi

printf "TOTAL_COVERAGE=%.1f\n" "$total_pct"

if [[ "$MODE" == "ci" ]]; then
  if awk -v total="$total_pct" -v threshold="$THRESHOLD" 'BEGIN { exit !(total + 0 < threshold + 0) }'; then
    echo "coverage gate failed: ${total_pct}% < ${THRESHOLD}%"
    exit 1
  fi
fi
