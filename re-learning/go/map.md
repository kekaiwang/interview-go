# map

## 什么是 map

- **删除 map 不存在的键值对时，不会报错，相当于没有任何作用**
- **获取不存在的键值对时，返回值类型对应的零值，所以返回 0**

### map 结构 hmap

`map` 类型的变量本质上是一个 `hmap` 类型的指针：

```go
type hmap struct {
    count     int    // 已经存储的键值对个数
    flags     uint8
    B         uint8  // 常规桶个数等于2^B; map底层的哈希表通过与运算的方式选择桶
    noverflow uint16 // 使用的溢出桶数量
    hash0     uint32 // hash seed
    buckets    unsafe.Pointer // 常规桶起始地址
    oldbuckets unsafe.Pointer // 扩容时保存原来常规桶的地址
    nevacuate  uintptr        // 渐进式扩容时记录下一个要被迁移的旧桶编号

    extra *mapextra
}

// 溢出桶相关信息
type mapextra struct {
    overflow    *[]*bmap // 把已经用到的溢出桶链起来
    oldoverflow *[]*bmap // 渐进式扩容时，保存旧桶用到的溢出桶
    nextOverflow *bmap   // 下一个尚未使用的溢出桶
}
```

1. `count` 表示当前哈希表中的元素数量；
2. `B` 表示当前哈希表持有的 `buckets` 数量，但是因为哈希表中桶的数量都 2 的倍数，所以该字段会存储对数，也就是 `len(buckets) == 2^B`；
3. `hash0` 是哈希的种子，它能为哈希函数的结果引入随机性，这个值在创建哈希表时确定，并在调用哈希函数时作为参数传入；
4. `oldbuckets` 是**哈希在扩容时用于保存之前 buckets 的字段**，它的大小是当前 `buckets` 的一半；
5. `nevacuate` 哈希扩容时迁移的进度，也就是下一个要迁移的位置
6. `extra` 溢出桶的相关信息

map使用的桶很有设计感，每个桶里可以存储8个键值对，并且为了内存使用更加紧凑，**8个键放一起，8个值放一起**。
**对应每个key只保存其哈希值的高8位（`tophash`）**。
**每个键值对的 tophash、key 和 value 的索引顺序一一对应**。这就是 map 使用的桶的内存布局。

```go
type bmap struct {
    tophash [bucketCnt]uint8
}
```

在运行期间，`runtime.bmap` 结构体其实不止包含 `tophash` 字段，因为哈希表中可能存储不同类型的键值对，而且 Go 语言也不支持泛型，所以键值对占据的内存空间大小只能在编译时进行推导。`runtime.bmap` 中的其他字段在运行时也都是通过计算内存地址的方式访问的，所以它的定义中就不包含这些字段，不过我们能根据编译期间的 cmd/compile/internal/gc.bmap 函数重建它的结构：

```go
type bmap struct {
    topbits  [8]uint8
    keys     [8]keytype
    values   [8]valuetype
    pad      uintptr
    overflow uintptr
}
```

#### 溢出桶

- 当桶的数量小于 2^4 时，由于数据较少、使用溢出桶的可能性较低，会省略创建的过程以减少额外开销；
- 当桶的数量多于 2^4 时，会额外创建 2^𝐵−4 个溢出桶；

溢出桶的内存布局与之前介绍的常规桶相同。
**如果哈希表要分配的桶的数目大于 `2^4`，就会预分配 `2^(B-4)` 个溢出桶备用**。这些常规桶和溢出桶在内存中是连续的，只是前 2^B 个用作常规桶，后面的用作溢出桶。

如果当前桶存满了以后，检查 `hmap.extra.nextoverflow` 还有可用的溢出桶，就在这个桶后面链上这个溢出桶，然后继续往这个溢出桶里存。而 `hmap.extra.nextoverflow` 继续指向下一个空闲的溢出桶。所以这里解决哈希冲突的方式应该属于拉链法。

