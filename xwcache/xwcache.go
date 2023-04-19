package xwcache

import (
	"fmt"
	"log"
	"sync"
	"xwace/xwcache/xwcache/lru"
	"xwace/xwcache/xwcache/singleflight"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type Group struct {
	name        string
	getter      Getter
	mainCache   cache
	peersPicker PeerPicker
	sg          *singleflight.Group
}

var (
	mu     sync.RWMutex // 这里是为了给groups全局变量加锁的
	groups = make(map[string]*Group)
)

// Getter是从db溯源的实现
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
		sg: &singleflight.Group{},
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
		log.Println("[Cache] hit")
		return value, nil
	}

	v, err := g.sg.Do(key, func() (interface{}, error) {
		return g.load(key)
	})
	return v.(*ImmutableByte), err
}

func (g *Group) load(key string) (i *ImmutableByte, err error) {
	if g.peersPicker != nil { //则先尝试从其他节点获取
		if peer, ok := g.peersPicker.PickPeer(key); ok {
			if bytes, err := peer.Get(g.name, key); err == nil {
				return &ImmutableByte{bytes}, nil
			}
			log.Println("[Cache] Failed to get from peer", err)
		}
	}
	//若还是没有，则直接调用回调函数，查db
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
