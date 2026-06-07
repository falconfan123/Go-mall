#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STACK_SCRIPT="$ROOT_DIR/scripts/ci-rpc-stack.sh"
GO_CMD="${GO_CMD:-go}"
GOTOOLCHAIN_VALUE="${GOTOOLCHAIN:-go1.25.10}"

run_local_tests() {
  cd "$ROOT_DIR/test/rpc"
  env GOWORK=off GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" "$GO_CMD" test -count=1 -v ./scenarios/gateway_smoke
}

if [[ "${GO_MALL_TEST_LOCAL_ONLY:-0}" == "1" ]]; then
  run_local_tests
  exit 0
fi

if [[ "${GO_MALL_TEST_LOCAL:-}" == "1" ]]; then
  "$STACK_SCRIPT" run-local-suite env GO_MALL_TEST_LOCAL_ONLY=1 GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" GO_CMD="$GO_CMD" bash "$0"
  exit 0
fi

echo "GO_MALL_TEST_LOCAL=1 is required for gateway smoke tests" >&2
exit 1
