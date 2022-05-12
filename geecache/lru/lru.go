/**
  @author: 黄睿楠
  @since: 2022/5/5
  @desc: lru 缓存淘汰策略
**/

package lru

import "container/list"

type Cache struct {
	maxBytes int64   // 允许使用的最大内存
	nbytes int64   // 当前已使用的内存，len(key)+len(value)
	// ll和cache共同构成了LinkedHashMap
	ll *list.List                          // 双向链表实现的队列
	cache map[string]*list.Element         // 字典，map的value为链表的节点的指针
	OnEvicted func(key string,value Value) // 某条记录被移除时的回调函数，可以为 nil
}

// list.Element包含4个字段：prev,next,list,Value interface{}，通过ele.Value.(*entry)可以得到节点的key和value值
type entry struct {
	key   string
	value Value
}

// Value 缓存中存的值可以是实现了 Value 接口的任意类型，该接口只包含了一个方法 Len() int，用于返回值所占用的内存大小。
type Value interface {
	Len() int
}

func New(maxBytes int64,onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes: maxBytes,
		ll: list.New(),
		cache: make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}



func (c *Cache) Add(key string,value Value) {
	// 修改，类似于删除操作
	if ele,ok:=c.cache[key];ok {
		c.ll.MoveToFront(ele)
		kv:=ele.Value.(*entry)
		c.nbytes=c.nbytes-int64(kv.value.Len())+int64(value.Len())
		kv.value=value
	} else {
		ele:=c.ll.PushFront(&entry{key,value})
		c.cache[key]=ele
		c.nbytes=c.nbytes+int64(len(key))+int64(value.Len())
	}
	for c.maxBytes!=0 && c.maxBytes<c.nbytes {
		c.removeOldest()
	}
}

func (c *Cache) removeOldest() {
	ele:=c.ll.Back()
	if ele!=nil {
		c.ll.Remove(ele)
		kv:=ele.Value.(*entry)
		delete(c.cache,kv.key)
		c.nbytes=c.nbytes- int64((len(kv.key))+kv.value.Len())
		if c.OnEvicted!=nil {
			c.OnEvicted(kv.key,kv.value)
		}
	}
}

func (c *Cache) Get(key string) (value Value,ok bool) {
	if ele,ok:=c.cache[key];ok{
		c.ll.MoveToFront(ele)
		// 由list.Element源码可知，这一步将元素ele的Value设置为entry类型，并将这个Value赋值给kv
		kv:=ele.Value.(*entry)
		return kv.value,true
	}
	return
}

// Len 获取cache中的元素个数,为了方便测试
func (c *Cache) Len() int {
	return c.ll.Len()
}
