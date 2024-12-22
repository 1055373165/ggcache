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
    curl \
    mysql-client \
    && rm -rf /var/lib/apt/lists/*

# 安装 etcd
RUN ETCD_VER=v3.5.11 && \
DOWNLOAD_URL=https://github.com/etcd-io/etcd/releases/download && \
curl -L ${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o etcd.tar.gz && \
tar xzvf etcd.tar.gz && \
mv etcd-${ETCD_VER}-linux-amd64/etcd /usr/local/bin/ && \
mv etcd-${ETCD_VER}-linux-amd64/etcdctl /usr/local/bin/ && \
rm -rf etcd-${ETCD_VER}-linux-amd64 etcd.tar.gz

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
ENV MYSQL_HOST=mysql
ENV MYSQL_PORT=3306
ENV MYSQL_USER=root
ENV MYSQL_PASSWORD=root
ENV MYSQL_DATABASE=ggcache

# 暴露服务端口
EXPOSE 9999 10000 10001 
EXPOSE 2222 2223 2224   
EXPOSE 6060 6061 6062   
EXPOSE 2379 22379 32379 

# 创建启动脚本
RUN echo '#!/bin/sh\n\
\n\
# 等待 MySQL 就绪\n\
until mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "SELECT 1" >/dev/null 2>&1; do\n\
  echo "Waiting for MySQL to be ready..."\n\
  sleep 2\n\
done\n\
echo "MySQL is ready!"\n\
\n\
# 更新配置文件中的 MySQL 连接信息\n\
sed -i "s/host: .*/host: $MYSQL_HOST/" /app/config/config.yml\n\
sed -i "s/port: .*/port: $MYSQL_PORT/" /app/config/config.yml\n\
sed -i "s/username: .*/username: $MYSQL_USER/" /app/config/config.yml\n\
sed -i "s/password: .*/password: $MYSQL_PASSWORD/" /app/config/config.yml\n\
\n\
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