#!/bin/bash
set -e

# ==============================================
# Go-Mall 统一启动脚本
# 功能：一键启动所有前后端服务，支持多种启动模式
# 兼容 macOS 自带 bash 3.x 版本
# ==============================================

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_ROOT"

# 全局变量
MODE="core"
CUSTOM_SERVICES=""
SKIP_CHECK=false
DAEMON=false
PID_FILE="/tmp/go-mall-pids.txt"
LOG_DIR="${PROJECT_ROOT}/logs"

# 服务定义 (bash 3.x 兼容，不用关联数组)
ALL_SERVICES="
auths:services/auths:auths.go:10000:rpc
audit:services/audit:audit.go:10008:rpc
users:services/users:users.go:10001:rpc
inventory:services/inventory:inventory.go:10011:rpc
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

# 模式定义
MODES_MINIMAL="auths,users,user-api"
MODES_CORE="auths,audit,users,inventory,product,user-api,product-api,gateway"
MODES_FULL="auths,audit,users,inventory,product,carts,coupons,order,checkout,payment,user-api,product-api,carts-api,order-api,checkout-api,payment-api,coupon-api,flash-api,gateway,frontend"

# 基础设施依赖
INFRA_DEPS="
Consul:8500
MySQL:3306
Redis:6379
Elasticsearch:9200
RabbitMQ:5672
"

# ==============================================
# 工具函数
# ==============================================

# 显示帮助信息
show_help() {
    cat << EOF
Usage: $0 [OPTIONS]

Go-Mall 统一启动脚本

Options:
  -m, --mode <mode>        启动模式: minimal, core, full, custom (默认: core)
  -s, --services <list>    自定义启动服务列表，逗号分隔（仅custom模式有效）
  --no-check               跳过基础设施依赖检查
  -d, --daemon             后台运行
  -h, --help               显示帮助信息
  stop                     停止所有运行中的服务

Examples:
  $0                          # 启动核心服务
  $0 --mode full              # 启动所有服务
  $0 --mode custom --services auths,product,gateway  # 启动指定服务
  $0 stop                     # 停止所有服务
EOF
    exit 0
}

# 获取服务信息
get_service_info() {
    local srv_name=$1
    echo "$ALL_SERVICES" | grep -E "^$srv_name:" | head -1
}

# 获取模式服务列表
get_mode_services() {
    local mode=$1
    case $mode in
        minimal) echo "$MODES_MINIMAL" ;;
        core) echo "$MODES_CORE" ;;
        full) echo "$MODES_FULL" ;;
        *) echo "" ;;
    esac
}

# 检查端口是否被占用
port_in_use() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# 清理旧进程
cleanup_old_processes() {
    echo "清理旧进程..."

    # 杀死PID文件中的进程
    if [ -f "$PID_FILE" ]; then
        source "$PID_FILE"
        for pid in $PIDS; do
            if kill -0 $pid 2>/dev/null; then
                kill $pid 2>/dev/null || true
            fi
        done
        rm -f "$PID_FILE"
    fi

    # 兜底清理
    pkill -f "go run" 2>/dev/null || true
    pkill -f "python3.*http.server 3000" 2>/dev/null || true

    # 等待进程退出
    sleep 2

    # 清理占用的端口
    local all_ports=""
    echo "$ALL_SERVICES" | while read -r line; do
        [ -z "$line" ] && continue
        IFS=':' read -r _ _ _ port _ <<< "$line"
        all_ports="$all_ports,$port"
    done
    all_ports="${all_ports:1}"

    lsof -ti:$all_ports 2>/dev/null | xargs kill -9 2>/dev/null || true

    echo "旧进程清理完成"
}

