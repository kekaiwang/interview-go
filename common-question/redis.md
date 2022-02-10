[toc]

## Redis 基础数据类型和结构

所有的 Redis 对象都有下面这个对象头结构：

```c
struct RedisObject {
    int4 type; // 4bits
    int4 encoding; // 4bits
    int24 lru; // 24bits
    int32 refcount; // 4bytes
    void *ptr; // 8bytes，64-bit system
} robj;
```

不同的对象具有不同的类型 `type(4bit)`，同一个类型的 type 会有不同的存储形式 `encoding(4bit)`，为了记录对象的 LRU 信息，使用了 24 个 bit 来记录 LRU 信息。每个对象都有个引用计数，当引用计数为零时，对象就会被销毁，内存被回收。`ptr` 指针将指向对象内容 (body) 的具体存储位置。这样一个 `RedisObject` 对象头需要占据 16 字节的存储空间。

### String 字符串类型

#### 基础介绍

字符串结构使用非常广泛，**一个常见的用途就是缓存用户信息、锁、计数器和限速器**等。

如果 value 值是一个整数，还可以对它进行自增操作。
**自增是有范围的，它的范围是 signed long 的最大最小值，超过了这个值，Redis 会报错**。

**Redis 规定字符串的长度不得超过 512M 字节**。

#### 常用命令

```shell
> set name Jugg 
OK
> exists name # 是否存在
(integer) 1
> del name # 删除
(integer) 1
> get name
(nil)

# 批量操作
> mset name1 boy name2 girl name3 unknown
> mget name1 name2 name3 # 返回一个列表
1) "Jugg"
2) "holycoder"
3) (nil)

// 过期
> expire name 5  # 5s 后过期
> get name # wait for 5s
(nil)
> ttl name
-1

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

> setex name 5 Jugg  # 5s 后过期，等价于 set+expire

> setnx name Jugg  # 如果 name 不存在就执行 set 创建，如果已经存在，set 创建不成功
(integer) 1 / 0
> get name
"Jugg"

// set 指令扩展
> set lock:Jugg true ex 5 nx
OK
```

#### 底层结构

```c
struct SDS<T> {
  T capacity; // 数组容量
  T len; // 数组长度
  byte flags; // 特殊标识位，不理睬它
  byte[] content; // 数组内容
}
```

Redis 的字符串是动态字符串，是可以修改的字符串，采用预分配冗余空间的方式来减少内存的频繁分配，*内部为当前字符串实际分配的空间 `capacity` 一般要高于实际字符串长度 `len`*。但是 Redis 创建字符串时 `len` 和 `capacity` 一样长，不会多分配冗余空间，这是因为绝大多数场景下我们不会使用 `append` 操作来修改字符串。

上面的 `SDS` 结构使用了范型 T，为什么不直接用 int 呢，这是因为当字符串比较短时，len 和 capacity 可以使用 `byte` 和 `short` 来表示，Redis 为了对内存做极致的优化，**不同长度的字符串使用不同的结构体来表示**。

#### embstr vs raw

Redis 的字符串有两种存储方式，**在长度特别短时**，使用 `emb` 形式存储 (embeded)，**当长度超过 44 时**，使用 `raw` 形式存储。

在字符串比较小时，SDS 对象头的大小是 capacity+3，至少是 3。意味着分配一个字符串的最小空间占用为 19 字节 (16+3)。RedisObject 占用 16 字节，SDS 占用 3 字节。

- **`embstr` 存储形式**: 它将 RedisObject 对象头和 SDS 对象连续存在一起，使用 `malloc` 方法一次分配。
- **`raw` 存储形式**: 它需要两次 `malloc`，两个对象头在内存地址上一般是不连续的。

字符串是由多个字节组成，每个字节又是由 8 个 bit 组成，如此便可以将一个字符串看成很多 bit 的组合，这便是 bitmap「位图」数据结构。

#### String 扩容

**当字符串长度小于 1M 时，扩容都是加倍现有的空间**，也就是保留 100% 的冗余空间。如果超过 1M，扩容时一次只会多扩 1M 的空间。**字符串最大长度为 512M。**

