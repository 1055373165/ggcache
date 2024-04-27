# gRPC Group Cache

## 项目介绍

支持 `HTTP`、`RPC` 和服务注册发现的分布式键值缓存系统；

## 功能介绍

- v1 version，实现了
  
    - 并发访问控制（singleFlight）

    - 负载均衡（consistenthash 算法）

    - 多种缓存淘汰策略（lru、lfu、fifo，策略类模式）

    - 分布式缓存节点间基于 http 协议的通信

    - 分布式缓存节点间基于 gRPC 协议的通信

    - 简单的服务注册发现（需要手动导入）

    - 高可用 etcd 集群（使用 goerman 进行配置）

    - 简单测试用例

- v2 version，实现了

    - 改进 singleFlight，增加结果缓存逻辑提升性能

    - 增加 TTL 机制，自动清理过期缓存

    - 实现了简单的业务逻辑，在数据库和缓存之间进行交互测试（两种测试用例）

    - 改进服务注册发现，使用 endpoint manager 对节点进行管理

    - 提供自动化测试脚本和环境搭建介绍
 
- v3 version，实现了

    - 服务注册发现最终版（使用 endpoint manager 和 watch channel实现类似于服务订阅发布的能力）

    - 使用类似于事件回调的处理机制，根据节点的 PUT、DEL 事件更新节点状态

    - 实现秒级节点之间拓扑结构的快速收敛（动态节点管理）

    - 增加 grpc client 测试重试逻辑

    - grpc server 崩溃恢复后，能够接着处理 client 发来的请求

    - 不同节点预热完成后，节点之间的 rpc 调用时延仅为 1ms（但是如果查询的是不存在的数据，延迟最高达到 999ms）
 
    - 即使其中一个节点崩溃或者宕机，也不会影响其他已经预热好的节点（key 和节点之间的关系不受影响，正常提供服务），而且该节点恢复后可以直接提供服务
 
    - 增加缓存穿透的防御策略（将不存在的 key 的空值存到缓存中，设置合理过期时间，防止不存在的 key 的大量并发请求打穿数据库）

## 项目结构
```
.
├── README.md
├── api
│   ├── groupcachepb        // grpc server idl.
│   ├── studentpb           // business idl.
│   └── website             
├── cmd
│   ├── grpc
│   │   └── main.go         // grpc server latest version implement
│   └── http                // http server implement
├── config                  // global config manage
│   ├── config.go
│   └── config.yml
├── go.mod
├── go.sum
├── internal                
│   ├── middleware 
│   │   └── etcd            
│   │       ├── cluster       // goerman etcd cluster manage
│   │       ├── discovery     // service registration discovery (latest version use discovery3 impl.) 
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
│       │   └── purge.go     // export distributed kv object
│       ├── consistenthash   // consistent hash algorithm for load balance
│       ├── group.go         
│       ├── groupcache.go    // group cache imp.
│       ├── grpc_fetcher.go  // grpc proxy 
│       ├── grpc_picker.go   // grpc peer selector
│       ├── http_fetcher.go  // http proxy
│       ├── http_helper.go   // http api server and http server start helper
│       ├── http_picker.go   // http peer selector
│       ├── interface.go     // grpc peer selector and grpc proxy abstract
│       └── singleflight     // single flight concurrent access control
├── main.go                  // grpc server default imp. equal to cmd/grpc/main.go
├── script                   // grpc and http service test
│   ├── test
│   │   ├── grpc
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
```

## 系统运行

### 预热阶段
https://github.com/1055373165/ggcache/assets/33158355/bcd6d2e7-979b-4b1a-b021-b09e654f4bf0



### 工作阶段
https://github.com/1055373165/ggcache/assets/33158355/5df39b9a-7dca-46f0-bb08-e48a84f43e19

## 使用

### 依赖

1. etcd 服务（"127.0.0.1:2379"）

2. mysql 服务（"127.0.0.1:3306"）

3. goreman (etcd 集群) `internal/middleware/etcd/cluster` 中有使用说明

    - 本项目默认依赖了 etcd 集群，需要先关闭本地的 etcd 服务，然后使用 goreman 同时启动三个 etcd 服务进行统一管理
    - 如果不想依赖 etcd 集群，那么在 config/config.yml 中将 22379 和 32379 删除然后启动本地 etcd 服务即可

5. gorm（数据库）

6. grpc（rpc 通信协议）

7. logrus（日志管理）

8. viper（配置管理）

9. protobuf（序列化）

### 运行测试

1. 将服务实例地址注册到 etcd （外部统一存储中心）

`./script/prepare/exec1.sh`
  
2. 启动三个 grpc server

- terminal1: `go run main.go -port 9999`
- terminal2: `go run main.go -port 10000`
- terminal3: `go run main.go -port 10001`

3. 执行 rpc call 测试

- `go run /script/test/grpc1/grpc_client1.go`
- `go run /script/test/grpc2/grpc_client2.go`

> http 测试与上面类似
> 在 script/test.md 有完整的测试介绍
> 上面只是以三个节点为例，系统将 server port 使用命令行参数导入，因此可以自定义任意数量的 grpc server（前提是需要将它的服务地址率先导入 etcd）

## 功能优化方向（todo）

1. 自动检测服务节点信息变化，动态增删节点（节点变化时，动态重构哈希环，对于系统的请求分发十分重要）

- 实现思路一：监听 server 启停信号，使用 `endpoint manager`管理

- 实现思路二：使用 etcd 官方 WatchChan() 提供的服务订阅发布机制）
    
2. 添加缓存命中率指标（动态调整缓存容量）

3. 负载均衡策略优化

- 添加 `arc`算法

- `LRU2` 算法升级（高低水位）

4. 增加请求限流（令牌桶算法）

5. 实现缓存和数据库的一致性（增加消息队列异步处理）

...

## 参考资源链接
1. [Geektutu 分布式缓存 `GeeCache`](https://geektutu.com/post/geecache.html) 
2. [gcache 缓存淘汰策略](https://github.com/bluele/gcache)
3. [groupcache 瑞士军刀](https://github.com/golang/groupcache) 
4. [ `gRPC` 官方文档](https://grpc.io/docs/languages/go/quickstart/)
5. [`protobuf` 官方文档](https://protobuf.dev/programming-guides/proto3/) 
6. [`protobuf` 编码原理](https://www.notion.so/blockchainsee/Protocol-Buffer-04cba19af055479299507f04d0a24862) 
7. [`protobuf` 个人学习笔记](https://www.notion.so/blockchainsee/protoscope-fbfe36c2eef64bfcb630be4f0bd673f5) 
8. [etcd 官方文档](https://etcd.io/docs/v3.5/) 
9. [ `etcd` 集群搭建](https://github.com/mattn/goreman)
10. [shell 脚本](https://www.shellscript.sh/) 
11. [golang 数据库框架](https://gorm.io/docs/models.html) 
12. [air 动态加载](https://github.com/cosmtrek/air) 
13. [极简、多彩的 `Go` 日志库](https://github.com/charmbracelet/log) 

