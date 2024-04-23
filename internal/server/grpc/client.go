package grpc

import (
	"context"
	"fmt"
	pb "ggcache/api"
	etcdservice "ggcache/internal/middleware/etcd"
	"ggcache/internal/middleware/logger"
	"ggcache/internal/service"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var _ service.Fetcher = (*grpcFetcher)(nil)

/*
1. grpcFetcher implements the ability to fetch cached values from the node providing the specified service via grpc protocol.
2. like httpFetcher, grpcFetcher implements the Fetcher interface.
*/
type grpcFetcher struct {
	serviceName string
}

func (c *grpcFetcher) Fetch(groupName string, key string) ([]byte, error) {
	logger.Logger.Info("...into fetcher Fetch")
	cli, err := clientv3.New(etcdservice.DefaultEtcdConfig)
	if err != nil {
		logger.Logger.Errorf("[Fetch] cli new, %v", err)
		return nil, err
	}
	defer cli.Close()
	logger.Logger.Info("...into fetcher Fetch Discovery")
	// get the client connection to the specified service, also called grpc channel
	conn, err := etcdservice.Discovery(cli, c.serviceName)
	logger.Logger.Info("...into fetcher Fetch Discovery end")
	if err != nil {
		logger.Logger.Errorf("[Fetch] discover, %v", err)
		return nil, err
	}
	defer conn.Close()

	logger.Logger.Infof("discover success, create client stub")
	clientStub := pb.NewGGCacheClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// clientStub mplements the same methods as the service
	resp, err := clientStub.Get(ctx, &pb.GetRequest{
		Group: groupName,
		Key:   key,
	})

	if err != nil {
		return nil, fmt.Errorf("get %s/%s from service peer %s failed: %v", groupName, key, c.serviceName, err)
	}

	return resp.Value, nil
}

func NewGrpcFetcher(serviceNameWithPrefix string) *grpcFetcher {
	return &grpcFetcher{serviceName: serviceNameWithPrefix}
}
