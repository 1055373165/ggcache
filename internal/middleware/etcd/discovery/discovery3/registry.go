package discovery3

import (
	"context"
	"fmt"
	"time"

	"github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/utils/logger"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
)

// service registration
func Register(service string, addr string, stop chan error) error {
	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		logger.LogrusObj.Fatalf("err: %v", err)
		return err
	}

	//  creates a new lease
	resp, err := cli.Grant(context.Background(), 5)
	if err != nil {
		return fmt.Errorf("grant creates a new lease failed: %v", err)
	}

	/*
		Resp is an wrap of the lease response. The core fields include ttl and lease Id.
		TTL: TimeToLive which means lease remaining validity time. entrie are removed from etcd after expiration.
		ID: unique identifier of the lease
	*/
	leaseId := resp.ID
	// associated service address and lease, lease expiration
	// the service address information is deleted from etcd when the lease expires
	err = etcdAdd(cli, leaseId, service, addr)
	if err != nil {
		return fmt.Errorf("failed to add services as endpoint to etcd endpoint Manager: %v", err)
	}

	// KeepAlive attempts to keep the given lease alive forever
	// ch used to receive lease keepalive responses
	ch, err := cli.KeepAlive(context.Background(), leaseId)
	if err != nil {
		return fmt.Errorf("set keepalive for lease failed: %v", err)
	}

	logger.LogrusObj.Debugf("[%s] register service success\n", addr)

	for {
		select {
		case err := <-stop: // service revocation signal
			etcdDel(cli, service, addr)
			if err != nil {
				logger.LogrusObj.Error(err.Error())
			}
			return err
		case <-cli.Ctx().Done(): // etcd client connect 断开
			return fmt.Errorf("etcd client connect broken")
		case _, ok := <-ch: // lease keepalive responses
			if !ok {
				logger.LogrusObj.Error("keepalive channel closed, revoke given lease") // 比如 etcd 断开服务，通知 server 停止
				etcdDel(cli, service, addr)
				return fmt.Errorf("keepalive channel closed, revoke given lease") // 返回非 nil 的 error，上层就会关闭 stopsChan 从而关闭 server
			}
		default:
			time.Sleep(200 * time.Millisecond)
		}
	}
}

/*
stored in etcd in the form of key - value,
the form of key is service/addr,
the form of value is endpoint{addr, metadata}.
*/
func etcdAdd(client *clientv3.Client, leaseId clientv3.LeaseID, service string, addr string) error {
	endPointsManager, err := endpoints.NewManager(client, service)
	if err != nil {
		return err
	}
	return endPointsManager.AddEndpoint(context.TODO(),
		fmt.Sprintf("%s/%s", service, addr),
		/*
			Addr is the server address on which a connection will be established.
			Metadata is the information associated with Addr, which may be used to make load balancing decision.
			Endpoint represents a single address the connection can be established with.
		*/
		endpoints.Endpoint{Addr: addr, Metadata: "ggcache services"},
		clientv3.WithLease(leaseId))
}

func etcdDel(client *clientv3.Client, service string, addr string) error {
	endPointsManager, err := endpoints.NewManager(client, service)
	if err != nil {
		return err
	}
	return endPointsManager.DeleteEndpoint(client.Ctx(),
		fmt.Sprintf("%s/%s", service, addr), nil)
}
