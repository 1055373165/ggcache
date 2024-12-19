package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// 缓存相关指标
var (
	cacheHits = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ggcache",
		Name:      "cache_hits_total",
		Help:      "Total number of cache hits",
	})

	cacheMisses = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ggcache",
		Name:      "cache_misses_total",
		Help:      "Total number of cache misses",
	})

	cacheEvictions = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ggcache",
		Name:      "cache_evictions_total",
		Help:      "Total number of cache evictions",
	})

	cacheSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "ggcache",
		Name:      "cache_size_bytes",
		Help:      "Current size of cache in bytes",
	})

	cacheItemCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "ggcache",
		Name:      "cache_items_total",
		Help:      "Total number of items in cache",
	})

	requestDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "ggcache",
		Name:      "request_duration_seconds",
		Help:      "Time taken to process cache requests",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10), // from 1ms to ~1s
	})

	requestErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ggcache",
		Name:      "request_errors_total",
		Help:      "Total number of request errors",
	})
)

func init() {
	// 注册所有指标
	prometheus.MustRegister(
		cacheHits,
		cacheMisses,
		cacheEvictions,
		cacheSize,
		cacheItemCount,
		requestDuration,
		requestErrors,
	)
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

// UpdateCacheSize 更新缓存大小
func UpdateCacheSize(size int64) {
	cacheSize.Set(float64(size))
}

// UpdateCacheItemCount 更新缓存项数量
func UpdateCacheItemCount(count int64) {
	cacheItemCount.Set(float64(count))
}

// RecordRequestDuration 记录请求处理时间
func RecordRequestDuration(duration time.Duration) {
	requestDuration.Observe(duration.Seconds())
}

// RecordError 记录请求错误
func RecordError() {
	requestErrors.Inc()
}

// GetCacheHitRatio 获取缓存命中率
func GetCacheHitRatio() float64 {
	var mHits, mMisses dto.Metric
	_ = cacheHits.(prometheus.Metric).Write(&mHits)
	_ = cacheMisses.(prometheus.Metric).Write(&mMisses)

	hits := *mHits.Counter.Value
	misses := *mMisses.Counter.Value
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total)
}
