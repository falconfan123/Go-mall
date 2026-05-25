#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TARGETS_FILE="$ROOT_DIR/scripts/mock_targets.txt"

mockgen_bin="${MOCKGEN_BIN:-}"
if [[ -z "$mockgen_bin" ]]; then
  if command -v mockgen >/dev/null 2>&1; then
    mockgen_bin="$(command -v mockgen)"
  else
    mockgen_bin="$(go env GOPATH)/bin/mockgen"
  fi
fi

if [[ ! -x "$mockgen_bin" ]]; then
  echo "mockgen not found; install with: go install go.uber.org/mock/mockgen@v0.6.0" >&2
  exit 1
fi

while read -r source_file package_path interface_name destination; do
  if [[ -z "${source_file}" ]] || [[ "${source_file}" == \#* ]]; then
    continue
  fi

  mkdir -p "$(dirname "$ROOT_DIR/$destination")"
  "$mockgen_bin" \
    -source "$ROOT_DIR/$source_file" \
    -package testmock \
    -destination "$ROOT_DIR/$destination"
done <"$TARGETS_FILE"
