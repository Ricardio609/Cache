# Cache(分布式缓存)
分布式缓存系统需要：
 * 速度，最关键的指标
 * 容量，越高表明可以适用于更多的场景
 * 低GC开销，键值缓存是临时留在堆上的，GC会一次又一次的对其进行扫描，因此不影响GC也是很重要的
 * 分布式与缓存一致。这是属于一个系统更高级的问题，单机意味着崩溃后就要重新再来，缓存一致性则会影响读取正确性。
 * 正确性，这其实是应该放在第一条的，一个错误百出的系统是毫无价值的。


goCache

1. 缓存数据结构  lru

```
资源限制问题
```
![image](https://user-images.githubusercontent.com/64991294/170391659-ee8a1009-47e1-4c2b-865d-ca7bd9272bd8.png)


2. 单机并发缓存
![image](https://user-images.githubusercontent.com/64991294/170391758-f9265e19-6641-4e72-b388-ebd9d284dc0d.png)
实现多个缓存表
核心数据结构：group

3. HTTP服务端
4. 一致性哈希（consistent hashing)
一致性哈希抽象的解释就是一个很大的环
```
远程节点选择问题。
```
![image](https://user-images.githubusercontent.com/64991294/170392063-816718c8-8843-4ed0-9445-3ed49d5ea8bf.png)


* consistent hashing 是从单节点走向分布式的一个重要环节(对于给定的key, 每一次都选择同一个节点）

  * 节点数量固定时，当一个节点收到请求时，如果该节点并没有存储缓存值，利用自定义的hash算法来实现将某些查找分配给固定的节点，避免：1）缓存效率低；2）各个节点上存储着相同的数据，浪费大量的存储空间。
  * 节点数量发生变化时，意味着几乎缓存值对应的节点都发生了改变（即几乎所有的缓存值都失效了），节点在接收到对应的请求时，均需要重新取数据源取数据，容易引起`缓存雪崩`。
    * `缓存雪崩`：缓存在同一时刻全部失效，造成瞬时DB请求量大、压力骤增，引起雪崩。长因为缓存服务器宕机，或缓存设置了相同的过期时间引起。
* consistent hashing算法原理

  * 将key映射到2^32的空间中，并将这个数字首尾相连，形成一个环，如下图。
    * 计算节点/机器（通常使用节点的名称、编号和IP地址）的哈希值，放置在环上
    * 计算key的哈希值，放置在环上，顺时针找到的第一个节点，就是应选取的节点/机器

![](https://geektutu.com/post/geecache-day4/add_peer.jpg)

* 该算法在新增/删除节点时，只需要重新定位该节点附件的一小部分数据，而不需要重新定位所有的节点。
* 该算法存在`数据倾斜`的问题
  * 如果服务器的节点过少，则容易引发key的倾斜。例如上面的图片中，peeer2、4、6分布在环的上半部分，下部分是空的，这就导致映射到下半部分的key都会被分配给peer2，key过度向peer2倾斜，缓存节点负载不均。
  * 针对该问题，引入虚拟节点（一个真实节点对应多个虚拟节点）

    * 示例：假设1个真实节点对应3个虚拟节点，那么peer1对应的虚拟节点是peer1-1、peer1-2、peer1-3（通常以添加编号的方式实现），其余节点也以相同的方式操作
    * 第一步：计算虚拟节点的hash值，放置在环上
    * 第二部：计算key的hash值，在环上顺时针寻找应选去的虚拟节点，例如是peer2-1，那么就对应着真实节点peer2
  * 引入虚拟节点的优势：虚拟节点扩充了节点的数量，解决了节点数量较少的情况下数据容易倾斜的问题。代价非常小，只需要增加一个字典(map)维护真实节点与虚拟节点的映射关系即可
 * 不足：当有一个真实节点添加进来的时候，所有值都要重新计算一遍。这在并发情况下，会造成一定拥塞。因为在重新计算期间，不能进行正确的访问操作。
5. HTTP客户端（多节点通信）

* 注册节点，并借助一致性哈希算法选择节点
* 实现HTTP客户端，与远程节点的服务端通信

哈希一致性算法能够保证同一个key每次访问的节点都相同，但这样会导致缓存被击穿，即一个key同时大量请求某个节点，并且请求的结果一致。

6. 缓存击穿

> 缓存雪崩：缓存在同一时间内全部失效，造成瞬时DB请求量大、压力骤增，引起雪崩。通常因为缓存服务器宕机、缓存的key设置了相同的过期时间等引起。
>
> 缓存击穿：一个存在的key，在缓存过期的一刻，同时有大量的请求，而这些请求都会击穿到DB，造成瞬时DB请求量大、压力骤增。
>
> 缓存穿透：查询一个不存在的数据，因为不存在则不会写道缓存中，所有每次都会去请求DB,如果瞬时流量过大，穿透到DB，会导致宕机。

采用singleflight，实现在并发场景下，同一个key,多次请求时，只发起一次请求。
