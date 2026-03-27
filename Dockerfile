# 第一阶段：构建阶段
FROM golang:1.25.3-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache git

# 设置工作目录
WORKDIR /build

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o blog-api main.go

# 第二阶段：运行阶段
FROM alpine:latest

# 安装必要的工具（如果需要调试）
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1001 app && adduser -D -u 1001 -G app app

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/blog-api .

# 设置文件权限
RUN chown -R app:app /app && chmod 400 .env

# 切换用户
USER app

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./blog-api"]
