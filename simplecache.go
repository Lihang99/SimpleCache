package SimpleCache

import (
	pb "SimpleCache/proto"
	"SimpleCache/singleflight"
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
	peers     PeerPicker
	loader    *singleflight.Group
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
		loader:    &singleflight.Group{},
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

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value CacheValue, err error) {
	cache, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return cache.(CacheValue), nil
	}
	return
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (CacheValue, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return CacheValue{}, nil
	}
	return CacheValue{b: res.Value}, nil
}
