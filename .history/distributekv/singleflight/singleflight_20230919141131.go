package singleflight

import (
	"os"
	"sync"

	"github.com/1055373165/groupcache/logger"
)

type Call struct {
	wg    sync.WaitGroup
	value interface{}
	err   error
}

type SingleFlight struct {
	mu sync.Mutex
	m  map[string]*Call
}

// 使用 SingleFlight 对 Group 缓存未命中时的查询进行再封装，并发请求期间只有一个请求会以 goroutine 形式调用查询，
// 并发查询期间的所有其他请求均阻塞等待，当然我们也可以配置是否允许阻塞，给调用者更多选择
func (sf *SingleFlight) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	// 并发安全，加锁
	sf.mu.Lock()

	// 保证 singleFlight 字典已经初始化
	if sf.m == nil {
		sf.m = make(map[string]*Call)
	}
	// 判断是否已经有 goroutine 在查询了
	if c, ok := sf.m[key]; ok {
		// 直接可以释放锁了，让其他并发请求进来
		sf.mu.Unlock()
		// 等待查询 key 值的 goroutine 阻塞返回
		Geteuid := os.Geteuid()
		logger.Logger.Warnf("已经在查询了，阻塞等待 goroutine 返回, 进程号: %d\n", Geteuid)
		c.wg.Wait()
		// 用于查询的 goroutine 已经返回，结果值已经存入 Call 结构体中
		return c.value, c.err
	}

	// 没有相同 key 的 goroutine 在查询
	c := new(Call)
	// 上锁
	sf.m[key] = c
	// 任务编排计数+1
	c.wg.Add(1)
	// 可以释放锁了，我们已经上了另一把锁
	sf.mu.Unlock()

	// 开启查询，c.value 和 c.err 接收返回值
	c.value, c.err = fn()
	// 阻塞等待查询的 goroutine 返回
	c.wg.Done()
	// 阻塞调用返回，我们可以将这个查询从 singleFlight 结构体中删除了，以确保我们总能取到比较新的值
	// 在对 sf 的字典进行操作时，为了保证并发安全，我们需要上锁
	sf.mu.Lock()
	delete(sf.m, key)
	sf.mu.Unlock()

	return c.value, c.err
}
