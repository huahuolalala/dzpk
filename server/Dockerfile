# 多阶段构建

# 阶段1: 构建
FROM golang:1.23-alpine AS builder

# 安装构建依赖 (sqlite3 需要 CGO)
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build

# 复制 go.mod 和 go.sum 先下载依赖（利用缓存）
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建 (启用 CGO)
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server

# 阶段2: 运行
FROM alpine:3.19

# 安装运行时依赖
RUN apk add --no-cache ca-certificates sqlite-libs tzdata

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/server .
# 复制初始数据库（如果有）
COPY --from=builder /build/dz-poker.db* ./

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/health || exit 1

# 运行
CMD ["./server"]