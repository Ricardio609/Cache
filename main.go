/* 测试http服务端 */

package main

import (
	"net/http"
	"fmt"
	"log"
	"geecache"
)

//使用map模拟数据源db
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main()  {
	//创建一个名为scores的Group,若缓存为空，回调函数会从db中获取数据并返回
	geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	peers := geecache.NewHTTPPool(addr)				//使用该函数在9999端口启动HTTP服务
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}