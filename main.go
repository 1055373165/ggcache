package main

import (
	"flag"
	"fmt"

	"github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/internal/bussiness/student/dao"
	"github.com/1055373165/ggcache/internal/cache"
	"github.com/1055373165/ggcache/internal/metrics"
	"github.com/1055373165/ggcache/pkg/common/logger"
	"github.com/1055373165/ggcache/pkg/etcd/discovery"
)

var (
	port        = flag.Int("port", 8888, "service node port")
	metricsPort = flag.Int("metricsPort", 2222, "metrics port")
)

func main() {
	config.InitConfig()
	dao.InitDB()
	flag.Parse()

	// 启动指标收集服务
	metrics.StartMetricsServer(*metricsPort)
	logger.LogrusObj.Infof("Metrics server started on port %d", *metricsPort)

	serviceAddr := fmt.Sprintf("localhost:%d", *port)
	gm := cache.NewGroupManager([]string{"scores", "website"}, serviceAddr)

	updateChan := make(chan struct{})
	svr, err := cache.NewServer(updateChan, serviceAddr)
	if err != nil {
		logger.LogrusObj.Errorf("acquire grpc server instance failed, %v", err)
		return
	}

	go discovery.DynamicServices(updateChan, config.Conf.Services["groupcache"].Name)

	peers, err := discovery.ListServicePeers(config.Conf.Services["groupcache"].Name)
	if err != nil {
		logger.LogrusObj.Fatalf("failed to discover peers: %v", err)
		return
	}

	svr.SetPeers(peers)

	gm["scores"].RegisterServer(svr)

	if err := svr.Start(); err != nil {
		logger.LogrusObj.Fatalf("server failed to start: %v", err)
	}
}
