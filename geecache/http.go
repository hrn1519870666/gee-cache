/**
  @author: 黄睿楠
  @since: 2022/5/6
  @desc: 提供被其他节点访问的能力(基于http)
**/

package geecache

import (
	"Gee-Cache/geecache/consistenthash"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	pb "Gee-Cache/geecache/geecachepb"
	"github.com/golang/protobuf/proto"
)

const (
	defaultBasePath = "/cache/"
	defaultReplicas = 50
)

// Peer HTTP服务端类
type Peer struct {
	addr     string
	basePath string
	mu          sync.Mutex
	m       *consistenthash.Map
	// 每一个远程节点对应一个 httpGetter，因为 httpGetter 与远程节点的地址 baseURL 有关
	// key是远程节点的名称，例如http://localhost:8001
	httpGetters map[string]*httpGetter
}

func NewPeer(addr string) *Peer {
	return &Peer{
		addr:     addr,
		basePath: defaultBasePath,
	}
}

func (p *Peer) Set(peerNames...string)  {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.m=consistenthash.New(defaultReplicas,nil)
	p.m.Add(peerNames...)
	p.httpGetters=make(map[string]*httpGetter,len(peerNames))
	for _,peerName := range peerNames {
		// 为每一个节点创建了一个HTTP客户端 httpGetter
		p.httpGetters[peerName] = &httpGetter{baseURL: peerName+p.basePath}
	}
}

// PickPeer 封装了一致性哈希算法的 Get() 方法，根据具体的key，选择节点，返回节点对应的HTTP客户端
func (p *Peer) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peerName:=p.m.Get(key);peerName != "" && peerName != p.addr {
		p.Log("Pick peer %s", peerName)
		return p.httpGetters[peerName], true
	}
	return nil, false
}

var _ PeerPicker = (*Peer)(nil)

// Log info with server name
func (p *Peer) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.addr, fmt.Sprintf(format, v...))
}

// ServeHTTP handle all http requests
func (p *Peer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}



// HTTP客户端类
type httpGetter struct {
	// 将要访问的远程节点的地址，例如 http://localhost:8001/cache/
	baseURL string
}

func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}

	return nil
}

var _ PeerGetter = (*httpGetter)(nil)