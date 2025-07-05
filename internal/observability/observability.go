package observability

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/1055373165/ggcache/internal/metrics"
	"github.com/1055373165/ggcache/pkg/common/logger"
	"github.com/sirupsen/logrus"
)

// Init starts pprof and Prometheus metrics servers based on provided ports.
// If a port is <=0, that specific component will not be started.
func Init(pprofPort, metricsPort int) {
	// ensure logger formatted as JSON & info level by default
	SetupLogging(logrus.InfoLevel, true)
	// pprof
	if pprofPort > 0 {
		go func() {
			addr := fmt.Sprintf("localhost:%d", pprofPort)
			logger.LogrusObj.Infof("Starting pprof server on %s", addr)
			if err := http.ListenAndServe(addr, nil); err != nil {
				logger.LogrusObj.Errorf("pprof server error: %v", err)
			}
		}()
	}

	// metrics
	if metricsPort > 0 {
		metrics.StartMetricsServer(metricsPort)
		logger.LogrusObj.Infof("Metrics server started on port %d", metricsPort)
	}
}
