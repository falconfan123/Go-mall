#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODULES_FILE="$ROOT_DIR/scripts/go_ci_modules.txt"
GO_CMD="${GO_CMD:-go}"
GO_VERSION="$(cat "$ROOT_DIR/.go-version" | tr -d '[:space:]')"
GOTOOLCHAIN_VALUE="${GOTOOLCHAIN:-go${GO_VERSION}}"

modules=()
while IFS= read -r line; do
  if [[ -n "$line" ]] && [[ ! "$line" =~ ^[[:space:]]*# ]]; then
    modules+=("$line")
  fi
done <"$MODULES_FILE"

if [[ ${#modules[@]} -eq 0 ]]; then
  echo "no Go modules configured for vet" >&2
  exit 1
fi

for module in "${modules[@]}"; do
  module_dir="$ROOT_DIR/$module"
  if [[ ! -d "$module_dir" ]]; then
    echo "module directory not found: $module_dir" >&2
    exit 1
  fi

  echo "running go vet in $module"
  (
    cd "$module_dir"
    GOWORK=off GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" "$GO_CMD" mod tidy
    GOWORK=off GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" "$GO_CMD" vet ./...
  )
done
