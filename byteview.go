/* 
 * byteview:只读，用来保存缓存值
 */

package Cache

/* 不可变的字节视图 */
type ByteView struct {
	b []type									/// b存储真实的缓存值。此处选择type类型是为了能够支持任意数据类型的存储，例如字符串、图片等。
}

/* 返回视图的长度 */
func (v ByteView) Len() int {					//实现Len()方法。在 lru.Cache 的实现中，要求被缓存对象必须实现 Value 接口，即 Len() int 方法，返回其所占的内存大小。
	return len(v.b)
}

/* 以字节切片的形式返回数据的副本 */
func (v ByteView) ByteSlice() []byte {			//b是只读的，利用ByteSlice()方法返回一个拷贝，防止缓存值被外部程序修改。
	return cloneBytes(v.b)
}

/* 将数据作为字符串返回，必要时进行复制 */
func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}