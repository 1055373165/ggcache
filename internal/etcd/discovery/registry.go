package discovery

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/pkg/common/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
)

// ServiceMetadata contains information about a service endpoint
type ServiceMetadata struct {
	Version     string    `json:"version"`     // Service version
	Weight      int       `json:"weight"`      // Load balancing weight
	Region      string    `json:"region"`      // Deployment region
	Environment string    `json:"environment"` // Runtime environment (e.g., prod, staging)
	StartTime   time.Time `json:"start_time"`  // Service start time
	CPU         int       `json:"cpu"`         // Number of CPU cores
	Memory      int64     `json:"memory"`      // Available memory in bytes
}

// Register adds a service endpoint to etcd with a lease-based TTL mechanism.
// It maintains the registration until either the service is stopped, the lease
// renewal fails, or the etcd connection is lost.
//
// Parameters:
//   - service: Name of the service to register
//   - addr: Address of the service endpoint (host:port)
//   - stop: Channel to signal service shutdown
//
// Returns an error if registration fails or when the service is stopped.
func Register(service string, addr string, stop chan error) error {
	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		logger.LogrusObj.Fatalf("failed to create etcd client: %v", err)
		return err
	}

	lease, err := cli.Grant(context.Background(), 5)
	if err != nil {
		return fmt.Errorf("failed to create lease: %v", err)
	}

	if err := etcdAddEndpoint(cli, lease.ID, service, addr); err != nil {
		return fmt.Errorf("failed to register endpoint: %v", err)
	}

	keepAlive, err := cli.KeepAlive(context.Background(), lease.ID)
	if err != nil {
		return fmt.Errorf("failed to keep lease alive: %v", err)
	}

	logger.LogrusObj.Debugf("registered service %s at %s", service, addr)

	for {
		select {
		case err := <-stop:
			return err
		case <-cli.Ctx().Done():
			return fmt.Errorf("etcd connection lost")
		case _, ok := <-keepAlive:
			if !ok {
				logger.LogrusObj.Error("lease keepalive channel closed")
				etcdDelEndpoint(cli, service, addr)
				return fmt.Errorf("lease expired")
			}
		default:
			time.Sleep(200 * time.Millisecond)
		}
	}
}

// etcdAddEndpoint registers a service endpoint in etcd with the given lease.
// The endpoint is stored as a key-value pair where:
// - Key: "{service}/{addr}"
// - Value: JSON-encoded endpoint information including address and metadata
func etcdAddEndpoint(client *clientv3.Client, leaseID clientv3.LeaseID, service, addr string) error {
	em, err := endpoints.NewManager(client, service)
	if err != nil {
		return err
	}

	hostname, _ := os.Hostname()
	metadata := ServiceMetadata{
		Version:     "v1.0.0",
		Weight:      10,
		Region:      os.Getenv("SERVICE_REGION"),
		Environment: os.Getenv("SERVICE_ENV"),
		StartTime:   time.Now(),
		CPU:         runtime.NumCPU(),
		Memory:      getAvailableMemory(),
	}

	endpoint := endpoints.Endpoint{
		Addr: addr,
		Metadata: map[string]interface{}{
			"metadata":    metadata,
			"hostname":    hostname,
			"health":      "healthy",
			"update_time": time.Now().Format(time.RFC3339),
		},
	}

	return em.AddEndpoint(context.TODO(),
		fmt.Sprintf("%s/%s", service, addr),
		endpoint,
		clientv3.WithLease(leaseID))
}

// etcdDelEndpoint removes a service endpoint from etcd.
func etcdDelEndpoint(client *clientv3.Client, service, addr string) error {
	em, err := endpoints.NewManager(client, service)
	if err != nil {
		return err
	}
	return em.DeleteEndpoint(client.Ctx(),
		fmt.Sprintf("%s/%s", service, addr),
		nil)
}

// getAvailableMemory returns the available system memory in bytes
func getAvailableMemory() int64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return int64(memStats.Sys)
}
