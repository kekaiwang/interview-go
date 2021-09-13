# Redis

## 1. string-字符串

字符串结构使用非常广泛，**一个常见的用途就是缓存用户信息**。我们将用户信息结构体使用 JSON 序列化成字符串，然后将序列化后的字符串塞进 Redis 来缓存。同样，取用户信息会经过一次反序列化的过程。

Redis 的字符串是动态字符串，是可以修改的字符串，内部结构实现上类似于 Java 的 `ArrayList`，采用预分配冗余空间的方式来减少内存的频繁分配，如图中所示，*内部为当前字符串实际分配的空间 `capacity` 一般要高于实际字符串长度 len*。  
**当字符串长度小于 1M 时，扩容都是加倍现有的空间，如果超过 1M，扩容时一次只会多扩 1M 的空间。需要注意的是字符串最大长度为 512M**。
> 创建字符串时 len 和 capacity 一样长，不会多分配冗余空间，这是因为绝大多数场景下我们不会使用  append 操作来修改字符串

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

### 过期和set命令扩展

```redis
// 过期
> expire name 5  # 5s 后过期
...  # wait for 5s
> get name
(nil)
> ttl name
-1

> setex name 5 codehole  # 5s 后过期，等价于 set+expire

> setnx name codehole  # 如果 name 不存在就执行 set 创建，如果已经存在，set 创建不成功
(integer) 1 / 0
> get name
"codehole"

// set 指令扩展
> set lock:codehole true ex 5 nx
OK
```

### 过期的 key 集合

redis 会将每个设置了过期时间的 key 放入到一个独立的字典中，以后会定时遍历这个字典来删除到期的 key。  
除了**定时遍历之外，它还会使用惰性策略来删除过期的 key**。
> 所谓惰性策略就是在客户端访问这个  key 的时候， redis 对  key 的过期时间进行检查，如果过期了就立即删除。

**定时删除是集中处理，惰性删除是零散处理**。

### 定时扫描策略

Redis 默认会每秒进行十次过期扫描，过期扫描不会遍历过期字典中所有的 key，而是采用了一种简单的**贪心策略**。

1. 从过期字典中随机 20 个 key
2. 删除这 20 个 key 中已经过期的 key
3. 如果过期的 key 比率超过 1/4，那就重复步骤 1

> 同时，为了保证过期扫描不会出现循环过度，导致线程卡死现象，算法还增加了扫描时间的上限，默认不会超过 25ms。

*设想一个大型的 Redis 实例中所有的 key 在同一时间过期了，会出现怎样的结果？*

> 毫无疑问，Redis 会持续扫描过期字典 (循环多次)，直到过期字典中过期的 key 变得稀疏，才会停止 (循环次数明显下降)。这就会导致线上读写请求出现明显的卡顿现象。导致这种卡顿的另外一种原因是内存管理器需要频繁回收内存页，这也会产生一定的 CPU 消耗。

当客户端请求到来时，服务器如果正好进入过期扫描状态，客户端的请求将会等待至少 25ms 后才会进行处理，如果客户端将超时时间设置的比较短，比如 10ms，那么就会出现大量的链接因为超时而关闭，业务端就会出现很多异常。而且这时你还无法从 Redis 的 slowlog 中看到慢查询记录，因为慢查询指的是逻辑处理过程慢，不包含等待时间。

所以业务开发人员一定要注意过期时间，**如果有大批量的 key 过期，要给过期时间设置一个随机范围，而不宜全部在同一时间过期，分散过期处理的压力**。

```redis
redis.expire_at(key, random.randint(86400) + expire_ts)
```

### 从库的过期策略

从库不会进行过期扫描，从库对过期的处理是被动的。主库在 key 到期时，会在 AOF 文件里增加一条 `del` 指令，同步到所有的从库，从库通过执行这条  del 指令来删除过期的 key。

> **字符串是由多个字节组成，每个字节又是由 8 个 bit 组成**，如此便可以将一个字符串看成很多 bit 的组合，这便是 bitmap「位图」数据结构

### string内部结构

**Redis 中的字符串是可以修改的字符串，在内存中它是以字节数组的形式存在的**。  
我们知道 C 语言里面的字符串标准形式是以 `NULL` 作为结束符，但是在 Redis 里面字符串不是这么表示的。因为要获取 `NULL` 结尾的字符串的长度使用的是 strlen 标准库函数，这个函数的算法复杂度是 O(n)，它需要对字节数组进行遍历扫描，作为单线程的 Redis 表示承受不起。

Redis 的字符串叫着「SDS」，也就是 `Simple Dynamic String`。它的结构是一个带长度信息的字节数组。

```c++
struct SDS<T> {
    T capacity; // 数组容量 - 1byte
    T len; // 数组长度 - 1byte
    byte flags; // 特殊标识位，不理睬它 - 1byte
    byte[] content; // 数组内容
}
```