![image](https://mmbiz.qpic.cn/mmbiz_png/ibjI8pEWI9L4iaD5VQHquw0dxSYrziavjtiaahHtecXecef74PXZBAc5AibkynD3rDbJSsicwOtgibyR1zAljN5VERt4w/640?wx_fmt=png&tp=webp&wxfrom=5&wx_lazy=1&wx_co=1)

**Go语言中可以通过 `==` 来比较是否相等的类型，都可以作为 map 的 key 类型。**

#### 字面量初始化

```go
hash := map[string]int{
    "1": 2,
    "3": 4,
    "5": 6,
}
```

1. **当哈希表中的元素数量少于或者等于 25 个时**，编译器会将字面量初始化的结构体将所有的键值对一次加入到哈希表中。
2. **一旦哈希表中元素的数量超过了 25 个**，编译器会创建两个数组分别存储键和值，这些键值对会通过如下所示的 for 循环加入哈希

```go
hash := make(map[string]int, 26)
vstatk := []string{"1", "2", "3", ... ， "26"}
vstatv := []int{1, 2, 3, ... , 26}
for i := 0; i < len(vstak); i++ {
    hash[vstatk[i]] = vstatv[i]
}
```

### hash 桶选择

假设哈希值为 `hash`，桶数量为 `m`。通常我们有两种方法来选择一个桶:

1. **取模法** `hash%m`
2. **与运算** `hash&(m-1)`

第二种更加高效，**但限制了 m 必须要是2的整数次幂，这样才能保证与运算结果落在 `[0,m-1]`，而不会出现有些桶注定不会被选中的情况**。
**所以 `hmap` 中并不直接记录桶的个数，而是记录这个数目是2的多少次幂**。

### hash 扩容

1. 如果超过负载因子(默认6.5)就触发**翻倍扩容**；

    `hmap.count / 2^hmap.B > 6.5`

    分配新桶数目是旧桶的 2 倍，`hmap.oldbuckets` 指向旧桶，`hmap.buckets` 指向新桶。`hmap.nevacuate` 为0，表示接下来要迁移编号为0的旧桶。

    **然后通过渐进式扩容的方式，每次读写map时检测到当前map处在扩容阶段（`hmap.oldbuckets != nil`），就执行一次迁移工作，把编号为 `hmap.nevacuate` 的旧桶迁移到新桶**，每个旧桶的键值对都会分流到两个新桶中。

    编号为 `hmap.nevacuate` 的旧桶迁移结束后会增加这个编号值，直到所有旧桶迁移完毕，把 `hmap.oldbuckets` 置为 nil，一次翻倍扩容结束。

2. 如果没有超过设置的负载因子上限，但是使用的溢出桶较多，也会触发扩容，不过这一次是**等量扩容**
    - 如果常规桶数目不大于 `2^15`，那么使用的溢出桶数目超过常规桶就算是多了；
    - 如果常规桶数目大于 `2^15`，那么使用溢出桶数目一旦超过 `2^15` 就算多了

    桶的负载因子没有超过上限值，却偏偏使用了很多溢出桶呢？因为是有很多键值对被删除的情况。
    如果把这些键值对重新安排到等量的新桶中，虽然哈希值没变，常规桶数目没变，每个键值对还是会选择与旧桶一样的新桶编号，**但是能够存储的更加紧凑，进而减少溢出桶的使用**。

**正是因为扩容过程中会发生键值对迁移，键值对的地址也会发生改变，所以才说 map 的元素是不可寻址的，如果要取一个 value 的地址则不能通过编译。**

### map 遍历

```go
// 生成随机数 r
r := uintptr(fastrand())
if h.B > 31-bucketCntBits {
    r += uintptr(fastrand()) << 31
}

// 从哪个 bucket 开始遍历
it.startBucket = r & (uintptr(1)<<h.B - 1)
// 从 bucket 的哪个 cell 开始遍历
it.offset = uint8(r >> h.B & (bucketCnt - 1))
```

例如，B = 2，那 uintptr(1)<<h.B - 1 结果就是 3，低 8 位为 0000 0011，将 r 与之相与，就可以得到一个 0~3 的 bucket 序号；bucketCnt - 1 等于 7，低 8 位为 0000 0111，将 r 右移 2 位后，与 7 相与，就可以得到一个 0~7 号的 cell。

首先根据 B 进行位运算得到起始 Buckets，然后与 7 进行位运算得到起始槽位 cell，然后开始进行遍历，最后再遍历当前桶槽位之前的 cell，如果是在扩容阶段则需要进行老 Buckets 是否迁移完成的判断，如果迁移完成直接进行遍历，没有进行判断是否是加倍扩容，加倍扩容则需要拿出到当前新桶的元素进行迁移，然后继续进行，知道返回起始桶的 cell 前位置。

### map 插入或更新

对 key 计算 hash 值，根据 hash 值按照之前的流程，找到要赋值的位置（可能是插入新 key，也可能是更新老 key），对相应位置进行赋值。

**核心还是一个双层循环，外层遍历 `bucket` 和它的 `overflow bucket`，内层遍历整个 `bucket` 的各个 cell**。

**插入的大致过程**：

1. **首先会检查 map 的标志位 `flags`**。**如果 `flags` 的写标志位此时被置 1 了，说明有其他协程在执行“写”操作，进而导致程序 panic**。
这也说明了 map 对协程是不安全的。

2. **定位 map key**

   - 如果 map 处在扩容的过程中，那么当 key 定位到了某个 bucket 后，需要确保这个 bucket 对应的老 bucket 完成了迁移过程。**即老 bucket 里的 key 都要迁移到新的 bucket 中来（分裂到 2 个新 bucket），才能在新的 bucket 中进行插入或者更新的操作**。

   - 准备两个指针，一个（`inserti`）指向 `key` 的 hash 值在 `tophash` 数组所处的位置，另一个(insertk)指向 `cell` 的位置（也就是 `key` 最终放置的地址），当然，对应 `value` 的位置就很容易定位出来了。这三者实际上都是关联的，在 `tophash` 数组中的索引位置决定了 `key` 在整个 bucket 中的位置（共 8 个 key），而 `value` 的位置需要“跨过” 8 个 key 的长度。

   - 在循环的过程中，`inserti` 和 `insertk` 分别指向第一个找到的空闲的 `cell`。如果之后在 `map` 没有找到 `key` 的存在，也就是说原来 `map` 中没有此 `key`，这意味着插入新 `key`。那最终 `key` 的安置地址就是第一次发现的“空位”（tophash 是 empty）。

   - 如果这个 `bucket` 的 8 个 `key` 都已经放置满了，那在跳出循环后，发现 `inserti` 和 `insertk` 都是空，这时候需要在 `bucket` 后面挂上 `overflow bucket`。当然，也有可能是在 `overflow bucket` 后面再挂上一个 `overflow bucket`。这就说明，太多 `key` `hash` 到了此 `bucket`。

3. **最后，会更新 map 相关的值**，如果是插入新 key，map 的元素数量字段 `count` 值会加 1；在函数之初设置的 `hashWriting` 写标志处会清零。

### map 删除

1. 它首先会检查 `h.flags` 标志，如果发现写标位是 1，直接 `panic`，因为这表明有其他协程同时在进行写操作。

2. 计算 `key` 的哈希，找到落入的 bucket。检查此 `map` 如果正在扩容的过程中，直接触发一次搬迁操作。
删除操作同样是两层循环，核心还是找到 key 的具体位置。寻找过程都是类似的，在 bucket 中挨个 cell 寻找。

3. 找到对应位置后，对 `key` 或者 `value` 进行“清零”操作：

4. 最后，将 `count` 值减 1，将对应位置的 `tophash` 值置成 `Empty`。

### map 总结

- **无法对 map 的key 和 value 取地址**
- **在查找、赋值、遍历、删除的过程中都会检测写标志**，一旦发现写标志置位（等于1），则直接 panic。赋值和删除函数在检测完写标志是复位之后，先将写标志位置位，才会进行之后的操作。
- **`map` 的 `value` 本身是不可寻址的**

    ```go
    type Student struct {
        Name string
    }

    func main() {
        student := map[string]*Student{"name": {"test"}}
        student["name"].Name = "a"
        fmt.Println(student["name"])
    }
    ```

    因为 `map` 中的值会在内存中移动，并且旧的指针地址在 `map` 改变时会变得⽆效。故如果需要修改 `map` 值，可以将 `map` 中的⾮指针类型 `value` ，修改为指针类型，⽐如使⽤ `map[string]*Student` 。

   **`student["name"].Name = "a"` 不能直接进行赋值，`student["name"]` 返回的是两个参数，可以先通过变量接收后进行赋值修改。**

- **map 是并发读写不安全的**

    ```go
    type UserAges struct {
        ages map[string]int
        sync.Mutex
    }

    func (ua *UserAges) Add(name string, age int) {
        ua.Lock()
        defer ua.Unlock()
        ua.ages[name] = age
    }
    func (ua *UserAges) Get(name string) int {
        if age, ok := ua.ages[name]; ok {
            return age
        }
        return -1
    }
    ```

    在执⾏ `Get` ⽅法时可能被 `panic`。
    虽然有使⽤ `sync.Mutex` 做写锁，但是 `map` 是并发读写不安全的。**`map` 属于引⽤类型，并发读写时多个协程⻅是通过指针访问同⼀个地址，即访问共享变量，此时同时读写资源存在竞争关系**。会报错误信息:`“fatal error: oncurrent map read and map write”`。

    因此，在 `Get` 中也需要加锁，因为这⾥只是读，建议使⽤读写 `sync.RWMutex` 。

- **sync.map 没有 len 方法**

- **map panic 无法 recover**
    当并发写入时会抛出 throw -> fatalthrow -> exit(2)
