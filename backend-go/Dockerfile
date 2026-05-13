# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git build-base

# 复制go.mod和go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o tomato-backend ./cmd

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 安装CA证书
RUN apk --no-cache add ca-certificates tzdata

# 从构建阶段复制二进制文件
COPY --from=builder /app/tomato-backend .

# 复制配置文件
COPY config/config.yaml ./config/

# 创建上传目录
RUN mkdir -p uploads logs

# 暴露端口
EXPOSE 8090

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8090/health || exit 1

# 运行应用
CMD ["./tomato-backend"]
