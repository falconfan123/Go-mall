#!/bin/bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

PID_FILE="/tmp/go-mall-pids.txt"
LOG_DIR="${PROJECT_ROOT}/scripts/logs"
DEPEND_STACK="${PROJECT_ROOT}/construct/depend/docker-compose.yaml"
LOKI_STACK="${PROJECT_ROOT}/infrastructure/docker-compose.yaml"

ALL_SERVICES="
system:services/system:10010:rpc
activity:services/activity:10011:rpc
auths:services/auths:10000:rpc
audit:services/audit:10008:rpc
users:services/users:10001:rpc
inventory:services/inventory:10007:rpc
product:services/product:10002:rpc
carts:services/carts:10003:rpc
coupons:services/coupons:10009:rpc
order:services/order:10004:rpc
checkout:services/checkout:10005:rpc
payment:services/payment:10006:rpc
admin:services/admin:10012:rpc
gateway:services/gateway:8888:gateway
"

INFRA_PORTS="2379 5432 5672 6379 9200 8088"
CORE_PORTS="10000 10001 10002 10003 10004 10005 10006 10007 10008 10009 8888"

port_in_use() {
    lsof -Pi :"$1" -sTCP:LISTEN -t >/dev/null 2>&1
}

wait_for_port() {
    local port="$1"
    local attempts="${2:-30}"
    local sleep_secs="${3:-2}"
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
    local attempts="${2:-40}"
    local sleep_secs="${3:-3}"
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

cleanup_old_processes() {
    echo "正在清理旧进程及占用端口..."

    if [[ -f "$PID_FILE" ]]; then
        awk -F: '{print $2}' "$PID_FILE" | while read -r pid; do
            [[ -n "$pid" ]] && kill -TERM -- "-$pid" 2>/dev/null || kill -9 "$pid" 2>/dev/null || true
        done
        rm -f "$PID_FILE"
    fi

    local ports
    ports="$(printf '%s\n%s\n' "$CORE_PORTS" "3000" | xargs)"
    for port in $ports; do
        local pids
        pids="$(lsof -ti :"$port" 2>/dev/null || true)"
        if [[ -n "$pids" ]]; then
            echo "释放端口 $port (PID: $pids)"
            echo "$pids" | xargs kill -9 2>/dev/null || true
        fi
    done

    (pkill -9 -f "air -c .air.toml" 2>/dev/null || true) >/dev/null 2>&1
    (pkill -9 -f "node proxy.js" 2>/dev/null || true) >/dev/null 2>&1
}

start_dependencies() {
    echo "启动本地依赖栈..."
    docker compose -f "$DEPEND_STACK" up -d

    local containers=(
        go-mall-postgres
        go-mall-redis
        go-mall-rabbitmq
        go-mall-etcd
        go-mall-elasticsearch
        go-mall-gorse
    )

    local container
    for container in "${containers[@]}"; do
        echo "等待 $container 健康..."
        wait_for_container_health "$container" || {
            echo "依赖未就绪: $container"
            docker logs --tail 100 "$container" || true
            exit 1
        }
    done

    local port
    for port in $INFRA_PORTS; do
        wait_for_port "$port" 20 1 || {
            echo "依赖端口未就绪: $port"
            exit 1
        }
    done
}

reconcile_postgres() {
    echo "校准 Postgres 本地角色和库..."
    docker exec -i go-mall-postgres psql -U root -d postgres <<'SQL'
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

    docker exec -i go-mall-postgres psql -U root -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='mall'" | grep -q 1 || \
        docker exec -i go-mall-postgres psql -U root -d postgres -c "CREATE DATABASE mall OWNER root;"

    docker exec -i go-mall-postgres psql -U root -d mall -c "GRANT ALL PRIVILEGES ON DATABASE mall TO root;"
    docker exec -i go-mall-postgres psql -U root -d mall -f /docker-entrypoint-initdb.d/01-init_all_tables_postgres.sql >/dev/null
}

reconcile_rabbitmq() {
    echo "校准 RabbitMQ 本地账号权限..."
    docker exec go-mall-rabbitmq rabbitmqctl await_startup >/dev/null
    docker exec go-mall-rabbitmq rabbitmqctl add_vhost / >/dev/null 2>&1 || true
    docker exec go-mall-rabbitmq rabbitmqctl add_user admin admin >/dev/null 2>&1 || \
        docker exec go-mall-rabbitmq rabbitmqctl change_password admin admin >/dev/null
    docker exec go-mall-rabbitmq rabbitmqctl set_permissions -p / admin ".*" ".*" ".*" >/dev/null
    docker exec go-mall-rabbitmq rabbitmqctl set_user_tags admin administrator >/dev/null
}

get_service_info() {
    echo "$ALL_SERVICES" | grep -E "^$1:" | head -1
}

start_service() {
    local srv_name="$1"
    local srv_info
    srv_info="$(get_service_info "$srv_name")"
    [[ -z "$srv_info" ]] && return 1

    IFS=':' read -r _ srv_dir srv_port srv_type <<< "$srv_info"

    if port_in_use "$srv_port"; then
        lsof -ti :"$srv_port" | xargs kill -9 2>/dev/null || true
        sleep 1
    fi

    echo "启动 $srv_name (端口 $srv_port)..."
    local log_file="${LOG_DIR}/${srv_name}.log"
    if command -v setsid >/dev/null 2>&1; then
        setsid bash -lc "
            cd '$PROJECT_ROOT/$srv_dir'
            if [[ -f .air.toml ]]; then
                exec air -c .air.toml > '$log_file' 2>&1
            fi
            entrypoint='${srv_name}.go'
            if [[ ! -f \"\$entrypoint\" ]]; then
                entrypoint='$(basename "$srv_dir").go'
            fi
            exec go run \"\$entrypoint\" > '$log_file' 2>&1
        " &
    else
        (
            cd "$PROJECT_ROOT/$srv_dir"
            if [[ -f .air.toml ]]; then
                exec air -c .air.toml > "$log_file" 2>&1
            fi
            local entrypoint
            entrypoint="${srv_name}.go"
            if [[ ! -f "$entrypoint" ]]; then
                entrypoint="$(basename "$srv_dir").go"
            fi
            exec go run "$entrypoint" > "$log_file" 2>&1
        ) &
    fi
    local pid=$!
    echo "$srv_name:$pid:$srv_port" >> "$PID_FILE"

    if wait_for_port "$srv_port" 20 1; then
        echo "服务就绪: $srv_name ($srv_port)"
    else
        echo "服务未在预期时间内监听: $srv_name"
        tail -n 80 "$log_file" || true
        exit 1
    fi
}

start_loki_stack() {
    if ! docker info >/dev/null 2>&1; then
        echo "Docker 未运行，跳过 Loki/Grafana"
        return
    fi
    echo "启动 Loki/Grafana..."
    docker compose -f "$LOKI_STACK" up -d
}

scan_core_ports() {
    local port
    for port in $CORE_PORTS; do
        wait_for_port "$port" 5 1 || {
            echo "关键端口未监听: $port"
            exit 1
        }
    done
}

stop_all() {
    cleanup_old_processes
    docker compose -f "$DEPEND_STACK" down >/dev/null 2>&1 || true
    docker compose -f "$LOKI_STACK" down >/dev/null 2>&1 || true
}

mkdir -p "$LOG_DIR"

if [[ "${1:-}" == "stop" ]]; then
    stop_all
    exit 0
fi

cleanup_old_processes
start_dependencies
reconcile_postgres
reconcile_rabbitmq

for srv in system activity auths audit users inventory product carts coupons order checkout payment admin gateway; do
    start_service "$srv"
done

scan_core_ports
start_loki_stack

echo "本地联调环境已启动。"
echo "关键依赖端口: $INFRA_PORTS"
echo "核心服务端口: $CORE_PORTS"
