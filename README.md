# Distributed_KV_Store

简介：支持 HTTP、RPC 和服务注册发现的分布式键值存储系统

包含详细的项目设计与分析（共 9 个部分）

# 项目结构

```
.
├── README.md
├── conf                            // 日志和 mysql 数据库配置
│   ├── conf.go
│   └── init.go
├── distributekv                    // 分布式键值存储系统
│   ├── byteview.go                 // 并发读写
│   ├── cache.go                    // 支持缓存淘汰策略的底层缓存
│   ├── client.go                   // gRPC 客户端实现
│   ├── consistenthash              // 一致性 hash 算法（负载均衡）
│   ├── group.go                    // 测试用的数据库和缓存数据
│   ├── groupcache.go               // 对底层缓存的封装（资源隔离、并发安全）
│   ├── peers.go                    // 服务发现
│   ├── server.go                   // gRPC 服务端实现
│   └── singleflight                // 并发请求处理优化（协程编排）
├── etcd
│   ├── cluster                     // etcd 3 节点集群
│   ├── discover.go                 // 服务发现
│   ├── getServerNodesByEtcd.go     // 从 etcd 获取服务节点信息
│   ├── getServicesAddrs.go         // 获取节点信息
│   ├── register.go                 // 服务注册（阻塞）
│   └── server_register_to_etcd     // 将服务节点信息注册到 etcd
├── go.mod
├── go.sum
├── grpc
│   ├── groupcachepb                // gRPC 
│   ├── rpcCallClient               // 简单 RPC 调用
│   ├── server                      // gRPC 服务端实现
│   └── serviceRegisterCall         // RPC 调用（以服务发现的方式）
├── main.go                             
├── middleware                      
│   ├── db
│   └── logger
├── policy                          // 缓存淘汰策略
│   ├── fifo.go                     // FIFO 淘汰策略
│   ├── fifo_test.go
│   ├── kv.go                       // 使用策略模式
│   ├── lfu.go                      // LFU 淘汰策略（高级局部性）
│   ├── lfu_single.go
│   ├── lfu_test.go
│   ├── lru.go                      // LRU 淘汰策略（基本局部性）
│   ├── lru_test.go
│   ├── priority_queue.go           // 基于堆实现的优先队列
│   └── priority_queue_test.go
├── resources
│   └── images
├── script                          // 自动化测试脚本
│   ├── build
│   ├── run.sh
│   ├── script.md                   // 脚本使用指南
│   └── test.sh
└── utils
    └── utils.go
```


