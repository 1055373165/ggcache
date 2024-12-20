#!/bin/bash

echo "Stopping all running services..."

# 停止所有运行的go进程（main.go实例）
echo "Stopping Go servers..."
pkill -f "go run main.go"

# 停止goreman管理的etcd集群
echo "Stopping etcd cluster..."
pkill -f goreman

# 停止所有运行的客户端
echo "Stopping clients..."
pkill -f "go run.*grpc_client.go"

# 停止并删除docker容器
echo "Stopping docker containers..."
docker-compose down

echo "All services have been stopped."
