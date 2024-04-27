#! /bin/bash


# 废弃：v3 版本无需执行改脚本
# 所有测试脚本均默认在 script 目录下运行
# test1.sh and test2.sh and test3.sh
# 按顺序执行 test1 -> test2 -> test3

# Save grpc server address to etcd cluster
./prepare/exec1.sh

# 如何 debug?
# etcdctl get "" --prefix 查询 put 结果（clusters 开头是为了拿到所有节点的地址列表，做负载均衡使用）
# 而在实际实现时，grpc.Dial 解析（反序列化）的是结构体形式，是以 endpointer 端点形式导入 etcd 的，服务注册对应的 key 的 prefix 是 "GroupCache" 和 pb 中定义的服务名字一致
# 如果出现解析报错，使用 etcdctl del "conflict key" 将错误的 key 删除即可

