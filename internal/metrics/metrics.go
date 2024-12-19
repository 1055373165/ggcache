package metrics

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// TaskMetrics 存储任务执行的性能指标
type TaskMetrics struct {
	// 系统资源指标
	CurrentGoroutines int64
	SystemMemoryUsed  int64
	SystemCPUUsage    float64

	// 缓存性能指标
	CacheSize      int64 // 当前缓存大小（字节）
	CacheItemCount int64 // 缓存项数量
	CacheHits      int64 // 缓存命中次数
	CacheMisses    int64 // 缓存未命中次数
	EvictionCount  int64 // 缓存驱逐次数

	// 请求处理指标
	RequestsTotal    int64   // 总请求数
	RequestErrors    int64   // 错误请求数
	RequestLatencies []int64 // 请求延迟记录（纳秒）

	// 内部状态
	startTime time.Time
	mutex     sync.RWMutex
}

var (
	GlobalMetrics = &TaskMetrics{
		RequestLatencies: make([]int64, 0, 1000), // 保留最近1000次请求的延迟
		startTime:        time.Now(),
	}
)

func (m *TaskMetrics) IncrementGoroutines() {
	atomic.AddInt64(&m.CurrentGoroutines, 1)
}

func (m *TaskMetrics) DecrementGoroutines() {
	atomic.AddInt64(&m.CurrentGoroutines, -1)
}

func (m *TaskMetrics) UpdateCacheSize(size int64) {
	atomic.StoreInt64(&m.CacheSize, size)
}

func (m *TaskMetrics) UpdateCacheItemCount(count int64) {
	atomic.StoreInt64(&m.CacheItemCount, count)
}

func (m *TaskMetrics) IncrementCacheHits() {
	atomic.AddInt64(&m.CacheHits, 1)
}

func (m *TaskMetrics) IncrementCacheMisses() {
	atomic.AddInt64(&m.CacheMisses, 1)
}

func (m *TaskMetrics) IncrementEvictions() {
	atomic.AddInt64(&m.EvictionCount, 1)
}

func (m *TaskMetrics) IncrementRequests() {
	atomic.AddInt64(&m.RequestsTotal, 1)
}

func (m *TaskMetrics) IncrementErrors() {
	atomic.AddInt64(&m.RequestErrors, 1)
}

func (m *TaskMetrics) RecordRequestLatency(duration time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 保持固定大小的延迟记录，超出容量时移除最旧的记录
	if len(m.RequestLatencies) >= 1000 {
		m.RequestLatencies = m.RequestLatencies[1:]
	}
	m.RequestLatencies = append(m.RequestLatencies, int64(duration))
}

func (m *TaskMetrics) GetLatencyStats() (min, max, avg int64, count int) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.RequestLatencies) == 0 {
		return 0, 0, 0, 0
	}

	var total int64
	min = m.RequestLatencies[0]
	max = m.RequestLatencies[0]

	for _, t := range m.RequestLatencies {
		if t < min {
			min = t
		}
		if t > max {
			max = t
		}
		total += t
	}

	return min, max, total / int64(len(m.RequestLatencies)), len(m.RequestLatencies)
}

func (m *TaskMetrics) GetSystemMetrics() map[string]float64 {
	metrics := make(map[string]float64)

	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		metrics["cpu_usage"] = cpuPercent[0]
		m.SystemCPUUsage = cpuPercent[0] // 直接赋值，因为这个值只在这里更新
	}

	memInfo, err := mem.VirtualMemory()
	if err == nil {
		metrics["memory_usage"] = memInfo.UsedPercent
		atomic.StoreInt64(&m.SystemMemoryUsed, int64(memInfo.Used))
	}

	metrics["goroutine_count"] = float64(runtime.NumGoroutine())

	return metrics
}

// GetMetricsSnapshot 获取当前指标快照
func (m *TaskMetrics) GetMetricsSnapshot() map[string]interface{} {
	min, max, avg, count := m.GetLatencyStats()
	systemMetrics := m.GetSystemMetrics()

	return map[string]interface{}{
		// 系统资源
		"current_goroutines": atomic.LoadInt64(&m.CurrentGoroutines),
		"system_memory_used": atomic.LoadInt64(&m.SystemMemoryUsed),
		"system_cpu_usage":   m.SystemCPUUsage, // 直接读取，因为只在 GetSystemMetrics 中更新

		// 缓存性能
		"cache_size_bytes": atomic.LoadInt64(&m.CacheSize),
		"cache_items":      atomic.LoadInt64(&m.CacheItemCount),
		"cache_hits":       atomic.LoadInt64(&m.CacheHits),
		"cache_misses":     atomic.LoadInt64(&m.CacheMisses),
		"cache_evictions":  atomic.LoadInt64(&m.EvictionCount),
		"cache_hit_ratio":  float64(atomic.LoadInt64(&m.CacheHits)) / float64(atomic.LoadInt64(&m.CacheHits)+atomic.LoadInt64(&m.CacheMisses)),

		// 请求统计
		"requests_total":  atomic.LoadInt64(&m.RequestsTotal),
		"request_errors":  atomic.LoadInt64(&m.RequestErrors),
		"error_rate":      float64(atomic.LoadInt64(&m.RequestErrors)) / float64(atomic.LoadInt64(&m.RequestsTotal)),
		"min_latency_ms":  float64(min) / float64(time.Millisecond),
		"max_latency_ms":  float64(max) / float64(time.Millisecond),
		"avg_latency_ms":  float64(avg) / float64(time.Millisecond),
		"latency_samples": count,

		// 运行时信息
		"uptime_seconds":       time.Since(m.startTime).Seconds(),
		"cpu_usage_percent":    systemMetrics["cpu_usage"],
		"memory_usage_percent": systemMetrics["memory_usage"],
		"total_goroutines":     systemMetrics["goroutine_count"],
	}
}
