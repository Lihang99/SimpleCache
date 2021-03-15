package SimpleCache

import (
	"fmt"
	"log"
	"sync"
)

//接口型函数
type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.Unlock()
	return g
}

//Use key to get value from cache
func (g *Group) Get(key string) (CacheValue, error) {
	if key == "" {
		return CacheValue{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok {
		log.Printf("[SimpleCache] Get Success,key = %s", key)
		return v, nil
	}
	return g.load(key)
}

func (g *Group) addCache(key string, value CacheValue) {
	g.mainCache.put(key, value)
}

func (g *Group) getLocally(key string) (CacheValue, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return CacheValue{}, err
	}
	value := CacheValue{b: cloneBytes(bytes)}
	g.addCache(key, value)
	return value, nil
}

func (g *Group) load(key string) (value CacheValue, err error) {
	return g.getLocally(key)
}
