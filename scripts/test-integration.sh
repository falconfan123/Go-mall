#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_CMD="${GO_CMD:-go}"
GOTOOLCHAIN_VALUE="${GOTOOLCHAIN:-go1.25.8}"
REPORT_DIR="${RPC_INTEGRATION_REPORT_DIR:-$ROOT_DIR/.artifacts/rpc-integration-report}"
RAW_LOG="$REPORT_DIR/go-test.jsonl"

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

mkdir -p "$REPORT_DIR"
rm -f "$REPORT_DIR"/{go-test.jsonl,junit.xml,index.html,summary.txt}

if [[ "${GO_MALL_TEST_LOCAL:-}" == "1" ]]; then
  cd "$ROOT_DIR/test/rpc"
  MONITOR_LOG="$REPORT_DIR/stack-monitor.log"
  rm -f "$MONITOR_LOG"
  monitor_pid=""

  env GO_MALL_TEST_LOCAL=1 GOWORK=off GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" "$GO_CMD" run ./cmd/readycheck -mode wait

  cleanup_monitor() {
    if [[ -n "${monitor_pid:-}" ]] && kill -0 "$monitor_pid" 2>/dev/null; then
      kill "$monitor_pid" 2>/dev/null || true
      wait "$monitor_pid" 2>/dev/null || true
    fi
  }

  env GO_MALL_TEST_LOCAL=1 GO_MALL_TEST_TIMEOUT="${GO_MALL_TEST_TIMEOUT:-30m}" GOWORK=off GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" \
    "$GO_CMD" run ./cmd/readycheck -mode monitor -timeout "${GO_MALL_TEST_TIMEOUT:-30m}" >"$MONITOR_LOG" 2>&1 &
  monitor_pid=$!
  trap cleanup_monitor EXIT

  if generate_report "$RAW_LOG" env GO_MALL_TEST_LOCAL=1 GOWORK=off GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" "$GO_CMD" test -json -count=1 ./...; then
    cleanup_monitor
    exit 0
  else
    status=$?
    if kill -0 "$monitor_pid" 2>/dev/null; then
      cleanup_monitor
    else
      if ! wait "$monitor_pid"; then
        echo "stack monitor detected failure:" >&2
        cat "$MONITOR_LOG" >&2 || true
        status=1
      fi
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
