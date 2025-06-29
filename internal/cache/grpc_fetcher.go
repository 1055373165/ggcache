package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "github.com/1055373165/ggcache/api/groupcachepb"
	"github.com/1055373165/ggcache/pkg/common/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ Fetcher = (*grpcFetcher)(nil)

type grpcFetcher struct {
	peerAddr string
	conn     *grpc.ClientConn
	mu       sync.RWMutex // protect connection state
}

func NewGrpcFetcher(peerAddr string) *grpcFetcher {
	conn, err := grpc.NewClient(peerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.LogrusObj.Errorf("failed to dial peer %s: %v", peerAddr, err)
		return nil
	}
	return &grpcFetcher{
		peerAddr: peerAddr,
		conn:     conn,
	}
}

// Fetch gets the corresponding cache value from remote peer
func (c *grpcFetcher) Fetch(group string, key string) ([]byte, error) {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()
	if conn == nil {
		return nil, fmt.Errorf("no connection to peer %s", c.peerAddr)
	}

	logger.LogrusObj.Infof("fetching %s/%s from peer %s reusing connection", group, key, c.peerAddr)

	grpcClient := pb.NewGroupCacheClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	start := time.Now()
	resp, err := grpcClient.Get(ctx, &pb.GetRequest{
		Group: group,
		Key:   key,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get %s/%s from peer %s", group, key, c.peerAddr)
	}

	logger.LogrusObj.Debugf("the duration of this grpc Call is: %v ms", time.Since(start).Milliseconds())

	return resp.Value, nil
}

func (c *grpcFetcher) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
