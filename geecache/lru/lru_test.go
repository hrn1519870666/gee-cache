/**
  @author: 黄睿楠
  @since: 2022/5/5
  @desc:
**/

package lru

import (
	"reflect"
	"testing"
)

// 需要让string类型实现Len方法
type String string

func (s String) Len() int{
	return len(s)
}

func TestAdd(t *testing.T) {
	// maxBytes=0时，不会执行removeOldest操作
	lru:= New(int64(0),nil)
	lru.Add("key", String("v1"))
	lru.Add("key", String("v2"))

	if lru.nbytes !=int64(len("key")+len("v2")) {
		t.Fatal("expected 5 but got",lru.nbytes)
	}
}

func TestRemoveOldest(t *testing.T) {
	lru:= New(int64(8),nil)
	lru.Add("k1", String("v1"))
	lru.Add("k2", String("v2"))
	lru.Add("k3", String("v3"))

	if _,ok:=lru.Get("k1");ok || lru.Len() != 2 {
		t.Fatalf("Removeoldest k1 failed")
	}
}

func TestGet(t *testing.T) {
	lru:= New(int64(0),nil)
	lru.Add("key", String("val"))

	if v,ok:=lru.Get("key");!ok || string(v.(String)) != "val" {
		t.Fatalf("cache hit key=val failed")
	}
}

func TestOnEvicted(t *testing.T) {
	keys:=make([]string,0)
	// 某条记录被移除时的回调函数
	callback:= func(key string,value Value) {
		keys = append(keys, key)
	}
	lru:= New(int64(4),callback)
	lru.Add("k1", String("v1"))
	lru.Add("k2", String("v2"))
	lru.Add("k3", String("v3"))

	expected:=[]string{"k1","k2"}
	if !reflect.DeepEqual(keys,expected) {
		t.Fatal("Call OnEvicted failed, expect keys equals to",expected)
	}
}