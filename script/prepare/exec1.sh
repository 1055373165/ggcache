#!/bin/bash

# 服务注册（模拟分布式场景下的节点服务注册）
go run ../../internal/middleware/etcd/put/put1/client_put1.go
go run ../../internal/middleware/etcd/put/put2/client_put2.go
go run ../../internal/middleware/etcd/put/put3/client_put3.go

# 进入 main.go 所在目录

# 在不同 terminal 运行程序 （为了方便 debug，自己手动运行）

# terminal 1
# go run main.go -port -9999 

# terminal 2
# go run main.go -port -10000

# terminal 3
# go run main.go -port -10001