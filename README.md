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

当前 `web/.env.test` 默认使用 Apifox Mock 地址。如果要让前端请求本地后端，请将其中的 `VITE_SERVICE_BASE_URL` 改为：

```env
VITE_SERVICE_BASE_URL=http://localhost:8080/api/v1
```

修改环境文件后需要重启前端开发进程。

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

## 配置位置

- 前端环境变量：`web/.env`、`web/.env.test`、`web/.env.prod`
- 后端本地配置：`server/config.toml`
- 后端 Docker 配置：`server/config.docker.toml`
- 基础设施编排：`server/docker-compose.yml`
