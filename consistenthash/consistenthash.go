package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash struct {
	replicas int
	mp       map[int]string
	hashFunc func(data []byte) uint32
	keys     []int
}

func NewConsistentHash(replicas int, hashFunc func(data []byte) uint32) *Hash {
	h := &Hash{
		replicas: replicas,
		mp:       make(map[int]string),
		hashFunc: hashFunc,
	}
	if hashFunc == nil {
		h.hashFunc = crc32.ChecksumIEEE
	}
	return h
}
func (h *Hash) AddNode(keys ...string) {
	for _, key := range keys {
		for i := 0; i < h.replicas; i++ {
			newKey := strconv.Itoa(i) + key
			hashcode := int(h.hashFunc([]byte(newKey))) // 转回int是为了成环search
			h.keys = append(h.keys, hashcode)
			h.mp[hashcode] = key
		}
	}
	//因为AddNode操作很少，所以可以在这里维护排序，Get的时候直接取就可以了
	sort.Ints(h.keys)
}

// 获得：要存储key，需要访问的机器号key
func (h *Hash) Get(key string) string {
	hashCode := int(h.hashFunc([]byte(key)))
	//找到第一个满足该函数的值
	idx := sort.Search(len(h.keys), func(i int) bool {
		return h.keys[i] >= hashCode
	})
	return h.mp[h.keys[idx%(len(h.keys))]]
}
