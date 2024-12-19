package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/internal/bussiness/student/dao"
	"github.com/1055373165/ggcache/internal/cache"
	"github.com/1055373165/ggcache/pkg/common/logger"
	"github.com/1055373165/ggcache/pkg/etcd/discovery"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	port        = flag.Int("port", 8888, "service node port")
	metricsPort = flag.Int("metricsPort", 2222, "metrics port")
)

func main() {
	config.InitConfig()
	dao.InitDB()
	flag.Parse()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	// 添加一个简单的重定向，从根路径跳转到 /metrics
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/metrics", http.StatusFound)
	})
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", *metricsPort), mux); err != nil {
			logger.LogrusObj.Fatalf("Failed to start metrics server: %v", err)
		}
	}()
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
