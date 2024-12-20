#!/bin/bash

# 配置参数
NUM_CLIENTS=20          # 要启动的客户端数量
INTERVAL=0.1           # 客户端启动间隔（秒）

# 获取脚本所在目录的绝对路径
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "Starting $NUM_CLIENTS gRPC clients..."

# 并发运行客户端
for ((i=1; i<=NUM_CLIENTS; i++))
do
    go run $SCRIPT_DIR/grpc_client.go &
    echo "Started client $i"
    sleep $INTERVAL
done

# 等待所有后台进程完成
echo "Waiting for all clients to finish..."
wait

echo "All clients finished."
