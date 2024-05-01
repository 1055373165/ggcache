# gRPC Group Cache(ggcache)

## 项目介绍

支持分布式节点间使用 `HTTP` 或 `gRPC`协议通信、

支持高可用的服务注册发现、

支持动态节点管理和网络拓扑快速收敛的分布式键值缓存系统；

## 功能介绍

- v1 version，实现了
  
    - 并发访问控制（singleFlight）

    - 负载均衡（consistenthash 算法）

    - 多种缓存淘汰策略（lru、lfu、fifo，策略类模式）

    - 分布式缓存节点间基于 http 协议的通信

    - 分布式缓存节点间基于 gRPC 协议的通信

    - 简单的服务注册发现（需要手动导入）

    - 高可用 etcd 集群（使用 goreman 进行配置）

    - 简单测试用例

- v2 version，实现了

    - 改进 singleFlight，增加结果缓存逻辑提升性能

    - 增加 TTL 机制（ 单独一个 goroutine），自动清理过期缓存

    - 实现了简单的业务逻辑，在数据库和缓存之间进行交互测试（两种测试用例）

    - 改进服务注册发现，使用 endpoint manager 对节点进行管理

    - 提供自动化测试脚本和环境搭建介绍
 
- v3 version，实现了

    - 服务注册发现最终版（使用 endpoint manager 和 watch channel 实现类似于服务订阅发布的能力）

    - 使用类似于事件回调的处理机制，根据节点的 PUT、DEL 事件更新节点状态

    - 实现秒级节点之间拓扑结构的快速收敛（动态节点管理）

    - 增加 grpc client 测试重试逻辑

    - failover 机制，节点失效后请求将转发到其他节点处理；即使所有节点下线，只要其中一个节点完成重启仍可继续提供服务（需要重新缓存预热）

    - 不同节点预热完成后，节点之间的 rpc 调用时延仅为 1ms（但是如果查询的是不存在的数据，延迟最高达到 999ms）
 
    - 即使其中一个节点崩溃或者宕机，也不会大范围影响其他已经预热好的节点（只会影响虚拟节点逆时针方向的一小段范围），而且该节点恢复后可以直接提供服务
 
    - 增加缓存穿透的防御策略（将不存在的 key 的空值存到缓存中，设置合理过期时间，防止不存在的 key 的大量并发请求打穿数据库）
 

> Rob Pike："Dont't communicate by sharing memory, share memory by communicating". 不要通过共享内存来通信，应该通过通信来共享内存。

> 这句话奠定了 Go 应用并发设计的主流风格：使用 channel 进行不同 goroutine 之间的通信。在动态节点管理实现时，使用 channel 实现了负责哈希环视图重建的 goroutine（g1）和负责监听 endpoint 事件变更 goroutine（g2）之间的通信。一旦系统新增或者移除了节点，g2 监听到了变更事件，通过 g1 和 g2 共享的信号 channel 告知 g1，g1 收到通知后上锁重建哈希环视图，从而实现并发安全的动态节点管理。


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

## 依赖

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


## 系统运行

### 预热阶段
https://github.com/1055373165/ggcache/assets/33158355/3215adb2-6615-481f-861d-6c1369179d7c

1. 使用 goreman 启动 etcd 集群（3 个 etcd 节点），实现高可用的、强一致性的外部存储中心

2. 启动多个 grpc service 服务节点，可以在秒级获取到节点信息变化，快速收敛

3. 启动多个 grpc client 客户端，不停地向 grpc 服务集群发送查询请求（其中特意添加了不存在的 key）

4. 每个节点刚启动时并没有缓存数据，都需要跑到后端数据库进行查询，实际上是对缓存进行预热

### 工作阶段

https://github.com/1055373165/ggcache/assets/33158355/7b95a2c2-98a5-40bd-8467-b19c55a3292e

1. 某些节点的缓存预热完毕后，可以直接从缓存中返回结果，因此处理速度肉眼可见的变快了

2. 发送终止信号给其中两个服务节点，正常工作节点不受影响，重新启动节点在加入后继续提供服务（后台会自动重建网络拓扑）

3. 所有节点全部关闭，客户端重试机制生效，最大重试次数为 3 次（可配置），分别退避 1s、2s、4s，在重试期间，只有有一个节点完成重启可以立即返回结果

4. 节点重启后，重新回到之前正常状态，节点间负责的 key 关系不会发生变化（hash 环上虚拟节点固定），因此不会导致大量缓存失效


### 动态节点管理

https://github.com/1055373165/ggcache/assets/33158355/1c771e10-c11c-493f-8488-a1cfa7e45f1e

动态获取速度可自定义配置，因为一般情况下网络拓扑相对比较稳定，没有必要后台启动一个长轮询任务一直监听不太可能发生的事件，这对于系统性能是一种浪费。


## 功能优化方向（todo）

1. 自动检测服务节点信息变化，动态增删节点（节点变化时，动态重构哈希环，对于系统的请求分发十分重要）✅

- 实现思路一：监听 server 启停信号，使用 `endpoint manager`管理 ✅

- 实现思路二：使用 etcd 官方 WatchChan() 提供的服务订阅发布机制）✅
    
2. 添加缓存命中率指标（动态调整缓存容量）todo

3. 负载均衡策略优化 ✅

- 添加 `arc`算法 todo

- `LRU2` 算法升级（高低水位） todo

4. 增加请求限流（令牌桶算法） todo

5. 增加对不存在 key 的特殊处理 ✅

6. 实现缓存和数据库的一致性（增加消息队列异步处理）（也可以通过缓存淘汰时的回调函数实现） todo（重点）

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

