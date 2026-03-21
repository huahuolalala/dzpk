#!/bin/bash

# 德州扑克服务器 Docker 构建脚本

set -e

# 配置
IMAGE_NAME="dz-poker-server"
IMAGE_TAG="${1:-latest}"
FULL_IMAGE="${IMAGE_NAME}:${IMAGE_TAG}"

echo "=== 构建 Docker 镜像 ==="
echo "镜像名称: ${FULL_IMAGE}"
echo ""

# 进入 server 目录
cd "$(dirname "$0")"

# 构建镜像
echo "开始构建..."
docker build -t "${FULL_IMAGE}" .

echo ""
echo "=== 构建完成 ==="
echo "镜像: ${FULL_IMAGE}"
echo ""
echo "运行容器:"
echo "  docker run -d -p 8080:8080 --name dz-poker ${FULL_IMAGE}"
echo ""
echo "推送到仓库:"
echo "  docker tag ${FULL_IMAGE} <registry>/${FULL_IMAGE}"
echo "  docker push <registry>/${FULL_IMAGE}"