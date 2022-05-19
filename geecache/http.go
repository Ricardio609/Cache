/* 
 * 提供被其他节点访问的能力（基于http) 。
 * 分布式缓存需要实现节点间通信，建立基于 HTTP 的通信机制是比较常见和简单的做法。
 * 如果一个节点启动了 HTTP 服务，那么这个节点就可以被其他节点访问。
 * 接下来为单机节点搭建 HTTP Server，这部分代码不与其他部分耦合，单独放在新的 http.go 中
 */
package geecache

import (
	"geecache/consistenthash"
	"sync"
	"io/ioutil"
	"net/url"
	"log"
	"strings"
	"net/http"
	"fmt"

)

/* 服务端 */

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50					//客户端
)

/* HTTPPool的功能: 1. 提供 HTTP 服务的能力; 2. 根据具体的 key，创建 HTTP 客户端从远程节点获取缓存值的能力*/
type HTTPPool struct {						//该结构体作为承载节点间HTTP通信的核心数据结构（包括服务端和客户端）

	self			string					//记录自己的地址，包括主机名/IP和端口
	basePath		string					//节点间通讯地址的前缀，默认是`/_geecache/`
	
	//以下变量是为了给HTTPPool添加节点选择的功能
	mu				sync.Mutex				//客户端。保护peers和httpGetter
	peers			*consistenthash.Map		//客户端。 类型是一致性哈希算法的Map，用来根据具体的key选择节点
	httpGetters		map[string]*httpGetter	//客户端。keyed by e.g. "http://10.0.0.2:8008"。映射远程节点与对应httpGetter。每个远程节点对应一个httpGetter，因为httpGetter与远程节点的地址baseURL有关
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

/* 客户端 */

type httpGetter struct {
	baseURL	string				//baseURL 表示将要访问的远程节点的地址，例如 http://example.com/_geecache/
}

func (h *httpGetter) Get(group string, key string) ([]byte, error)  {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	res, err := http.Get(u)		//使用 http.Get() 方式获取返回值，并转换为 []bytes 类型

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading responese body: %v", err)
	}

	return bytes, nil
}

var _ PeerGetter = (*httpGetter)(nil)

/* 实例化一个一致性哈希算法，并添加传入的节点 */
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {											//为每个节点创建一个HTTP客户端的httpGetter
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

/*  包装了一致性哈希算法的 Get() 方法，根据具体的 key，选择节点，返回节点对应的 HTTP 客户端 */
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)