![image](https://mail.wangkekai.cn/D1DA79F5-652B-4DD4-B665-A4A4616176D6.png)

capacity **表示所分配数组的长度**，len **表示字符串的实际长度**。前面我们提到字符串是可以修改的字符串，它要支持 append 操作。如果数组没有冗余空间，那么追加操作必然涉及到分配新数组，然后将旧内容复制过来，再 append 新内容。如果字符串的长度非常长，这样的内存分配和复制开销就会非常大。

> 上面的 SDS 结构使用了范型 T，为什么不直接用 int 呢，这是因为当字符串比较短时，len 和 capacity 可以使用 byte 和 short 来表示，Redis 为了对内存做极致的优化，不同长度的字符串使用不同的结构体来表示。

### embstr vs raw

Redis 的字符串有两种存储方式，在长度特别短时，使用 emb 形式存储 (embeded)，当长度超过 44 时，使用  raw 形式存储。

所有的 Redis 对象都有下面的这个结构头:

```c++
struct RedisObject {
    int4 type; // 4bits
    int4 encoding; // 4bits
    int24 lru; // 24bits
    int32 refcount; // 4bytes
    void *ptr; // 8bytes，64-bit system
} robj;
```

接着我们再看 SDS 结构体的大小，在字符串比较小时，SDS 对象头的大小是 capacity+3，至少是 3。意味着分配一个字符串的最小空间占用为 19 字节 (16+3)。

```c++
struct SDS {
    int8 capacity; // 1byte
    int8 len; // 1byte
    int8 flags; // 1byte
    byte[] content; // 内联数组，长度为 capacity
}
```

![image](https://mail.wangkekai.cn/9528D65C-6E40-4157-9797-280B11D23E1A.png)

embstr 存储形式是这样一种存储形式，它将 RedisObject 对象头和 SDS 对象连续存在一起，使用 malloc 方法一次分配。而 raw 存储形式不一样，它需要两次 malloc，两个对象头在内存地址上一般是不连续的。

而内存分配器 jemalloc/tcmalloc 等分配内存大小的单位都是 2、4、8、16、32、64 等等，为了能容纳一个完整的 embstr 对象，jemalloc 最少会分配 32 字节的空间，如果字符串再稍微长一点，那就是 64 字节的空间。如果总体超出了 64 字节，Redis 认为它是一个大字符串，不再使用 emdstr 形式存储，而该用 raw 形式。

前面我们提到 SDS 结构体中的 content 中的字符串是以字节\0结尾的字符串，之所以多出这样一个字节，是为了便于直接使用 glibc 的字符串处理函数，以及为了便于字符串的调试打印输出。

![image](https://mail.wangkekai.cn/349F73F3-2726-45AD-8AFB-A10AC276BBB4.png)

看上面这张图可以算出，留给 content 的长度最多只有 45(64-19) 字节了。字符串又是以\0结尾，所以 embstr 最大能容纳的字符串长度就是 44。

### 扩容策略

字符串在长度小于 1M 之前，扩容空间采用加倍策略，也就是保留 100% 的冗余空间。当长度超过 1M 之后，为了避免加倍后的冗余空间过大而导致浪费，每次扩容只会多分配 1M 大小的冗余空间。

## 2. list - 列表

Redis 的列表相当于 Java 语言里面的 LinkedList，*注意它是链表而不是数组*。  
**这意味着 list 的插入和删除操作非常快，时间复杂度为 O(1)，但是索引定位很慢，时间复杂度为 O(n)，这点让人非常意外**。

当列表弹出了最后一个元素之后，该数据结构自动被删除，内存被回收。  
**Redis 的列表结构常用来做异步队列使用**。将需要延后处理的任务结构体序列化成字符串塞进 Redis 的列表，另一个线程从这个列表中轮询数据进行处理。

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

index 相当于 Java 链表的 get(int index)方法，它需要对链表进行遍历，性能随着参数 index 增大而变差。

trim 和字面上的含义不太一样，个人觉得它叫 lretain(保留) 更合适一些，因为 ltrim 跟的两个参数 start_index 和 end_index 定义了一个区间，在这个区间内的值，ltrim 要保留，区间之外统统砍掉。我们可以通过 ltrim 来实现一个定长的链表，这一点非常有用。

index 可以为负数，index=-1 表示倒数第一个元素，同样 index=-2 表示倒数第二个元素。

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

### 快速列表

如果再深入一点，你会发现 Redis 底层存储的还不是一个简单的 linkedlist，而是称之为快速链表 `quicklist` 的一个结构。

*首先在列表元素较少的情况下会使用一块连续的内存存储*，这个结构是 ziplist，也就是**压缩列表**。  
**它将所有的元素紧挨着一起存储，*分配的是一块连续的内存***。当数据量比较多的时候才会改成 quicklist。因为普通的链表需要的附加指针空间太大，会比较浪费空间，而且会加重内存的碎片化。比如这个列表里存的只是 int 类型的数据，结构上还需要两个额外的指针 prev 和 next 。所以 Redis 将链表和 ziplist 结合起来组成了 quicklist。也就是将多个 ziplist 使用双向指针串起来使用。这样既满足了快速的插入删除性能，又不会出现太大的空间冗余。

Redis 早期版本存储 list 列表数据结构使用的是压缩列表 ziplist 和普通的双向链表 linkedlist，也就是元素少时用 ziplist，元素多时用 linkedlist。

```c++
// 链表的节点
struct listNode<T> {
    listNode* prev;
    listNode* next;
    T value;
}
// 链表
struct list {
    listNode *head;
    listNode *tail;
    long length;
}
```

考虑到链表的附加空间相对太高，prev 和 next 指针就要占去 16 个字节 (64bit 系统的指针是 8 个字节)，另外每个节点的内存都是单独分配，会加剧内存的碎片化，影响内存管理效率。后续版本对列表数据结构进行了改造，使用 quicklist 代替了 ziplist 和 linkedlist。

```redis
> rpush codehole go java python
(integer) 3
> debug object codehole
Value at:0x7fec2dc2bde0 refcount:1 encoding:quicklist serializedlength:31 lru:6101643 lru_seconds_idle:5 ql_nodes:1 ql_avg_node:3.00 ql_ziplist_max:-2 ql_compressed:0 ql_uncompressed_size:29
```

**注意观察上面输出字段 encoding 的值。 quicklist 是 ziplist 和 linkedlist 的混合体，它将  linkedlist 按段切分，每一段使用 ziplist 来紧凑存储，多个 ziplist 之间使用双向指针串接起来。**
![image](https://mail.wangkekai.cn/7AAB35D9-BBDB-46C0-879F-04746234203D.png)

### 压缩列表

Redis 为了节约内存空间使用，zset 和 hash 容器对象在元素个数较少的时候，采用压缩列表 (ziplist) 进行存储。  
**压缩列表是一块连续的内存空间，元素之间紧挨着存储，没有任何冗余空隙。**

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

```c++
struct ziplist<T> {
    int32 zlbytes; // 整个压缩列表占用字节数
    int32 zltail_offset; // 最后一个元素距离压缩列表起始位置的偏移量，用于快速定位到最后一个节点
    int16 zllength; // 元素个数
    T[] entries; // 元素内容列表，挨个挨个紧凑存储
    int8 zlend; // 标志压缩列表的结束，值恒为 0xFF
}
```

> 压缩列表为了支持双向遍历，所以才会有 ztail_offset 这个字段，用来快速定位到最后一个元素，然后倒着遍历。

### 增加元素

因为 ziplist 都是紧凑存储，没有冗余空间 (对比一下 Redis 的字符串结构)。意味着插入一个新的元素就需要调用 realloc 扩展内存。取决于内存分配器算法和当前的 ziplist 内存大小，realloc 可能会重新分配新的内存空间，并将之前的内容一次性拷贝到新的地址，也可能在原有的地址上进行扩展，这时就不需要进行旧内容的内存拷贝。

如果 ziplist 占据内存太大，重新分配内存和拷贝内存就会有很大的消耗。所以 **ziplist 不适合存储大型字符串，存储的元素也不宜过多**。

### IntSet 小整数集合

当 set 集合容纳的元素都是整数并且元素个数较小时，Redis 会使用 intset 来存储结合元素。intset 是紧凑的数组结构，同时支持 16 位、32 位和 64 位整数。

```c++
struct intset<T> {
    int32 encoding; // 决定整数位宽是 16 位、32 位还是 64 位
    int32 length; // 元素个数
    int<T> contents; // 整数数组，可以是 16 位、32 位和 64 位
}
```

### 每个 ziplist 存多少元素？

`quicklist` 内部默认单个 ziplist 长度为 8k 字节，超出了这个字节数，就会新起一个 ziplist。ziplist 的长度由配置参数 list-max-ziplist-size 决定。

### 压缩深度

![image](https://mail.wangkekai.cn/71ABE178-6788-4362-90F7-EB1F74495E9E.png)
quicklist 默认的压缩深度是 0，也就是不压缩。压缩的实际深度由配置参数 list-compress-depth 决定。为了支持快速的 push/pop 操作，quicklist 的首尾两个 ziplist 不压缩，此时深度就是 1。如果深度为 2，就表示 quicklist 的首尾第一个 ziplist 以及首尾第二个 ziplist 都不压缩。

## 3. hash - 字典

Redis 的字典相当于 Java 语言里面的 HashMap，它是**无序字典**。内部实现结构上同 Java 的 HashMap 也是一致的，**同样的数组 + 链表二维结构。第一维 hash 的数组位置碰撞时，就会将碰撞的元素使用链表串接起来**。
![image](https://mail.wangkekai.cn/999C4B1B-8CD5-42E5-BF26-74022C9B2326.png)

Redis 的字典的值只能是字符串，另外它们 rehash 的方式不一样，因为 Java 的 HashMap 在字典很大时，rehash 是个耗时的操作，需要一次性全部 rehash。**Redis 为了高性能，不能堵塞服务，所以采用了渐进式 rehash 策略。**
![image](https://mail.wangkekai.cn/C5E5EF39-9A78-4D29-A1F7-C3983C4F4A58.png)

渐进式 rehash 会在 rehash 的同时，保留新旧两个 hash 结构，查询时会同时查询两个 hash 结构，然后在后续的定时任务中以及 hash 操作指令中，循序渐进地将旧 hash 的内容一点点迁移到新的 hash 结构中。当搬迁完成了，就会使用新的 hash 结构取而代之。

> 当 hash 移除了最后一个元素之后，该数据结构自动被删除，内存被回收。

hash 结构也可以用来存储用户信息，不同于字符串一次性需要全部序列化整个对象，hash 可以对用户结构中的每个字段单独存储。这样当我们需要获取用户信息时可以进行部分获取。而以整个字符串的形式去保存用户信息的话就只能一次性全部读取，这样就会比较浪费网络流量。

> hash 也有缺点，hash 结构的存储消耗要高于单个字符串，到底该使用 hash 还是字符串，需要根据实际情况再三权衡。

```redis
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

### dict - 内部结构

dict 是 Redis 服务器中出现最为频繁的复合型数据结构，除了 hash 结构的数据会用到字典外，**整个 Redis 数据库的所有 key 和 value 也组成了一个全局字典**，还有带过期时间的 key 集合也是一个字典。  
**zset 集合中存储 value 和 score 值的映射关系也是通过 dict 结构实现的。**

```c++
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

`dict` 结构内部包含两个 hashtable，通常情况下只有一个 hashtable 是有值的。但是在 dict 扩容缩容时，需要分配新的 hashtable，然后进行渐进式搬迁，这时候两个 hashtable 存储的分别是旧的 hashtable 和新的 hashtable。待搬迁结束后，旧的 hashtable 被删除，新的 hashtable 取而代之。

### 渐进式 rehash

大字典的扩容是比较耗时间的，需要重新申请新的数组，然后将旧字典所有链表中的元素重新挂接到新的数组下面，这是一个 O(n) 级别的操作，作为单线程的Redis表示很难承受这样耗时的过程。步子迈大了会扯着蛋，所以 Redis 使用渐进式 rehash 小步搬迁。虽然慢一点，但是肯定可以搬完。

### 查找的过程

插入和删除操作都依赖于查找，先必须把元素找到，才可以进行数据结构的修改操作。hashtable 的元素是在第二维的链表上，所以首先我们得想办法定位出元素在哪个链表上。

```c++
func get(key) {
    let index = hash_func(key) % size;
    let entry = table[index];
    while(entry != NULL) {
        if entry.key == target {
            return entry.value;
        }
        entry = entry.next;
    }
}
```

值得注意的是代码中的 hash_func，它会将 key 映射为一个整数，不同的 key 会被映射成分布比较均匀散乱的整数。只有 hash 值均匀了，整个 hashtable 才是平衡的，所有的二维链表的长度就不会差距很远，查找算法的性能也就比较稳定。

### hash 函数

hashtable 的性能好不好完全取决于 hash 函数的质量。hash 函数如果可以将 key 打散的比较均匀，那么这个 hash 函数就是个好函数。Redis 的字典默认的 hash 函数是 siphash。siphash 算法即使在输入 key 很小的情况下，也可以产生随机性特别好的输出，而且它的性能也非常突出。对于 Redis 这样的单线程来说，字典数据结构如此普遍，字典操作也会非常频繁，hash 函数自然也是越快越好。

### 扩容条件

**正常情况下，当 hash 表中元素的个数等于第一维数组的长度时，就会开始扩容，扩容的新数组是原数组大小的 2 倍**。不过如果 Redis 正在做 bgsave，为了减少内存页的过多分离 (Copy On Write)，Redis 尽量不去扩容 (dict_can_resize)，但是如果 hash 表已经非常满了，元素的个数已经达到了第一维数组长度的 5 倍  (dict_force_resize_ratio)，说明 hash 表已经过于拥挤了，这个时候就会强制扩容。

### 缩容条件

当 hash 表因为元素的逐渐删除变得越来越稀疏时，Redis 会对 hash 表进行缩容来减少 hash 表的第一维数组空间占用。  
**缩容的条件是元素个数低于数组长度的 10%。缩容不会考虑 Redis 是否正在做 bgsave。**

### set 的结构

Redis 里面 set 的结构底层实现也是字典，只不过所有的 value 都是 NULL，其它的特性和字典一模一样。

## 4.  set - 集合

Redis 的集合相当于 Java 语言里面的 HashSet，**它内部的键值对是无序的唯一的**。它的内部实现相当于一个特殊的字典，字典中所有的 value 都是一个值 NULL。
> 当集合中最后一个元素移除之后，数据结构自动删除，内存被回收。

set 结构可以用来存储活动中奖的用户 ID，因为有去重功能，可以保证同一个用户不会中奖两次。

```redis
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

## 5. zset - 有序集合

它类似于 Java 的 SortedSet 和 HashMap 的结合体，一方面它是一个 set，保证了内部 value 的唯一性，另一方面它可以给每个 value 赋予一个 score，代表这个 value 的排序权重。它的内部实现用的是一种叫做「跳跃列表」的数据结构。

zset 可以用来存粉丝列表或学生的成绩，value 值是粉丝的用户/学生 ID，score 是关注时间/成绩。我们可以对粉丝列表按关注时间/成绩进行排序。

```redis
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

### 跳跃列表

zset 内部的排序功能是通过「**跳跃列表**」数据结构来实现的，它的结构非常特殊，也比较复杂。

因为 zset 要支持随机的插入和删除，所以它不好使用数组来表示。我们先看一个普通的链表结构。

![image](https://mail.wangkekai.cn/230C465E-1643-4B51-ACC7-C16D772F8512.png)

我们需要这个链表按照 score 值进行排序。这意味着当有新元素需要插入时，要定位到特定位置的插入点，这样才可以继续保证链表是有序的。通常我们会通过二分查找来找到插入点，但是二分查找的对象必须是数组，只有数组才可以支持快速位置定位，链表做不到，那该怎么办？

> 跳跃列表就是类似于这种层级制，最下面一层所有的元素都会串起来。然后每隔几个元素挑选出一个代表来，再将这几个代表使用另外一级指针串起来。然后在这些代表里再挑出二级代表，再串起来。最终就形成了金字塔结构。

![image](https://mail.wangkekai.cn/57E68050-B9D5-476E-BF90-B2FA661018F4.png)

「**跳跃列表**」之所以「跳跃」，是因为内部的元素可能「身兼数职」，比如上图中间的这个元素，同时处于 L0、L1 和 L2 层，可以快速在不同层次之间进行「跳跃」。

定位插入点时，先在顶层进行定位，然后下潜到下一级定位，一直下潜到最底层找到合适的位置，将新元素插进去。你也许会问，那新插入的元素如何才有机会「身兼数职」呢？

跳跃列表采取一个随机策略来决定新元素可以兼职到第几层。

首先 L0 层肯定是 100% 了，L1 层只有 50% 的概率，L2 层只有 25% 的概率，L3 层只有 12.5% 的概率，一直随机到最顶层 L31 层。绝大多数元素都过不了几层，只有极少数元素可以深入到顶层。列表中的元素越多，能够深入的层次就越深，能进入到顶层的概率就会越大。

### 查找过程

设想如果跳跃列表只有一层会怎样？插入删除操作需要定位到相应的位置节点 (定位到最后一个比「我」小的元素，也就是第一个比「我」大的元素的前一个)，定位的效率肯定比较差，复杂度将会是 O(n)，因为需要挨个遍历。也许你会想到二分查找，但是二分查找的结构只能是有序数组。跳跃列表有了多层结构之后，这个定位的算法复杂度将会降到 O(lg(n))。
![image](https://mail.wangkekai.cn/4272EE6D-0DAF-4B09-8503-D0595D21CED7.png)

我们要定位到那个紫色的 kv，需要从 header 的最高层开始遍历找到第一个节点 (最后一个比「我」小的元素)，然后从这个节点开始降一层再遍历找到第二个节点 (最后一个比「我」小的元素)，然后一直降到最底层进行遍历就找到了期望的节点 (最底层的最后一个比我「小」的元素)。

我们将中间经过的一系列节点称之为「**搜索路径**」，它是从最高层一直到最底层的每一层最后一个比「我」小的元素节点列表。

有了这个搜索路径，我们就可以插入这个新节点了。不过这个插入过程也不是特别简单。因为新插入的节点到底有多少层，得有个算法来分配一下，跳跃列表使用的是随机算法。

### 随机层数

对于每一个新插入的节点，都需要调用一个随机算法给它分配一个合理的层数。直观上期望的目标是 50% 的 Level1，25% 的 Level2，12.5% 的 Level3，一直到最顶层2^-63，因为这里每一层的晋升概率是 50%。

不过 Redis 标准源码中的晋升概率只有 25%，也就是代码中的 ZSKIPLIST_P 的值。所以官方的跳跃列表更加的扁平化，层高相对较低，在单个层上需要遍历的节点数量会稍多一点。

也正是因为层数一般不高，所以遍历的时候从顶层开始往下遍历会非常浪费。跳跃列表会记录一下当前的最高层数maxLevel，遍历时从这个 maxLevel 开始遍历性能就会提高很多。

### 元素排名是怎么算出来的？

Redis 在 skiplist 的 forward 指针上进行了优化，给每一个 forward 指针都增加了 span 属性，span 是「跨度」的意思，表示从当前层的当前节点沿着 forward 指针跳到下一个节点中间跳过多少个节点。Redis 在插入删除操作时会小心翼翼地更新 span 值的大小。

```c++
struct zslforward {
  zslnode* item;
  long span;  // 跨度
}

struct zslnode {
  String value;
  double score;
  zslforward*[] forwards;  // 多层连接指针
  zslnode* backward;  // 回溯指针
}
```

这样当我们要计算一个元素的排名时，只需要将「搜索路径」上的经过的所有节点的跨度 span 值进行叠加就可以算出元素的最终 rank 值。

## 6. 分布式锁

分布式锁本质上要实现的目标就是在 Redis 里面占一个“茅坑”，当别的进程也要来占时，发现已经有人蹲在那里了，就只好放弃或者稍后再试。

占坑一般是使用 `setnx(set if not exists)` 指令，只允许被一个客户端占坑。
如果在 setnx 和 expire 之间服务器进程突然挂掉了，可能是因为机器掉电或者是被人为杀掉的，就会导致 expire 得不到执行，也会造成死锁。

> Redis 2.8 版本中作者加入了 set 指令的扩展参数，使得 `setnx` 和 `expire` 指令可以一起执行，彻底解决了分布式锁的乱象。

```redis
> setnx lock:codehole true // 这里的冒号:就是一个普通的字符，没特别含义，它可以是任意其它字符，不要误解
OK
... do something critical ...
> del lock:codehole
(integer) 1
```

比如在 Sentinel 集群中，主节点挂掉时，从节点会取而代之，客户端上却并没有明显感知。原先第一个客户端在主节点中申请成功了一把锁，但是这把锁还没有来得及同步到从节点，主节点突然挂掉了。然后从节点变成了主节点，这个新的节点内部没有这个锁，所以当另一个客户端过来请求加锁时，立即就批准了。这样就会导致系统中同样一把锁被两个客户端同时持有，不安全性由此产生。

### Redlock 算法

为了使用 Redlock，需要提供多个 Redis 实例，这些实例之前相互独立没有主从关系。同很多分布式算法一样，redlock 也使用 「*大多数机制*」。

**加锁时**，它会向过半节点发送 `set(key, value, nx=True, ex=xxx)` 指令，*只要过半节点 set 成功，那就认为加锁成功*。释放锁时，需要向所有节点发送 del 指令。不过 Redlock 算法还需要考虑出错重试、时钟漂移等很多细节问题，同时因为 Redlock 需要向多个节点进行读写，意味着相比单实例 Redis 性能会下降一些。

### Redlock 使用场景

在乎高可用性，希望挂了一台 redis 完全不受影响，那就应该考虑 redlock。不过代价也是有的，需要更多的 redis 实例，性能也下降了，代码上还需要引入额外的 library，运维上也需要特殊对待，这些都是需要考虑的成本。

## 7. 延时队列

常用 Rabbitmq 和 Kafka 作为消息队列中间件进行异步消息传递。

### 异步消息队列

Redis 的 list(列表) 数据结构常用来作为异步消息队列使用，使用 rpush/lpush 操作入队列，使用 lpop 和 rpop 来出队列。
![image](https://mail.wangkekai.cn/1607610238093.jpg)

```redis
> rpush notify-queue apple banana pear
(integer) 3
> llen notify-queue
(integer) 3
> lpop notify-queue
"apple"
> llen notify-queue
(integer) 2
...
> lpop notify-queue
"pear"
> llen notify-queue
(integer) 0
> lpop notify-queue
(nil)
```

### 队列延迟

可是如果队列空了，客户端就会陷入 pop 的死循环，不停地 pop，没有数据，接着再 pop，又没有数据。这就是浪费生命的空轮询。空轮询不但拉高了客户端的 CPU，redis 的 QPS 也会被拉高，如果这样空轮询的客户端有几十来个，Redis 的慢查询可能会显著增多。

*阻塞读在队列没有数据的时候，会立即进入休眠状态，一旦数据到来，则立刻醒过来。消息的延迟几乎为零*。用 `blpop/brpop` 替代前面的 lpop/rpop，就完美解决了上面的问题。

### 空闲连接自动断开

如果线程一直阻塞在哪里，Redis 的客户端连接就成了闲置连接，闲置过久，服务器一般会主动断开连接，减少闲置资源占用。这个时候 blpop/brpop 会抛出异常来。 **注意捕获异常，还要重试。**

*延时队列可以通过 Redis 的 zset(有序列表) 来实现*。  
**我们将消息序列化成一个字符串作为 zset 的 value，这个消息的到期处理时间作为 score，然后用多个线程轮询 zset 获取到期的任务进行处理，多个线程是为了保障可用性，万一挂了一个线程还有其它线程可以继续处理**。  
因为有多个线程，所以需要考虑并发争抢任务，确保任务不能被多次执行。

```python
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

Redis 的 zrem 方法是多线程多进程争抢任务的关键，它的返回值决定了当前实例有没有抢到任务，因为 loop 方法可能会被多个线程、多个进程调用，同一个任务可能会被多个进程线程抢到，通过 zrem 来决定唯一的属主。

## 8. 位图

Redis 提供了位图数据结构，这样每天的签到记录只占据一个位，365 天就是 365 个位，46 个字节 (一个稍长一点的字符串) 就可以完全容纳下，这就大大节约了存储空间。

*位图不是特殊的数据结构，它的内容其实就是普通的字符串，也就是 byte 数组*。我们可以使用普通的 get/set 直接获取和设置整个位图的内容，也可以使用位图操作 `getbit/setbit` 等将 byte 数组看成「位数组」来处理。

### 统计和查找

Redis 提供了位图统计指令 `bitcount` 和位图查找指令 `bitpos`，bitcount 用来统计指定位置范围内 1 的个数，kbitpos 用来查找指定范围内出现的第一个 0 或 1。

## 9. HyperLogLog

Redis 提供了 HyperLogLog 数据结构就是用来解决这种统计问题的。HyperLogLog 提供不精确的去重计数方案，虽然不精确但是也不是非常不精确，标准误差是 0.81%，这样的精确度已经可以满足上面的 UV 统计需求了。

HyperLogLog 数据结构是 Redis 的高级数据结构，它非常有用，但是令人感到意外的是，使用过它的人非常少。

### 使用方法

HyperLogLog 提供了两个指令 `pfadd` 和 `pfcount`，根据字面意义很好理解，**一个是增加计数，一个是获取计数**。

```redis
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

### pfmerge 适合什么场合用？

HyperLogLog 除了上面的 pfadd 和 pfcount 之外，还提供了第三个指令 `pfmerge`，用于将多个 pf 计数值累加在一起形成一个新的 pf 值。

### 注意事项

HyperLogLog 它需要占据一定 12k 的存储空间，所以它不适合统计单个用户相关的数据。如果你的用户上亿，可以算算，这个空间成本是非常惊人的。但是相比 set 存储方案，HyperLogLog 所使用的空间那真是可以使用千斤对比四两来形容了。

不过你也不必过于担心，因为 Redis 对 HyperLogLog 的存储进行了优化，***在计数比较小时，它的存储空间采用稀疏矩阵存储，空间占用很小*，仅仅在计数慢慢变大，稀疏矩阵占用空间渐渐超过了阈值时才会一次性转变成稠密矩阵，才会占用 12k 的空间**。

### 见缝插针 —— 探索 HyperLogLog 内部

HyperLogLog算法是一种非常巧妙的近似统计海量去重元素数量的算法。它内部维护了 16384 个桶（bucket）来记录各自桶的元素数量。当一个元素到来时，它会散列到其中一个桶，以一定的概率影响这个桶的计数值。因为是概率算法，所以单个桶的计数值并不准确，但是将所有的桶计数值进行调合均值累加起来，结果就会非常接近真实的计数值。

![image](https://mail.wangkekai.cn/2FE9CC12-3E6B-491F-9240-BC26DCD6ACD9.png)

为了便于理解 HyperLogLog 算法，我们先简化它的计数逻辑。因为是去重计数，如果是准确的去重，肯定需要用到 set 集合，使用集合来记录所有的元素，然后使用 scard 指令来获取集合大小就可以得到总的计数。因为元素特别多，单个集合会特别大，所以将集合打散成 16384 个小集合。当元素到来时，通过 hash 算法将这个元素分派到其中的一个小集合存储，同样的元素总是会散列到同样的小集合。这样总的计数就是所有小集合大小的总和。使用这种方式精确计数除了可以增加元素外，还可以减少元素。

```python
# coding:utf-8
import hashlib

class ExactlyCounter:

    def __init__(self):
        # 先分配16384个空集合
        self.buckets = []
        for i in range(16384):
            self.buckets.append(set([]))
        # 使用md5哈希算法
        self.hash = lambda x: int(hashlib.md5(x).hexdigest(), 16)
        self.count = 0

    def add(self, element):
        h = self.hash(element)
        idx = h % len(self.buckets)
        bucket = self.buckets[idx]
        old_len = len(bucket)
        bucket.add(element)
        if len(bucket) > old_len:
            # 如果数量变化了，总数就+1
            self.count += 1

    def remove(self, element):
        h = self.hash(element)
        idx = h % len(self.buckets)
        bucket = self.buckets[idx]
        old_len = len(bucket)
        bucket.remove(element)
        if len(bucket) < old_len:
            # 如果数量变化了，总数-1
            self.count -= 1


if __name__ == '__main__':
    c = ExactlyCounter()
    for i in range(100000):
        c.add("element_%d" % i)
    print c.count
    for i in range(100000):
        c.remove("element_%d" % i)
    print c.count
```

集合打散并没有什么明显好处，因为总的内存占用并没有减少。HyperLogLog 肯定不是这个算法，它需要对这个小集合进行优化，压缩它的存储空间，让它的内存变得非常微小。HyperLogLog 算法中每个桶所占用的空间实际上只有 6 个 bit，这 6 个 bit 自然是无法容纳桶中所有元素的，它记录的是桶中元素数量的对数值。

为了说明这个对数值具体是个什么东西，我们先来考虑一个小问题。一个随机的整数值，这个整数的尾部有一个 0 的概率是 50%，要么是 0 要么是 1。同样，尾部有两个 0 的概率是 25%，有三个零的概率是 12.5%，以此类推，有 k 个 0 的概率是 2^(-k)。如果我们随机出了很多整数，整数的数量我们并不知道，但是我们记录了整数尾部连续 0 的最大数量 K。我们就可以通过这个 K 来近似推断出整数的数量，这个数量就是 2^K。

当然结果是非常不准确的，因为可能接下来你随机了非常多的整数，但是末尾连续零的最大数量 K 没有变化，但是估计值还是 2^K。你也许会想到要是这个 K 是个浮点数就好了，每次随机一个新元素，它都可以稍微往上涨一点点，那么估计值应该会准确很多。

HyperLogLog 通过分配 16384 个桶，然后对所有的桶的最大数量 K 进行调合平均来得到一个平均的末尾零最大数量 K# ，K# 是一个浮点数，使用平均后的 2^K# 来估计元素的总量相对而言就会准确很多。不过这只是简化算法，真实的算法还有很多修正因子，因为涉及到的数学理论知识过于繁多，这里就不再精确描述。

### 密集存储结构

不论是稀疏存储还是密集存储，Redis 内部都是使用字符串位图来存储 HyperLogLog 所有桶的计数值。密集存储的结构非常简单，就是连续 16384 个 6bit 串成的字符串位图。

![image](https://mail.wangkekai.cn/7C44C48F-0FA2-429B-AB7A-9FB63B6B788F.png)

假设桶的编号为idx，这个 6bit 计数值的起始字节位置偏移用 offset_bytes 表示，它在这个字节的起始比特位置偏移用 offset_bits 表示。我们有

```redis
offset_bytes = (idx * 6) / 8
offset_bits = (idx * 6) % 8
```

前者是商，后者是余数。比如 bucket 2 的字节偏移是 1，也就是第 2 个字节。它的位偏移是4，也就是第 2 个字节的第 5 个位开始是 bucket 2 的计数值。需要注意的是字节位序是左边低位右边高位，而通常我们使用的字节都是左边高位右边低位，我们需要在脑海中进行倒置。

![image](https://mail.wangkekai.cn/FA64324C-3A44-43CC-8816-644949272566.png)

如果 offset_bits 小于等于 2，那么这 6bit 在一个字节内部，可以直接使用下面的表达式得到计数值 val

![image](https://mail.wangkekai.cn/599BB0E3-31E2-49F7-A303-C00EF8C96556.png)

```shell
val = buffer[offset_bytes] >> offset_bits  # 向右移位
```

如果 offset_bits 大于 2，那么就会跨越字节边界，这时需要拼接两个字节的位片段。

![image](https://mail.wangkekai.cn/627C3EC9-4BE3-4EC1-9F13-61433C3EFEB1.png)

```shell
# 低位值
low_val = buffer[offset_bytes] >> offset_bits
# 低位个数
low_bits = 8 - offset_bits
# 拼接，保留低6位
val = (high_val << low_bits | low_val) & 0b111111
```

不过下面 Redis 的源码要晦涩一点，看形式它似乎只考虑了跨越字节边界的情况。这是因为如果 6bit 在单个字节内，上面代码中的 high_val 的值是零，所以这一份代码可以同时照顾单字节和双字节。

### 稀疏存储结构

稀疏存储适用于很多计数值都是零的情况。下图表示了一般稀疏存储计数值的状态。

![image](https://user-gold-cdn.xitu.io/2018/8/31/1658e631c1730628?imageView2/0/w/1280/h/960/ignore-error/1)

**当多个连续桶的计数值都是零时，Redis 使用了一个字节来表示接下来有多少个桶的计数值都是零：00xxxxxx。**前缀两个零表示接下来的 6bit 整数值加 1 就是零值计数器的数量，注意这里要加 1 是因为数量如果为零是没有意义的。比如 00010101 表示连续 22 个零值计数器。6bit 最多只能表示连续 64 个零值计数器，所以 Redis 又设计了连续多个多于 64 个的连续零值计数器，它使用两个字节来表示：01xxxxxx yyyyyyyy，后面的 14bit 可以表示最多连续 16384 个零值计数器。这意味着 HyperLogLog 数据结构中 16384 个桶的初始状态，所有的计数器都是零值，可以直接使用 2 个字节来表示。

如果连续几个桶的计数值非零，那就使用形如 1vvvvvxx 这样的一个字节来表示。中间 5bit 表示计数值，尾部 2bit 表示连续几个桶。它的意思是连续 （xx +1） 个计数值都是 （vvvvv + 1）。比如 10101011 表示连续 4 个计数值都是 11。注意这两个值都需要加 1，因为任意一个是零都意味着这个计数值为零，那就应该使用零计数值的形式来表示。注意计数值最大只能表示到32，而 HyperLogLog 的密集存储单个计数值用 6bit 表示，最大可以表示到 63。当稀疏存储的某个计数值需要调整到大于 32 时，Redis 就会立即转换 HyperLogLog 的存储结构，将稀疏存储转换成密集存储。

![image](https://mail.wangkekai.cn/EC8A16A9-E67B-4DA4-A4ED-9950783BAE73.png)

Redis 为了方便表达稀疏存储，它将上面三种字节表示形式分别赋予了一条指令。

1. `ZERO`: len 单个字节表示 00[len-1]，连续最多64个零计数值
2. `VAL`: value,len 单个字节表示 1[value-1][len-1]，连续 len 个值为 value 的计数值
3. `XZERO`: len 双字节表示 01[len-1]，连续最多16384个零计数值

上图可以用指令形式表示如下：
![image](https://mail.wangkekai.cn/B3480193-7761-4327-92B7-C498BEEC62A9.png)

### 存储转换

当计数值达到一定程度后，稀疏存储将会不可逆一次性转换为密集存储。转换的条件有两个，任意一个满足就会立即发生转换

1. 任意一个计数值从 32 变成 33，因为 VAL 指令已经无法容纳，它能表示的计数值最大为 32
2. 稀疏存储占用的总字节数超过 3000 字节，这个阈值可以通过 hll_sparse_max_bytes 参数进行调整。

### 计数缓存

前面提到 HyperLogLog 表示的总计数值是由 16384 个桶的计数值进行调和平均后再基于因子修正公式计算得出来的。它需要遍历所有的桶进行计算才可以得到这个值，中间还涉及到很多浮点运算。这个计算量相对来说还是比较大的。

所以 Redis 使用了一个额外的字段来缓存总计数值，这个字段有 64bit，最高位如果为 1 表示该值是否已经过期，如果为 0， 那么剩下的 63bit 就是计数值。

当 HyperLogLog 中任意一个桶的计数值发生变化时，就会将计数缓存设为过期，但是不会立即触发计算。而是要等到用户显示调用 pfcount 指令时才会触发重新计算刷新缓存。缓存刷新在密集存储时需要遍历 16384 个桶的计数值进行调和平均，但是稀疏存储时没有这么大的计算量。也就是说只有当计数值比较大时才可能产生较大的计算量。另一方面如果计数值比较大，那么大部分 pfadd 操作根本不会导致桶中的计数值发生变化。

这意味着在一个极具变化的 HLL 计数器中频繁调用 pfcount 指令可能会有少许性能问题。关于这个性能方面的担忧在 Redis 作者 antirez 的博客中也提到了。不过作者做了仔细的压力的测试，发现这是无需担心的，pfcount 指令的平均时间复杂度就是 O(1)。

### 对象头

HyperLogLog 除了需要存储 16384 个桶的计数值之外，它还有一些附加的字段需要存储，比如总计数缓存、存储类型。所以它使用了一个额外的对象头来表示。

```c++
struct hllhdr {
    char magic[4];      /* 魔术字符串"HYLL" */
    uint8_t encoding;   /* 存储类型 HLL_DENSE or HLL_SPARSE. */
    uint8_t notused[3]; /* 保留三个字节未来可能会使用 */
    uint8_t card[8];    /* 总计数缓存 */
    uint8_t registers[]; /* 所有桶的计数器 */
};
```

所以 HyperLogLog 整体的内部结构就是 HLL 对象头 加上 16384 个桶的计数值位图。它在 Redis 的内部结构表现就是一个字符串位图。你可以把 HyperLogLog 对象当成普通的字符串来进行处理。

```redis
127.0.0.1:6379> pfadd codehole python java golang
(integer) 1
127.0.0.1:6379> get codehole
"HYLL\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x80C\x03\x84MK\x80P\xb8\x80^\xf3"
```

但是不可以使用 HyperLogLog 指令来操纵普通的字符串，因为它需要检查对象头魔术字符串是否是 "HYLL"。

```shell
127.0.0.1:6379> set codehole python
OK
127.0.0.1:6379> pfadd codehole java golang
(error) WRONGTYPE Key is not a valid HyperLogLog string value.
```

但是如果字符串以 "HYLL\x00" 或者 "HYLL\x01" 开头，那么就可以使用 HyperLogLog 的指令。

```shell
127.0.0.1:6379> set codehole "HYLL\x01whatmagicthing"
OK
127.0.0.1:6379> get codehole
"HYLL\x01whatmagicthing"
127.0.0.1:6379> pfadd codehole python java golang
(integer) 1
```

这是因为 HyperLogLog 在执行指令前需要对内容进行格式检查，这个检查就是查看对象头的 magic 魔术字符串是否是 "HYLL" 以及 encoding 字段是否是 HLL_SPARSE=0 或者 HLL_DENSE=1 来判断当前的字符串是否是 HyperLogLog 计数器。如果是密集存储，还需要判断字符串的长度是否恰好等于密集计数器存储的长度。

```c++
int isHLLObjectOrReply(client *c, robj *o) {
    ...
    /* Magic should be "HYLL". */
    if (hdr->magic[0] != 'H' || hdr->magic[1] != 'Y' ||
        hdr->magic[2] != 'L' || hdr->magic[3] != 'L') goto invalid;

    if (hdr->encoding > HLL_MAX_ENCODING) goto invalid;

    if (hdr->encoding == HLL_DENSE &&
        stringObjectLen(o) != HLL_DENSE_SIZE) goto invalid;

    return C_OK;

invalid:
    addReplySds(c,
        sdsnew("-WRONGTYPE Key is not a valid "
               "HyperLogLog string value.\r\n"));
    return C_ERR;
}
```

HyperLogLog 和 字符串的关系就好比 Geo 和 zset 的关系。你也可以使用任意 zset 的指令来访问 Geo 数据结构，因为 Geo 内部存储就是使用了一个纯粹的 zset来记录元素的地理位置。

## 10. 布隆过滤器

布隆过滤器可以理解为一个不怎么精确的 set 结构，当你使用它的 contains 方法判断某个对象是否存在时，它可能会误判。但是布隆过滤器也不是特别不精确，只要参数设置的合理，它的精确度可以控制的相对足够精确，只会有小小的误判概率。  
**当布隆过滤器说某个值存在时，这个值可能不存在；当它说不存在时，那就肯定不存在。**

### 布隆过滤器基本使用

布隆过滤器有二个基本指令，bf.add 添加元素，bf.exists 查询元素是否存在，它的用法和 set 集合的 sadd 和 sismember 差不多。注意 bf.add 只能一次添加一个元素，如果想要一次添加多个，就需要用到 bf.madd 指令。同样如果需要一次查询多个元素是否存在，就需要用到 bf.mexists 指令。

```redis
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

**这个限流需求中存在一个滑动时间窗口**，想想 zset 数据结构的 score 值，是不是可以通过 score 来圈出这个时间窗口来。**而且我们只需要保留这个时间窗口，窗口之外的数据都可以砍掉**。那这个 zset 的 value 填什么比较合适呢？它只需要保证唯一性即可，用 uuid 会比较浪费空间，那就改用毫秒时间戳吧。

```python
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

用一个 zset 结构记录用户的行为历史，每一个行为都会作为 zset 中的一个 key 保存下来。同一个用户同一种行为用一个 zset 记录。  
为节省内存，我们只需要保留时间窗口内的行为记录，同时如果用户是冷用户，滑动时间窗口内的行为是空记录，那么这个 zset 就可以从内存中移除，不再占用空间。

## 12. 漏斗限流

Redis 4.0 提供了一个限流 Redis 模块，它叫 redis-cell。该模块也使用了漏斗算法，并提供了原子的限流指令。有了这个模块，限流问题就非常简单了。

```redis
> cl.throttle laoqian:reply 15 30 60 1
                      ▲     ▲  ▲  ▲  ▲
                      |     |  |  |  └───── need 1 quota (可选参数，默认值也是1)
                      |     |  └──┴─────── 30 operations / 60 seconds 这是漏水速率
                      |     └───────────── 15 capacity 这是漏斗容量
                      └─────────────────── key laoqian
```

上面这个指令的意思是允许「用户老钱回复行为」的频率为每 60s 最多 30 次(漏水速率)，漏斗的初始容量为 15，也就是说一开始可以连续回复 15 个帖子，然后才开始受漏水速率的影响。我们看到这个指令中漏水速率变成了 2 个参数，替代了之前的单个浮点数。用两个参数相除的结果来表达漏水速率相对单个浮点数要更加直观一些。

```redis
> cl.throttle laoqian:reply 15 30 60
1) (integer) 0   # 0 表示允许，1表示拒绝
2) (integer) 15  # 漏斗容量capacity
3) (integer) 14  # 漏斗剩余空间left_quota
4) (integer) -1  # 如果拒绝了，需要多长时间后再试(漏斗有空间了，单位秒)
5) (integer) 2   # 多长时间后，漏斗完全空出来(left_quota==capacity，单位秒)
```

在执行限流指令时，如果被拒绝了，就需要丢弃或重试。cl.throttle 指令考虑的非常周到，连重试时间都帮你算好了，直接取返回结果数组的第四个值进行 sleep 即可，如果不想阻塞线程，也可以异步定时任务来重试。

## 13. GeoHash

GeoHash 算法将二维的经纬度数据映射到一维的整数，这样所有的元素都将在挂载到一条线上，距离靠近的二维坐标映射到一维后的点之间距离也会很接近。

在使用 Redis 进行 Geo 查询时，我们要时刻想到它的内部结构实际上只是一个 zset(skiplist)。通过 zset 的 score 排序就可以得到坐标附近的其它元素 (实际情况要复杂一些，不过这样理解足够了)，通过将 score 还原成坐标值就可以得到元素的原始坐标。

### Geo 基本指令使用

#### 增加

geoadd 指令携带集合名称以及多个经纬度名称三元组，注意这里可以加入多个三元组。

```redis
127.0.0.1:6379> geoadd company 116.48105 39.996794 juejin
(integer) 1
// ...
127.0.0.1:6379> geoadd company 116.562108 39.787602 jd 116.334255 40.027400 xiaomi
(integer) 2
```

> 也许你会问为什么 Redis 没有提供 geo 删除指令？前面我们提到 geo 存储结构上使用的是 zset，意味着我们可以使用 zset 相关的指令来操作 geo 数据，所以删除指令可以直接使用 zrem 指令即可。

##### 距离

 geodist 指令可以用来计算两个元素之间的距离，携带集合名称、2 个名称和距离单位。

```redis
127.0.0.1:6379> geodist company juejin ireader km
"10.5501"
// ...
127.0.0.1:6379> geodist company juejin juejin km
"0.0000"
```

##### 获取元素位置

geopos 指令可以获取集合中任意元素的经纬度坐标，可以一次获取多个。

```redis
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

geohash 可以获取元素的经纬度编码字符串，上面已经提到，它是 base32 编码。 你可以使用这个编码值去 geohash.org/${hash}中进行直… geohash 的标准编码值。

```redis
127.0.0.1:6379> geohash company ireader
1) "wx4g52e1ce0"
127.0.0.1:6379> geohash company juejin
1) "wx4gd94yjn0"
```

##### 附近的公司

georadiusbymember 指令是最为关键的指令，它可以用来查询指定元素附近的其它元素，它的参数非常复杂。

```redis
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

除了 `georadiusbymember` 指令根据元素查询附近的元素，Redis 还提供了根据坐标值来查询附近的元素，这个指令更加有用，它可以根据用户的定位来计算「附近的车」，「附近的餐馆」等。它的参数和 georadiusbymember 基本一致，除了将目标元素改成经纬度坐标值。

```redis
127.0.0.1:6379> georadius company 116.514202 39.905409 20 km withdist count 3 asc
1) 1) "ireader"
   2) "0.0000"
2) 1) "juejin"
   2) "10.5501"
3) 1) "meituan"
   2) "11.5748"
```

## 14. scan

在平时线上 Redis 维护工作中，有时候需要从 Redis 实例成千上万的 key 中找出特定前缀的 key 列表来手动处理数据，可能是修改它的值，也可能是删除 key。这里就有一个问题，如何从海量的 key 中找出满足特定前缀的 key 列表来？

keys 用来列出所有满足特定正则字符串规则的 key。

这个指令使用非常简单，提供一个简单的正则字符串即可，但是有很明显的两个缺点。

1. 没有 offset、limit 参数，一次性吐出所有满足条件的 key，万一实例中有几百 w 个 key 满足条件，当你看到满屏的字符串刷的没有尽头时，你就知道难受了。
2. keys 算法是遍历算法，复杂度是 O(n)，如果实例中有千万级以上的 key，这个指令就会导致 Redis 服务卡顿，所有读写 Redis 的其它的指令都会被延后甚至会超时报错，因为 Redis 是单线程程序，顺序执行所有指令，其它指令必须等到当前的 keys 指令执行完了才可以继续。

> scan 相比 keys 具备有以下特点:

1. 复杂度虽然也是 O(n)，但是它是通过游标分步进行的，不会阻塞线程
2. 提供 limit 参数，可以控制每次返回结果的最大条数，limit 只是一个 hint，返回的结果可多可少
3. 同 keys 一样，它也提供模式匹配功能
4. 服务器不需要为游标保存状态，游标的唯一状态就是 scan 返回给客户端的游标整数
5. 返回的结果可能会有重复，需要客户端去重复，这点非常重要
6. 遍历的过程中如果有数据修改，改动后的数据能不能遍历到是不确定的
7. 单次返回的结果是空的并不意味着遍历结束，而要看返回的游标值是否为零

### 基础使用

scan 参数提供了三个参数，第一个是 cursor 整数值，第二个是 key 的正则模式，第三个是遍历的 limit hint。第一次遍历时，cursor 值为 0，然后将返回结果中第一个整数值作为下一次遍历的 cursor。一直遍历到返回的 cursor 值为 0 时结束。

```redis
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

从上面的过程可以看到虽然提供的 limit 是 1000，但是返回的结果只有 10 个左右。**因为这个 limit 不是限定返回结果的数量，而是限定服务器单次遍历的字典槽位数量(约等于)**。如果将 limit 设置为 10，你会发现返回结果是空的，*但是游标值不为零，意味着遍历还没结束*。

```redis
127.0.0.1:6379> scan 0 match key99* count 10
1) "3072"
2) (empty list or set)
```

#### 字典的结构

在 Redis 中所有的 key 都存储在一个很大的字典中，这个字典的结构和 Java 中的 HashMap 一样，是一维数组 + 二维链表结构，第一维数组的大小总是 2^n(n>=0)，扩容一次数组大小空间加倍，也就是 n++
![image](https://mail.wangkekai.cn/5DED180C-BA39-4CD4-95C3-0A34144E80E0.png)

scan 指令返回的游标就是第一维数组的位置索引，我们将这个位置索引称为槽 (slot)。如果不考虑字典的扩容缩容，直接按数组下标挨个遍历就行了。 limit 参数就表示需要遍历的槽位数，*之所以返回的结果可能多可能少，是因为不是所有的槽位上都会挂接链表，有些槽位可能是空的，还有些槽位上挂接的链表上的元素可能会有多个*。每一次遍历都会将 limit 数量的槽位上挂接的所有链表元素进行模式匹配过滤后，一次性返回给客户端。

#### scan 遍历顺序

scan 的遍历顺序非常特别。它不是从第一维数组的第 0 位一直遍历到末尾，而是采用了高位进位加法来遍历。之所以使用这样特殊的方式进行遍历，是考虑到字典的扩容和缩容时避免槽位的遍历重复和遗漏。
![image](https://mail.wangkekai.cn/16469760d12e0cbd.gif)

从动画中可以看出高位进位法从左边加，进位往右边移动，同普通加法正好相反。但是最终它们都会遍历所有的槽位并且没有重复。

#### 字典扩容

Java 中的 HashMap 有扩容的概念，当 loadFactor 达到阈值时，需要重新分配一个新的 2 倍大小的数组，然后将所有的元素全部 rehash 挂到新的数组下面。rehash 就是将元素的 hash 值对数组长度进行取模运算，因为长度变了，所以每个元素挂接的槽位可能也发生了变化。又因为数组的长度是 2^n 次方，所以取模运算等价于位与操作。

#### 渐进式rehash

Java 的 HashMap 在扩容时会一次性将旧数组下挂接的元素全部转移到新数组下面。如果 HashMap 中元素特别多，**线程就会出现卡顿现象**。Redis 为了解决这个问题，它采用渐进式 rehash。

它会同时保留旧数组和新数组，然后在定时任务中以及后续对  hash 的指令操作中渐渐地将旧数组中挂接的元素迁移到新数组上。这意味着要操作处于  rehash 中的字典，需要同时访问新旧两个数组结构。如果在旧数组下面找不到元素，还需要去新数组下面去寻找。

scan 也需要考虑这个问题，对与 rehash 中的字典，它需要同时扫描新旧槽位，然后将结果融合后返回给客户端。

#### 更多的 scan 指令

scan 指令是一系列指令，除了可以遍历所有的 key 之外，还可以对指定的容器集合进行遍历。比如 zscan 遍历 zset 集合元素，hscan 遍历 hash 字典的元素、sscan 遍历 set 集合的元素。

它们的原理同 scan 都会类似的，因为 hash 底层就是字典，set 也是一个特殊的 hash(所有的 value 指向同一个元素)，zset 内部也使用了字典来存储所有的元素内容，所以这里不再赘述。

#### 大 key 扫描

为了避免对线上 Redis 带来卡顿，这就要用到 scan 指令，对于扫描出来的每一个 key，使用 type 指令获得  key 的类型，然后使用相应数据结构的 size 或者 len 方法来得到它的大小，对于每一种类型，保留大小的前 N 名作为扫描结果展示出来。

上面这样的过程需要编写脚本，比较繁琐，不过 Redis 官方已经在 redis-cli 指令中提供了这样的扫描功能，我们可以直接拿来即用。

```redis
redis-cli -h 127.0.0.1 -p 7001 –-bigkeys
```

如果你担心这个指令会大幅抬升 Redis 的 ops 导致线上报警，还可以增加一个休眠参数。

```redis
redis-cli -h 127.0.0.1 -p 7001 –-bigkeys -i 0.1
```

上面这个指令每隔 100 条 scan 指令就会休眠 0.1s，ops 就不会剧烈抬升，但是扫描的时间会变长。

## 15. 线程 IO 模型

Redis 是个单线程程序！

### Redis 单线程为什么还能这么快？

因为它所有的数据都在内存中，所有的运算都是内存级别的运算。正因为 Redis 是单线程，所以要小心使用 Redis 指令，对于那些时间复杂度为 O(n) 级别的指令，一定要谨慎使用，一不小心就可能会导致 Redis 卡顿。

#### Redis 单线程如何处理那么多的并发客户端连接？

这个问题，有很多中高级程序员都无法回答，因为他们没听过多路复用这个词汇，不知道 select 系列的事件轮询 API，没用过非阻塞 IO。

#### 非阻塞 IO

当我们调用套接字的读写方法，默认它们是阻塞的，比如 read 方法要传递进去一个参数 n，表示最多读取这么多字节后再返回，如果一个字节都没有，那么线程就会卡在那里，直到新的数据到来或者连接关闭了，read 方法才可以返回，线程才能继续处理。而 write 方法一般来说不会阻塞，除非内核为套接字分配的写缓冲区已经满了，write 方法就会阻塞，直到缓存区中有空闲空间挪出来了。
![image](https://mail.wangkekai.cn/E505D0AC-47FC-4790-9812-5E7BCCF628DE.png)

非阻塞 IO 在套接字对象上提供了一个选项 Non_Blocking，当这个选项打开时，**读写方法不会阻塞，而是能读多少读多少，能写多少写多少**。能读多少取决于内核为套接字分配的读缓冲区内部的数据字节数，能写多少取决于内核为套接字分配的写缓冲区的空闲空间字节数。读方法和写方法都会通过返回值来告知程序实际读写了多少字节。

有了非阻塞 IO 意味着线程在读写 IO 时可以不必再阻塞了，读写可以瞬间完成然后线程可以继续干别的事了。

#### 事件轮询 (多路复用)

非阻塞 IO 有个问题，那就是线程要读数据，结果读了一部分就返回了，线程如何知道何时才应该继续读。也就是当数据到来时，线程如何得到通知。写也是一样，如果缓冲区满了，写不完，剩下的数据何时才应该继续写，线程也应该得到通知。
![image](https://mail.wangkekai.cn/5A55BFDD-3FCB-48E1-81DF-566BC251B2CF.png)

事件轮询 API 就是用来解决这个问题的，最简单的事件轮询 API 是 select 函数，它是操作系统提供给用户程序的 API。输入是读写描述符列表 read_fds & write_fds，输出是与之对应的可读可写事件。同时还提供了一个 timeout 参数，如果没有任何事件到来，那么就最多等待 timeout 时间，线程处于阻塞状态。一旦期间有任何事件到来，就可以立即返回。时间过了之后还是没有任何事件到来，也会立即返回。拿到事件后，线程就可以继续挨个处理相应的事件。处理完了继续过来轮询。于是线程就进入了一个死循环，我们把这个死循环称为事件循环，一个循环为一个周期。

#### 指令队列

Redis 会将每个客户端套接字都关联一个指令队列。客户端的指令通过队列来排队进行顺序处理，先到先服务。

#### 响应队列

Redis 同样也会为每个客户端套接字关联一个响应队列。Redis 服务器通过响应队列来将指令的返回结果回复给客户端。 如果队列为空，那么意味着连接暂时处于空闲状态，不需要去获取写事件，也就是可以将当前的客户端描述符从 write_fds 里面移出来。等到队列有数据了，再将描述符放进去。避免 select 系统调用立即返回写事件，结果发现没什么数据可以写。出这种情况的线程会飙高 CPU。

#### 定时任务

服务器处理要响应 IO 事件外，还要处理其它事情。比如定时任务就是非常重要的一件事。如果线程阻塞在 select 系统调用上，定时任务将无法得到准时调度。那 Redis 是如何解决这个问题的呢？

Redis 的定时任务会记录在一个称为最小堆的数据结构中。这个堆中，最快要执行的任务排在堆的最上方。在每个循环周期，Redis 都会将最小堆里面已经到点的任务立即进行处理。处理完毕后，将最快要执行的任务还需要的时间记录下来，这个时间就是 select 系统调用的 timeout 参数。因为 Redis 知道未来 timeout 时间内，没有其它定时任务需要处理，所以可以安心睡眠 timeout 的时间。

Nginx 和 Node 的事件处理原理和 Redis 也是类似的

## 16. 通信协议

## 17. 持久化

Redis 的持久化机制有两种，*第一种是快照，第二种是 AOF 日志*。**快照是一次全量备份，AOF 日志是连续的增量备份**。

快照是内存数据的二进制序列化形式，在存储上非常紧凑，而 *AOF 日志记录* 的是内存数据修改的指令记录文本。

AOF 日志在长期的运行过程中会变的无比庞大，数据库重启时需要加载 AOF 日志进行指令重放，这个时间就会无比漫长。所以需要定期进行 AOF 重写，给 AOF 日志进行瘦身。

### 快照原理

我们知道 Redis 是单线程程序，这个线程要同时负责多个客户端套接字的并发读写操作和内存数据结构的逻辑读写。

在服务线上请求的同时，Redis 还需要进行内存快照，内存快照要求 Redis 必须进行文件 IO 操作，可文件 IO 操作是不能使用多路复用 API。

这意味着单线程同时在服务线上的请求还要进行文件 IO 操作，文件 IO 操作会严重拖垮服务器请求的性能。还有个重要的问题是为了不阻塞线上的业务，就需要边持久化边响应客户端请求。持久化的同时，内存数据结构还在改变，比如一个大型的 hash 字典正在持久化，结果一个请求过来把它给删掉了，还没持久化完呢，这尼玛要怎么搞？

Redis 使用操作系统的多进程 COW(Copy On Write) 机制来实现快照持久化，这个机制很有意思，也很少人知道。多进程 COW 也是鉴定程序员知识广度的一个重要指标。

#### fork(多进程)

Redis 在持久化时会调用 glibc 的函数 fork 产生一个子进程，快照持久化完全交给子进程来处理，父进程继续处理客户端请求。子进程刚刚产生时，它和父进程共享内存里面的代码段和数据段。这时你可以将父子进程想像成一个连体婴儿，共享身体。这是 Linux 操作系统的机制，为了节约内存资源，所以尽可能让它们共享起来。在进程分离的一瞬间，内存的增长几乎没有明显变化。

**子进程做数据持久化，它不会修改现有的内存数据结构，它只是对数据结构进行遍历读取，然后序列化写到磁盘中。但是父进程不一样，它必须持续服务客户端请求，然后对内存数据结构进行不间断的修改。**

随着父进程修改操作的持续进行，越来越多的共享页面被分离出来，内存就会持续增长。但是也不会超过原有数据内存的 2 倍大小。**另外一个 Redis 实例里冷数据占的比例往往是比较高的，所以很少会出现所有的页面都会被分离，被分离的往往只有其中一部分页面。*每个页面的大小只有 4K，一个 Redis 实例里面一般都会有成千上万的页面。***。

子进程因为数据没有变化，它能看到的内存里的数据在进程产生的一瞬间就凝固了，再也不会改变，这也是为什么 Redis 的持久化叫「快照」的原因。接下来子进程就可以非常安心的遍历数据了进行序列化写磁盘了。

### AOF 原理

**AOF 日志存储的是 Redis 服务器的顺序指令序列，AOF 日志只记录对内存进行修改的指令记录**。

假设 AOF 日志记录了自 Redis 实例创建以来所有的修改性指令序列，那么就可以通过对一个空的 Redis 实例顺序执行所有的指令，也就是「重放」，来恢复 Redis 当前实例的内存数据结构的状态。

Redis 会在收到客户端修改指令后，进行参数校验进行逻辑处理后，如果没问题，就立即将该指令文本存储到 AOF 日志中，也就是*先执行指令才将日志存盘*。这点不同于 leveldb、hbase 等存储引擎，它们都是先存储日志再做逻辑处理。

Redis 在长期运行的过程中，AOF 的日志会越变越长。如果实例宕机重启，重放整个 AOF 日志会非常耗时，导致长时间 Redis 无法对外提供服务。所以需要对 AOF 日志瘦身。

#### AOF 重写

Redis 提供了 `bgrewriteaof` 指令用于对 AOF 日志进行瘦身。**其原理就是开辟一个子进程对内存进行遍历转换成一系列 Redis 的操作指令，序列化到一个新的 AOF 日志文件中**。序列化完毕后再将操作期间发生的增量 AOF 日志追加到这个新的 AOF 日志文件中，追加完毕后就立即替代旧的 AOF 日志文件了，瘦身工作就完成了。

#### fsync

AOF 日志是以文件的形式存在的，当程序对 AOF 日志文件进行写操作时，*实际上是将内容写到了内核为文件描述符分配的一个内存缓存中，然后内核会异步将脏数据刷回到磁盘的*。

Linux 的 glibc提供了 fsync(int fd) 函数可以将指定文件的内容强制从内核缓存刷到磁盘。**只要 Redis 进程实时调用 fsync 函数就可以保证 aof 日志不丢失。但是 fsync 是一个磁盘 IO 操作，它很慢**！如果 Redis 执行一条指令就要 fsync 一次，那么 Redis 高性能的地位就不保了。

所以在生产环境的服务器中，Redis 通常是每隔 1s 左右执行一次 fsync 操作，周期 1s 是可以配置的。这是在数据安全性和性能之间做了一个折中，在保持高性能的同时，尽可能使得数据少丢失。

Redis 同样也提供了另外两种策略，**一个是永不 fsync——让操作系统来决定何时同步磁盘，很不安全，另一个是来一个指令就 fsync 一次——非常慢**。但是在生产环境基本不会使用，了解一下即可。

> 通常 Redis 的主节点是不会进行持久化操作，持久化操作主要在从节点进行。从节点是备份节点，没有来自客户端请求的压力，它的操作系统资源往往比较充沛。

#### Redis 4.0 混合持久化

> 重启 Redis 时，我们很少使用 rdb 来恢复内存状态，因为会丢失大量数据。我们通常使用 AOF 日志重放，但是重放 AOF 日志性能相对 rdb 来说要慢很多，这样在 Redis 实例很大的情况下，启动需要花费很长的时间。

将 rdb 文件的内容和增量的 AOF 日志文件存在一起。这里的 AOF 日志不再是全量的日志，而是自持久化开始到持久化结束的这段时间发生的增量 AOF 日志，通常这部分 AOF 日志很小。
![image](https://mail.wangkekai.cn/3354A8D7-3098-4448-81D8-ED5BBF6B0CC6.png)

于是在 Redis 重启的时候，**可以先加载 rdb 的内容，然后再重放增量 AOF 日志就可以完全替代之前的 AOF 全量文件重放，重启效率因此大幅得到提升**。

## 18. 管道

### redis 的消息交互

![image](https://mail.wangkekai.cn/166AC5E1-3578-4428-8CC9-9A8904D404CE.png)

两个连续的写操作和两个连续的读操作总共只会花费一次网络来回，就好比连续的 write 操作合并了，连续的 read 操作也合并了一样。

客户端通过对管道中的指令列表改变读写顺序就可以大幅节省 IO 时间。管道中指令越多，效果越好。

### 深入理解管道本质

![image](https://mail.wangkekai.cn/05CA3F48-4FDB-4A07-9A43-6DAB8D752064.png)

1. 客户端进程调用 write 将消息写到操作系统内核为套接字分配的发送缓冲 send buffer。
2. 客户端操作系统内核将发送缓冲的内容发送到网卡，网卡硬件将数据通过「网际路由」送到服务器的网卡。
3. 服务器操作系统内核将网卡的数据放到内核为套接字分配的接收缓冲 recv buffer。
4. 服务器进程调用 read 从接收缓冲中取出消息进行处理。
5. 服务器进程调用 write 将响应消息写到内核为套接字分配的发送缓冲 send buffer。
6. 服务器操作系统内核将发送缓冲的内容发送到网卡，网卡硬件将数据通过「网际路由」送到客户端的网卡。
7. 客户端操作系统内核将网卡的数据放到内核为套接字分配的接收缓冲 recv buffer。
8. 客户端进程调用 read 从接收缓冲中取出消息返回给上层业务逻辑进行处理。
9. 结束。

## 19. 事物

### 基本使用

multi/exec/discard。multi 指示事务的开始，exec 指示事务的执行，discard 指示事务的丢弃。

```redis
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

所有的指令在 exec 之前不执行，**而是缓存在服务器的一个事务队列中**，服务器一旦收到 exec 指令，才开执行整个事务队列，执行完毕后一次性返回所有指令的运行结果。因为 Redis 的单线程特性，它不用担心自己在执行队列的时候被其它指令打搅，可以保证他们能得到的「原子性」执行。

### 原子性

```redis
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

你应该明白 Redis 的**事务根本不能算「原子性」，而仅仅是满足了事务的「隔离性」**，隔离性中的串行化——当前执行的事务有着不被其它事务打断的权利。

### 优化

上面的 Redis 事务在发送每个指令到事务缓存队列时都要经过一次网络读写，当一个事务内部的指令较多时，需要的网络 IO 时间也会线性增长。所以通常 Redis 的客户端在执行事务时都会结合 pipeline 一起使用，这样可以将多次 IO 操作压缩为单次 IO 操作。

```redis
pipe = redis.pipeline(transaction=true)
pipe.multi()
pipe.incr("books")
pipe.incr("books")
values = pipe.execute()
```

### watch

有多个客户端会并发进行操作。我们可以通过 Redis 的**分布式锁**来避免冲突，这是一个很好的解决方案。**分布式锁是一种悲观锁**，那是不是可以使用乐观锁的方式来解决冲突呢？

Redis 提供了这种 watch 的机制，它就是一种乐观锁。有了 watch 我们又多了一种可以用来解决并发修改的方法。watch 的使用方式如下：

```redis
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

## 20.  PubSub - 小道消息

### 消息多播

消息多播允许生产者生产一次消息，中间件负责将消息复制到多个消息队列，每个消息队列由相应的消费组进行消费。它是分布式系统常用的一种解耦方式，用于将多个消费组的逻辑进行拆分。支持了消息多播，多个消费组的逻辑就可以放到不同的子系统中。

### PubSub

为了支持消息多播，Redis 不能再依赖于那 5 种基本数据类型了。**它单独使用了一个模块来支持消息多播，这个模块的名字叫着  PubSub，也就是 PublisherSubscriber，发布者订阅者模型**。我们使用 Python 语言来演示一下 PubSub 如何使用。

```redis
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

客户端发起订阅命令后，Redis 会立即给予一个反馈消息通知订阅成功。因为有网络传输延迟，在 subscribe 命令发出后，需要休眠一会，再通过 get_message 才能拿到反馈消息。客户端接下来执行发布命令，发布了一条消息。同样因为网络延迟，在 publish 命令发出后，需要休眠一会，再通过 get_message 才能拿到发布的消息。**如果当前没有消息，get_message 会返回空，告知当前没有消息，所以它不是阻塞的**。

Redis PubSub 的生产者和消费者是不同的连接，也就是上面这个例子实际上使用了两个 Redis 的连接。这是必须的，因为 Redis 不允许连接在 subscribe 等待消息时还要进行其它的操作。

### 模式订阅

简化订阅的繁琐，redis 提供了模式订阅功能 Pattern Subscribe，这样就可以一次订阅多个主题，即使生产者新增加了同模式的主题，消费者也可以立即收到消息。

```redis
> psubscribe codehole.*  # 用模式匹配一次订阅多个主题，主题以 codehole. 字符开头的消息都可以收到
1) "psubscribe"
2) "codehole.*"
3) (integer) 1

```

### PubSub 缺点

PubSub 的生产者传递过来一个消息，Redis 会直接找到相应的消费者传递过去。如果一个消费者都没有，那么消息直接丢弃。如果开始有三个消费者，一个消费者突然挂掉了，生产者会继续发送消息，另外两个消费者可以持续收到消息。但是挂掉的消费者重新连上的时候，这断连期间生产者发送的消息，对于这个消费者来说就是彻底丢失了。

如果 Redis 停机重启，PubSub 的消息是不会持久化的，毕竟 Redis 宕机就相当于一个消费者都没有，所有的消息直接被丢弃。

## 21. 小对象压缩

### 32bit vs 64bit

Redis 如果使用 32bit 进行编译，内部所有数据结构所使用的指针空间占用会少一半，如果你对 Redis 使用内存不超过 4G，可以考虑使用 32bit 进行编译，可以节约大量内存。

### 小对象压缩存储 ( ziplist)

如果 Redis 内部管理的集合数据结构很小，它会使用紧凑存储形式压缩存储。

这就好比 HashMap 本来是二维结构，但是如果内部元素比较少，使用二维结构反而浪费空间，还不如使用一维数组进行存储，需要查找时，因为元素少进行遍历也很快，甚至可以比 HashMap 本身的查找还要快。比如下面我们可以使用数组来模拟 HashMap 的增删改操作。

Redis 的 ziplist 是一个紧凑的字节数组结构，如下图所示，每个元素之间都是紧挨着的。我们不用过于关心 zlbytes/zltail 和 zlend 的含义，稍微了解一下就好。
![image](https://mail.wangkekai.cn/D6AC88EE-8543-4E17-BB49-B258DBF9AF2B.png)

> Redis 的 intset 是一个紧凑的整数数组结构，它用于存放元素都是整数的并且元素个数较少的  set 集合。

如果整数可以用 uint16 表示，那么 intset 的元素就是 16 位的数组，如果新加入的整数超过了 uint16 的表示范围，那么就使用 uint32 表示，如果新加入的元素超过了 uint32 的表示范围，那么就使用 uint64 表示，Redis 支持 set 集合动态从 uint16 升级到 uint32，再升级到 uint64。

### 内存回收机制

如果当前  Redis 内存有  10G，当你删除了  1GB 的  key 后，再去观察内存，你会发现内存变化不会太大。**原因是操作系统回收内存是以页为单位，<font color=red>如果这个页上只要有一个 key 还在使用，那么它就不能被回收</font>**。Redis 虽然删除了 1GB 的 key，但是这些 key 分散到了很多页面中，每个页面都还有其它 key 存在，这就导致了内存不会立即被回收。

Redis 虽然无法保证立即回收已经删除的  key 的内存，但是它会重用那些尚未回收的空闲内存。这就好比电影院里虽然人走了，但是座位还在，下一波观众来了，直接坐就行。而操作系统回收内存就好比把座位都给搬走了。

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

Redis 同步的是指令流，主节点会将那些对自己的状态产生修改性影响的指令记录在本地的内存 buffer 中，然后异步将 buffer 中的指令同步到从节点，从节点一边执行同步的指令流来达到和主节点一样的状态，一边向主节点反馈自己同步到哪里了 (偏移量)。

因为内存的 buffer 是有限的，所以 Redis 主库不能将所有的指令都记录在内存 buffer 中。Redis 的复制内存 buffer 是一个定长的环形数组，如果数组内容满了，就会从头开始覆盖前面的内容。

如果因为网络状况不好，从节点在短时间内无法和主节点进行同步，那么当网络状况恢复时，Redis 的主节点中那些没有同步的指令在 buffer 中有可能已经被后续的指令覆盖掉了，从节点将无法直接通过指令流来进行同步，这个时候就需要用到更加复杂的同步机制 —— **快照同步**。

### 快照同步

快照同步是一个非常耗费资源的操作，它首先需要在主库上进行一次 bgsave 将当前内存的数据全部快照到磁盘文件中，然后再将快照文件的内容全部传送到从节点。从节点将快照文件接受完毕后，立即执行一次全量加载，加载之前先要将当前内存的数据清空。加载完毕后通知主节点继续进行增量同步。

在整个快照同步进行的过程中，主节点的复制 buffer 还在不停的往前移动，如果快照同步的时间过长或者复制 buffer 太小，都会导致同步期间的增量指令在复制 buffer 中被覆盖，这样就会导致快照同步完成后无法进行增量复制，然后会再次发起快照同步，如此极有可能会陷入快照同步的死循环。
![image](https://mail.wangkekai.cn/5EAD43EE-6D7B-4AD6-8383-B8F49C43F89A.png)
### 无盘复制

主节点在进行快照同步时，会进行很重的文件  IO 操作，特别是对于非 SSD 磁盘存储时，快照会对系统的负载产生较大影响。特别是当系统正在进行 AOF 的 fsync 操作时如果发生快照，fsync 将会被推迟执行，这就会严重影响主节点的服务效率。

**所谓<font color=red>无盘复制</font>是指主服务器直接通过套接字将快照内容发送到从节点，生成快照是一个遍历的过程，主节点会一边遍历内存，一边将序列化的内容发送到从节点，从节点还是跟之前一样，先将接收到的内容存储到磁盘文件中，再进行一次性加载**。

### Wait 指令

Redis 的复制是异步进行的，wait 指令可以让异步复制变身同步复制，确保系统的强一致性 (不严格)。 wait 指令是 Redis3.0 版本以后才出现的。

```redis
> set key value
OK
> wait 1 0
(integer) 1
```

wait 提供两个参数，第一个参数是从库的数量 N，第二个参数是时间 t，以毫秒为单位。它表示等待 wait 指令之前的所有写操作同步到 N 个从库 (也就是确保 N 个从库的同步没有滞后)，最多等待时间 t。如果时间 t=0，表示无限等待直到 N 个从库同步完成达成一致。

假设此时出现了网络分区，wait 指令第二个参数时间 t=0，主从同步无法继续进行，wait 指令会永远阻塞，Redis 服务器将丧失可用性。

## 23.  Sentinel 哨兵

![image](https://mail.wangkekai.cn/34911A9B-E5F0-4694-980D-751714B960D7.png)

它负责**持续监控主从节点的健康，当主节点挂掉时，自动选择一个最优的从节点切换为主节点**。客户端来连接集群时，会首先连接 sentinel，通过 sentinel 来查询主节点的地址，然后再去连接主节点进行数据交互。当主节点发生故障时，客户端会重新向 sentinel 要地址， sentinel 会将最新的主节点地址告诉客户端。如此应用程序将无需重启即可自动完成节点切换。比如上图的主节点挂掉后，集群将可能自动调整为下图所示结构。

### 消息丢失

Redis 主从采用异步复制，意味着当主节点挂掉时，从节点可能没有收到全部的同步消息，这部分未同步的消息就丢失了。如果主从延迟特别大，那么丢失的数据就可能会特别多。 sentinel 无法保证消息完全不丢失，但是也尽可能保证消息少丢失。它有两个选项可以限制主从延迟过大。

```redis
min-slaves-to-write 1
min-slaves-max-lag 10
```

第一个参数表示主节点必须至少有一个从节点在进行正常复制，否则就停止对外写服务，丧失可用性。

第二个参数控制的，它的单位是秒，表示如果 10s 没有收到从节点的反馈，就意味着从节点同步不正常，要么网络断开了，要么一直没有给反馈。

## 24. Cluster

![image](https://mail.wangkekai.cn/F42ED9EA-2F1A-4716-B3BC-D7CC72467CC9.png)

它是去中心化的，该集群有三个 Redis 节点组成，每个节点负责整个集群的一部分数据，每个节点负责的数据多少可能不一样。这三个节点相互连接组成一个对等的集群，它们之间通过一种特殊的二进制协议相互交互集群信息。
![image](https://mail.wangkekai.cn/1614777854058.jpg)

Redis Cluster 将所有数据划分为 16384 的 slots，它比 Codis 的 1024 个槽划分的更为精细，每个节点负责其中一部分槽位。槽位的信息存储于每个节点中，它不像 Codis，<u>它不需要另外的分布式存储来存储节点槽位信息。</u>

当 Redis Cluster 的客户端来连接集群时，它也会得到一份集群的槽位配置信息。这样当客户端要查找某个 key 时，可以直接定位到目标节点。

RedisCluster 的每个节点会将集群的配置信息持久化到配置文件中，所以必须确保配置文件是可写的，而且尽量不要依靠人工修改配置文件。

### 槽位定位算法

Cluster 默认会对 key 值使用 crc16 算法进行 hash 得到一个整数值，然后用这个整数值对 16384 进行取模来得到具体槽位。

Cluster 还允许用户强制某个 key 挂在特定槽位上，通过在 key 字符串里面嵌入 tag 标记，这就可以强制 key 所挂在的槽位等于 tag 所在的槽位。

### 跳转

当客户端向一个错误的节点发出了指令，该节点会发现指令的 key 所在的槽位并不归自己管理，这时它会向客户端发送一个特殊的跳转指令携带目标操作的节点地址，告诉客户端去连这个节点去获取数据。

```redis
GET x
-MOVED 3999 127.0.0.1:6381
```

MOVED 指令的第一个参数 3999 是 key 对应的槽位编号，后面是目标节点地址。MOVED 指令前面有一个减号，表示该指令是一个错误消息。

### 迁移

迁移过程
![image](https://mail.wangkekai.cn/1614778678632.jpg)

Redis 迁移的单位是槽，Redis 一个槽一个槽进行迁移，当一个槽正在迁移时，这个槽就处于中间过渡状态。这个槽在原节点的状态为 migrating，在目标节点的状态为 importing，表示数据正在从源流向目标。

迁移工具 redis-trib 首先会在源和目标节点设置好中间过渡状态，然后一次性获取源节点槽位的所有 key 列表( keysinslot 指令，可以部分获取)，再挨个 key 进行迁移。每个 key 的迁移过程是以原节点作为目标节点的「客户端」，原节点对当前的 key 执行 dump 指令得到序列化内容，然后通过「客户端」向目标节点发送指令 restore 携带序列化的内容作为参数，目标节点再进行反序列化就可以将内容恢复到目标节点的内存中，然后返回「客户端」OK，原节点「客户端」收到后再把当前节点的 key 删除掉就完成了单个 key 迁移的整个过程。

**从源节点获取内容 => 存到目标节点 => 从源节点删除内容。**

注意这里的迁移过程是同步的，在目标节点执行 restore 指令到原节点删除 key 之间，原节点的主线程会处于阻塞状态，直到 key 被成功删除。  
如果迁移过程中突然出现网络故障，整个 slot 的迁移只进行了一半。这时两个节点依旧处于中间过渡状态。待下次迁移工具重新连上时，会提示用户继续进行迁移。  
在迁移过程中，如果每个 key 的内容都很小，migrate 指令执行会很快，它就并不会影响客户端的正常访问。如果 key 的内容很大，因为 migrate 指令是阻塞指令会同时导致原节点和目标节点卡顿，影响集群的稳定型。**所以在集群环境下业务逻辑要尽可能避免大 key 的产生。**

在迁移过程中，客户端访问的流程会有很大的变化。

首先新旧两个节点对应的槽位都存在部分 key 数据。客户端先尝试访问旧节点，如果对应的数据还在旧节点里面，那么旧节点正常处理。如果对应的数据不在旧节点里面，那么有两种可能，要么该数据在新节点里，要么根本就不存在。旧节点不知道是哪种情况，所以它会向客户端返回一个- ASK targetNodeAddr 的重定向指令。客户端收到这个重定向指令后，先去目标节点执行一个不带任何参数的asking指令，然后在目标节点再重新执行原先的操作指令。

为什么需要执行一个不带参数的asking指令呢？

因为在迁移没有完成之前，按理说这个槽位还是不归新节点管理的，如果这个时候向目标节点发送该槽位的指令，节点是不认的，它会向客户端返回一个-MOVED重定向指令告诉它去源节点去执行。如此就会形成 重定向循环。asking指令的目标就是打开目标节点的选项，告诉它下一条指令不能不理，而要当成自己的槽位来处理。

### 容错

Redis Cluster 可以为每个主节点设置若干个从节点，单主节点故障时，集群会自动将其中某个从节点提升为主节点。如果某个主节点没有从节点，那么当它发生故障时，集群将完全处于不可用状态。不过 Redis 也提供了一个参数cluster-require-full-coverage可以允许部分节点故障，其它节点还可以继续提供对外访问。

### 网络抖动

Redis Cluster 提供了一种选项 cluster-node-timeout，表示当某个节点持续 timeout 的时间失联时，才可以认定该节点出现故障，需要进行主从切换。如果没有这个选项，网络抖动会导致主从频繁切换 (数据的重新复制)。

还有另外一个选项 cluster-slave-validity-factor 作为倍乘系数来放大这个超时时间来宽松容错的紧急程度。如果这个系数为零，那么主从切换是不会抗拒网络抖动的。如果这个系数大于 1，它就成了主从切换的松弛系数。

### 可能下线 (PFAIL-Possibly Fail) 与确定下线 (Fail)

因为 Redis Cluster 是去中心化的，一个节点认为某个节点失联了并不代表所有的节点都认为它失联了。所以<u>集群还得经过一次协商的过程，只有当大多数节点都认定了某个节点失联了，集群才认为该节点需要进行主从切换来容错。</u>

Redis 集群节点采用 Gossip 协议来广播自己的状态以及自己对整个集群认知的改变。比如一个节点发现某个节点失联了 (PFail)，它会将这条信息向整个集群广播，其它节点也就可以收到这点失联信息。如果一个节点收到了某个节点失联的数量 (PFail Count) 已经达到了集群的大多数，就可以标记该节点为确定下线状态 (Fail)，然后向整个集群广播，强迫其它节点也接收该节点已经下线的事实，并立即对该失联节点进行主从切换。

### Cluster 基本使用

### 槽位迁移感知

客户端保存了槽位和节点的映射关系表，它需要即时得到更新，client 才可以正常地将某条指令发到正确的节点中。

**Cluster 有两个特殊的 error 指令，一个是 moved，一个是 asking。**

第一个 moved 是用来纠正槽位的。如果我们将指令发送到了错误的节点，该节点发现对应的指令槽位不归自己管理，就会将目标节点的地址随同 moved 指令回复给客户端通知客户端去目标节点去访问。这个时候客户端就会刷新自己的槽位关系表，然后重试指令，后续所有打在该槽位的指令都会转到目标节点。

第二个 asking 指令和 moved 不一样，它是用来临时纠正槽位的。如果当前槽位正处于迁移中，指令会先被发送到槽位所在的旧节点，如果旧节点存在数据，那就直接返回结果了，如果不存在，那么它可能真的不存在也可能在迁移目标节点上。所以旧节点会通知客户端去新节点尝试一下拿数据，看看新节点有没有。这时候就会给客户端返回一个 asking error 携带上目标节点的地址。客户端收到这个 asking error 后，就会去目标节点去尝试。客户端不会刷新槽位映射关系表，因为它只是临时纠正该指令的槽位信息，不影响后续指令。

**重试 2 次**

moved 和 asking 指令都是重试指令，客户端会因为这两个指令多重试一次。客户端有可能重试 2 次呢？这种情况是存在的，比如一条指令被发送到错误的节点，这个节点会先给你一个 moved 错误告知你去另外一个节点重试。所以客户端就去另外一个节点重试了，结果刚好这个时候运维人员要对这个槽位进行迁移操作，于是给客户端回复了一个 asking 指令告知客户端去目标节点去重试指令。所以这里客户端重试了 2 次。

**重试多次**

在某些特殊情况下，客户端甚至会重试多次  
正是因为存在多次重试的情况，所以客户端的源码里在执行指令时都会有一个循环，然后会设置一个最大重试次数，Java 和 Python 都有这个参数，只是设置的值不一样。当重试次数超过这个值时，客户端会直接向业务层抛出异常。

### 集群变更感知

1. 目标节点挂掉了，客户端会抛出一个 ConnectionError，紧接着会随机挑一个节点来重试，这时被重试的节点会通过 moved error 告知目标槽位被分配到的新的节点地址。
2. 运维手动修改了集群信息，将 master 切换到其它节点，并将旧的 master 移除集群。这时打在旧节点上的指令会收到一个 ClusterDown 的错误，告知当前节点所在集群不可用 (当前节点已经被孤立了，它不再属于之前的集群)。这时客户端就会关闭所有的连接，清空槽位映射关系表，然后向上层抛错。待下一条指令过来时，就会重新尝试初始化节点信息。

## 25. 耳听八方 —— Stream

Redis5.0 最大的新特性就是多出了一个数据结构 Stream，它是一个新的强大的支持多播的可持久化的消息队列，作者坦言 Redis Stream 狠狠地借鉴了 Kafka 的设计。
![image](https://mail.wangkekai.cn/1614865676414.jpg)

Redis Stream 的结构如上图所示，它有一个消息链表，将所有加入的消息都串起来，每个消息都有一个唯一的 ID 和对应的内容。消息是持久化的，Redis 重启后，内容还在。

每个 Stream 都有唯一的名称，它就是 Redis 的 key，在我们首次使用 xadd 指令追加消息时自动创建。

每个 Stream 都可以挂多个消费组，每个消费组会有个游标 last_delivered_id 在 Stream 数组之上往前移动，表示当前消费组已经消费到哪条消息了。每个消费组都有一个 Stream 内唯一的名称，消费组不会自动创建，它需要单独的指令 xgroup create 进行创建，需要指定从 Stream 的某个消息 ID 开始消费，这个 ID 用来初始化 last_delivered_id 变量。

同一个消费组 (Consumer Group) 可以挂接多个消费者 (Consumer)，这些消费者之间是竞争关系，任意一个消费者读取了消息都会使游标 last_delivered_id 往前移动。每个消费者有一个组内唯一名称。

消费者 (Consumer) 内部会有个状态变量 pending_ids，它记录了当前已经被客户端读取的消息，但是还没有 ack。如果客户端没有 ack，这个变量里面的消息 ID 会越来越多，一旦某个消息被 ack，它就开始减少。这个 pending_ids 变量在 Redis 官方被称之为 PEL，也就是 Pending Entries List，这是一个很核心的数据结构，它用来确保客户端至少消费了消息一次，而不会在网络传输的中途丢失了没处理。

### 消息 ID

消息 ID 的形式是 timestampInMillis-sequence，例如 1527846880572-5，它表示当前的消息在毫米时间戳 1527846880572 时产生，并且是该毫秒内产生的第 5 条消息。消息 ID 可以由服务器自动生成，也可以由客户端自己指定，但是形式必须是整数-整数，而且必须是后面加入的消息的 ID 要大于前面的消息 ID。

### 消息内容

消息内容就是键值对，形如 hash 结构的键值对。

### 增删改查

1. xadd 追加消息
2. xdel 删除消息，这里的删除仅仅是设置了标志位，不影响消息总长度
3. xrange 获取消息列表，会自动过滤已经删除的消息
4. xlen 消息长度
5. del 删除 Stream

```redis
# * 号表示服务器自动生成 ID，后面顺序跟着一堆 key/value
#  名字叫 laoqian，年龄 30 岁
127.0.0.1:6379> xadd codehole * name laoqian age 30  
1527849609889-0  # 生成的消息 ID
...
127.0.0.1:6379> xlen codehole
(integer) 3
# -表示最小值 , + 表示最大值
127.0.0.1:6379> xrange codehole - +
127.0.0.1:6379> xrange codehole - +
1) 1) 1527849609889-0
   2) 1) "name"
      2) "laoqian"
      3) "age"
      4) "30"
2) 1) 1527849629172-0
   2) 1) "name"
      2) "xiaoyu"
      3) "age"
      4) "29"
