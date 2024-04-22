#!/bin/bash
trap "rm server; kill 0" EXIT

go build -o ./server ../cmd/http/main.go
./server -port=8001 &
./server -port=8002 &
./server -port=8003 -api=1 &

sleep 2
echo ">>> start test"

# 定义名字数组
names=("张三" "李四" "王五" "李六" "赵七" "孙八" "钱九" "周十" "one" "two" "three" "four" "five" "six" "unknwon")

# 在后台启动无限循环
(
    while true; do
        for name in "${names[@]}"; do
            curl "http://localhost:9999/api?key=$name" &
            sleep 0.3  # 每次请求之间暂停 0.3 秒，可根据需要调整
        done
    done
) &

wait
