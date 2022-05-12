/**
  @author: 黄睿楠
  @since: 2022/5/6
  @desc:
**/

package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 采取依赖注入的方式，允许替换成自定义的 Hash 函数
type Hash func(data []byte) uint32

// Map 一致性哈希算法的主数据结构
type Map struct {
	hash Hash   // 哈希算法
	replicas int   // 虚拟节点倍数
	hashCycle []int   // 哈希环，保存所有虚拟节点的hash值
	hashMap map[int]string   // 虚拟节点与真实节点的映射表，键是虚拟节点的哈希值，值是真实节点的名称
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash: fn,
		replicas: replicas,
		hashMap: make(map[int]string),
	}
	if m.hash==nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 将多个真实节点扩展成虚拟节点，然后计算hash值，并保存到哈希环上
func (m *Map) Add(peerNames ...string) {   // 允许传入0或多个真实节点的名称
	// _是index
	for _,peerName := range peerNames {
		for i:=0;i<m.replicas;i++ {
			// strconv.Itoa()函数的参数是一个整型数字，它可以将数字转换成对应的字符串类型的数字
			hash:=int(m.hash([]byte(strconv.Itoa(i)+peerName)))
			m.hashCycle=append(m.hashCycle,hash)
			m.hashMap[hash] = peerName
		}
	}
	// 环上的哈希值排序
	sort.Ints(m.hashCycle)
}

// Get 根据传入的参数key，返回最近的虚拟节点，对应的真实节点的名称
func (m *Map) Get(key string) string {
	if len(key)==0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	// 使用二分法来搜索某指定切片[0:n]，并返回能够使f(i)=true的最小的i,（0<=i<n），如果无法找到该i，则该方法为返回n
	idx := sort.Search(len(m.hashCycle), func(i int) bool {
		return m.hashCycle[i] >= hash
	})
	// m.hashCycle是一个int切片，如果hash(key)比切片中的所有元素都大，根据哈希环的思想，这个key应该选择第0个节点
	// 在二分搜索时，找不到满足 m.hashCycle[i] >= hash 的i，所以idx=len(m.hashCycle)，取余后为0
	return m.hashMap[m.hashCycle[idx%len(m.hashCycle)]]
}