3) 1) 1527849637634-0
   2) 1) "name"
      2) "xiaoqian"
      3) "age"
      4) "1"
# 指定最小消息 ID 的列表
127.0.0.1:6379> xrange codehole 1527849629172-0 +  
1) 1) 1527849629172-0
   2) 1) "name"
      2) "xiaoyu"
      3) "age"
      4) "29"
2) 1) 1527849637634-0
   2) 1) "name"
      2) "xiaoqian"
      3) "age"
      4) "1"
# 指定最大消息 ID 的列表
127.0.0.1:6379> xrange codehole - 1527849629172-0
1) 1) 1527849609889-0
   2) 1) "name"
      2) "laoqian"
      3) "age"
      4) "30"
2) 1) 1527849629172-0
   2) 1) "name"
      2) "xiaoyu"
      3) "age"
      4) "29"
127.0.0.1:6379> xdel codehole 1527849609889-0
(integer) 1
# 长度不受影响
127.0.0.1:6379> xlen codehole
(integer) 3
# 被删除的消息没了
127.0.0.1:6379> xrange codehole - +
1) 1) 1527849629172-0
   2) 1) "name"
      2) "xiaoyu"
      3) "age"
      4) "29"
2) 1) 1527849637634-0
   2) 1) "name"
      2) "xiaoqian"
      3) "age"
      4) "1"
