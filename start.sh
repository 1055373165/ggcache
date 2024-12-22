#!/bin/bash

# 设置默认值
MYSQL_HOST=${MYSQL_HOST:-"127.0.0.1"}
MYSQL_PORT=${MYSQL_PORT:-"3307"}
MYSQL_USER=${MYSQL_USER:-"root"}
MYSQL_PASSWORD=${MYSQL_PASSWORD:-"root"}
MYSQL_DATABASE=${MYSQL_DATABASE:-"ggcache"}

# 检查是否在容器内部运行
if [ -f "/.dockerenv" ]; then
    echo "Running inside container, skipping docker-compose..."
else
    echo "Running outside container, starting dependency containers..."
    # 复制本地开发环境专用的 Prometheus 配置
    cp deploy/prometheus/prometheus.local.yml deploy/prometheus/prometheus.yml
    # 启动依赖服务
    docker-compose -f docker-compose-local.yml up -d mysql prometheus grafana
fi

# 等待 MySQL 就绪
echo "Trying to connect to MySQL at ${MYSQL_HOST}:${MYSQL_PORT} with user ${MYSQL_USER}"
until mysql -h "${MYSQL_HOST}" -P "${MYSQL_PORT}" -u "${MYSQL_USER}" -p"${MYSQL_PASSWORD}" -e "SELECT 1" >/dev/null 2>&1; do
    echo "MySQL is not ready yet... retrying in 2 seconds"
    sleep 2
done
echo "MySQL is ready!"

# 启动 etcd 集群
echo "Starting etcd cluster..."
goreman -f pkg/etcd/cluster/Procfile start &
sleep 5

# 启动 ggcache 服务
echo "Starting ggcache services..."
go run main.go -port 9999 &
sleep 3

go run main.go -port 10000 -metricsPort 2223 -pprofPort 6061 &
sleep 3

go run main.go -port 10001 -metricsPort 2224 -pprofPort 6062 &
sleep 3

# 启动客户端测试
echo "Starting client tests..."
./test/grpc/run_clients.sh

# 等待所有后台进程
wait
