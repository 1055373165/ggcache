#  Local multi-member cluster

在 etcd git 仓库的基础上提供了一个 Procfile，用于轻松配置本地多成员集群。

安装 goreman 以控制基于 Procfile 的应用程序：

## install goreman

> go install github.com/mattn/goreman@latest

使用 etcd 的 Procfile 启动一个集群。

依赖环境安装（以 macos 为例）
```
brew install etcd 
brew services stop etcd
sudo cp /opt/homebrew/bin/etcd /usr/local/bin/
rm -rf ./infra*
```

在 cluster 目录下运行 

> goreman -f Procfile start

三个节点开始运行。

它们分别通过 localhost: 2379、localhost:22379 和 localhost:32379 监听客户端请求。

## Interacting with the cluster

使用etcdctl与正在运行的集群交互：

打印成员列表：

> etcdctl --write-out=table --endpoints=localhost:12379 member list

+------------------+---------+--------+------------------------+------------------------+
|        ID        | STATUS  |  NAME  |       PEER ADDRS       |      CLIENT ADDRS      |
+------------------+---------+--------+------------------------+------------------------+
| 8211f1d0f64f3269 | started | infra1 | http://127.0.0.1:2380  | http://127.0.0.1:2379  |
| 91bc3c398fb3c146 | started | infra2 | http://127.0.0.1:22380 | http://127.0.0.1:22379 |
| fd422379fda50e48 | started | infra3 | http://127.0.0.1:32380 | http://127.0.0.1:32379 |
+------------------+---------+--------+------------------------+------------------------+

## Store an key-value pair in the cluster

> etcdctl put foo bar

如果打印OK，则存储键值对成功。

## 测试容错性

杀死其中一个节点后向这个节点请求 key：

> goreman run stop etcd2

往集群节点中存入 key-value（正常，因为还有两个正常节点）

> etcdctl put key hello

尝试从集群获取 key 的 value 值（正常）

> etcdctl get key

hello

尝试从已经停止的 node 获取值：

> etcdctl --endpoints=localhost:22379 get key

2023/06/18 23:07:35 grpc: Conn.resetTransport failed to create client transport: connection error: desc = "transport: dial tcp 127.0.0.1:22379: getsockopt: connection refused"; Reconnecting to "localhost:22379"
Error:  grpc: timed out trying to connect

重启已经停止的 node：

> goreman run restart etcd2

从重新启动的成员中获取 key 的值：

> etcdctl --endpoints=localhost:22379 get key

hello