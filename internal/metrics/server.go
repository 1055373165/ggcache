package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// StartMetricsServer 启动 Prometheus metrics 服务器
func StartMetricsServer(port int) error {
	// 注册 Prometheus metrics handler
	http.Handle("/metrics", promhttp.Handler())

	// 启动 HTTP 服务器
	addr := fmt.Sprintf(":%d", port)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			panic(fmt.Sprintf("Failed to start metrics server: %v", err))
		}
	}()

	return nil
}
