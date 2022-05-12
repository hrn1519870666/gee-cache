/**
  @author: 黄睿楠
  @since: 2022/5/6
  @desc: 并发控制
**/

package geecache

import (
	"Gee-Cache/geecache/lru"
	"sync"
)

type cache struct {
	mu sync.Mutex
	lru *lru.Cache
	cacheBytes int64
}

func (c *cache) add(key string,value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 延迟初始化
	if c.lru==nil {
		c.lru=lru.New(c.cacheBytes,nil)
	}
	c.lru.Add(key,value)
}

func (c *cache) get(key string) (value ByteView,ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru==nil {
		// 方法签名中定义了返回值名称之后，才能直接 return
		return
	}
	if v,ok:=c.lru.Get(key);ok {
		// v是Value接口类型（只有一个Len方法），类似于空接口，可以通过v.(ByteView)进行类型转换
		return v.(ByteView),ok
	}
	return
}