package consistenthash

import (
	"testing"

	"github.com/1055373165/groupcache/conf"
	"github.com/1055373165/groupcache/db"
	"github.com/1055373165/groupcache/logger"
)

func init() {
	conf.Init()
	logger.Init()
	db.Init()
}

func TestConsistentHash(t *testing.T) {
	// 使用 crc32.ChecksumIEEE hash 算法
	ch := NewConsistentHash(2, nil)
	logger.Logger.Info("NewConsistentHash Success...")
	// 先计算 key = 1 key = 2 key = 3 时的 hash 值

	// 6 16 26
	// 4 14 24
	// 2 12 22
	ch.AddTruthNode("2", "4")
	for _, virtualhash := range ch.virtualNodes {
		logger.Logger.Infof("虚拟节点 hash 值：%d, 对应的真实节点为：%s", virtualhash, ch.hashMap[virtualhash])
	}
	// node2 值：2322626082 值：4252452532
	// node4 值：2871910706 值：3693793700
	// 随机计算两个键
	key1, key2 := "key1", "key2"
	logger.Logger.Info(ch.hash([]byte(key1))) // 744252496
	logger.Logger.Info(ch.hash([]byte(key2))) // 3042260458
	// 因此 key1 应该打到节点 2 上，key2 应该打到节点 4 上
	// 2322626082 2871910706 3693793700 4252452532
	expectK1, expectK2 := "2", "4"
	if ch.GetTruthNode(key1) != expectK1 || ch.GetTruthNode(key2) != expectK2 {
		t.Fatal("GetTruthNode 错误")
	}
}
