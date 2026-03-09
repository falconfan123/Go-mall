#!/bin/bash

# Go-Mall 完整启动脚本
# 同时启动后端服务和前端静态文件服务器

echo ""
echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║                    🛒  Go-Mall 电商平台                         ║"
echo "╠═══════════════════════════════════════════════════════════════╣"
echo "║  基于 Go-Zero 微服务架构的现代电商系统                         ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo ""

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_ROOT"

# 清理旧进程
echo "清理旧进程..."
pkill -f "go run" 2>/dev/null || true
pkill -f "python3.*http.server" 2>/dev/null || true
pkill -f "python.*SimpleHTTPServer" 2>/dev/null || true
sleep 1

# 检查前端目录
if [ ! -d "frontend" ]; then
    echo "错误: 前端目录不存在!"
    exit 1
fi

# 启动前端服务器
echo "启动前端服务器..."
cd frontend
if command -v python3 &> /dev/null; then
    python3 -m http.server 3000 > /tmp/frontend-server.log 2>&1 &
    FRONTEND_PID=$!
elif command -v python &> /dev/null; then
    python -m SimpleHTTPServer 3000 > /tmp/frontend-server.log 2>&1 &
    FRONTEND_PID=$!
else
    echo "警告: 未找到 Python，无法启动前端服务器"
    FRONTEND_PID=""
fi
cd "$PROJECT_ROOT"

# 等待前端启动
sleep 2

# 启动后端服务
echo "启动后端服务..."

# 如果有 cmd/server 目录，用它
if [ -f "cmd/server/main.go" ]; then
    cd cmd/server
    go run main.go -services=auths,users,product,inventory > /tmp/backend-server.log 2>&1 &
    BACKEND_PID=$!
    cd "$PROJECT_ROOT"
elif [ -f "run.go.backup" ]; then
    # 用备份的 run.go
    mv run.go.backup run.go.temp
    go run run.go.temp -services=auths,users,product,inventory > /tmp/backend-server.log 2>&1 &
    BACKEND_PID=$!
    mv run.go.temp run.go.backup
else
    # 直接用我们的脚本启动各个服务
    echo "使用简单模式启动服务..."

    # 启动 auths
    cd services/auths
    go run auths.go > /tmp/auths.log 2>&1 &
    AUTHS_PID=$!
    cd "$PROJECT_ROOT"

    # 启动 users
    sleep 1
    cd services/users
    go run users.go > /tmp/users.log 2>&1 &
    USERS_PID=$!
    cd "$PROJECT_ROOT"

    # 启动 product
    sleep 1
    cd services/product
    go run product.go > /tmp/product.log 2>&1 &
    PRODUCT_PID=$!
    cd "$PROJECT_ROOT"

    # 启动 inventory
    sleep 1
    cd services/inventory
    go run inventory.go > /tmp/inventory.log 2>&1 &
    INVENTORY_PID=$!
    cd "$PROJECT_ROOT"

    # 启动 APIs
    sleep 2
    cd apis/user
    go run user.go > /tmp/user-api.log 2>&1 &
    USER_API_PID=$!
    cd "$PROJECT_ROOT"

    sleep 1
    cd apis/product
    go run product.go > /tmp/product-api.log 2>&1 &
    PRODUCT_API_PID=$!
    cd "$PROJECT_ROOT"
fi

# 保存 PID
cat > /tmp/go-mall-pids.txt <<EOF
FRONTEND_PID=$FRONTEND_PID
BACKEND_PID=$BACKEND_PID
AUTHS_PID=$AUTHS_PID
USERS_PID=$USERS_PID
PRODUCT_PID=$PRODUCT_PID
INVENTORY_PID=$INVENTORY_PID
USER_API_PID=$USER_API_PID
PRODUCT_API_PID=$PRODUCT_API_PID
EOF

# 等待服务启动
echo "等待服务启动..."
sleep 5

echo ""
echo "═══════════════════════════════════════════════════════════════"
if [ -n "$FRONTEND_PID" ]; then
    echo "  🎨  前端界面:  http://localhost:3000"
fi
echo "  🔧  Consul UI:  http://localhost:8500"
echo "  🔍  Elasticsearch: http://localhost:9200"
echo "  🐰  RabbitMQ:  http://localhost:15672 (admin/admin)"
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo "📝  日志文件:"
echo "    前端: /tmp/frontend-server.log"
echo "    后端: /tmp/backend-server.log"
echo "    Auths: /tmp/auths.log"
echo "    Users: /tmp/users.log"
echo "    Product: /tmp/product.log"
echo "    Inventory: /tmp/inventory.log"
echo ""
echo "按 Ctrl+C 停止所有服务"
echo ""

# 清理函数
cleanup() {
    echo ""
    echo "正在停止所有服务..."

    # 读取并杀死所有 PID
    if [ -f /tmp/go-mall-pids.txt ]; then
        source /tmp/go-mall-pids.txt
        for pid in $FRONTEND_PID $BACKEND_PID $AUTHS_PID $USERS_PID $PRODUCT_PID $INVENTORY_PID $USER_API_PID $PRODUCT_API_PID; do
            if [ -n "$pid" ]; then
                kill $pid 2>/dev/null || true
            fi
        done
        rm -f /tmp/go-mall-pids.txt
    fi

    # 兜底清理
    pkill -f "go run" 2>/dev/null || true
    pkill -f "python3.*http.server" 2>/dev/null || true

    echo "所有服务已停止!"
    exit 0
}

# 注册信号处理
trap cleanup SIGINT SIGTERM

# 等待
wait
