#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODULES_FILE="$ROOT_DIR/scripts/go_ci_modules.txt"
GO_CMD="${GO_CMD:-go}"
GOTOOLCHAIN_VALUE="${GOTOOLCHAIN:-go1.25.8}"

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

cd "$ROOT_DIR"
GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" "$GO_CMD" work sync

for module in "${modules[@]}"; do
  module_dir="$ROOT_DIR/$module"
  if [[ ! -d "$module_dir" ]]; then
    echo "module directory not found: $module_dir" >&2
    exit 1
  fi

  echo "running go vet in $module"
  (
    cd "$module_dir"
    GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" "$GO_CMD" vet ./...
  )
done
