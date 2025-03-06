# 阶段一：构建阶段
FROM golang:1.24.1-alpine3.21 AS builder

# 设置工作目录
WORKDIR /build

# 复制Go程序文件到容器中
COPY . .

# 构建Go程序
RUN go build -o tarumt-wifi-autoconnect

# 阶段二：运行阶段
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从第一阶段中复制生成的可执行文件到当前容器
COPY --from=builder /build/tarumt-wifi-autoconnect /app/tarumt-wifi-autoconnect

# 定义启动容器时运行的命令
CMD ["/app/tarumt-wifi-autoconnect"]
