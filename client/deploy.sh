#!/bin/bash

# 德州扑克前端部署脚本
# 使用方法: ./deploy.sh [build|up|down|logs|push]

set -e

IMAGE_NAME="dz-poker-web"
CONTAINER_NAME="dz-poker-web"

# 使用阿里云镜像加速
DOCKER_MIRROR="--registry-mirror=https://registry.cn-hangzhou.aliyuncs.com"

case "$1" in
    build)
        echo "构建镜像..."
        docker build -t ${IMAGE_NAME}:latest .
        echo "构建完成!"
        ;;

    up)
        echo "启动服务..."
        docker compose up -d
        echo "服务已启动，访问 http://localhost:8088"
        ;;

    down)
        echo "停止服务..."
        docker compose down
        echo "服务已停止"
        ;;

    logs)
        docker logs -f ${CONTAINER_NAME}
        ;;

    push)
        if [ -z "$2" ]; then
            echo "用法: ./deploy.sh push <registry>"
            echo "示例: ./deploy.sh push registry.cn-hangzhou.aliyuncs.com/your-namespace"
            exit 1
        fi
        echo "推送镜像到 $2..."
        docker tag ${IMAGE_NAME}:latest $2/${IMAGE_NAME}:latest
        docker push $2/${IMAGE_NAME}:latest
        echo "推送完成!"
        ;;

    *)
        echo "德州扑克前端部署脚本"
        echo ""
        echo "用法: $0 {build|up|down|logs|push}"
        echo ""
        echo "命令说明:"
        echo "  build       构建镜像"
        echo "  up          启动服务"
        echo "  down        停止服务"
        echo "  logs        查看日志"
        echo "  push        推送镜像到仓库"
        echo "              示例: $0 push registry.cn-hangzhou.aliyuncs.com/your-namespace"
        ;;
esac