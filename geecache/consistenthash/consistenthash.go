package consistenthash
/* 实现一致性哈希算法 */

import (
	"strconv"
	"sort"
	"hash/crc32"
)

/* 定义函数类型hash, 采取依赖注入的方式，允许用于替换自定义的Hash函数，也便于测试时替换，默认为crc32。ChecksumIEEE算法*/
type Hash func(data []byte) uint32

/* Map是一致性哈希算法的主结构函数 */
type Map struct {
	hash	 	Hash				//hash函数
	replicas	int					//虚拟节点倍数
	keys		[]int				//Sorted。哈希环
	hashMap		map[int]string		//虚拟节点与真实节点的映射表，其中键是虚拟节点的哈希值，值是真实节点的名称
}

/* 构造函数，允许自定义虚拟节点倍数和Hash函数*/
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas:	replicas,
		hash:		fn,
		hashMap:	make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

/* 添加真实节点/机器 */
func (m *Map) Add(keys ...string) {									//允许传入0个或多个真实节点的名称
	for _, key := range keys {										//对每一个真实节点 key，对应创建 m.replicas 个虚拟节点
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))		//虚拟节点的名称是：strconv.Itoa(i) + key，即通过添加编号的方式区分不同虚拟节点
			m.keys = append(m.keys, hash)							//使用 m.hash() 计算虚拟节点的哈希值，使用 append(m.keys, hash) 添加到环上
			m.hashMap[hash] = key									//在 hashMap 中增加虚拟节点和真实节点的映射关系
		}
	}
	sort.Ints(m.keys)												//环上的哈希值排序
}

/* 选择节点 */
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))						//计算hash值

	idx := sort.Search(len(m.keys), func(i int) bool {		//顺时针找到第一个匹配的虚拟节点的下标 idx，从 m.keys 中获取到对应的哈希值
		return m.keys[i] >= hash							//如果 idx == len(m.keys)，说明应选择 m.keys[0]，因为 m.keys 是一个环状结构，所以用取余数的方式来处理这种情况
	})

	return m.hashMap[m.keys[idx % len(m.keys)]]				//通过 hashMap 映射得到真实的节点
}