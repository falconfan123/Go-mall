#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SERVICES_FILE="$ROOT_DIR/scripts/build_services.txt"
GO_CMD="${GO_CMD:-go}"
GOTOOLCHAIN_VALUE="${GOTOOLCHAIN:-go1.25.10}"

load_services() {
  local services=()
  while IFS= read -r line; do
    if [[ -n "$line" ]] && [[ ! "$line" =~ ^[[:space:]]*# ]]; then
      services+=("$line")
    fi
  done <"$SERVICES_FILE"
  printf '%s\n' "${services[@]}"
}

usage() {
  cat <<'EOF'
usage: go-ci-build.sh [--all | <service> ...]

Builds the CI service whitelist defined in scripts/build_services.txt.
EOF
}

configured_services=()
while IFS= read -r service; do
  configured_services+=("$service")
done < <(load_services)
if [[ ${#configured_services[@]} -eq 0 ]]; then
  echo "no build services configured" >&2
  exit 1
fi

targets=()
if [[ $# -eq 0 ]] || [[ "$1" == "--all" ]]; then
  targets=("${configured_services[@]}")
else
  for arg in "$@"; do
    if [[ "$arg" == "--help" || "$arg" == "-h" ]]; then
      usage
      exit 0
    fi
    targets+=("$arg")
  done
fi

cd "$ROOT_DIR"
GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" "$GO_CMD" work sync

# Disable work mode when building individual services to prevent module resolution issues
export GOWORK=disable

for service in "${targets[@]}"; do
  if [[ ! " ${configured_services[*]} " =~ [[:space:]]${service}[[:space:]] ]]; then
    echo "service is not in build whitelist: $service" >&2
    exit 1
  fi

  service_dir="$ROOT_DIR/services/$service"
  if [[ ! -d "$service_dir" ]]; then
    echo "service directory not found: $service_dir" >&2
    exit 1
  fi

  echo "building $service"
  (
    cd "$service_dir"
    GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" "$GO_CMD" build ./...
  )
done
