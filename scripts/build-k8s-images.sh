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

# 镜像仓库前缀。默认与 Helm values 中的镜像仓库命名保持一致。
IMAGE_REGISTRY="${IMAGE_REGISTRY:-go-mall}"

# 镜像标签
IMAGE_TAG="${IMAGE_TAG:-latest}"

# 目标架构。默认自动从集群节点探测。
TARGET_ARCH="${TARGET_ARCH:-}"

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

detect_target_arch() {
    if [ -n "$TARGET_ARCH" ]; then
        return
    fi

    TARGET_ARCH="$(kubectl get nodes -o jsonpath='{.items[0].status.nodeInfo.architecture}' 2>/dev/null || true)"
    case "$TARGET_ARCH" in
        amd64|arm64)
            ;;
        aarch64)
            TARGET_ARCH="arm64"
            ;;
        x86_64)
            TARGET_ARCH="amd64"
            ;;
        *)
            TARGET_ARCH="amd64"
            ;;
    esac
}

resolve_base_image() {
    local svc=$1
    local candidates=(
        "${IMAGE_REGISTRY}/${svc}:latest"
        "${IMAGE_REGISTRY}/${svc}:v2"
        "${IMAGE_REGISTRY}/${svc}:fixed"
    )

    for image in "${candidates[@]}"; do
        if docker image inspect "$image" >/dev/null 2>&1; then
            echo "$image"
            return 0
        fi
    done

    return 1
}

# 构建单个服务镜像
build_image() {
    local svc=$1
    local image_name="${IMAGE_REGISTRY}/${svc}:${IMAGE_TAG}"
    local build_dir="/tmp/go-mall-image-build/${svc}"
    local base_image

    echo ""
    echo ">>> 构建 $image_name ..."

    if ! base_image="$(resolve_base_image "$svc")"; then
        echo "✗ 未找到 $svc 的本地基础镜像，无法离线重建"
        return 1
    fi

    rm -rf "$build_dir"
    mkdir -p "$build_dir/etc"

    (
        cd "$PROJECT_ROOT/services/$svc" && \
        TMPDIR=/tmp \
        GOCACHE="$PROJECT_ROOT/.gocache" \
        CGO_ENABLED=0 \
        GOOS=linux \
        GOARCH="$TARGET_ARCH" \
        go build -o "$build_dir/$svc" .
    )

    cp -R "$PROJECT_ROOT/services/$svc/etc/." "$build_dir/etc/"

    cat > "$build_dir/Dockerfile" <<EOF
FROM $base_image
COPY $svc /app/$svc
COPY etc/ /app/etc/
EOF

    docker build -t "$image_name" "$build_dir"

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
    detect_target_arch
    echo "目标架构: $TARGET_ARCH"

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
