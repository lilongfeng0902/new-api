#!/bin/bash

# ============================================
# ai-backend-neolink-api 构建脚本
# ============================================

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置变量
MODULE_NAME="ai-backend-neolink-api"
IMAGE_TAG="${1:-latest}"
ENVIRONMENT="${2:-dev}"  # 默认dev环境

# 根据环境设置镜像仓库配置
case "$ENVIRONMENT" in
    "dev")
        REGISTRY="220.181.114.186:80"
        ;;
    "prod")
        REGISTRY="220.181.114.188:80"
        ;;
    *)
        REGISTRY="220.181.114.186:80"
        ;;
esac
IMAGE_NAME="${REGISTRY}/ai-backend-go/${MODULE_NAME}"

# 打印带颜色的消息
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查 Docker 是否可用
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker 未安装或不在 PATH 中"
        exit 1
    fi
    print_info "Docker 检查通过"
}

# Docker 构建
docker_build() {
    print_info "开始 Docker 构建..."
    print_info "镜像名称: ${IMAGE_NAME}:${IMAGE_TAG}"

    docker build -t "${IMAGE_NAME}:${IMAGE_TAG}" .

    print_info "Docker 构建完成"
}

# Docker 推送
docker_push() {
    print_info "推送镜像到仓库..."
    docker push "${IMAGE_NAME}:${IMAGE_TAG}"

    # 如果是 latest 标签，同时推送 latest
    if [ "$IMAGE_TAG" != "latest" ]; then
        print_info "同时推送 latest 标签..."
        docker tag "${IMAGE_NAME}:${IMAGE_TAG}" "${IMAGE_NAME}:latest"
        docker push "${IMAGE_NAME}:latest"
    fi

    print_info "镜像推送完成"
}

# 显示构建结果
show_result() {
    print_info "构建完成！"
    echo ""
    echo "镜像信息："
    echo "  - ${IMAGE_NAME}:${IMAGE_TAG}"
    if [ "$IMAGE_TAG" != "latest" ]; then
        echo "  - ${IMAGE_NAME}:latest"
    fi
    echo ""
    echo "部署命令："
    echo "  cd k8s && ./deploy.sh ${IMAGE_TAG}"
}

# 主函数
main() {
    print_info "开始构建 ${MODULE_NAME}..."
    echo ""

    # 检查依赖
    check_docker
    if [ $? -ne 0 ]; then exit 1; fi

    # 构建流程
    docker_build
    if [ $? -ne 0 ]; then exit 1; fi
    docker_push
    if [ $? -ne 0 ]; then exit 1; fi
    show_result
}

# 执行主函数
main "$@"
