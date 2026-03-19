# 构建阶段
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 复制 go.mod
COPY go.mod ./

# 生成 go.sum 并下载依赖
RUN go mod tidy && go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

# 运行阶段
FROM alpine:latest

# 安装 ca-certificates 和 wget（用于健康检查）
RUN apk --no-cache add ca-certificates wget

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/server .

# 复制配置文件目录
COPY --from=builder /app/configs ./configs

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动命令（使用 Docker 专用配置）
CMD ["./server", "-config", "./configs/config.docker.yaml"]
