package discovery

import (
	"context"

	"github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/pkg/common/logger"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Discovery creates a gRPC client connection to a service using etcd for service discovery.
// It takes an etcd client and a service name, returning a connection that can be used
// to make gRPC calls to any instance of the service.
func Discovery(c *clientv3.Client, service string) (*grpc.ClientConn, error) {
	etcdResolver, err := resolver.NewBuilder(c)
	if err != nil {
		return nil, err
	}

	return grpc.Dial("etcd:///"+service,
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
}

// ListServicePeers retrieves all available endpoints for a given service from etcd.
// It returns a list of address strings in the format "host:port" that can be used
// to connect to the service instances.
func ListServicePeers(serviceName string) ([]string, error) {
	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		logger.LogrusObj.Errorf("failed to connected to etcd: %v", err)
		return nil, err
	}

	em, err := endpoints.NewManager(cli, serviceName)
	if err != nil {
		logger.LogrusObj.Errorf("failed to create endpoints manager: %v", err)
		return nil, err
	}

	eps, err := em.List(context.Background())
	if err != nil {
		logger.LogrusObj.Errorf("failed to list endpoints: %v", err)
		return nil, err
	}

	var peers []string
	for key, ep := range eps {
		peers = append(peers, ep.Addr)
		logger.LogrusObj.Infof("found endpoint: %s at %s with metadata: %v", key, ep.Addr, ep.Metadata)
	}

	return peers, nil
}

// DynamicServices monitors service endpoint changes in etcd and notifies through a channel.
// It watches for any changes (additions, updates, or deletions) to service endpoints and
// sends a notification through the update channel when changes occur.
//
// Parameters:
//   - update: Channel to send notifications when service endpoints change
//   - service: Name of the service to monitor
//
// The function runs indefinitely until the etcd client connection is closed.
func DynamicServices(update chan struct{}, service string) {
	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		logger.LogrusObj.Errorf("failed to connect to etcd: %v", err)
		return
	}
	defer cli.Close()

	watchChan := cli.Watch(context.Background(), service, clientv3.WithPrefix())
	for resp := range watchChan {
		for _, ev := range resp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				update <- struct{}{} // Notify about endpoint addition/update
				logger.LogrusObj.Infof("endpoint added/updated: %s", string(ev.Kv.Value))
			case clientv3.EventTypeDelete:
				update <- struct{}{} // Notify about endpoint removal
				logger.LogrusObj.Infof("endpoint removed: %s", string(ev.Kv.Key))
			}
		}
	}
}
