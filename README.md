# Blog-BE

Blog-BE 是一个使用 Go + Gin + GORM + PostgreSQL 构建的博客后端服务，提供文章管理、分类和标签查询、用户注册登录、邮件验证码、JWT 鉴权、Redis 验证码缓存等能力。

## 功能

- 用户注册、登录、刷新令牌、获取当前用户信息
- 邮件验证码发送、校验与冷却时间限制
- 文章列表、详情、创建、更新、删除
- 分类、标签列表查询
- JWT access / refresh 双令牌鉴权
- Redis 用于验证码和冷却信息缓存
- Docker Compose 一键启动 PostgreSQL、Redis 和后端服务

## 技术栈

- Go 1.25.3
- Gin
- GORM
- PostgreSQL 15
- Redis 7
- gomail
- bcrypt
- RSA JWT

## 项目结构

```text
blog-BE/
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── init.sql
├── main.go
├── README.md
├── secrets/
│   ├── jwt_private.pem
│   └── jwt_public.pem
└── src/
    ├── config/
    ├── dao/
    ├── handler/
    ├── middleware/
    ├── models/
    │   └── request/
    ├── routers/
    ├── service/
    └── utils/
```

## 运行前提

- Go 1.25 或更高版本
- PostgreSQL 15 或兼容版本
- Redis 7 或兼容版本
- 可用的 SMTP 邮箱账号，用于发送验证码
- RSA JWT 私钥和公钥文件，或通过环境变量直接注入 PEM 内容

## 配置说明

应用会优先读取项目根目录下的 `.env` 文件；如果文件不存在，则继续使用系统环境变量和默认值。

### 数据库

| 变量名 | 说明 | 默认值 |
| --- | --- | --- |
| `DB_HOST` | PostgreSQL 地址 | `localhost` |
| `DB_PORT` | PostgreSQL 端口 | `5432` |
| `DB_HOST_PORT` | Docker Compose 下宿主机映射端口 | `5432` |
| `DB_USER` | 数据库用户名 | `admin` |
| `DB_PASSWORD` | 数据库密码 | `123456` |
| `DB_NAME` | 数据库名称 | `mydb` |

### Redis

| 变量名 | 说明 | 默认值 |
| --- | --- | --- |
| `REDIS_ADDR` | Redis 地址 | `localhost:6379` |
| `REDIS_HOST_PORT` | Docker Compose 下宿主机映射端口 | `6379` |
| `REDIS_PASSWORD` | Redis 密码 | 空 |

### 邮件

| 变量名 | 说明 | 默认值 |
| --- | --- | --- |
| `MAIL_HOST` | SMTP 服务地址 | `smtp.qq.com` |
| `MAIL_PORT` | SMTP 端口 | `465` |
| `MAIL_USERNAME` | 发件邮箱 | 空 |
| `MAIL_PASSWORD` | 邮箱授权码 | 空 |

### 服务

| 变量名 | 说明 | 默认值 |
| --- | --- | --- |
| `APP_PORT` | 服务端口 | `8080` |
| `GIN_MODE` | Gin 运行模式 | `debug` |
| `GIN_TRUSTED_PROXIES` | Gin 信任代理列表 | `127.0.0.1,::1` |

### JWT

| 变量名 | 说明 | 默认值 |
| --- | --- | --- |
| `JWT_PRIVATE_KEY_PATH` | JWT 私钥文件路径 | `/app/secrets/jwt_private.pem` |
| `JWT_PUBLIC_KEY_PATH` | JWT 公钥文件路径 | `/app/secrets/jwt_public.pem` |
| `JWT_ACCESS_TTL` | Access token 有效期 | `15m` |
| `JWT_REFRESH_TTL` | Refresh token 有效期 | `168h` |
| `JWT_SECRETS_HOST_DIR` | Docker Compose 下宿主机 JWT 密钥目录 | `./secrets` |

### CORS

| 变量名 | 说明 | 默认值 |
| --- | --- | --- |
| `CORS_ALLOW_ORIGINS` | 允许跨域来源 | `http://localhost:5173,http://127.0.0.1:5173` |
| `CORS_ALLOW_METHODS` | 允许的 HTTP 方法 | `GET,POST,PUT,PATCH,DELETE,OPTIONS` |
| `CORS_ALLOW_HEADERS` | 允许的请求头 | `Origin,Content-Type,Accept,Authorization,X-Requested-With` |
| `CORS_EXPOSE_HEADERS` | 暴露给前端的响应头 | `Content-Length` |
| `CORS_ALLOW_CREDENTIALS` | 是否允许携带凭证 | `false` |

### 配置文件建议

建议复制 [.env.example](.env.example) 为本地 `.env`，再根据运行方式调整：

- 本地直连运行时，`DB_HOST=localhost`，`REDIS_ADDR=localhost:6379`
- 使用 Docker Compose 时，应用容器会自动注入 `DB_HOST=db` 和 `REDIS_ADDR=redis:6379`
- JWT 密钥既可以挂载文件，也可以直接通过 `JWT_PRIVATE_KEY` 和 `JWT_PUBLIC_KEY` 注入 PEM 内容

## JWT 密钥初始化

先在项目根目录创建 `secrets` 目录，然后生成一对 RSA PEM 文件：

```bash
mkdir secrets
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out secrets/jwt_private.pem
openssl rsa -pubout -in secrets/jwt_private.pem -out secrets/jwt_public.pem
```

