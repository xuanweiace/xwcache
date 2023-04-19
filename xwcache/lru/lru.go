package lru

import "container/list"

type Value interface {
	Len() int64 //返回字节数
}
type entry struct {
	key   string
	value Value
}

type ICache interface {
	Get(key string) (value Value, ok bool)
	Add(key string, value Value)
	RemoveOldest()
}
type Cache struct {
	capBytes  int64
	lenBytes  int64
	li        *list.List
	cache     map[string]*list.Element
	onEvicted func(key string, value Value)
}

func NewCache(capacity int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		capBytes:  capacity,
		lenBytes:  0,
		li:        list.New(),
		cache:     make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.li.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return nil, false
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.li.MoveToFront(ele)
		c.lenBytes -= ele.Value.(*entry).value.Len()
		ele.Value.(*entry).value = value
		c.lenBytes += value.Len()
	} else {
		e := &entry{
			key:   key,
			value: value,
		}
		ele := c.li.PushFront(e)
		c.cache[key] = ele
		c.lenBytes += int64(len(key)) + value.Len()
	}
	for c.capBytes > 0 && c.lenBytes > c.capBytes {
		c.RemoveOldest()
	}

}

func (c *Cache) RemoveOldest() {
	back := c.li.Back()
	//if back != nil {}
	c.li.Remove(back)
	key := back.Value.(*entry).key
	value := back.Value.(*entry).value
	delete(c.cache, key)
	c.lenBytes -= int64(len(key)) + value.Len()
	if c.onEvicted != nil {
		c.onEvicted(key, value)
	}
}

func (c *Cache) Len() int {
	return len(c.cache)
}
