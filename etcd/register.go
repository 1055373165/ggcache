package etcd

import (
	"context"
	"fmt"
	"time"

	"github.com/1055373165/Distributed_KV_Store/logger"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
)

// The Register module provides the ability to register services to etcd
var (
	DefaultEtcdConfig = clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	}
)

// Register registers a service to etcd
// Note that Register will not return (if there is no error)
func Register(service string, addr string, stop chan error) error {
	cli, err := clientv3.New(DefaultEtcdConfig)
	if err != nil {
		return fmt.Errorf("create etcd client falied: %v", err)
	}
	defer cli.Close()

	resp, err := cli.Grant(context.Background(), 5)
	if err != nil {
		return fmt.Errorf("create lease failed: %v", err)
	}
	leaseId := resp.ID

	// Associate service address with lease
	err = etcdAdd(cli, leaseId, service, addr)
	if err != nil {
		panic(err)
	}

	// Set up keep-alive heartbeat detection
	ch, err := cli.KeepAlive(context.Background(), leaseId)
	if err != nil {
		return fmt.Errorf("set keepalive failed: %v", err)
	}

	for {
		select {
		case err := <-stop:
			if err != nil {
				logger.Logger.Errorf(err.Error())
			}
			return err
		case <-cli.Ctx().Done():
			logger.Logger.Infof("service closed")
			return nil
		case _, ok := <-ch: // Listen for lease revocation signals
			if !ok {
				logger.Logger.Info("keepalive channel closed")
				_, err := cli.Revoke(context.Background(), leaseId)
				return err
			}
		}
	}
}

// etcdAdd Adds a pair of kvs to etcd in lease mode.
func etcdAdd(client *clientv3.Client, leaseId clientv3.LeaseID, service string, addr string) error {
	em, err := endpoints.NewManager(client, service)
	if err != nil {
		return err
	}

	// key: groupcache+"/"+addr value: endpoints.Endpoint{Addr: addr} bind leaseId
	return em.AddEndpoint(client.Ctx(), fmt.Sprintf("%s/%s", service, addr), endpoints.Endpoint{Addr: addr, Metadata: "groupcache services"}, clientv3.WithLease(leaseId))
}
