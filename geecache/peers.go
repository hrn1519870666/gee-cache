/**
  @author: 黄睿楠
  @since: 2022/5/9
  @desc:
**/

package geecache

import pb "Gee-Cache/geecache/geecachepb"

type PeerPicker interface {
	// PickPeer 根据传入的 key 选择相应节点 PeerGetter
	PickPeer(key string) (peer PeerGetter,ok bool)
}

// PeerGetter 节点类型
type PeerGetter interface {
	// Get 从对应 group 查找缓存值
	Get(in *pb.Request, out *pb.Response) error
}