# 阶段1: 构建
FROM node:20-alpine AS builder

# 使用国内 npm 镜像
RUN npm config set registry https://registry.npmmirror.com

WORKDIR /app

# 复制依赖文件
COPY package*.json ./

# 安装依赖
RUN npm ci --registry=https://registry.npmmirror.com

# 复制源代码
COPY . .

# 构建
RUN npm run build

# 阶段2: 生产镜像
FROM nginx:alpine

# 使用阿里云镜像源（如果需要安装包）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制 nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 暴露端口
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]