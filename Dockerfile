FROM golang:alpine
MAINTAINER 1667834841@qq.com
RUN go env -w GO111MODULE=on && go env -w GOPROXY=https://goproxy.io,direct

RUN mkdir /opt/go
WORKDIR /opt/go
COPY . /opt/go/
RUN cd /opt/go
RUN go build -o smzdmPusher
#CMD ./smzdmPusher

RUN apk add --no-cache gcc musl-dev sqlite-dev  dropbear shadow bash
RUN wget https://github.com/erebe/wstunnel/releases/download/v10.4.4/wstunnel_10.4.4_linux_amd64.tar.gz && \
    tar -xzf wstunnel_10.4.4_linux_amd64.tar.gz && \
    chmod +x ./wstunnel

CMD ["sh", "-c", "if [ -z \"$ROOT_PASSWORD\" ]; then echo \"警告：ROOT_PASSWORD 环境变量未设置，将无法通过密码登录 SSH。\" >&2; else echo \"root:$ROOT_PASSWORD\" | chpasswd; fi && \
                     ./wstunnel server ws://0.0.0.0:8080 & \
                     dropbear -F -p 22 & \
                     ./smzdmPusher"]
RUN mkdir -p /data
VOLUME /data