### List 列表

#### List 基础介绍

**Redis 的列表结构常用来做异步队列使用**。比如秒杀场景将需要延后处理的任务结构体序列化成字符串塞进 Redis 的列表，另一个线程从这个列表中轮询数据进行处理。

Redis 的列表相当于 Java 语言里面的 LinkedList，***注意它是链表而不是数组***。  
**这意味着 list 的插入和删除操作非常快，时间复杂度为 O(1)，但是索引定位很慢，时间复杂度为 O(n)，这点让人非常意外**。

#### List 常用命令

```shell
# 对列 右进左出
> rpush books python java golang
(integer) 3
> llen books
(integer) 3
> lpop books
"python"
...
> lpop books
(nil)

# 栈 - 右进右出
> rpush books python java golang
(integer) 3
> rpop books
"golang"
...
> rpop books
(nil)

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

#### 慢操作

`lindex` 方法，它需要对链表进行遍历，性能随着参数 `index` 增大而变差。

`ltrim` 和字面上的含义不太一样，个人觉得它叫 `lretain`(保留) 更合适一些，因为 ltrim 跟的两个参数 `start_index` 和 `end_index` 定义了一个区间，在这个区间内的值，ltrim 要保留，区间之外统统砍掉。我们可以通过 ltrim 来实现一个定长的链表，这一点非常有用。

index 可以为负数，`index=-1` 表示倒数第一个元素，同样 `index=-2` 表示倒数第二个元素。

**Redis 底层存储的还不是一个简单的 linkedlist，而是称之为 *快速链表* `quicklist` 的一个结构**。

#### 快速列表

**首先在列表元素较少的情况下会使用一块连续的内存存储**，这个结构是 `ziplist`，也就是**压缩列表**。*它将所有的元素紧挨着一起存储，分配的是一块连续的内存*。  
**当数据量比较多的时候才会改成 `quicklist`**。

考虑到链表的附加空间相对太高，`prev` 和 `next` 指针就要占去 16 个字节 (64bit 系统的指针是 8 个字节)，另外每个节点的内存都是单独分配，会加剧内存的碎片化，影响内存管理效率。  
后续版本对列表数据结构进行了改造，**使用 quicklist 代替了 ziplist 和 linkedlist**。

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

*quicklist 是 ziplist 和 linkedlist 的混合体，它将 linkedlist 按段切分，每一段使用 ziplist 来紧凑存储，多个 ziplist 之间使用双向指针串接起来*。
![image](https://mail.wangkekai.cn/7AAB35D9-BBDB-46C0-879F-04746234203D.png)

##### 每个 ziplist 存多少元素？

`quicklist` 内部默认单个 `ziplist` 长度为 8k 字节，超出了这个字节数，就会新起一个 ziplist。ziplist 的长度由配置参数 `list-max-ziplist-size` 决定。

##### 压缩深度

![image](https://mail.wangkekai.cn/71ABE178-6788-4362-90F7-EB1F74495E9E.png)
quicklist 默认的压缩深度是 0，也就是不压缩。压缩的实际深度由配置参数 `list-compress-depth` 决定。为了支持快速的 `push/pop` 操作，quicklist 的首尾两个 `ziplist` 不压缩，此时深度就是 1。如果深度为 2，就表示 quicklist 的首尾第一个 ziplist 以及首尾第二个 ziplist 都不压缩。

#### 压缩列表

Redis 为了节约内存空间使用，**`zset` 和 `hash` 对象在元素个数较少的时候，采用压缩列表 (ziplist) 进行存储**。  
**压缩列表是一块连续的内存空间，元素之间紧挨着存储，没有任何冗余空隙**。

```c++
struct ziplist<T> {
    int32 zlbytes; // 整个压缩列表占用字节数
    int32 zltail_offset; // 最后一个元素距离压缩列表起始位置的偏移量，用于快速定位到最后一个节点
    int16 zllength; // 元素个数
    T[] entries; // 元素内容列表，挨个挨个紧凑存储
    int8 zlend; // 标志压缩列表的结束，值恒为 0xFF
}

