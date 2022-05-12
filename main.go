/**
  @author: 黄睿楠
  @since: 2022/5/6
  @desc:
**/

package main

import (
	"Gee-Cache/geecache"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var db=map[string]string{
	"Tom":  "100",
	"Jack": "90",
	"Sam":  "80",
}

func createGroup() *geecache.Group {
	return geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

// 启动缓存服务器，用户不感知
func startCacheServer(addr string,addrs []string,gee *geecache.Group)  {
	// 创建 Peer
	peer := geecache.NewPeer(addr)
	// 添加节点信息
	peer.Set(addrs...)
	// 注册到 gee 中
	gee.RegisterPeers(peer)
	log.Println("geecache is running at", addr)
	// 启动 HTTP 服务
	log.Fatal(http.ListenAndServe(addr[7:], peer))
}

// 启动一个 API 服务（端口 9999），与用户进行交互，用户感知
func startAPIServer(apiAddr string, gee *geecache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := gee.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())

		}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

// main() 函数需要命令行传入 port 和 api 2 个参数，用来在指定端口启动 HTTP 服务
func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()
	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}
	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	gee := createGroup()
	if api {
		go startAPIServer(apiAddr, gee)
	}
	startCacheServer(addrMap[port], addrs, gee)
}

