#! /bin/bash

# 2. 在三个 terminal 分别启动 grpc server（为了方便调试，手动启动）
# terminal 1 启动
# terminal 2 启动
# terminal 3 启动

# 一键脚本的话，用后台启动命令实现
# terminal 1
# go run ../main.go -port 9999 &
# terminal 2
# go run ../main.go -port 10000 &
# terminal 3
# go run ../main.go -port 10001 

# ps -ef | grep 9999
# lsof -i:9999
# kill -9 pid