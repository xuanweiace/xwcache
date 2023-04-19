package xwcache

import (
	"fmt"
	"log"
	"sync"
	"xwace/xwcache/xwcache/lru"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type Group struct {
	name        string
	getter      Getter
	mainCache   cache
	peersPicker PeerPicker
}

var (
	mu     sync.RWMutex // 这里是为了给groups全局变量加锁的
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheByte int64, getter Getter) *Group {
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:   name,
		getter: getter,
		mainCache: cache{
			mu:         sync.Mutex{},
			lru:        lru.NewCache(cacheByte, nil),
			cacheBytes: cacheByte,
		},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (*ImmutableByte, error) {
	if key == "" {
		return nil, fmt.Errorf("不支持null key")
	}
	if value, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return value, nil
	}
	return g.load(key)
}

func (g *Group) load(key string) (*ImmutableByte, error) {
	//目前直接调用回调函数
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (*ImmutableByte, error) {
	bytes, err := g.getter.Get(key)
	value := ImmutableByte{b: bytes}
	// 加回缓存 注意不能写在load里
	g.mainCache.add(key, &value)

	// todo 为什么要克隆一下？
	//return &ImmutableByte{b: cloneByteSlice(b)}, err
	return &value, err
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peersPicker != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peersPicker = peers
}
