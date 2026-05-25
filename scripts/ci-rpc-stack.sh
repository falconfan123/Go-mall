#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEPEND_STACK="$ROOT_DIR/construct/depend/docker-compose.yaml"
STATE_DIR="${GO_MALL_CI_STATE_DIR:-$ROOT_DIR/.artifacts/ci-rpc-stack}"
PID_FILE="$STATE_DIR/pids.txt"
LOG_DIR="${GO_MALL_CI_LOG_DIR:-$ROOT_DIR/scripts/logs}"
DEPENDENCY_LOG_DIR="${GO_MALL_CI_DEPENDENCY_LOG_DIR:-$ROOT_DIR/.artifacts/dependency-logs}"
CONFIG_DIR="$STATE_DIR/configs"
GO_CMD="${GO_CMD:-go}"
GOTOOLCHAIN_VALUE="${GOTOOLCHAIN:-go1.25.8}"
ETCD_HOST_LOCAL="${GO_MALL_CI_ETCD_HOST:-127.0.0.1:2379}"
POSTGRES_HOST_LOCAL="${GO_MALL_CI_POSTGRES_HOST:-127.0.0.1}"
REDIS_HOST_LOCAL="${GO_MALL_CI_REDIS_HOST:-127.0.0.1}"
ELASTICSEARCH_HOST_LOCAL="${GO_MALL_CI_ELASTICSEARCH_HOST:-127.0.0.1}"

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

mkdir -p "$STATE_DIR" "$LOG_DIR" "$DEPENDENCY_LOG_DIR" "$CONFIG_DIR"

port_in_use() {
  lsof -Pi :"$1" -sTCP:LISTEN -t >/dev/null 2>&1
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
  rm -f "$CONFIG_DIR"/*.yaml
}

render_service_config() {
  local service_dir="$1"
  local service_name="$2"
  local source_config
  local rendered_config

  source_config="$service_dir/etc/${service_name}.yaml"
  if [[ "$service_name" == "gateway" ]]; then
    source_config="$service_dir/etc/gateway.yaml"
  fi

  if [[ ! -f "$source_config" ]]; then
    echo "$source_config"
    return 0
  fi

  rendered_config="$CONFIG_DIR/${service_name}.yaml"
  sed \
    -e "s/localhost:2379/${ETCD_HOST_LOCAL}/g" \
    -e "s/postgres:\\/\\/root:fht3825099@localhost:/postgres:\\/\\/root:fht3825099@${POSTGRES_HOST_LOCAL}:/g" \
    -e "s/Host: localhost:6379/Host: ${REDIS_HOST_LOCAL}:6379/g" \
    -e "s/Addr: \"http:\\/\\/localhost:9200\"/Addr: \"http:\\/\\/${ELASTICSEARCH_HOST_LOCAL}:9200\"/g" \
    -e "s/http:\\/\\/localhost:9200/http:\\/\\/${ELASTICSEARCH_HOST_LOCAL}:9200/g" \
    "$source_config" >"$rendered_config"
  echo "$rendered_config"
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
  local config_file

  if [[ ! -f "$service_dir/$entrypoint" ]]; then
    echo "entrypoint not found for $name: $service_dir/$entrypoint" >&2
    exit 1
  fi

  config_file="$(render_service_config "$service_dir" "$name")"

  if port_in_use "$port"; then
    lsof -ti :"$port" | xargs kill -KILL 2>/dev/null || true
    sleep 1
  fi

  echo "starting $name on $port"
  if command -v setsid >/dev/null 2>&1; then
    setsid bash -lc "
      cd '$service_dir'
      exec env GOTOOLCHAIN='$GOTOOLCHAIN_VALUE' '$GO_CMD' run '$entrypoint' -f '$config_file' > '$log_file' 2>&1
    " &
  else
    (
      cd "$service_dir"
      exec env GOTOOLCHAIN="$GOTOOLCHAIN_VALUE" "$GO_CMD" run "$entrypoint" -f "$config_file" >"$log_file" 2>&1
    ) &
  fi
  local pid=$!
  echo "$name:$pid:$port" >>"$PID_FILE"

  if ! wait_for_port "$port" 90 1; then
    tail -n 120 "$log_file" >&2 || true
    echo "service failed to listen: $name" >&2
    exit 1
  fi
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
  local port
  for port in "${CORE_PORTS[@]}"; do
    if ! wait_for_port "$port" 10 1; then
      echo "core port not ready: $port" >&2
      exit 1
    fi
  done
}

snapshot_dependency_logs() {
  local service
  for service in "${DEPENDENCY_SERVICES[@]}"; do
    docker logs --timestamps "go-mall-${service}" >"$DEPENDENCY_LOG_DIR/${service}.log" 2>&1 || true
  done
}

status() {
  local spec
  for spec in "${SERVICES[@]}"; do
    IFS=: read -r name _ _ port <<<"$spec"
    if port_in_use "$port"; then
      echo "$name:$port ready"
    else
      echo "$name:$port missing" >&2
      return 1
    fi
  done
}

stop_stack() {
  snapshot_dependency_logs || true
  cleanup_processes
  cleanup_ports
  docker compose -f "$DEPEND_STACK" down >/dev/null 2>&1 || true
}

run_local_suite() {
  local status=0

  trap 'status=$?; stop_stack; exit "$status"' EXIT

  reset_logs
  cleanup_processes
  cleanup_ports
  start_dependencies
  reconcile_postgres
  reconcile_rabbitmq
  start_services
  scan_ports
  status

  "$@"
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
  snapshot-dependency-logs)
    snapshot_dependency_logs
    ;;
  run-local-suite)
    shift
    if [[ $# -eq 0 ]]; then
      echo "usage: ci-rpc-stack.sh run-local-suite <command> [args...]" >&2
      exit 1
    fi
    run_local_suite "$@"
    ;;
  *)
    echo "usage: ci-rpc-stack.sh [start|stop|status|snapshot-dependency-logs|run-local-suite]" >&2
    exit 1
    ;;
esac
