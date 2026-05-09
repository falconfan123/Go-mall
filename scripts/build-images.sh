#!/bin/bash
# 兼容旧入口，统一委托给新的 K8s 构建脚本
set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
exec "$PROJECT_ROOT/scripts/build-k8s-images.sh"
