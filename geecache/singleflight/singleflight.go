package singleflight

import (
	"sync"
)

/* call表示正在进行中或已结束的请求 */
type call struct {
	wg		sync.WaitGroup		//使用WaitGroup锁避免重入
	val		interface{}
	err		error
}

/* singleflight的主结构函数，管理不同key的请求（call) */
type Group struct {
	mu sync.Mutex							//并发协程之间不需要消息传递，非常适合 sync.WaitGroup
	m	map[string]*call
}

/* 针对相同的key，无论Do被调用多少次，函数fn都只会被调用一次，等待fn调用结束，返回返回值或错误 */
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error)  {
	g.mu.Lock()								//保护Group的成员变量m不被并发读写而加上的锁
	if g.m == nil {							//延迟初始化，以提高内存使用效率
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()							//如果请求正在进行中，则等待。阻塞，指导锁被释放
		return c.val, c.err					//请求结束，返回结果
	}
	c := new(call)
	c.wg.Add(1)								//发起请求前加锁。Add(1):锁加1
	g.m[key] = c							//添加到g.m，表明key已有对应的请求在处理
	g.mu.Unlock()

	c.val, c.err = fn()						//调用fn，发起请求
	c.wg.Done()								//请求结束。锁减一

	g.mu.Lock()
	delete(g.m, key)						//更新g.m
	g.mu.Unlock()

	return c.val, c.err						//返回结果
}
