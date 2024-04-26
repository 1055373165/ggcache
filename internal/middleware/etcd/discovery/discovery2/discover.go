package discovery2

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*
ClientConn 表示与概念端点的虚拟连接，用于执行 RPC，ClientConn 可根据配置、负载等情况，与端点自由建立零个或多个实际连接。
*/
func EtcdDial(c *clientv3.Client, service string) (*grpc.ClientConn, error) {
	// NewBuilder creates a parser builder. It is used to parse the request path sent by the client to identify the object to connect to
	etcdResolver, err := resolver.NewBuilder(c)
	if err != nil {
		return nil, err
	}

	return grpc.Dial(
		"etcd:///"+service,
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), // Block and wait until connected
	)
}
