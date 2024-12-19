package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	registryOnce sync.Once
	registry     *prometheus.Registry

	// 当前goroutine数量
	currentGoroutines = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "netdev_check",
		Name:      "current_goroutines",
		Help:      "Current number of goroutines",
	})

	// CPU使用率
	cpuUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "netdev_check",
		Name:      "cpu_usage_percent",
		Help:      "CPU usage percentage",
	})

	// 内存使用率
	memoryUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "netdev_check",
		Name:      "memory_usage_percent",
		Help:      "Memory usage percentage",
	})

	// 任务完成总数
	tasksCompleted = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "netdev_check",
		Name:      "tasks_completed_total",
		Help:      "Total number of completed tasks",
	})

	// 所有指标的列表，用于批量注册
	metrics = []prometheus.Collector{
		currentGoroutines,
		cpuUsage,
		memoryUsage,
		tasksCompleted,
	}
)

// InitializePrometheusRegistry 初始化并注册所有指标
func InitializePrometheusRegistry() *prometheus.Registry {
	registryOnce.Do(func() {
		registry = prometheus.NewRegistry()

		// 注册所有指标
		for _, metric := range metrics {
			if err := registry.Register(metric); err != nil {
				// 如果指标已经注册，尝试取消注册后重新注册
				registry.Unregister(metric)
				if err := registry.Register(metric); err != nil {
					panic(err)
				}
			}
		}
	})
	return registry
}

// PrometheusMetricsCollector 实现了 prometheus.Collector 接口
type PrometheusMetricsCollector struct {
	metrics *TaskMetrics
}

// NewPrometheusMetricsCollector 创建一个新的 Prometheus 指标收集器
func NewPrometheusMetricsCollector(m *TaskMetrics) *PrometheusMetricsCollector {
	// 确保指标已经注册
	InitializePrometheusRegistry()
	return &PrometheusMetricsCollector{metrics: m}
}

// UpdateMetrics 更新所有 Prometheus 指标
func (c *PrometheusMetricsCollector) UpdateMetrics() {
	m := c.metrics
	snapshot := m.GetMetricsSnapshot()

	currentGoroutines.Set(float64(snapshot["current_goroutines"].(int64)))
	cpuUsage.Set(snapshot["cpu_usage_percent"].(float64))
	memoryUsage.Set(snapshot["memory_usage_percent"].(float64))
}

// StartPrometheusMetricsUpdater 启动一个后台 goroutine 定期更新 Prometheus 指标
func StartPrometheusMetricsUpdater(collector *PrometheusMetricsCollector, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			collector.UpdateMetrics()
		}
	}()
}
