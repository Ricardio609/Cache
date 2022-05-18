/*
 *单机并发缓存
 *并发控制。
 *为lru.Cache添加并发特性。
 *实现：实例化lru，封装get与add方法，并添加互斥锁。
 */

 package geecache
 
 import (
	 "geecache/lru"
	"sync"
 )

 type cache struct {
	 mu				sync.Mutex
	 lru			*lru.Cache
	 cacheBytes		int64
 }

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {									//延迟初始化（lazy initialization），一个对象的延迟初始化意味着该对象的创建将会延迟至第一次使用该对象时。主要用于提高性能，并减少程序内存要求。
		c.lru = lru.New(c.cacheBytes, nil)				//判断了 c.lru 是否为 nil，如果等于 nil 再创建实例
	}
	c.lru.Add(key, value)
}



 func (c *cache) get(key string) (value ByteView, ok bool) {
	 c.mu.Lock()
	 defer c.mu.Unlock()
	 if c.lru == nil {
		 return
	 }

	 if v, ok := c.lru.Get(key); ok {
		 return v.(ByteView), ok
	 }

	 return
 } 