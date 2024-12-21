package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/pkg/common/logger"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
)

// Register registers the {addr} for the specified {service}. During normal service provision, this function will not return.
// This is returned only when the 1. application is stopped 2. the lease renewal fails 3. the etcd connection is lost.
func Register(service string, addr string, stop chan error) error {
	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		logger.LogrusObj.Fatalf("err: %v", err)
		return err
	}

	//  Create a lease with a timeout of 5 seconds.
	leaseGrantResp, err := cli.Grant(context.Background(), 5)
	if err != nil {
		return fmt.Errorf("grant creates a new lease failed: %v", err)
	}

	// leaseGrantResp is an wrap of the lease response. The core fields include ttl and lease Id.
	// - TTL: TimeToLive which means lease remaining validity time. entrie are removed from etcd after expiration.
	// - ID: unique identifier of the lease.
	leaseId := leaseGrantResp.ID

	// Associate the service address with the lease and delete the service address information from etcd when the lease expires.
	// If a service address wants to continue to provide services, it needs to renew the lease, which is also called lease keepalive.
	err = etcdAddEndpoint(cli, leaseId, service, addr)
	if err != nil {
		return fmt.Errorf("failed to add services as endpoint to etcd endpoint Manager: %v", err)
	}

	// KeepAlive attempts to keep the given lease alive forever.
	// Each time a connected client receives a response from the keepalive channel,
	// it can assume that the server has completed the renewal.
	alive, err := cli.KeepAlive(context.Background(), leaseId)
	if err != nil {
		return fmt.Errorf("set keepalive for lease failed: %v", err)
	}

	// During the lease period, the server corresponding to addr can provide services normally.
	logger.LogrusObj.Debugf("[%s] register service success", addr)

	for {
		select {
		case err := <-stop: // Application-level stop signal.
			return err
		case <-cli.Ctx().Done(): // Etcd client broken.
			return fmt.Errorf("etcd client connect broken")
		case _, ok := <-alive: // Lease keepalive response.
			if !ok {
				logger.LogrusObj.Error("keepalive channel closed, revoke given lease")
				// Delete the endpoint from etcd
				if err := etcdDelEndpoint(cli, service, addr); err != nil {
					logger.LogrusObj.Errorf("Failed to delete endpoint: %v", err)
				}
				return fmt.Errorf("keepalive channel closed, revoke given lease") //If a non-nil error is returned, the upper layer will close stopsChan  which in turn shuts down the server.
			}
		default:
			time.Sleep(2 * time.Second) // Prevent excessive resource usage.
		}
	}
}

// The registration information for the service endpoint is stored in etcd as a key value.
// the form of key is {service}/{addr},
// the form of value is {addr, metadata}.
func etcdAddEndpoint(client *clientv3.Client, leaseId clientv3.LeaseID, service string, addr string) error {
	endpointsManager, err := endpoints.NewManager(client, service)
	if err != nil {
		return err
	}

	// Use string metadata to ensure comparability.
	metadata := endpoints.Endpoint{
		Addr:     addr,
		Metadata: "weight:10;version:v1.0.0",
	}

	// Addr is the server address on which a connection will be established.
	// Metadata is the information associated with Addr, which may be used to make load balancing decision.
	// Endpoint represents a single address the connection can be established with.
	return endpointsManager.AddEndpoint(context.TODO(),
		fmt.Sprintf("%s/%s", service, addr),
		metadata,
		clientv3.WithLease(leaseId))
}

func etcdDelEndpoint(client *clientv3.Client, service string, addr string) error {
	endpointsManager, err := endpoints.NewManager(client, service)
	if err != nil {
		return err
	}
	// Delete endpoint by Key({service}/{addr}).
	return endpointsManager.DeleteEndpoint(client.Ctx(),
		fmt.Sprintf("%s/%s", service, addr), nil)
}
