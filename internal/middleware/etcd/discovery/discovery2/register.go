package discovery2

import (
	"context"
	"fmt"

	"github.com/1055373165/ggcache/config"

	"github.com/1055373165/ggcache/utils/logger"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
)

// Register registers a service to etcd
// Note that Register will not return (if there is no error)
func Register(service string, addr string, stop chan error) error {
	cli, err := clientv3.New(config.DefaultEtcdConfig)
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
	ch1, err := cli.KeepAlive(context.Background(), leaseId)
	if err != nil {
		return fmt.Errorf("set keepalive failed: %v", err)
	}

	manager, _ := endpoints.NewManager(cli, service)
	ch2, _ := manager.NewWatchChannel(context.Background())

	for {
		select {
		case err := <-stop:
			if err != nil {
				logger.LogrusObj.Errorf(err.Error())
			}
			return err
		case <-cli.Ctx().Done():
			logger.LogrusObj.Infof("service closed")
			return nil
		case _, ok := <-ch1: // Listen for lease revocation signals
			if !ok {
				logger.LogrusObj.Info("keepalive channel closed")
				_, err := cli.Revoke(context.Background(), leaseId)
				return err
			}
		case <-ch2:
			// map[string]Endpoint
			key2EndpointMap, _ := manager.List(context.Background())
			var addrs []string
			for _, endpoint := range key2EndpointMap {
				addrs = append(addrs, endpoint.Addr)
			}

		}
	}
}

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