# 删除整个 Stream
127.0.0.1:6379> del codehole
(integer) 1
```

### 独立消费

我们可以在不定义消费组的情况下进行 Stream 消息的独立消费，当 Stream 没有新消息时，甚至可以阻塞等待。Redis 设计了一个单独的消费指令 xread，可以将 Stream 当成普通的消息队列 (list) 来使用。使用 xread 时，我们可以完全忽略消费组 (Consumer Group) 的存在，就好比 Stream 就是一个普通的列表 (list)。

```redis
# 从 Stream 头部读取两条消息
127.0.0.1:6379> xread count 2 streams codehole 0-0
1) 1) "codehole"
   2) 1) 1) 1527851486781-0
         2) 1) "name"
            2) "laoqian"
            3) "age"
            4) "30"
      2) 1) 1527851493405-0
         2) 1) "name"
            2) "yurui"
            3) "age"
            4) "29"
# 从 Stream 尾部读取一条消息，毫无疑问，这里不会返回任何消息
127.0.0.1:6379> xread count 1 streams codehole $
(nil)
# 从尾部阻塞等待新消息到来，下面的指令会堵住，直到新消息到来
127.0.0.1:6379> xread block 0 count 1 streams codehole $
# 我们从新打开一个窗口，在这个窗口往 Stream 里塞消息
127.0.0.1:6379> xadd codehole * name youming age 60
1527852774092-0
# 再切换到前面的窗口，我们可以看到阻塞解除了，返回了新的消息内容
# 而且还显示了一个等待时间，这里我们等待了 93s
127.0.0.1:6379> xread block 0 count 1 streams codehole $
1) 1) "codehole"
   2) 1) 1) 1527852774092-0
         2) 1) "name"
            2) "youming"
            3) "age"
            4) "60"
