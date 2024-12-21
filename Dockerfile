# 使用官方 Go 镜像作为基础镜像
FROM golang:1.21

# 安装必要的系统依赖
RUN apt-get update && apt-get install -y \
    bash \
    git \
    make \
    gcc \
    libc-dev \
    lsof \
    && rm -rf /var/lib/apt/lists/*

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 安装 goreman (禁用 CGO 并使用特定版本)
ENV CGO_ENABLED=0
RUN go install github.com/mattn/goreman@v0.3.15

# 设置环境变量
ENV ETCD_CLUSTER_ENDPOINTS="http://127.0.0.1:2379,http://127.0.0.1:22379,http://127.0.0.1:32379"

# 暴露服务端口
EXPOSE 9999 10000 10001 
EXPOSE 2222 2223 2224   
EXPOSE 6060 6061 6062   
EXPOSE 2379 22379 32379 

# 创建启动脚本
RUN echo '#!/bin/sh\n\
# 启动 etcd 集群\n\
goreman -f pkg/etcd/cluster/Procfile start &\n\
sleep 5\n\
\n\
# 启动服务器\n\
go run main.go -port 9999 &\n\
sleep 3\n\
\n\
go run main.go -port 10000 -metricsPort 2223 -pprofPort 6061 &\n\
sleep 3\n\
\n\
go run main.go -port 10001 -metricsPort 2224 -pprofPort 6062 &\n\
sleep 3\n\
\n\
# 启动客户端测试\n\
./test/grpc/run_clients.sh\n\
\n\
# 保持容器运行\n\
tail -f /dev/null' > /app/docker-entrypoint.sh \
    && chmod +x /app/docker-entrypoint.sh

# 设置入口点
ENTRYPOINT ["/app/docker-entrypoint.sh"]