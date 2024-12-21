package discovery

import (
	"context"
	"time"

	"github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/pkg/common/logger"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Returns a client connection (established) to the specified server.
func Discovery(c *clientv3.Client, service string) (*grpc.ClientConn, error) {
	etcdResolver, err := resolver.NewBuilder(c)
	if err != nil {
		return nil, err
	}

	// Note that the name of the service here must be consistent
	// with the name of the service when it is registered.
	return grpc.NewClient("etcd:///"+service,
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`))
}

// Go to the service registration center to find a list of
// available service nodes based on the service name.
func ListServicePeers(serviceName string) ([]string, error) {
	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		logger.LogrusObj.Errorf("failed to connected to etcd, error: %v", err)
		return []string{}, err
	}

	// Endpoints are actually ip:port combinations, which can also be regarded as socket in Unix.
	// An endpoint manager stores both an etcd client object and the name of the requested service.
	endpointsManager, err := endpoints.NewManager(cli, serviceName)
	if err != nil {
		logger.LogrusObj.Errorf("create endpoints manager failed, %v", err)
		return []string{}, err
	}

	// List returns all endpoints of the current service in the form of a map.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	Key2EndpointMap, err := endpointsManager.List(ctx)
	if err != nil {
		logger.LogrusObj.Errorf("list endpoint nodes for target service failed, error: %s", err.Error())
		return []string{}, err
	}

	var peersAddr []string
	for key, endpoint := range Key2EndpointMap {
		peersAddr = append(peersAddr, endpoint.Addr) // Addr is the server address on which a connection will be established.
		logger.LogrusObj.Infof("found endpoint addr: %s (%s):(%v)", key, endpoint.Addr, endpoint.Metadata)
	}

	return peersAddr, nil
}

// DynamicServices provides the ability to dynamically build global hash views
// for the cache system and allowing for second-level view convergence.
func DynamicServices(update chan struct{}, service string) {
	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		logger.LogrusObj.Errorf("failed to connected to etcd, error: %v", err)
		return
	}
	defer cli.Close()

	// Subscription and publishing mechanism.
	// Can also be seen as an observer pattern.
	// Monitor the changes of the {service} key or KV pairs prefixed with {service},
	// and return the corresponding events, notify through the returned channel.
	watchChan := cli.Watch(context.Background(), service, clientv3.WithPrefix())

	// Each time a user adds or removes a new instance address to a given service, the watchChan backend daemon
	// can scan for changes in the number of instances via WithPrefix() and return them as watchResp.Events events.
	for watchResp := range watchChan {
		for _, ev := range watchResp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				update <- struct{}{} //When a change occurs, send a signal to update channel telling endpoint manager to rebuild the hash map.
				logger.LogrusObj.Warnf("Service endpoint added or updated: %s", string(ev.Kv.Value))
			case clientv3.EventTypeDelete:
				update <- struct{}{} //When a change occurs, send a signal to update channel telling endpoint manager to rebuild the hash map.
				logger.LogrusObj.Warnf("Service endpoint removed: %s", string(ev.Kv.Key))
			}
		}
	}
}
