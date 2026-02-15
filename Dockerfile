# ===================================
# 多阶段构建 Dockerfile
# Stage 1: 构建阶段
# Stage 2: 运行阶段（轻量镜像）
# ===================================

# Stage 1: 构建
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app

# 设置 Go 代理加速（国内推荐使用 goproxy.cn）
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

# 在 Alpine 中编译需要 build-base，但不需要 git 和 ca-certificates
RUN apk add --no-cache build-base

# 复制 go.mod 和 go.sum（利用Docker缓存层）
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译（CGO_ENABLED=0 生成静态链接的二进制文件）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/server \
    ./cmd/server/main.go

# ===================================
# Stage 2: 运行
FROM alpine:latest

# 安装运行时依赖（最小化）
RUN apk --no-cache add ca-certificates tzdata wget

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户运行应用
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/server .

COPY config.prod.yaml ./config.yaml

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 切换到非root用户
USER appuser

# 启动应用
CMD ["./server"]
