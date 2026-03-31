# Blog-BE

一个使用 Go + Gin + GORM + PostgreSQL 构建的博客后端服务，提供用户注册、登录、邮件验证码、JWT 鉴权等基础能力。

## 功能

- 用户注册、登录、刷新令牌
- 邮件验证码发送与校验
- JWT access / refresh 双令牌鉴权
- 密码使用 bcrypt 哈希存储与校验
- Docker Compose 一键启动 PostgreSQL + 后端服务

## 技术栈

- Go 1.25.3
- Gin
- GORM
- PostgreSQL 15
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
└── src/
    ├── config/
    ├── dao/
    ├── handler/
    ├── middleware/
    ├── models/
    ├── routers/
    ├── service/
    └── utils/
```

## 环境要求

- Go 1.25 或更高版本
- PostgreSQL 15 或兼容版本
- 可用的 SMTP 邮箱账号，用于发送验证码
- RSA JWT 私钥和公钥文件

## 环境变量

项目通过 `.env` 文件读取配置。Docker Compose 会把数据库、邮件和 JWT 相关变量显式注入到 `app` 容器中。

| 变量名 | 说明 | 默认值 |
| --- | --- | --- |
| `DB_HOST` | PostgreSQL 地址 | `localhost` |
| `DB_PORT` | PostgreSQL 端口 | `5432` |
| `DB_HOST_PORT` | Docker Compose 下宿主机映射端口 | `5432` |
| `DB_USER` | 数据库用户名 | `admin` |
| `DB_PASSWORD` | 数据库密码 | `123456` |
| `DB_NAME` | 数据库名称 | `mydb` |
| `MAIL_HOST` | SMTP 服务地址 | `smtp.qq.com` |
| `MAIL_PORT` | SMTP 端口 | `465` |
| `MAIL_USERNAME` | 发件邮箱 | 空 |
| `MAIL_PASSWORD` | 邮箱授权码 | 空 |
| `APP_PORT` | 服务端口 | `8080` |
| `GIN_MODE` | Gin 运行模式 | `release` |
| `GIN_TRUSTED_PROXIES` | Gin 信任代理列表 | `127.0.0.1,::1` |
| `JWT_PRIVATE_KEY_PATH` | JWT 私钥文件路径 | `/app/secrets/jwt_private.pem` |
| `JWT_PUBLIC_KEY_PATH` | JWT 公钥文件路径 | `/app/secrets/jwt_public.pem` |
| `JWT_ACCESS_TTL` | Access token 有效期 | `15m` |
| `JWT_REFRESH_TTL` | Refresh token 有效期 | `168h` |
| `JWT_SECRETS_HOST_DIR` | 宿主机 JWT 密钥目录 | `./secrets` |

## 初始化 JWT 密钥

先在项目根目录创建 `secrets` 目录，然后生成一对 RSA PEM 文件：

```bash
mkdir secrets
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out secrets/jwt_private.pem
openssl rsa -pubout -in secrets/jwt_private.pem -out secrets/jwt_public.pem
```

Docker Compose 会把 `./secrets` 挂载到容器内的 `/app/secrets`，应用启动时自动读取。

## 本地运行

1. 创建 `.env` 文件

   ```bash
   # Windows
   copy .env.example .env

   # macOS / Linux
   cp .env.example .env
   ```

2. 安装依赖并启动

   ```bash
   go mod download
   go run main.go
   ```

服务默认监听 `http://localhost:8080`。

## Docker 运行

项目提供 `docker-compose.yml`，会同时启动 PostgreSQL 和后端服务。

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
- `app` 服务会显式注入数据库、邮件和 JWT 相关环境变量
- `app` 服务会读取挂载到 `/app/secrets` 的 JWT PEM 文件
- 数据会持久化到 `postgres_data` 卷

## API

当前路由前缀为 `/api/v1`。

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

### 注册

- `POST /api/v1/auth/register`
- 请求体：`username`、`email`、`password`、`code`
- 注册时会先校验验证码，再使用 bcrypt 哈希保存密码

## 数据库说明

- 启动时会自动执行 GORM 迁移，创建或更新 `users` 表结构
- `init.sql` 主要用于初始化扩展或补充手工 SQL
- `users.password` 现在存储 bcrypt 哈希，不再存放明文密码

## 运行注意事项

- 旧的明文密码账号需要重置密码或手动迁移为 bcrypt 哈希
- 邮件发送依赖 SMTP 配置，配置错误会导致验证码发送失败
- JWT 密钥建议通过文件挂载或 Secret 注入，不要提交到仓库
- 修改代码后，如果镜像内代码有变更，使用 `docker compose up -d --build`
