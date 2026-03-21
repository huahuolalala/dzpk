#!/bin/bash

# 德州扑克服务器一键部署脚本

set -e

echo "=== 德州扑克服务器一键部署 ==="

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo "错误: 需要安装 Docker"
    echo "安装方法: https://docs.docker.com/get-docker/"
    exit 1
fi

# 检查 docker compose (新版命令)
if ! docker compose version &> /dev/null; then
    # 检查旧版 docker-compose
    if ! command -v docker-compose &> /dev/null; then
        echo "错误: 需要安装 Docker Compose"
        echo "安装方法: https://docs.docker.com/compose/install/"
        exit 1
    fi
    COMPOSE_CMD="docker-compose"
else
    COMPOSE_CMD="docker compose"
fi

# 进入 server 目录
cd "$(dirname "$0")"

# 创建数据目录
mkdir -p data

echo ""
echo "1. 构建 Docker 镜像..."
$COMPOSE_CMD build

echo ""
echo "2. 启动服务..."
$COMPOSE_CMD up -d

echo ""
echo "3. 等待服务启动..."
sleep 3

echo ""
echo "4. 检查服务状态..."
if curl -s http://localhost:8080/health > /dev/null; then
    echo "✓ 服务启动成功"
else
    echo "✗ 服务启动失败，请检查日志"
    echo "  $COMPOSE_CMD logs"
    exit 1
fi

echo ""
echo "=== 部署完成 ==="
echo ""
echo "服务地址:"
echo "  WebSocket: ws://localhost:8080/ws"
echo "  HTTP API:  http://localhost:8080"
echo "  健康检查:  http://localhost:8080/health"
echo ""
echo "常用命令:"
echo "  查看日志:   $COMPOSE_CMD logs -f"
echo "  停止服务:   $COMPOSE_CMD down"
echo "  重启服务:   $COMPOSE_CMD restart"
echo "  查看状态:   $COMPOSE_CMD ps"