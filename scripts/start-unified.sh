#!/bin/bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

PID_FILE="/tmp/go-mall-pids.txt"
STATE_DIR="/tmp/go-mall-supervisors"
LOG_DIR="${PROJECT_ROOT}/scripts/logs"
DEPEND_STACK="${PROJECT_ROOT}/construct/depend/docker-compose.yaml"
LOKI_STACK="${PROJECT_ROOT}/infrastructure/docker-compose.yaml"
FRONTEND_DIR="${PROJECT_ROOT}/frontend"
FRONTEND_PORT="3000"
FRONTEND_LOG="${LOG_DIR}/frontend.log"
STRIPE_FORWARD_LOG="${LOG_DIR}/stripe-listen.log"
SUPERVISOR_SCRIPT="${PROJECT_ROOT}/scripts/service-supervisor.sh"
STRIPE_WEBHOOK_URL="http://127.0.0.1:11112/stripe/webhook"

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

DEPENDENCY_CONTAINERS="
go-mall-postgres:5432:Postgres
go-mall-redis:6379:Redis
go-mall-rabbitmq:5672:RabbitMQ
go-mall-etcd:2379:etcd
go-mall-elasticsearch:9200:Elasticsearch
go-mall-gorse:8088:Gorse
"

INFRA_PORTS="2379 5432 5672 6379 9200 8088"
CORE_PORTS="10000 10001 10002 10003 10004 10005 10006 10007 10008 10009 10010 10011 10012 8888"
MANAGED_PORTS="${CORE_PORTS} ${FRONTEND_PORT}"

usage() {
    cat <<'EOF'
Usage:
  ./scripts/start-unified.sh          Start the full Go-mall local environment
  ./scripts/start-unified.sh status   Show dependency/service/frontend status
  ./scripts/start-unified.sh stop     Stop services started by this script
  ./scripts/start-unified.sh restart  Restart the full Go-mall local environment
EOF
}

port_in_use() {
    lsof -Pi :"$1" -sTCP:LISTEN -t >/dev/null 2>&1
}

