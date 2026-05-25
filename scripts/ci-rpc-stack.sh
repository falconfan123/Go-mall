#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEPEND_STACK="$ROOT_DIR/construct/depend/docker-compose.yaml"
STATE_DIR="${GO_MALL_CI_STATE_DIR:-$ROOT_DIR/.artifacts/ci-rpc-stack}"
PID_FILE="$STATE_DIR/pids.txt"
LOG_DIR="${GO_MALL_CI_LOG_DIR:-$ROOT_DIR/scripts/logs}"
DEPENDENCY_LOG_DIR="${GO_MALL_CI_DEPENDENCY_LOG_DIR:-$ROOT_DIR/.artifacts/dependency-logs}"
EXIT_DIR="$STATE_DIR/exits"
BIN_DIR="$STATE_DIR/bin"
MONITOR_PID_FILE="$STATE_DIR/monitor.pid"
GO_CMD="${GO_CMD:-go}"
GOTOOLCHAIN_VALUE="${GOTOOLCHAIN:-go1.25.8}"

DEPENDENCY_PORTS=(2379 5432 5672 6379 9200 8088)
CORE_PORTS=(10000 10001 10002 10003 10004 10005 10006 10007 10008 10009 10010 10011 10012 8888)
DEPENDENCY_SERVICES=(postgres redis rabbitmq etcd elasticsearch gorse)

SERVICES=(
  "system:services/system:system.go:10010"
  "activity:services/activity:activity.go:10011"
  "auths:services/auths:auths.go:10000"
  "users:services/users:users.go:10001"
  "product:services/product:product.go:10002"
  "carts:services/carts:carts.go:10003"
  "order:services/order:order.go:10004"
  "checkout:services/checkout:checkout.go:10005"
  "payment:services/payment:payment.go:10006"
  "inventory:services/inventory:inventory.go:10007"
  "audit:services/audit:audit.go:10008"
  "coupons:services/coupons:coupons.go:10009"
  "admin:services/admin:admin.go:10012"
  "gateway:services/gateway:gateway.go:8888"
)

mkdir -p "$STATE_DIR" "$LOG_DIR" "$DEPENDENCY_LOG_DIR" "$EXIT_DIR" "$BIN_DIR"

port_in_use() {
  lsof -Pi :"$1" -sTCP:LISTEN -t >/dev/null 2>&1
}

listener_pid() {
  local port="$1"
  lsof -ti :"$port" -sTCP:LISTEN 2>/dev/null | head -n 1
}

wait_for_port() {
  local port="$1"
  local attempts="${2:-60}"
  local sleep_secs="${3:-1}"
  local i

  for ((i = 1; i <= attempts; i++)); do
    if port_in_use "$port"; then
      return 0
    fi
    sleep "$sleep_secs"
  done
  return 1
}

pid_alive() {
  local pid="$1"
  [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null
}

tail_service_log() {
  local name="$1"
  local log_file="$LOG_DIR/$name.log"
  if [[ -f "$log_file" ]]; then
    tail -n 120 "$log_file" >&2 || true
  fi
}

last_service_log_line() {
  local name="$1"
  local log_file="$LOG_DIR/$name.log"
  if [[ -f "$log_file" ]]; then
    tail -n 1 "$log_file" 2>/dev/null || true
  fi
}

wait_for_service_ready() {
  local name="$1"
  local pid="$2"
  local port="$3"
  local attempts="${4:-90}"
  local sleep_secs="${5:-1}"
  local i

  for ((i = 1; i <= attempts; i++)); do
    if ! pid_alive "$pid"; then
      echo "service process exited before listen: $name" >&2
      tail_service_log "$name"
      return 1
    fi
    if port_in_use "$port"; then
      return 0
    fi
    sleep "$sleep_secs"
  done

  echo "service failed to listen: $name" >&2
  tail_service_log "$name"
  return 1
}

assert_service_healthy() {
  local name="$1"
  local pid="$2"
  local port="$3"

  if ! pid_alive "$pid"; then
    echo "service process not running: $name" >&2
    tail_service_log "$name"
    return 1
  fi

  if ! port_in_use "$port"; then
    echo "service port not ready: $name:$port" >&2
    tail_service_log "$name"
    return 1
  fi

  return 0
}

wait_for_container_health() {
  local container="$1"
  local attempts="${2:-60}"
  local sleep_secs="${3:-2}"
  local i

  for ((i = 1; i <= attempts; i++)); do
    local status
    status="$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$container" 2>/dev/null || true)"
    if [[ "$status" == "healthy" || "$status" == "running" ]]; then
      return 0
    fi
    sleep "$sleep_secs"
  done
  return 1
}

