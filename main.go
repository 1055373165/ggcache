package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/internal/bussiness/student/dao"
	"github.com/1055373165/ggcache/internal/cache"
	"github.com/1055373165/ggcache/internal/metrics"
	"github.com/1055373165/ggcache/pkg/common/logger"
	"github.com/1055373165/ggcache/pkg/etcd/discovery"
)

var (
	port        = flag.Int("port", 9999, "service node port")
	metricsPort = flag.Int("metricsPort", 2222, "metrics port")
	pprofPort   = flag.Int("pprofPort", 6060, "pprof port")
)

func main() {
	config.InitConfig()
	dao.InitDB()
	flag.Parse()

	// 启动pprof服务器
	go func() {
		pprofAddr := fmt.Sprintf("localhost:%d", *pprofPort)
		logger.LogrusObj.Infof("Starting pprof server on %s", pprofAddr)
		if err := http.ListenAndServe(pprofAddr, nil); err != nil {
			logger.LogrusObj.Errorf("Failed to start pprof server: %v", err)
		}
	}()

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
		logger.LogrusObj.Fatalf("failed to start server: %v", err)
		return
	}
}
