package app

import (
	"fmt"

	"github.com/1055373165/ggcache/configs"
	"github.com/1055373165/ggcache/internal/bussiness/student/dao"
	"github.com/1055373165/ggcache/internal/cache/core"
	cache "github.com/1055373165/ggcache/internal/cache/transport/grpc"
	"github.com/1055373165/ggcache/internal/observability"
	"github.com/1055373165/ggcache/pkg/common/logger"
	"github.com/1055373165/ggcache/pkg/etcd/discovery"
)

// Run starts the ggcache application.
// It encapsulates all bootstrapping logic so that cmd/ggcache/main.go remains minimal.
func Run(port, metricsPort, pprofPort int) error {
	// if flags not set, fallback to config values
	if port == 0 {
		port = configs.Conf.Server.Port
	}
	if metricsPort == 0 {
		metricsPort = configs.Conf.Server.MetricsPort
	}
	if pprofPort == 0 {
		pprofPort = configs.Conf.Server.PprofPort
	}

	if err := dao.InitDB(); err != nil {
		return fmt.Errorf("failed to init db: %w", err)
	}

	// observability (pprof + metrics + logging setup)
	observability.Init(pprofPort, metricsPort)

	serviceAddr := fmt.Sprintf("localhost:%d", port)
	gm := core.NewGroupManager([]string{"scores", "website"}, serviceAddr)

	updateChan := make(chan struct{})
	svr, err := cache.NewServer(updateChan, serviceAddr)
	if err != nil {
		return fmt.Errorf("create grpc server: %w", err)
	}

	// watch etcd for dynamic peers
	go discovery.DynamicServices(updateChan, configs.Conf.Services["groupcache"].Name)

	peers, err := discovery.ListServicePeers(configs.Conf.Services["groupcache"].Name)
	if err != nil {
		return fmt.Errorf("discover peers: %w", err)
	}
	svr.SetPeers(peers)

	gm["scores"].RegisterServer(svr)

	logger.LogrusObj.Infof("Starting ggcache node on %s", serviceAddr)
	if err := svr.Start(); err != nil {
		return fmt.Errorf("grpc server start: %w", err)
	}
	return nil
}
