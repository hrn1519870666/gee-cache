/**
  @author: 黄睿楠
  @since: 2022/5/6
  @desc:
**/

package geecache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

var db=map[string]string{
	"Tom":  "100",
	"Jack": "90",
	"Sam":  "80",
}

func TestGetter(t *testing.T) {
	// 借助 GetterFunc 的类型转换，将一个匿名回调函数转换成了接口 f Getter
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	// 调用该接口的方法 f.Get(key string)，实际上就是在调用匿名回调函数
	// 对于array、slice、map、struct等类型，当比较两个值是否相等时，不能使用==
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("callback failed")
	}
}

func TestGet(t *testing.T) {

	// 使用 loadCounts 统计每个键调用回调函数的次数，如果次数大于1，则表示调用了多次回调函数，没有缓存
	loadCounts := make(map[string]int, len(db))
	gee := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key]++
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	for k, v := range db {
		// 第一次执行gee.Get(k)，缓存不命中，通过回调函数查数据库，查到之后将数据写到缓存中
		if view, err := gee.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		}
		// 第二次执行gee.Get(k)，缓存命中，打印 [GeeCache] hit
		if _, err := gee.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}

	if view, err := gee.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}

func TestGetGroup(t *testing.T) {
	groupName := "scores"
	NewGroup(groupName, 2<<10, GetterFunc(
		func(key string) (bytes []byte, err error) { return }))
	if group := GetGroup(groupName); group == nil || group.name != groupName {
		t.Fatalf("group %s not exist", groupName)
	}

	if group := GetGroup(groupName + "111"); group != nil {
		t.Fatalf("expect nil, but %s got", group.name)
	}
}