Docker Compose 会把 `./secrets` 挂载到容器内的 `/app/secrets`，应用启动时会自动读取对应文件。

## 本地运行

1. 准备配置文件

   ```bash
   copy .env.example .env
   ```

2. 修改 `.env`

   - 设置 PostgreSQL、Redis 和 SMTP 参数
   - 设置 `JWT_PRIVATE_KEY_PATH` 和 `JWT_PUBLIC_KEY_PATH`
   - 或者直接提供 PEM 内容给 `JWT_PRIVATE_KEY` 和 `JWT_PUBLIC_KEY`

3. 安装依赖并启动

   ```bash
   go mod download
   go run main.go
   ```

服务默认监听 `http://localhost:8080`。

## Docker 运行

项目提供 `docker-compose.yml`，会同时启动 PostgreSQL、Redis 和后端服务。

```bash
docker compose up -d --build
```

常用命令：

```bash
docker compose ps
docker compose logs -f app
docker compose restart app
```

说明：

- `db` 服务使用 `.env` 中的数据库配置
- `redis` 服务使用 `.env` 中的 Redis 配置
- `app` 服务会显式注入数据库、邮件、Redis 和 JWT 相关环境变量
- `app` 服务会读取挂载到 `/app/secrets` 的 JWT PEM 文件
- 数据分别持久化到 `postgres_data` 和 `redis_data` 卷

## API 概览

当前路由前缀为 `/api/v1`。所有响应都使用统一结构：

```json
{
  "message": "...",
  "data": {},
  "requestId": "...",
  "code": "..."
}
```

### 公共接口

- `GET /api/v1/categories`
- `GET /api/v1/tags`
- `GET /api/v1/articles`
- `GET /api/v1/articles/:id`

### 认证接口

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `GET /api/v1/auth/me`
- `POST /api/v1/auth/sendcode`
- `POST /api/v1/auth/verify-email`
- `POST /api/v1/auth/register`

### 文章管理接口

- `POST /api/v1/articles`
- `PUT /api/v1/articles/:id`
- `DELETE /api/v1/articles/:id`

## 接口说明

### 登录

- `POST /api/v1/auth/login`
- 请求体：`email`、`password`
- 返回：用户信息和 access / refresh token

### 刷新令牌

- `POST /api/v1/auth/refresh`
- 请求体：`refresh_token`，或通过 `Authorization: Bearer <token>` 传入

### 当前用户

- `GET /api/v1/auth/me`
- 请求头：`Authorization: Bearer <access_token>`

### 发送验证码

- `POST /api/v1/auth/sendcode`
- 请求体：`email`
- 仅允许未注册邮箱发送
- 同一邮箱会有发送冷却时间

### 校验验证码

- `POST /api/v1/auth/verify-email`
- 请求体：`email`、`code`
- 校验成功后会返回短 TTL、一次性的 `registration_token`

### 注册

- `POST /api/v1/auth/register`
- 请求体：`username`、`email`、`password`、`registration_token`
- 兼容过渡期内也接受旧字段 `code`
- 仅传 `code` 时，沿用旧流程校验验证码；传 `registration_token` 时走新流程消费注册令牌
- 注册前会原子消费 `registration_token`，随后使用 bcrypt 哈希保存密码

### 文章列表

- `GET /api/v1/articles`
- 查询参数：`page`、`page_size`、`status`
- `status` 支持 `published`、`draft`、`archived`
- 默认仅返回已发布文章
- 已登录用户可以查看自己的未发布文章，管理员可以查看更多状态的文章

### 文章详情

- `GET /api/v1/articles/:id`
- 已发布文章可直接访问
- 未发布文章仅作者或管理员可访问

### 创建文章

- `POST /api/v1/articles`
- 需要 JWT 认证
- 请求体：`title`、`content`、`summary`、`cover_image`、`status`
- `status` 可选，默认 `draft`

### 更新文章

- `PUT /api/v1/articles/:id`
- 需要 JWT 认证
- 请求体：`title`、`content`、`summary`、`cover_image`、`status`
- 以上字段均为可选，提交哪些字段就更新哪些字段
- 仅作者或管理员可修改

### 删除文章

- `DELETE /api/v1/articles/:id`
- 需要 JWT 认证
- 仅作者或管理员可删除

## 数据模型概览

- `users`：用户、邮箱、密码哈希和角色信息
- `articles`：文章主体、作者、分类、标签、状态、浏览量
- `categories`：分类，支持父子分类和排序
- `tags`：标签，支持颜色和 slug
- `article_tags`：文章与标签的多对多关联

启动时会执行 GORM 迁移，自动创建或更新上述表结构。

## 常见问题

- 如果本地启动报 Redis 连接失败，先确认 `REDIS_ADDR` 和 `REDIS_PASSWORD` 是否和实际环境一致。
- 如果邮件发送失败，通常是 SMTP 主机、端口、授权码或发件邮箱配置错误。
- 如果 JWT 鉴权失败，检查私钥和公钥路径是否正确，或者 PEM 内容是否损坏。
- 如果 Docker 容器内读不到 JWT 文件，确认 `./secrets` 已挂载到容器内的 `/app/secrets`。

## 开发提示

- `main.go` 启动时会先加载配置，再初始化 Redis、数据库和路由。
- `src/utils/response.go` 定义了统一响应格式，前端对接时可以直接按 `message/data/code/requestId` 处理。
- `src/middleware/cors.go` 会读取 CORS 相关环境变量，适合前端本地开发和部署环境切换。
