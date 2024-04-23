package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
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

/*
1. put real nodes into the hash ring in the form of virtual nodes
2. sort the new hash ring
*/
func (ch *ConsistentHash) AddTruthNode(nodes []string) {
	for _, node := range nodes {
		for i := 0; i < ch.replicas; i++ {
			hash := int(ch.hash([]byte(node + strconv.Itoa(i))))
			ch.virtualNodes = append(ch.virtualNodes, hash)
			ch.hashMap[hash] = node
		}
	}
	sort.Ints(ch.virtualNodes)
}

// key hit the specified node
func (ch *ConsistentHash) GetTruthNode(key string) string {
	if len(ch.virtualNodes) == 0 {
		return ""
	}

	hash := int(ch.hash([]byte(key)))
	idx := sort.Search(len(ch.virtualNodes), func(i int) bool {
		return ch.virtualNodes[i] >= hash
	})

	return ch.hashMap[ch.virtualNodes[idx%len(ch.virtualNodes)]]
}

// remove the malfunction node from the hash ring, re-construct hash ring
func (ch *ConsistentHash) RemoveNode(node string) {
	virtualHash := []int{}
	for key, v := range ch.hashMap {
		if v == node {
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
	// logger.Logger.Warn("故障节点移除成功，相同请求应该打到其他节点上")
}
