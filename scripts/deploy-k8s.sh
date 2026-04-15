#!/bin/bash
# ==============================================
# 部署 Go-mall 到 K8s (K3s/Minikube/Kind)
# 使用 Helm Chart
# ==============================================

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Helm chart 路径
CHART_PATH="k8s/helm/go-mall"

# 命名空间
NAMESPACE="${NAMESPACE:-go-mall}"

# 部署环境
ENVIRONMENT="${ENVIRONMENT:-dev}"

# 镜像仓库
IMAGE_REGISTRY="${IMAGE_REGISTRY:-minikube}"

# 显示帮助
show_help() {
    echo "用法: $0 [命令] [选项]"
    echo ""
    echo "命令:"
    echo "  deploy     部署到 K8s"
    echo "  dry-run   模拟部署 (helm template)"
    echo "  status    查看部署状态"
    echo "  logs      查看网关日志"
    echo "  delete    删除部署"
    echo ""
    echo "选项:"
    echo "  -n, --namespace     命名空间 (默认: go-mall)"
    echo "  -e, --env          环境 (dev/prod, 默认: dev)"
    echo "  -r, --registry    镜像仓库 (默认: minikube)"
    echo "  -h, --help        显示帮助"
    echo ""
    echo "示例:"
    echo "  $0 deploy                          # 部署到默认命名空间"
    echo "  $0 deploy -n go-mall -e prod     # 生产环境部署"
    echo "  $0 dry-run                        # 模拟部署"
    echo "  $0 delete                         # 删除部署"
}

# 检查依赖
check_dependencies() {
    if ! command -v kubectl &> /dev/null; then
        echo "错误: kubectl 未安装"
        exit 1
    fi

    if ! command -v helm &> /dev/null; then
        echo "错误: helm 未安装"
        exit 1
    fi
}

# 部署
deploy() {
    echo "========================================="
    echo "部署 Go-mall 到 K8s"
    echo "命名空间: $NAMESPACE"
    echo "环境: $ENVIRONMENT"
    echo "镜像仓库: $IMAGE_REGISTRY"
    echo "========================================="

    # 创建命名空间
    echo ""
    echo ">>> 创建命名空间..."
    kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

    # 部署基础设施
    echo ""
    echo ">>> 部署基础设施 (Consul, Redis, Postgres)..."
    kubectl apply -f k8s/infrastructure/ -n "$NAMESPACE"

    # 等待基础设施就绪
    echo ""
    echo ">>> 等待基础设施就绪..."
    kubectl wait --for=condition=available deployment/consul -n "$NAMESPACE" --timeout=60s || true
    kubectl wait --for=condition=available deployment/redis -n "$NAMESPACE" --timeout=60s || true
    kubectl wait --for=condition=available deployment/postgres -n "$NAMESPACE" --timeout=60s || true

    # 使用 Helm 部署
    echo ""
    echo ">>> 使用 Helm 部署微服务..."
    helm upgrade --install go-mall "$CHART_PATH" \
        --namespace "$NAMESPACE" \
        --set global.namespace="$NAMESPACE" \
        --set global.imageRegistry="$IMAGE_REGISTRY" \
        --set gateway.image="$IMAGE_REGISTRY/gateway:latest" \
        --wait \
        --timeout 5m

    echo ""
    echo "========================================="
    echo "部署完成!"
    echo "========================================="
    echo ""
    echo "查看服务状态:"
    echo "  kubectl get pods -n $NAMESPACE"
    echo ""
    echo "查看网关日志:"
    echo "  kubectl logs -n $NAMESPACE -l app=gateway"
    echo ""
    echo "访问网关:"
    echo "  kubectl port-forward -n $NAMESPACE svc/gateway 8888:8888"
}

# 模拟部署
dry_run() {
    echo "========================================="
    echo "模拟部署 (Helm Template)"
    echo "========================================="

    helm template go-mall "$CHART_PATH" \
        --namespace "$NAMESPACE" \
        --set global.namespace="$NAMESPACE" \
        --set global.imageRegistry="$IMAGE_REGISTRY"
}

# 查看状态
status() {
    echo "========================================="
    echo "部署状态 - $NAMESPACE"
    echo "========================================="

    echo ""
    echo ">>> Pods:"
    kubectl get pods -n "$NAMESPACE" -o wide

    echo ""
    echo ">>> Services:"
    kubectl get svc -n "$NAMESPACE"

    echo ""
    echo ">>> ConfigMaps:"
    kubectl get configmaps -n "$NAMESPACE"

    echo ""
    echo ">>> Deployments:"
    kubectl get deployments -n "$NAMESPACE"
}

# 查看日志
logs() {
    echo "========================================="
    echo "网关日志"
    echo "========================================="

    kubectl logs -n "$NAMESPACE" -l app=gateway --tail=100 -f
}

# 删除部署
delete() {
    echo "========================================="
    echo "删除 Go-mall 部署"
    echo "命名空间: $NAMESPACE"
    echo "========================================="

    # Helm 删除
    echo ""
    echo ">>> 删除 Helm 发布..."
    helm uninstall go-mall -n "$NAMESPACE" || true

    # 删除基础设施
    echo ""
    echo ">>> 删除基础设施..."
    kubectl delete -f k8s/infrastructure/ -n "$NAMESPACE" --ignore-not-found=true

    # 删除命名空间
    echo ""
    echo ">>> 删除命名空间..."
    kubectl delete namespace "$NAMESPACE" --ignore-not-found=true

    echo ""
    echo "========================================="
    echo "删除完成!"
    echo "========================================="
}

# 解析参数
COMMAND=""
while [[ $# -gt 0 ]]; do
    case $1 in
        deploy|dry-run|status|logs|delete)
            COMMAND="$1"
            shift
            ;;
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -r|--registry)
            IMAGE_REGISTRY="$2"
            shift 2
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

# 检查依赖
check_dependencies

# 执行命令
case "$COMMAND" in
    deploy)
        deploy
        ;;
    dry-run)
        dry_run
        ;;
    status)
        status
        ;;
    logs)
        logs
        ;;
    delete)
        delete
        ;;
    *)
        show_help
        exit 1
        ;;
esac