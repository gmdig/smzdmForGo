# 基于 Golang Alpine 镜像，包含 Go 语言环境
FROM golang:alpine
MAINTAINER 1667834841@qq.com

# 设置 Go 代理
RUN go env -w GO111MODULE=on && go env -w GOPROXY=https://goproxy.io,direct

# 安装 Go 应用的依赖和 SSH 所需的工具
# dropbear: 轻量级 SSH 服务器
# shadow: 提供 chpasswd 命令来设置密码
# bash: 提供更强大的脚本能力
RUN apk add --no-cache gcc musl-dev sqlite-dev dropbear shadow bash

# 创建 Go 应用的工作目录并复制所有项目文件
RUN mkdir /opt/go
WORKDIR /opt/go
COPY . /opt/go/

# 编译 Go 应用
RUN go build -o smzdmPusher

# 创建数据目录
RUN mkdir -p /data
VOLUME /data

# 暴露端口，这里假设 Go 应用和 SSH 都需要暴露端口
# 8080 用于 SSH over WebSocket，8081 用于 Go 应用（如果需要的话）
EXPOSE 8080
EXPOSE 8081

# 启动容器时执行的命令
# 1. 检查环境变量 ROOT_PASSWORD 是否设置
# 2. 如果设置了，就用它来给 root 用户设置密码
# 3. 启动 wstunnel 服务，并把它放到后台运行
# 4. 启动 dropbear SSH 服务，并把它放到后台运行
# 5. 最后，启动你的主应用 smzdmPusher，保持在前台运行
CMD ["sh", "-c", "if [ -z \"$ROOT_PASSWORD\" ]; then echo \"警告：ROOT_PASSWORD 环境变量未设置，将无法通过密码登录 SSH。\" >&2; else echo \"root:$ROOT_PASSWORD\" | chpasswd; fi && \
                     wstunnel --server --external-port 8080 --internal-port 22 & \
                     /usr/bin/dropbear -F -p 22 & \
                     ./smzdmPusher"]
