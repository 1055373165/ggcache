#!/bin/bash

# 启动依赖的容器服务
echo "Starting dependency containers..."
docker-compose up -d mysql prometheus grafana

# 等待 MySQL 就绪
echo "Waiting for MySQL to be ready..."
until mysql -h 127.0.0.1 -P 3307 -u root -proot -e "SELECT 1" >/dev/null 2>&1; do
    echo "MySQL is not ready yet..."
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
