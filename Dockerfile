# ----------------------------------
# 第一阶段：编译 Go 应用
# 使用 go:alpine 作为构建环境
# ----------------------------------
FROM golang:alpine AS builder
MAINTAINER 1667834841@qq.com

# 设置 Go 代理
RUN go env -w GO111MODULE=on && go env -w GOPROXY=https://goproxy.io,direct

# 创建工作目录并复制项目文件，包括配置文件
WORKDIR /app
COPY . /app/

# 编译 Go 应用
RUN go build -o smzdmPusher

# ----------------------------------
# 第二阶段：设置运行时环境
# 基于 alpine 镜像，并添加 SSH 和应用
# ----------------------------------
FROM alpine:latest
MAINTAINER 1667834841@qq.com

# 安装所需的包，包括 Go 应用依赖和 SSH 工具
RUN apk add --no-cache gcc musl-dev sqlite-dev dropbear shadow bash

# 设置工作目录，确保应用能找到其配置文件
WORKDIR /app

# 从构建阶段复制编译好的 Go 应用和配置文件
COPY --from=builder /app/smzdmPusher /app/smzdmPusher
COPY --from=builder /app/config/ /app/config/

# 创建数据目录
RUN mkdir -p /data
VOLUME /data

# 暴露端口，这里假设 Go 应用和 SSH 都需要暴露端口
EXPOSE 8080
EXPOSE 9090

# 启动容器时执行的命令，集成 SSH 配置和 Go 应用启动
CMD ["sh", "-c", "if [ -z \"$ROOT_PASSWORD\" ]; then echo \"Error: ROOT_PASSWORD environment variable is not set.\" && exit 1; fi && \
                     echo \"root:$ROOT_PASSWORD\" | chpasswd && \
                     wstunnel --server --external-port 8080 --internal-port 22 & \
                     ./smzdmPusher"]
