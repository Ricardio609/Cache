package lru
/*
 *Cache缓存淘汰策略：lru
 */
import "container/list"

/*
 * LRU有两个核心数据结构：
 * 1. 字典，存储键和值的映射关系。根据某个键(key)查找对应的值(value)的复杂是O(1)，在字典中插入一条记录的复杂度也是O(1)
 * 2. 双向链表实现的队列，将所有的值存储在双向链表中。当访问到某个值时，将其移动到队尾的复杂度是O(1)，在队尾新增一条记录以及删除一条记录的复杂度均为O(1)
 */

type Cache struct {
	maxBytes	int64											//允许使用的最大内存
	nbytes	int64												//当前已使用的内存

	ll	*list.List												//直接利用Go标准库实现的双向链表list.List
	cache	map[string]*list.Element
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value)						//某条记录被移除时的回调函数，可以是nil
}

/* 键值对entry是双向链表节点的类型。在链表中仍保存每个值对应的 key 的好处在于，淘汰队首节点时，需要用 key 从字典中删除对应的映射 */
type entry struct {
	key string 
	value Value
}

/* 为了通用性，这里允许值是实现了 Value 接口的任意类型，该接口只包含了一个方法 Len() int，用于返回值所占用的内存大小。不太理解。*/
type Value interface {
	Len()	int													//使用len来计算它使用了多少字节
}


/* cache的构造函数*/
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:	maxBytes,
		ll:			list.New(),
		cache:		make(map[string]*list.Element),
		OnEvicted:	onEvicted,
	}
}

/* 查找一个键的值 */
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {								//如果键对应着的链表节点（即值）存在，那么将对应节点移动到队尾，并返回查找到的值
		c.ll.MoveToFront(ele)										//MoveToFront()，将链表中的节点ele移动到队尾。注意：双向对队列作为队列，队首队尾是相对的，这里以front为队尾
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

/* 删除最近最少访问的节点（队首），即缓存淘汰 */
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()												//取队首节点，从链表中删除
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)										//从字典c.cache中删除该节点的映射关系
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())		//更新当前所用内存
		if c.OnEvicted  != nil {									//如果回调函数不为nil,则调用回调函数
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

/* 在cache上添加和修改*/
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {								//如果键存在，则更新对应节点的值，并移动该节点到队尾
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)				
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {														//如果不存在，则是添加节点
		ele := c.ll.PushFront(&entry{key, value})					//在队尾Fron添加新节点，并在字典中更新key与节点间的映射关系
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {					//更新nbytes，如果超过最大值maxByetes,则移除最少访问的节点
		c.RemoveOldest()
	}
}

/* 为了便于测试，实现Len(), 用来获取添加了多少数据 */
func (c *Cache) Len() int {
	return c.ll.Len()
}