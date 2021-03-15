package lru

import (
	"container/list"
)

type Cache struct {
	//最大内存
	maxBytes int64
	//已使用内存
	usedBytes int64
	list      *list.List
	cache     map[string]*list.Element
	//缓存记录被淘汰的时候的回调函数
	OnEliminate func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

//Constructor of Cache
func New(maxBytes int64, onEliminate func(string, Value)) *Cache {
	return &Cache{
		maxBytes:    maxBytes,
		list:        list.New(),
		cache:       make(map[string]*list.Element),
		OnEliminate: onEliminate,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.list.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, ok
	}
	return
}

//on the basic of LRU to remove item
func (c *Cache) RemoveOldestItem() {
	ele := c.list.Back()
	if ele != nil {
		c.list.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.usedBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEliminate != nil {
			c.OnEliminate(kv.key, kv.value)
		}
	}
}

func (c *Cache) Put(key string, value Value) {
	//existed in cache
	if ele, ok := c.cache[key]; ok {
		kv := ele.Value.(*entry)
		kv.value = value
		c.list.MoveToFront(ele)
		c.usedBytes += int64(value.Len()) - int64(kv.value.Len())
	} else {
		//1.new element
		//2.update cache map
		//3.calculate used memory
		//4.if cap is full ,remove item(lru)
		ele := c.list.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.usedBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.usedBytes {
		c.RemoveOldestItem()
	}
}

// Len the number of cache entries
func (c *Cache) Len() int {
	return c.list.Len()
}
