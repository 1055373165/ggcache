// Package cache implements a distributed cache system with various features.
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/1055373165/ggcache/internal/middleware/etcd/discovery/discovery3"
	"github.com/1055373165/ggcache/api/groupcachepb"
	"github.com/1055373165/ggcache/utils/logger"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Client implements the Fetcher interface using gRPC.
type Client struct {
	serviceName string // Name of the service to fetch from
}

// NewClient creates a new gRPC client for fetching cache values.
func NewClient(service string) *Client {
	return &Client{serviceName: service}
}

// Fetch retrieves a cache value from a remote peer using gRPC.
//
// Parameters:
//   - group: The cache group name
//   - key: The key to fetch
//
// Returns:
//   - []byte: The fetched value
//   - error: Any error encountered during the fetch
//
// The function performs service discovery using etcd, establishes a gRPC
// connection to the selected peer, and makes a Get request with timeout.
func (c *Client) Fetch(group string, key string) ([]byte, error) {
	// Create etcd client
	cli, err := clientv3.NewFromURL("http://localhost:2379")
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	// Discover service and establish connection
	start := time.Now()
	conn, err := discovery3.Discovery(cli, c.serviceName)
	logger.LogrusObj.Warnf("gRPC dial took: %v ms", time.Since(start).Milliseconds())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Create gRPC client and context with timeout
	grpcClient := groupcachepb.NewGroupCacheClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Make gRPC call
	start = time.Now()
	resp, err := grpcClient.Get(ctx, &groupcachepb.GetRequest{
		Group: group,
		Key:   key,
	})
	logger.LogrusObj.Warnf("gRPC call took: %v ms", time.Since(start).Milliseconds())
	if err != nil {
		return nil, fmt.Errorf("failed to get %s/%s from peer %s: %v", group, key, c.serviceName, err)
	}

	return resp.Value, nil
}
