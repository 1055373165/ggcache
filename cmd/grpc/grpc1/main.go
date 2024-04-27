package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/1055373165/ggcache/api/groupcachepb"
	"github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/internal/middleware/etcd"
	"github.com/1055373165/ggcache/internal/middleware/etcd/discovery/discovery1"
	"github.com/1055373165/ggcache/internal/pkg/student/dao"
	"github.com/1055373165/ggcache/internal/service"
	"github.com/1055373165/ggcache/utils/logger"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 9999, "listen port")
)

func main() {
	config.InitConfig()
	dao.InitDB()
	flag.Parse()
	addr := fmt.Sprintf("localhost:%d", *port)

	groupManager := service.NewGroupManager([]string{"scores", "student"}, addr)

	updateChan := make(chan bool)
	svr, err := service.NewServer(updateChan, addr)
	if err != nil {
		panic(err)
	}
	// put clusters/nodeadress nodeadress
	addrs, err := etcd.GetPeers("clusters")
	if err != nil {
		addrs = []string{"localhost:9999"}
	}
	svr.SetPeers(addrs)
	groupManager["scores"].RegisterServer(svr)

	Start(svr)
}

func UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		logger.LogrusObj.Warnf("call %s", info.FullMethod)
		resp, err = handler(ctx, req)
		return resp, err
	}
}

func Start(svr *service.Server) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		s := <-ch
		discovery1.EtcdUnRegister(svr.Addr)
		if i, ok := s.(syscall.Signal); ok {
			os.Exit(int(i))
		} else {
			os.Exit(0)
		}
	}()

	err := discovery1.EtcdRegister(svr.Addr)
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(UnaryInterceptor()))
	pb.RegisterGroupCacheServer(grpcServer, svr)

	lis, err := net.Listen("tcp", svr.Addr)
	if err != nil {
		panic(err)
	}

	logger.LogrusObj.Debugf("ggcache service is running at ", svr.Addr)
	if err := grpcServer.Serve(lis); err != nil {
		panic(err)
	}
}