http_ready() {
    local url="$1"
    curl -fsS "$url" >/dev/null 2>&1
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

wait_for_http() {
    local url="$1"
    local attempts="${2:-30}"
    local sleep_secs="${3:-2}"
    local i
    for ((i = 1; i <= attempts; i++)); do
        if http_ready "$url"; then
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

container_status() {
    docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$1" 2>/dev/null || echo "missing"
}

pid_running() {
    local pid="${1:-}"
    [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null
}

supervisor_pid_path() {
    echo "${STATE_DIR}/$1.supervisor.pid"
}

child_pid_path() {
    echo "${STATE_DIR}/$1.child.pid"
}

read_pid_value() {
    local pid_file="$1"
    if [[ -r "$pid_file" ]]; then
        tr -d '[:space:]' < "$pid_file" 2>/dev/null || true
    fi
}

service_label() {
    echo "dev.go-mall.$1"
}

stripe_cli_authenticated() {
    stripe listen --print-secret >/dev/null 2>&1
}

fetch_stripe_webhook_secret() {
    stripe listen --print-secret 2>/dev/null | tr -d '[:space:]'
}

append_pid() {
    local name="$1"
    local pid="$2"
    local port="$3"
    local tmp_file
    tmp_file="$(mktemp)"
    if [[ -f "$PID_FILE" ]]; then
        grep -v "^${name}:" "$PID_FILE" > "$tmp_file" || true
    fi
    echo "${name}:${pid}:${port}" >> "$tmp_file"
    mv "$tmp_file" "$PID_FILE"
}

print_log_tail() {
    local log_file="$1"
    if [[ -f "$log_file" ]]; then
        tail -n 80 "$log_file" || true
    else
        echo "日志文件不存在: $log_file"
    fi
}

cleanup_old_processes() {
    echo "正在清理旧进程及占用端口..."

    if [[ -f "$PID_FILE" ]]; then
        awk -F: '{print $1 ":" $2}' "$PID_FILE" | while IFS=: read -r name pid; do
            if command -v launchctl >/dev/null 2>&1; then
                launchctl remove "$(service_label "$name")" >/dev/null 2>&1 || true
            fi
            [[ -n "$pid" ]] && kill -TERM "$pid" 2>/dev/null || true
        done
        rm -f "$PID_FILE"
    fi

    if [[ -d "$STATE_DIR" ]]; then
        find "$STATE_DIR" -name '*.pid' -type f -maxdepth 1 -print0 2>/dev/null | while IFS= read -r -d '' pid_file; do
            local pid
            pid="$(read_pid_value "$pid_file")"
            [[ -n "$pid" ]] && kill -TERM "$pid" 2>/dev/null || true
        done
        rm -rf "$STATE_DIR"
    fi

    local port
    for port in $MANAGED_PORTS; do
        local pids
        pids="$(lsof -ti :"$port" 2>/dev/null || true)"
        if [[ -n "$pids" ]]; then
            echo "释放端口 $port (PID: $pids)"
            echo "$pids" | xargs kill -9 2>/dev/null || true
        fi
    done

    (pkill -9 -f "air -c .air.toml" 2>/dev/null || true) >/dev/null 2>&1
}

start_dependencies() {
    echo "启动本地依赖栈..."
    docker compose -f "$DEPEND_STACK" up -d

    local line
    while IFS=: read -r container port _label; do
        [[ -z "$container" ]] && continue
        echo "等待 ${container} 健康..."
        wait_for_container_health "$container" || {
            echo "依赖未就绪: $container"
            docker logs --tail 100 "$container" || true
            exit 1
        }
        wait_for_port "$port" 20 1 || {
            echo "依赖端口未就绪: $port ($container)"
            exit 1
        }
    done <<< "$DEPENDENCY_CONTAINERS"
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
    local schema_ready
    schema_ready="$(docker exec -i go-mall-postgres psql -U root -d mall -tAc \
        "SELECT CASE
            WHEN to_regclass('public.products') IS NOT NULL
             AND to_regclass('public.users') IS NOT NULL
             AND to_regclass('public.inventory') IS NOT NULL
            THEN 1 ELSE 0
         END")"

    if [[ "$schema_ready" == "1" ]]; then
        echo "保留现有 mall 数据库数据，不重建表结构"
    else
        echo "初始化 mall 数据库表结构..."
        docker exec -i go-mall-postgres psql -U root -d mall -f /docker-entrypoint-initdb.d/01-init_all_tables_postgres.sql >/dev/null
    fi
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

service_command() {
    local srv_name="$1"
    local srv_dir="$2"
    local entrypoint="${srv_name}.go"
    local go_bin
    local binary_path

    go_bin="$(command -v go)"
    binary_path="${STATE_DIR}/bin/${srv_name}"

    if [[ ! -f "${PROJECT_ROOT}/${srv_dir}/${entrypoint}" ]]; then
        entrypoint="$(basename "$srv_dir").go"
    fi

    if [[ -f "${PROJECT_ROOT}/${srv_dir}/.air.toml" ]] && command -v air >/dev/null 2>&1; then
        local air_bin
        air_bin="$(command -v air)"
        echo "exec ${air_bin} -c .air.toml"
        return 0
    fi

    echo "mkdir -p ${STATE_DIR}/bin && ${go_bin} build -o ${binary_path} ${entrypoint} && exec ${binary_path}"
}

start_supervised_process() {
    local name="$1"
    local workdir="$2"
    local port="$3"
    local log_file="$4"
    local command="$5"
    local supervisor_pid_file
    local child_pid_file
    local label

    supervisor_pid_file="$(supervisor_pid_path "$name")"
    child_pid_file="$(child_pid_path "$name")"
    label="$(service_label "$name")"

    : > "$log_file"
    if command -v launchctl >/dev/null 2>&1; then
        launchctl remove "$label" >/dev/null 2>&1 || true
        launchctl submit -l "$label" -- "$SUPERVISOR_SCRIPT" "$name" "$workdir" "$log_file" \
            "$supervisor_pid_file" "$child_pid_file" "$command"
    else
        nohup "$SUPERVISOR_SCRIPT" "$name" "$workdir" "$log_file" \
            "$supervisor_pid_file" "$child_pid_file" "$command" >/dev/null 2>&1 < /dev/null &
        local supervisor_pid=$!
        append_pid "$name" "$supervisor_pid" "$port"
        return 0
    fi
    local supervisor_pid
    local attempts=20
    local i
    for ((i = 1; i <= attempts; i++)); do
        supervisor_pid="$(read_pid_value "$supervisor_pid_file")"
        if pid_running "$supervisor_pid"; then
            break
        fi
        sleep 1
    done
    append_pid "$name" "$supervisor_pid" "$port"
}

start_service() {
    local srv_name="$1"
    local srv_info
    srv_info="$(get_service_info "$srv_name")"
    [[ -z "$srv_info" ]] && return 1

    IFS=':' read -r _ srv_dir srv_port _srv_type <<< "$srv_info"

    if port_in_use "$srv_port"; then
        lsof -ti :"$srv_port" | xargs kill -9 2>/dev/null || true
        sleep 1
    fi

    echo "启动 $srv_name (端口 $srv_port)..."
    local log_file="${LOG_DIR}/${srv_name}.log"
    local command
    command="$(service_command "$srv_name" "$srv_dir")"
    if [[ "$srv_name" == "payment" ]]; then
        local webhook_secret="${STRIPE_WEBHOOK_SECRET:-}"
        if [[ -n "$webhook_secret" ]]; then
            command="export STRIPE_WEBHOOK_SECRET='${webhook_secret}'; ${command}"
        fi
    fi
    start_supervised_process "$srv_name" "${PROJECT_ROOT}/${srv_dir}" "$srv_port" "$log_file" "$command"

    if wait_for_port "$srv_port" 20 1; then
        echo "服务就绪: $srv_name ($srv_port)"
    else
        echo "服务未在预期时间内监听: $srv_name"
        print_log_tail "$log_file"
        exit 1
    fi
}

start_stripe_forwarder() {
    if ! command -v stripe >/dev/null 2>&1; then
        echo "未找到 Stripe CLI，请先安装 stripe 后再启动本地支付链路"
        exit 1
    fi
    if ! stripe_cli_authenticated; then
        echo "Stripe CLI 未登录，请先执行 stripe login"
        exit 1
    fi

    local webhook_secret
    webhook_secret="$(fetch_stripe_webhook_secret)"
    if [[ -z "$webhook_secret" ]]; then
        echo "无法获取 Stripe webhook signing secret"
        exit 1
    fi
    export STRIPE_WEBHOOK_SECRET="$webhook_secret"

    echo "启动 stripe listen webhook 转发..."
    start_supervised_process "stripe-listen" "$PROJECT_ROOT" "11112" "$STRIPE_FORWARD_LOG" \
        "exec stripe listen --skip-update --forward-to ${STRIPE_WEBHOOK_URL}"

    sleep 2
    local stripe_supervisor_pid
    stripe_supervisor_pid="$(read_pid_value "$(supervisor_pid_path stripe-listen)")"
    if ! pid_running "$stripe_supervisor_pid"; then
        echo "stripe listen 启动失败"
        print_log_tail "$STRIPE_FORWARD_LOG"
        exit 1
    fi
}

ensure_frontend_dependencies() {
    if [[ -d "${FRONTEND_DIR}/node_modules" ]]; then
        return 0
    fi

    echo "frontend/node_modules 缺失，安装前端依赖..."
    : > "$FRONTEND_LOG"
    if [[ -f "${FRONTEND_DIR}/package-lock.json" ]]; then
        (
            cd "$FRONTEND_DIR"
            npm ci
        ) >> "$FRONTEND_LOG" 2>&1 || {
            echo "前端依赖安装失败，请检查 ${FRONTEND_LOG}"
            print_log_tail "$FRONTEND_LOG"
            exit 1
        }
    else
        (
            cd "$FRONTEND_DIR"
            npm install
        ) >> "$FRONTEND_LOG" 2>&1 || {
            echo "前端依赖安装失败，请检查 ${FRONTEND_LOG}"
            print_log_tail "$FRONTEND_LOG"
            exit 1
        }
    fi
}

start_frontend() {
    ensure_frontend_dependencies

    if port_in_use "$FRONTEND_PORT"; then
        lsof -ti :"$FRONTEND_PORT" | xargs kill -9 2>/dev/null || true
        sleep 1
    fi

    echo "启动 frontend (端口 ${FRONTEND_PORT})..."
    local node_bin
    local vite_entry
    node_bin="$(command -v node)"
    vite_entry="${FRONTEND_DIR}/node_modules/vite/bin/vite.js"
    start_supervised_process "frontend" "$FRONTEND_DIR" "$FRONTEND_PORT" "$FRONTEND_LOG" \
        "exec ${node_bin} ${vite_entry} --host 127.0.0.1 --port ${FRONTEND_PORT}"

    if wait_for_http "http://127.0.0.1:${FRONTEND_PORT}" 30 2; then
        echo "前端就绪: frontend (${FRONTEND_PORT})"
    else
        echo "前端未在预期时间内可访问"
        print_log_tail "$FRONTEND_LOG"
        exit 1
    fi
}

start_loki_stack() {
    if ! docker info >/dev/null 2>&1; then
        echo "Docker 未运行，跳过 Loki/Grafana"
        return
    fi
    if [[ ! -f "$LOKI_STACK" ]]; then
        echo "未找到 Loki/Grafana compose 文件，跳过日志栈"
        return
    fi

    echo "启动 Loki/Grafana..."
    if ! docker compose -f "$LOKI_STACK" up -d loki promtail grafana >/dev/null 2>&1; then
        echo "Loki/Grafana 启动失败，继续启动主联调环境"
    fi
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

scan_frontend() {
    wait_for_http "http://127.0.0.1:${FRONTEND_PORT}" 5 1 || {
        echo "前端未就绪: ${FRONTEND_PORT}"
        exit 1
    }
}

service_status_line() {
    local name="$1"
    local port="$2"
    local log_file="$3"
    local supervisor_pid
    supervisor_pid="$(read_pid_value "$(supervisor_pid_path "$name")")"
    if port_in_use "$port" && pid_running "$supervisor_pid"; then
        printf '  [up]   %-10s port=%-5s log=%s\n' "$name" "$port" "$log_file"
    else
        printf '  [down] %-10s port=%-5s log=%s\n' "$name" "$port" "$log_file"
        return 1
    fi
}

dependency_status_line() {
    local container="$1"
    local port="$2"
    local label="$3"
    local status
    status="$(container_status "$container")"
    if [[ "$status" == "healthy" || "$status" == "running" ]]; then
        printf '  [up]   %-13s port=%-5s container=%s (%s)\n' "$label" "$port" "$container" "$status"
    else
        printf '  [down] %-13s port=%-5s container=%s (%s)\n' "$label" "$port" "$container" "$status"
        return 1
    fi
}

status_all() {
    local overall=0

    echo "Go-mall 本地联调环境状态"
    echo
    echo "Dependencies:"
    local line
    while IFS=: read -r container port label; do
        [[ -z "$container" ]] && continue
        dependency_status_line "$container" "$port" "$label" || overall=1
    done <<< "$DEPENDENCY_CONTAINERS"

    echo
    echo "Core services:"
    while IFS=: read -r name _dir port _type; do
        [[ -z "$name" ]] && continue
        service_status_line "$name" "$port" "${LOG_DIR}/${name}.log" || overall=1
    done <<< "$ALL_SERVICES"
    local frontend_supervisor_pid
    frontend_supervisor_pid="$(read_pid_value "$(supervisor_pid_path frontend)")"
    if http_ready "http://127.0.0.1:${FRONTEND_PORT}" && pid_running "$frontend_supervisor_pid"; then
        printf '  [up]   %-10s url=%s log=%s\n' "frontend" "http://127.0.0.1:${FRONTEND_PORT}" "$FRONTEND_LOG"
    else
        printf '  [down] %-10s url=%s log=%s\n' "frontend" "http://127.0.0.1:${FRONTEND_PORT}" "$FRONTEND_LOG"
        overall=1
    fi

    local stripe_supervisor_pid
    stripe_supervisor_pid="$(read_pid_value "$(supervisor_pid_path stripe-listen)")"
    if pid_running "$stripe_supervisor_pid"; then
        printf '  [up]   %-10s target=%s log=%s\n' "stripe" "$STRIPE_WEBHOOK_URL" "$STRIPE_FORWARD_LOG"
    else
        printf '  [down] %-10s target=%s log=%s\n' "stripe" "$STRIPE_WEBHOOK_URL" "$STRIPE_FORWARD_LOG"
        overall=1
    fi

    echo
    echo "Access URLs:"
    echo "  frontend: http://127.0.0.1:${FRONTEND_PORT}"
    echo "  gateway:  http://127.0.0.1:8888"
    echo "  rabbitmq: http://127.0.0.1:15672"
    echo "  gorse:    http://127.0.0.1:8088"
    echo "  grafana:  http://127.0.0.1:3001"
    echo "  loki:     http://127.0.0.1:3100"
    echo "  stripe webhook: ${STRIPE_WEBHOOK_URL}"
    echo
    echo "Logs:"
    echo "  ${LOG_DIR}"
    echo "  frontend log: ${FRONTEND_LOG}"
    echo "  stripe log:   ${STRIPE_FORWARD_LOG}"

    return "$overall"
}

stop_all() {
    cleanup_old_processes
    docker compose -f "$DEPEND_STACK" down >/dev/null 2>&1 || true
    docker compose -f "$LOKI_STACK" down >/dev/null 2>&1 || true
}

start_all() {
    cleanup_old_processes
    mkdir -p "$STATE_DIR"
    : > "$PID_FILE"
    start_dependencies
    reconcile_postgres
    reconcile_rabbitmq

    for srv in system activity auths audit users inventory product carts coupons order checkout payment admin gateway; do
        start_service "$srv"
    done

    scan_core_ports
    start_stripe_forwarder
    start_frontend
    scan_frontend
    start_loki_stack

    echo
    echo "本地联调环境已启动。"
    status_all
}

mkdir -p "$LOG_DIR"

case "${1:-start}" in
    start)
        start_all
        ;;
    stop)
        stop_all
        ;;
    status)
        status_all
        ;;
    restart)
        stop_all
        start_all
        ;;
    *)
        usage
        exit 1
        ;;
esac
