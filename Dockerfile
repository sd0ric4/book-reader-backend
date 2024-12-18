# 使用 Golang 官方镜像作为构建环境
FROM golang:1.23-alpine as builder

# 设置工作目录
WORKDIR /app

# 复制应用代码到容器
COPY . .

# 下载依赖并编译
RUN go mod tidy && go build -o main .

# 第二阶段：构建运行镜像
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates

# 从构建阶段复制二进制文件
COPY --from=builder /app/main /main

# 设置容器启动命令
ENTRYPOINT ["/main"]

# 暴露应用运行端口
EXPOSE 8080
