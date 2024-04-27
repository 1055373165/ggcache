# 测试介绍

- 已经实现全自动节点管理，拥有节点动态增删和一致性视图重构能力，无需先往 etcd 中导入实例地址

## test

前置：一定要先运行 etcd 集群，在 internal/middleware/cluster 有详细介绍（先将本地的 2379 etcd 服务停掉）

### grpc test

- 在三个 terminal 下分别运行
  - go run main.go -port 9999
  - go run main.go -port 10000
  - go run main.go -port 10001

- 服务实例全部成功启动后，就可以运行 grpc 客户端示例进行 rpc 调用测试了

在实际的测试用例中：往一个数组中插入了一些学生名，然后使用 rand.Shuffle （洗牌算法）将学生名打散，从而更好模拟实际查询情况。
- 除非调用发生失败，否则该测试用例将会一直运行；一般经过两个阶段：
  - 预热阶段：所有查询请求结果都没有被缓存，所有请求都需要请求数据库（慢速查询），然后从数据库加载到缓存
  - 工作阶段：当缓存基本构建完成后，大部分热点请求会命中缓存直接返回，因此请求处理速度明显加快
- 测试用例对 grpc server 返回的未查询到的请求做了特殊处理，假设 grpc server 从数据库中没有查询到某个学生的成绩，默认返回 "record not found"，那么客户端需要根据 grpc 返回的调用结果进行拦截处理，防止 panic
- 增加了重试机制（在服务器崩溃恢复后可以立即获取响应）
- 将 "record not found" 视作正常结果而不是错误

运行：go run /test/grpc/grpc_client.go 进行测试，也可以在完成了上述操作后执行 ./test3.sh 进行测试

### http
- http_test1.sh: 有限个查询用例，之所以请求 http://localhost:9999 是因为 api server 运行在 ":9999"，然后对请求进行分发（一致性哈希算法实现的负载均衡）；
- http_test2.sh: 无限迭代查询，http server 是通过 api server 实现的代理请求转发。
