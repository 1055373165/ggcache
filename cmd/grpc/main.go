package main

import (
	"flag"
	"fmt"

	"github.com/1055373165/ggcache/config"
	conf "github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/internal/middleware/etcd/discovery/discovery3"
	"github.com/1055373165/ggcache/internal/pkg/student/dao"
	"github.com/1055373165/ggcache/internal/service"
	"github.com/1055373165/ggcache/utils/logger"
)

var (
	port = flag.Int("port", 9999, "service node port")
)

func main() {
	conf.InitConfig()
	dao.InitDB()
	flag.Parse()

	// grpc node local service address
	serviceAddr := fmt.Sprintf("localhost:%d", *port)
	gm := service.NewGroupManager([]string{"scores", "website"}, serviceAddr)

	// get a grpc service instance（通过通信来共享内存而不是通过共享内存来通信）
	updateChan := make(chan bool)
	svr, err := service.NewServer(updateChan, serviceAddr)
	if err != nil {
		logger.LogrusObj.Errorf("acquire grpc server instance failed, %v", err)
		return
	}

	go discovery3.DynamicServices(updateChan, config.Conf.Services["groupcache"].Name)

	// Server implemented Pick interface, register a node selector for ggcache
	peers, err := discovery3.ListServicePeers(config.Conf.Services["groupcache"].Name)
	if err != nil {
		peers = []string{"serviceAddr"}
	}

	svr.SetPeers(peers)

	gm["scores"].RegisterServer(svr)

	// start grpc service
	svr.Start()
}