cleanup_processes() {
  if [[ -f "$PID_FILE" ]]; then
    while IFS=: read -r _ pid _; do
      if [[ -n "$pid" ]]; then
        kill -TERM -- "-$pid" 2>/dev/null || kill -TERM "$pid" 2>/dev/null || true
      fi
    done <"$PID_FILE"

    sleep 2

    while IFS=: read -r _ pid _; do
      if [[ -n "$pid" ]]; then
        kill -KILL -- "-$pid" 2>/dev/null || kill -KILL "$pid" 2>/dev/null || true
      fi
    done <"$PID_FILE"
  fi

  : >"$PID_FILE"
}

cleanup_ports() {
  local port
  for port in "${CORE_PORTS[@]}"; do
    local pids
    pids="$(lsof -ti :"$port" 2>/dev/null || true)"
    if [[ -n "$pids" ]]; then
      echo "$pids" | xargs kill -KILL 2>/dev/null || true
    fi
  done
}

reset_logs() {
  rm -f "$LOG_DIR"/*.log "$DEPENDENCY_LOG_DIR"/*.log
  rm -f "$EXIT_DIR"/*.status
}

json_escape() {
  local value="$1"
  python3 - <<'PY' "$value"
import json
import sys
print(json.dumps(sys.argv[1]))
PY
}

postgres_exec() {
  local database="$1"
  shift

  if docker exec go-mall-postgres psql -U root -d "$database" -c 'SELECT 1;' >/dev/null 2>&1; then
    docker exec -i go-mall-postgres psql -U root -d "$database" "$@"
    return
  fi
  docker exec -i go-mall-postgres psql -U postgres -d "$database" "$@"
}

start_dependencies() {
  docker compose -f "$DEPEND_STACK" up -d

  local service
  for service in "${DEPENDENCY_SERVICES[@]}"; do
    local container="go-mall-${service}"
    echo "waiting for $container"
    if ! wait_for_container_health "$container"; then
      docker logs "$container" >"$DEPENDENCY_LOG_DIR/${service}.log" 2>&1 || true
      echo "dependency not healthy: $container" >&2
      exit 1
    fi
  done

  local port
  for port in "${DEPENDENCY_PORTS[@]}"; do
    if ! wait_for_port "$port"; then
      echo "dependency port not ready: $port" >&2
      exit 1
    fi
  done
}

reconcile_postgres() {
  postgres_exec postgres <<'SQL'
DO
$do$
BEGIN
   IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'root') THEN
      CREATE ROLE root LOGIN PASSWORD 'fht3825099' SUPERUSER;
   ELSE
      ALTER ROLE root WITH LOGIN PASSWORD 'fht3825099' SUPERUSER;
   END IF;
END
$do$;
SQL

  postgres_exec postgres -tAc "SELECT 1 FROM pg_database WHERE datname='mall'" | grep -q 1 || \
    postgres_exec postgres -c "CREATE DATABASE mall OWNER root;"

  postgres_exec mall -c "GRANT ALL PRIVILEGES ON DATABASE mall TO root;"
  postgres_exec mall -f /docker-entrypoint-initdb.d/01-init_all_tables_postgres.sql >/dev/null
}

reconcile_rabbitmq() {
  docker exec go-mall-rabbitmq rabbitmqctl await_startup >/dev/null
  docker exec go-mall-rabbitmq rabbitmqctl add_vhost / >/dev/null 2>&1 || true
  docker exec go-mall-rabbitmq rabbitmqctl add_user admin admin >/dev/null 2>&1 || \
    docker exec go-mall-rabbitmq rabbitmqctl change_password admin admin >/dev/null
  docker exec go-mall-rabbitmq rabbitmqctl set_permissions -p / admin ".*" ".*" ".*" >/dev/null
  docker exec go-mall-rabbitmq rabbitmqctl set_user_tags admin administrator >/dev/null
}

start_service() {
  local name="$1"
  local rel_dir="$2"
  local entrypoint="$3"
  local port="$4"
  local service_dir="$ROOT_DIR/$rel_dir"
  local log_file="$LOG_DIR/$name.log"
  local exit_file="$EXIT_DIR/$name.status"
  local bin_file="$BIN_DIR/$name"

  if [[ ! -f "$service_dir/$entrypoint" ]]; then
    echo "entrypoint not found for $name: $service_dir/$entrypoint" >&2
    exit 1
  fi

  if port_in_use "$port"; then
    lsof -ti :"$port" | xargs kill -KILL 2>/dev/null || true
    sleep 1
  fi

  echo "starting $name on $port"
  rm -f "$exit_file"
  (
    cd "$service_dir"
    env GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" "$GO_CMD" build -o "$bin_file" "$entrypoint"
  )
  if command -v setsid >/dev/null 2>&1; then
    nohup setsid bash -lc "
      cd '$service_dir'
      '$bin_file' > '$log_file' 2>&1 < /dev/null
      status=\$?
      printf 'exit_code=%s\nfinished_at=%s\n' \"\$status\" \"\$(date -u +%Y-%m-%dT%H:%M:%SZ)\" > '$exit_file'
      exit \"\$status\"
    " >/dev/null 2>&1 < /dev/null &
  else
    nohup bash -lc "
      cd '$service_dir'
      '$bin_file' > '$log_file' 2>&1 < /dev/null
      status=\$?
      printf 'exit_code=%s\nfinished_at=%s\n' \"\$status\" \"\$(date -u +%Y-%m-%dT%H:%M:%SZ)\" > '$exit_file'
      exit \"\$status\"
    " >/dev/null 2>&1 < /dev/null &
  fi
  local launcher_pid=$!

  if ! wait_for_service_ready "$name" "$launcher_pid" "$port" 90 1; then
    exit 1
  fi

  local pid
  pid="$(listener_pid "$port")"
  if [[ -z "$pid" ]]; then
    echo "service listener pid not found: $name:$port" >&2
    tail_service_log "$name"
    exit 1
  fi
  echo "$name:$pid:$port" >>"$PID_FILE"
}

start_services() {
  : >"$PID_FILE"

  local spec
  for spec in "${SERVICES[@]}"; do
    IFS=: read -r name rel_dir entrypoint port <<<"$spec"
    start_service "$name" "$rel_dir" "$entrypoint" "$port"
  done
}

scan_ports() {
  local name pid port
  while IFS=: read -r name pid port; do
    [[ -z "${name:-}" ]] && continue
    if ! assert_service_healthy "$name" "$pid" "$port"; then
      exit 1
    fi
  done <"$PID_FILE"
}

snapshot_dependency_logs() {
  local service
  for service in "${DEPENDENCY_SERVICES[@]}"; do
    docker logs --timestamps "go-mall-${service}" >"$DEPENDENCY_LOG_DIR/${service}.log" 2>&1 || true
  done
}

status() {
  local name pid port
  while IFS=: read -r name pid port; do
    [[ -z "${name:-}" ]] && continue
    if assert_service_healthy "$name" "$pid" "$port"; then
      echo "$name:$port ready"
    else
      return 1
    fi
  done <"$PID_FILE"
}

inspect_stack_json() {
  local first=1
  local name pid port
  printf '['
  while IFS=: read -r name pid port; do
    [[ -z "${name:-}" ]] && continue
    local alive="false"
    local listening="false"
    local exit_code=""
    local finished_at=""
    local exit_file="$EXIT_DIR/$name.status"
    local log_path="$LOG_DIR/$name.log"
    local last_log_line
    last_log_line="$(last_service_log_line "$name")"

    if pid_alive "$pid"; then
      alive="true"
    fi
    if port_in_use "$port"; then
      listening="true"
    fi
    if [[ -f "$exit_file" ]]; then
      exit_code="$(awk -F= '/^exit_code=/{print $2}' "$exit_file" | tail -n 1)"
      finished_at="$(awk -F= '/^finished_at=/{print $2}' "$exit_file" | tail -n 1)"
    fi

    if [[ "$first" -eq 0 ]]; then
      printf ','
    fi
    first=0
    printf '{"name":%s,"pid":%s,"port":%s,"alive":%s,"listening":%s,"exitCode":%s,"finishedAt":%s,"logPath":%s,"lastLogLine":%s}' \
      "$(json_escape "$name")" \
      "${pid:-0}" \
      "${port:-0}" \
      "$alive" \
      "$listening" \
      "$(json_escape "$exit_code")" \
      "$(json_escape "$finished_at")" \
      "$(json_escape "$log_path")" \
      "$(json_escape "$last_log_line")"
  done <"$PID_FILE"
  printf ']'
}

stop_stack() {
  snapshot_dependency_logs || true
  cleanup_processes
  cleanup_ports
  docker compose -f "$DEPEND_STACK" down >/dev/null 2>&1 || true
}

command="${1:-start}"
case "$command" in
  start)
    reset_logs
    cleanup_processes
    cleanup_ports
    start_dependencies
    reconcile_postgres
    reconcile_rabbitmq
    start_services
    scan_ports
    status
    ;;
  stop)
    stop_stack
    ;;
  status)
    status
    ;;
  inspect)
    if [[ "${2:-}" == "--json" ]]; then
      inspect_stack_json
    else
      status
    fi
    ;;
  snapshot-dependency-logs)
    snapshot_dependency_logs
    ;;
  *)
    echo "usage: ci-rpc-stack.sh [start|stop|status|inspect [--json]|snapshot-dependency-logs]" >&2
    exit 1
    ;;
esac
