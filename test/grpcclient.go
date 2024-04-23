package main

import (
	"context"
	"fmt"
	"math/rand"

	pb "ggcache/api"
	conf "ggcache/configs"

	etcdservice "ggcache/internal/middleware/etcd"
	"ggcache/internal/middleware/logger"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	group   = "scores"
	service = "ggcache"
)

func main() {
	conf.Init()
	// get the client connection to the specified service, also called grpc channel
	peers := etcdservice.ListServicePeers(service)
	if len(peers) == 0 {
		logger.Logger.Fatalf("no available service node")
		return
	}

	// var conn *grpc.ClientConn
	// var err error
	// for {
	// 	conn, err = grpc.Dial(shuffle(peers), grpc.WithTransportCredentials(insecure.NewCredentials()))
	// 	if err == nil {
	// 		break
	// 	}
	// }
	// defer conn.Close()
	logger.Logger.Info("...into fetcher Fetch")
	cli, err := clientv3.New(etcdservice.DefaultEtcdConfig)
	if err != nil {
		logger.Logger.Errorf("[Fetch] cli new, %v", err)
		return
	}
	defer cli.Close()

	conn, err := etcdservice.Discovery(cli, service)
	if err != nil {
		logger.Logger.Errorf("[Fetch] Discover, %v", err)
	}
	defer conn.Close()

	// the client stub implements the same methods as the grpc server
	clientStub := pb.NewGGCacheClient(conn)

	// populate requests in protobuf format
	names := []string{"张三", "李四", "one", "two", "1", "2", "3", "4", "5", "6", "7"}
	for _, name := range names {
		resp, err := clientStub.Get(context.Background(), &pb.GetRequest{Key: name, Group: group})
		if err != nil {
			fmt.Printf("failed to get %s's score via rpc request\n", name)
		} else {
			fmt.Printf("the rpc call succeeds, and the score of %s is %s\n", name, string(resp.GetValue()))
		}
	}
}

func shuffle(peers []string) string {
	rand.Shuffle(len(peers), func(i, j int) {
		peers[i], peers[j] = peers[j], peers[i]
	})
	return peers[len(peers)/2]
}
