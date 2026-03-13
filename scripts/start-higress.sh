#!/bin/bash
# Higress 启动脚本
# 使用方法: ./start-higress.sh

set -e

echo "=== Go-mall Higress 网关启动脚本 ==="

# 确保 docker 网络存在
echo "检查 Docker 网络..."
if ! docker network inspect go-mall >/dev/null 2>&1; then
    echo "创建 go-mall 网络..."
    docker network create go-mall
else
    echo "go-mall 网络已存在"
fi

# 启动 Higress
echo "启动 Higress..."
cd "$(dirname "$0")"
docker-compose -f configs/docker-compose.higress.yml up -d

echo ""
echo "=== 启动完成 ==="
echo ""
echo "Higress 网关:     http://localhost:8888"
echo "Higress UI:       http://localhost:8889"
echo ""
echo "下一步: 访问 http://localhost:8889 配置路由"
echo "  - 在 UI 中添加 'McpBridge' 配置后端服务"
echo "  - 添加 'Http2Rpc' 配置 HTTP 到 gRPC 的路由"
echo ""