(93.11s)
```

客户端如果想要使用 xread 进行顺序消费，一定要记住当前消费到哪里了，也就是返回的消息 ID。下次继续调用 xread 时，将上次返回的最后一个消息 ID 作为参数传递进去，就可以继续消费后续的消息。

block 0 表示永远阻塞，直到消息到来，block 1000 表示阻塞 1s，如果 1s 内没有任何消息到来，就返回 nil。

```redis
127.0.0.1:6379> xread block 1000 count 1 streams codehole $
(nil)
(1.07s)
```

### 创建消费组

![image](https://mail.wangkekai.cn/1614866113969.jpg)

Stream 通过 xgroup create 指令创建消费组 (Consumer Group)，需要传递起始消息 ID 参数用来初始化 last_delivered_id 变量。

```redis
#  表示从头开始消费
127.0.0.1:6379> xgroup create codehole cg1 0-0
OK
# $ 表示从尾部开始消费，只接受新消息，当前 Stream 消息会全部忽略
127.0.0.1:6379> xgroup create codehole cg2 $
OK
# 获取 Stream 信息
127.0.0.1:6379> xinfo stream codehole
 1) length
 2) (integer) 3  # 共 3 个消息
 3) radix-tree-keys
 4) (integer) 1
 5) radix-tree-nodes
 6) (integer) 2
 7) groups
 8) (integer) 2  # 两个消费组
 9) first-entry  # 第一个消息
10) 1) 1527851486781-0
    2) 1) "name"
       2) "laoqian"
       3) "age"
       4) "30"
11) last-entry  # 最后一个消息
12) 1) 1527851498956-0
    2) 1) "name"
       2) "xiaoqian"
       3) "age"
       4) "1"
# 获取 Stream 的消费组信息
127.0.0.1:6379> xinfo groups codehole
1) 1) name
   2) "cg1"
   3) consumers
   4) (integer) 0  # 该消费组还没有消费者
   5) pending
   6) (integer) 0  # 该消费组没有正在处理的消息
2) 1) name
   2) "cg2"
   3) consumers  # 该消费组还没有消费者
   4) (integer) 0
   5) pending
   6) (integer) 0  # 该消费组没有正在处理的消息
```

### 消费

Stream 提供了 xreadgroup 指令可以进行消费组的组内消费，需要提供消费组名称、消费者名称和起始消息 ID。它同 xread 一样，也可以阻塞等待新消息。读到新消息后，对应的消息 ID 就会进入消费者的 PEL(正在处理的消息) 结构里，客户端处理完毕后使用 xack 指令通知服务器，本条消息已经处理完毕，该消息 ID 就会从 PEL 中移除。

```redis
# > 号表示从当前消费组的 last_delivered_id 后面开始读
# 每当消费者读取一条消息，last_delivered_id 变量就会前进
127.0.0.1:6379> xreadgroup GROUP cg1 c1 count 1 streams codehole >
1) 1) "codehole"
   2) 1) 1) 1527851486781-0
         2) 1) "name"
            2) "laoqian"
            3) "age"
            4) "30"
127.0.0.1:6379> xreadgroup GROUP cg1 c1 count 1 streams codehole >
1) 1) "codehole"
   2) 1) 1) 1527851493405-0
         2) 1) "name"
            2) "yurui"
            3) "age"
            4) "29"
127.0.0.1:6379> xreadgroup GROUP cg1 c1 count 2 streams codehole >
1) 1) "codehole"
   2) 1) 1) 1527851498956-0
         2) 1) "name"
            2) "xiaoqian"
            3) "age"
            4) "1"
      2) 1) 1527852774092-0
         2) 1) "name"
            2) "youming"
            3) "age"
            4) "60"
# 再继续读取，就没有新消息了
127.0.0.1:6379> xreadgroup GROUP cg1 c1 count 1 streams codehole >
(nil)
# 那就阻塞等待吧
127.0.0.1:6379> xreadgroup GROUP cg1 c1 block 0 count 1 streams codehole >
# 开启另一个窗口，往里塞消息
127.0.0.1:6379> xadd codehole * name lanying age 61
1527854062442-0
# 回到前一个窗口，发现阻塞解除，收到新消息了
127.0.0.1:6379> xreadgroup GROUP cg1 c1 block 0 count 1 streams codehole >
1) 1) "codehole"
   2) 1) 1) 1527854062442-0
         2) 1) "name"
            2) "lanying"
            3) "age"
            4) "61"
(36.54s)
# 观察消费组信息
127.0.0.1:6379> xinfo groups codehole
1) 1) name
   2) "cg1"
   3) consumers
   4) (integer) 1  # 一个消费者
   5) pending
   6) (integer) 5  # 共 5 条正在处理的信息还有没有 ack
2) 1) name
   2) "cg2"
   3) consumers
   4) (integer) 0  # 消费组 cg2 没有任何变化，因为前面我们一直在操纵 cg1
   5) pending
   6) (integer) 0
# 如果同一个消费组有多个消费者，我们可以通过 xinfo consumers 指令观察每个消费者的状态
127.0.0.1:6379> xinfo consumers codehole cg1  # 目前还有 1 个消费者
1) 1) name
   2) "c1"
   3) pending
   4) (integer) 5  # 共 5 条待处理消息
   5) idle
   6) (integer) 418715  # 空闲了多长时间 ms 没有读取消息了
# 接下来我们 ack 一条消息
127.0.0.1:6379> xack codehole cg1 1527851486781-0
(integer) 1
127.0.0.1:6379> xinfo consumers codehole cg1
1) 1) name
   2) "c1"
   3) pending
   4) (integer) 4  # 变成了 5 条
   5) idle
   6) (integer) 668504
# 下面 ack 所有消息
127.0.0.1:6379> xack codehole cg1 1527851493405-0 1527851498956-0 1527852774092-0 1527854062442-0
(integer) 4
127.0.0.1:6379> xinfo consumers codehole cg1
1) 1) name
   2) "c1"
   3) pending
   4) (integer) 0  # pel 空了
   5) idle
   6) (integer) 745505
```

### Stream 消息太多怎么办?

它提供了一个定长 Stream 功能。在 xadd 的指令提供一个定长长度 maxlen，就可以将老的消息干掉，确保最多不超过指定长度。

```redis
127.0.0.1:6379> xlen codehole
(integer) 5
127.0.0.1:6379> xadd codehole maxlen 3 * name xiaorui age 1
1527855160273-0
127.0.0.1:6379> xlen codehole
(integer) 3
```

### 消息如果忘记 ACK 会怎样?

Stream 在每个消费者结构中保存了正在处理中的消息 ID 列表 PEL，如果消费者收到了消息处理完了但是没有回复 ack，就会导致 PEL 列表不断增长，如果有很多消费组的话，那么这个 PEL 占用的内存就会放大。

### PEL 如何避免消息丢失?

在客户端消费者读取 Stream 消息时，Redis 服务器将消息回复给客户端的过程中，客户端突然断开了连接，消息就丢失了。但是 PEL 里已经保存了发出去的消息 ID。待客户端重新连上之后，可以再次收到 PEL 中的消息 ID 列表。不过此时 xreadgroup 的起始消息 ID 不能为参数>，而必须是任意有效的消息 ID，一般将参数设为 0-0，表示读取所有的 PEL 消息以及自 last_delivered_id 之后的新消息。

### Stream 的高可用

Stream 的高可用是建立主从复制基础上的，它和其它数据结构的复制机制没有区别，也就是说在 Sentinel 和 Cluster 集群环境下 Stream 是可以支持高可用的。不过鉴于 Redis 的指令复制是异步的，在 failover 发生时，Redis 可能会丢失极小部分数据，这点 Redis 的其它数据结构也是一样的。

### 分区 Partition

Redis 的服务器没有原生支持分区能力，如果想要使用分区，那就需要分配多个 Stream，然后在客户端使用一定的策略来生产消息到不同的 Stream。你也许会认为 Kafka 要先进很多，它是原生支持 Partition 的。关于这一点，我并不认同。记得 Kafka 的客户端也存在 HashStrategy 么，因为它也是通过客户端的 hash 算法来将不同的消息塞入不同分区的。

## 26. 无所不知 —— Info 指令

Info 指令显示的信息非常繁多，分为 9 大块，每个块都有非常多的参数，这 9 个块分别是:

1. Server 服务器运行的环境参数
2. Clients 客户端相关信息
3. Memory 服务器运行内存统计数据
4. Persistence 持久化信息
5. Stats 通用统计数据
6. Replication 主从复制相关信息
7. CPU CPU 使用情况
8. Cluster 集群信息
9. KeySpace 键值对统计数量信息

Info 可以一次性获取所有的信息，也可以按块取信息。

```redis
# 获取所有信息
> info
# 获取内存相关信息
> info memory
# 获取复制相关信息
> info replication
```

### Redis 每秒执行多少次指令？

```redis
# ops_per_sec: operations per second，也就是每秒操作数
> redis-cli info stats |grep ops
instantaneous_ops_per_sec:789
```

以上，表示 ops 是 789，也就是所有客户端每秒会发送 789 条指令到服务器执行。极限情况下，Redis 可以每秒执行 10w 次指令，CPU 几乎完全榨干。**如果 qps 过高，可以考虑通过 monitor 指令快速观察一下究竟是哪些 key 访问比较频繁，从而在相应的业务上进行优化，以减少 IO 次数。**monitor 指令会瞬间吐出来巨量的指令文本，所以一般在执行 monitor 后立即 ctrl+c中断输出。

### Redis 连接了多少客户端？

```redis
> redis-cli info clients
# Clients
connected_clients:124  # 这个就是正在连接的客户端数量
client_longest_output_list:0
client_biggest_input_buf:0
blocked_clients:0
```

这个信息也是比较有用的，通过观察这个数量可以确定是否存在意料之外的连接。如果发现这个数量不对劲，接着就可以使用 client list 指令列出所有的客户端链接地址来确定源头。

关于客户端的数量还有个重要的参数需要观察，那就是 rejected_connections，它表示因为超出最大连接数限制而被拒绝的客户端连接次数，如果这个数字很大，意味着服务器的最大连接数设置的过低需要调整 maxclients 参数。

```redis
> redis-cli info stats |grep reject
rejected_connections:0
```

### Redis 内存占用多大 ?

```redis
> redis-cli info memory | grep used | grep human
used_memory_human:827.46K # 内存分配器 (jemalloc) 从操作系统分配的内存总量
used_memory_rss_human:3.61M  # 操作系统看到的内存占用 ,top 命令看到的内存
used_memory_peak_human:829.41K  # Redis 内存消耗的峰值
used_memory_lua_human:37.00K # lua 脚本引擎占用的内存大小
```

### 复制积压缓冲区多大？

```redis
> redis-cli info replication |grep backlog
repl_backlog_active:0
repl_backlog_size:1048576  # 这个就是积压缓冲区大小
repl_backlog_first_byte_offset:0
repl_backlog_histlen:0
```

<font color=red>复制积压缓冲区大小非常重要，它严重影响到主从复制的效率。</font>当从库因为网络原因临时断开了主库的复制，然后网络恢复了，又重新连上的时候，这段断开的时间内发生在 master 上的修改操作指令都会放在积压缓冲区中，这样从库可以通过积压缓冲区恢复中断的主从同步过程。

积压缓冲区是环形的，后来的指令会覆盖掉前面的内容。如果从库断开的时间过长，或者缓冲区的大小设置的太小，都会导致从库无法快速恢复中断的主从同步过程，因为中间的修改指令被覆盖掉了。这时候从库就会进行全量同步模式，非常耗费 CPU 和网络资源。

如果有多个从库复制，积压缓冲区是共享的，它不会因为从库过多而线性增长。如果实例的修改指令请求很频繁，那就把积压缓冲区调大一些，几十个 M 大小差不多了，如果很闲，那就设置为几个 M。

```redis
> redis-cli info stats | grep sync
sync_full:0
sync_partial_ok:0
sync_partial_err:0  # 半同步失败次数
```

**通过查看 sync_partial_err 变量的次数来决定是否需要扩大积压缓冲区，它表示主从半同步复制失败的次数。**

## 27. 优胜劣汰 —— LRU

当 Redis 内存超出物理内存限制时，内存的数据会开始和磁盘产生频繁的交换 (swap)。交换会让 Redis 的性能急剧下降，对于访问量比较频繁的 Redis 来说，这样龟速的存取效率基本上等于不可用。  
为了限制最大使用内存，Redis 提供了配置参数 maxmemory 来限制内存超出期望大小。

1. **noeviction** 不删除策略 - 不会继续服务写请求 (DEL 请求可以继续服务)，读请求可以继续进行。这样可以保证不会丢失数据，但是会让线上的业务不能持续进行。这是默认的淘汰策略。
2. **volatile-lru** 尝试淘汰设置了过期时间的 key，最少使用的 key 优先被淘汰。没有设置过期时间的 key 不会被淘汰，这样可以保证需要持久化的数据不会突然丢失。
3. **volatile-ttl** 除了淘汰的策略不是 LRU，而是 key 的剩余寿命 ttl 的值，ttl 越小越优先被淘汰。
4. **volatile-random** 淘汰的 key 是过期 key 集合中随机的 key。
5. **allkeys-lru** 区别于 volatile-lru，这个策略要淘汰的 key 对象是全体的 key 集合，而不只是过期的 key 集合。这意味着没有设置过期时间的 key 也会被淘汰。
6. **allkeys-random** 跟上面一样，不过淘汰的策略是随机的 key。

### LRU 算法

实现 LRU 算法除了需要 key/value 字典外，还需要附加一个链表，链表中的元素按照一定的顺序进行排列。当空间满的时候，会踢掉链表尾部的元素。当字典的某个元素被访问时，它在链表中的位置会被移动到表头。所以链表的元素排列顺序就是元素最近被访问的时间顺序。

> 位于链表尾部的元素就是不被重用的元素，所以会被踢掉。位于表头的元素就是最近刚刚被人用过的元素，所以暂时不会被踢。

```python
from collections import OrderedDict

class LRUDict(OrderedDict):

    def __init__(self, capacity):
        self.capacity = capacity
        self.items = OrderedDict()

    def __setitem__(self, key, value):
        old_value = self.items.get(key)
        if old_value is not None:
            self.items.pop(key)
            self.items[key] = value
        elif len(self.items) < self.capacity:
            self.items[key] = value
        else:
            self.items.popitem(last=True)
            self.items[key] = value

    def __getitem__(self, key):
        value = self.items.get(key)
        if value is not None:
            self.items.pop(key)
            self.items[key] = value
        return value

    def __repr__(self):
        return repr(self.items)


d = LRUDict(10)

for i in range(15):
    d[i] = i
