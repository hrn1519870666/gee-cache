/**
  @author: 黄睿楠
  @since: 2022/5/6
  @desc: 缓存值的抽象与封装
**/

package geecache

type ByteView struct {
	// 存储真实的缓存值。选择byte类型是为了能够支持任意的数据类型的存储，例如字符串、图片等。
	b []byte
}

// Len 实现Value接口
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 返回一个拷贝，防止缓存值被外部程序修改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func cloneBytes(b []byte) []byte{
	c:=make([]byte,len(b))
	copy(c,b)
	return c
}

func (v ByteView) String() string {
	return string(v.b)
}
