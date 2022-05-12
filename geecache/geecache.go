/**
  @author: 黄睿楠
  @since: 2022/5/6
  @desc: 负责与外部交互，控制缓存存储和获取的主流程
**/

package geecache

import (
	pb "Gee-Cache/geecache/geecachepb"
	"Gee-Cache/geecache/singleflight"
	"fmt"
	"log"
	"sync"
)

type Getter interface {
	// Get 回调函数，在缓存不存在时，调用这个函数，得到源数据
	Get(key string) ([]byte,error)
}

// GetterFunc 定义函数类型 GetterFunc
type GetterFunc func(key string) ([]byte,error)

// Get 并实现 Getter 接口的 Get 方法
// 函数类型实现某一个接口，称之为接口型函数，方便使用者在调用时既能够传入函数作为参数，也能够传入实现了该接口的结构体作为参数
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name string
	// 缓存未命中时获取源数据的回调(callback)
	getter Getter
	mainCache cache
	peers     PeerPicker
	loader *singleflight.Group
}

var(
	mu sync.RWMutex
	groups=make(map[string]*Group)
)

func NewGroup(name string,cacheBytes int64,getter Getter) *Group {
	if getter==nil {
		panic("nil Getter")
	}

	mu.Lock()
	defer mu.Unlock()
	g:=&Group{
		name: name,
		getter: getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader: &singleflight.Group{},
	}
	groups[name]=g
	return g
}

func GetGroup(name string) *Group {
	// 只读锁
	mu.RLock()
	g:=groups[name]
	mu.RUnlock()
	return g
}

// RegisterPeers 将 实现了 PeerPicker 接口的Peer注入到 Group 中
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) Get(key string) (ByteView,error) {
	if key=="" {
		return ByteView{},fmt.Errorf("key is required")
	}
	if v,ok:=g.mainCache.get(key);ok {
		log.Println("[GeeCache] hit")
		return v,nil
	}
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	// 将 load 的逻辑，使用 g.loader.Do 包裹起来，确保并发场景下针对相同的 key，load 过程只会调用一次
	view, err := g.loader.Do(key, func() (interface{}, error) {
		// 如果该节点注册过
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
		return view.(ByteView), nil
	}
	return
}

// 使用实现了 PeerGetter 接口的 httpGetter 访问远程节点，获取缓存值
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}

func (g *Group) getLocally(key string) (ByteView, error) {
	// 调用用户回调函数 g.getter.Get() 获取源数据
	bytes,err:=g.getter.Get(key)
	if err!=nil {
		return ByteView{}, err
	}
	value:=ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value,nil
}

// 将源数据添加到缓存 mainCache 中
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key,value)
}
