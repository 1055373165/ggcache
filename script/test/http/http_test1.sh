#!/bin/bash
trap "rm server;kill 0" EXIT

go build -o ./server ../cmd/http/main.go
./server -port=8001 &
./server -port=8002 &
./server -port=8003 -api=1 &

sleep 2
echo ">>> start test"
curl "http://localhost:9999/api?key=张三" &
curl "http://localhost:9999/api?key=张三" &
curl "http://localhost:9999/api?key=张三" &

# 服务一直启动，剩下的可以自己手动进行测试；http_test2.sh 提供全自动测试
wait