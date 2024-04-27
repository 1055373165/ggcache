# gRPC Group Cache

## 项目介绍

支持 `HTTP`、`RPC` 和服务注册发现的分布式键值缓存系统；

## 功能介绍
- 支持 RPC （`gRPC` 框架）
- 支持多种缓存淘汰策略替换（策略类模式：`LRU`、`LFU`、`FIFO`）
- 并发访问控制（`singleFlight`机制）
- 负载均衡策略（一致性哈希算法）
- 键值分组管理、键值分布式存储（扩展系统吞吐量和可用性）
- 外部存储高可用（`etcd` 集群模式）
- 服务注册发现（`etcd endpoint manager`）
- 提供了自动化测试脚本和相对完整的测试用例（使用查询学生分数进行模拟）

# 项目结构
```
.
├── README.md
├── api
│   ├── ggcache
│   ├── groupcachepb        // grpc server idl.
│   ├── studentpb           // business idl.
│   └── website             
├── assets
│   ├── image
│   └── sql
├── cmd
│   ├── grpc
│   │   ├── grpc1           // grpc server implement1
│   │   │   └── main.go  
│   │   ├── grpc2           // grpc server implement2
│   │   │   └── main.go
│   │   └── main.go         // grpc server implement
│   └── http                // http server implement
│       └── main.go
├── config                  // global config manage
│   ├── config.go
│   └── config.yml
├── go.mod
├── go.sum
├── internal                
│   ├── middleware            // depend on
│   │   └── etcd            
│   │       ├── cluster       // goreman etcd cluster manage
│   │       ├── discovery     // service registration discovery (three implement)
│   │       ├── list_peers.go
│   │       └── put           // service instance address put
│   ├── pkg
│   │   ├── student           // business logic
│   │   │   ├── dao
│   │   │   ├── ecode
│   │   │   ├── model
│   │   │   └── service
│   │   └── website          
│   └── service              // grpc group cache service imp.
│       ├── byteview.go      // read-only 
│       ├── cache.go         // concurrency-safe caching imp.
│       ├── cachepurge       // cache eviction algorithm implemented in strategy mode
│       │   ├── fifo         
│       │   ├── interfaces   // cache eviction algorithm interface abstract
│       │   ├── lfu
│       │   ├── lru
│       │   └── purge.go     // export
│       ├── consistenthash   // consistent hash algorithm for load balance
│       ├── group.go         
│       ├── groupcache.go    // group cache imp.
│       ├── grpc_fetcher.go  // grpc proxy 
│       ├── grpc_picker.go   // grpc peer selector
│       ├── http_fetcher.go  // http proxy
│       ├── http_helper.go   // http api server and http server start helper
│       ├── http_picker.go   // http peer selector
│       ├── interface.go     // grpc peer selector and grpc proxy abstract
│       ├── policy           // old version cache eviction algorithm
│       └── singleflight     // single flight concurrent access control
├── main.go                  // grpc server default imp.
├── script                   // grpc and http service test
│   ├── prepare
│   ├── test
│   │   ├── grpc1
│   │   ├── grpc2
│   │   └── http
│   ├── test.md              // test step
│   ├── test0.sh
│   ├── test1.sh
│   ├── test2.sh
│   └── test3.sh
└── utils
    ├── logger
    ├── shutdown            // goroutine gracefully_shutdown
    ├── trace
    └── validate            // ip address validation

49 directories, 66 files
```

# 系统运行

## 预热阶段
https://github.com/1055373165/ggcache/assets/33158355/bcd6d2e7-979b-4b1a-b021-b09e654f4bf0


## 工作阶段
https://github.com/1055373165/ggcache/assets/33158355/5df39b9a-7dca-46f0-bb08-e48a84f43e19

# 使用

## 依赖

- etcd 服务（"127.0.0.1:2379"）
- mysql 服务（"127.0.0.1:3306"）
- etcd 集群（高可用）
    - `internal/middleware/etcd/cluster` 中有使用说明

## 运行测试

1. 将服务实例地址注册到 etcd （外部统一存储中心）

`./script/prepare/exec1.sh`
  
2. 启动三个 grpc server

系统将 server port 使用命令行参数导入，因此可以自定义任意数量的 grpc server

- terminal1: `go run main.go -port 9999`
- terminal2: `go run main.go -port 10000`
- terminal3: `go run main.go -port 10001`

3. 执行 rpc call 测试

- `go run /script/test/grpc1/grpc_client1.go`
- `go run /script/test/grpc2/grpc_client2.go`

> http 测试与上面类似
> 在 script/test.md 有完整的测试介绍

# 功能优化方向（todo）
- 添加缓存命中率指标（动态调整缓存容量）
- 自动检测服务节点信息变化，动态增删节点（节点变化时，动态重构哈希环，对于系统的请求分发十分重要）；
    - 实现思路一：监听 server 启停信号，使用 `endpoint manager`管理
    - 实现思路二：使用 etcd 官方 WatchChan() 提供的服务订阅发布机制）
- 增加更多的负载均衡策略
    - 添加 `arc`算法
    - `LRU2` 算法升级
- 增加请求限流（令牌桶算法）
- 实现缓存和数据库的一致性（增加消息队列异步处理）
- ......

# 参考资源链接
1. [ Geektutu]( https://geektutu.com/post/geecache.html) 分布式缓存 `GeeCache`
2. [gcache](https://github.com/bluele/gcache) 缓存淘汰策略（基于策略模式）
3. [groupcache](https://github.com/golang/groupcache) 常作为 `memcached` 替代
4. [grpc](https://grpc.io/docs/languages/go/quickstart/) `gRPC` 官方文档
5. [proto3](https://protobuf.dev/programming-guides/proto3/) `protobuf` 官方文档
6. [protobuf](https://www.notion.so/blockchainsee/Protocol-Buffer-04cba19af055479299507f04d0a24862) `protobuf` 编码原理
7. [protoscope](https://www.notion.so/blockchainsee/protoscope-fbfe36c2eef64bfcb630be4f0bd673f5) `protobuf` 个人学习笔记
8. [etcd](https://etcd.io/docs/v3.5/) 官方文档
9. [goreman](https://github.com/mattn/goreman) `etcd` 集群搭建
10. [shell](https://www.shellscript.sh/) shell 脚本
11. [gorm](https://gorm.io/docs/models.html) 快速搭建后端数据库
12. [air](https://github.com/cosmtrek/air) 动态加载（方便调试）
13. [log](https://github.com/charmbracelet/log) 极简、多彩的 `Go` 日志库