print d
```

### 近似 LRU 算法

Redis 使用的是一种近似 LRU 算法，它跟 LRU 算法还不太一样。之所以不使用 LRU 算法，是因为需要消耗大量的额外的内存，需要对现有的数据结构进行较大的改造。<u>近似 LRU 算法则很简单，在现有数据结构的基础上使用随机采样法来淘汰元素，能达到和 LRU 算法非常近似的效果。</u>Redis 为实现近似 LRU 算法，它给每个 key 增加了一个额外的小字段，这个字段的长度是 24 个 bit，也就是最后一次被访问的时间戳。

处理 key 过期方式分为集中处理和懒惰处理，LRU 淘汰不一样，它的处理方式只有懒惰处理。  
当 Redis 执行写操作时，发现内存超出 maxmemory，就会执行一次 LRU 淘汰算法。这个算法也很简单，就是随机采样出 5(可以配置) 个 key，然后淘汰掉最旧的 key，如果淘汰后内存还是超出 maxmemory，那就继续随机采样淘汰，直到内存低于 maxmemory 为止。

如何采样就是看 maxmemory-policy 的配置，如果是 allkeys 就是从所有的 key 字典中随机，如果是 volatile 就从带过期时间的 key 字典中随机。每次采样多少个 key 看的是 maxmemory_samples 的配置，默认为 5。
![image](https://mail.wangkekai.cn/A0A1AF6D-B708-4261-BCB5-69973CDE8FFF.png)

绿色部分是新加入的 key，深灰色部分是老旧的 key，浅灰色部分是通过 LRU 算法淘汰掉的 key。从图中可以看出采样数量越大，近似 LRU 算法的效果越接近严格 LRU 算法。同时 Redis3.0 在算法中增加了淘汰池，进一步提升了近似 LRU 算法的效果。  
淘汰池是一个数组，它的大小是 maxmemory_samples，在每一次淘汰循环中，新随机出来的 key 列表会和淘汰池中的 key 列表进行融合，淘汰掉最旧的一个 key 之后，保留剩余较旧的 key 列表放入淘汰池中留待下一个循环。

## 28. 平波缓进 —— 懒惰删除

### Redis 为什么要懒惰删除(lazy free)？

删除指令 del 会直接释放对象的内存，大部分情况下，这个指令非常快，没有明显延迟。不过如果删除的 key 是一个非常大的对象，比如一个包含了千万元素的 hash，那么删除操作就会导致单线程卡顿。

Redis 为了解决这个卡顿问题，在 4.0 版本引入了 unlink 指令，**它能对删除操作进行懒处理，丢给后台线程来异步回收内存。**

```redis
> unlink key
OK
```

如果担心这里的线程安全问题？  
可以将整个 Redis 内存里面所有有效的数据想象成一棵大树。当 unlink 指令发出时，它只是把大树中的一个树枝别断了，然后扔到旁边的火堆里焚烧 (异步线程池)。树枝离开大树的一瞬间，它就再也无法被主线程中的其它指令访问到了，因为主线程只会沿着这颗大树来访问。

### flush

Redis 提供了 flushdb 和 flushall 指令，用来清空数据库，这也是极其缓慢的操作。Redis 4.0 同样给这两个指令也带来了异步化，在指令后面增加 async 参数就可以将整棵大树连根拔起，扔给后台线程慢慢焚烧。

```redis
> flushall async
OK
```

### 异步队列

主线程将对象的引用从「大树」中摘除后，会将这个 key 的内存回收操作包装成一个任务，塞进异步任务队列，后台线程会从这个异步队列中取任务。任务队列被主线程和异步线程同时操作，所以必须是一个线程安全的队列。
![image](https://mail.wangkekai.cn/7428EC70-56B7-46E4-8A16-6F5683DD751A.png)

不是所有的 unlink 操作都会延后处理，如果对应 key 所占用的内存很小，延后处理就没有必要了，这时候 Redis 会将对应的 key 内存立即回收，跟 del 指令一样。

### AOF Sync也很慢

Redis需要每秒一次(可配置)同步AOF日志到磁盘，确保消息尽量不丢失，需要调用sync函数，这个操作会比较耗时，会导致主线程的效率下降，所以Redis也将这个操作移到异步线程来完成。执行AOF Sync操作的线程是一个独立的异步线程，和前面的懒惰删除线程不是一个线程，同样它也有一个属于自己的任务队列，队列里只用来存放AOF Sync任务。

### 更多异步删除点

Redis 回收内存除了 del 指令和 flush 之外，还会存在于在 key 的过期、LRU 淘汰、rename 指令以及从库全量同步时接受完 rdb 文件后会立即进行的 flush 操作。

Redis4.0 为这些删除点也带来了异步删除机制，打开这些点需要额外的配置选项。

1. slave-lazy-flush 从库接受完 rdb 文件后的 flush 操作
2. lazyfree-lazy-eviction 内存达到 maxmemory 时进行淘汰
3. lazyfree-lazy-expire key 过期删除
4. lazyfree-lazy-server-del rename 指令删除 destKey

## 29. 居安思危 —— 保护 Redis

### 指令安全

Redis 有一些非常危险的指令，这些指令会对 Redis 的稳定以及数据安全造成非常严重的影响。<font color=red>比如 keys 指令会导致 Redis 卡顿，flushdb 和 flushall 会让 Redis 的所有数据全部清空。</font>如何避免人为操作失误导致这些灾难性的后果也是运维人员特别需要注意的风险点之一。

Redis 在配置文件中提供了 rename-command 指令用于将某些危险的指令修改成特别的名称，用来避免人为误操作。比如在配置文件的 security 块增加下面的内容:

```redis
rename-command keys abckeysabc
```

如果还想执行 keys 方法，那就不能直接敲 keys 命令了，而需要键入 abckeysabc。 如果想完全封杀某条指令，可以将指令 rename 成空串，就无法通过任何字符串指令来执行这条指令了。

```redis
rename-command flushall ""
```

### 端口安全

Redis 默认会监听 *:6379，如果当前的服务器主机有外网地址，Redis 的服务将会直接暴露在公网上，任何一个初级黑客使用适当的工具对 IP 地址进行端口扫描就可以探测出来。

```redis
bind 10.100.20.13
```

<font color=red>运维人员务必在 Redis 的配置文件中指定监听的 IP 地址，避免这样的惨剧发生。</font>更进一步，还可以增加 Redis 的密码访问限制，客户端必须使用 auth 指令传入正确的密码才可以访问 Redis，这样即使地址暴露出去了，普通黑客也无法对 Redis 进行任何指令操作。

```redis
requirepass yoursecurepasswordhereplease
```

密码控制也会影响到从库复制，从库必须在配置文件里使用 masterauth 指令配置相应的密码才可以进行复制操作。

```redis
masterauth yoursecurepasswordhereplease
```

### Lua 脚本安全

开发者必须禁止 Lua 脚本由用户输入的内容 (UGC) 生成，这可能会被黑客利用以植入恶意的攻击代码来得到 Redis 的主机权限。  
同时，我们应该让 Redis 以普通用户的身份启动，这样即使存在恶意代码黑客也无法拿到 root 权限。

### SSL 代理

SSL 代理比较常见的有 ssh，不过 Redis 官方推荐使用 spiped 工具，可能是因为 spiped 的功能相对比较单一，使用也比较简单，易于理解。下面这张图是使用 spiped 对 ssh 通道进行二次加密 (因为 ssh 通道也可能存在 bug)。
![image](https://mail.wangkekai.cn/605CF567-B0EF-4D23-9A8C-37052FCD4A6B.png)

同样 SSL 代理也可以用在主从复制上，如果 Redis 主从实例需要跨机房复制，spiped 也可以派上用场。

## 30. 隔墙有耳 —— Redis 安全通信

### spiped 原理

spiped 会在客户端和服务器各启动一个 spiped 进程。
![image](https://mail.wangkekai.cn/5DDEE410-C225-4425-AA8B-6D72181995BE.png)

左边的 spiped 进程 A 负责接受来自 Redis Client 发送过来的请求数据，加密后传送到右边的 spiped 进程 B。spiped B 将接收到的数据解密后传递到 Redis Server。然后 Redis Server 再走一个反向的流程将响应回复给 Redis Client。

每一个 spiped 进程都会有一个监听端口 (server socket) 用来接收数据，同时还会作为一个客户端 (socket client) 将数据转发到目标地址。

spiped 进程需要成对出现，相互之间需要使用相同的共享密钥来加密消息。

### spiped 使用入门

```shell
> brew install spiped
```

1. 使用 Docker 启动 redis-server，注意要绑定本机的回环127.0.0.1；

```shell
> docker run -d -p127.0.0.1:6379:6379 --name redis-server-6379 redis
12781661ec47faa8a8a967234365192f4da58070b791262afb8d9f64fce61835
> docker ps
CONTAINER ID        IMAGE               COMMAND                  CREATED                  STATUS              PORTS                      NAMES
12781661ec47        redis               "docker-entrypoint.s…"   Less than a second ago   Up 1 second         127.0.0.1:6379->6379/tcp   redis-server-6379
```

2. 生成随机的密钥文件

```shell
# 随机的 32 个字节
> dd if=/dev/urandom bs=32 count=1 of=spiped.key
1+0 records in
1+0 records out
32 bytes transferred in 0.000079 secs (405492 bytes/sec)
> ls -l
rw-r--r--  1 qianwp  staff  32  7 24 18:13 spiped.key
```

3. 使用密钥文件启动服务器 spiped 进程，172.16.128.81是我本机的公网 IP 地址；

```shell
# -d 表示 decrypt(对输入数据进行解密)，-s 为源监听地址，-t 为转发目标地址
> spiped -d -s '[172.16.128.81]:6479' -t '[127.0.0.1]:6379' -k spiped.key
> ps -ef|grep spiped
501 30673     1   0  7:29 下午 ??         0:00.04 spiped -d -s [172.16.128.81]:6479 -t [127.0.0.1]:6379 -k spiped.key
```

这个 spiped 进程监听公网 IP 的 6479 端口接收公网上的数据，将数据解密后转发到本机回环地址的 6379 端口，也就是 redis-server 监听的端口。

4. 使用密钥文件启动客户端 spiped 进程，172.16.128.81是我本机的公网 IP 地址

```shell
# -e 表示 encrypt，对输入数据进行加密
> spiped -e -s '[127.0.0.1]:6579' -t '[172.16.128.81]:6479' -k spiped.key
> ps -ef|grep spiped
501 30673     1   0  7:29 下午 ??         0:00.04 spiped -d -s [172.16.128.81]:6479 -t [127.0.0.1]:6379 -k spiped.key
501 30696     1   0  7:30 下午 ??         0:00.03 spiped -e -s [127.0.0.1]:6579 -t [172.16.128.81]:6479 -k spiped.key
```

客户端 spiped 进程监听了本地回环地址的 6579 端口，将该端口上收到的数据加密转发到服务器 spiped 进程。

5. 启动客户端链接，因为 Docker 里面的客户端不好访问宿主机的回环地址，所以 Redis 的客户端我们使用 Python 代码来启动；

```shell
>> import redis
>> c=redis.StrictRedis(host="localhost", port=6579)
>> c.ping()
>> c.info('cpu')
{'used_cpu_sys': 4.83,
 'used_cpu_sys_children': 0.0,
 'used_cpu_user': 0.93,
 'used_cpu_user_children': 0.0}
```

可以看出客户端和服务器已经通了，如果我们尝试直接链接服务器 spiped 进程 (加密的端口 6379)，看看会发生什么。

```shell
>>> import redis
>>> c=redis.StrictRedis(host="172.16.128.81", port=6479)
>>> c.ping()
Traceback (most recent call last):
  File "<stdin>", line 1, in <module>
  File "/Users/qianwp/source/animate/juejin-redis/.py/lib/python2.7/site-packages/redis/client.py", line 777, in ping
    return self.execute_command('PING')
  File "/Users/qianwp/source/animate/juejin-redis/.py/lib/python2.7/site-packages/redis/client.py", line 674, in execute_command
    return self.parse_response(connection, command_name, **options)
  File "/Users/qianwp/source/animate/juejin-redis/.py/lib/python2.7/site-packages/redis/client.py", line 680, in parse_response
    response = connection.read_response()
  File "/Users/qianwp/source/animate/juejin-redis/.py/lib/python2.7/site-packages/redis/connection.py", line 624, in read_response
    response = self._parser.read_response()
  File "/Users/qianwp/source/animate/juejin-redis/.py/lib/python2.7/site-packages/redis/connection.py", line 284, in read_response
    response = self._buffer.readline()
  File "/Users/qianwp/source/animate/juejin-redis/.py/lib/python2.7/site-packages/redis/connection.py", line 216, in readline
    self._read_from_socket()
  File "/Users/qianwp/source/animate/juejin-redis/.py/lib/python2.7/site-packages/redis/connection.py", line 191, in _read_from_socket
    (e.args,))
redis.exceptions.ConnectionError: Error while reading from socket: ('Connection closed by server.',)
```

从输出中可以看出来请求是发送过去了，但是却出现了读超时，要么是服务器在默认的超时时间内没有返回数据，要么是服务器没有返回客户端想要的数据。

spiped 可以同时支持多个客户端链接的数据转发工作，它还可以通过参数来限定允许的最大客户端连接数。但是对于服务器 spiped，它不能同时支持多个服务器之间的转发。意味着在集群环境下，需要为每一个 server 节点启动一个 spiped 进程来代收消息，在运维实践上这可能会比较繁琐。

## 31. 法力无边 Redis lua 脚本执行原理

可以向服务器发送 lua 脚本来执行自定义动作，获取脚本的响应数据。Redis 服务器会单线程原子性执行 lua 脚本，保证 lua 脚本在处理的过程中不会被任意其它请求打断。
![image](https://mail.wangkekai.cn/2599F979-1B21-48AD-9266-C5052F336A97.png)

比如在分布式锁小节，我们提到了 del_if_equals 伪指令，它可以将匹配 key 和删除 key 合并在一起原子性执行，Redis 原生没有提供这样功能的指令，它可以使用 lua 脚本来完成。

```lua
if redis.call("get",KEYS[1]) == ARGV[1] then
    return redis.call("del",KEYS[1])
else
    return 0
end
```

那上面这个脚本可以使用 EVAL 指令执行

```redis
127.0.0.1:6379> set foo bar
OK
127.0.0.1:6379> eval 'if redis.call("get",KEYS[1]) == ARGV[1] then return redis.call("del",KEYS[1]) else return 0 end' 1 foo bar
(integer) 1
127.0.0.1:6379> eval 'if redis.call("get",KEYS[1]) == ARGV[1] then return redis.call("del",KEYS[1]) else return 0 end' 1 foo bar
(integer) 0
```

EVAL 指令的第一个参数是脚本内容字符串，上面的例子我们将 lua 脚本压缩成一行以单引号围起来是为了方便命令行执行。然后是 key 的数量以及每个 key 串，最后是一系列附加参数字符串。附加参数的数量不需要和 key 保持一致，可以完全没有附加参数。

```redis
EVAL SCRIPT KEY_NUM KEY1 KEY2 ... KEYN ARG1 ARG2 ....
```

上面的例子中只有 1 个 key，它就是 foo，紧接着 bar 是唯一的附加参数。在 lua 脚本中，数组下标是从 1 开始，所以通过 KEYS[1] 就可以得到 第一个 key，通过 ARGV[1] 就可以得到第一个附加参数。redis.call 函数可以让我们调用 Redis 的原生指令，上面的代码分别调用了 get 指令和 del 指令。return 返回的结果将会返回给客户端。

### SCRIPT LOAD 和 EVALSHA 指令

如果脚本的内容很长，而且客户端需要频繁执行，那么每次都需要传递冗长的脚本内容势必比较浪费网络流量。<u>所以 Redis 还提供了 SCRIPT LOAD 和 EVALSHA 指令来解决这个问题。</u>

![image](https://mail.wangkekai.cn/4D5FC590-6CE3-4F5D-A0AE-694E2E9B8D93.png)

SCRIPT LOAD 指令用于将客户端提供的 lua 脚本传递到服务器而不执行，但是会得到脚本的唯一 ID，这个唯一 ID 是用来唯一标识服务器缓存的这段 lua 脚本，它是由 Redis 使用 sha1 算法揉捏脚本内容而得到的一个很长的字符串。有了这个唯一 ID，后面客户端就可以通过 EVALSHA 指令反复执行这个脚本了。 我们知道 Redis 有 incrby 指令可以完成整数的自增操作，但是没有提供自乘这样的指令。

```shell
incrby key value  ==> $key = $key + value
mulby key value ==> $key = $key * value
```

下面我们使用 SCRIPT LOAD 和 EVALSHA 指令来完成自乘运算。

```lua
local curVal = redis.call("get", KEYS[1])
if curVal == false then
  curVal = 0
else
  curVal = tonumber(curVal)
end
curVal = curVal * tonumber(ARGV[1])
redis.call("set", KEYS[1], curVal)
return curVal

# 将上面的语句单行化
local curVal = redis.call("get", KEYS[1]); if curVal == false then curVal = 0 else curVal = tonumber(curVal) end; curVal = curVal * tonumber(ARGV[1]); redis.call("set", KEYS[1], curVal); return curVal
```

加载脚本

```redis
127.0.0.1:6379> script load 'local curVal = redis.call("get", KEYS[1]); if curVal == false then curVal = 0 else curVal = tonumber(curVal) end; curVal = curVal * tonumber(ARGV[1]); redis.call("set", KEYS[1], curVal); return curVal'
"be4f93d8a5379e5e5b768a74e77c8a4eb0434441"
```

命令行输出了很长的字符串，它就是脚本的唯一标识，下面我们使用这个唯一标识来执行指令

```redis
127.0.0.1:6379> evalsha be4f93d8a5379e5e5b768a74e77c8a4eb0434441 1 notexistskey 5
(integer) 0
127.0.0.1:6379> evalsha be4f93d8a5379e5e5b768a74e77c8a4eb0434441 1 notexistskey 5
(integer) 0
127.0.0.1:6379> set foo 1
OK
127.0.0.1:6379> evalsha be4f93d8a5379e5e5b768a74e77c8a4eb0434441 1 foo 5
(integer) 5
127.0.0.1:6379> evalsha be4f93d8a5379e5e5b768a74e77c8a4eb0434441 1 foo 5
(integer) 25
```

### 错误处理

上面的脚本参数要求传入的附加参数必须是整数，如果没有传递整数会怎样呢？

```redis
127.0.0.1:6379> evalsha be4f93d8a5379e5e5b768a74e77c8a4eb0434441 1 foo bar
(error) ERR Error running script (call to f_be4f93d8a5379e5e5b768a74e77c8a4eb0434441): @user_script:1: user_script:1: attempt to perform arithmetic on a nil value
```

可以看到客户端输出了服务器返回的通用错误消息，注意这是一个动态抛出的异常，Redis 会保护主线程不会因为脚本的错误而导致服务器崩溃，近似于在脚本的外围有一个很大的 try catch 语句包裹。在 lua 脚本执行的过程中遇到了错误，同 redis 的事务一样，那些通过 redis.call 函数已经执行过的指令对服务器状态产生影响是无法撤销的，在编写 lua 代码时一定要小心，避免没有考虑到的判断条件导致脚本没有完全执行。

![image](https://mail.wangkekai.cn/4192BA00-6088-44BE-A2A6-2B6DE1974124.png)

如果读者对 lua 语言有所了解就知道 lua 原生没有提供 try catch 语句，那上面提到的异常包裹语句究竟是用什么来实现的呢？lua 的替代方案是内置了 pcall(f) 函数调用。pcall 的意思是 protected call，它会让 f 函数运行在保护模式下，f 如果出现了错误，pcall 调用会返回 false 和错误信息。而普通的 call(f) 调用在遇到错误时只会向上抛出异常。在 Redis 的源码中可以看到 lua 脚本的执行被包裹在 pcall 函数调用中。

Redis 在 lua 脚本中除了提供了 redis.call 函数外，同样也提供了 redis.pcall 函数。前者遇到错误向上抛出异常，后者会返回错误信息。使用时一定要注意 call 函数出错时会中断脚本的执行，为了保证脚本的原子性，要谨慎使用。

### 错误传递

redis.call 函数调用会产生错误，脚本遇到这种错误会返回怎样的信息呢？

```redis
127.0.0.1:6379> hset foo x 1 y 2
(integer) 2
127.0.0.1:6379> eval 'return redis.call("incr", "foo")' 0
(error) ERR Error running script (call to f_8727c9c34a61783916ca488b366c475cb3a446cc): @user_script:1: WRONGTYPE Operation against a key holding the wrong kind of value
```

客户端输出的依然是一个通用的错误消息，而不是 incr 调用本应该返回的 WRONGTYPE 类型的错误消息。Redis 内部在处理 redis.call 遇到错误时是向上抛出异常，外围的用户看不见的 pcall调用捕获到脚本异常时会向客户端回复通用的错误信息。如果我们将上面的 call 改成 pcall，结果就会不一样，它可以将内部指令返回的特定错误向上传递。

```redis
127.0.0.1:6379> eval 'return redis.pcall("incr", "foo")' 0
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

### 脚本死循环怎么办？

Redis 的指令执行是个单线程，这个单线程还要执行来自客户端的 lua 脚本。如果 lua 脚本中来一个死循环，是不是 Redis 就完蛋了？Redis 为了解决这个问题，<u>它提供了 script kill 指令用于动态杀死一个执行时间超时的 lua 脚本。</u>不过 script kill 的执行有一个重要的前提，那就是当前正在执行的脚本没有对 Redis 的内部数据状态进行修改，因为 Redis 不允许 script kill 破坏脚本执行的原子性。比如脚本内部使用了 redis.call("set", key, value) 修改了内部的数据，那么 script kill 执行时服务器会返回错误。下面我们来尝试以下 script kill 指令。

```redis
127.0.0.1:6379> eval 'while(true) do print("hello") end' 0
```

eval 指令执行后，可以明显看出来 redis 卡死了，死活没有任何响应，如果去观察 Redis 服务器日志可以看到日志在疯狂输出 hello 字符串。这时候就必须重新开启一个 redis-cli 来执行 script kill 指令。

```redis
127.0.0.1:6379> script kill
OK
(2.58s)
```

再回过头看 eval 指令的输出

```redis
127.0.0.1:6379> eval 'while(true) do print("hello") end' 0
(error) ERR Error running script (call to f_d395649372f578b1a0d3a1dc1b2389717cadf403): @user_script:1: Script killed by user with SCRIPT KILL...
(6.99s)
```

有几个疑点：  
第一个是 script kill 指令为什么执行了 2.58 秒  
第二个是脚本都卡死了，Redis 哪里来的闲功夫接受 script kill 指令。  
如果你自己尝试了在第二个窗口执行 redis-cli 去连接服务器，你还会发现第三个疑点，redis-cli 建立连接有点慢，大约顿了有 1 秒左右。

### Script Kill 的原理

下面我就要开始揭秘 kill 的原理了，lua 脚本引擎功能太强大了，它提供了各式各样的钩子函数，它允许在内部虚拟机执行指令时运行钩子代码。比如每执行 N 条指令执行一次某个钩子函数，Redis 正是使用了这个钩子函数。