struct entry {
    int<var> prevlen; // 前一个 entry 的字节长度
    int<var> encoding; // 元素类型编码
    optional byte[] content; // 元素内容
}
```

压缩列表为了支持双向遍历，所以才会有 `ztail_offset` 这个字段，用来快速定位到最后一个元素，然后倒着遍历。`entry` 的 `prevlen` 字段表示前一个 `entry` 的字节长度，当压缩列表倒着遍历时，需要通过这个字段来快速定位到下一个元素的位置。

##### 增加元素

因为 `ziplist` 都是紧凑存储，没有冗余空间 (对比一下 Redis 的字符串结构)。意味着插入一个新的元素就需要调用 `realloc` 扩展内存。取决于内存分配器算法和当前的 ziplist 内存大小，realloc 可能会重新分配新的内存空间，并将之前的内容一次性拷贝到新的地址，也可能在原有的地址上进行扩展，这时就不需要进行旧内容的内存拷贝。

如果 ziplist 占据内存太大，重新分配内存和拷贝内存就会有很大的消耗。所以 **ziplist 不适合存储大型字符串，存储的元素也不宜过多**。

##### 级联更新

每个 entry 都会有一个 `prevlen` 字段存储前一个 entry 的长度。如果内容小于 254 字节，prevlen 用 1 字节存储，否则就是 5 字节。

如果 ziplist 里面每个 entry 恰好都存储了 253 字节的内容，那么第一个 entry 内容的修改就会导致后续所有 entry 的级联更新，这就是一个比较耗费计算资源的操作。

##### IntSet 小整数集合

当 set 集合容纳的元素都是整数并且元素个数较小时，Redis 会使用 `intset` 来存储结合元素。intset 是紧凑的数组结构，同时支持 16 位、32 位和 64 位整数。

```c++
struct intset<T> {
    int32 encoding; // 决定整数位宽是 16 位、32 位还是 64 位
    int32 length; // 元素个数
    int<T> contents; // 整数数组，可以是 16 位、32 位和 64 位
}
```

```c++
> sadd codehole 1 2 3
(integer) 3
> debug object codehole
Value at:0x7fec2dc2bde0 refcount:1 encoding:intset serializedlength:15 lru:6065795 lru_seconds_idle:4
> sadd codehole go java python
(integer) 3
> debug object codehole
Value at:0x7fec2dc2bde0 refcount:1 encoding:hashtable serializedlength:22 lru:6065810 lru_seconds_idle:5
```

注意观察 debug object 的输出字段 encoding 的值，可以发现**当 set 里面放进去了非整数值时，存储形式立即从 intset 转变成了 hash 结构**

### hash 字典

#### hash 基础介绍

hash 可以记录结构体信息，如帖子的标题、摘要、作者和封面信息、点赞数、评论数和点击数；缓存近期热帖内容

Redis 的字典相当于 Java 语言里面的 HashMap，它是**无序字典**。内部实现结构是**数组 + 链表二维结构。第一维 hash 的数组位置碰撞时，就会将碰撞的元素使用链表串接起来**。
![image](https://mail.wangkekai.cn/999C4B1B-8CD5-42E5-BF26-74022C9B2326.png)

**Redis 为了高性能，不能堵塞服务，所以采用了渐进式 rehash 策略**。
![image](https://mail.wangkekai.cn/C5E5EF39-9A78-4D29-A1F7-C3983C4F4A58.png)

渐进式 rehash 会在 rehash 的同时，保留新旧两个 hash 结构，查询时会同时查询两个 hash 结构，然后在后续的定时任务中以及 hash 操作指令中，循序渐进地将旧 hash 的内容一点点迁移到新的 hash 结构中。当搬迁完成了，就会使用新的 hash 结构取而代之。

**当 hash 移除了最后一个元素之后，该数据结构自动被删除，内存被回收**。

hash 结构也可以用来存储用户信息，不同于字符串一次性需要全部序列化整个对象，hash 可以对用户结构中的每个字段单独存储。这样当我们需要获取用户信息时可以进行部分获取。而以整个字符串的形式去保存用户信息的话就只能一次性全部读取，这样就会比较浪费网络流量。

hash 也有缺点，*hash 结构的存储消耗要高于单个字符串*，到底该使用 hash 还是字符串，需要根据实际情况再三权衡。

**下面的情况 hash 会使用 ziplist 存储 hash 对象**:

1. 所有的键值对的健和值的字符串长度都小于等于64byte（一个英文字母一个字节）
2. 哈希对象保存的键值对数量小于512个。

#### hash 常用命令

```shell
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

