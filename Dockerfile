# ===================================
# Dockerfile
# 使用本地交叉编译的二进制文件
# 快速构建 Docker 镜像（无需编译）
# ===================================

FROM alpine:latest

# 安装运行时依赖（最小化）
RUN apk --no-cache add ca-certificates tzdata wget

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户运行应用
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# 直接复制预编译的二进制文件（本地交叉编译的产物）
COPY server .

# 复制生产配置
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
