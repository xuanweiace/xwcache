package xwcache

import (
	"fmt"
	"log"
	"testing"
)

func TestA(t *testing.T) {
	b := []byte{1, 2, 3}
	c := cloneByteSlice(b)
	var bb []byte
	fmt.Println(bb)
	fmt.Println(c)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	bytes, err := f(key)
	return bytes, err
}

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGet(t *testing.T) {
	cnt := make(map[string]int)
	group := NewGroup("test", 100, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				if _, ok := cnt[key]; !ok {
					cnt[key] = 0
				}
				cnt[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	// test
	for k, v := range db {
		if view, err := group.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		} // load from callback function
		if _, err := group.Get(k); err != nil || cnt[k] > 1 {
			t.Fatalf("cache %s miss", k)
		} // cache hit
	}

	if view, err := group.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}
