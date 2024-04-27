# 测试介绍

## prepare
- prepare 中包含了系统运行前的准备操作，需要先将服务实例的地址列表导入到 etcd 中保存，这是 ggcache service 启动时构建完整哈希环需要的；
- 导入的 key 组合形式为："clusters" + "ip:port"
- 导入的 value 值为：服务实例的地址 "ip:port"
- 由于还未实现全自动化管理节点，因此还需要一个导入工作，实现了动态节点管理后，就无需这步操作了。

## test

### grpc1

- 这个是 grpc 客户端的一个示例，它提供了有限个查询测试，可以用来测试 rpc 客户端和服务器的连通性
- grpc server 提供了三种实现，分别位于
  - cmd/main.go 
  - cmd/grpc1/main.go
  - cmd/grpc2/main.go
- 在三个 terminal 下分别运行
  - go run main.go -port 9999
  - go run main.go -port 10000
  - go run main.go -port 10001
- 服务实例全部成功启动后，就可以运行这个 grpc 客户端示例进行 rpc 调用测试了
  - go run grpc_client1.go

### grpc2

- 不同于 grpc1 中的少量测试，这个测试用例中往一个数组中插入了一些学生名，然后使用 rand.Shuffle （洗牌算法）将学生名打散，从而更好模拟实际情况。
- 除非调用发生失败，否则该测试用例将会一直运行；一般经过两个阶段：
  - 预热阶段：所有查询请求结果都没有被缓存，所有请求都需要请求数据库（慢速查询），然后从数据库加载到缓存
  - 工作阶段：当缓存基本构建完成后，大部分热点请求会命中缓存直接返回，因此请求处理速度明显加快（从 README.md 提供的两个视频中可以观察到差别）
- 测试用例对 grpc server 返回的未查询到的请求做了特殊处理，假设 grpc server 从数据库中没有查询到某个学生的成绩，默认返回 "record not found"，那么客户端需要根据 grpc 返回的调用结果进行拦截处理，防止 panic
- 将 "record not found" 视作正常结果而不是错误
- 同样是想启动 grpc server 后运行 `go run grpc_client2.go`

## http
- http_test1.sh: 有限个查询用例，之所以请求 http://localhost:9999 是因为 api server 运行在 ":9999"，然后对请求进行分发（一致性哈希算法实现的负载均衡）；
- http_test2.sh: 无限迭代查询，http server 是通过 api server 实现的代理请求转发。