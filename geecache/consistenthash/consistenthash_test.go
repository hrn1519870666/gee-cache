/**
  @author: 黄睿楠
  @since: 2022/5/6
  @desc:
**/

package consistenthash

import (
	"strconv"
	"testing"
)

func TestConsistentHash(t *testing.T) {
	// 要进行测试，那么我们需要明确地知道每一个传入的 key 的哈希值，那使用默认的 crc32.ChecksumIEEE 算法显然达不到目的
	// 所以在这里使用了自定义的 Hash 算法
	m := New(3, func(key []byte) uint32 {
		// 将字符串转换为数字
		i,_ := strconv.Atoi(string(key))
		return uint32(i)
	})
	// 扩展出的虚拟节点为: 02,12,22   04,14,24   06,16,26
	// 排序之后的哈希环: 2,4,6,12,14,16,22,24,26
	m.Add("2","4","6")

	// 键为要查找的key，值为该key对应的真实节点
	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	for k,v := range testCases{
		if m.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}
	// 添加一个真实节点 8，对应虚拟节点的哈希值是 08/18/28
	// 此时，用例 27 对应的虚拟节点从 02 变更为 28，即真实节点 8
	m.Add("8")
	testCases["27"]="8"
	for k,v := range testCases{
		if m.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}
}
