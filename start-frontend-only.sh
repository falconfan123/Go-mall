#!/bin/bash

# Go-Mall 前端启动脚本
# 仅启动前端服务器，用于预览前端开发和演示

echo ""
echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║                    🎨  Go-Mall 前端服务器                       ║"
echo "╠═══════════════════════════════════════════════════════════════╣"
echo "║  提供前端静态文件服务                                   ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo ""

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_ROOT"

# 检查前端目录
if [ ! -d "frontend" ]; then
    echo "❌ 错误: 前端目录不存在!"
    exit 1
fi

# 清理旧进程
echo "🧹 清理旧进程..."
pkill -f "python3.*http.server" 2>/dev/null || true
pkill -f "python.*SimpleHTTPServer" 2>/dev/null || true
sleep 1

# 启动前端服务器
cd frontend
echo "🚀 启动前端服务器..."

if command -v python3 &> /dev/null; then
    echo "   使用 Python 3 启动..."
    python3 -m http.server 3000
elif command -v python &> /dev/null; then
    echo "   使用 Python 2 启动..."
    python -m SimpleHTTPServer 3000
else
    echo "❌ 错误: 未找到 Python!"
    echo ""
    echo "请安装 Python 或使用其他 HTTP 服务器："
    echo "  - Node.js: npx serve -p 3000"
    echo "  - Node.js: npm install -g serve && serve -p 3000"
    echo "  - 使用 IDE 内置服务器"
    exit 1
fi
