#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODULES_FILE="$ROOT_DIR/scripts/go_ci_modules.txt"
GO_CMD="${GO_CMD:-go}"
GOTOOLCHAIN_VALUE="${GOTOOLCHAIN:-go1.25.11}"
GOVULNCHECK_BIN="${GOVULNCHECK_BIN:-}"

if [[ -z "$GOVULNCHECK_BIN" ]]; then
  if command -v govulncheck >/dev/null 2>&1; then
    GOVULNCHECK_BIN="$(command -v govulncheck)"
  else
    GOVULNCHECK_BIN="$("$GO_CMD" env GOPATH)/bin/govulncheck"
  fi
fi

if [[ ! -x "$GOVULNCHECK_BIN" ]]; then
  echo "govulncheck not found; install with: go install golang.org/x/vuln/cmd/govulncheck@latest" >&2
  exit 1
fi

modules=()
while IFS= read -r line; do
  if [[ -n "$line" ]] && [[ ! "$line" =~ ^[[:space:]]*# ]]; then
    modules+=("$line")
  fi
done <"$MODULES_FILE"

if [[ ${#modules[@]} -eq 0 ]]; then
  echo "no Go modules configured for govulncheck" >&2
  exit 1
fi

for module in "${modules[@]}"; do
  module_dir="$ROOT_DIR/$module"
  if [[ ! -d "$module_dir" ]]; then
    echo "module directory not found: $module_dir" >&2
    exit 1
  fi

  echo "running govulncheck in $module"
  (
    cd "$module_dir"
    GOWORK=off GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" "$GO_CMD" mod tidy
    GOWORK=off GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" "$GOVULNCHECK_BIN" ./...
  )
done
