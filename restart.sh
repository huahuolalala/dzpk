#!/bin/bash

# 进入服务器目录
cd "$(dirname "$0")/server"

# 杀掉旧进程
echo "Stopping old server..."
lsof -ti:8080 | xargs kill -9 2>/dev/null
lsof -ti:8081 | xargs kill -9 2>/dev/null

# 等待端口释放
sleep 1

# 编译并启动
echo "Building and starting server..."
go build -o server ./cmd/server && ./server