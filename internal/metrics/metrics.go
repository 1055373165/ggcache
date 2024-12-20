// Package metrics provides cache metrics collection and reporting functionality.
package metrics

import (
	"fmt"
	"net/http"

	"github.com/1055373165/ggcache/pkg/common/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// 缓存命中相关指标
	cacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ggcache_hits_total",
		Help: "The total number of cache hits",
	})

	cacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ggcache_misses_total",
		Help: "The total number of cache misses",
	})

	cacheEvictions = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ggcache_evictions_total",
		Help: "The total number of cache evictions",
	})

	// 总请求数指标
	requestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ggcache_requests_total",
		Help: "The total number of requests received",
	})

	// 缓存大小相关指标
	cacheSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ggcache_size_bytes",
		Help: "The current size of the cache in bytes",
	})

	cacheItemCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ggcache_items_total",
		Help: "The total number of items in the cache",
	})

	// ARC 缓存特定指标
	arcT1Size = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ggcache_arc_t1_size",
		Help: "Number of items in ARC T1 list (recently used once)",
	})

	arcT2Size = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ggcache_arc_t2_size",
		Help: "Number of items in ARC T2 list (frequently used)",
	})

	arcB1Size = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ggcache_arc_b1_size",
		Help: "Number of items in ARC B1 list (ghost entries for T1)",
	})

	arcB2Size = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ggcache_arc_b2_size",
		Help: "Number of items in ARC B2 list (ghost entries for T2)",
	})

	arcTargetSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ggcache_arc_target_size",
		Help: "Target size for T1 (p) in ARC algorithm",
	})

	// 请求延迟指标
	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ggcache_request_duration_seconds",
			Help:    "Time spent processing cache requests",
			Buckets: prometheus.ExponentialBuckets(0.00001, 2, 20), // from 10µs to ~5s
		},
		[]string{"operation"},
	)
)

// StartMetricsServer 启动指标收集服务器
func StartMetricsServer(port int) {
	mux := http.NewServeMux()

	// 注册 /metrics 端点
	mux.Handle("/metrics", promhttp.Handler())

	// 添加从根路径到 /metrics 的重定向
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/metrics", http.StatusFound)
	})

	// 异步启动服务器
	go func() {
		addr := fmt.Sprintf(":%d", port)
		logger.LogrusObj.Infof("Starting metrics server on %s", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			logger.LogrusObj.Fatalf("Failed to start metrics server: %v", err)
		}
	}()
}

// RecordCacheHit 记录缓存命中
func RecordCacheHit() {
	cacheHits.Inc()
}

// RecordCacheMiss 记录缓存未命中
func RecordCacheMiss() {
	cacheMisses.Inc()
}

// RecordEviction 记录缓存驱逐
func RecordEviction() {
	cacheEvictions.Inc()
}

// UpdateCacheSize 更新缓存大小（字节）
func UpdateCacheSize(size int64) {
	cacheSize.Set(float64(size))
}

// UpdateCacheItemCount 更新缓存项数量
func UpdateCacheItemCount(count int64) {
	cacheItemCount.Set(float64(count))
}

// UpdateARCMetrics updates all ARC-specific metrics
func UpdateARCMetrics(t1Size, t2Size, b1Size, b2Size, targetSize int) {
	arcT1Size.Set(float64(t1Size))
	arcT2Size.Set(float64(t2Size))
	arcB1Size.Set(float64(b1Size))
	arcB2Size.Set(float64(b2Size))
	arcTargetSize.Set(float64(targetSize))
}

// ObserveRequestDuration records the duration of a cache operation
func ObserveRequestDuration(operation string, duration float64) {
	requestDuration.WithLabelValues(operation).Observe(duration)
}

// RecordRequest increments the total request counter
func RecordRequest() {
	requestsTotal.Inc()
}
