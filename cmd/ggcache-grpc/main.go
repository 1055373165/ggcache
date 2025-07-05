package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/1055373165/ggcache/configs"
	"github.com/1055373165/ggcache/internal/bussiness/student/dao"
	"github.com/1055373165/ggcache/internal/cache/core"
	grpctransport "github.com/1055373165/ggcache/internal/cache/transport/grpc"
	"github.com/1055373165/ggcache/pkg/common/logger"
	"github.com/1055373165/ggcache/pkg/etcd/discovery"
)

var (
	port = flag.Int("port", 9999, "service node port")
)

func main() {
	configs.InitConfig()
	if err := dao.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	flag.Parse()

	// grpc node local service address
	serviceAddr := fmt.Sprintf("localhost:%d", *port)
	gm := core.NewGroupManager([]string{"scores", "website"}, serviceAddr)

	// get a grpc service instance
	updateChan := make(chan struct{})
	svr, err := grpctransport.NewServer(updateChan, serviceAddr)
	if err != nil {
		logger.LogrusObj.Errorf("acquire grpc server instance failed, %v", err)
		return
	}

	go discovery.DynamicServices(updateChan, configs.Conf.Services["ggcache"].Name)

	// Server implemented Pick interface, register a node selector for ggcache
	peers, err := discovery.ListServicePeers(configs.Conf.Services["ggcache"].Name)
	if err != nil {
		peers = []string{"serviceAddr"}
	}

	svr.SetPeers(peers)

	gm["scores"].RegisterServer(svr)

	// Start the server
	if err := svr.Start(); err != nil {
		logger.LogrusObj.Fatalf("Failed to start server: %v", err)
	}
}