![image](https://mail.wangkekai.cn/319C7609-AD91-4DA8-9E9E-8C7E38F518FC.png)

Redis 在钩子函数里会忙里偷闲去处理客户端的请求，并且只有在发现 lua 脚本执行超时之后才会去处理请求，这个超时时间默认是 5 秒。于是上面提出的三个疑点也就烟消云散了。

## 32. 短小精悍 - 命令行工具的妙用

### 执行单条命令

平时在访问 Redis 服务器，一般都会使用 redis-cli 进入交互模式，然后一问一答来读写服务器，这种情况下我们使用的是它的「交互模式」。还有另外一种「直接模式」，通过将命令参数直接传递给 redis-cli 来执行指令并获取输出结果。

```shell
$ redis-cli incrby foo 5
(integer) 5
$ redis-cli incrby foo 5
(integer) 10
```

如果输出的内容较大，还可以将输出重定向到外部文件

```shell
$ redis-cli info > info.txt
$ wc -l info.txt
     120 info.txt
```

上面的命令指向的服务器是默认服务器地址，如果想指向特定的服务器可以这样

```shell
// -n 2 表示使用第2个库，相当于 select 2
$ redis-cli -h localhost -p 6379 -n 2 ping
PONG
```

### 批量执行命令

在平时线上的开发过程中，有时候我们免不了要手工造数据，然后导入 Redis。通常我们会编写脚本程序来做这件事。不过还有另外一种比较便捷的方式，那就是直接使用 redis-cli 来批量执行一系列指令。

```shell
$ cat cmds.txt
set foo1 bar1
set foo2 bar2
set foo3 bar3
......
$ cat cmds.txt | redis-cli
OK
OK
OK
...
```

上面的指令使用了 Unix 管道将 cat 指令的标准输出连接到 redis-cli 的标准输入。其实还可以直接使用输入重定向来批量执行指令。

```shell
$ redis-cli < cmds.txt
OK

...
```

### set 多行字符串

如果一个字符串有多行，你希望将它传入 set 指令，redis-cli 要如何做？<u>可以使用 -x 选项，该选项会使用标准输入的内容作为最后一个参数。</u>

```shell
$ cat str.txt
Ernest Hemingway once wrote,
"The world is a fine place and worth fighting for."
I agree with the second part.
$ redis-cli -x set foo < str.txt
OK
$ redis-cli get foo
"Ernest Hemingway once wrote,\n\"The world is a fine place and worth fighting for.\"\nI agree with the second part.\n"
```

### 重复执行指令

redis-cli 还支持重复执行指令多次，每条指令执行之间设置一个间隔时间，如此便可以观察某条指令的输出内容随时间变化。

```shell
// 间隔1s，执行5次，观察qps的变化
$ redis-cli -r 5 -i 1 info | grep ops
instantaneous_ops_per_sec:43469
instantaneous_ops_per_sec:47460
instantaneous_ops_per_sec:47699
instantaneous_ops_per_sec:46434
instantaneous_ops_per_sec:47216
```

如果将次数设置为 -1 那就是重复无数次永远执行下去。如果不提供 -i 参数，那就没有间隔，连续重复执行。在交互模式下也可以重复执行指令，形式上比较怪异，在指令前面增加次数

```redis
127.0.0.1:6379> 5 ping
PONG
PONG
PONG
PONG
PONG
# 下面的指令很可怕，你的屏幕要愤怒了
127.0.0.1:6379> 10000 info
.......
```

### 导出 csv

redis-cli 不能一次导出整个库的内容为 csv，但是可以导出单条指令的输出为 csv 格式。

```shell
$ redis-cli rpush lfoo a b c d e f g
(integer) 7
$ redis-cli --csv lrange lfoo 0 -1
"a","b","c","d","e","f","g"
$ redis-cli hmset hfoo a 1 b 2 c 3 d 4
OK
$ redis-cli --csv hgetall hfoo
"a","1","b","2","c","3","d","4"
```

当然这种导出功能比较弱，仅仅是一堆字符串用逗号分割开来。不过你可以结合命令的批量执行来看看多个指令的导出效果。

```redis
$ redis-cli --csv -r 5 hgetall hfoo
"a","1","b","2","c","3","d","4"
"a","1","b","2","c","3","d","4"
"a","1","b","2","c","3","d","4"
"a","1","b","2","c","3","d","4"
"a","1","b","2","c","3","d","4"
```

### 执行 lua 脚本

在 lua 脚本小节，我们使用 eval 指令来执行脚本字符串，每次都是将脚本内容压缩成单行字符串再调用 eval 指令，这非常繁琐，而且可读性很差。redis-cli 考虑到了这点，它可以直接执行脚本文件。

```redis
127.0.0.1:6379> eval "return redis.pcall('mset', KEYS[1], ARGV[1], KEYS[2], ARGV[2])" 2 foo1 foo2 bar1 bar2
OK
127.0.0.1:6379> eval "return redis.pcall('mget', KEYS[1], KEYS[2])" 2 foo1 foo2
1) "bar1"
2) "bar2"
```

下面我们以脚本的形式来执行上面的指令，参数形式有所不同，KEY 和 ARGV 之间需要使用逗号分割，并且不需要提供 KEY 的数量参数

```shell
$ cat mset.txt
return redis.pcall('mset', KEYS[1], ARGV[1], KEYS[2], ARGV[2])
$ cat mget.txt
return redis.pcall('mget', KEYS[1], KEYS[2])
$ redis-cli --eval mset.txt foo1 foo2 , bar1 bar2
OK
$ redis-cli --eval mget.txt foo1 foo2
1) "bar1"
2) "bar2"
```

如果你的 lua 脚本太长，--eval 将大有用处。

### 监控服务器状态

我们可以使用 --stat 参数来实时监控服务器的状态，间隔 1s 实时输出一次。

```redis
$ redis-cli --stat
------- data ------ --------------------- load -------------------- - child -
keys       mem      clients blocked requests            connections
2          6.66M    100     0       11591628 (+0)       335
2          6.66M    100     0       11653169 (+61541)   335
2          6.66M    100     0       11706550 (+53381)   335
2          6.54M    100     0       11758831 (+52281)   335
2          6.66M    100     0       11803132 (+44301)   335
2          6.66M    100     0       11854183 (+51051)   335
```

如果你觉得间隔太长或是太短，可以使用 -i 参数调整输出间隔。

### 扫描大 KEY

这个功能太实用了，我已经在线上试过无数次了。每次遇到 Redis 偶然卡顿问题，第一个想到的就是实例中是否存在大 KEY，大 KEY 的内存扩容以及释放都会导致主线程卡顿。如果知道里面有没有大 KEY，可以自己写程序扫描，不过这太繁琐了。redis-cli 提供了 --bigkeys 参数可以很快扫出内存里的大 KEY，使用 -i 参数控制扫描间隔，避免扫描指令导致服务器的 ops 陡增报警。

```shell
$ ./redis-cli --bigkeys -i 0.01
# Scanning the entire keyspace to find biggest keys as well as
# average sizes per key type.  You can use -i 0.1 to sleep 0.1 sec
# per 100 SCAN commands (not usually needed).

[00.00%] Biggest zset   found so far 'hist:aht:main:async_finish:20180425:17' with 1440 members
[00.00%] Biggest zset   found so far 'hist:qps:async:authorize:20170311:27' with 2465 members
[00.00%] Biggest hash   found so far 'job:counters:6ya9ypu6ckcl' with 3 fields
[00.01%] Biggest string found so far 'rt:aht:main:device_online:68:{-4}' with 4 bytes
[00.01%] Biggest zset   found so far 'machine:load:20180709' with 2879 members
[00.02%] Biggest string found so far '6y6fze8kj7cy:{-7}' with 90 bytes
```

redis-cli 对于每一种对象类型都会记录长度最大的 KEY，对于每一种对象类型，刷新一次最高记录就会立即输出一次。它能保证输出长度为 Top1 的 KEY，但是 Top2、Top3 等 KEY 是无法保证可以扫描出来的。一般的处理方法是多扫描几次，或者是消灭了 Top1 的 KEY 之后再扫描确认还有没有次大的 KEY。

### 采样服务器指令

现在线上有一台 Redis 服务器的 OPS 太高，有很多业务模块都在使用这个 Redis，如何才能判断出来是哪个业务导致了 OPS 异常的高。这时可以对线上服务器的指令进行采样，观察采样的指令大致就可以分析出 OPS 占比高的业务点。这时就要使用 monitor 指令，它会将服务器瞬间执行的指令全部显示出来。不过使用的时候要注意即使使用 ctrl+c 中断，否则你的显示器会噼里啪啦太多的指令瞬间让你眼花缭乱。

```shell
$ redis-cli --host 192.168.x.x --port 6379 monitor
1539853410.458483 [0 10.100.90.62:34365] "GET" "6yax3eb6etq8:{-7}"
1539853410.459212 [0 10.100.90.61:56659] "PFADD" "growth:dau:20181018" "2klxkimass8w"
1539853410.462938 [0 10.100.90.62:20681] "GET" "6yax3eb6etq8:{-7}"
1539853410.467231 [0 10.100.90.61:40277] "PFADD" "growth:dau:20181018" "2kei0to86ps1"
1539853410.470319 [0 10.100.90.62:34365] "GET" "6yax3eb6etq8:{-7}"
1539853410.473927 [0 10.100.90.61:58128] "GET" "6yax3eb6etq8:{-7}"
1539853410.475712 [0 10.100.90.61:40277] "PFADD" "growth:dau:20181018" "2km8sqhlefpc"
1539853410.477053 [0 10.100.90.62:61292] "GET" "6yax3eb6etq8:{-7}"
```

### 诊断服务器时延

平时我们诊断两台机器的时延一般是使用 Unix 的 ping 指令。Redis 也提供了时延诊断指令，不过它的原理不太一样，它是诊断当前机器和 Redis 服务器之间的指令(PING指令)时延，它不仅仅是物理网络的时延，还和当前的 Redis 主线程是否忙碌有关。如果你发现 Unix 的 ping 指令时延很小，而 Redis 的时延很大，那说明 Redis 服务器在执行指令时有微弱卡顿。

```shell
$ redis-cli --host 192.168.x.x --port 6379 --latency
min: 0, max: 5, avg: 0.08 (305 samples)
```

时延单位是 ms。redis-cli 还能显示时延的分布情况，而且是图形化输出。

```shell
$ redis-cli --latency-dist
```

### 远程 rdb 备份

执行下面的命令就可以将远程的 Redis 实例备份到本地机器，远程服务器会执行一次bgsave操作，然后将 rdb 文件传输到客户端。远程 rdb 备份让我们有一种“秀才不出门，全知天下事”的感觉。

```shell
$ ./redis-cli --host 192.168.x.x --port 6379 --rdb ./user.rdb
SYNC sent to master, writing 2501265095 bytes to './user.rdb'
Transfer finished with success.
```

### 模拟从库

如果你想观察主从服务器之间都同步了那些数据，可以使用 redis-cli 模拟从库。

```shell
$ ./redis-cli --host 192.168.x.x --port 6379 --slave
SYNC with master, discarding 51778306 bytes of bulk transfer...
SYNC done. Logging commands from master.
...
```

从库连上主库的第一件事是全量同步，所以看到上面的指令卡顿这很正常，待首次全量同步完成后，就会输出增量的 aof 日志。

## 33. 破旧立新 —— 探索「紧凑列表」内部

Redis 5.0 又引入了一个新的数据结构 listpack，它是对 ziplist 结构的改进，在存储空间上会更加节省，而且结构上也比 ziplist 要精简。它的整体形式和 ziplist 还是比较接近的，如果你认真阅读了 ziplist 的内部结构分析，那么 listpack 也是比较容易理解的。

```c++
struct listpack<T> {
    int32 total_bytes; // 占用的总字节数
    int16 size; // 元素个数
    T[] entries; // 紧凑排列的元素列表
    int8 end; // 同 zlend 一样，恒为 0xFF
}
```

![image](https://mail.wangkekai.cn/DE875CCC-11B6-4B3F-B413-1C616CA2EB99.png)

首先这个 listpack 跟 ziplist 的结构几乎一摸一样，只是少了一个zltail_offset字段。ziplist 通过这个字段来定位出最后一个元素的位置，用于逆序遍历。不过 listpack 可以通过其它方式来定位出最后一个元素的位置，所以zltail_offset 字段就省掉了。

```c++
struct lpentry {
    int<var> encoding;
    optional byte[] content;
    int<var> length;
}
```

元素的结构和 ziplist 的元素结构也很类似，都是包含三个字段。不同的是长度字段放在了元素的尾部，而且存储的不是上一个元素的长度，是当前元素的长度。正是因为长度放在了尾部，所以可以省去了 zltail_offset 字段来标记最后一个元素的位置，这个位置可以通过 total_bytes 字段和最后一个元素的长度字段计算出来。

长度字段使用 varint 进行编码，不同于 skiplist 元素长度的编码为 1 个字节或者 5 个字节，listpack 元素长度的编码可以是 1、2、3、4、5 个字节。同 UTF8 编码一样，它通过字节的最高为是否为 1 来决定编码的长度。

![image](https://mail.wangkekai.cn/3ED7D7B2-BED4-4743-B0AF-102AD1FEBD6F.png)

同样，Redis 为了让 listpack 元素支持很多类型，它对 encoding 字段也进行了较为复杂的设计。

1. 0xxxxxxx 表示非负小整数，可以表示0~127。
2. 10xxxxxx 表示小字符串，长度范围是0~63，content字段为字符串的内容。
3. 110xxxxx yyyyyyyy 表示有符号整数，范围是-2048~2047。
4. 1110xxxx yyyyyyyy 表示中等长度的字符串，长度范围是0~4095，content字段为字符串的内容。
5. 11110000 aaaaaaaa bbbbbbbb cccccccc dddddddd 表示大字符串，四个字节表示长度，content字段为字符串内容。
6. 11110001 aaaaaaaa bbbbbbbb 表示 2 字节有符号整数。
7. 11110010 aaaaaaaa bbbbbbbb cccccccc 表示 3 字节有符号整数。
8. 11110011 aaaaaaaa bbbbbbbb cccccccc dddddddd 表示 4 字节有符号整数。
9. 11110011 aaaaaaaa ... hhhhhhhh 表示 8 字节有符号整数。
10. 11111111 表示 listpack 的结束符号，也就是0xFF。

### 级联更新

listpack 的设计彻底消灭了 ziplist 存在的级联更新行为，元素与元素之间完全独立，不会因为一个元素的长度变长就导致后续的元素内容会受到影响。

### 取代 ziplist

listpack 的设计的目的是用来取代 ziplist，不过当下还没有做好替换 ziplist 的准备，因为有很多兼容性的问题需要考虑，ziplist 在 Redis 数据结构中使用太广泛了，替换起来复杂度会非常之高。它目前只使用在了新增加的 Stream 数据结构中。

## 34. 金枝玉叶 —— 探索「基数树」内部

Rax 是 Redis 内部比较特殊的一个数据结构，它是一个有序字典树 (基数树 Radix Tree)，按照 key 的字典序排列，支持快速地定位、插入和删除操作。Redis 五大基础数据结构里面，能作为字典使用的有 hash 和 zset。hash 不具备排序功能，zset 则是按照 score 进行排序的。rax 跟 zset 的不同在于它是按照 key 进行排序的。Redis 作者认为 rax 的结构非常易于理解，但是实现却有相当的复杂度，需要考虑很多的边界条件，需要处理节点的分裂、合并，一不小心就会出错。

![image](https://mail.wangkekai.cn/800FFAD3-9A26-412A-9868-DCA09424CF47.png)

### 应用

你可以将一本英语字典看成一棵 radix tree，它所有的单词都是按照字典序进行排列，每个词汇都会附带一个解释，这个解释就是 key 对应的 value。有了这棵树，你就可以快速地检索单词，还可以查询以某个前缀开头的单词有哪些。

![image](https://mail.wangkekai.cn/FC20E91F-4436-4D76-8D00-1A45427F050D.png)

我们经常使用的 Web 服务器的 Router 它也是一棵 radix tree。这棵树上挂满了 URL 规则，每个 URL 规则上都会附上一个请求处理器。当一个请求到来时，我们拿这个请求的 URL 沿着树进行遍历，找到相应的请求处理器来处理。因为 URL 中可能存在正则 pattern，而且同一层的节点对顺序没有要求，所以它不算是一棵严格的 radix tree。

```go
# golang 的 HttpRouter 库
The router relies on a tree structure which makes heavy use of *common prefixes*
it is basically a *compact* [*prefix tree*](https://en.wikipedia.org/wiki/Trie) 
(or just [*Radix tree*](https://en.wikipedia.org/wiki/Radix_tree)). 
Nodes with a common prefix also share a common parent. 
Here is a short example what the routing tree for the `GET` request method could look like:

Priority   Path             Handle
9          \                *<1>
3          ├s               nil
2          |├earch\         *<2>
1          |└upport\        *<3>
2          ├blog\           *<4>
1          |    └:post      nil
1          |         └\     *<5>
2          ├about-us\       *<6>
1          |        └team\  *<7>
1          └contact\        *<8>
```

![image](https://mail.wangkekai.cn/3AE42A99-2525-475D-940F-897CF44C34F0.png)

Rax 被用在 Redis Stream 结构里面用于存储消息队列，在 Stream 里面消息 ID 的前缀是时间戳 + 序号，这样的消息可以理解为时间序列消息。使用 Rax 结构进行存储就可以快速地根据消息 ID 定位到具体的消息，然后继续遍历指定消息之后的所有消息。

Rax 被用在 Redis Cluster 中用来记录槽位和key的对应关系，这个对应关系的变量名成叫着 slots_to_keys。这个 raxNode 的 key 是由槽位编号 hashslot 和 key 组合而成的。我们知道 cluster 的槽位数量是 16384，它需要2个字节来表示，所以 rax 节点里存的 key 就是 2 个字节的 hashslot 和对象 key 拼接起来的字符串。

![image](https://mail.wangkekai.cn/FBC2A041-862A-4A94-85F5-94B2B16B0BBD.png)

因为 rax 的 key 是按照 key 前缀顺序挂接的，意味着同样的 hashslot 的对象 key 将会挂在同一个 raxNode 下面。这样我们就可以快速遍历具体某个槽位下面的所有对象 key。

### 结构

rax 中有非常多的节点，根节点、叶节点和中间节点，有些中间节点带有 value，有些中间节点纯粹是结构性需要没有对应的 value。

```c++
struct raxNode {
    int<1> isKey; // 是否没有 key，没有 key 的是根节点
    int<1> isNull; // 是否没有对应的 value，无意义的中间节点
    int<1> isCompressed; // 是否压缩存储，这个压缩的概念比较特别
    int<29> size; // 子节点的数量或者是压缩字符串的长度 (isCompressed)
    byte[] data; // 路由键、子节点指针、value 都在这里
}
```

rax 是一棵比较特殊的 radix tree，它在结构上不是标准的 radix tree。如果一个中间节点有多个子节点，那么路由键就只是一个字符。如果只有一个子节点，那么路由键就是一个字符串。后者就是所谓的「压缩」形式，多个字符压在一起的字符串。比如前面的那棵字典树在 Rax 算法中将呈现出如下结构：

![image](https://mail.wangkekai.cn/05F2E3F3-DA96-48C0-B798-2A4328CFBDA1.png)

深蓝色节点就是「压缩」节点。

接下来我们再细看raxNode.data里面存储的到底是什么东西，它是一个比较复杂的结构，按照压缩与否分为两种结构

**压缩结构** 子节点如果只有一个，那就是压缩结构，data 字段如下伪代码所示：

```c++
struct data {
    optional struct { // 取决于 header 的 size 字段是否为零
        byte[] childKey; // 路由键
        raxNode* childNode; // 子节点指针
    } child;
    optional string value; // 取决于 header 的 isNull 字段
}
```

![image](https://mail.wangkekai.cn/D9EE3A2C-4BA3-4C00-BEE7-5F10E6376A5B.png)

如果是叶子节点，child 字段就不存在。如果是无意义的中间节点 (isNull)，那么 value 字段就不存在。

非压缩节点 如果子节点有多个，那就不是压缩结构，存在多个路由键，一个键是一个字符。

```c++
struct data {
    byte[] childKeys; // 路由键字符列表
    raxNode*[] childNodes; // 多个子节点指针
    optional string value; // 取决于 header 的 isNull 字段
}
```

![image](https://mail.wangkekai.cn/FC4B9C08-111B-4F5F-B047-FF3B64D3BAC6.png)

也许你会想到如果子节点只有一个，并且路由键字符串的长度为 1 呢，那到底算压缩还是非压缩？仔细思考一下，在这种情况下，压缩和非压缩在数据结构表现形式上是一样的，不管 isCompressed 是 0 还好是 1，结构都是一样的。

### 增删节点

## 35. 精益求精 —— LFU vs LRU

Antirez 在 Redis 4.0 里引入了一个新的淘汰策略 —— LFU 模式，作者认为它比 LRU 更加优秀。  
LFU 的全称是 Least Frequently Used，表示按最近的访问频率进行淘汰，它比 LRU 更加精准地表示了一个 key 被访问的热度。

如果一个 key 长时间不被访问，只是刚刚偶然被用户访问了一下，那么在使用 LRU 算法下它是不容易被淘汰的，因为 LRU 算法认为当前这个 key 是很热的。而 LFU 是需要追踪最近一段时间的访问频率，如果某个 key 只是偶然被访问一次是不足以变得很热的，它需要在近期一段时间内被访问很多次才有机会被认为很热。

### Redis 对象的热度

Redis 的所有对象结构头中都有一个 24bit 的字段，这个字段用来记录对象的「热度」。

```c++
// redis 的对象头
typedef struct redisObject {
    unsigned type:4; // 对象类型如 zset/set/hash 等等
    unsigned encoding:4; // 对象编码如 ziplist/intset/skiplist 等等
    unsigned lru:24; // 对象的「热度」
    int refcount; // 引用计数
    void *ptr; // 对象的 body
} robj;
```

### LRU 模式

在 LRU 模式下，lru 字段存储的是 Redis 时钟 server.lruclock，Redis 时钟是一个 24bit 的整数，默认是 Unix 时间戳对 2^24 取模的结果，大约 97 天清零一次。当某个 key 被访问一次，它的对象头的 lru 字段值就会被更新为server.lruclock。

默认 Redis 时钟值每毫秒更新一次，在定时任务 serverCron 里主动设置。Redis 的很多定时任务都是在 serverCron 里面完成的，比如大型 hash 表的渐进式迁移、过期 key 的主动淘汰、触发 bgsave、bgaofrewrite 等等。

如果 server.lruclock 没有折返 (对 2^24 取模)，它就是一直递增的，这意味着对象的 LRU 字段不会超过 server.lruclock 的值。如果超过了，说明 server.lruclock 折返了。通过这个逻辑就可以精准计算出对象多长时间没有被访问——对象的空闲时间。

![image](https://mail.wangkekai.cn/8A54445C-051E-4785-9A70-AC4C98B0915C.png)

```c++
// 计算对象的空闲时间，也就是没有被访问的时间，返回结果是毫秒
unsigned long long estimateObjectIdleTime(robj *o) {
    unsigned long long lruclock = LRU_CLOCK(); // 获取 redis 时钟，也就是 server.lruclock 的值
    if (lruclock >= o->lru) {
        // 正常递增
        return (lruclock - o->lru) * LRU_CLOCK_RESOLUTION; // LRU_CLOCK_RESOLUTION 默认是 1000
    } else {
        // 折返了
        return (lruclock + (LRU_CLOCK_MAX - o->lru)) * // LRU_CLOCK_MAX 是 2^24-1
                    LRU_CLOCK_RESOLUTION;
    }
}
```

有了对象的空闲时间，就可以相互之间进行比较谁新谁旧，随机 LRU 算法靠的就是比较对象的空闲时间来决定谁该被淘汰了。

### LFU 模式

在 LFU 模式下，lru 字段 24 个 bit 用来存储两个值，分别是 ldt(last decrement time) 和 logc(logistic counter)。

![image](https://mail.wangkekai.cn/8850C21C-D38C-4135-85FB-FFA563475C4E.png)

logc 是 8 个 bit，用来存储访问频次，因为 8 个 bit 能表示的最大整数值为 255，存储频次肯定远远不够，所以这 8 个 bit 存储的是频次的对数值，并且这个值还会随时间衰减。如果它的值比较小，那么就很容易被回收。为了确保新创建的对象不被回收，新对象的这 8 个 bit 会初始化为一个大于零的值，默认是 LFU_INIT_VAL=5。

![image](https://mail.wangkekai.cn/59B5EFFB-E08E-499A-92BF-E6EFB0AAE521.png)

ldt 是 16 个位，用来存储上一次 logc 的更新时间，因为只有 16 位，所以精度不可能很高。它取的是分钟时间戳对 2^16 进行取模，大约每隔 45 天就会折返。同 LRU 模式一样，我们也可以使用这个逻辑计算出对象的空闲时间，只不过精度是分钟级别的。图中的 server.unixtime 是当前 redis 记录的系统时间戳，和 server.lruclock 一样，它也是每毫秒更新一次。

```c++
// nowInMinutes
// server.unixtime 为 redis 缓存的系统时间戳
unsigned long LFUGetTimeInMinutes(void) {
    return (server.unixtime/60) & 65535;
}

// idle_in_minutes
unsigned long LFUTimeElapsed(unsigned long ldt) {
    unsigned long now = LFUGetTimeInMinutes();
    if (now >= ldt) return now-ldt; // 正常比较
    return 65535-ldt+now; // 折返比较
}
```

ldt 的值和 LRU 模式的 lru 字段不一样的是 ldt 不是在对象被访问时更新的。它在 Redis 的淘汰逻辑进行时进行更新，淘汰逻辑只会在内存达到 maxmemory 的设置时才会触发，在每一个指令的执行之前都会触发。每次淘汰都是采用随机策略，随机挑选若干个 key，更新这个 key 的「热度」，淘汰掉「热度」最低的。因为 Redis 采用的是随机算法，如果 key 比较多的话，那么 ldt 更新的可能会比较慢。不过既然它是分钟级别的精度，也没有必要更新的过于频繁。

ldt 更新的同时也会一同衰减 logc 的值，衰减也有特定的算法。它将现有的 logc 值减去对象的空闲时间 (分钟数) 除以一个衰减系数，默认这个衰减系数lfu_decay_time是 1。如果这个值大于 1，那么就会衰减的比较慢。如果它等于零，那就表示不衰减，它是可以通过配置参数lfu-decay-time进行配置。

```c++
// 衰减 logc
unsigned long LFUDecrAndReturn(robj *o) {
    unsigned long ldt = o->lru >> 8; // 前 16bit
    unsigned long counter = o->lru & 255; // 后 8bit 为 logc
    // num_periods 为即将衰减的数量
    unsigned long num_periods = server.lfu_decay_time ? LFUTimeElapsed(ldt) / server.lfu_decay_time : 0;
    if (num_periods)
        counter = (num_periods > counter) ? 0 : counter - num_periods;
    return counter;
}
```

logc 的更新和 LRU 模式的 lru 字段一样，都会在 key 每次被访问的时候更新，只不过它的更新不是简单的+1，而是采用概率法进行递增，因为 logc 存储的是访问计数的对数值，不能直接+1。

```c++
/* Logarithmically increment a counter. The greater is the current counter value
 * the less likely is that it gets really implemented. Saturate it at 255. */
// 对数递增计数值
uint8_t LFULogIncr(uint8_t counter) {
    if (counter == 255) return 255; // 到最大值了，不能在增加了
    double baseval = counter - LFU_INIT_VAL; // 减去新对象初始化的基数值 (LFU_INIT_VAL 默认是 5)
    // baseval 如果小于零，说明这个对象快不行了，不过本次 incr 将会延长它的寿命
    if (baseval < 0) baseval = 0; 
    // 当前计数越大，想要 +1 就越困难
    // lfu_log_factor 为困难系数，默认是 10
    // 当 baseval 特别大时，最大是 (255-5)，p 值会非常小，很难会走到 counter++ 这一步
    // p 就是 counter 通往 [+1] 权力的门缝，baseval 越大，这个门缝越窄，通过就越艰难
    double p = 1.0/(baseval*server.lfu_log_factor+1);
    // 开始随机看看能不能从门缝挤进去
    double r = (double)rand()/RAND_MAX; // 0 < r < 1
    if (r < p) counter++;
    return counter;
}
```

### 为什么 Redis 要缓存系统时间戳？

我们平时使用系统时间戳时，常常是不假思索地使用 System.currentTimeInMillis 或者 time.time() 来获取系统的毫秒时间戳。Redis 不能这样，因为每一次获取系统时间戳都是一次系统调用，系统调用相对来说是比较费时间的，作为单线程的 Redis 表示承受不起，所以它需要对时间进行缓存，获取时间都直接从缓存中直接拿。

### redis 为什么在获取 lruclock 时使用原子操作？

我们知道 Redis 是单线程的，那为什么 lruclock 要使用原子操作 atomicGet 来获取呢？

```c++
unsigned int LRU_CLOCK(void) {
    unsigned int lruclock;
    if (1000/server.hz <= LRU_CLOCK_RESOLUTION) {
        // 这里原子操作，通常会走这里，我们只需要注意这里
        atomicGet(server.lruclock,lruclock);  
    } else {
        // 直接通过系统调用获取时间戳，hz 配置的太低 (一般不会这么干)，lruclock 更新不及时，需要实时获取系统时间戳
        lruclock = getLRUClock(); 
    }
    return lruclock;
}
```

因为 Redis 实际上并不是单线程，它背后还有几个异步线程也在默默工作。这几个线程也要访问 Redis 时钟，所以 lruclock 字段是需要支持多线程读写的。使用 atomic 读写能保证多线程 lruclock 数据的一致性。

### 如何打开 LFU 模式？

Redis 4.0 给淘汰策略配置参数 maxmemory-policy 增加了 2 个选项，分别是 volatile-lfu 和 allkeys-lfu，分别是对带过期时间的 key 进行 lfu 淘汰以及对所有的 key 执行 lfu 淘汰算法。打开了这个选项之后，就可以使用 object freq 指令获取对象的 lfu 计数值了。

```redis
> config set maxmemory-policy allkeys-lfu
OK
> set codehole yeahyeahyeah
OK
// 获取计数值，初始化为 LFU_INIT_VAL=5
> object freq codehole
(integer) 5
// 访问一次
> get codehole
"yeahyeahyeah"
// 计数值增加了
> object freq codehole
(integer) 6
```

## 36. 如履薄冰 —— 懒惰删除的巨大牺牲

**异步线程在 Redis 内部有一个特别的名称，它就是 BIO，全称是 Background IO，意思是在背后默默干活的 IO 线程。**不过内存回收本身并不是什么 IO 操作，只是 CPU 的计算消耗可能会比较大而已。

### 懒惰删除的最初实现不是异步线程

Antirez 实现懒惰删除时，它并不是一开始就想到了异步线程。<u>最初的尝试是在主线程里，使用类似于字典渐进式搬迁那样来实现渐进式删除回收。</u>比如对于一个非常大的字典来说，懒惰删除是采用类似于 scan 操作的方法，通过遍历第一维数组来逐步删除回收第二维链表的内容，等到所有链表都回收完了，再一次性回收第一维数组。这样也可以达到删除大对象时不阻塞主线程的效果。

但是说起来容易做起来却很难。渐进式回收需要仔细控制回收频率，它不能回收的太猛，这会导致 CPU 资源占用过多，也不能回收的蜗牛慢，因为内存回收不及时可能导致内存持续增长。

Antirez 需要采用合适的自适应算法来控制回收频率。他首先想到的是检测内存增长的趋势是增长 (+1) 还是下降 (-1) 来渐进式调整回收频率系数，这样的自适应算法实现也很简单。但是测试后发现在服务繁忙的时候，QPS 会下降到正常情况下 65% 的水平，这点非常致命。

所以 Antirez 才使用了如今使用的方案——异步线程。异步线程这套方案就简单多了，释放内存不用为每种数据结构适配一套渐进式释放策略，也不用搞个自适应算法来仔细控制回收频率。将对象从全局字典中摘掉，然后往队列里一扔，主线程就干别的去了。异步线程从队列里取出对象来，直接走正常的同步释放逻辑就可以了。

不过使用异步线程也是有代价的，主线程和异步线程之间在内存回收器 (jemalloc) 的使用上存在竞争。这点竞争消耗是可以忽略不计的，因为 Redis 的主线程在内存的分配与回收上花的时间相对整体运算时间而言是极少的。

### 异步线程方案其实也相当复杂

Redis 的内部对象有共享机制，严重阻碍了异步线程方案的改造，这点非常可怕。

比如集合的并集操作 sunionstore 用来将多个集合合并成一个新集合

```redis
> sadd src1 value1 value2 value3
(integer) 3
> sadd src2 value3 value4 value5
(integer) 3
> sunionstore dest src1 src2
(integer) 5
> smembers dest
1) "value2"
2) "value3"
3) "value1"
4) "value4"
5) "value5"
```

我们看到新的集合包含了旧集合的所有元素。但是这里有一个我们没看到的 trick。那就是底层的字符串对象被共享了。

![image](https://mail.wangkekai.cn/321A6E01-4D67-4CCD-892A-3927A3D71BC5.png)

为什么对象共享是懒惰删除的巨大障碍呢？因为懒惰删除相当于彻底砍掉某个树枝，将它扔到异步删除队列里去。注意这里必须是彻底删除，而不能藕断丝连。**如果底层对象是共享的，那就做不到彻底删除。**

![image](https://mail.wangkekai.cn/0AC74D70-4E9E-411E-AC61-6482AA0C0E36.png)

所以 Antirez 为了支持懒惰删除，将对象共享机制彻底抛弃，它将这种对象结构称为「share-nothing」，也就是无共享设计。但是甩掉对象共享谈何容易！这种对象共享机制散落在源代码的各个角落，牵一发而动全身，改起来犹如在布满地雷的道路上小心翼翼地行走。

不过 Antirez 还是决心改了，他将这种改动描述为「绝望而疯狂」，可见改动之大之深之险，前后花了好几周的时间才改完。不过效果也是很明显的，对象的删除操作再也不会导致主线程卡顿了。

### 异步删除的实现

主线程需要将删除任务传递给异步线程，它是通过一个普通的双向链表来传递的。因为链表需要支持多线程并发操作，所以它需要有锁来保护。

执行懒惰删除时，Redis 将删除操作的相关参数封装成一个 bio_job 结构，然后追加到链表尾部。异步线程通过遍历链表摘取 job 元素来挨个执行异步任务。

```c++
struct bio_job {
    time_t time;  // 时间字段暂时没有使用，应该是预留的
    void *arg1, *arg2, *arg3;
};
```

我们注意到这个 job 结构有三个参数，为什么删除对象需要三个参数呢？我们继续看代码：

```c++
   /* What we free changes depending on what arguments are set:
     * arg1 -> free the object at pointer.
     * arg2 & arg3 -> free two dictionaries (a Redis DB).
     * only arg3 -> free the skiplist. */
   if (job->arg1)
      // 释放一个普通对象，string/set/zset/hash 等等，用于普通对象的异步删除
      lazyfreeFreeObjectFromBioThread(job->arg1);
   else if (job->arg2 && job->arg3)
      // 释放全局 redisDb 对象的 dict 字典和 expires 字典，用于 flushdb
      lazyfreeFreeDatabaseFromBioThread(job->arg2,job->arg3);
   else if (job->arg3)
      // 释放 Cluster 的 slots_to_keys 对象，参见源码篇的「基数树」小节
      lazyfreeFreeSlotsMapFromBioThread(job->arg3);
```

可以看到通过组合这三个参数可以实现不同结构的释放逻辑。接下来我们继续追踪普通对象的异步删除 lazyfreeFreeObjectFromBioThread 是如何进行的，请仔细阅读代码注释。

```java
void lazyfreeFreeObjectFromBioThread(robj *o) {
    decrRefCount(o); // 降低对象的引用计数，如果为零，就释放
    atomicDecr(lazyfree_objects,1); // lazyfree_objects 为待释放对象的数量，用于统计
}

// 减少引用计数
void decrRefCount(robj *o) {
    if (o->refcount == 1) {
        // 该释放对象了
        switch(o->type) {
        case OBJ_STRING: freeStringObject(o); break;
        case OBJ_LIST: freeListObject(o); break;
        case OBJ_SET: freeSetObject(o); break;
        case OBJ_ZSET: freeZsetObject(o); break;
        case OBJ_HASH: freeHashObject(o); break;  // 释放 hash 对象，继续追踪
        case OBJ_MODULE: freeModuleObject(o); break;
        case OBJ_STREAM: freeStreamObject(o); break;
        default: serverPanic("Unknown object type"); break;
        }
        zfree(o);
    } else {
        if (o->refcount <= 0) serverPanic("decrRefCount against refcount <= 0");
        if (o->refcount != OBJ_SHARED_REFCOUNT) o->refcount--; // 引用计数减 1
    }
}

// 释放 hash 对象
void freeHashObject(robj *o) {
    switch (o->encoding) {
    case OBJ_ENCODING_HT:
        // 释放字典，我们继续追踪
        dictRelease((dict*) o->ptr);
        break;
    case OBJ_ENCODING_ZIPLIST:
        // 如果是压缩列表可以直接释放
        // 因为压缩列表是一整块字节数组
        zfree(o->ptr);
        break;
    default:
        serverPanic("Unknown hash encoding type");
        break;
    }
}

// 释放字典，如果字典正在迁移中，ht[0] 和 ht[1] 分别存储旧字典和新字典
void dictRelease(dict *d)
{
    _dictClear(d,&d->ht[0],NULL); // 继续追踪
    _dictClear(d,&d->ht[1],NULL);
    zfree(d);
}

// 这里要释放 hashtable 了
// 需要遍历第一维数组，然后继续遍历第二维链表，双重循环
int _dictClear(dict *d, dictht *ht, void(callback)(void *)) {
    unsigned long i;

    /* Free all the elements */
    for (i = 0; i < ht->size && ht->used > 0; i++) {
        dictEntry *he, *nextHe;

        if (callback && (i & 65535) == 0) callback(d->privdata);

        if ((he = ht->table[i]) == NULL) continue;
        while(he) {
            nextHe = he->next;
            dictFreeKey(d, he); // 先释放 key
            dictFreeVal(d, he); // 再释放 value
            zfree(he); // 最后释放 entry
            ht->used--;
            he = nextHe;
        }
    }
    /* Free the table and the allocated cache structure */
    zfree(ht->table); // 可以回收第一维数组了
    /* Re-initialize the table */
    _dictReset(ht);
    return DICT_OK; /* never fails */
}
```

### 队列安全

前面提到任务队列是一个不安全的双向链表，需要使用锁来保护它。当主线程将任务追加到队列之前它需要加锁，追加完毕后，再释放锁，还需要唤醒异步线程，如果它在休眠的话。

```c++
void bioCreateBackgroundJob(int type, void *arg1, void *arg2, void *arg3) {
    struct bio_job *job = zmalloc(sizeof(*job));

    job->time = time(NULL);
    job->arg1 = arg1;
    job->arg2 = arg2;
    job->arg3 = arg3;
    pthread_mutex_lock(&bio_mutex[type]); // 加锁
    listAddNodeTail(bio_jobs[type],job); // 追加任务
    bio_pending[type]++; // 计数
    pthread_cond_signal(&bio_newjob_cond[type]); // 唤醒异步线程
    pthread_mutex_unlock(&bio_mutex[type]); // 释放锁
}
```

异步线程需要对任务队列进行轮训处理，依次从链表表头摘取元素逐个处理。摘取元素的时候也需要加锁，摘出来之后再解锁。如果一个元素都没有，它需要等待，直到主线程来唤醒它继续工作。

```c++
// 异步线程执行逻辑
void *bioProcessBackgroundJobs(void *arg) {
...
   pthread_mutex_lock(&bio_mutex[type]); // 先加锁
   ...
   // 循环处理
   while(1) {
      listNode *ln;

      /* The loop always starts with the lock hold. */
      if (listLength(bio_jobs[type]) == 0) {
         // 对列空，那就睡觉吧
         pthread_cond_wait(&bio_newjob_cond[type],&bio_mutex[type]);
         continue;
      }
      /* Pop the job from the queue. */
      ln = listFirst(bio_jobs[type]); // 获取队列头元素
      job = ln->value;
      /* It is now possible to unlock the background system as we know have
      * a stand alone job structure to process.*/
      pthread_mutex_unlock(&bio_mutex[type]); // 释放锁

      // 这里是处理过程，为了省纸，就略去了
      ...
      
      // 释放任务对象
      zfree(job);

      ...
      
      // 再次加锁继续处理下一个元素
      pthread_mutex_lock(&bio_mutex[type]);
      // 因为任务已经处理完了，可以放心从链表中删除节点了
      listDelNode(bio_jobs[type],ln);
      bio_pending[type]--; // 计数减 1
   }
```

研究完这些加锁解锁的代码后，我开始有点当心主线程的性能。我们都知道加锁解锁是一个相对比较耗时的操作，尤其是悲观锁最为耗时。如果删除很频繁，主线程岂不是要频繁加锁解锁。所以这里肯定还有优化空间，Java 的 ConcurrentLinkQueue 就没有使用这样粗粒度的悲观锁，它优先使用 cas 来控制并发。

## 37. 跋山涉水 —— 深入字典遍历

![image](https://user-gold-cdn.xitu.io/2018/8/15/1653bd1a8cbe5f4d?imageView2/0/w/1280/h/960/ignore-error/1)

### 一边遍历一边修改

我们知道 Redis 对象树的主干是一个字典，如果对象很多，这个主干字典也会很大。当我们使用 keys 命令搜寻指定模式的 key 时，它会遍历整个主干字典。值得注意的是，在遍历的过程中，如果满足模式匹配条件的 key 被找到了，还需要判断 key 指向的对象是否已经过期。如果过期了就需要从主干字典中将该 key 删除。

### 重复遍历

字典在扩容的时候要进行渐进式迁移，会存在新旧两个 hashtable。遍历需要对这两个 hashtable 依次进行，先遍历完旧的 hashtable，再继续遍历新的 hashtable。如果在遍历的过程中进行了 rehashStep，将已经遍历过的旧的 hashtable 的元素迁移到了新的 hashtable 中，那么遍历会不会出现元素的重复？这也是遍历需要考虑的疑难之处，下面我们来看看 Redis 是如何解决这个问题的。

### 迭代器的结构

Redis 为字典的遍历提供了 2 种迭代器，一种是安全迭代器，另一种是不安全迭代器。

```c++
typedef struct dictIterator {
   dict *d; // 目标字典对象
   long index; // 当前遍历的槽位置，初始化为-1
   int table; // ht[0] or ht[1]
   int safe; // 这个属性非常关键，它表示迭代器是否安全
   dictEntry *entry; // 迭代器当前指向的对象
   dictEntry *nextEntry; // 迭代器下一个指向的对象
   long long fingerprint; // 迭代器指纹，放置迭代过程中字典被修改
} dictIterator;

// 获取非安全迭代器，只读迭代器，允许rehashStep
dictIterator *dictGetIterator(dict *d)
{
   dictIterator *iter = zmalloc(sizeof(*iter));

   iter->d = d;
   iter->table = 0;
   iter->index = -1;
   iter->safe = 0;
   iter->entry = NULL;
   iter->nextEntry = NULL;
   return iter;
}

// 获取安全迭代器，允许触发过期处理，禁止rehashStep
dictIterator *dictGetSafeIterator(dict *d) {
   dictIterator *i = dictGetIterator(d);

   i->safe = 1;
   return i;
}
```

迭代器的「安全」指的是在遍历过程中可以对字典进行查找和修改，不用感到担心，因为查找和修改会触发过期判断，会删除内部元素。「安全」的另一层意思是迭代过程中不会出现元素重复，为了保证不重复，就会禁止 rehashStep。

而「不安全」的迭代器是指遍历过程中字典是只读的，你不可以修改，你只能调用 dictNext 对字典进行持续遍历，不得调用任何可能触发过期判断的函数。不过好处是不影响 rehash，代价就是遍历的元素可能会出现重复。

安全迭代器在刚开始遍历时，会给字典打上一个标记，有了这个标记，rehashStep 就不会执行，遍历时元素就不会出现重复。

```c++
typedef struct dict {
    dictType *type;
    void *privdata;
    dictht ht[2];
    long rehashidx;
    // 这个就是标记，它表示当前加在字典上的安全迭代器的数量
    unsigned long iterators;
} dict;

// 如果存在安全的迭代器，就禁止rehash
static void _dictRehashStep(dict *d) {
    if (d->iterators == 0) dictRehash(d,1);
}
```

### 迭代过程

安全的迭代器在遍历过程中允许删除元素，意味着字典第一维数组下面挂接的链表中的元素可能会被摘走，元素的 next 指针就会发生变动，这是否会影响迭代过程呢？下面我们仔细研究一下迭代函数的代码逻辑。

```c++
dictEntry *dictNext(dictIterator *iter)
{
    while (1) {
        if (iter->entry == NULL) {
            // 遍历一个新槽位下面的链表，数组的index往前移动了
            dictht *ht = &iter->d->ht[iter->table];
            if (iter->index == -1 && iter->table == 0) {
                // 第一次遍历，刚刚进入遍历过程
                // 也就是ht[0]数组的第一个元素下面的链表
                if (iter->safe) {
                  // 给字典打安全标记，禁止字典进行rehash
                  iter->d->iterators++;
                } else {
                  // 记录迭代器指纹，就好比字典的md5值
                  // 如果遍历过程中字典有任何变动，指纹就会改变
                  iter->fingerprint = dictFingerprint(iter->d);
                }      
            }
            iter->index++; // index=0，正式进入第一个槽位
            if (iter->index >= (long) ht->size) {
                // 最后一个槽位都遍历完了
                if (dictIsRehashing(iter->d) && iter->table == 0) {
                    // 如果处于rehash中，那就继续遍历第二个 hashtable
                    iter->table++;
                    iter->index = 0;
                    ht = &iter->d->ht[1];
                } else {
                    // 结束遍历
                    break;
                }
            }
            // 将当前遍历的元素记录到迭代器中
            iter->entry = ht->table[iter->index];
        } else {
            // 直接将下一个元素记录为本次迭代的元素
            iter->entry = iter->nextEntry;
        }
        if (iter->entry) {
            // 将下一个元素也记录到迭代器中，这点非常关键
            // 防止安全迭代过程中当前元素被过期删除后，找不到下一个需要遍历的元素
            
            // 试想如果后面发生了rehash，当前遍历的链表被打散了，会发生什么
            // 这里要使劲发挥自己的想象力来理解
            // 旧的链表将一分为二，打散后重新挂接到新数组的两个槽位下
            // 结果就是会导致当前链表上的元素会重复遍历
            
            // 如果rehash的链表是index前面的链表，那么这部分链表也会被重复遍历
            iter->nextEntry = iter->entry->next;
            return iter->entry;
        }
    }
    return NULL;
}

// 遍历完成后要释放迭代器，安全迭代器需要去掉字典的禁止rehash的标记
// 非安全迭代器还需要检查指纹，如果有变动，服务器就会奔溃(failfast)
void dictReleaseIterator(dictIterator *iter)
{
    if (!(iter->index == -1 && iter->table == 0)) {
        if (iter->safe)
            iter->d->iterators--; // 去掉禁止rehash的标记
        else
            assert(iter->fingerprint == dictFingerprint(iter->d));
    }
    zfree(iter);
}

// 计算字典的指纹，就是将字典的关键字段进行按位糅合到一起
// 这样只要有任意的结构变动，指纹都会发生变化
// 如果只是某个元素的value被修改了，指纹不会发生变动
long long dictFingerprint(dict *d) {
    long long integers[6], hash = 0;
    int j;

    integers[0] = (long) d->ht[0].table;
    integers[1] = d->ht[0].size;
    integers[2] = d->ht[0].used;
    integers[3] = (long) d->ht[1].table;
    integers[4] = d->ht[1].size;
    integers[5] = d->ht[1].used;

    for (j = 0; j < 6; j++) {
        hash += integers[j];
        hash = (~hash) + (hash << 21);
        hash = hash ^ (hash >> 24);
        hash = (hash + (hash << 3)) + (hash << 8);
        hash = hash ^ (hash >> 14);
        hash = (hash + (hash << 2)) + (hash << 4);
        hash = hash ^ (hash >> 28);
        hash = hash + (hash << 31);
    }
    return hash;
}
```

值得注意的是在字典扩容时进行 rehash，将旧数组中的链表迁移到新的数组中。某个具体槽位下的链表只可能会迁移到新数组的两个槽位中。

```c++
hash mod 2^n = k
hash mod 2^(n+1) = k or k+2^n
```

### 迭代器的选择

除了 keys 指令使用了安全迭代器，因为结果不允许重复。那还有其它的地方使用了安全迭代器么，什么情况下遍历适合使用非安全迭代器呢？

简单一点说，那就是如果遍历过程中不允许出现重复，那就使用 SafeIterator，比如下面的两种情况

1. bgaofrewrite 需要遍历所有对象转换称操作指令进行持久化，绝对不允许出现重复
2. bgsave 也需要遍历所有对象来持久化，同样不允许出现重复

如果遍历过程中需要处理元素过期，需要对字典进行修改，那也必须使用 SafeIterator，因为非安全的迭代器是只读的。
