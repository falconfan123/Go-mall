#!/bin/bash
# ==============================================
# Go-Mall 统一启动脚本 (增强版 - 强制清理端口)
# ==============================================

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# 全局变量
MODE="core"
PID_FILE="/tmp/go-mall-pids.txt"
LOG_DIR="${PROJECT_ROOT}/scripts/logs"

# 服务定义
ALL_SERVICES="
auths:services/auths:auths.go:10000:rpc
audit:services/audit:audit.go:10008:rpc
users:services/users:users.go:10001:rpc
inventory:services/inventory:inventory.go:10007:rpc
product:services/product:product.go:10002:rpc
carts:services/carts:carts.go:10003:rpc
coupons:services/coupons:coupons.go:10009:rpc
order:services/order:order.go:10004:rpc
checkout:services/checkout:checkout.go:10005:rpc
payment:services/payment:payment.go:10006:rpc
user-api:apis/user:user.go:8001:api
product-api:apis/product:product.go:8002:api
carts-api:apis/carts:carts.go:8003:api
order-api:apis/order:order.go:8004:api
checkout-api:apis/checkout:checkout.go:8005:api
payment-api:apis/payment:payment.go:8006:api
coupon-api:apis/coupon:coupon.go:8009:api
flash-api:apis/flash_sale:flash.go:8008:api
gateway:services/gateway:gateway.go:8888:gateway
frontend:frontend:python3 -m http.server 3000:3000:frontend
"

# 基础设施依赖
INFRA_DEPS="Consul:8500 MySQL:3306 Redis:6379 Elasticsearch:9200 RabbitMQ:5672"

# --- 核心修改部分：强化清理函数 ---
cleanup_old_processes() {
    echo "正在强制清理旧进程及占用端口..."

    # 1. 尝试通过 PID 文件杀掉进程
    if [ -f "$PID_FILE" ]; then
        # 兼容旧版读取方式
        PIDS_TO_KILL=$(grep -o '[0-9]\+' "$PID_FILE" || true)
        for pid in $PIDS_TO_KILL; do
            kill -9 $pid 2>/dev/null || true
        done
        rm -f "$PID_FILE"
    fi

    # 2. 批量清理所有服务定义的端口 (核心修复)
    # 提取 ALL_SERVICES 中所有的端口号
    local ports=$(echo "$ALL_SERVICES" | awk -F':' '{print $4}' | xargs)

    for port in $ports; do
        if [ -n "$port" ]; then
            # 查找占用该端口的 PID
            local port_pids=$(lsof -ti :$port 2>/dev/null)
            if [ -n "$port_pids" ]; then
                echo "释放端口 $port (PID: $port_pids)..."
                echo "$port_pids" | xargs kill -9 2>/dev/null || true
            fi
        fi
    done

    # 3. 兜底清理关键字进程
    pkill -9 -f "go run" 2>/dev/null || true
    pkill -9 -f "python3.*http.server 3000" 2>/dev/null || true

    sleep 1
    echo "旧进程清理完成"
}

# 检查端口是否被占用
port_in_use() {
    lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null 2>&1
}

# 获取服务信息
get_service_info() {
    echo "$ALL_SERVICES" | grep -E "^$1:" | head -1
}

# 启动服务逻辑 (略作优化以保证稳定性)
start_service() {
    local srv_name=$1
    local srv_info=$(get_service_info "$srv_name")
    [ -z "$srv_info" ] && return 1

    IFS=':' read -r _ srv_dir srv_cmd srv_port srv_type <<< "$srv_info"

    # 二次检查端口，防止清理失败
    if port_in_use $srv_port; then
        echo "⚠️  端口 $srv_port 仍被占用，尝试最后一次强制清理..."
        lsof -ti :$srv_port | xargs kill -9 2>/dev/null || true
        sleep 1
    fi

    echo "启动 $srv_name (端口 $srv_port)..."
    local log_file="${LOG_DIR}/${srv_name}.log"

    cd "$srv_dir"
    if [[ "$srv_type" == "frontend" ]]; then
        python3 -m http.server 3000 > "$log_file" 2>&1 &
    else
        go run "$srv_cmd" > "$log_file" 2>&1 &
    fi
    local pid=$!
    cd "$PROJECT_ROOT"

    echo "$srv_name:$pid:$srv_port" >> "$PID_FILE"

    # 快速检查
    sleep 1.5
    if kill -0 $pid 2>/dev/null; then
        echo "✅ $srv_name 启动成功 (PID: $pid)"
    else
        echo "❌ $srv_name 启动失败，查看: $log_file"
    fi
}

# --- 以下为脚本主逻辑 (简化调用) ---

# 处理停止逻辑
if [ "$1" == "stop" ]; then
    cleanup_old_processes
    exit 0
fi

mkdir -p "$LOG_DIR"
cleanup_old_processes

# 默认启动核心服务
START_SERVICES="auths audit users inventory product carts order checkout payment gateway"
for srv in $START_SERVICES; do
    start_service "$srv"
done

echo "所有服务处理完毕。"