# 检查基础设施依赖
check_infra() {
    if [ "$SKIP_CHECK" = true ]; then
        echo "跳过基础设施检查"
        return
    fi

    echo "检查基础设施依赖..."
    local all_ok=true

    echo "$INFRA_DEPS" | while read -r line; do
        [ -z "$line" ] && continue
        IFS=':' read -r name port <<< "$line"
        if port_in_use $port; then
            echo "✅ $name (端口 $port) 运行正常"
        else
            echo "❌ $name (端口 $port) 未运行"
            all_ok=false
        fi
    done

    if [ "$all_ok" = false ]; then
        echo ""
        echo "错误: 部分基础设施未启动，请先启动必要的依赖服务"
        echo "提示: 可以使用 docker-compose up -d 启动所有基础设施"
        echo "      或使用 --no-check 参数跳过检查"
        exit 1
    fi

    echo "所有基础设施检查通过"
}

# 启动服务
start_service() {
    local srv_name=$1
    local srv_info=$(get_service_info "$srv_name")

    if [ -z "$srv_info" ]; then
        echo "❌ 未知服务: $srv_name"
        return 1
    fi

    IFS=':' read -r _ srv_dir srv_cmd srv_port srv_type <<< "$srv_info"

    if [ ! -d "$srv_dir" ]; then
        echo "❌ 服务目录不存在: $srv_dir"
        return 1
    fi

    # 检查端口是否被占用
    if port_in_use $srv_port; then
        echo "⚠️  端口 $srv_port 已被占用，跳过启动 $srv_name"
        return 0
    fi

    echo "启动 $srv_name (端口 $srv_port)..."

    local log_file="${LOG_DIR}/${srv_name}.log"
    local pid

    cd "$srv_dir"
    if [[ "$srv_type" == "frontend" ]]; then
        # 前端服务特殊处理
        if command -v python3 &> /dev/null; then
            python3 -m http.server 3000 > "$log_file" 2>&1 &
            pid=$!
        elif command -v python &> /dev/null; then
            python -m SimpleHTTPServer 3000 > "$log_file" 2>&1 &
            pid=$!
        else
            echo "❌ 未找到Python，无法启动前端服务"
            cd "$PROJECT_ROOT"
            return 1
        fi
    else
        # Go服务
        go run "$srv_cmd" > "$log_file" 2>&1 &
        pid=$!
    fi
    cd "$PROJECT_ROOT"

    # 保存PID
    PIDS="$PIDS $pid"
    # 用字符串保存服务PID信息，格式: "name:pid:port name2:pid2:port2"
    SERVICE_PIDS_STR="$SERVICE_PIDS_STR $srv_name:$pid:$srv_port"

    # 等待服务启动
    sleep 2

    # 检查是否启动成功
    if kill -0 $pid 2>/dev/null; then
        echo "✅ $srv_name 启动成功 (PID: $pid, 日志: $log_file)"
        return 0
    else
        echo "❌ $srv_name 启动失败，请查看日志: $log_file"
        return 1
    fi
}

# 优雅停止所有服务
stop_all_services() {
    echo ""
    echo "正在停止所有服务..."

    if [ -f "$PID_FILE" ]; then
        source "$PID_FILE"
        for pid in $PIDS; do
            if kill -0 $pid 2>/dev/null; then
                kill $pid 2>/dev/null || true
            fi
        done
        rm -f "$PID_FILE"
    fi

    # 兜底清理
    pkill -f "go run" 2>/dev/null || true
    pkill -f "python3.*http.server 3000" 2>/dev/null || true

    echo "所有服务已停止"
    exit 0
}

