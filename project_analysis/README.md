# Distributed_KV_Store
本项目参考 GeeCache 和 GroupCache 实现，并在 GeeCache 基础上进行了功能扩展。
GroupCache 源码简单走读：https://www.notion.so/blockchainsee/4f4cc29098e84c79ac8ff2e94c9dd9ee


# 分模块实现
> 所有测试用例均可正常使用
## 缓存淘汰策略模块
- 缓存淘汰策略模块主要实现缓存淘汰算法，包括 LRU、LFU、FIFO 等
为了方便区分，项目各个模块设计分析部分的目录使用大写开头命名；
在 /1-CacheEvicAlgorithm/LRU 部分对 LRU 算法的原理进行了解释，并实现了 LRU 算法；
在 /1-CacheEvicAlgorithm/LFU 部分对 LFU 算法的原理进行了解释，并实现了 LFU 算法；
在 /1-CacheEvicAlgorithm/FIFO 部分对 FIFO 算法的原理进行了解释，并实现了 FIFO 算法；
在 /1-CacheEvicAlgorithm/FIFO_LRU_LFU 部分对比了三种淘汰算法，并从基本局部性和高级局部性角度分析了LRU 和 LFU 的区别；

## 资源隔离模块（Group）

## HTTP 模块（负责基于 HTTP 的节点间通信）

## 分布式一致性hash 模块（负载均衡）

## SingleFlight 模块（防止缓存雪崩）

## gRPC 模块（负责基于 RPC 协议的通信）

## etcd 模块（负责服务注册发现等）


