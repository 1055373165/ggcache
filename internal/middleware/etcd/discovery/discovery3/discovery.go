package discovery3

import (
	"context"
	"math/rand"

	"github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/utils/logger"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// create a client connection to the given service
func Discovery(c *clientv3.Client, service string) (*grpc.ClientConn, error) {
	etcdResolver, err := resolver.NewBuilder(c)
	if err != nil {
		return nil, err
	}

	return grpc.Dial(
		"etcd:///"+service,
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
}

// find the list of available service nodes based on the service name
func ListServicePeers(serviceName string) ([]string, error) {
	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		logger.LogrusObj.Errorf("failed to connected to etcd, error: %v", err)
		return []string{}, err
	}

	endPointsManager, err := endpoints.NewManager(cli, serviceName)
	if err != nil {
		logger.LogrusObj.Errorf("create endpoints manager failed, %v", err)
		return []string{}, err
	}

	Key2EndpointMap, err := endPointsManager.List(context.Background())
	if err != nil {
		logger.LogrusObj.Errorf("enpoint manager list op failed, %v", err)
		return []string{}, err
	}

	var peers []string
	for key, endpoint := range Key2EndpointMap {
		peers = append(peers, endpoint.Addr)
		logger.LogrusObj.Infof("found endpoint %s (%s):(%s)", key, endpoint.Addr, endpoint.Metadata)
	}

	return peers, nil
}

func DynamicServices(update chan bool, service string) {
	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		logger.LogrusObj.Errorf("failed to connected to etcd, error: %v", err)
		return
	}
	defer cli.Close()

	// Subscription and publishing mechanism
	watchChan := cli.Watch(context.Background(), service, clientv3.WithPrefix())

	// 每次用户往指定的服务中添加或者删除新的实例地址时，watchChan 后台都能通过 WithPrefix() 扫描到实例数量的变化并以  watchResp.Events 事件的方式返回
	// 当发生变更时，往 update channel 发送一个信号，告知 endpoint manager 重新构建哈希映射
	for watchResp := range watchChan {
		for _, ev := range watchResp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				update <- true // 通知 endpoint manager 重新构建节点视图
				logger.LogrusObj.Warnf("Service endpoint added or updated: %s", string(ev.Kv.Value))
			case clientv3.EventTypeDelete:
				update <- true // 通知 endpoint manager 重新构建节点视图
				logger.LogrusObj.Warnf("Service endpoint removed: %s", string(ev.Kv.Key))
			}
		}
	}
}

func shuffle(peers []string) string {
	rand.Shuffle(len(peers), func(i, j int) {
		peers[i], peers[j] = peers[j], peers[i]
	})
	return peers[len(peers)/2]
}

// func EtcdDial(c *clientv3.Client, service string) (*grpc.ClientConn, error) {
// 	// NewBuilder 创建一个解析器生成器。用于解析客户端发来的请求路径，从而确认要连接的对象
// 	etcdResolver, err := resolver.NewBuilder(c)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Dial 创建到给定目标的客户端连接（有了通道就可以建立与服务端的连接了）
// 	// WithResolvers 允许在 ClientConn 本地注册一系列解析器实现，而无需通过 resolver.Register 进行全局注册。
// 	// 它们将仅与当前 Dial 使用的方案进行匹配，并优先于全局注册。
// 	return grpc.Dial(
// 		"etcd:///"+service,
// 		grpc.WithResolvers(etcdResolver),
// 		grpc.WithTransportCredentials(insecure.NewCredentials()),
// 		grpc.WithBlock(), //阻塞等待直至连接 up
// 	)
// }
