#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_CMD="${GO_CMD:-go}"
GOTOOLCHAIN_VALUE="${GOTOOLCHAIN:-auto}"
REPORT_DIR="${RPC_INTEGRATION_REPORT_DIR:-$ROOT_DIR/.artifacts/rpc-integration-report}"
RAW_LOG="$REPORT_DIR/go-test.jsonl"
STACK_SCRIPT="$ROOT_DIR/scripts/ci-rpc-stack.sh"
LOCAL_ONLY_MODE="${GO_MALL_TEST_LOCAL_ONLY:-0}"

generate_report() {
  local log_path="$1"
  shift

  if "$@" | tee "$log_path" | python3 "$ROOT_DIR/scripts/test_report.py" \
    --suite rpc-integration \
    --title "RPC Integration Test Report" \
    --out-dir "$REPORT_DIR"; then
    return 0
  fi

  local statuses=("${PIPESTATUS[@]}")
  for status in "${statuses[@]}"; do
    if [[ "$status" -ne 0 ]]; then
      return "$status"
    fi
  done
  return 0
}

write_startup_failure_report() {
  local message="$1"

  mkdir -p "$REPORT_DIR"
  printf '%s\n' "RPC Integration Test Report" \
    "packages: 0" \
    "tests: 0" \
    "failures: 1" \
    "skipped: 0" \
    "duration: 0s" \
    "" \
    "failed tests:" \
    "- startup/bootstrap: ${message}" >"$REPORT_DIR/summary.txt"

  cat >"$REPORT_DIR/junit.xml" <<EOF
<?xml version='1.0' encoding='utf-8'?>
<testsuites name="rpc-integration" tests="1" failures="1" skipped="0" time="0">
  <testsuite name="startup" tests="1" failures="1" skipped="0" time="0">
    <testcase classname="startup" name="bootstrap" time="0">
      <failure message="${message}">${message}</failure>
    </testcase>
  </testsuite>
</testsuites>
EOF

  cat >"$REPORT_DIR/index.html" <<EOF
<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>RPC Integration Test Report</title></head>
<body>
  <h1>RPC Integration Test Report</h1>
  <p>startup/bootstrap failed</p>
  <pre>${message}</pre>
</body>
</html>
EOF
}

mkdir -p "$REPORT_DIR"
rm -f "$REPORT_DIR"/{go-test.jsonl,junit.xml,index.html,summary.txt}

run_local_tests() {
  cd "$ROOT_DIR/test/rpc"
  generate_report "$RAW_LOG" env GOWORK=off GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" "$GO_CMD" test -json -count=1 ./...
}

if [[ "$LOCAL_ONLY_MODE" == "1" ]]; then
  if run_local_tests; then
    exit 0
  else
    status=$?
    exit "$status"
  fi
fi

if [[ "${GO_MALL_TEST_LOCAL:-}" == "1" ]]; then
  if "$STACK_SCRIPT" run-local-suite env GO_MALL_TEST_LOCAL_ONLY=1 GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" GO_CMD="$GO_CMD" bash "$0"; then
    exit 0
  else
    status=$?
    if [[ ! -f "$REPORT_DIR/summary.txt" ]]; then
      write_startup_failure_report "local integration bootstrap failed before go test execution"
    fi
    exit "$status"
  fi
fi

NAMESPACE="${GO_MALL_TEST_NAMESPACE:-go-mall}"
RUN_ID="${GO_MALL_TEST_RUN_ID:-$(date +%s)-${RANDOM:-0}}"
TIMEOUT="${GO_MALL_TEST_TIMEOUT:-30s}"
JOB_TIMEOUT="${GO_MALL_TEST_JOB_TIMEOUT:-30m}"
IMAGE="${GO_MALL_TEST_IMAGE:-go-mall-rpc-tests:${RUN_ID}}"
SERVICE_ENDPOINTS="${GO_MALL_TEST_SERVICE_ENDPOINTS:-}"
JOB_NAME="go-mall-rpc-tests-${RUN_ID}"
JOB_TEMPLATE="$ROOT_DIR/k8s/tests/rpc-integration-job.yaml"
RENDERED_JOB="$(mktemp)"

cleanup() {
  rm -f "$RENDERED_JOB"
}

trap cleanup EXIT

docker build -f "$ROOT_DIR/test/rpc/Dockerfile" -t "$IMAGE" "$ROOT_DIR"
if command -v minikube >/dev/null 2>&1; then
  minikube image load "$IMAGE"
fi

sed \
  -e "s|__IMAGE__|$IMAGE|g" \
  -e "s|__RUN_ID__|$RUN_ID|g" \
  -e "s|__NAMESPACE__|$NAMESPACE|g" \
  -e "s|__TIMEOUT__|$TIMEOUT|g" \
  -e "s|__SERVICE_ENDPOINTS__|$SERVICE_ENDPOINTS|g" \
  "$JOB_TEMPLATE" > "$RENDERED_JOB"

kubectl apply -f "$RENDERED_JOB"

if kubectl wait --for=condition=complete "job/$JOB_NAME" -n "$NAMESPACE" --timeout="$JOB_TIMEOUT"; then
  if generate_report "$RAW_LOG" kubectl logs "job/$JOB_NAME" -n "$NAMESPACE"; then
    report_status=0
  else
    report_status=$?
  fi
  if [[ "${GO_MALL_TEST_KEEP_JOB:-}" != "1" ]]; then
    kubectl delete job "$JOB_NAME" -n "$NAMESPACE" --ignore-not-found=true
  fi
  if [[ "$report_status" -ne 0 ]]; then
    exit "$report_status"
  fi
else
  if ! generate_report "$RAW_LOG" kubectl logs "job/$JOB_NAME" -n "$NAMESPACE"; then
    true
  fi
  kubectl describe job "$JOB_NAME" -n "$NAMESPACE" || true
  exit 1
fi
