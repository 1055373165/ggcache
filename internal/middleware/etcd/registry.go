package etcd

import (
	"context"
	"fmt"
	"time"

	"ggcache/internal/middleware/logger"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
)

var (
	DefaultEtcdConfig = clientv3.Config{
		Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: 5 * time.Second,
	}
)

// service registration
func Register(service string, addr string, stop chan error) error {
	cli, err := clientv3.New(DefaultEtcdConfig)
	if err != nil {
		return fmt.Errorf("create etcd client v3 failed: %v", err)
	}
	defer cli.Close()

	//  creates a new lease
	resp, err := cli.Grant(context.Background(), 5)
	if err != nil {
		return fmt.Errorf("grant 5s lease failed: %v", err)
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

	logger.Logger.Infof("[%s] register service success\n", addr)

	for {
		select {
		case err := <-stop: // service revocation signal
			if err != nil {
				etcdDel(cli, service, addr)
				logger.Logger.Error(err.Error())
			}
			return err
		case <-cli.Ctx().Done():
			logger.Logger.Infof("node %v stopped the service %v", addr, service)
			etcdDel(cli, service, addr)
			return nil
		case _, ok := <-ch: // lease keepalive responses
			if !ok {
				logger.Logger.Error("keepalive channel closed, revoke given lease")
				_, err := cli.Revoke(context.Background(), leaseId)
				etcdDel(cli, service, addr)
				return err
			}
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
	return endPointsManager.AddEndpoint(client.Ctx(),
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

func RegisterPeerToEtcd(serviceName, addr string) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   DefaultEtcdConfig.Endpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		fmt.Println("new clientv3 failed,err:", err)
		return
	}
	defer cli.Close()

	fmt.Println("connect to etcd success!")
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// key {environment/zone/servicename}
	// value {service peer address}
	_, err = cli.Put(ctx, fmt.Sprintf("%s/%s", serviceName, addr), addr)
	if err != nil {
		logger.Logger.Errorf("put service peer address %s to etcd failed, %v", addr, err)
	}
	logger.Logger.Info("put node address to etcd done")
}
