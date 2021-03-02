# Redis

## 1. <code>string</code>-字符串
字符串结构使用非常广泛，**一个常见的用途就是缓存用户信息**。我们将用户信息结构体使用 <code>JSON</code> 序列化成字符串，然后将序列化后的字符串塞进 Redis 来缓存。同样，取用户信息会经过一次反序列化的过程。

<code>Redis</code> 的字符串是动态字符串，是可以修改的字符串，内部结构实现上类似于 Java 的 ArrayList，采用预分配冗余空间的方式来减少内存的频繁分配，如图中所示，++内部为当前字符串实际分配的空间<code> capacity </code>一般要高于实际字符串长度 len++。    
<span style="color:red">当字符串长度小于 1M 时，扩容都是加倍现有的空间，如果超过 1M，扩容时一次只会多扩 1M 的空间。需要注意的是字符串最大长度为 512M</span>。
> 创建字符串时 <code>len</code> 和 <code>capacity</code> 一样长，不会多分配冗余空间，这是因为绝大多数场景下我们不会使用 <code>append</code> 操作来修改字符串

```redis
> set name codehole 
OK
> exists name // 是否存在
(integer) 1
> del name // 删除
(integer) 1
> get name
(nil)
> mset name1 boy name2 girl name3 unknown --- 批量操作
> mget name1 name2 name3 # 返回一个列表
1) "codehole"
2) "holycoder"
3) (nil)

// 计数 -- 如果 value 值是一个整数，还可以对它进行自增操作。
// 自增是有范围的，它的范围是 signed long 的最大最小值，超过了这个值，Redis 会报错。
> set age 30
OK
> incr age
(integer) 31
> incrby age 5
(integer) 36
> incrby age -5
(integer) 31
> set codehole 9223372036854775807  # Long.Max
OK
> incr codehole
(error) ERR increment or decrement would overflow
```
#### 过期和set命令扩展
```
// 过期
> expire name 5  # 5s 后过期
...  # wait for 5s
> get name
(nil)
> ttl name
-1

> setex name 5 codehole  # 5s 后过期，等价于 set+expire

> setnx name codehole  # 如果 name 不存在就执行 set 创建
(integer) 1
> get name
"codehole"
> setnx name holycoder
(integer) 0  # 因为 name 已经存在，所以 set 创建不成功

// set 指令扩展
> set lock:codehole true ex 5 nx
OK
```
##### 过期的 key 集合
redis 会将每个设置了过期时间的 <code>key</code> 放入到一个独立的字典中，以后会定时遍历这个字典来删除到期的 key。          
除了<code style="color:red">定时遍历</code>之外，它还会使<code style="color:red">用惰性策略</code>来删除过期的 key，
> 所谓惰性策略就是在客户端访问这个<code> key</code> 的时候，<code>redis</code> 对 <code>key</code> 的过期时间进行检查，如果过期了就立即删除。

<span style="color:red">**定时删除是集中处理，惰性删除是零散处理**。</span>
##### 定时扫描策略
<code>Redis</code> 默认会每秒进行十次过期扫描，过期扫描不会遍历过期字典中所有的<code>key</code>，而是采用了一种简单的**贪心策略**。
1. 从过期字典中随机 20 个 key；
1. 删除这 20 个 key 中已经过期的 key；
1. 如果过期的 key 比率超过 1/4，那就重复步骤 1；

> 同时，为了保证过期扫描不会出现循环过度，导致线程卡死现象，算法还增加了扫描时间的上限，默认不会超过<code style="color:red">25ms</code>。

<u>设想一个大型的 Redis 实例中所有的 key 在同一时间过期了，会出现怎样的结果？</u>

> 毫无疑问，Redis 会持续扫描过期字典 (循环多次)，直到过期字典中过期的 key 变得稀疏，才会停止 (循环次数明显下降)。这就会导致线上读写请求出现明显的卡顿现象。导致这种卡顿的另外一种原因是内存管理器需要频繁回收内存页，这也会产生一定的 CPU 消耗。     
当客户端请求到来时，服务器如果正好进入过期扫描状态，客户端的请求将会等待至少 25ms 后才会进行处理，如果客户端将超时时间设置的比较短，比如 10ms，那么就会出现大量的链接因为超时而关闭，业务端就会出现很多异常。而且这时你还无法从 Redis 的 slowlog 中看到慢查询记录，因为慢查询指的是逻辑处理过程慢，不包含等待时间。

所以业务开发人员一定要注意过期时间，如果有大批量的 key 过期，要给过期时间设置一个随机范围，而不宜全部在同一时间过期，分散过期处理的压力。

```
redis.expire_at(key, random.randint(86400) + expire_ts)
```
##### 从库的过期策略
从库不会进行过期扫描，从库对过期的处理是被动的。主库在 key 到期时，会在 AOF 文件里增加一条 <code> del </code> 指令，同步到所有的从库，从库通过执行这条<code> del </code>指令来删除过期的 key。

> **++字符串是由多个字节组成，每个字节又是由 8 个 bit 组成++**，如此便可以将一个字符串看成很多 bit 的组合，这便是 <code>bitmap「位图」</code>数据结构

#### <code>string</code>内部结构
<u>Redis 中的字符串是可以修改的字符串，在内存中它是以字节数组的形式存在的。</u>        
我们知道 C 语言里面的字符串标准形式是以 NULL 作为结束符，但是在 Redis 里面字符串不是这么表示的。因为要获取 NULL 结尾的字符串的长度使用的是 strlen 标准库函数，这个函数的算法复杂度是 O(n)，它需要对字节数组进行遍历扫描，作为单线程的 Redis 表示承受不起。

Redis 的字符串叫着<code>「SDS」</code>，也就是<code>Simple Dynamic String</code>。它的结构是一个带长度信息的字节数组。

```redis
struct SDS<T> {
  T capacity; // 数组容量 - 1byte
  T len; // 数组长度 - 1byte
  byte flags; // 特殊标识位，不理睬它 - 1byte
  byte[] content; // 数组内容
}
```

