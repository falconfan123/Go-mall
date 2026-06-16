#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PACKAGES_FILE="$ROOT_DIR/scripts/test_unit_packages.txt"
GO_CMD="${GO_CMD:-go}"
GOTOOLCHAIN_VALUE="${GOTOOLCHAIN:-auto}"
REPORT_DIR="${UNIT_REPORT_DIR:-$ROOT_DIR/.artifacts/unit-report}"
RAW_LOG="$REPORT_DIR/go-test.jsonl"

packages=()
while IFS= read -r line; do
  if [[ -n "$line" ]] && [[ ! "$line" =~ ^[[:space:]]*# ]]; then
    packages+=("$line")
  fi
done <"$PACKAGES_FILE"

if [[ ${#packages[@]} -eq 0 ]]; then
  echo "no unit test packages configured" >&2
  exit 1
fi

mkdir -p "$REPORT_DIR"
rm -f "$REPORT_DIR"/{go-test.jsonl,junit.xml,index.html,summary.txt}

cd "$ROOT_DIR"
if GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" "$GO_CMD" test -json -race -count=1 "${packages[@]}" \
  | tee "$RAW_LOG" \
  | python3 "$ROOT_DIR/scripts/test_report.py" \
    --suite unit \
    --title "Unit Test Report" \
    --out-dir "$REPORT_DIR"; then
  exit 0
fi

statuses=("${PIPESTATUS[@]}")
for status in "${statuses[@]}"; do
  if [[ "$status" -ne 0 ]]; then
    exit "$status"
  fi
done
