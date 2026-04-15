#!/bin/bash
# ==============================================
# 清理 K8s 资源
# 删除 Go-mall 部署及基础设施
# ==============================================

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# 命名空间
NAMESPACE="${NAMESPACE:-go-mall}"

# 显示帮助
show_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -n, --namespace    命名空间 (默认: go-mall)"
    echo "  --keep-namespace   保留命名空间，只删除资源"
    echo "  --force            强制删除（不确认）"
    echo "  -h, --help         显示帮助"
    echo ""
    echo "示例:"
    echo "  $0                  # 删除所有（包括命名空间）"
    echo "  $0 --keep-namespace # 只删除资源，保留命名空间"
    echo "  $0 --force          # 强制删除，不确认"
}

# 选项
KEEP_NAMESPACE=false
FORCE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        --keep-namespace)
            KEEP_NAMESPACE=true
            shift
            ;;
        --force)
            FORCE=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
done

# 确认删除
if [ "$FORCE" != "true" ]; then
    echo "========================================="
    echo "警告: 此操作将删除以下资源:"
    echo "  - Helm 发布: go-mall"
    echo "  - 所有 Pods, Services, ConfigMaps"
    if [ "$KEEP_NAMESPACE" != "true" ]; then
        echo "  - 命名空间: $NAMESPACE"
    fi
    echo "========================================="
    read -p "确认删除? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "取消删除"
        exit 0
    fi
fi

# 检查依赖
if ! command -v kubectl &> /dev/null; then
    echo "错误: kubectl 未安装"
    exit 1
fi

if ! command -v helm &> /dev/null; then
    echo "错误: helm 未安装"
    exit 1
fi

echo "========================================="
echo "开始清理 K8s 资源"
echo "命名空间: $NAMESPACE"
echo "========================================="

# 删除 Helm 发布
echo ""
echo ">>> 删除 Helm 发布..."
helm uninstall go-mall -n "$NAMESPACE" 2>/dev/null || true

# 删除基础设施
echo ""
echo ">>> 删除基础设施..."
kubectl delete -f k8s/infrastructure/ -n "$NAMESPACE" --ignore-not-found=true 2>/dev/null || true

# 删除 go-mall 资源
echo ""
echo ">>> 删除 go-mall 服务资源..."
kubectl delete deploy -n "$NAMESPACE" --all 2>/dev/null || true
kubectl delete svc -n "$NAMESPACE" --all 2>/dev/null || true
kubectl delete configmaps -n "$NAMESPACE" --all 2>/dev/null || true

# 删除命名空间
if [ "$KEEP_NAMESPACE" != "true" ]; then
    echo ""
    echo ">>> 删除命名空间..."
    kubectl delete namespace "$NAMESPACE" --ignore-not-found=true
fi

echo ""
echo "========================================="
echo "清理完成!"
echo "========================================="