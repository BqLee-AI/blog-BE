# blog-BE AGENTS 规范

你是 blog-BE 后端工程 Agent。目标不是写更多代码，而是在控制面约束下交付可验证的工程产出。

## 1) 核心职责

- 按控制面 Spec + OpenAPI 契约实现 API 端点
- 遵循 handler → service → models(含 DAO) → dao(连接) 的分层架构
- 每次变更必须通过 `go build ./...`，变更完成后本地验证
- API 变更必须同步更新 `blog-Docs/contracts/api/openapi.yaml`

## 2) 非目标

- 不改统一响应格式 `{ message, data, requestId, code }`
- 不改 JWT RSA256 签名机制
- 不改 `/api/v1` 路由前缀
- 不在本仓库实现前端逻辑、CLI 工具或外部系统对接
- 不引入 gRPC、GraphQL、消息队列
- 不在 handler 中直接调用 `dao.DB`

## 3) 技术边界

### 分层职责

- `handler/`：HTTP 请求解析、响应组装，不写业务逻辑，不直连数据库。auth 分布在 `handler.go`（login/register/refresh/me）和 `auth_handler.go`（验证码发送/邮箱验证）
- `service/`：跨模型编排、外部调用（邮件 `gomail.go`、验证码 `verification.go`），不感知 HTTP
- `models/`：GORM 模型定义 + 单表 CRUD，DAO 方法挂在 model 上
- `models/request/`：请求 DTO
- `dao/`：数据库连接单例，只暴露 `dao.DB`
- `config/`：配置加载（支持 .env + yaml），Redis 连接
- `logger/`：结构化日志
- `middleware/`：认证、CORS、请求 ID、请求日志，不夹带业务逻辑
- `utils/`：JWT、密码哈希、响应构造，不依赖 models

### 禁止跨越

- handler 禁止直接调用 `dao.DB`
- service 禁止直接操作 HTTP 请求/响应
- 新增路由必须在 `src/routers/router.go` 注册

### 技术栈锁定

Go 1.25 · Gin · GORM · PostgreSQL 15 · Redis · JWT(RSA256) · bcrypt · gomail · zap(结构化日志)

## 4) 开发命令

```bash
go build ./...          # 编译检查（PR 前必须通过）
go test ./...           # 运行测试
go vet ./...            # 静态分析
go run main.go          # 本地启动（需 .env 或 config.yaml + secrets/ 下 RSA 密钥）
docker compose up -d    # 启动 PostgreSQL + Redis
```

配置支持 `.env` 和 `config.*.yaml` 两种方式，参照 `.env.example` 和 `config.development.yaml`。敏感文件（`.env`、`secrets/*.pem`）不入库。

## 5) 控制面读取顺序

接任务前必须按以下顺序加载控制面上下文：

1. **控制面边界规则**：`blog-Docs/docs/control-plane/mvp-scope.md` → 确认任务在冻结范围内
2. **契约规则**：`blog-Docs/contracts/api/openapi.yaml` → 确认端点定义、请求/响应结构
3. **Harness 清单**：`blog-Docs/docs/harness/harness-gate-baseline.md` → 确认验证标准

如果执行中发现对全局有长远影响的关键约束（如新发现的安全风险、架构边界冲突、容易踩的坑），必须回写到本文件（AGENTS.md）对应章节，不允许只在代码里默默修掉。

## 6) 任务执行原则

- 先读代码再改：修改前先看 handler → model → dao 的完整链路，不凭猜测改
- 变更最小：只改完成任务所需的文件，不顺手重构无关代码
- 契约优先：新增或修改 API 端点时，先确认 openapi.yaml 中是否有对应定义，没有则先提 OpenSpec change
- 错误码规范：使用 `utils.NewResponse` + 语义化 code（如 `AUTH_FAILED`、`NOT_FOUND`、`INVALID_REQUEST`），不发明新格式

## 7) 代码规范

### 命名

- 文件：`snake_case.go`
- 导出函数：`PascalCase`
- 请求结构体放 `models/request/`，响应结构体定义在使用它的 handler 文件内
- Handler 函数命名：`GetXxx`、`CreateXxx`、`UpdateXxx`、`DeleteXxx`

### 编码

- 密码操作必须通过 `utils.HashPassword` / `utils.CheckPassword`
- 软删除使用 GORM `DeletedAt`，不做物理删除
- 新增配置项必须同步 `.env.example`

## 8) 变更边界

### 可以直接做

- 新增 handler/service/model
- 修复 bug、补充错误处理
- 补充单测
- 调整已有端点的内部实现（不改接口签名）

### 先确认再做

- 新增 API 端点（需先确认 openapi.yaml）
- 修改已有端点的请求/响应结构（影响 FE）
- 新增第三方依赖
- 修改中间件行为

### 不要做

- 在 handler 中写 SQL 或调用 `dao.DB`
- 修改 `utils/response.go` 的响应格式
- 提交 `.env`、`secrets/`、`*.pem` 等敏感文件
- 在实现中偷偷加 Spec 范围外的功能

## 9) 交付定义（DoD）

代码变更完成必须满足：

- `go build ./...` 通过
- API 变更已同步到 `blog-Docs/contracts/api/openapi.yaml`
- Commit 符合 Angular 规范
- 无敏感信息进入变更
- 变更为单一目标、无无关文件修改
