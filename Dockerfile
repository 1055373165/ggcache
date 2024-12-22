# 设置入口点
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
    default-mysql-client \
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

# 使用 Docker 专用配置文件
RUN cp config/config.docker.yml config/config.yml

# 安装 goreman (禁用 CGO 并使用特定版本)
ENV CGO_ENABLED=0
RUN go install github.com/mattn/goreman@v0.3.15

# 设置环境变量
ENV ETCD_CLUSTER_ENDPOINTS="http://127.0.0.1:2379,http://127.0.0.1:22379,http://127.0.0.1:32379"
ENV MYSQL_HOST="mysql"
ENV MYSQL_PORT="3306"
ENV MYSQL_USER="root"
ENV MYSQL_PASSWORD="root"
ENV MYSQL_DATABASE="ggcache"

# 暴露服务端口
EXPOSE 9999 10000 10001 
EXPOSE 2222 2223 2224   
EXPOSE 6060 6061 6062   
EXPOSE 2379 22379 32379 

# 确保 start.sh 有执行权限
RUN chmod +x /app/start.sh

# 设置入口点
ENTRYPOINT ["/app/start.sh"]