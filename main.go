// /* 测试http服务端 */

package main

import (
	"flag"
	"net/http"
	"log"
	"fmt"
	"geecache"
)

// import (
// 	"net/http"
// 	"fmt"
// 	"log"
// 	"geecache"
// )

// //使用map模拟数据源db
// var db = map[string]string{
// 	"Tom":  "630",
// 	"Jack": "589",
// 	"Sam":  "567",
// }

// func main()  {
// 	//创建一个名为scores的Group,若缓存为空，回调函数会从db中获取数据并返回
// 	geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
// 		func(key string) ([]byte, error) {
// 			log.Println("[SlowDB] search key", key)
// 			if v, ok := db[key]; ok {
// 				return []byte(v), nil
// 			}
// 			return nil, fmt.Errorf("%s not exist", key)
// 		}))

// 	addr := "localhost:9999"
// 	peers := geecache.NewHTTPPool(addr)				//使用该函数在9999端口启动HTTP服务
// 	log.Println("geecache is running at", addr)
// 	log.Fatal(http.ListenAndServe(addr, peers))
// }


/* 测试客户端 */

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
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

/* 启动缓存服务器 */
func startCacheServer(addr string, addrs []string, gee *geecache.Group) {
	peers := geecache.NewHTTPPool(addr)						//创建HTTPPool
	peers.Set(addrs...)										//添加节点信息
	gee.RegisterPeers(peers)								//注册到gee中
	log.Println("geecache is running at", addr)				//启动http服务（共3个端口，8001/8002/8003），用户不感知
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

/* 启动一个API服务（端口9999），与用户进行交互，用户感知 */
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
	startCacheServer(addrMap[port], []string(addrs), gee)
}