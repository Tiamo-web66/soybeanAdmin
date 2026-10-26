# 第一阶段：构建应用
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/soybean-admin-go ./cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/init-data ./cmd/gen_test/main.go

# 第二阶段：运行应用
FROM alpine:latest

# 安装必要的工具
RUN apk add --no-cache postgresql-client redis bash curl

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的应用
COPY --from=builder /app/soybean-admin-go /app/
COPY --from=builder /app/init-data /app/

# 复制配置文件
COPY config.docker.toml /app/config.toml

# 创建启动脚本
COPY docker-entrypoint.sh /app/
RUN chmod +x /app/docker-entrypoint.sh

# 暴露应用端口
EXPOSE 8080

# 设置启动命令
ENTRYPOINT ["/app/docker-entrypoint.sh"]