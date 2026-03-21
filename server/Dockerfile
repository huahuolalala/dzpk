FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装编译依赖
RUN apk add --no-cache gcc musl-dev

# 复制源码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# 运行镜像
FROM alpine:latest

WORKDIR /app

# 安装 CA 证书（用于 HTTPS 请求）
RUN apk add --no-cache ca-certificates

# 复制编译好的二进制文件
COPY --from=builder /app/server .

# 暴露端口
EXPOSE 8080 8081

# 运行
CMD ["./server"]
