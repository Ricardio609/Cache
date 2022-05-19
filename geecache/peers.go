/*
 *geecache 的流程（2）：geecache.go文件中首次提到，这里细化该过程。
 * 
 *使用一致性哈希选择节点        是                                    是
 *  |-----> 是否是远程节点 -----> HTTP 客户端访问远程节点 --> 成功？-----> 服务端返回返回值
 *                   |  否                                    ↓  否
 *                   |----------------------------> 回退到本地节点处理。
 */

 package geecache
 
 /* 接口。根据传入的key选择相应的节点 */
 type PeerPicker interface {
	 PickPeer(key string) (peer PeerGetter, ok bool)
 }

 /* 接口。从对应group查找缓存值。该方法对应于上述流程中的HTTP客户端 */
 type PeerGetter interface {
	 Get(group string, key string) ([]byte, error)
 }

 

