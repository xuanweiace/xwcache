package singleflight

import (
	"sync"
)

type call struct {
	val interface{}
	err error
	wg  sync.WaitGroup
}

type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// 同步方法Do
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil { //懒加载
		g.m = make(map[string]*call)
	}
	//重复请求
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	//首次请求
	c := call{
		val: nil,
		err: nil,
		wg:  sync.WaitGroup{},
	}
	g.m[key] = &c
	c.wg.Add(1) // 必须要在这后面Unlock
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
