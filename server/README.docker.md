# Soybean Admin Go - Docker 部署指南

本文档提供了使用 Docker 和 Docker Compose 部署 Soybean Admin Go 项目的详细说明。

## 前提条件

- 安装 [Docker](https://docs.docker.com/get-docker/)
- 安装 [Docker Compose](https://docs.docker.com/compose/install/)

## 快速开始

### 使用 Docker Compose 启动所有服务

```bash
# 在项目根目录下执行
docker-compose up -d
```

这将启动以下服务：
- PostgreSQL 数据库 (端口: 15432)
- Redis 服务 (端口: 16379)
- Soybean Admin Go 应用 (端口: 8080)

### 查看日志

```bash
# 查看所有服务的日志
docker-compose logs -f

# 只查看应用的日志
docker-compose logs -f app
```

### 停止服务

```bash
docker-compose down
```

## 配置说明

### 数据持久化

- PostgreSQL 数据存储在名为 `postgres_data` 的 Docker 卷中
- Redis 数据存储在名为 `redis_data` 的 Docker 卷中
- 应用日志存储在宿主机的 `./logs` 目录中

### 环境配置

项目使用 `config.docker.toml` 作为 Docker 环境的配置文件，该文件在构建 Docker 镜像时会被复制到容器中的 `/app/config.toml`。

如需修改配置，请编辑 `config.docker.toml` 文件，然后重新构建镜像：

```bash
docker-compose build app
docker-compose up -d app
```

## 自定义构建

如果需要单独构建应用镜像，可以执行：

```bash
docker build -t soybean-admin-go .
```

## 故障排除

### 数据库连接问题

如果应用无法连接到数据库，请检查：

1. PostgreSQL 服务是否正常运行：`docker-compose ps postgres`
2. 数据库配置是否正确：检查 `config.docker.toml` 中的数据库配置

### Redis 连接问题

如果应用无法连接到 Redis，请检查：

1. Redis 服务是否正常运行：`docker-compose ps redis`
2. Redis 配置是否正确：检查 `config.docker.toml` 中的 Redis 配置