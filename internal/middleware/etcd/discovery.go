package etcd

import (
	"context"
	"ggcache/internal/middleware/logger"
	"log"
	"math/rand"

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
		// grpc.WithBlock(),
	)
}

// find the list of available service nodes based on the service name
func ListServicePeers(serviceName string) []string {
	cli, err := clientv3.New(DefaultEtcdConfig)
	if err != nil {
		logger.Logger.Errorf("failed to connect etcd client, %v", err)
		return []string{}
	}
	defer cli.Close()

	endPointsManager, err := endpoints.NewManager(cli, serviceName)
	if err != nil {
		logger.Logger.Errorf("create endpoints manager failed, %v", err)
		return []string{}
	}

	Key2EndpointMap, err := endPointsManager.List(cli.Ctx())
	if err != nil {
		logger.Logger.Errorf("enpoint manager list op failed, %v", err)
		return []string{}
	}

	var peers []string
	for key, endpoint := range Key2EndpointMap {
		peers = append(peers, endpoint.Addr)
		logger.Logger.Infof("found endpoint %s (%s):(%s)", key, endpoint.Addr, endpoint.Metadata)
	}

	return peers
}

func DynamicServices(update chan bool, service string) {
	client, err := clientv3.New(DefaultEtcdConfig)
	if err != nil {
		logger.Logger.Fatalf("create etcd client v3 failed: %v", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	watchChan := client.Watch(ctx, service, clientv3.WithPrefix())

	for watchResp := range watchChan {
		for _, ev := range watchResp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				update <- true
				logger.Logger.Infof("Service endpoint added or updated: %s", string(ev.Kv.Value))

			case clientv3.EventTypeDelete:
				update <- true
				log.Printf("Service endpoint removed: %s", string(ev.Kv.Key))
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
