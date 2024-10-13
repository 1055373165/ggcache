# 项目介绍

已经实现全自动节点管理，拥有节点动态增删和一致性视图重构能力，无需先往 etcd 中导入实例地址

## 测试

前置

1. 运行 etcd 集群，参考 etcd/cluster/runcluster.md；
2. 初始化测试需要使用的 student 服务的数据库，参考 test/sql/create_sql.md
3. ggcache 服务和 student 服务通过 grpc 进行进程间通信，参考 api/generate.md

### 服务启动

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

### 测试 grpc 通信

1. 可以直接在项目根目录运行 `go run test/client/grpc_client.go` 进行测试
2. 或者直接进入 test/client 目录下，运行 `./client.sh` 进行测试

