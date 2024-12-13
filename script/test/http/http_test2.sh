#!/bin/bash
trap "rm server; kill 0" EXIT

go build -o ./server ../cmd/http/main.go
./server -port=8001 -api=1 &
./server -port=8002 &
./server -port=8003 &

echo "Waiting for HTTP API server to start on port 9999..."
while ! nc -z localhost 9999; do
  sleep 0.1
done

echo ">>> Start test"

# 定义名字数组
names=("钱九" "周十" "my" "oh" "ohmy" "李四" "王五" "李六" "赵七" "孙八" "钱九" "周十" "one" "two" "three" "four" "five" "six" "unknown" "钱九" "周十" "my" "oh" "ohmy" "李四" "王五" "李六" "赵七" "孙八" "钱九" "周十" "one" "two" "three" "four" "five" "six" "unknown")

for (( i=1; i<=10; i++ )); do
    names+=("$i")
done

for (( i=1; i<=10; i++ )); do
    names+=("$i")
done

# 在后台启动无限循环
(
    while true; do
        for name in "${names[@]}"; do
            # 执行 curl 请求，捕获输出并添加前缀
            result=$(curl -s "http://localhost:9999/api?key=$name")
            echo "curl 请求结果：$result"
            sleep 0.2  # 每次请求之间暂停 0.2 秒，可根据需要调整
        done
    done
) &

wait
