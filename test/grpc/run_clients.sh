#!/bin/bash

# 配置参数
NUM_CLIENTS=20          # 要启动的客户端数量
INTERVAL=0.1           # 客户端启动间隔（秒）
LOG_DIR="logs"         # 日志目录

# 获取脚本所在目录的绝对路径
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# 创建日志目录
mkdir -p "$SCRIPT_DIR/$LOG_DIR"

echo "Starting $NUM_CLIENTS gRPC clients..."
echo "Logs will be saved to $SCRIPT_DIR/$LOG_DIR/"

# 并发运行客户端
for ((i=1; i<=NUM_CLIENTS; i++))
do
    LOG_FILE="$SCRIPT_DIR/$LOG_DIR/client_${i}.log"
    go run $SCRIPT_DIR/grpc_client.go > "$LOG_FILE" 2>&1 &
    echo "Started client $i - logging to $LOG_FILE"
    sleep $INTERVAL
done

# 等待所有后台进程完成
echo "Waiting for all clients to finish..."
wait

# 汇总结果
echo "All clients finished. Summarizing results..."
SUCCESS_COUNT=$(grep -l "Success" "$SCRIPT_DIR/$LOG_DIR"/*.log | wc -l)
ERROR_COUNT=$(grep -l "Error" "$SCRIPT_DIR/$LOG_DIR"/*.log | wc -l)

echo "Summary:"
echo "- Total Clients: $NUM_CLIENTS"
echo "- Successful: $SUCCESS_COUNT"
echo "- Failed: $ERROR_COUNT"
echo "Detailed logs available in: $SCRIPT_DIR/$LOG_DIR/"
