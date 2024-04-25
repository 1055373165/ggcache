package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/1055373165/Distributed_KV_Store/conf"
	kv "github.com/1055373165/Distributed_KV_Store/distributekv"
	"github.com/1055373165/Distributed_KV_Store/etcd"
	pb "github.com/1055373165/Distributed_KV_Store/grpc/groupcachepb"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 9999, "listen port")
)

func main() {
	conf.Init()
	flag.Parse()
	addr := fmt.Sprintf("localhost:%d", *port)

	g := kv.NewGroupInstance("scores")
	svr, err := kv.NewServer(addr)
	if err != nil {
		panic(err)
	}
	// put clusters/nodeadress nodeadress
	addrs, err := etcd.GetPeers("clusters")
	if err != nil {
		addrs = []string{"localhost:9999"}
	}
	svr.SetPeers(addrs)
	g.RegisterServer(svr)

	Start(svr)
}

func UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		log.Printf("call %s\n", info.FullMethod)
		resp, err = handler(ctx, req)
		return resp, err
	}
}

func Start(svr *kv.Server) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		s := <-ch
		etcdUnRegister(svr.Addr)
		if i, ok := s.(syscall.Signal); ok {
			os.Exit(int(i))
		} else {
			os.Exit(0)
		}
	}()

	err := etcdRegister(svr.Addr)
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(UnaryInterceptor()))
	pb.RegisterGroupCacheServer(grpcServer, svr)

	lis, err := net.Listen("tcp", svr.Addr)
	if err != nil {
		panic(err)
	}

	log.Println("groupcache is running at ", svr.Addr)
	if err := grpcServer.Serve(lis); err != nil {
		panic(err)
	}
}
