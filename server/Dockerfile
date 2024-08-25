# 使用官方的 Go 镜像作为基础镜像
FROM golang:1.23-alpine AS builder

# 设置工作目录为 /app
WORKDIR /app

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 下载依赖包
RUN go mod download

# 复制必要的源代码文件
COPY main.go ./  
COPY internal/ ./internal/  
COPY config/ ./config/

# 编译 Go 应用程序
RUN go build -o server-app ./main.go

# 使用一个更小的基础镜像
FROM alpine:latest

# 设置工作目录为 /root
WORKDIR /root/

# 从构建阶段复制编译后的二进制文件和配置文件
COPY --from=builder /app/server-app .
COPY --from=builder /app/config /root/config

# 创建日志目录
RUN mkdir -p /root/logs

# 启动应用程序
CMD ["./server-app"]
