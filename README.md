# soybeanAdmin

这是一个前后端分离的 Soybean Admin 项目，前端和后端分别维护在 `web`、`server` 目录中。

## 目录结构

```text
soybeanAdmin/
├── web/             # Vue 3 + TypeScript 前端，独立 Git 仓库
├── server/          # Go 后端 API，独立 Git 仓库
├── scripts/
│   └── dev.ps1      # 同时启动前后端的开发脚本
└── README.md
```

`web` 和 `server` 是两个独立的 Git 仓库，同时作为本仓库的 Git 子模块记录。提交代码、查看历史或处理分支时，请先进入对应目录；根目录用于统一组织项目和提供开发脚本。

首次克隆整个项目时，建议递归初始化子模块：

```powershell
git clone --recurse-submodules https://github.com/Tiamo-web66/soybeanAdmin.git
```

如果已经克隆了根仓库，可以补充初始化子模块：

```powershell
git submodule update --init --recursive
```

## 环境要求

- Node.js `>=20.19.0`
- pnpm `>=10.5.0`（项目固定版本为 `pnpm@12.10.1`）
- Go `1.24+`
- Docker Desktop（可选；用于自动启动 PostgreSQL 和 Redis）

后端默认连接以下本地服务：

- PostgreSQL：`localhost:15432`
- Redis：`localhost:16379`

## 首次安装

在项目根目录执行：

```powershell
Set-Location .\web
pnpm install

Set-Location ..\server
go mod download all
```

如果使用 Docker 提供数据库和缓存，可以先启动基础设施：

```powershell
Set-Location .\server
docker compose up -d postgres redis
```

首次使用空数据库时，可在 `server` 目录执行初始化脚本：

```powershell
go run cmd/gen_test/main.go
```

## 统一启动

回到项目根目录后，使用一个命令同时启动 Go 后端和 Vue 前端：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\dev.ps1
```

如果还需要自动启动 PostgreSQL 和 Redis：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\dev.ps1 -WithInfra
```

脚本会分别在 `server`、`web` 目录中启动进程，因此后端可以正确读取 `server/config.toml`。按 `Ctrl+C` 会停止脚本启动的前端和后端进程；脚本不会自动停止 Docker 服务。

前端启动脚本会直接调用 `web/node_modules/.bin/vpr.cmd`，运行 `web/package.json` 中的 `dev` 脚本；如果提示找不到该文件，请先完成前端依赖安装。

启动后访问：

- 前端开发地址：<http://localhost:9527>
- 后端 API 根地址：<http://localhost:8080/api/v1>

`web/.env.test` 已配置为本地后端地址。开发模式下，前端请求会先发送到 Vite 代理，再转发到后端：

```env
VITE_SERVICE_BASE_URL=http://localhost:8080/api/v1
```

请求路径会经过 `/proxy-default` 代理前缀，Vite 会移除该前缀，因此前端的 `/auth/login` 最终请求为后端的 `/api/v1/auth/login`。修改环境文件后需要重启前端开发进程。

`server/config.toml` 是本地 Go 进程使用的配置，数据库和 Redis 默认配置如下：

```toml
[database]
host = "localhost"
port = 15432
user = "postgres"
password = "wangbadan"
dbname = "soybean"
sslmode = "disable"

[redis]
host = "localhost"
port = 16379
password = "wangbadan"
db = 0
```

使用 `docker compose up -d postgres redis` 时，Compose 会创建 `soybean` 数据库；如果之前已经创建过 `postgres_data` 卷，需手动创建一次：

```powershell
docker compose exec postgres createdb -U postgres soybean
```

## 单独启动

只启动前端：

```powershell
Set-Location .\web
pnpm dev
```

只启动后端：

```powershell
Set-Location .\server
go run cmd/main.go
```

前端构建预览端口为 `9725`，可在 `web` 目录执行：

```powershell
pnpm build:test
pnpm preview
```

## 生产环境部署

生产环境建议将 Go 后端、PostgreSQL 和 Redis 运行在 Docker Compose 中，再使用 Nginx 托管前端静态文件，并把 `/api/` 反向代理到后端。这样前后端使用同一个域名，不需要额外处理浏览器跨域问题。

### 1. 部署后端和基础设施

在部署服务器执行：

```bash
cd server
docker compose up -d --build postgres redis app
docker compose ps
docker compose logs -f app
```

生产容器使用 `server/config.docker.toml`。修改数据库密码、Redis 密码、JWT 密钥或其他配置后，需要重新构建应用镜像：

```bash
docker compose build app
docker compose up -d app
```

Compose 会持久化 `postgres_data`、`redis_data` 两个数据卷，并将后端 API 暴露在服务器的 `8080` 端口。首次启动时，应用会等待 PostgreSQL 和 Redis 就绪，创建 `soybean` 数据库，并执行数据库表迁移和初始化数据脚本。默认初始化账号为 `Soybean / 123456`，上线后请立即修改密码和密钥。

生产服务器应限制 `15432` 和 `16379` 端口只允许内网访问，或者在生产 Compose 配置中删除 PostgreSQL、Redis 的 `ports` 映射；前端和外部客户端只需要访问 `80/443` 以及后端 API 入口。

### 2. 配置并构建前端

生产环境的 API 地址会在构建时写入前端文件。编辑 `web/.env.prod`，将 Mock 地址替换为实际地址：

```env
# 使用独立 API 域名
VITE_SERVICE_BASE_URL=https://api.example.com/api/v1
```

独立 API 域名需要后端配置 CORS。当前项目推荐使用下面的同域名 Nginx 代理，因此通常将 API 地址设置为相对路径：

如果前端和后端使用同一个域名，推荐使用相对路径：

```env
VITE_SERVICE_BASE_URL=/api/v1
```

然后构建生产文件：

```bash
cd web
pnpm install --frozen-lockfile
pnpm build
```

构建产物位于 `web/dist`。修改 `.env.prod` 后必须重新执行 `pnpm build`，运行中的静态文件不会自动更新。

### 3. 使用 Nginx 托管前端

将 `web/dist` 的内容复制到 Nginx 的站点目录，例如 `/var/www/soybean-admin`，并配置 history 路由回退和 API 代理：

```nginx
server {
    listen 80;
    server_name admin.example.com;

    root /var/www/soybean-admin;
    index index.html;

    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

检查并重载 Nginx：

```bash
nginx -t
nginx -s reload
```

部署完成后访问 `http://admin.example.com`。如果使用 HTTPS，请将 `VITE_SERVICE_BASE_URL` 和 Nginx 配置中的域名替换为 HTTPS 地址，并配置证书。

## 配置位置

- 前端环境变量：`web/.env`、`web/.env.test`、`web/.env.prod`
- 后端本地配置：`server/config.toml`
- 后端 Docker 配置：`server/config.docker.toml`
- 基础设施编排：`server/docker-compose.yml`
