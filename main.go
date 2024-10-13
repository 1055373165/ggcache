package main

import (
	"flag"
	"fmt"

	"github.com/1055373165/ggcache/config"
	discovery "github.com/1055373165/ggcache/discovery"
	grpcservice "github.com/1055373165/ggcache/internal"
	"github.com/1055373165/ggcache/test/servicefortest/student/dao"
	"github.com/1055373165/ggcache/utils/logger"
)

var (
	port = flag.Int("port", 9999, "service node port")
)

func main() {
	config.InitConfig()
	dao.InitDB()
	flag.Parse()

	// grpc node local service address
	serviceAddr := fmt.Sprintf("localhost:%d", *port)
	gm := grpcservice.NewGroupManager([]string{"scores", "website"}, serviceAddr)

	// get a grpc service instance
	updateChan := make(chan struct{})
	svr, err := grpcservice.NewServer(updateChan, serviceAddr)
	if err != nil {
		logger.LogrusObj.Errorf("acquire grpc server instance failed, %v", err)
		return
	}

	go discovery.DynamicServices(updateChan, config.Conf.Services["ggcache"].Name)

	// Server implemented Pick interface, register a node selector for ggcache
	peers, err := discovery.ListServicePeers(config.Conf.Services["ggcache"].Name)
	if err != nil {
		peers = []string{"serviceAddr"}
	}

	svr.SetPeers(peers)

	gm["scores"].RegisterServer(svr)

	// start grpc service
	svr.Start()
}
