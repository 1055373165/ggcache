#! /bin/bash
trap "rm kv;kill 0" EXIT

# 单独非后台执行（为了观察日志输出）
cd ..
go build -o kv
./kv -port 9999