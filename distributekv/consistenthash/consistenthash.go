package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"

	"github.com/1055373165/groupcache/middleware/logger"
)

type Hash func(data []byte) uint32

type ConsistentHash struct {
	hash         Hash
	replicas     int
	virtualNodes []int
	hashMap      map[int]string
}

func NewConsistentHash(replicas int, hash Hash) *ConsistentHash {
	if hash == nil {
		hash = crc32.ChecksumIEEE
	}

	return &ConsistentHash{
		hash:     hash,
		replicas: replicas,
		hashMap:  map[int]string{},
	}
}

func (ch *ConsistentHash) AddTruthNode(nodes ...string) {
	for _, node := range nodes {
		for i := 0; i < ch.replicas; i++ {
			hash := int(ch.hash([]byte(node + strconv.Itoa(i))))
			ch.virtualNodes = append(ch.virtualNodes, hash)
			ch.hashMap[hash] = node
		}
	}
	sort.Ints(ch.virtualNodes)
}

// 选择真实节点
func (ch *ConsistentHash) GetTruthNode(key string) string {
	if len(ch.virtualNodes) == 0 {
		return ""
	}

	hash := int(ch.hash([]byte(key)))
	idx := sort.Search(len(ch.virtualNodes), func(i int) bool {
		return ch.virtualNodes[i] >= hash
	})
	logger.Logger.Infof("计算出 key 的 hash: %d, 顺时针选择的虚拟节点下标 idx: %d", hash, idx)
	logger.Logger.Infof("2322626082 2871910706 3693793700 4252452532")
	logger.Logger.Infof("选择的真实节点：%s", ch.hashMap[ch.virtualNodes[idx%len(ch.virtualNodes)]])
	return ch.hashMap[ch.virtualNodes[idx%len(ch.virtualNodes)]]
}

func (ch *ConsistentHash) RemovePeer(peer string) {
	// 将真实节点从 hash 环中删除
	logger.Logger.Warn("peer:", peer)
	virtualHash := []int{}
	for key, v := range ch.hashMap {
		logger.Logger.Warn("peers:", v)
		if v == peer {
			delete(ch.hashMap, key)
			virtualHash = append(virtualHash, key)
		}
	}

	for i := 0; i < len(virtualHash); i++ {
		for index, value := range ch.virtualNodes {
			if virtualHash[i] == value {
				ch.virtualNodes = append(ch.virtualNodes[:index], ch.virtualNodes[index+1:]...)
			}
		}
	}

	logger.Logger.Warn("故障节点移除成功，相同请求应该打到其他节点上")
	logger.Logger.Warn(len(ch.hashMap))
	logger.Logger.Warn(len(ch.virtualNodes))
}