![image](https://mail.wangkekai.cn/D1DA79F5-652B-4DD4-B665-A4A4616176D6.png)

<code>capacity</code> **表示所分配数组的长度**，<code>len</code> **表示字符串的实际长度**。前面我们提到字符串是可以修改的字符串，它要支持 <code>append</code> 操作。如果数组没有冗余空间，那么追加操作必然涉及到分配新数组，然后将旧内容复制过来，再 append 新内容。如果字符串的长度非常长，这样的内存分配和复制开销就会非常大。

> 上面的 <code>SDS</code> 结构使用了范型 <code>T</code>，为什么不直接用 <code>int</code> 呢，这是因为当字符串比较短时，<code>len</code> 和 <code>capacity</code> 可以使用 <code>byte</code> 和 <code>short</code> 来表示，Redis 为了对内存做极致的优化，不同长度的字符串使用不同的结构体来表示。

##### embstr vs raw
<code>Redis</code> 的字符串有两种存储方式，在长度特别短时，使用 <code>emb</code> 形式存储 <code>(embeded)</code>，当长度超过 44 时，使用 <code>raw</code> 形式存储。

![image](https://mail.wangkekai.cn/9528D65C-6E40-4157-9797-280B11D23E1A.png)

<code>embstr</code> 存储形式是这样一种存储形式，它将 <code>RedisObject</code> 对象头和 <code>SDS</code> 对象连续存在一起，使用 <code>malloc</code> 方法一次分配。而 <code>raw</code> 存储形式不一样，它需要两次 <code>malloc</code>，两个对象头在内存地址上一般是不连续的。

## 2. <code>list</code> - 列表
<code>Redis</code> 的列表相当于 <code>Java</code> 语言里面的 <code>LinkedList</code>，<span style="color:red">注意它是链表而不是数组</span>。      
**++这意味着 <code>list</code> 的插入和删除操作非常快，时间复杂度为 <code>O(1)</code>，但是索引定位很慢，时间复杂度为 <code>O(n)</code>，这点让人非常意外++**。

当列表弹出了最后一个元素之后，该数据结构自动被删除，内存被回收。     
**++Redis 的列表结构常用来做异步队列使用++**。将需要延后处理的任务结构体序列化成字符串塞进 Redis 的列表，另一个线程从这个列表中轮询数据进行处理。

```redis
// 队列 - 右进左出
> rpush books python java golang
(integer) 3
> llen books
(integer) 3
> lpop books
"python"
...
> lpop books
(nil)

// 栈 - 右进右出
> rpush books python java golang
(integer) 3
> rpop books
"golang"
...
> rpop books
(nil)
```

<code>lindex</code> 相当于 Java 链表的<code>get(int index)</code>方法，它需要对链表进行遍历，性能随着参数<code>index</code>增大而变差。

<code>ltrim</code> 和字面上的含义不太一样，个人觉得它叫 <code>lretain(保留)</code> 更合适一些，因为 <code>ltrim</code> 跟的两个参数<code>start_index和end_index</code>定义了一个区间，在这个区间内的值，<code>ltrim</code> 要保留，区间之外统统砍掉。我们可以通过<code>ltrim</code>来实现一个定长的链表，这一点非常有用。

<code>index</code> 可以为负数，<code>index=-1</code>表示倒数第一个元素，同样<code>index=-2</code>表示倒数第二个元素。

```redis
> rpush books python java golang
(integer) 3
> lindex books 1  # O(n) 慎用
"java"
> lrange books 0 -1  # 获取所有元素，O(n) 慎用
1) "python"
2) "java"
3) "golang"
> ltrim books 1 -1 # O(n) 慎用
OK
> lrange books 0 -1
1) "java"
2) "golang"
> ltrim books 1 0 # 这其实是清空了整个列表，因为区间范围长度为负
OK
> llen books
(integer) 0
```

##### 快速列表
如果再深入一点，你会发现 <code>Redis</code> 底层存储的还不是一个简单的 <code>linkedlist</code>，而是称之为快速链表 <code>quicklist</code> 的一个结构。

++首先在列表元素较少的情况下会使用一块连续的内存存储++，这个结构是 <code>ziplist</code>，也即是<span style="color:red">压缩列表</span>。**它将所有的元素紧挨着一起存储，<u>分配的是一块连续的内存</u>**。当数据量比较多的时候才会改成 <code>quicklist</code>。因为普通的链表需要的附加指针空间太大，会比较浪费空间，而且会加重内存的碎片化。比如这个列表里存的只是 int 类型的数据，结构上还需要两个额外的指针 <code>prev</code> 和 <code>next</code> 。所以 Redis 将链表和 <code>ziplist</code> 结合起来组成了 <code>quicklist</code>。也就是将多个 ziplist 使用双向指针串起来使用。这样既满足了快速的插入删除性能，又不会出现太大的空间冗余。

#### 压缩列表
Redis 为了节约内存空间使用，<code>zset</code> 和 <code>hash</code> 容器对象在元素个数较少的时候，采用<code>压缩列表 (ziplist)</code> 进行存储。     
**<u>压缩列表是一块连续的内存空间，元素之间紧挨着存储，没有任何冗余空隙</u>。**

```redis
//  encoding 字段都是 ziplist
> zadd programmings 1.0 go 2.0 python 3.0 java
(integer) 3
> debug object programmings
Value at:0x7fec2de00020 refcount:1 encoding:ziplist serializedlength:36 lru:6022374 lru_seconds_idle:6
> hmset books go fast python slow java fast
OK
> debug object books
Value at:0x7fec2de000c0 refcount:1 encoding:ziplist serializedlength:48 lru:6022478 lru_seconds_idle:1
```

```redis
struct ziplist<T> {
    int32 zlbytes; // 整个压缩列表占用字节数
    int32 zltail_offset; // 最后一个元素距离压缩列表起始位置的偏移量，用于快速定位到最后一个节点
    int16 zllength; // 元素个数
    T[] entries; // 元素内容列表，挨个挨个紧凑存储
    int8 zlend; // 标志压缩列表的结束，值恒为 0xFF
}
```
> 压缩列表为了支持双向遍历，所以才会有 ztail_offset这个字段，用来快速定位到最后一个元素，然后倒着遍历。
##### 增加元素

因为 <code>ziplist</code> 都是紧凑存储，没有冗余空间 (对比一下 <code>Redis</code> 的字符串结构)。意味着插入一个新的元素就需要调用 <code>realloc</code> 扩展内存。取决于内存分配器算法和当前的 <code>ziplist</code> 内存大小，<code>realloc</code> 可能会重新分配新的内存空间，并将之前的内容一次性拷贝到新的地址，也可能在原有的地址上进行扩展，这时就不需要进行旧内容的内存拷贝。

如果 <code>ziplist</code> 占据内存太大，重新分配内存和拷贝内存就会有很大的消耗。所以 <code style="color:red">ziplist 不适合存储大型字符串</code>，存储的元素也不宜过多。

##### <code>IntSet</code> 小整数集合
当 <code>set</code> 集合容纳的元素都是整数并且元素个数较小时，<code>Redis</code> 会使用 <code>intset</code> 来存储结合元素。<code>intset</code> 是紧凑的数组结构，同时支持 16 位、32 位和 64 位整数。

```redis
struct intset<T> {
    int32 encoding; // 决定整数位宽是 16 位、32 位还是 64 位
    int32 length; // 元素个数
    int<T> contents; // 整数数组，可以是 16 位、32 位和 64 位
}
```

#### 快速列表
<code>Redis</code> 早期版本存储 <code>list</code> 列表数据结构使用的是压缩列表 <code>ziplist</code> 和普通的双向链表 <code>linkedlist</code>，也就是元素少时用 <code>ziplist</code>，元素多时用 <code>linkedlist</code>。

考虑到链表的附加空间相对太高，<code>prev</code> 和 <code>next</code> 指针就要占去 16 个字节 (64bit 系统的指针是 8 个字节)，另外每个节点的内存都是单独分配，会加剧内存的碎片化，影响内存管理效率。后续版本对列表数据结构进行了改造，使用 <code>quicklist</code> 代替了 <code>ziplist</code> 和 <code>linkedlist</code>。

```redis
> rpush codehole go java python
(integer) 3
> debug object codehole
Value at:0x7fec2dc2bde0 refcount:1 encoding:quicklist serializedlength:31 lru:6101643 lru_seconds_idle:5 ql_nodes:1 ql_avg_node:3.00 ql_ziplist_max:-2 ql_compressed:0 ql_uncompressed_size:29
```

**注意观察上面输出字段 <code>encoding</code> 的值。<code>quicklist</code> 是 <code>ziplist</code> 和 <code>linkedlist</code> 的混合体，它将 <code>linkedlist</code> 按段切分，每一段使用 <code>ziplist</code> 来紧凑存储，多个 <code>ziplist</code> 之间使用双向指针串接起来。**
![image](https://mail.wangkekai.cn/7AAB35D9-BBDB-46C0-879F-04746234203D.png)

##### 每个 ziplist 存多少元素？
<code style="color:red">quicklist 内部默认单个 ziplist 长度为 8k 字节，超出了这个字节数，就会新起一个 ziplist。</code>ziplist 的长度由配置参数<code>list-max-ziplist-size</code>决定。

##### 压缩深度
![image](https://mail.wangkekai.cn/71ABE178-6788-4362-90F7-EB1F74495E9E.png)
<code>quicklist</code> 默认的压缩深度是 0，也就是不压缩。压缩的实际深度由配置参数<code>list-compress-depth</code>决定。为了支持快速的 <code>push/pop</code> 操作，<code>quicklist</code> 的首尾两个 <code>ziplist</code> 不压缩，此时深度就是 1。如果深度为 2，就表示 <code>quicklist</code> 的首尾第一个 <code>ziplist</code> 以及首尾第二个 <code>ziplist</code> 都不压缩。

## 3. <code>hash</code> - 字典

Redis 的字典相当于 Java 语言里面的 <code>HashMap</code>，它是<span style="color:red">无序字典</span>。内部实现结构上同 Java 的 <code>HashMap</code> 也是一致的，**<u>同样的数组 + 链表二维结构。第一维 hash 的数组位置碰撞时，就会将碰撞的元素使用链表串接起来</u>**。
![image](https://mail.wangkekai.cn/999C4B1B-8CD5-42E5-BF26-74022C9B2326.png)

Redis 的字典的值只能是字符串，另外它们 <code>rehash</code> 的方式不一样，因为 Java 的 HashMap 在字典很大时，rehash 是个耗时的操作，需要一次性全部 rehash。<code style="color:red">Redis 为了高性能，不能堵塞服务，所以采用了渐进式 rehash 策略</code>。
![image](https://mail.wangkekai.cn/C5E5EF39-9A78-4D29-A1F7-C3983C4F4A58.png)

渐进式 rehash 会在 rehash 的同时，保留新旧两个 hash 结构，查询时会同时查询两个 hash 结构，然后在后续的定时任务中以及 hash 操作指令中，循序渐进地将旧 hash 的内容一点点迁移到新的 hash 结构中。当搬迁完成了，就会使用新的hash结构取而代之。

> 当 hash 移除了最后一个元素之后，该数据结构自动被删除，内存被回收。

hash 结构也可以用来存储用户信息，不同于字符串一次性需要全部序列化整个对象，hash 可以对用户结构中的每个字段单独存储。这样当我们需要获取用户信息时可以进行部分获取。而以整个字符串的形式去保存用户信息的话就只能一次性全部读取，这样就会比较浪费网络流量。

> hash 也有缺点，hash 结构的存储消耗要高于单个字符串，到底该使用 hash 还是字符串，需要根据实际情况再三权衡。
```
> hset books java "think in java"  # 命令行的字符串如果包含空格，要用引号括起来
(integer) 1
// ...
> hgetall books  # entries()，key 和 value 间隔出现
1) "java"
2) "think in java"
3) "golang"
4) "concurrency in go"
5) "python"
6) "python cookbook"
> hlen books
(integer) 3
> hget books java
"think in java"
> hset books golang "learning go programming"  # 因为是更新操作，所以返回 0
(integer) 0
> hget books golang
"learning go programming"
> hmset books java "effective java" python "learning python" golang "modern golang programming"  # 批量 set
OK

// 单个字符计数
> hincrby user-laoqian age 1
```

#### <code>dict</code> - 内部结构
dict 是 Redis 服务器中出现最为频繁的复合型数据结构，除了 hash 结构的数据会用到字典外，**++整个 <code>Redis</code> 数据库的所有 <code>key</code> 和 <code>value</code> 也组成了一个全局字典++**，还有带过期时间的 key 集合也是一个字典。    
**++zset 集合中存储 value 和 score 值的映射关系也是通过 dict 结构实现的++。**
```
struct RedisDb {
    dict* dict; // all keys  key=>value
    dict* expires; // all expired keys key=>long(timestamp)
    ...
}

struct zset {
    dict *dict; // all values  value=>score
    zskiplist *zsl;
}
```

![image](https://mail.wangkekai.cn/0EE01A26-57D5-455A-A842-0586AA9D885D.png)

<code style="color:red">dict 结构内部包含两个 hashtable，通常情况下只有一个 hashtable 是有值的</code>。但是在 <code>dict</code> 扩容缩容时，需要分配新的 <code>hashtable</code>，然后进行渐进式搬迁，这时候两个 <code>hashtable</code> 存储的分别是旧的 <code>hashtable</code> 和新的 <code>hashtable</code>。待搬迁结束后，旧的 <code>hashtable</code> 被删除，新的 <code>hashtable</code> 取而代之。

##### 渐进式 <code>rehash</code>

大字典的扩容是比较耗时间的，需要重新申请新的数组，然后将旧字典所有链表中的元素重新挂接到新的数组下面，这是一个O(n)级别的操作，作为单线程的Redis表示很难承受这样耗时的过程。步子迈大了会扯着蛋，所以Redis使用渐进式rehash小步搬迁。虽然慢一点，但是肯定可以搬完。

##### 查找过程
插入和删除操作都依赖于查找，先必须把元素找到，才可以进行数据结构的修改操作。hashtable 的元素是在第二维的链表上，所以首先我们得想办法定位出元素在哪个链表上。

##### 扩容条件
<span style="color:red">正常情况下，当 hash 表中元素的个数等于第一维数组的长度时，就会开始扩容，扩容的新数组是原数组大小的 2 倍。</span>不过如果 Redis 正在做 <code>bgsave</code>，为了减少内存页的过多分离 <code>(Copy On Write)</code>，Redis 尽量不去扩容 <code>(dict_can_resize)</code>，但是如果 hash 表已经非常满了，元素的个数已经达到了第一维数组长度的 5 倍 <code>(dict_force_resize_ratio)</code>，说明 hash 表已经过于拥挤了，这个时候就会强制扩容。

##### 缩容条件
当 hash 表因为元素的逐渐删除变得越来越稀疏时，Redis 会对 hash 表进行缩容来减少 hash 表的第一维数组空间占用。    
<span style="color:red">缩容的条件是元素个数低于数组长度的 10%。缩容不会考虑 Redis 是否正在做 bgsave。</span>

## 4. <code>set</code> - 集合
Redis 的集合相当于 Java 语言里面的 <code>HashSet</code>，**++它内部的键值对是无序的唯一的++**。它的内部实现相当于一个特殊的字典，字典中所有的 <code>value</code> 都是一个值 <code>NULL</code>。
> 当集合中最后一个元素移除之后，数据结构自动删除，内存被回收。

set 结构可以用来存储活动中奖的用户 ID，因为有去重功能，可以保证同一个用户不会中奖两次。

```
> sadd books python
(integer) 1
> sadd books python  #  重复
(integer) 0
> sadd books java golang
(integer) 2
> smembers books  # 注意顺序，和插入的并不一致，因为 set 是无序的
1) "java"
2) "python"
3) "golang"
> sismember books java  # 查询某个 value 是否存在，相当于 contains(o)
(integer) 1
> sismember books rust
(integer) 0
> scard books  # 获取长度相当于 count()
(integer) 3
> spop books  # 弹出一个
"java"
```

## 5. <code>zset</code> - 有序集合
它类似于 Java 的 <code>SortedSet</code> 和 <code>HashMap</code> 的结合体，一方面它是一个 set，保证了内部 <code>value</code> 的唯一性，另一方面它可以给每个 <code>value</code> 赋予一个 <code>score</code>，代表这个 <code>value</code> 的排序权重。它的内部实现用的是一种叫做<code>「跳跃列表」</code>的数据结构。

<code>zset</code> 可以用来存粉丝列表或学生的成绩，<code>value</code> 值是粉丝的用户/学生 ID，<code>score</code> 是关注时间/成绩。我们可以对粉丝列表按关注时间/成绩进行排序。
```
> zadd books 9.0 "think in java"
(integer) 1
// ...
> zrange books 0 -1  # 按 score 排序列出，参数区间为排名范围
1) "java cookbook"
2) "java concurrency"
3) "think in java"
> zrevrange books 0 -1  # 按 score 逆序列出，参数区间为排名范围
// ...
> zcard books  # 相当于 count()
(integer) 3
> zscore books "java concurrency"  # 获取指定 value 的 score
"8.9000000000000004"  # 内部 score 使用 double 类型进行存储，所以存在小数点精度问题
> zrank books "java concurrency"  # 排名
(integer) 1
> zrangebyscore books 0 8.91  # 根据分值区间遍历 zset
1) "java cookbook"
2) "java concurrency"
> zrangebyscore books -inf 8.91 withscores # 根据分值区间 (-∞, 8.91] 遍历 zset，同时返回分值。inf 代表 infinite，无穷大的意思。
1) "java cookbook"
2) "8.5999999999999996"
3) "java concurrency"
4) "8.9000000000000004"
> zrem books "java concurrency"  # 删除 value
(integer) 1
> zrange books 0 -1
1) "java cookbook"
2) "think in java"
```

#### 跳跃列表
zset 内部的排序功能是通过「跳跃列表」数据结构来实现的，它的结构非常特殊，也比较复杂。

因为 zset 要支持随机的插入和删除，所以它不好使用数组来表示。我们先看一个普通的链表结构。

![image](https://mail.wangkekai.cn/230C465E-1643-4B51-ACC7-C16D772F8512.png)

我们需要这个链表按照 score 值进行排序。这意味着当有新元素需要插入时，要定位到特定位置的插入点，这样才可以继续保证链表是有序的。通常我们会通过二分查找来找到插入点，但是二分查找的对象必须是数组，只有数组才可以支持快速位置定位，链表做不到，那该怎么办？

> 跳跃列表就是类似于这种层级制，最下面一层所有的元素都会串起来。然后每隔几个元素挑选出一个代表来，再将这几个代表使用另外一级指针串起来。然后在这些代表里再挑出二级代表，再串起来。最终就形成了金字塔结构。

![image](https://mail.wangkekai.cn/57E68050-B9D5-476E-BF90-B2FA661018F4.png)

「跳跃列表」之所以「跳跃」，是因为内部的元素可能「身兼数职」，比如上图中间的这个元素，同时处于 L0、L1 和 L2 层，可以快速在不同层次之间进行「跳跃」。

定位插入点时，先在顶层进行定位，然后下潜到下一级定位，一直下潜到最底层找到合适的位置，将新元素插进去。你也许会问，那新插入的元素如何才有机会「身兼数职」呢？

跳跃列表采取一个随机策略来决定新元素可以兼职到第几层。

首先 L0 层肯定是 100% 了，L1 层只有 50% 的概率，L2 层只有 25% 的概率，L3 层只有 12.5% 的概率，一直随机到最顶层 L31 层。绝大多数元素都过不了几层，只有极少数元素可以深入到顶层。列表中的元素越多，能够深入的层次就越深，能进入到顶层的概率就会越大。

## 6. 分布式锁
分布式锁本质上要实现的目标就是在 Redis 里面占一个“茅坑”，当别的进程也要来占时，发现已经有人蹲在那里了，就只好放弃或者稍后再试。

占坑一般是使用 setnx(set if not exists) 指令，只允许被一个客户端占坑。
如果在 <code>setnx</code> 和 <code>expire</code> 之间服务器进程突然挂掉了，可能是因为机器掉电或者是被人为杀掉的，就会导致 <code>expire</code> 得不到执行，也会造成死锁。

++<code>Redis 2.8</code> 版本中作者加入了 <code>set</code> 指令的扩展参数，使得 <code>setnx</code> 和 <code>expire</code> 指令可以一起执行，彻底解决了分布式锁的乱象++。
```
// 这里的冒号:就是一个普通的字符，没特别含义，它可以是任意其它字符，不要误解
> setnx lock:codehole true
OK
... do something critical ...
> del lock:codehole
(integer) 1
```

比如在 Sentinel 集群中，主节点挂掉时，从节点会取而代之，客户端上却并没有明显感知。原先第一个客户端在主节点中申请成功了一把锁，但是这把锁还没有来得及同步到从节点，主节点突然挂掉了。然后从节点变成了主节点，这个新的节点内部没有这个锁，所以当另一个客户端过来请求加锁时，立即就批准了。这样就会导致系统中同样一把锁被两个客户端同时持有，不安全性由此产生。

##### Redlock 算法

为了使用 <code>Redlock</code>，需要提供多个 <code>Redis</code> 实例，这些实例之前相互独立没有主从关系。同很多分布式算法一样，<code>redlock</code> 也使用<code>「大多数机制」</code>。

==加锁时==，它会向过半节点发送 <code>set(key, value, nx=True, ex=xxx)</code> 指令，<code style="color:red">只要过半节点 set 成功，那就认为加锁成功</code>。释放锁时，需要向所有节点发送 del 指令。不过 Redlock 算法还需要考虑出错重试、时钟漂移等很多细节问题，同时因为 Redlock 需要向多个节点进行读写，意味着相比单实例 Redis 性能会下降一些。

##### Redlock 使用场景
<span style="color:red">在乎高可用性</span>，希望挂了一台 redis 完全不受影响，那就应该考虑 redlock。不过代价也是有的，需要更多的 redis 实例，性能也下降了，代码上还需要引入额外的 library，运维上也需要特殊对待，这些都是需要考虑的成本。

## 7. 延时队列
常用 <code>Rabbitmq</code> 和 <code>Kafka</code> 作为消息队列中间件进行异步消息传递。

#### 异步消息队列
<code>Redis</code> 的 <code>list(列表)</code> 数据结构常用来作为异步消息队列使用，使用<code>rpush/lpush</code>操作入队列，使用<code>lpop 和 rpop</code>来出队列。
![image](https://mail.wangkekai.cn/1607610238093.jpg)

```
> rpush notify-queue apple banana pear
(integer) 3
> llen notify-queue
(integer) 3
> lpop notify-queue
"apple"
> llen notify-queue
(integer) 2
> lpop notify-queue
"banana"
> llen notify-queue
(integer) 1
> lpop notify-queue
"pear"
> llen notify-queue
(integer) 0
> lpop notify-queue
(nil)
```

#### 队列延迟
可是如果队列空了，客户端就会陷入 <code>pop</code> 的死循环，不停地 <code>pop</code>，没有数据，接着再 <code>pop</code>，又没有数据。这就是浪费生命的空轮询。空轮询不但拉高了客户端的 <code>CPU</code>，<code>redis</code> 的 <code>QPS</code> 也会被拉高，如果这样空轮询的客户端有几十来个，<code>Redis</code> 的慢查询可能会显著增多。

<span style="color:red">阻塞读在队列没有数据的时候，会立即进入休眠状态，一旦数据到来，则立刻醒过来。</span>消息的延迟几乎为零。用<code>blpop/brpop</code>替代前面的<code>lpop/rpop</code>，就完美解决了上面的问题。

##### 空闲连接自动断开
如果线程一直阻塞在哪里，Redis 的客户端连接就成了闲置连接，闲置过久，服务器一般会主动断开连接，减少闲置资源占用。这个时候blpop/brpop会抛出异常来。 <span style="color:red">注意捕获异常，还要重试。</span>

<span style="color:red">延时队列可以通过 <code>Redis</code> 的 <code>zset(有序列表)</code> 来实现。</span>**我们将消息序列化成一个字符串作为 zset 的value，这个消息的到期处理时间作为score，然后用多个线程轮询 zset 获取到期的任务进行处理，多个线程是为了保障可用性，万一挂了一个线程还有其它线程可以继续处理**。因为有多个线程，所以需要考虑并发争抢任务，确保任务不能被多次执行。

```
def delay(msg):
    msg.id = str(uuid.uuid4())  # 保证 value 值唯一
    value = json.dumps(msg)
    retry_ts = time.time() + 5  # 5 秒后重试
    redis.zadd("delay-queue", retry_ts, value)


def loop():
    while True:
        # 最多取 1 条
        values = redis.zrangebyscore("delay-queue", 0, time.time(), start=0, num=1)
        if not values:
            time.sleep(1)  # 延时队列空的，休息 1s
            continue
        value = values[0]  # 拿第一条，也只有一条
        success = redis.zrem("delay-queue", value)  # 从消息队列中移除该消息
        if success:  # 因为有多进程并发的可能，最终只会有一个进程可以抢到消息
            msg = json.loads(value)
            handle_msg(msg)
```
<code style="color:red">Redis 的 zrem 方法是多线程多进程争抢任务的关键</code>，它的返回值决定了当前实例有没有抢到任务，因为 <code>loop</code> 方法可能会被多个线程、多个进程调用，同一个任务可能会被多个进程线程抢到，通过 <code>zrem</code> 来决定唯一的属主。

## 8. 位图
<code>Redis 提供了位图数据结构</code>，这样每天的签到记录只占据一个位，365 天就是 365 个位，46 个字节 (一个稍长一点的字符串) 就可以完全容纳下，这就大大节约了存储空间。

位图不是特殊的数据结构，它的内容其实就是普通的字符串，也就是 <code>byte 数组</code>。我们可以使用普通的 <code>get/set</code> 直接获取和设置整个位图的内容，也可以使用位图操作 <code>getbit/setbit</code> 等将 byte 数组看成「位数组」来处理。

#### 统计和查找
Redis 提供了位图统计指令 <code>bitcount</code> 和位图查找指令 <code>bitpos</code>，<code>bitcount</code> 用来统计指定位置范围内 1 的个数，<code>bitpos</code> 用来查找指定范围内出现的第一个 0 或 1。

## 9. HyperLogLog
Redis 提供了 <code>HyperLogLog</code> 数据结构就是用来解决这种统计问题的。<code>HyperLogLog</code> 提供不精确的去重计数方案，虽然不精确但是也不是非常不精确，标准误差是 0.81%，这样的精确度已经可以满足上面的 UV 统计需求了。

HyperLogLog 数据结构是 Redis 的高级数据结构，它非常有用，但是令人感到意外的是，使用过它的人非常少。
#### 使用方法
HyperLogLog 提供了两个指令 <code>pfadd</code> 和 <code>pfcount</code>，根据字面意义很好理解，**++一个是增加计数，一个是获取计数++**。
```
127.0.0.1:6379> pfadd codehole user1
(integer) 1
127.0.0.1:6379> pfcount codehole
(integer) 1
// ...
127.0.0.1:6379> pfadd codehole user7 user8 user9 user10
(integer) 1
127.0.0.1:6379> pfcount codehole
(integer) 10
```

#### pfmerge 适合什么场合用？
HyperLogLog 除了上面的 <code>pfadd</code> 和 <code>pfcount</code> 之外，还提供了第三个指令 <code style="color:red">pfmerge</code>，用于将多个 pf 计数值累加在一起形成一个新的 pf 值。

#### 注意事项
HyperLogLog 它需要占据一定<code>12k</code>的存储空间，所以它不适合统计单个用户相关的数据。如果你的用户上亿，可以算算，这个空间成本是非常惊人的。但是相比 set 存储方案，HyperLogLog 所使用的空间那真是可以使用千斤对比四两来形容了。

不过你也不必过于担心，因为 Redis 对 HyperLogLog 的存储进行了优化，**<span style="color:red">在计数比较小时，它的存储空间采用稀疏矩阵存储，空间占用很小</span>，仅仅在计数慢慢变大，稀疏矩阵占用空间渐渐超过了阈值时才会一次性转变成稠密矩阵，才会占用 12k 的空间**。

## 10. 布隆过滤器
布隆过滤器可以理解为一个不怎么精确的 <code>set</code> 结构，当你使用它的 <code>contains</code> 方法判断某个对象是否存在时，它可能会误判。但是布隆过滤器也不是特别不精确，只要参数设置的合理，它的精确度可以控制的相对足够精确，只会有小小的误判概率。    
<code style="color:red">当布隆过滤器说某个值存在时，这个值可能不存在；当它说不存在时，那就肯定不存在。</code>
#### 布隆过滤器基本使用
布隆过滤器有二个基本指令，<code>bf.add 添加元素</code>，<code>bf.exists 查询元素是否存在</code>，它的用法和 <code>set</code> 集合的 <code>sadd</code> 和 <code>sismember</code> 差不多。注意 <code>bf.add</code> 只能一次添加一个元素，如果想要一次添加多个，就需要用到 <code>bf.madd</code> 指令。同样如果需要一次查询多个元素是否存在，就需要用到 <code>bf.mexists 指令</code>。
```
127.0.0.1:6379> bf.add codehole user1
(integer) 1
// ...
127.0.0.1:6379> bf.exists codehole user1
(integer) 1
// ...
127.0.0.1:6379> bf.exists codehole user4
(integer) 0
127.0.0.1:6379> bf.madd codehole user4 user5 user6
1) (integer) 1
2) (integer) 1
3) (integer) 1
127.0.0.1:6379> bf.mexists codehole user4 user5 user6 user7
1) (integer) 1
2) (integer) 1
3) (integer) 1
4) (integer) 0
```
## 11. 简单限流
系统要限定用户的某个行为在指定的时间里只能允许发生 N 次，如何使用 Redis 的数据结构来实现这个限流的功能？

<span style="color:red">这个限流需求中存在一个滑动时间窗口</span>，想想 <code>zset</code> 数据结构的 <code>score</code> 值，是不是可以通过 <code>score</code> 来圈出这个时间窗口来。**++而且我们只需要保留这个时间窗口，窗口之外的数据都可以砍掉++**。那这个 zset 的 value 填什么比较合适呢？它只需要保证唯一性即可，用 uuid 会比较浪费空间，那就改用毫秒时间戳吧。

```
def is_action_allowed(user_id, action_key, period, max_count):
    key = 'hist:%s:%s' % (user_id, action_key)
    now_ts = int(time.time() * 1000)  # 毫秒时间戳
    with client.pipeline() as pipe:  # client 是 StrictRedis 实例
        # 记录行为
        pipe.zadd(key, now_ts, now_ts)  # value 和 score 都使用毫秒时间戳
        # 移除时间窗口之前的行为记录，剩下的都是时间窗口内的
        pipe.zremrangebyscore(key, 0, now_ts - period * 1000)
        # 获取窗口内的行为数量
        pipe.zcard(key)
        # 设置 zset 过期时间，避免冷用户持续占用内存
        # 过期时间应该等于时间窗口的长度，再多宽限 1s
        pipe.expire(key, period + 1)
        # 批量执行
        _, _, current_count, _ = pipe.execute()
    # 比较数量是否超标
    return current_count <= max_count


for i in range(20):
    print is_action_allowed("laoqian", "reply", 60, 5)
```
![image](https://mail.wangkekai.cn/F6E160B6-E181-4ACE-95BA-B46717C97E3D.png)

用一个 <code>zset</code> 结构记录用户的行为历史，每一个行为都会作为 <code>zset</code> 中的一个 <code>key</code> 保存下来。同一个用户同一种行为用一个 <code>zset</code> 记录。    
为节省内存，我们只需要保留时间窗口内的行为记录，同时如果用户是冷用户，滑动时间窗口内的行为是空记录，那么这个 zset 就可以从内存中移除，不再占用空间。

## 12. 漏斗限流
Redis 4.0 提供了一个限流 Redis 模块，它叫 <code>redis-cell</code>。该模块也使用了漏斗算法，并提供了原子的限流指令。有了这个模块，限流问题就非常简单了。

```
> cl.throttle laoqian:reply 15 30 60 1
                      ▲     ▲  ▲  ▲  ▲
                      |     |  |  |  └───── need 1 quota (可选参数，默认值也是1)
                      |     |  └──┴─────── 30 operations / 60 seconds 这是漏水速率
                      |     └───────────── 15 capacity 这是漏斗容量
                      └─────────────────── key laoqian
```
上面这个指令的意思是允许「用户老钱回复行为」的频率为每 60s 最多 30 次(漏水速率)，漏斗的初始容量为 15，也就是说一开始可以连续回复 15 个帖子，然后才开始受漏水速率的影响。我们看到这个指令中漏水速率变成了 2 个参数，替代了之前的单个浮点数。用两个参数相除的结果来表达漏水速率相对单个浮点数要更加直观一些。

```
> cl.throttle laoqian:reply 15 30 60
1) (integer) 0   # 0 表示允许，1表示拒绝
2) (integer) 15  # 漏斗容量capacity
3) (integer) 14  # 漏斗剩余空间left_quota
4) (integer) -1  # 如果拒绝了，需要多长时间后再试(漏斗有空间了，单位秒)
5) (integer) 2   # 多长时间后，漏斗完全空出来(left_quota==capacity，单位秒)
```

在执行限流指令时，如果被拒绝了，就需要丢弃或重试。cl.throttle 指令考虑的非常周到，连重试时间都帮你算好了，直接取返回结果数组的第四个值进行 sleep 即可，如果不想阻塞线程，也可以异步定时任务来重试。

## 13. GeoHash
<code>GeoHash</code> 算法将二维的经纬度数据映射到一维的整数，这样所有的元素都将在挂载到一条线上，距离靠近的二维坐标映射到一维后的点之间距离也会很接近。

在使用 <code>Redis</code> 进行 <code>Geo</code> 查询时，我们要时刻想到它的内部结构实际上只是一个 <code>zset(skiplist)</code>。通过 <code>zset</code> 的 <code>score</code> 排序就可以得到坐标附近的其它元素 (实际情况要复杂一些，不过这样理解足够了)，通过将 <code>score</code> 还原成坐标值就可以得到元素的原始坐标。

#### <code>Geo</code> 基本指令使用
##### 增加

<code>geoadd</code> 指令携带集合名称以及多个经纬度名称三元组，注意这里可以加入多个三元组

```
127.0.0.1:6379> geoadd company 116.48105 39.996794 juejin
(integer) 1
// ...
127.0.0.1:6379> geoadd company 116.562108 39.787602 jd 116.334255 40.027400 xiaomi
(integer) 2
```
> 也许你会问为什么 <code>Redis</code> 没有提供 <code>geo</code> 删除指令？前面我们提到 <code>geo</code> 存储结构上使用的是 <code>zset</code>，意味着我们可以使用 <code>zset</code> 相关的指令来操作 <code>geo</code> 数据，所以删除指令可以直接使用 <code>zrem</code> 指令即可。

##### 距离

<code>geodist</code> 指令可以用来计算两个元素之间的距离，携带集合名称、2 个名称和距离单位。

```
127.0.0.1:6379> geodist company juejin ireader km
"10.5501"
// ...
127.0.0.1:6379> geodist company juejin juejin km
"0.0000"
```
##### 获取元素位置

<code>geopos</code> 指令可以获取集合中任意元素的经纬度坐标，可以一次获取多个。

```
127.0.0.1:6379> geopos company juejin
1) 1) "116.48104995489120483"
   2) "39.99679348858259686"
// ...
127.0.0.1:6379> geopos company juejin ireader
1) 1) "116.48104995489120483"
   2) "39.99679348858259686"
2) 1) "116.5142020583152771"
   2) "39.90540918662494363"
```

##### 获取元素的 hash 值

<code>geohash</code> 可以获取元素的经纬度编码字符串，上面已经提到，它是 <code>base32</code> 编码。 你可以使用这个编码值去 geohash.org/${hash}中进行直… <code>geohash</code> 的标准编码值。

```
127.0.0.1:6379> geohash company ireader
1) "wx4g52e1ce0"
127.0.0.1:6379> geohash company juejin
1) "wx4gd94yjn0"
```
##### 附近的公司

<code style="color:red">georadiusbymember</code> 指令是最为关键的指令，它可以用来查询指定元素附近的其它元素，它的参数非常复杂。

```
# 范围 20 公里以内最多 3 个元素按距离正排，它不会排除自身
127.0.0.1:6379> georadiusbymember company ireader 20 km count 3 asc
1) "ireader"
2) "juejin"
3) "meituan"
# 范围 20 公里以内最多 3 个元素按距离倒排
127.0.0.1:6379> georadiusbymember company ireader 20 km count 3 desc
1) "jd"
2) "meituan"
3) "juejin"
# 三个可选参数 withcoord withdist withhash 用来携带附加参数
# withdist 很有用，它可以用来显示距离
127.0.0.1:6379> georadiusbymember company ireader 20 km withcoord withdist withhash count 3 asc
1) 1) "ireader"
   2) "0.0000"
   3) (integer) 4069886008361398
   4) 1) "116.5142020583152771"
      2) "39.90540918662494363"
2) 1) "juejin"
   2) "10.5501"
   3) (integer) 4069887154388167
   4) 1) "116.48104995489120483"
      2) "39.99679348858259686"
3) 1) "meituan"
   2) "11.5748"
   3) (integer) 4069887179083478
   4) 1) "116.48903220891952515"
      2) "40.00766997707732031"
```
除了 <code>georadiusbymember</code> 指令根据元素查询附近的元素，<code>Redis</code> 还提供了根据坐标值来查询附近的元素，这个指令更加有用，它可以根据用户的定位来计算「附近的车」，「附近的餐馆」等。它的参数和 <code>georadiusbymember</code> 基本一致，除了将目标元素改成经纬度坐标值。


```
127.0.0.1:6379> georadius company 116.514202 39.905409 20 km withdist count 3 asc
1) 1) "ireader"
   2) "0.0000"
2) 1) "juejin"
   2) "10.5501"
3) 1) "meituan"
   2) "11.5748"
```
## 14. <code>scan</code>
在平时线上 Redis 维护工作中，有时候需要从 Redis 实例成千上万的 key 中找出特定前缀的 key 列表来手动处理数据，可能是修改它的值，也可能是删除 key。这里就有一个问题，如何从海量的 key 中找出满足特定前缀的 key 列表来？

<code>keys</code> 用来列出所有满足特定正则字符串规则的 <code>key</code>。

这个指令使用非常简单，提供一个简单的正则字符串即可，但是有很明显的两个缺点。

1. 没有 <code>offset、limit</code> 参数，一次性吐出所有满足条件的 <code>key</code>，万一实例中有几百 w 个 <code>key</code> 满足条件，当你看到满屏的字符串刷的没有尽头时，你就知道难受了。
2. <code>keys</code> 算法是遍历算法，复杂度是 <code>O(n)</code>，如果实例中有千万级以上的 <code>key</code>，这个指令就会导致 <code>Redis</code> 服务卡顿，所有读写 <code>Redis</code> 的其它的指令都会被延后甚至会超时报错，因为 <code>Redis</code> 是单线程程序，顺序执行所有指令，其它指令必须等到当前的 <code>keys</code> 指令执行完了才可以继续。
3. 
> <code>scan</code> 相比 <code>keys</code> 具备有以下特点:
1. 复杂度虽然也是 <code>O(n)</code>，但是它是通过游标分步进行的，不会阻塞线程;
1. 提供 <code>limit</code> 参数，可以控制每次返回结果的最大条数，<code>limit</code> 只是一个 <code>hint</code>，返回的结果可多可少;
1. 同 keys 一样，它也提供模式匹配功能;
1. 服务器不需要为游标保存状态，游标的唯一状态就是 scan 返回给客户端的游标整数;
1. 返回的结果可能会有重复，需要客户端去重复，这点非常重要;
1. 遍历的过程中如果有数据修改，改动后的数据能不能遍历到是不确定的;
1. 单次返回的结果是空的并不意味着遍历结束，而要看返回的游标值是否为零;

#### 基础使用
<code>scan</code> 参数提供了三个参数，第一个是 <code>cursor</code> 整数值，第二个是 <code>key</code> 的正则模式，第三个是遍历的 <code>limit hint</code>。第一次遍历时，<code>cursor</code> 值为 0，然后将返回结果中第一个整数值作为下一次遍历的 <code>cursor</code>。一直遍历到返回的 <code>cursor</code> 值为 0 时结束。


```
127.0.0.1:6379> scan 0 match key99* count 1000
1) "13976"
2)  1) "key9911"
    2) "key9974"
    // ...
127.0.0.1:6379> scan 13976 match key99* count 1000
1) "1996"
2)  1) "key9982"
    2) "key9997"
    // ...
127.0.0.1:6379> scan 1996 match key99* count 1000
1) "12594"
2) 1) "key9939"
   2) "key9941"
......
127.0.0.1:6379> scan 11687 match key99* count 1000
1) "0"
2)  1) "key9969"
    2) "key998"
    // ...
```
从上面的过程可以看到虽然提供的 <code>limit</code> 是 1000，但是返回的结果只有 10 个左右。++**因为这个 <code>limit</code> 不是限定返回结果的数量，而是限定服务器单次遍历的字典槽位数量(约等于)**++。如果将 <code>limit</code> 设置为 10，你会发现返回结果是空的，<font color=red>但是游标值不为零，意味着遍历还没结束</font>。

```
127.0.0.1:6379> scan 0 match key99* count 10
1) "3072"
2) (empty list or set)
```
#### 字典的结构
在 <code>Redis</code> 中所有的 <code>key</code> 都存储在一个很大的字典中，这个字典的结构和 Java 中的 <code>HashMap</code> 一样，是一维数组 + 二维链表结构，第一维数组的大小总是 <code>2^n(n>=0)</code>，扩容一次数组大小空间加倍，也就是 <code>n++</code>
![image](https://mail.wangkekai.cn/5DED180C-BA39-4CD4-95C3-0A34144E80E0.png)

<code>scan</code> 指令返回的游标就是第一维数组的位置索引，我们将这个位置索引称为槽 (slot)。如果不考虑字典的扩容缩容，直接按数组下标挨个遍历就行了。<code>limit</code> 参数就表示需要遍历的槽位数，<font color=red>++之所以返回的结果可能多可能少，是因为不是所有的槽位上都会挂接链表，有些槽位可能是空的，还有些槽位上挂接的链表上的元素可能会有多个++</font>。每一次遍历都会将 <code>limit</code> 数量的槽位上挂接的所有链表元素进行模式匹配过滤后，一次性返回给客户端。

#### scan 遍历顺序
<code>scan</code> 的遍历顺序非常特别。它不是从第一维数组的第 0 位一直遍历到末尾，而是采用了高位进位加法来遍历。之所以使用这样特殊的方式进行遍历，是考虑到字典的扩容和缩容时避免槽位的遍历重复和遗漏。
![image](https://mail.wangkekai.cn/16469760d12e0cbd.gif)

从动画中可以看出高位进位法从左边加，进位往右边移动，同普通加法正好相反。但是最终它们都会遍历所有的槽位并且没有重复。

#### 字典扩容
Java 中的 HashMap 有扩容的概念，当 loadFactor 达到阈值时，需要重新分配一个新的 2 倍大小的数组，然后将所有的元素全部 rehash 挂到新的数组下面。rehash 就是将元素的 hash 值对数组长度进行取模运算，因为长度变了，所以每个元素挂接的槽位可能也发生了变化。又因为数组的长度是 2^n 次方，所以取模运算等价于位与操作。

#### 渐进式 <code>rehash</code>
Java 的 <code>HashMap</code> 在扩容时会一次性将旧数组下挂接的元素全部转移到新数组下面。如果 <code>HashMap</code> 中元素特别多，**++线程就会出现卡顿现象++**。Redis 为了解决这个问题，它采用渐进式 rehash。

它会同时保留旧数组和新数组，然后在定时任务中以及后续对 <code>hash</code> 的指令操作中渐渐地将旧数组中挂接的元素迁移到新数组上。这意味着要操作处于 <code>rehash</code> 中的字典，需要同时访问新旧两个数组结构。如果在旧数组下面找不到元素，还需要去新数组下面去寻找。

scan 也需要考虑这个问题，对与 rehash 中的字典，它需要同时扫描新旧槽位，然后将结果融合后返回给客户端。

#### 更多的 scan 指令
scan 指令是一系列指令，除了可以遍历所有的 key 之外，还可以对指定的容器集合进行遍历。比如 zscan 遍历 zset 集合元素，hscan 遍历 hash 字典的元素、sscan 遍历 set 集合的元素。

它们的原理同 scan 都会类似的，因为 hash 底层就是字典，set 也是一个特殊的 hash(所有的 value 指向同一个元素)，zset 内部也使用了字典来存储所有的元素内容，所以这里不再赘述。

#### 大 key 扫描
为了避免对线上 <code>Redis</code> 带来卡顿，这就要用到 <code>scan</code> 指令，对于扫描出来的每一个 <code>key</code>，使用 <code>type</code> 指令获得 <code>key</code> 的类型，然后使用相应数据结构的 <code>size</code> 或者 <code>len</code> 方法来得到它的大小，对于每一种类型，保留大小的前 N 名作为扫描结果展示出来。

上面这样的过程需要编写脚本，比较繁琐，不过 <code>Redis</code> 官方已经在 <code>redis-cli</code> 指令中提供了这样的扫描功能，我们可以直接拿来即用。


```
redis-cli -h 127.0.0.1 -p 7001 –-bigkeys
```
如果你担心这个指令会大幅抬升 Redis 的 ops 导致线上报警，还可以增加一个休眠参数。

```
redis-cli -h 127.0.0.1 -p 7001 –-bigkeys -i 0.1
```
上面这个指令每隔 100 条 scan 指令就会休眠 0.1s，ops 就不会剧烈抬升，但是扫描的时间会变长。

## 15. 线程 <code>IO</code> 模型
Redis 是个单线程程序！

### Redis 单线程为什么还能这么快？
因为它所有的数据都在内存中，所有的运算都是内存级别的运算。正因为 Redis 是单线程，所以要小心使用 <code>Redis</code> 指令，对于那些时间复杂度为 <code>O(n)</code> 级别的指令，一定要谨慎使用，一不小心就可能会导致 <code>Redis</code> 卡顿。

#### Redis 单线程如何处理那么多的并发客户端连接？
这个问题，有很多中高级程序员都无法回答，因为他们没听过多路复用这个词汇，不知道 <code>select</code> 系列的事件轮询 <code>API</code>，没用过非阻塞 <code>IO</code>。

#### 非阻塞 IO
当我们调用套接字的读写方法，默认它们是阻塞的，比如 <code>read</code> 方法要传递进去一个参数 <code>n</code>，表示最多读取这么多字节后再返回，如果一个字节都没有，那么线程就会卡在那里，直到新的数据到来或者连接关闭了，<code>read</code> 方法才可以返回，线程才能继续处理。而 <code>write</code> 方法一般来说不会阻塞，除非内核为套接字分配的写缓冲区已经满了，<code>write</code> 方法就会阻塞，直到缓存区中有空闲空间挪出来了。
![image](https://mail.wangkekai.cn/E505D0AC-47FC-4790-9812-5E7BCCF628DE.png)

非阻塞 IO 在套接字对象上提供了一个选项 <code>Non_Blocking</code>，当这个选项打开时，**++读写方法不会阻塞，而是能读多少读多少，能写多少写多少++**。能读多少取决于内核为套接字分配的读缓冲区内部的数据字节数，能写多少取决于内核为套接字分配的写缓冲区的空闲空间字节数。读方法和写方法都会通过返回值来告知程序实际读写了多少字节。

有了非阻塞 <code>IO</code> 意味着线程在读写 <code>IO</code> 时可以不必再阻塞了，读写可以瞬间完成然后线程可以继续干别的事了。

#### 事件轮询 (多路复用)
非阻塞 IO 有个问题，那就是线程要读数据，结果读了一部分就返回了，线程如何知道何时才应该继续读。也就是当数据到来时，线程如何得到通知。写也是一样，如果缓冲区满了，写不完，剩下的数据何时才应该继续写，线程也应该得到通知。
![image](https://mail.wangkekai.cn/5A55BFDD-3FCB-48E1-81DF-566BC251B2CF.png)

事件轮询 API 就是用来解决这个问题的，最简单的事件轮询 API 是select函数，它是操作系统提供给用户程序的 API。输入是读写描述符列表 <code>read_fds & write_fds</code>，输出是与之对应的可读可写事件。同时还提供了一个 <code>timeout</code> 参数，如果没有任何事件到来，那么就最多等待 <code>timeout</code> 时间，线程处于阻塞状态。一旦期间有任何事件到来，就可以立即返回。时间过了之后还是没有任何事件到来，也会立即返回。拿到事件后，线程就可以继续挨个处理相应的事件。处理完了继续过来轮询。于是线程就进入了一个死循环，我们把这个死循环称为事件循环，一个循环为一个周期。

#### 指令队列
Redis 会将每个客户端套接字都关联一个指令队列。客户端的指令通过队列来排队进行顺序处理，先到先服务。

#### 响应队列
Redis 同样也会为每个客户端套接字关联一个响应队列。Redis 服务器通过响应队列来将指令的返回结果回复给客户端。 如果队列为空，那么意味着连接暂时处于空闲状态，不需要去获取写事件，也就是可以将当前的客户端描述符从<code>write_fds</code>里面移出来。等到队列有数据了，再将描述符放进去。避免<code>select</code>系统调用立即返回写事件，结果发现没什么数据可以写。出这种情况的线程会飙高 CPU。
#### 定时任务
服务器处理要响应 IO 事件外，还要处理其它事情。比如定时任务就是非常重要的一件事。如果线程阻塞在 <code>select</code> 系统调用上，定时任务将无法得到准时调度。那 Redis 是如何解决这个问题的呢？

Redis 的定时任务会记录在一个称为最小堆的数据结构中。这个堆中，最快要执行的任务排在堆的最上方。在每个循环周期，Redis 都会将最小堆里面已经到点的任务立即进行处理。处理完毕后，将最快要执行的任务还需要的时间记录下来，这个时间就是<code>select</code>系统调用的<code>timeout</code>参数。因为 Redis 知道未来<code>timeout</code>时间内，没有其它定时任务需要处理，所以可以安心睡眠<code>timeout</code>的时间。

Nginx 和 Node 的事件处理原理和 Redis 也是类似的

## 16. 通信协议


## 17. 持久化
Redis 的持久化机制有两种，<font color=red>第一种是快照，第二种是 AOF 日志</font>。**++快照是一次全量备份，AOF 日志是连续的增量备份++**。

==快照==是内存数据的二进制序列化形式，在存储上非常紧凑，而 ==AOF 日志记录==的是内存数据修改的指令记录文本。

AOF 日志在长期的运行过程中会变的无比庞大，数据库重启时需要加载 AOF 日志进行指令重放，这个时间就会无比漫长。所以需要定期进行 AOF 重写，给 AOF 日志进行瘦身。

### 快照原理
我们知道 Redis 是单线程程序，这个线程要同时负责多个客户端套接字的并发读写操作和内存数据结构的逻辑读写。

在服务线上请求的同时，<code>Redis</code> 还需要进行内存快照，内存快照要求 <code>Redis</code> 必须进行文件 <code>IO</code> 操作，可文件 <code>IO</code> 操作是不能使用多路复用 <code>API</code>。

这意味着单线程同时在服务线上的请求还要进行文件 IO 操作，文件 IO 操作会严重拖垮服务器请求的性能。还有个重要的问题是为了不阻塞线上的业务，就需要边持久化边响应客户端请求。持久化的同时，内存数据结构还在改变，比如一个大型的 hash 字典正在持久化，结果一个请求过来把它给删掉了，还没持久化完呢，这尼玛要怎么搞？

Redis 使用操作系统的多进程 <code>COW(Copy On Write)</code> 机制来实现快照持久化，这个机制很有意思，也很少人知道。多进程 <code>COW</code> 也是鉴定程序员知识广度的一个重要指标。

#### fork(多进程)
<code>Redis</code> 在持久化时会调用 <code>glibc</code> 的函数 <code>fork</code> 产生一个子进程，快照持久化完全交给子进程来处理，父进程继续处理客户端请求。子进程刚刚产生时，它和父进程共享内存里面的代码段和数据段。这时你可以将父子进程想像成一个连体婴儿，共享身体。这是 <code>Linux</code> 操作系统的机制，为了节约内存资源，所以尽可能让它们共享起来。在进程分离的一瞬间，内存的增长几乎没有明显变化。

**子进程做数据持久化，它不会修改现有的内存数据结构，它只是对数据结构进行遍历读取，然后序列化写到磁盘中。但是父进程不一样，它必须持续服务客户端请求，然后对内存数据结构进行不间断的修改。**

随着父进程修改操作的持续进行，越来越多的共享页面被分离出来，内存就会持续增长。但是也不会超过原有数据内存的 2 倍大小。**另外一个 Redis 实例里冷数据占的比例往往是比较高的，所以很少会出现所有的页面都会被分离，被分离的往往只有其中一部分页面。<font color=red>每个页面的大小只有 4K，一个 Redis 实例里面一般都会有成千上万的页面。</font>**

子进程因为数据没有变化，它能看到的内存里的数据在进程产生的一瞬间就凝固了，再也不会改变，这也是为什么 Redis 的持久化叫「快照」的原因。接下来子进程就可以非常安心的遍历数据了进行序列化写磁盘了。

### AOF 原理
**<code>AOF</code> 日志存储的是 <code>Redis</code> 服务器的<font color=red>顺序指令序列，<code>AOF</code> 日志只记录对内存进行修改的指令记录</font>**。

假设 AOF 日志记录了自 Redis 实例创建以来所有的修改性指令序列，那么就可以通过对一个空的 Redis 实例顺序执行所有的指令，也就是「重放」，来恢复 Redis 当前实例的内存数据结构的状态。

Redis 会在收到客户端修改指令后，进行参数校验进行逻辑处理后，如果没问题，就立即将该指令文本存储到 AOF 日志中，也就是<font color=red>先执行指令才将日志存盘</font>。这点不同于leveldb、hbase等存储引擎，它们都是先存储日志再做逻辑处理。

Redis 在长期运行的过程中，AOF 的日志会越变越长。如果实例宕机重启，重放整个 AOF 日志会非常耗时，导致长时间 Redis 无法对外提供服务。所以需要对 AOF 日志瘦身。

#### AOF 重写
Redis 提供了 <code>bgrewriteaof</code> 指令用于对 <code>AOF</code> 日志进行瘦身。<font color=red>其原理就是开辟一个子进程对内存进行遍历转换成一系列 <code>Redis</code> 的操作指令，序列化到一个新的 <code>AOF</code> 日志文件中。</font> ++序列化完毕后再将操作期间发生的增量 <code>AOF</code> 日志追加到这个新的 <code>AOF</code> 日志文件中，追加完毕后就立即替代旧的 <code>AOF</code> 日志文件了，瘦身工作就完成了++。

#### fsync
AOF 日志是以文件的形式存在的，当程序对 AOF 日志文件进行写操作时，++实际上是将内容写到了内核为文件描述符分配的一个内存缓存中，然后内核会异步将脏数据刷回到磁盘的++。

Linux 的 <code>glibc</code>提供了<code>fsync(int fd)</code> 函数可以将指定文件的内容强制从内核缓存刷到磁盘。**只要 Redis 进程实时调用 <code>fsync</code> 函数就可以保证 <code>aof</code> 日志不丢失。但是 <code>fsync</code> 是一个磁盘 IO 操作，它很慢**！如果 Redis 执行一条指令就要 fsync 一次，那么 Redis 高性能的地位就不保了。

所以在生产环境的服务器中，<code>Redis</code> 通常是每隔 1s 左右执行一次 <code>fsync</code> 操作，周期 1s 是可以配置的。这是在数据安全性和性能之间做了一个折中，在保持高性能的同时，尽可能使得数据少丢失。

<code>Redis</code> 同样也提供了另外两种策略，<font color=red>一个是永不 <code>fsync</code>——让操作系统来决定何时同步磁盘，很不安全，另一个是来一个指令就 <code>fsync</code> 一次——非常慢</font>。但是在生产环境基本不会使用，了解一下即可。

> 通常 Redis 的主节点是不会进行持久化操作，持久化操作主要在从节点进行。从节点是备份节点，没有来自客户端请求的压力，它的操作系统资源往往比较充沛。

#### Redis 4.0 混合持久化
> 重启 Redis 时，我们很少使用 rdb 来恢复内存状态，因为会丢失大量数据。我们通常使用 AOF 日志重放，但是重放 AOF 日志性能相对 rdb 来说要慢很多，这样在 Redis 实例很大的情况下，启动需要花费很长的时间。

将 <code>rdb</code> 文件的内容和增量的 <code>AOF</code> 日志文件存在一起。这里的 <code>AOF</code> 日志不再是全量的日志，而是自持久化开始到持久化结束的这段时间发生的增量 <code>AOF</code> 日志，通常这部分 <code>AOF</code> 日志很小。
![image](https://mail.wangkekai.cn/3354A8D7-3098-4448-81D8-ED5BBF6B0CC6.png)

于是在 <code>Redis</code> 重启的时候，<font color=red>可以先加载 <code>rdb</code> 的内容，然后再重放增量 <code>AOF</code> 日志就可以完全替代之前的 <code>AOF</code> 全量文件重放，重启效率因此大幅得到提升。</font>

## 18. 管道
### redis 的消息交互

![image](https://mail.wangkekai.cn/166AC5E1-3578-4428-8CC9-9A8904D404CE.png)

两个连续的写操作和两个连续的读操作总共只会花费一次网络来回，就好比连续的 write 操作合并了，连续的 read 操作也合并了一样。

客户端通过对管道中的指令列表改变读写顺序就可以大幅节省 IO 时间。管道中指令越多，效果越好。

### 深入理解管道本质
![image](https://mail.wangkekai.cn/05CA3F48-4FDB-4A07-9A43-6DAB8D752064.png)
1. 客户端进程调用 <code>write</code> 将消息写到操作系统内核为套接字分配的发送缓冲 <code>send buffer</code>。
1. 客户端操作系统内核将发送缓冲的内容发送到网卡，网卡硬件将数据通过「网际路由」送到服务器的网卡。
1. 服务器操作系统内核将网卡的数据放到内核为套接字分配的接收缓冲 <code>recv buffer</code>。
1. 服务器进程调用 <code>read</code> 从接收缓冲中取出消息进行处理。
1. 服务器进程调用 <code>write</code> 将响应消息写到内核为套接字分配的发送缓冲 <code>send buffer</code>。
1. 服务器操作系统内核将发送缓冲的内容发送到网卡，网卡硬件将数据通过「网际路由」送到客户端的网卡。
1. 客户端操作系统内核将网卡的数据放到内核为套接字分配的接收缓冲 <code>recv buffer</code>。
1. 客户端进程调用 <code>read</code> 从接收缓冲中取出消息返回给上层业务逻辑进行处理。
1. 结束。

## 19. 事物
### 基本使用
<code>multi/exec/discard</code>。<code>multi</code> 指示事务的开始，<code>exec</code> 指示事务的执行，<code>discard</code> 指示事务的丢弃。

```
> multi
OK
> incr books
QUEUED
> incr books
QUEUED
> exec
(integer) 1
(integer) 2
```
所有的指令在 <code>exec</code> 之前不执行，**而是缓存在服务器的一个事务队列中**，服务器一旦收到 <code>exec</code> 指令，才开执行整个事务队列，执行完毕后一次性返回所有指令的运行结果。因为 <code>Redis</code> 的单线程特性，它不用担心自己在执行队列的时候被其它指令打搅，可以保证他们能得到的「原子性」执行。
### 原子性

```
// 特别的例子
> multi
OK
> set books iamastring
QUEUED
> incr books
QUEUED
> set poorman iamdesperate
QUEUED
> exec
1) OK
2) (error) ERR value is not an integer or out of range
3) OK
> get books
"iamastring"
>  get poorman
"iamdesperate
```
上面的例子是事务执行到中间遇到失败了，因为我们不能对一个字符串进行数学运算，事务在遇到指令执行失败后，后面的指令还继续执行，所以 poorman 的值能继续得到设置。

<font color=red>你应该明白 Redis 的**事务根本不能算「原子性」，而仅仅是满足了事务的「隔离性」**，隔离性中的串行化——当前执行的事务有着不被其它事务打断的权利。</font>
### 优化
上面的 <code>Redis</code> 事务在发送每个指令到事务缓存队列时都要经过一次网络读写，当一个事务内部的指令较多时，需要的网络 <code>IO</code> 时间也会线性增长。所以通常 <code>Redis</code> 的客户端在执行事务时都会结合 <code>pipeline</code> 一起使用，这样可以将多次 <code>IO</code> 操作压缩为单次 <code>IO</code> 操作。

```
pipe = redis.pipeline(transaction=true)
pipe.multi()
pipe.incr("books")
pipe.incr("books")
values = pipe.execute()
```
### <code>watch</code>
有多个客户端会并发进行操作。我们可以通过 Redis 的==分布式锁==来避免冲突，这是一个很好的解决方案。**分布式锁是一种悲观锁**，那是不是可以使用乐观锁的方式来解决冲突呢？

<code>Redis</code> 提供了这种 <code>watch</code> 的机制，它就是一种乐观锁。有了 <code>watch</code> 我们又多了一种可以用来解决并发修改的方法。 <code>watch</code> 的使用方式如下：
```
while True:
    do_watch()
    commands()
    multi()
    send_commands()
    try:
        exec()
        break
    except WatchError:
        continue
```
## 20. <code>PubSub</code> - 小道消息
### 消息多播
消息多播允许生产者生产一次消息，中间件负责将消息复制到多个消息队列，每个消息队列由相应的消费组进行消费。它是分布式系统常用的一种解耦方式，用于将多个消费组的逻辑进行拆分。支持了消息多播，多个消费组的逻辑就可以放到不同的子系统中。

### PubSub
为了支持消息多播，Redis 不能再依赖于那 5 种基本数据类型了。**它==单独使用了一个模块来支持消息多播==，这个模块的名字叫着 <code>PubSub</code>，也就是 PublisherSubscriber，发布者订阅者模型**。我们使用 Python 语言来演示一下 PubSub 如何使用。
```
# -*- coding: utf-8 -*-
import time
import redis

client = redis.StrictRedis()
p = client.pubsub()
p.subscribe("codehole")
time.sleep(1)
print p.get_message()
client.publish("codehole", "java comes")
time.sleep(1)
print p.get_message()
client.publish("codehole", "python comes")
time.sleep(1)
print p.get_message()
print p.get_message()
```
客户端发起订阅命令后，<code>Redis</code> 会立即给予一个反馈消息通知订阅成功。因为有网络传输延迟，在 <code>subscribe</code> 命令发出后，需要休眠一会，再通过 <code>get\_message</code> 才能拿到反馈消息。客户端接下来执行发布命令，发布了一条消息。同样因为网络延迟，在 <code>publish</code> 命令发出后，需要休眠一会，再通过 <code>get\_message</code> 才能拿到发布的消息。**如果当前没有消息，<code>get\_message</code> 会返回空，告知当前没有消息，所以它不是阻塞的**。

==<code>Redis PubSub</code> 的生产者和消费者是不同的连接==，也就是上面这个例子实际上使用了两个 <code>Redis</code> 的连接。这是必须的，因为 <code>Redis</code> 不允许连接在 <code>subscribe</code> 等待消息时还要进行其它的操作。
### 模式订阅
简化订阅的繁琐，<code>redis</code> 提供了模式订阅功能 <code>Pattern Subscribe</code>，这样就可以一次订阅多个主题，++即使生产者新增加了同模式的主题，消费者也可以立即收到消息++。
```
> psubscribe codehole.*  # 用模式匹配一次订阅多个主题，主题以 codehole. 字符开头的消息都可以收到
1) "psubscribe"
2) "codehole.*"
3) (integer) 1

```
### PubSub 缺点
PubSub 的生产者传递过来一个消息，Redis 会直接找到相应的消费者传递过去。如果一个消费者都没有，那么消息直接丢弃。如果开始有三个消费者，一个消费者突然挂掉了，生产者会继续发送消息，另外两个消费者可以持续收到消息。但是挂掉的消费者重新连上的时候，这断连期间生产者发送的消息，对于这个消费者来说就是彻底丢失了。

如果 Redis 停机重启，<code>PubSub</code> 的消息是不会持久化的，毕竟 <code>Redis</code> 宕机就相当于一个消费者都没有，所有的消息直接被丢弃。

## 21. 小对象压缩
### 32bit vs 64bit
<code>Redis</code> 如果使用 <code>32bit</code> 进行编译，内部所有数据结构所使用的指针空间占用会少一半，如果你对 <code>Redis</code> 使用内存不超过 <code>4G</code>，可以考虑使用 <code>32bit</code> 进行编译，可以节约大量内存。

### 小对象压缩存储 (<code>ziplist</code>)
++如果 Redis 内部管理的集合数据结构很小，它会使用==紧凑存储形式压缩存储==。++

这就好比 <code>HashMap</code> 本来是二维结构，但是如果内部元素比较少，使用二维结构反而浪费空间，还不如使用一维数组进行存储，需要查找时，因为元素少进行遍历也很快，甚至可以比 <code>HashMap</code> 本身的查找还要快。比如下面我们可以使用数组来模拟 <code>HashMap</code> 的增删改操作。

Redis 的 <code>ziplist</code> 是一个紧凑的字节数组结构，如下图所示，每个元素之间都是紧挨着的。我们不用过于关心 zlbytes/zltail 和 zlend 的含义，稍微了解一下就好。
![image](https://mail.wangkekai.cn/D6AC88EE-8543-4E17-BB49-B258DBF9AF2B.png)

> Redis 的 <code style="color:red">intset</code> 是一个紧凑的整数数组结构，它用于存放元素都是整数的并且元素个数较少的 <code>set</code> 集合。

如果整数可以用 <code>uint16</code> 表示，那么 <code>intset</code> 的元素就是 16 位的数组，如果新加入的整数超过了 <code>uint16</code> 的表示范围，那么就使用 <code>uint32</code> 表示，如果新加入的元素超过了 <code>uint32</code> 的表示范围，那么就使用 <code>uint64</code> 表示，Redis 支持 <code>set</code> 集合动态从 <code>uint16</code> 升级到 <code>uint32</code>，再升级到 <code>uint64</code>。

### 内存回收机制
如果当前 <code>Redis</code> 内存有 <code>10G</code>，当你删除了 <code>1GB</code> 的 <code>key</code> 后，再去观察内存，你会发现内存变化不会太大。**原因是操作系统回收内存是以页为单位，<font color=red>如果这个页上只要有一个 key 还在使用，那么它就不能被回收</font>**。Redis 虽然删除了 <code>1GB</code> 的 <code>key</code>，但是这些 <code>key</code> 分散到了很多页面中，每个页面都还有其它 key 存在，这就导致了内存不会立即被回收。

Redis 虽然无法保证立即回收已经删除的 <code>key</code> 的内存，但是它会重用那些尚未回收的空闲内存。这就好比电影院里虽然人走了，但是座位还在，下一波观众来了，直接坐就行。而操作系统回收内存就好比把座位都给搬走了。
## 22. 主从同步
### CAP 原理
- C - Consistent ，一致性
- A - Availability ，可用性
- P - Partition tolerance ，分区容忍性

分布式系统的节点往往都是分布在不同的机器上进行网络隔离开的，这意味着必然会有网络断开的风险，这个网络断开的场景的专业词汇叫着「**网络分区**」。
> 网络分区发生时，一致性和可用性两难全。

### 最终一致
Redis 的主从数据是异步同步的，所以分布式的 Redis 系统并不满足「一致性」要求。当客户端在 Redis 的主节点修改了数据后，立即返回，即使在主从网络断开的情况下，主节点依旧可以正常对外提供修改服务，所以 Redis 满足「**可用性**」。

Redis 保证「**最终一致性**」，从节点会努力追赶主节点，最终从节点的状态会和主节点的状态将保持一致。如果网络断开了，主从节点的数据将会出现大量不一致，一旦网络恢复，从节点会采用多种策略努力追赶上落后的数据，继续尽力保持和主节点一致。

### 增量同步
<code>Redis</code> 同步的是指令流，主节点会将那些对自己的状态产生修改性影响的指令记录在本地的内存 <code>buffer</code> 中，然后异步将 <code>buffer</code> 中的指令同步到从节点，从节点一边执行同步的指令流来达到和主节点一样的状态，一边向主节点反馈自己同步到哪里了 (偏移量)。

因为内存的 buffer 是有限的，所以 Redis 主库不能将所有的指令都记录在内存 buffer 中。Redis 的复制内存 buffer 是一个定长的环形数组，如果数组内容满了，就会从头开始覆盖前面的内容。

如果因为网络状况不好，从节点在短时间内无法和主节点进行同步，那么当网络状况恢复时，Redis 的主节点中那些没有同步的指令在 buffer 中有可能已经被后续的指令覆盖掉了，从节点将无法直接通过指令流来进行同步，这个时候就需要用到更加复杂的同步机制 —— **快照同步**。

### 快照同步
快照同步是一个非常耗费资源的操作，它首先需要在主库上进行一次 <code>bgsave</code> 将当前内存的数据全部快照到磁盘文件中，然后再将快照文件的内容全部传送到从节点。从节点将快照文件接受完毕后，立即执行一次全量加载，加载之前先要将当前内存的数据清空。加载完毕后通知主节点继续进行增量同步。

在整个快照同步进行的过程中，主节点的复制 <code>buffer</code> 还在不停的往前移动，如果快照同步的时间过长或者复制 <code>buffer</code> 太小，都会导致同步期间的增量指令在复制 <code>buffer</code> 中被覆盖，这样就会导致快照同步完成后无法进行增量复制，然后会再次发起快照同步，如此极有可能会陷入快照同步的死循环。
![image](https://mail.wangkekai.cn/5EAD43EE-6D7B-4AD6-8383-B8F49C43F89A.png)
### 无盘复制
主节点在进行快照同步时，会进行很重的文件 <code>IO</code> 操作，特别是对于非 SSD 磁盘存储时，快照会对系统的负载产生较大影响。特别是当系统正在进行 <code>AOF</code> 的 <code>fsync</code> 操作时如果发生快照，<code>fsync</code> 将会被推迟执行，这就会严重影响主节点的服务效率。

**所谓<font color=red>无盘复制</font>是指主服务器直接通过套接字将快照内容发送到从节点，生成快照是一个遍历的过程，主节点会一边遍历内存，一边将序列化的内容发送到从节点，从节点还是跟之前一样，先将接收到的内容存储到磁盘文件中，再进行一次性加载**。

### Wait 指令
Redis 的复制是异步进行的，<code>wait</code> 指令可以让异步复制变身同步复制，确保系统的强一致性 (不严格)。<code>wait</code> 指令是 <code>Redis3.0</code> 版本以后才出现的。

```
> set key value
OK
> wait 1 0
(integer) 1
```

wait 提供两个参数，第一个参数是从库的数量 N，第二个参数是时间 t，以毫秒为单位。它表示等待 wait 指令之前的所有写操作同步到 N 个从库 (也就是确保 N 个从库的同步没有滞后)，最多等待时间 t。如果时间 t=0，表示无限等待直到 N 个从库同步完成达成一致。

假设此时出现了网络分区，wait 指令第二个参数时间 t=0，主从同步无法继续进行，wait 指令会永远阻塞，Redis 服务器将丧失可用性。

## 23. <code>Sentinel</code> 哨兵
![image](https://mail.wangkekai.cn/34911A9B-E5F0-4694-980D-751714B960D7.png)

它负责**持续监控主从节点的健康，当主节点挂掉时，自动选择一个最优的从节点切换为主节点**。客户端来连接集群时，会首先连接 <code>sentinel</code>，通过 <code>sentinel</code> 来查询主节点的地址，然后再去连接主节点进行数据交互。当主节点发生故障时，客户端会重新向 <code>sentinel</code> 要地址，<code>sentinel</code> 会将最新的主节点地址告诉客户端。如此应用程序将无需重启即可自动完成节点切换。比如上图的主节点挂掉后，集群将可能自动调整为下图所示结构。

### 消息丢失
Redis 主从采用异步复制，意味着当主节点挂掉时，从节点可能没有收到全部的同步消息，这部分未同步的消息就丢失了。如果主从延迟特别大，那么丢失的数据就可能会特别多。<code>sentinel</code> 无法保证消息完全不丢失，但是也尽可能保证消息少丢失。它有两个选项可以限制主从延迟过大。
```
min-slaves-to-write 1
min-slaves-max-lag 10
```
第一个参数表示主节点必须至少有一个从节点在进行正常复制，否则就停止对外写服务，丧失可用性。

第二个参数控制的，它的单位是秒，表示如果 10s 没有收到从节点的反馈，就意味着从节点同步不正常，要么网络断开了，要么一直没有给反馈。

## 24. Cluster
![image](https://mail.wangkekai.cn/F42ED9EA-2F1A-4716-B3BC-D7CC72467CC9.png)

它是去中心化的，如图所示，该集群有三个 Redis 节点组成，每个节点负责整个集群的一部分数据，每个节点负责的数据多少可能不一样。这三个节点相互连接组成一个对等的集群，它们之间通过一种特殊的二进制协议相互交互集群信息。