# 单个字符计数
> hincrby user-test age 1
```

#### dict - 内部结构

dict 是 Redis 服务器中出现最为频繁的复合型数据结构，除了 hash 结构的数据会用到字典外，**整个 Redis 数据库的所有 key 和 value 也组成了一个全局字典**，还有带过期时间的 key 集合也是一个字典。  
**zset 集合中存储 value 和 score 值的映射关系也是通过 dict 结构实现的**。

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

![image](https://mail.wangkekai.cn/1639922760460.jpg)

字典数据结构的精华就落在了 hashtable 结构上了。hashtable 的结构和 Java 的 HashMap 几乎是一样的，都是通过分桶的方式解决 hash 冲突。**第一维是数组，第二维是链表。数组中存储的是第二维链表的第一个元素的指针**。

#### 渐进式 rehash

大字典的扩容是比较耗时间的，需要重新申请新的数组，然后将旧字典所有链表中的元素重新挂接到新的数组下面，这是一个 O(n) 级别的操作，作为单线程的Redis表示很难承受这样耗时的过程。步子迈大了会扯着蛋，所以 Redis 使用渐进式 rehash 小步搬迁。虽然慢一点，但是肯定可以搬完。

搬迁操作埋伏在当前字典的后续指令中(来自客户端的hset/hdel指令等)，但是有可能客户端闲下来了，没有了后续指令来触发这个搬迁，Redis还会在**定时任务**中对字典进行主动搬迁。

#### 查找的过程

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

#### hash 函数

hashtable 的性能好不好完全取决于 hash 函数的质量。hash 函数如果可以将 key 打散的比较均匀，那么这个 hash 函数就是个好函数。Redis 的字典默认的 hash 函数是 siphash。siphash 算法即使在输入 key 很小的情况下，也可以产生随机性特别好的输出，而且它的性能也非常突出。对于 Redis 这样的单线程来说，字典数据结构如此普遍，字典操作也会非常频繁，hash 函数自然也是越快越好。

#### hash 攻击

如果 hash 函数存在偏向性，黑客就可能利用这种偏向性对服务器进行攻击。存在偏向性的 hash 函数在特定模式下的输入会导致 hash 第二维链表长度极为不均匀，甚至所有的元素都集中到个别链表中，直接导致查找效率急剧下降，从O(1)退化到O(n)。有限的服务器计算能力将会被 hashtable 的查找效率彻底拖垮。这就是所谓 hash 攻击。

#### 扩容条件

**正常情况下，当 hash 表中元素的个数等于第一维数组的长度时，就会开始扩容，扩容的新数组是原数组大小的 2 倍**。不过如果 Redis 正在做 bgsave，为了减少内存页的过多分离 (Copy On Write)，Redis 尽量不去扩容 (dict_can_resize)，但是如果 hash 表已经非常满了，元素的个数已经达到了第一维数组长度的 5 倍  (dict_force_resize_ratio)，说明 hash 表已经过于拥挤了，这个时候就会强制扩容。

#### 缩容条件

当 hash 表因为元素的逐渐删除变得越来越稀疏时，Redis 会对 hash 表进行缩容来减少 hash 表的第一维数组空间占用。  
**缩容的条件是元素个数低于数组长度的 10%。缩容不会考虑 Redis 是否正在做 bgsave**。

#### set 的结构

Redis 里面 set 的结构底层实现也是字典，只不过所有的 value 都是 NULL，其它的特性和字典一模一样。

## 1. redis 实现消息队列的方案

### list

### PubSub

### Stream

## Redis 为什么这么快？

**也可以问单线程怎么支持高并发？**

1. **纯内存操作**

    不论读写操作都是在内存上完成的，跟传统的磁盘文件数据存储相比，避免了通过磁盘 IO 读取到内存这部分的开销。

2. **单线程模型**

    避免了频繁的上下文切换和竞争锁机制，也不会出现频繁切换线程导致CPU消耗，不会存在多线程的死锁等一系列问题。

    **单线程指的是 Redis 键值对读写请求的执行是单线程**。Redis 服务在执行一些其他命令时就会使用多线程，对于 Redis 的持久化、集群数据同步、异步删除的指令如 `UNLINK`、`FLUSHALL ASYNC`、`FLUSHDB ASYNC` 等非阻塞的删除操作。

3. **I/O 多路复用模型**

    **Redis 采用 I/O 多路复用技术，并发处理连接。**
    Redis 作为一个内存服务器，它需要处理很多来自外部的网络请求，它**使用 I/O 多路复用机制同时监听多个文件描述符的可读和可写状态，一旦受到网络请求就会在内存中快速处理**，由于绝大多数的操作都是纯内存的，所以处理的速度会非常地快。

4. **高效的数据结构**

    redis 共有 string、list、hash、set、sortedset 五种数据机构

    **SDS 简单动态字符串**

    1 .SDS 中 `len` 保存这字符串的长度，`O(1)` 时间复杂度查询字符串长度信息。  
    2 .**空间预分配**：SDS 被修改后，程序不仅会为 SDS 分配所需要的必须空间，还会分配额外的未使用空间。  
    3 .**惰性空间释放**：当对 SDS 进行缩短操作时，程序并不会回收多余的内存空间，而是使用 free 字段将这些字节数量记录下来不释放，后面如果需要 append 操作，则直接使用 free 中未使用的空间，减少了内存的分配。

    **zipList 压缩列表**

    压缩列表是 List 、hash、 sorted Set 三种数据类型底层实现之一。  
    当一个列表只有少量数据的时候，并且每个列表项要么就是小整数值，要么就是长度比较短的字符串，那么 Redis 就会使用压缩列表来做列表键的底层实现。

    **quicklist**

    **首先在列表元素较少的情况下会使用一块连续的内存存储**，这个结构是 `ziplist`，也就是**压缩列表**。  
    **它将所有的元素紧挨着一起存储，*分配的是一块连续的内存*。当数据量比较多的时候才会改成 `quicklist`**。
    因为普通的链表需要的附加指针空间太大，会比较浪费空间，而且会加重内存的碎片化。

    **skipList 跳跃表**

    sorted set 类型的排序功能便是通过「跳跃列表」数据结构来实现。  
    跳跃表（skiplist）是一种有序数据结构，它通过在每个节点中维持多个指向其他节点的指针，从而达到快速访问节点的目的。  
    跳表在链表的基础上，增加了多层级索引，通过索引位置的几个跳转，实现数据的快速定位

    **IntSet 小整数集合**

    当 `set` 集合容纳的元素都是整数并且元素个数较小时，Redis 会使用 `intset` 来存储结合元素。intset 是紧凑的数组结构，同时支持 16 位、32 位和 64 位整数。

5. **简单的 RESP 通信协议**

    RESP 是 Redis 序列化协议的简写。它是一种直观的文本协议，优势在于实现异常简单，解析性能极好。

    Redis 协议将传输的结构数据分为 5 种最小单元类型，单元结束时统一加上回车换行符号 `\r\n`。

    1.**单行字符串**以 + 符号开头。  
    2.**多行字符串** 以 $ 符号开头，后跟字符串长度。  
    3.**整数值**以 : 符号开头，后跟整数的字符串形式。  
    4.**错误消息**以 - 符号开头。  
    5.**数组**以 * 号开头，后跟数组的长度。  

#### 为什么选择单线程？

1. 使用单线程模型能带来更好的可维护性，方便开发和调试
2. 使用单线程模型也能并发的处理客户端的请求
3. Redis 服务中运行的绝大多数操作的性能瓶颈都不是 CPU

## 2. 过期策略

1. **定时删除**  
    redis 会将每个设置了过期时间的 key 放入到一个独立的字典中，以后会定时遍历这个字典来删除到期的 key。

2. **惰性删除**  
    惰性策略就是在客户端访问这个 key 的时候，redis 对 key 的过期时间进行检查，如果过期了就立即删除。
    **定时删除是集中处理，惰性删除是零散处理。**

3. **定期删除**  
    Redis 默认会每 100ms 进行一次过期扫描，过期扫描不会遍历过期字典中所有的 key，而是采用了一种简单的**贪心策略**。

    1. 从过期字典中随机 20 个 key
    2. 删除这 20 个 key 中已经过期的 key
    3. **如果过期的 key 比率超过 1/4，那就重复步骤 1**

    同时，为了保证过期扫描不会出现循环过度，导致线程卡死现象，**算法还增加了扫描时间的上限，默认不会超过 25ms。**

### 对比

**定时删除**对内存友好，能够在 key 过期后立即从内存中删除，但是对 CPU 不友好，如果过期键较多会占用 CPU 对一些时间

**惰性删除**对 CPU 友好，只有在键用到的时候才会进行检查，对于很多用不到的 key 不用浪费时间进行检查，但是对内存不友好，过期 key 如果一直没用到就会一直在内存中，内存就得不到释放，从而造成内存泄漏。

**定期删除**可以通过限制操作时长和频率减少删除对 CPU 的影响，同时也能释放过期 key 占用的内存；但是频率和时长不太好控制，执行频繁了和定时一样占用 CPU，执行太少和惰性删除又一样对内存不好。

一般会使用组合策略 惰性删除 和 定期删除 组合使用。

#### 从库的过期策略

**从库不会进行过期扫描，从库对过期的处理是被动的**。主库在 key 到期时，会在 AOF 文件里增加一条 `del` 指令，同步到所有的从库，从库通过执行这条 `del` 指令来删除过期的 key。

#### Redis 为什么要懒惰删除(lazy free)？

删除指令 `del` 会直接释放对象的内存，大部分情况下，这个指令非常快，没有明显延迟。不过如果删除的 key 是一个非常大的对象，比如一个包含了千万元素的 hash，那么删除操作就会导致单线程卡顿。

Redis 为了解决这个卡顿问题，**在 4.0 版本引入了 `unlink` 指令，它能对删除操作进行懒处理，丢给后台线程来异步回收内存**。

```shell
> unlink key
OK
```

#### flush

Redis 提供了 `flushdb` 和 `flushall` 指令，用来清空数据库，这也是极其缓慢的操作。Redis 4.0 同样给这两个指令也带来了异步化，在指令后面增加 `async` 参数就可以将整棵大树连根拔起，扔给后台线程慢慢焚烧。

```shell
> flushall async
OK
```

#### 异步队列

主线程将对象的引用从「大树」中摘除后，会将这个 key 的内存回收操作包装成一个任务，塞进异步任务队列，后台线程会从这个异步队列中取任务。任务队列被主线程和异步线程同时操作，所以必须是一个线程安全的队列。
![image](https://mail.wangkekai.cn/7428EC70-56B7-46E4-8A16-6F5683DD751A.png)

不是所有的 `unlink` 操作都会延后处理，如果对应 key 所占用的内存很小，延后处理就没有必要了，这时候 Redis 会将对应的 key 内存立即回收，跟 `del` 指令一样。

#### AOF Sync也很慢

*Redis需要每秒一次(可配置)同步 AOF 日志到磁盘，确保消息尽量不丢失*，需要调用 `sync` 函数，这个操作会比较耗时，会导致主线程的效率下降，所以 Redis 也将这个操作移到异步线程来完成。**执行 AOF Sync 操作的线程是一个独立的异步线程，和前面的懒惰删除线程不是一个线程，同样它也有一个属于自己的任务队列，队列里只用来存放 AOF Sync 任务**。

#### 更多异步删除点

Redis 回收内存除了 `del` 指令和 `flush` 之外，还会存在于在 key 的过期、LRU 淘汰、rename 指令以及从库全量同步时接受完 rdb 文件后会立即进行的 flush 操作。

Redis4.0 为这些删除点也带来了异步删除机制，打开这些点需要额外的配置选项。

1. `slave-lazy-flush` 从库接受完 rdb 文件后的 flush 操作
2. `lazyfree-lazy-eviction` 内存达到 maxmemory 时进行淘汰
3. `lazyfree-lazy-expire` key 过期删除
4. `lazyfree-lazy-server-del` rename 指令删除 destKey

## 3. 内存淘汰策略

为了限制最大使用内存，Redis 提供了配置参数 `maxmemory` 来限制内存超出期望大小。
Redis 提供了几种可选策略 (`maxmemory-policy`) 来让用户自己决定该如何腾出新的空间以继续提供读写服务

1. **noeviction 不删除策略**  
    不会继续服务写请求 (DEL 请求可以继续服务)，读请求可以继续进行。这样可以保证不会丢失数据，但是会让线上的业务不能持续进行。这是默认的淘汰策略。

2. **volatile-lru 尝试淘汰设置了过期时间的 key**  
    最少使用的 key 优先被淘汰。没有设置过期时间的 key 不会被淘汰，这样可以保证需要持久化的数据不会突然丢失。

3. **volatile-ttl 除了淘汰的策略不是 LRU**  
    而是 key 的剩余寿命 ttl 的值，ttl 越小越优先被淘汰。

4. **volatile-random**  
    淘汰的 key 是过期 key 集合中随机的 key。

5. **allkeys-lru**  
    区别于 volatile-lru，这个策略要淘汰的 key 对象是全体的 key 集合，而不只是过期的 key 集合。这意味着没有设置过期时间的 key 也会被淘汰。

6. **allkeys-random**  
    跟上面一样，不过淘汰的策略是随机的 key。

### LRU 算法

实现 LRU 算法除了需要 key/value 字典外，还需要附加一个链表，链表中的元素按照一定的顺序进行排列。当空间满的时候，会踢掉链表尾部的元素。当字典的某个元素被访问时，它在链表中的位置会被移动到表头。所以链表的元素排列顺序就是元素最近被访问的时间顺序。

位于链表尾部的元素就是不被重用的元素，所以会被踢掉。位于表头的元素就是最近刚刚被人用过的元素，所以暂时不会被踢。

### LRU 热度

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

### 近似 LRU 算法

**Redis 使用的是一种近似 LRU 算法**，它跟 LRU 算法还不太一样。之所以不使用 LRU 算法，是因为需要消耗大量的额外的内存，需要对现有的数据结构进行较大的改造。使用链表加hash(O(1))可以快速删除。

*近似 LRU 算法则很简单，在现有数据结构的基础上使用随机采样法来淘汰元素，能达到和 LRU 算法非常近似的效果*。Redis 为实现近似 LRU 算法，它给每个 key 增加了一个额外的小字段，**这个字段的长度是 24 个 bit，也就是最后一次被访问的时间戳**，相当于是字段热度。

**处理 key 过期方式分为集中处理和懒惰处理，LRU 淘汰不一样，它的处理方式只有懒惰处理**。当 Redis 执行写操作时，发现内存超出 maxmemory，就会执行一次 LRU 淘汰算法。这个算法也很简单，就是随机采样出 5(可以配置) 个 key，然后淘汰掉最旧的 key，如果淘汰后内存还是超出 maxmemory，那就继续随机采样淘汰，直到内存低于 maxmemory 为止。

采样按照 `maxmemory-policy` 的配置，如果是 `allkeys` 就是从所有的 key 字典中随机，如果是 `volatile` 就从带过期时间的 key 字典中随机。每次采样多少个 key 看的是 `maxmemory_samples` 的配置，默认为 5。

同时 Redis3.0 在算法中增加了淘汰池，进一步提升了近似 LRU 算法的效果。  
淘汰池是一个数组，它的大小是 `maxmemory_samples`，在每一次淘汰循环中，新随机出来的 key 列表会和淘汰池中的 key 列表进行融合，淘汰掉最旧的一个 key 之后，保留剩余较旧的 key 列表放入淘汰池中留待下一个循环。

在 LRU 模式下，lru 字段存储的是 Redis 时钟 `server.lruclock`，Redis 时钟是一个 24bit 的整数，默认是 Unix 时间戳对 2^24 取模的结果，大约 97 天清零一次。当某个 key 被访问一次，它的对象头的 lru 字段值就会被更新为 `server.lruclock`。

默认 Redis 时钟值每毫秒更新一次，在定时任务 `serverCron` 里主动设置。**Redis 的很多定时任务都是在 `serverCron` 里面完成的，比如大型 hash 表的渐进式迁移、过期 key 的主动淘汰、触发 `bgsave、bgaofrewrite` 等等**。

如果 server.lruclock 没有折返 (对 2^24 取模)，它就是一直递增的，这意味着对象的 LRU 字段不会超过 server.lruclock 的值。如果超过了，说明 server.lruclock 折返了。通过这个逻辑就可以精准计算出对象多长时间没有被访问——对象的空闲时间。

![image](https://mail.wangkekai.cn/8A54445C-051E-4785-9A70-AC4C98B0915C.png)

### LFU

Redis 4.0 里引入了一个新的淘汰策略 —— LFU 模式，全称是 Least Frequently Used，表示按最近的访问频率进行淘汰，它比 LRU 更加精准地表示了一个 key 被访问的热度。

在 LFU 模式下，lru 字段 24 个 bit 用来存储两个值，分别是 `ldt(last decrement time)` 和 `logc(logistic counter)`。

![image](https://mail.wangkekai.cn/8850C21C-D38C-4135-85FB-FFA563475C4E.png)

`logc` 是 8 个 bit，**用来存储访问频次**，因为 8 个 bit 能表示的最大整数值为 255，存储频次肯定远远不够，*所以这 8 个 bit 存储的是频次的对数值，并且这个值还会随时间衰减*。如果它的值比较小，那么就很容易被回收。为了确保新创建的对象不被回收，新对象的这 8 个 bit 会初始化为一个大于零的值，默认是 `LFU_INIT_VAL=5`。

![image](https://mail.wangkekai.cn/59B5EFFB-E08E-499A-92BF-E6EFB0AAE521.png)

ldt 是 16 个位，用来存储上一次 logc 的更新时间，因为只有 16 位，所以精度不可能很高。它取的是分钟时间戳对 2^16 进行取模，大约每隔 45 天就会折返。同 LRU 模式一样，我们也可以使用这个逻辑计算出对象的空闲时间，只不过精度是分钟级别的。图中的 `server.unixtime` 是当前 redis 记录的系统时间戳，和 `server.lruclock` 一样，它也是每毫秒更新一次。

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

Redis 4.0 给淘汰策略配置参数 maxmemory-policy 增加了 2 个选项，分别是 `volatile-lfu` 和 `allkeys-lfu`，分别是对带过期时间的 key 进行 lfu 淘汰以及对所有的 key 执行 lfu 淘汰算法。打开了这个选项之后，就可以使用 object freq 指令获取对象的 lfu 计数值了。

```shell
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

## 持久化

Redis 的持久化机制有两种，*第一种是快照，第二种是 AOF 日志*。**快照是一次全量备份，AOF 日志是连续的增量备份**。

***快照*是内存数据的二进制序列化形式，在存储上非常紧凑，而 *AOF 日志记录* 的是内存数据修改的指令记录文本**。

AOF 日志在长期的运行过程中会变的无比庞大，数据库重启时需要加载 AOF 日志进行指令重放，这个时间就会无比漫长。所以需要定期进行 AOF 重写，给 AOF 日志进行瘦身。

