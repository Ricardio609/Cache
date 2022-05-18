/* 
 * 提供被其他节点访问的能力（基于http) 。
 * 分布式缓存需要实现节点间通信，建立基于 HTTP 的通信机制是比较常见和简单的做法。
 * 如果一个节点启动了 HTTP 服务，那么这个节点就可以被其他节点访问。
 * 接下来为单机节点搭建 HTTP Server，这部分代码不与其他部分耦合，单独放在新的 http.go 中
 */
package geecache

import (
	"log"
	"strings"
	"net/http"
	"fmt"

)

const defaultBasePath = "/_geecache/"

type HTTPPool struct {				//该结构体作为承载节点间HTTP通信的核心数据结构（包括服务端和客户端），目前只实现服务端

	self		string				//记录自己的地址，包括主机名/IP和端口
	basePath	string				//节点间通讯地址的前缀，默认是`/_geecache/`
}

/* 初始化peers的HTTP池 */
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:		self,
		basePath:	defaultBasePath,
	}
}

/* 带有服务器名称的log信息 */
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

/* 核心方法。处理所有的HTTP请求 */
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {						//实现判断访问的路径前缀是否是basePath，若不是，则返回错误
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)

	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	
	groupName := parts[0]
	key := parts[1]

	//约定访问路径格式为 /<basepath>/<groupname>/<key>，通过 groupname 得到 group 实例，再使用 group.Get(key) 获取缓存数据
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: " + groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//最终使用 w.Write() 将缓存值作为 httpResponse 的 body 返回
	w.Header().Set("content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}
