#!/bin/bash
# ==============================================
# 构建 K8s 部署所需的 Docker 镜像
# 支持 minikube / k3s 环境
# ==============================================

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# 服务列表
SERVICES="auths users product carts order checkout payment inventory audit coupons system activity admin gateway"

# 默认使用 minikube，可通过环境变量切换
KUBE_ENV="${KUBE_ENV:-minikube}"

# 镜像仓库前缀
IMAGE_REGISTRY="${IMAGE_REGISTRY:-minikube}"

# 镜像标签
IMAGE_TAG="${IMAGE_TAG:-latest}"

# 检查 kube 环境
check_kube_env() {
    case "$KUBE_ENV" in
        minikube)
            if ! command -v minikube &> /dev/null; then
                echo "错误: minikube 未安装"
                exit 1
            fi
            echo "使用 minikube Docker 环境"
            eval $(minikube docker-env)
            ;;
        k3d)
            if ! command -v k3d &> /dev/null; then
                echo "错误: k3d 未安装"
                exit 1
            fi
            echo "使用 k3d 镜像仓库"
            # k3d 使用内置 registry
            eval $(k3d node list K3S_REGISTRY)
            ;;
        kind)
            if ! command -v kind &> /dev/null; then
                echo "错误: kind 未安装"
                exit 1
            fi
            echo "使用 kind Docker 环境"
            # kind 加载镜像到集群
            ;;
        *)
            echo "未知环境: $KUBE_ENV，支持: minikube, k3d, kind"
            exit 1
            ;;
    esac
}

# 构建单个服务镜像
build_image() {
    local svc=$1
    local image_name="${IMAGE_REGISTRY}/${svc}:${IMAGE_TAG}"

    echo ""
    echo ">>> 构建 $image_name ..."

    # 进入服务目录
    cd "$PROJECT_ROOT/services/$svc"

    # 检查 Dockerfile 是否存在
    if [ ! -f "Dockerfile" ]; then
        echo "警告: $svc/Dockerfile 不存在，跳过"
        return 1
    fi

    # 构建镜像
    docker build -t "$image_name" .

    if [ $? -eq 0 ]; then
        echo "✓ $image_name 构建成功"
    else
        echo "✗ $image_name 构建失败"
        return 1
    fi
}

# 主流程
main() {
    echo "========================================="
    echo "开始构建 Go-mall K8s 镜像"
    echo "环境: $KUBE_ENV"
    echo "镜像仓库: $IMAGE_REGISTRY"
    echo "标签: $IMAGE_TAG"
    echo "========================================="

    # 检查并配置 kube 环境
    check_kube_env

    # 构建所有服务镜像
    failed=0
    for svc in $SERVICES; do
        if ! build_image "$svc"; then
            failed=1
        fi
    done

    cd "$PROJECT_ROOT"

    echo ""
    echo "========================================="
    if [ $failed -eq 0 ]; then
        echo "所有镜像构建完成!"
    else
        echo "部分镜像构建失败，请检查上述错误"
    fi
    echo "========================================="

    echo ""
    echo "当前镜像列表:"
    docker images | grep go-mall

    return $failed
}

main "$@"