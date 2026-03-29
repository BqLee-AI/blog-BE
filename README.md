# Blog-BE

一个使用 Go + Gin + GORM + PostgreSQL 构建的博客后端服务。当前版本主要提供用户注册、登录以及邮件验证码能力，数据库表会在启动时自动迁移创建。

## 技术栈

- Go 1.25.3
- Gin
- GORM
- PostgreSQL
- gomail
- RSA JWT

## 功能说明

- 基于 Gin 提供 HTTP API
- 使用 PostgreSQL 作为数据存储
- 启动时自动连接数据库并迁移 `users` 表
- 注册时发送邮件验证码
- 登录时根据邮箱和密码校验用户身份，并签发 RSA access/refresh 双 token

## 目录结构

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
    ├── models/
    ├── routers/
    ├── service/
    └── utils/
```

## 环境要求

- Go 1.25 或更高版本
- PostgreSQL 15 或兼容版本
- 可用的 SMTP 邮箱账号，用于发送验证码邮件

## 环境变量

项目通过 `.env` 文件读取配置，可以直接参考 [.env.example](.env.example) 创建本地配置。

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
| `JWT_PRIVATE_KEY_PATH` | JWT 私钥文件路径 | `./secrets/jwt_private.pem` |
| `JWT_PUBLIC_KEY_PATH` | JWT 公钥文件路径 | `./secrets/jwt_public.pem` |
| `JWT_ACCESS_TTL` | Access token 有效期 | `15m` |
| `JWT_REFRESH_TTL` | Refresh token 有效期 | `168h` |

## 生成 JWT 密钥

先在项目根目录创建 `secrets` 目录，然后生成一对 RSA PEM 文件：

```bash
mkdir secrets
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out secrets/jwt_private.pem
openssl rsa -pubout -in secrets/jwt_private.pem -out secrets/jwt_public.pem
```

生成后，Docker Compose 会把 `./secrets` 目录挂载到容器内的 `/app/secrets`，应用会自动读取这两个文件。

## 本地运行

1. 创建 `.env` 文件

   ```bash
   # Windows
   copy .env.example .env
   
   # macOS / Linux
   cp .env.example .env

2. 启动 PostgreSQL

   可以使用本机数据库，或者直接使用 Docker 启动数据库容器。

3. 安装依赖并启动服务

   ```bash
   go mod download
   go run main.go
   ```

服务启动后默认监听 `http://localhost:8080`。

## Docker 运行

项目提供了 `docker-compose.yml`，会同时启动 PostgreSQL 和后端服务。

```bash
docker compose up --build
```

说明：

- `db` 服务会使用 `.env` 中的数据库配置
- `app` 服务会自动连接 `db` 容器
- `app` 服务会读取挂载到 `/app/secrets` 的 JWT PEM 文件
- 数据会持久化到名为 `postgres_data` 的卷

## API 接口

当前路由前缀为 `/api/v1`。

### 登录

- 方法：`POST`
- 路径：`/api/v1/auth/login`
- 参数：`email`、`password`
- 参数来源：JSON payload，请求头建议设置 `Content-Type: application/json`

响应示例：

```json
{
  "message": "Login successful",
  "data": {
    "user": {
      "user_id": 1,
      "username": "demo",
      "email": "demo@example.com",
      "role_id": 0
    },
    "tokens": {
      "token_type": "Bearer",
      "access_token": "eyJ...",
      "refresh_token": "eyJ...",
      "access_expires_at": "2026-03-28T12:00:00Z",
      "refresh_expires_at": "2026-04-04T12:00:00Z"
    }
  },
  "requestId": "trace-id"
}
```

### 刷新令牌

- 方法：`POST`
- 路径：`/api/v1/auth/refresh`
- 参数：`refresh_token`，或者通过 `Authorization: Bearer <token>` 传入
- 行为：校验 refresh token 后签发新的 access/refresh 双 token

### 当前用户

- 方法：`GET`
- 路径：`/api/v1/auth/me`
- 参数：`Authorization: Bearer <access_token>`
- 行为：校验 access token 并返回 token 中的用户信息

### 注册

- 方法：`POST`
- 路径：`/api/v1/auth/register`
- 参数：`username`、`email`、`password`、`code`
- 行为：先发送验证码邮件，再校验 `code`，校验通过后创建用户

响应示例：

```json
{
  "message": "Registration successful",
  "data": {
    "user_id": 1
  },
  "requestId": "trace-id"
}
```

## 数据库说明

- 启动时会自动执行 GORM 迁移，创建或更新 `users` 表结构
- `init.sql` 主要用于初始化扩展或补充手工 SQL
- 当前用户表结构包括 `id`、`username`、`password`、`email`、`role_id`

## 注意事项

- 当前代码里密码是直接明文比对的，生产环境应改为哈希存储与校验
- 注册接口依赖邮件服务，SMTP 配置错误会导致注册失败
- JWT 私钥和公钥建议通过文件挂载或 Secret 注入，不要提交到仓库
- 默认 Docker 镜像不会打包 `.env` 文件，请通过环境变量（或在 `docker-compose.yml` 中挂载 `.env`）向容器注入配置，避免在镜像中硬编码敏感信息

## 开发建议

- 接口返回已经统一封装为 `message`、`data`、`requestId`、`code`
- 后续可以补充 JWT 鉴权、密码哈希、用户资料接口和更完整的错误码规范
