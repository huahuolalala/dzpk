#!/bin/bash

# 德州扑克服务器部署脚本

set -e

echo "=== 德州扑克服务器部署 ==="

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo "错误: 需要安装 Docker"
    exit 1
fi

# 检查 docker-compose
if ! command -v docker-compose &> /dev/null; then
    echo "错误: 需要安装 docker-compose"
    exit 1
fi

echo "1. 构建 Docker 镜像..."
docker-compose build

echo "2. 启动服务..."
docker-compose up -d

echo "3. 等待服务启动..."
sleep 2

echo "4. 检查服务状态..."
curl -s http://localhost:8081/health && echo " ✓ 服务启动成功"

echo ""
echo "=== 部署完成 ==="
echo "WebSocket: ws://localhost:8080"
echo "HTTP: http://localhost:8081"
echo ""
echo "停止服务: docker-compose down"
echo "查看日志: docker-compose logs -f"
