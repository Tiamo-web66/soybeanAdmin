# Soybean Admin Go

Soybean Admin Go 是一个基于 Go 语言的后端 API 实现，为 [Soybean Admin](https://github.com/soybeanjs/soybean-admin) 前端项目提供完整的后端服务支持。

## 项目简介

本项目是 [Soybean Admin](https://github.com/soybeanjs/soybean-admin) 开源项目的配套后端，使用 Go 语言构建，提供了一套完整的 RESTful API 服务，支持前端项目中的所有功能需求。项目采用了现代化的 Go 应用架构和最佳实践，使用了一系列主流的库和框架。

## 技术栈

- **Web 框架**：[Gin](https://github.com/gin-gonic/gin) v1.10.0 - 高性能 HTTP Web 框架
- **ORM**：[GORM](https://gorm.io/gorm) v1.25.12 - 强大的 ORM 库，用于 PostgreSQL 数据库交互
- **代码生成**：[Gen](https://gorm.io/gen) v0.3.27 - GORM 代码生成器，用于生成数据访问层代码
- **缓存**：[go-redis](https://github.com/redis/go-redis) v9.7.3 - Redis 客户端库
- **工具函数**：[lo](https://github.com/samber/lo) v1.49.1 - 提供函数式编程工具集
- **日志**：lumberjack - 用于日志文件的轮转和管理
- **测试**：[testify](https://github.com/stretchr/testify) v1.10.0 - 测试断言库
- **认证**：[jwt](https://github.com/golang-jwt/jwt) v5.2.2 - JWT 认证库
- **配置**：[go-toml](https://github.com/pelletier/go-toml) v2.2.4 - TOML 配置文件解析

## 项目结构

```
├── api          # API 层，处理 HTTP 请求和响应
├── cmd          # 应用程序入口点和命令行工具
│   ├── main.go  # 主应用入口点
│   └── gen_test # 数据库初始化和数据生成工具
├── config       # 配置文件和配置结构
├── global       # 全局变量，如数据库连接池
├── init         # 初始化代码，如数据库连接
├── middleware   # Gin 中间件
├── model        # 数据模型
│   ├── dto      # 数据传输对象，用于接收前端参数
│   ├── vo       # 视图对象，用于返回给前端
│   ├── po       # 持久化对象，对应数据库实体
│   └── bo       # 业务对象，由 Gen 生成
├── pkg          # 第三方库封装
├── router       # 路由配置
├── service      # 业务逻辑层
├── task         # 定时任务
├── test         # 测试代码
└── utils        # 工具函数
```

## soybean 项目的固定api

- /auth/login
- /auth/getUserInfo
- /auth/refreshToken
- /auth/error
- /route/getUserRoutes
- /route/isRouteExist
- /route/getConstantRoutes
- /route/getReactUserRoutes
- /systemManage/getRoleList
- /systemManage/getUserList
- /systemManage/getAllRoles
- /systemManage/getMenuList
- /systemManage/getMenuList/v2
- /systemManage/getAllPages
- /systemManage/getMenuTree
- /settings/getUserSettings
- /settings/saveUserSettings
- /api/users/login

## 数据响应结构

API 统一返回格式:

```json
{
    "code": "0000", // 0000 表示成功，非 0000 表示失败
    "msg": "",      // 消息说明
    "data": {}      // 实际数据，类型根据接口不同而变化
}
```

## 快速开始

### 环境要求

- Go 1.24+
- PostgreSQL 数据库
- Redis 缓存服务

### 安装

1. 克隆项目

```bash
git clone https://github.com/yourusername/soybean-admin-go.git
cd soybean-admin-go
```

2. 安装依赖

```bash
go mod tidy
```

3. 配置

修改 `config.toml` 文件，设置数据库、Redis 等连接信息：

```toml
# 数据库配置
[database]
host = "localhost"
port = 15432
user = "postgres"
password = "yourpassword"
dbname = "soybean"
sslmode = "disable"
max_idle = 10
max_open = 100

# Redis配置
[redis]
host = "localhost"
port = 16379
password = "yourpassword"
db = 0
```

### 初始化数据库和数据

项目提供了便捷的初始化数据功能，可以通过运行以下命令来初始化系统必要的数据：

```bash
go run cmd/gen_test/main.go
```

该命令将：

1. 创建必要的数据库表结构
2. 创建超级管理员账户 (用户名: Soybean，密码: 123456)
3. 创建基本角色（超级管理员、管理员、普通用户等）
4. 创建系统菜单和权限
5. 建立用户-角色关联关系

> 初始化脚本中还包含了 `cleanupData()` 函数，可以用于清除数据库中的所有数据，默认是注释状态。如需清除数据，可以取消注释后运行。

### 运行

```bash
go run cmd/main.go
```

默认情况下，API 服务将在 `:8080` 端口启动。

## 项目设计思想

本项目采用了典型的三层架构设计：

1. **表示层（API 层）**：位于 `api` 目录，负责接收和响应 HTTP 请求
2. **业务逻辑层（Service 层）**：位于 `service` 目录，实现核心业务逻辑
3. **数据访问层（DAO 层）**：使用 GORM Gen 生成的代码，位于 `model/bo` 目录

同时，项目还采用了以下设计原则：

- **依赖注入**：通过 `global` 包管理全局依赖
- **中间件**：使用 Gin 中间件实现认证、日志、错误处理等横切关注点
- **统一响应**：所有 API 返回统一的响应格式
- **模块化**：按功能划分目录，保持代码的高内聚低耦合

## 主要功能模块

1. **用户认证**：基于 JWT 的用户登录、注册和认证
2. **RBAC 权限管理**：角色、权限、资源的管理
3. **菜单管理**：系统菜单的配置和管理

## 构建与部署

### 构建

```bash
go build -o soybean-admin-go cmd/main.go
```

### Docker 部署

项目提供了 Dockerfile，可以使用 Docker 进行部署：

详细步骤请查看 [Docker 构建与部署](./README.docker.md)

## 开发指南

1. **添加新 API**:
   - 在 `api` 目录下创建新的 API 处理函数
   - 在 `router` 目录下注册路由
   - 在 `service` 目录下实现业务逻辑

2. **添加新数据模型**:
   - 在 `model/po` 目录下定义数据库模型
   - 使用 GORM Gen 生成相应的查询代码
   - 在 `model/dto` 和 `model/vo` 目录下定义请求和响应模型

3. **运行测试**:
   ```bash
   go test -v ./test/...
   ```

## 贡献

欢迎提交 Issues 和 Pull Requests！

## 协议

MIT License - 详见 [LICENSE](LICENSE) 文件 