# 显示启动完成信息
show_startup_complete() {
    echo ""
    echo "╔═══════════════════════════════════════════════════════════════╗"
    echo "║                    🛒  Go-Mall 启动成功!                         ║"
    echo "╚═══════════════════════════════════════════════════════════════╝"
    echo ""

    echo "📋  服务列表:"
    for srv_entry in $SERVICE_PIDS_STR; do
        [ -z "$srv_entry" ] && continue
        IFS=':' read -r srv_name pid port <<< "$srv_entry"
        printf "  %-15s 端口: %-6s PID: %-6s\n" "$srv_name" "$port" "$pid"
    done

    echo ""
    echo "🌐  访问地址:"
    echo "$SERVICE_PIDS_STR" | grep -q "frontend:" && echo "  前端界面:        http://localhost:3000"
    echo "$SERVICE_PIDS_STR" | grep -q "gateway:" && echo "  API网关:         http://localhost:8888"
    echo "$SERVICE_PIDS_STR" | grep -q "user-api:" && echo "  用户API:         http://localhost:8001"
    echo "$SERVICE_PIDS_STR" | grep -q "product-api:" && echo "  商品API:         http://localhost:8002"

    echo ""
    echo "🧪  测试URL:"
    echo "$SERVICE_PIDS_STR" | grep -q "gateway:" && echo "  商品列表(网关):  curl http://localhost:8888/douyin/product/list?page=1&size=10"
    echo "$SERVICE_PIDS_STR" | grep -q "product-api:" && echo "  商品列表(API):   curl http://localhost:8002/douyin/product/list?page=1&size=10"

    echo ""
    echo "📝  日志文件:"
    echo "  所有日志位于:    ${LOG_DIR}/"
    echo "  查看实时日志:    tail -f ${LOG_DIR}/<服务名>.log"

    echo ""
    echo "🛠️   管理命令:"
    echo "  停止所有服务:    $0 stop 或按 Ctrl+C"

    echo ""
    echo "💡  提示: 如果服务无法访问，请检查对应日志文件排查问题"
}

# ==============================================
# 主流程
# ==============================================

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -m|--mode)
            MODE="$2"
            shift 2
            ;;
        -s|--services)
            CUSTOM_SERVICES="$2"
            shift 2
            ;;
        --no-check)
            SKIP_CHECK=true
            shift
            ;;
        -d|--daemon)
            DAEMON=true
            shift
            ;;
        -h|--help)
            show_help
            ;;
        stop)
            cleanup_old_processes
            exit 0
            ;;
        *)
            echo "未知参数: $1"
            show_help
            ;;
    esac
done

# 检查模式是否合法
if [[ ! "$MODE" =~ ^(minimal|core|full|custom)$ ]]; then
    echo "错误: 不支持的模式 $MODE"
    echo "支持的模式: minimal, core, full, custom"
    exit 1
fi

# custom模式需要提供服务列表
if [ "$MODE" == "custom" ] && [ -z "$CUSTOM_SERVICES" ]; then
    echo "错误: custom模式需要指定 --services 参数"
    exit 1
fi

# 创建日志目录
mkdir -p "$LOG_DIR"

# 处理停止命令
if [ "$1" == "stop" ]; then
    stop_all_services
fi

# 注册信号处理
trap stop_all_services SIGINT SIGTERM

# 显示启动信息
cat << "EOF"
╔═══════════════════════════════════════════════════════════════╗
║                    🛒  Go-Mall 电商平台                         ║
║           基于 Go-Zero 微服务架构的现代电商系统                 ║
╚═══════════════════════════════════════════════════════════════╝
EOF
echo ""
echo "启动模式: $MODE"
echo ""

# 清理旧进程
cleanup_old_processes

# 检查基础设施
check_infra

# 确定要启动的服务列表
if [ "$MODE" == "custom" ]; then
    START_SERVICES=(${CUSTOM_SERVICES//,/ })
else
    MODE_SERVICES=$(get_mode_services "$MODE")
    START_SERVICES=(${MODE_SERVICES//,/ })
fi

echo ""
echo "即将启动以下服务: ${START_SERVICES[*]}"
echo ""

# 按顺序启动服务
PIDS=""
SERVICE_PIDS_STR=""

for srv in "${START_SERVICES[@]}"; do
    start_service "$srv"
done

# 保存PID到文件
cat > "$PID_FILE" << EOF
PIDS="$PIDS"
EOF

# 等待服务完全启动
echo ""
echo "等待服务完全启动..."
sleep 5

# 显示启动完成信息
show_startup_complete

# 后台运行或前台等待
if [ "$DAEMON" = true ]; then
    echo "服务已在后台运行，PID文件: $PID_FILE"
    exit 0
else
    echo "按 Ctrl+C 停止所有服务"
    wait
fi
