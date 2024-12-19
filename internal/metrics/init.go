package metrics

import "time"

// Initialize 初始化所有指标收集相关的组件
func Initialize() {
	// 启动 Prometheus 指标服务器
	StartMetricsServer(9091)

	// 创建并启动 Prometheus 指标收集器
	collector := NewPrometheusMetricsCollector(GlobalMetrics)
	StartPrometheusMetricsUpdater(collector, 5*time.Second)
}
