# Dockerfile for Go API Starter
# 多阶段构建：builder阶段用于编译Go应用
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 设置Go模块代理加速下载
ENV GOPROXY=https://goproxy.cn,direct

# 安装必要的工具
RUN apk add --no-cache git

# 复制源代码
COPY . .

# 下载依赖并编译应用
RUN go build -o main ./cmd/server

# 最终镜像阶段
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 从builder阶段复制编译好的二进制文件
COPY --from=builder /app/main .

# 复制配置文件
COPY configs/config.yaml ./configs/
COPY migrations ./migrations/

# 暴露端口
EXPOSE 8080

# 启动应用
CMD ["./main"]
