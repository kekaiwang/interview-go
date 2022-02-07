# go

[toc]

## 1. 常见基础题

### 1.1 make 和 new 的区别

- **`make`只用于内建类型 `map、slice 和channel` 的内存分配**, 返回初始化后的（非零）值。
    make只能创建 `slice`、`map` 和 `channel`，并且返回一个有初始值(非零)的T类型，而不是 `*T`。对于 `slice`、`map` 和 `channel` 来说，`make` 初始化了内部的数据结构，填充适当的值。

  - 向一个为 `nil` 的 `map` 读写数据，将会导致 `panic`，使用 `make` 可以指定 `map` 的初始空间大小，以容纳元素。如果未指定，则初始空间比较小。
  - 向一个为 `nil` 的 `chan` 读写数据，会导致 `deadlock!`，使用 `make` 可以初始化`chan`，并指定 `chan` 是缓存 `chan（make(chan T,size)）`，还是非缓存 `chan（make(chan T)）`
  - 使用 `make` 可以在创建 `slice` 时，定义切片的 `len`、`cap` 的大小
    使用 `make` 可以创建一个非零值的引用对象

- **`new` 用于各种类型的内存分配**，new 返回指针。
    分配一片零值的内存空间，并返回指向这片内存空间的指针 `value *T`

### 1.2 提高性能的写法

1. 可以使用 CAS，则使用 CAS 操作
2. 针对热点代码要做针对性优化
3. 不要忽略 GC 的影响，尤其是高性能低延迟的服务
4. 合理的对象复用可以取得非常好的优化效果
5. 尽量避免反射，在高性能服务中杜绝反射的使用
6. 有些情况下可以尝试调优 “GOGC”参数
7. 新版本稳定的前提下，尽量升级新的 Go 版本，因为旧版本永远不会变得更好

### 1.3 结构体能比较吗？

- Go 结构体有时候并不能直接比较，**当其基本类型包含：`slice`、`map`、`function` 时，是不能比较的**。若强行比较，就会导致出现例子中的直接报错的情况。
- 指针引用，其虽然都是 `new(string)`，从表象来看是一个东西，但其具体返回的地址是不一样的
    可以使用反射方法 `reflect.DeepEqual` 进行比较

    ```go
    // 如果是指针类型，可以这样比较
    type Gender Struct {
        Name string
        Gender *string
    }
    func main() {
        gender := new(string)
        v1 := Value{Name: "Viper", Gender: gender}
        v2 := Value{Name: "Viper", Gender: gender}
        ...
    }
    ```

**反射比较方法 `reflect.DeepEqual` 常用于判定两个值是否深度一致**，其规则如下：

- 指针类型比较的是指针地址，非指针类型比较的是每个属性的值
- 相同类型的值是深度相等的，不同类型的值永远不会深度相等。
- 当数组值（array）的对应元素深度相等时，数组值是深度相等的。
- 当结构体（struct）值如果其对应的字段（包括导出和未导出的字段）都是深度相等的，则该值是深度相等的。
- 当函数（func）值如果都是零，则是深度相等；否则就不是深度相等。
- 当接口（interface）值如果持有深度相等的具体值，则深度相等。

### 1.4 gin 框架优势

- **快速轻量**，基于 Radix 树的路由，小内存占用。没有反射。可预测的 API 性能。

- **支持中间件**，新建一个中间件非常简单，传入的 HTTP 请求可以由一系列中间件和最终操作来处理。例如：Logger，Authorization，GZIP，最终操作 DB。

- **Crash 处理**，Gin 可以 catch 一个发生在 HTTP 请求中的 panic 并 recover 它。这样，你的服务器将始终可用。例如，你可以向 Sentry 报告这个 panic！

- **JSON 验证**，Gin 可以解析并验证请求的 JSON，例如检查所需值的存在。

- **路由组**，更好地组织路由。是否需要授权，不同的 API 版本…… 此外，这些组可以无限制地嵌套而不会降低性能。

- **错误管理**，Gin 提供了一种方便的方法来收集 HTTP 请求期间发生的所有错误。最终，中间件可以将它们写入日志文件，数据库并通过网络发送。

- **内置渲染**，Gin 为 JSON，XML 和 HTML 渲染提供了易于使用的 API。

### 1.5 gRPC VS Restful

- **文档规范**，gRPC使用proto文件编写接口（API），文档规范比Restful更好，因为proto文件的语法和形式是定死的，所以更为严谨、风格统一清晰；而Restful由于可以使用多种工具进行编写（只要人看得懂就行），每家公司、每个人的攥写风格又各有差异，难免让人觉得比较混乱。

- **消息编码**，消息编码这块，gRPC使用protobuf进行消息编码，而Restful一般使用JSON进行编码

- **传输协议**，gRPC使用HTTP/2作为底层传输协议，而RestFul则使用 HTTP 或则其他协议 https/tcp等。

- **传输性能**，由于 gRPC 使用 protobuf 进行消息编码（即序列化），而经 protobuf 序列化后的消息体积很小（传输内容少，传输相对就快）；再加上HTTP/2协议的加持（HTTP1.1的进一步优化），使得gRPC的传输性能要优于Restful。

- **传输形式**，gRPC最大的优势就是支持流式传输，传输形式具体可以分为四种（unary、client stream、server stream、bidirectional stream），这个后面我们会讲到；而Restful是不支持流式传输的。

- 浏览器的支持度
不知道是不是gRPC发展较晚的原因，目前浏览器对gRPC的支持度并不是很好，而对Restful的支持可谓是密不可分，这也是gRPC的一个劣势，如果后续浏览器对gRPC的支持度越来越高，不知道gRPC有没有干饭Restful的可能呢？

- **消息的可读性和安全性**，由于gRPC序列化的数据是二进制，且如果你不知道定义的Request和Response是什么，你几乎是没办法解密的，所以gRPC的安全性也非常高，但随着带来的就是可读性的降低，调试会比较麻烦；而Restful则相反（现在有HTTPS，安全性其实也很高）

- **代码的编写**，由于gRPC调用的函数，以及字段名，都是使用stub文件的，所以从某种角度看，代码更不容易出错，联调成本也会比较低，不会出现低级错误，比如字段名写错、写漏。

总的来说：

1. **gRPC主要用于公司内部的服务调用，性能消耗低，传输效率高，服务治理方便。**
2. **Restful主要用于对外，比如提供接口给前端调用，提供外部服务给其他人调用等，**

### 1.5 context

#### 1.5.1 context 引入

多并发情况下：

- 使用 waitgroup 等待协程结束
  - 优点是使用等待组的并发控制模型，尤其适用于好多个goroutine协同做一件事情的时候，因为每个goroutine做的都是这件事情的一部分，只有全部的 goroutine 都完成，这件事情才算完成；
  - 这种方式的缺陷：在实际生产中，需要我们主动的通知某一个 goroutine 结束。
- 使用通道 channel + select
  - 优点：比较优雅，
  - 劣势：如果有很多 goroutine 都需要控制结束怎么办？，如果这些 goroutine 又衍生了其它更多的goroutine 怎么办？

context 主要用来在 goroutine 之间传递上下文信息，包括：**取消信号、超时时间、截止时间、k-v** 等。

1. `Deadline` — 返回 `context.Context` 被取消的时间，也就是完成工作的截止日期；
2. `Done` — 返回一个 Channel，这个 Channel 会在当前工作完成或者上下文被取消后关闭，多次调用 `Done` 方法会返回同一个 Channel；
3. `Err` — 返回 `context.Context` 结束的原因，它只会在 `Done` 方法对应的 Channel 关闭时返回非空的值；
    - 如果 `context.Context` 被取消，会返回 `Canceled` 错误；
    - 如果 `context.Context` 超时，会返回 `DeadlineExceeded` 错误；
4. `Value` — 从 `context.Context` 中获取键对应的值，对于同一个上下文来说，多次调用 `Value` 并传入相同的 `Key` 会返回相同的结果，该方法可以用来传递请求特定的数据；

- `Background()` 主要用于 `main` 函数、初始化以及测试代码中，作为 `Context` 这个树结构的最顶层的 Context，也就是根 Context。
- `TODO()`，它目前还不知道具体的使用场景，如果我们不知道该使用什么 Context 的时候，可以使用这个。

它们两个本质上都是 `emptyCtx` 结构体类型，是一个不可取消，没有设置截止时间，没有携带任何值的 Context。

#### 1.5.2 Context的继承衍生

```go
func WithCancel(parent Context) (ctx Context, cancel CancelFunc)
func WithDeadline(parent Context, deadline time.Time) (Context, CancelFunc)
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc)
func WithValue(parent Context, key, val interface{}) Context
```

这四个 `With` 函数，接收的都有一个 `partent` 参数，就是父 Context，我们要基于这个父 `Context` 创建出子 Context 的意思，这种方式可以理解为子 Context 对父 Context 的继承，也可以理解为基于父 Context 的衍生。通过这些函数，就创建了一颗 Context 树，树的每个节点都可以有任意多个子节点，节点层级可以有任意多个。

- `WithCancel` 函数，传递一个父 Context 作为参数，返回子 Context，以及一个取消函数用来取消 Context。`WithDeadline` 函数，和 `WithCancel` 差不多，它会多传递一个截止时间参数，意味着到了这个时间点，会自动取消 Context，当然我们也可以不等到这个时候，可以提前通过取消函数进行取消。

- `WithTimeout` 和 `WithDeadline` 基本上一样，这个表示是超时自动取消，是多少时间后自动取消 Context 的意思。

- `WithValue` 函数和取消 Context 无关，它是为了生成一个绑定了一个键值对数据的 Context，即给context设置值，这个绑定的数据可以通过 Context.Value 方法访问到.
`WithValue` 创建 context 节点的过程实际上就是创建链表节点的过程。两个节点的 `key` 值是可以相等的，但它们是两个不同的 context 节点。查找的时候，会向上查找到最后一个挂载的 context 节点，也就是离得比较近的一个父节点 context。所以，整体上而言，用 `WithValue` 构造的其实是一个低效率的链表。

上面3个函数都会返回一个取消函数CancelFunc，这是一个函数类型，它的定义非常简单 `type CancelFunc func()`,该函数可以取消一个 Context，以及这个节点 Context下所有的所有的 Context，不管有多少层级。

#### 1.5.3 最佳实践

- 不要将 Context 塞到结构体里。直接将 Context 类型作为函数的第一参数，而且一般都命名为 `ctx`。
- 不要向函数传入一个 `nil` 的 context，如果你实在不知道传什么就用 `context：todo`。
- 不要把本应该作为函数参数的类型塞到 context 中，context 存储的应该是一些共同的数据。例如：登陆的 session、cookie 等。
- 同一个 context 可能会被传递到多个 goroutine，别担心，context 是并发安全的。

### 1.6 map

#### map 遍历

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

#### map 插入或更新

对 key 计算 hash 值，根据 hash 值按照之前的流程，找到要赋值的位置（可能是插入新 key，也可能是更新老 key），对相应位置进行赋值。

**核心还是一个双层循环，外层遍历 `bucket` 和它的 `overflow bucket`，内层遍历整个 `bucket` 的各个 cell**。

**插入的大致过程**：

1. **首先会检查 map 的标志位 `flags`**。**如果 `flags` 的写标志位此时被置 1 了，说明有其他协程在执行“写”操作，进而导致程序 panic**。这也说明了 map 对协程是不安全的。

2. **定位 map key**

   - 如果 map 处在扩容的过程中，那么当 key 定位到了某个 bucket 后，需要确保这个 bucket 对应的老 bucket 完成了迁移过程。**即老 bucket 里的 key 都要迁移到新的 bucket 中来（分裂到 2 个新 bucket），才能在新的 bucket 中进行插入或者更新的操作**。

   - 准备两个指针，一个（`inserti`）指向 `key` 的 hash 值在 `tophash` 数组所处的位置，另一个(insertk)指向 `cell` 的位置（也就是 `key` 最终放置的地址），当然，对应 `value` 的位置就很容易定位出来了。这三者实际上都是关联的，在 `tophash` 数组中的索引位置决定了 `key` 在整个 bucket 中的位置（共 8 个 key），而 `value` 的位置需要“跨过” 8 个 key 的长度。

   - 在循环的过程中，`inserti` 和 `insertk` 分别指向第一个找到的空闲的 `cell`。如果之后在 `map` 没有找到 `key` 的存在，也就是说原来 `map` 中没有此 `key`，这意味着插入新 `key`。那最终 `key` 的安置地址就是第一次发现的“空位”（tophash 是 empty）。

   - 如果这个 `bucket` 的 8 个 `key` 都已经放置满了，那在跳出循环后，发现 `inserti` 和 `insertk` 都是空，这时候需要在 `bucket` 后面挂上 `overflow bucket`。当然，也有可能是在 `overflow bucket` 后面再挂上一个 `overflow bucket`。这就说明，太多 `key` hash 到了此 `bucket`。

3. **最后，会更新 map 相关的值**，如果是插入新 key，map 的元素数量字段 `count` 值会加 1；在函数之初设置的 `hashWriting` 写标志出会清零。

#### map 删除

1. 它首先会检查 `h.flags` 标志，如果发现写标位是 1，直接 `panic`，因为这表明有其他协程同时在进行写操作。

2. 计算 `key` 的哈希，找到落入的 bucket。检查此 `map` 如果正在扩容的过程中，直接触发一次搬迁操作。
删除操作同样是两层循环，核心还是找到 key 的具体位置。寻找过程都是类似的，在 bucket 中挨个 cell 寻找。

3. 找到对应位置后，对 `key` 或者 `value` 进行“清零”操作：

4. 最后，将 `count` 值减 1，将对应位置的 `tophash` 值置成 `Empty`。

#### map 总结

- **无法对 map 的key 和 value 取地址**
- **在查找、赋值、遍历、删除的过程中都会检测写标志**，一旦发现写标志置位（等于1），则直接 panic。赋值和删除函数在检测完写标志是复位之后，先将写标志位置位，才会进行之后的操作。
- **`map` 的 `value` 本身是不可寻址的**

    因为 `map` 中的值会在内存中移动，并且旧的指针地址在 `map` 改变时会变得⽆效。故如果需要修改 `map` 值，可以将 `map` 中的⾮指针类型 `value` ，修改为指针类型，⽐如使⽤ `map[string]*Student` .

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

### 1.7 接口

### 自动检测类型是否实现接口

```go
var _ io.Writer = (*myWriter)(nil)
var _ io.Writer = myWriter{}
```

上述赋值语句会发生隐式地类型转换，在转换的过程中，编译器会检测等号右边的类型是否实现了等号左边接口所规定的函数。

### 1.8 go 内存泄漏

go 中的内存泄露一般都是 goroutine 泄露，就是 goroutine 没有被关闭，或者没有添加超时控制，让goroutine一只处于阻塞状态，不能被 `GC`。

#### 1. 暂时性内存泄露

- 获取长字符串中的一段导致长字符串未释放
- 获取长 `slice` 中的一段导致长 `slice` 未释放
- 在长 `slice` 新建 `slice` 导致泄漏

`string` 相比于切片少了一个容量的 `cap` 字段，可以把 `string` 当成一个只读的切片类型。
**获取长`string` 或切片中的一段内容，由于新生成的对象和老的 `string` 或切片共用一个内存空间，会导致老的 `string` 和切片资源暂时得不到释放，造成短暂的内存泄露**。
比如切片引用很短的一段字符串，但是字符串又足够长，重复使用并长时间不结束，未被引用的部分就造成了泄漏。

#### 2. 永久性内存泄露

- `goroutine` 泄漏
    goroutine 引用的 `channel`，是不会被 gc，并且 channel 会使当前引用的 goroutine 一直阻塞，直到接收到退出的信号。

    1. **发送端 channel 满了**
    goroutine 作为生产者向 channel 发送信息，但是没有消费的 goroutine，或者消费的 goroutine 被错误的关闭了。导致 channel 被打满。
    2. **接收端消费的 channel 为空**
    作为消费者的 goroutine,等待消费 channel，但是上游的生产者不存在.
    3. **生产者消费者异常退出，导致 channel 满了或者 channel 为空**
    作为生产者的 goroutine 如果没有数据发送了，就需要主动退出当前的 goroutine,并且发出退出信号，这样下游消费的 goroutine,才能在 channel 消费完的时候，优雅的退出，不至于阻塞在没有发送者的channel 中。

    作为消费者的 goroutine 一定要在 channel 没数据了，并且上游发送数据的 goroutine 已经退出的情况下，退出。这样，才不至于上游的发送者阻塞到一个没有消费者的 channel 中。
- `time.Ticker` 未关闭导致泄漏
- `Finalizer` 导致泄漏
- `Deferring Function Call` 导致泄漏

### 1.9 go读写文件数据有没有进入系统调用

进入了系统调用

```go
func Create(name string) (*File, error) {
    return OpenFile(name, O_RDWR|O_CREATE|O_TRUNC, 0666)
}

r, e = syscall.Open(name, flag|syscall.O_CLOEXEC, syscallMode(perm))
```

第一个参数是文件名，第二个参数是文件模式，第三个参数是文件权限，默认权限是0666
O_RDWR O_CREATE O_TRUNC是file.go文件中定义好的一些常量，标识文件以什么模式打开，常见的模式有读写，只写，只读，权限依次降低。

看到syscall对于文件的操作进行了封装，继续进入

```go
func Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno)
func Syscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err Errno)
func RawSyscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno)
func RawSyscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err Errno)
```

这里，看到相似的函数有四个，后面没数字的，就是4个参数，后面为6的，就是6个参数。调用的是操作系统封装好的API。

## 4 select

golang 的 select 就是监听 IO 操作，当 IO 操作发生时，触发相应的动作每个case语句里必须是一个IO操作，确切的说，应该是一个面向 channel 的IO操作。

1. select 语句只能用于信道的读写操作
2. select 中的 case 条件(非阻塞)是并发执行的，select 会选择先操作成功的那个 case 条件去执行，**如果多个同时返回，则随机选择一个执行**，此时将无法保证执行顺序。对于阻塞的 case 语句会直到其中有信道可以操作，如果有多个信道可操作，会随机选择其中一个 case 执行
3. 对于 case 条件语句中，如果存在通道值为 `nil` 的读写操作（也就是 `var chan int`），则该分支将被忽略，可以理解为从 `select` 语句中删除了这个 `case` 语句; **并且会报 deadlock**
4. 如果有超时条件语句，`case <-time.After(2 * time.Second)`，判断逻辑为如果在这个时间段内一直没有满足条件的 case,则执行这个超时 case。如果此段时间内出现了可操作的 case,则直接执行这个 case。一般用超时语句代替了 default 语句
5. 对于空的 select{}，会引起死锁
6. 对于 for 中的 select{}, 也有可能会引起cpu占用过高的问题，比如增加一个监听退出信号的case 当前是处于阻塞状态，又加一个 default 分支什么都不做，这时候就会莫名拉高 cpu

### 4.1 直接阻塞

**空的 select 语句会直接阻塞当前 Goroutine，导致 Goroutine 进入无法被唤醒的永久休眠状态。**
当 select 结构中不包含任何 case。它直接将类似 `select {}` 的语句转换成调用 `runtime.block` 函数：
`runtime.block` 的实现非常简单，它会调用 `runtime.gopark` 让出当前 Goroutine 对处理器的使用权并传入等待原因 `waitReasonSelectNoCases`。

### 4.2 单一管道

如果当前的 select 条件只包含一个 `case`，那么编译器会将 select 改写成 `if` 条件语句。
**当 case 中的 Channel 是空指针时，会直接挂起当前 Goroutine 并陷入永久休眠。**

```go
// 改写前
select {
case v, ok <-ch: // case ch <- v
    ...    
}

// 改写后
if ch == nil {
    block()
}
v, ok := <-ch // case ch <- v
...

func block() {
    gopark(nil, nil, waitReasonSelectNoCases, traceEvGoStop, 1)
}
```

### 4.3 非阻塞操作

当 select 中仅包含两个 `case`，并且其中一个是 `default` 时，Go 语言的编译器就会认为这是一次非阻塞的收发操作

#### 发送

首先是 Channel 的发送过程，当 `case` 中表达式的类型是 `OSEND` 时，编译器会使用条件语句和 `runtime.selectnbsend` 函数改写代码：

```go
select {
case ch <- i:
    ...
default:
    ...
}

if selectnbsend(ch, i) {
    ...
} else {
    ...
}

func selectnbsend(c *hchan, elem unsafe.Pointer) (selected bool) {
    return chansend(c, elem, false, getcallerpc())
}
```

这段代码中最重要的就是 `runtime.selectnbsend`，它为我们提供了向 Channel 非阻塞地发送数据的能力。向 Channel 发送数据的 `runtime.chansend` 函数包含一个 `block` 参数，该参数会决定这一次的发送是不是阻塞的;
**由于我们向 `runtime.chansend` 函数传入了非阻塞，所以在不存在接收方或者缓冲区空间不足时，当前 Goroutine 都不会阻塞而是会直接返回。**

#### 接收

返回值数量不同会导致使用函数的不同，两个用于非阻塞接收消息的函数 `runtime.selectnbrecv` 和 `runtime.selectnbrecv2` 只是对 `runtime.chanrecv` 返回值的处理稍有不同。
因为接收方不需要，所以 `runtime.selectnbrecv` 会直接忽略返回的布尔值，而 `runtime.selectnbrecv2` 会将布尔值回传给调用方。
**与 `runtime.chansend` 一样，`runtime.chanrecv` 也提供了一个 block 参数用于控制这次接收是否阻塞。**

```go
func selectnbrecv(elem unsafe.Pointer, c *hchan) (selected bool) {
    selected, _ = chanrecv(c, elem, false)
    return
}

func selectnbrecv2(elem unsafe.Pointer, received *bool, c *hchan) (selected bool) {
    selected, *received = chanrecv(c, elem, false)
    return
}
```

### 4.4 常见流程

在默认的情况下，编译器会使用如下的流程处理 select 语句：

1. 将所有的 case 转换成包含 Channel 以及类型等信息的 `runtime.scase` 结构体；
2. 调用运行时函数 `runtime.selectgo` 从多个准备就绪的 Channel 中选择一个可执行的 `runtime.scase` 结构体；
3. 通过 `for` 循环生成一组 `if` 语句，在语句中判断自己是不是被选中的 case；

#### 初始化

`runtime.selectgo` 函数首先会进行执行必要的初始化操作并决定处理 `case` 的两个顺序 — **轮询顺序 `pollOrder` 和加锁顺序 `lockOrder`**：

- **轮询顺序**：通过 `runtime.fastrandn` 函数引入随机性；
- **加锁顺序**：按照 `Channel` 的地址排序后确定加锁顺序；

**随机的轮询顺序可以避免 Channel 的饥饿问题，保证公平性；而根据 Channel 的地址顺序确定加锁顺序能够避免死锁的发生**。这段代码最后调用的 `runtime.sellock` 会按照之前生成的加锁顺序锁定 `select` 语句中包含所有的 Channel。

#### 循环

当我们为 select 语句锁定了所有 `Channel` 之后就会进入 `runtime.selectgo` 函数的主循环，它会分三个阶段查找或者等待某个 `Channel` 准备就绪：

1. 查找是否已经存在准备就绪的 `Channel`，即可以执行收发操作；
    **主要职责是查找所有 `case` 中是否有可以立刻被处理的 Channel**。无论是在等待的 Goroutine 上还是缓冲区中，只要存在数据满足条件就会立刻处理，如果不能立刻找到活跃的 Channel 就会进入循环的下一阶段，按照需要将当前 Goroutine 加入到 Channel 的 `sendq` 或者 `recvq` 队列中
2. 将当前 Goroutine 加入 `Channel` 对应的收发队列上并等待其他 Goroutine 的唤醒；
3. 当前 Goroutine 被唤醒之后找到满足条件的 `Channel` 并进行处理；

`runtime.selectgo` 函数会根据不同情况通过 goto 语句跳转到函数内部的不同标签执行相应的逻辑，其中包括：

1. `bufrecv`：可以从缓冲区读取数据；
2. `bufsend`：可以向缓冲区写入数据；
3. `recv`：可以从休眠的发送方获取数据；
4. `send`：可以向休眠的接收方发送数据；
5. `rclose`：可以从关闭的 Channel 读取 EOF；
6. `sclose`：向关闭的 Channel 发送数据；
7. `retc`：结束调用并返回；

我们先来分析循环执行的第一个阶段，**查找已经准备就绪的 Channel**。循环会遍历所有的 `case` 并找到需要被唤起的 `runtime.sudog` 结构，在这个阶段，我们会根据 `case` 的四种类型分别处理：

1. 当 `case` 不包含 `Channel` 时；
   - 这种 `case` 会被跳过；
2. 当 `case` 会从 `Channel` 中接收数据时；
   - 如果当前 `Channel` 的 `sendq` 上有等待的 Goroutine，就会跳到 `recv` 标签并从缓冲区读取数据后将等待 Goroutine 中的数据放入到缓冲区中相同的位置；
   - 如果当前 `Channel` 的缓冲区不为空，就会跳到 `bufrecv` 标签处从缓冲区获取数据；
   - 如果当前 `Channel` 已经被关闭，就会跳到 `rclose` 做一些清除的收尾工作；
3. 当 `case` 会向 `Channel` 发送数据时；
   - 如果当前 `Channel` 已经被关，闭就会直接跳到 `sclose` 标签，触发 `panic` 尝试中止程序；
   - 如果当前 `Channel` 的 `recvq` 上有等待的 Goroutine，就会跳到 `send` 标签向 Channel 发送数据；
   - 如果当前 `Channel` 的缓冲区存在空闲位置，就会将待发送的数据存入缓冲区；
4. 当 `select` 语句中包含 `default` 时；
   - 表示前面的所有 `case` 都没有被执行，这里会解锁所有 `Channel` 并返回，意味着当前 `select` 结构中的收发都是非阻塞的；

**除了将当前 Goroutine 对应的 `runtime.sudog` 结构体加入队列之外，这些结构体都会被串成链表附着在 Goroutine 上。在入队之后会调用 `runtime.gopark` 挂起当前 Goroutine 等待调度器的唤醒。**

## 2. 同步原语&互斥锁

### 2.1 mutex

- Mutex 是互斥锁同一线程内不能重复加锁

- `sync/rwmutex.go` 中注释可以知道，读写锁当有⼀个协程在等待写锁时，其他协程是不能获得读锁的

- 加锁后复制变量，会将锁的状态也复制，所以 mu1 其实是已经加锁状态，再加锁会死锁。

```go
type MyMutex struct {
    count int
    sync.Mutex
}
func main() {
    var mu MyMutex
    mu.Lock()
    var mu2 = mu
    mu.count++
    mu.Unlock()
    mu2.Lock()
    mu2.count++
    mu2.Unlock()
    fmt.Println(mu.count, mu2.count)
}
```

### 2.2 mutex 加锁与解锁

#### 加锁

```go
func (m *Mutex) Lock() {
    if atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked) {
        return
    }
    m.lockSlow()
}

func (m *Mutex) lockSlow() {
    var waitStartTime int64
    starving := false
    awoke := false
    iter := 0
    old := m.state
    for {
        if old&(mutexLocked|mutexStarving) == mutexLocked && runtime_canSpin(iter) {
            if !awoke && old&mutexWoken == 0 && old>>mutexWaiterShift != 0 &&
                atomic.CompareAndSwapInt32(&m.state, old, old|mutexWoken) {
                awoke = true
            }
            runtime_doSpin()
            iter++
            old = m.state
            continue
        }
```

如果互斥锁的状态不是 0 时就会调用 `sync.Mutex.lockSlow` 尝试通过自旋（Spinnig）等方式等待锁的释放，该方法的主体是一个非常大 `for` 循环，这里将它分成几个部分介绍获取锁的过程：

1. 判断当前 Goroutine 能否进入自旋；
2. 通过自旋等待互斥锁的释放；
3. 计算互斥锁的最新状态；
4. 更新互斥锁的状态并获取锁；

**进入自旋的条件**：

1. 互斥锁只有在普通模式才能进入自旋；
2. `runtime.sync_runtime_canSpin` 需要返回 true：
   1. 运行在多 CPU 的机器上；
   2. 当前 Goroutine 为了获取该锁进入自旋的次数小于四次；
   3. 当前机器上至少存在一个正在运行的处理器 P 并且处理的运行队列为空；

如果没有通过 `CAS` 获得锁，会调用 `runtime.sync_runtime_SemacquireMutex` 通过信号量保证资源不会被两个 Goroutine 获取。`runtime.sync_runtime_SemacquireMutex` 会在方法中不断尝试获取锁并陷入休眠等待信号量的释放，一旦当前 Goroutine 可以获取信号量，它就会立刻返回，`sync.Mutex.Lock` 的剩余代码也会继续执行。

- 在正常模式下，这段代码会设置唤醒和饥饿标记、重置迭代次数并重新执行获取锁的循环；
- 在饥饿模式下，当前 Goroutine 会获得互斥锁，如果等待队列中只存在当前 Goroutine，互斥锁还会从饥饿模式中退出；

#### 解锁

互斥锁的解锁过程 `sync.Mutex.Unlock` 与加锁过程相比就很简单，该过程会先使用 `sync/atomic.AddInt32` 函数快速解锁，这时会发生下面的两种情况：

- 果该函数返回的新状态等于 0，当前 Goroutine 就成功解锁了互斥锁；
- 如果该函数返回的新状态不等于 0，这段代码会调用 `sync.Mutex.unlockSlow` 开始慢速解锁：

`sync.Mutex.unlockSlow` 解锁

- 在正常模式下，`sync.Mutex.unlockSlow` 会使用如下所示的处理过程：
  - 如果互斥锁不存在等待者或者互斥锁的 mutexLocked、mutexStarving、mutexWoken 状态不都为 0，那么当前方法可以直接返回，不需要唤醒其他等待者；
  - 如果互斥锁存在等待者，会通过 `sync.runtime_Semrelease` 唤醒等待者并移交锁的所有权；
- 在饥饿模式下，上述代码会直接调用 `sync.runtime_Semrelease` 将当前锁交给下一个正在尝试获取锁的等待者，等待者被唤醒后会得到锁，在这时互斥锁还不会退出饥饿状态；

**互斥锁的加锁过程比较复杂，它涉及自旋、信号量以及调度等概念：**

- 如果互斥锁处于初始化状态，会通过置位 `mutexLocked` 加锁；
- 如果互斥锁处于 `mutexLocked` 状态并且在普通模式下工作，会进入自旋，执行 30 次 `PAUSE` 指令消耗 CPU 时间等待锁的释放；
- 如果当前 Goroutine 等待锁的时间超过了 1ms，互斥锁就会切换到饥饿模式；
- 互斥锁在正常情况下会通过 `runtime.sync_runtime_SemacquireMutex` 将尝试获取锁的 Goroutine 切换至休眠状态，等待锁的持有者唤醒；
- 如果当前 Goroutine 是互斥锁上的最后一个等待的协程或者等待的时间小于 1ms，那么它会将互斥锁切换回正常模式；

互斥锁的解锁过程与之相比就比较简单，其代码行数不多、逻辑清晰，也比较容易理解：

- 当互斥锁已经被解锁时，调用 `sync.Mutex.Unlock` 会直接抛出异常；
- 当互斥锁处于饥饿模式时，将锁的所有权交给队列中的下一个等待者，等待者会负责设置 `mutexLocked` 标志位；
- 当互斥锁处于普通模式时，如果没有 Goroutine 等待锁的释放或者已经有被唤醒的 Goroutine 获得了锁，会直接返回；在其他情况下会通过 `sync.runtime_Semrelease` 唤醒对应的 Goroutine；

### 2.3 读写锁实现原理

**读写锁区别与互斥锁的主要区别就是读锁之间是共享的，多个 goroutine 可以同时加读锁，但是写锁与写锁、写锁与读锁之间则是互斥的**。

因为读锁是共享的，所以如果当前已经有读锁，那后续goroutine继续加读锁正常情况下是可以加锁成功，但是如果一直有读锁进行加锁，那尝试加写锁的goroutine则可能会长期获取不到锁，这就是因为读锁而导致的写锁饥饿问题

```go
type RWMutex struct {
    w           Mutex  // held if there are pending writers
    writerSem   uint32 // 用于writer等待读完成排队的信号量
    readerSem   uint32 // 用于reader等待写完成排队的信号量
    readerCount int32  // 读锁的计数器
    readerWait  int32  // 等待读锁释放的数量
}
```

1. **加读锁**

    ```go
    func (rw *RWMutex) RLock() {
        if race.Enabled {
            _ = rw.w.state
            race.Disable()
        }
        // 累加reader计数器，如果小于0则表明有writer正在等待
        if atomic.AddInt32(&rw.readerCount, 1) < 0 {
            // 当前有writer正在等待读锁，读锁就加入排队
            runtime_SemacquireMutex(&rw.readerSem, false)
        }
        if race.Enabled {
            race.Enable()
            race.Acquire(unsafe.Pointer(&rw.readerSem))
        }
    }
    ```

2. **释放读锁**

    ```go
    func (rw *RWMutex) RUnlock() {
        if race.Enabled {
            _ = rw.w.state
            race.ReleaseMerge(unsafe.Pointer(&rw.writerSem))
            race.Disable()
        }
        // redercount 如果小于0，则表明当前有 writer 正在等待
        // 不小于 0 直接释放锁
        if r := atomic.AddInt32(&rw.readerCount, -1); r < 0 {
            if r+1 == 0 || r+1 == -rwmutexMaxReaders {
                race.Enable()
                throw("sync: RUnlock of unlocked RWMutex")
            }
            // 将等待reader的计数减1，证明当前是已经有一个读的，
            // 如果值==0，则进行唤醒等待的，不等于零直接释放读锁
            if atomic.AddInt32(&rw.readerWait, -1) == 0 {
                // The last reader unblocks the writer.
                runtime_Semrelease(&rw.writerSem, false)
            }
        }
        if race.Enabled {
            race.Enable()
        }
    }
    ```

3. **加写锁**

    ```go
    func (rw *RWMutex) Lock() {
        if race.Enabled {
            _ = rw.w.state
            race.Disable()
        }
        // 首先获取mutex锁，同时多个goroutine只有一个可以进入到下面的逻辑
        rw.w.Lock()
        // 对readerCounter进行进行抢占，通过递减rwmutexMaxReaders允许最大读的数量
        // 来实现写锁对读锁的抢占
        r := atomic.AddInt32(&rw.readerCount, -rwmutexMaxReaders) + rwmutexMaxReaders
        // 记录需要等待多少个reader完成,如果发现不为0，则表明当前有reader正在读取，当前goroutine
        // 需要进行排队等待
        if r != 0 && atomic.AddInt32(&rw.readerWait, r) != 0 {
            runtime_SemacquireMutex(&rw.writerSem, false)
        }
        if race.Enabled {
            race.Enable()
            race.Acquire(unsafe.Pointer(&rw.readerSem))
            race.Acquire(unsafe.Pointer(&rw.writerSem))
        }
    }
    ```

    - 调用结构体持有的 `sync.Mutex` 结构体的 `sync.Mutex.Lock` 阻塞后续的写操作；
    - 因为互斥锁已经被获取，其他 Goroutine 在获取写锁时会进入自旋或者休眠；
    - 调用 `sync/atomic.AddInt32` 函数阻塞后续的读操作：
    - 如果仍然有其他 Goroutine 持有互斥锁的读锁，**该 Goroutine 会调用 `runtime.sync_runtime_SemacquireMutex` 进入休眠状态等待所有读锁所有者执行结束后释放 `writerSem` 信号量将当前协程唤醒；**

4. **释放写锁**

    ```go
    func (rw *RWMutex) Unlock() {
        if race.Enabled {
            _ = rw.w.state
            race.Release(unsafe.Pointer(&rw.readerSem))
            race.Disable()
        }

        // 将reader计数器复位，上面减去了一个rwmutexMaxReaders现在再重新加回去即可复位
        r := atomic.AddInt32(&rw.readerCount, rwmutexMaxReaders)
        if r >= rwmutexMaxReaders {
            race.Enable()
            throw("sync: Unlock of unlocked RWMutex")
        }
        // 唤醒所有的读锁
        for i := 0; i < int(r); i++ {
            runtime_Semrelease(&rw.readerSem, false)
        }
        // 释放mutex
        rw.w.Unlock()
        if race.Enabled {
            race.Enable()
        }
    }
    ```

**RWMutex 小结**。

- 调用 `sync.RWMutex.Lock` 尝试获取写锁时；
  - 每次 `sync.RWMutex.RUnlock` 都会将 `readerCount` 其减一，当它归零时该 Goroutine 会获得写锁；
  - 将 `readerCount` 减少 `rwmutexMaxReaders` 个数以阻塞后续的读操作；
- 调用 `sync.RWMutex.Unlock` 释放写锁时，会先通知所有的读操作，然后才会释放持有的互斥锁；

### 2.4. waitgroup

- `WaitGroup` 在调⽤ `Wait` 之后是不能再调⽤ `Add` ⽅法的。

通过对 `sync.WaitGroup` 的分析和研究，我们能够得出以下结论：

- `sync.WaitGroup` 必须在 `sync.WaitGroup.Wait` 方法返回之后才能被重新使用；
- `sync.WaitGroup.Done` 只是对 `sync.WaitGroup.Add` 方法的简单封装，我们可以向 `sync.WaitGroup.Add` 方法传入任意负数（需要保证计数器非负）快速将计数器归零以唤醒等待的 Goroutine；
- **可以同时有多个 Goroutine 等待当前 `sync.WaitGroup` 计数器的归零，这些 Goroutine 会被同时唤醒；**

### 2.5 once

```go
type Once struct {
    done uint32
    m    Mutex
}
```

作为用于保证函数执行次数的 `sync.Once` 结构体，它使用互斥锁和 sync/atomic 包提供的方法实现了某个函数在程序运行期间只能执行一次的语义。在使用该结构体时，我们也需要注意以下的问题：

- `sync.Once.Do` 方法中传入的函数只会被执行一次，哪怕函数中发生了 `panic`；
- 两次调用 `sync.Once.Do` 方法传入不同的函数只会执行第一次调传入的函数；


### 2.6 cond

`sync.Cond` 不是一个常用的同步机制，但是在条件长时间无法满足时，与使用 `for {}` 进行忙碌等待相比，`sync.Cond` 能够让出处理器的使用权，提高 CPU 的利用率。使用时我们也需要注意以下问题：

- `sync.Cond.Wait` 在调用之前一定要使用获取互斥锁，否则会触发程序崩溃；
- `sync.Cond.Signal` 唤醒的 Goroutine 都是队列最前面、等待最久的 Goroutine；
- `sync.Cond.Broadcast` 会按照一定顺序广播通知等待的全部 Goroutine；

## 3. unsafe

unsafe 包提供了 2 点重要的能力：

1. 任何类型的指针和 `unsafe.Pointer` 可以相互转换。
2. `uintptr` 类型和 `unsafe.Pointer` 可以相互转换。

**`uintptr` 并没有指针的语义，意思就是 `uintptr` 所指向的对象会被 gc 无情地回收。而 `unsafe.Pointer` 有指针语义，可以保护它所指向的对象在“有用”的时候不会被垃圾回收。**

### 3.1 unsfae 获取 slice &map

1. 可以通过 unsafe.Pointer 和 uintptr 进行转换，得到 slice 的字段值。

    ```go
    func main() {
        s := make([]int, 9, 20)
        var Len = *(*int)(unsafe.Pointer(uintptr(unsafe.Pointer(&s)) + uintptr(8)))
        fmt.Println(Len, len(s)) // 9 9

        var Cap = *(*int)(unsafe.Pointer(uintptr(unsafe.Pointer(&s)) + uintptr(16)))
        fmt.Println(Cap, cap(s)) // 20 20
    }

    // Len: &s => pointer => uintptr => pointer => *int => int
    // Cap: &s => pointer => uintptr => pointer => *int => int
    ```

2. 能通过 unsafe.Pointer 和 uintptr 进行转换，得到 hamp 字段的值，只不过，现在 count 变成二级指针了：
 
    ```go
    func main() {
        mp := make(map[string]int)
        mp["qcrao"] = 100
        mp["stefno"] = 18

        count := **(**int)(unsafe.Pointer(&mp))
        fmt.Println(count, len(mp)) // 2 2
    }

    // &mp => pointer => **int => int
    ```

### 3.2 修改私有成员

对于一个结构体，通过 `offset` 函数可以获取结构体成员的偏移量，进而获取成员的地址，读写该地址的内存，就可以达到改变成员值的目的。

这里有一个内存分配相关的事实：结构体会被分配一块连续的内存，结构体的地址也代表了第一个成员的地址。

```go
type Programmer struct {
    name string
    language string
}

func main() {
    p := Programmer{"stefno", "go"}
    fmt.Println(p)
    
    name := (*string)(unsafe.Pointer(&p))
    *name = "qcrao"

    lang := (*string)(unsafe.Pointer(uintptr(unsafe.Pointer(&p)) + unsafe.Offsetof(p.language)))
    *lang = "Golang"

    fmt.Println(p) // {qcrao Golang}
}
```

`name` 是结构体的第一个成员，因此可以直接将 `&p` 解析成 `*string`。这一点，在前面获取 `map` 的 `count` 成员时，用的是同样的原理。

**对于结构体的私有成员，现在有办法可以通过 `unsafe.Pointer` 改变它的值了。**

多加一个 age 字段，并且放在其他包，这样在 main 函数中，它的三个字段都是私有成员变量，不能直接修改。**通过 `unsafe.Sizeof()` 函数可以获取成员大小，进而计算出成员的地址，直接修改内存**。

```go
type Programmer struct {
    name string
    age int
    language string
}

func main() {
    p := Programmer{"stefno", 18, "go"}
    fmt.Println(p)

    lang := (*string)(unsafe.Pointer(uintptr(unsafe.Pointer(&p)) + unsafe.Sizeof(int(0)) + unsafe.Sizeof(string(""))))
    *lang = "Golang"

    fmt.Println(p) // {stefno 18 Golang}
}
```

### 3.3 字符串和 byte 转换

实现字符串和 `bytes` 切片之间的转换，要求是 `zero-copy`。

上面是反射包下的结构体，路径：src/reflect/value.go。只需要共享底层 Data 和 Len 就可以实现 zero-copy。

```go
func string2bytes(s string) []byte {
    return *(*[]byte)(unsafe.Pointer(&s))
}
func bytes2string(b []byte) string{
    return *(*string)(unsafe.Pointer(&b))
}
```

原理上是利用指针的强转，代码比较简单，不作详细解释。

## 2. channel

### 介绍 channel

Go 语言中，**不要通过共享内存来通信，而要通过通信来实现内存共享**。Go 的 `CSP`(Communicating Sequential Process)并发模型，中文叫做**通信顺序进程**，是通过 goroutine 和 channel 来实现的。

**channel 收发遵循先进先出 FIFO，分为有缓存和无缓存**，channel 中大致有 `buffer`(当缓冲区大小部位 0 时，是个 `ring buffer`)、`sendx` 和 `recvx` 收发的位置(`ring buffer` 记录实现)、`sendq`、`recvq` 当前 channel 因为缓冲区不足而阻塞的队列、使用双向链表存储、还有一个 `mutex` 锁控制并发、其他原属等。

### 2.1 向通道发送数据

```go
func chansend(c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool {
    lock(&c.lock)

    if c.closed != 0 {
        unlock(&c.lock)
        panic(plainError("send on closed channel"))
    }
```

**在发送数据的逻辑执行之前会先为当前 Channel 加锁，防止多个线程并发修改数据。如果 Channel 已经关闭，那么向该 Channel 发送数据时会报 “send on closed channel” 错误并中止程序**。

执行过程分成以下的三个部分：

- **当存在等待的接收者时**，通过 `runtime.send` 直接将数据发送给阻塞的接收者；
- **当缓冲区存在空余空间时**，将发送的数据写入 Channel 的缓冲区；
- **当不存在缓冲区或者缓冲区已满时**，等待其他 Goroutine 从 Channel 接收数据；

#### 2.1.1 发送数据

- 如果当前 Channel 的 `recvq` 上存在已经被阻塞的 Goroutine，那么会直接将数据发送给当前 Goroutine 并将其设置成下一个运行的 Goroutine，**也就是将接收方的 Goroutine 放到了处理器的 `runnext` 中，程序没有立刻执行该 Goroutine**
    > 向一个非缓冲型的 channel 发送数据、从一个无元素的（非缓冲型或缓冲型但空）的 channel 接收数据，都会导致一个 goroutine 直接操作另一个 goroutine 的栈, 由于 GC 假设对栈的写操作只能发生在 goroutine 正在运行中并且由当前 goroutine 来写, 所以这里实际上违反了这个假设。可能会造成一些问题，**所以需要用到写屏障来规避**
- 如果 Channel 存在缓冲区并且其中还有空闲的容量，我们会直接将数据存储到缓冲区 `sendx` 所在的位置上；
- 如果不满足上面的两种情况，会创建一个 `runtime.sudog` 结构并将其加入 Channel 的 `sendq` 队列中，当前 Goroutine 也会陷入阻塞等待其他的协程从 Channel 接收数据；

### 2.2 接收数据

```go
func chanrecv(c *hchan, ep unsafe.Pointer, block bool) (selected, received bool) {
    if c == nil {
        if !block {
            return
        }
        gopark(nil, nil, waitReasonChanReceiveNilChan, traceEvGoStop, 2)
        throw("unreachable")
    }

    lock(&c.lock)

    if c.closed != 0 && c.qcount == 0 {
        unlock(&c.lock)
        if ep != nil {
            typedmemclr(c.elemtype, ep)
        }
        return true, false
    }
```

**当我们从一个空 Channel 接收数据时会直接调用 `runtime.gopark` 让出处理器的使用权。**

使用 `runtime.chanrecv` 从 Channel 接收数据时还包含以下三种不同情况：

1. **当存在等待的发送者时**，通过 `runtime.recv` 从阻塞的发送者或者缓冲区中获取数据；
2. **当缓冲区存在数据时**，从 Channel 的缓冲区中接收数据；
3. **当缓冲区中不存在数据时**，等待其他 Goroutine 向 Channel 发送数据；

直接接收数据时根据缓冲区的大小分别处理不同的情况：

- **如果 Channel 不存在缓冲区**；
   1. 调用 `runtime.recvDirect` 将 Channel 发送队列中 Goroutine 存储的 `elem` 数据拷贝到目标内存地址中；
- **如果 Channel 存在缓冲区**；
   1. 将队列中的数据拷贝到接收方的内存地址；
   2. 将发送队列头的数据拷贝到缓冲区中，释放一个阻塞的发送方；

**无论发生哪种情况，运行时都会调用 `runtime.goready` 将当前处理器的 `runnext` 设置成发送数据的 Goroutine，在调度器下一次调度时将阻塞的发送方唤醒。**

#### 2.2.1 从通道接收数据

1. 如果 Channel 为空，那么会直接调用 `runtime.gopark` 挂起当前 Goroutine；
2. 如果 Channel 已经关闭并且缓冲区没有任何数据，`runtime.chanrecv` 会直接返回；
3. 如果 Channel 的 `sendq` 队列中存在挂起的 Goroutine，会将 `recvx` 索引所在的数据拷贝到接收变量所在的内存空间上并将 `sendq` 队列中 Goroutine 的数据拷贝到缓冲区；
4. 如果 Channel 的缓冲区中包含数据，那么直接读取 `recvx` 索引对应的数据；
5. 在默认情况下会挂起当前的 Goroutine，将 `runtime.sudog` 结构加入 `recvq` 队列并陷入休眠等待调度器的唤醒；

### 2.3 触发调度时机

1. **发送数据时**
   1. 发送数据时发现 Channel 上存在等待接收数据的 Goroutine，立刻设置处理器的 `runnext` 属性，但是并不会立刻触发调度；
   2. 发送数据时并没有找到接收方并且缓冲区已经满了，这时会将自己加入 Channel 的 `sendq` 队列并调用 `runtime.goparkunlock` 触发 Goroutine 的调度让出处理器的使用权；
2. **从 Channel 接收数据时，会触发 Goroutine 调度的两个时机**：

   1. 当 Channel 为空时；
   2. 当缓冲区中不存在数据并且也不存在数据的发送者时；

### 2.4 最佳实践

- **读已经关闭的 `chan` 能⼀直读到东⻄，但是读到的内容根据通道内关闭前是否有元素⽽不同。**
   1. 如果 `chan` 关闭前，`buffer` 内有元素还未读, 会正确读到 `chan` 内的值，且返回的第⼆个 `bool` 值（是否读成功）为 `true`。
   2. 如果 `chan` 关闭前，`buffer` 内有元素已经被读完，`chan` 内⽆值，接下来所有接收的值都会⾮阻塞直接成功，返回 `channel` 元素的零值，但是第⼆个 `bool` 值⼀直为 `false`。
- **触发 `panic` 的情况**
   1. **向已经关闭的 `chan` 发送数据会 `panic`**
   2. 关闭一个 `nil` 的 channel；
   3. 重复关闭一个 channel。

- 向 `nil` 的通道发送或接收数据会调用 `gopark` 挂起，并陷入永久阻塞
- channel 泄漏
    **泄漏的原因是 goroutine 操作 channel 后，处于发送或接收阻塞状态，而 channel 处于满或空的状态，一直得不到改变**。同时，垃圾回收器也不会回收此类资源，进而导致 gouroutine 会一直处于等待队列中，不见天日。

## 3. defer

`defer` 关键字的实现跟 `go` 关键字很类似，不同的是它调⽤的是 `runtime.deferproc` ⽽不是 `runtime.newproc`。 在 `defer` 出现的地⽅，插⼊了指令 `call runtime.deferproc`，然后在函数返回之前的地⽅，插⼊指令 `call runtime.deferreturn`。 goroutine 的控制结构中，有⼀张表记录 `defer`，调⽤ `runtime.deferproc` 时会将需要 `defer` 的表达式记录在表中，⽽在调⽤ `runtime.deferreturn` 的时候，则会依次从 `defer` 表中出栈并执⾏。
因此，题⽬最后输出顺序应该是 defer 定义顺序的倒序。 panic 错误并不能终⽌ defer 的执⾏。

## 4. 内存分配

### 4.x 内存逃逸

**逃逸分析最基本的原则**: 编译器会分析代码的特征和代码生命周期，Go 中的变量只有在编译器可以证明在函数返回后不会再被引用的，才分配到栈上，其他情况下都是分配到堆上。

*引起变量逃逸到堆上的典型情况*：

- **在⽅法内把局部变量指针返回**, 局部变量原本应该在栈中分配，在栈中回收。但是由于返回时被外部引⽤，因此其⽣命周期⼤于栈，则溢出。
- **发送指针或带有指针的值到 `channel` 中**。在编译时，是没有办法知道哪个 `goroutine` 会在 `channel` 上接收数据。所以编译器没法知道变量什么时候才会被释放。
- **在⼀个切⽚上存储指针或带指针的值**。⼀个典型的例⼦就是 `[]*string`。这会导致切⽚的内容逃逸。尽管其后⾯的数组可能是在栈上分配的，但其引⽤的值⼀定是在堆上。
- **`slice` 的背后数组被重新分配了，因为 `append` 时可能会超出其容量(`cap`)**。`slice` 初始化的地⽅在编译时是可以知道的，它最开始会在栈上分配。如果切⽚背后的存储要基于运⾏时的数据进⾏扩充，就会在堆上分配。
- **在 `interface` 类型上调⽤⽅法**。在 `interface` 类型上调⽤⽅法都是动态调度的⽅法的真正实现只能在运⾏时知道。想像⼀个 `io.Reader` 类型的变量 `r` , 调⽤ `r.Read(b)` 会使得 `r` 的值和切⽚ `b` 的背后存储都逃逸掉，所以会在堆上分配。

- **闭包的捕获变量也会分配到堆上，还有就是大对象 大于 32KB**

### new 一个对象最后在堆上还是栈上就根据上面的逃逸分析进行回答

## 5. GC 垃圾回收

### 5.1 GC 策略

1. 内存达到上限触发 GC
就像上面的 `GOGC` 配置那样，当程序达到驻留内存的相应倍数时候，触发 GC， 默认值就是两倍于当前内存。
--
    > GOGC 参数
    `GOGC` 默认值是100，举个例子：你程序的上一次 GC 完，驻留内存是100MB，由于你 GOGC 设置的是100，所以下次你的内存达到 200MB 的时候就会触发一次 GC，如果你 GOGC 设置的是200，那么下次你的内存达到300MB的时候就会触发 GC。

2. 时间达到触发 GC 的阈值
垃圾收集器关注的第二个指标是两个垃圾收集器之间的延迟。如果它没有被触发超过两分钟，一个循环将被强制。
3. **主动触发**，上面两种是**被动触发**，通过调用 `runtime.GC` 来触发 GC，此调用阻塞式地等待当前 GC 运行完毕。

### 5.2 GC 如何调优

- 通过 `go tool pprof` 和 `go tool trace` 等工具
- 控制内存分配的速度，限制 goroutine 的数量，从而提高赋值器对 CPU 的利用率。
- 减少并复用内存，例如使用 `sync.Pool` 来复用需要频繁创建临时对象，例如提前分配足够的内存来降低多余的拷贝。
- 需要时，增大 `GOGC` 的值，降低 GC 的运行频率。

## 3. GMP&调度器

### 3.1 goroutine 与线程有什么区别？

- **内存占用**
    创建一个 goroutine 的栈内存消耗为 2 KB，实际运行过程中，如果栈空间不够用，会自动进行扩容。创建一个 thread 则需要消耗 1 MB 栈内存，而且还需要一个被称为 “a guard page” 的区域用于和其他 thread 的栈空间进行隔离。

    对于一个用 Go 构建的 HTTP Server 而言，对到来的每个请求，创建一个 goroutine 用来处理是非常轻松的一件事。而如果用一个使用线程作为并发原语的语言构建的服务，例如 Java 来说，每个请求对应一个线程则太浪费资源了，很快就会出 OOM 错误（OutOfMemoryError）。

- **创建和销毀**
    Thread 创建和销毀都会有巨大的消耗，因为要和操作系统打交道，是内核级的，通常解决的办法就是线程池。而 goroutine 因为是由 `Go runtime` 负责管理的，创建和销毁的消耗非常小，是用户态的。

- **切换**
    当 threads 切换时，需要保存各种寄存器，以便将来恢复

    而 goroutines 切换只需保存三个寄存器：Program Counter, Stack Pointer and BP。

    一般而言，线程切换会消耗 1000-1500 纳秒，一个纳秒平均可以执行 12-18 条指令。所以由于线程切换，执行指令的条数会减少 12000-18000。

    Goroutine 的切换约为 200 ns，相当于 2400-3600 条指令。

    因此，goroutines 切换成本比 threads 要小得多。

### M:N 模型

Runtime 会在程序启动的时候，创建 `M` 个线程（CPU 执行调度的单位），之后创建的 `N` 个 goroutine 都会依附在这 `M` 个线程上执行。

在同一时刻，一个线程上只能跑一个 goroutine。当 goroutine 发生阻塞（例如上篇文章提到的向一个 channel 发送数据，被阻塞）时，runtime 会把当前 goroutine 调度走，让其他 goroutine 来执行。

### 3.2 schedule 调度



*`Runtime` 运行时维护所有的 `goroutines`，并通过 `scheduler` 来进行调度。`Goroutines` 和 `threads` 是独立的，但是 `goroutines` 要依赖 `threads` 才能执行。*

#### 调度的核心思想

- `reuse threads`
- 限制同时运行（不包含阻塞）的线程数为 `N`，`N` 等于 CPU 的核心数目
- 线程私有的 `runqueues`，并且可以从其他线程 stealing goroutine 来运行，线程阻塞后，可以将 `runqueues` 传递给其他线程。

#### GMP

- `g` 代表一个 goroutine，它包含：表示 goroutine 栈的一些字段，指示当前 `goroutine` 的状态，指示当前运行到的指令地址，也就是 PC 值。
- `m` 表示操作系统内核线程，包含正在运行的 `goroutine` 等字段。
- `p` 代表一个虚拟的 `Processor`，它维护一个处于 `Runnable` 状态的 `g` 队列，`m` 需要获得 `p` 才能运行 `g`。

Go 程序启动后，会给每个逻辑核心分配一个 `P`（Logical Processor）；同时，会给每个 `P` 分配一个 `M`（Machine，表示内核线程），这些内核线程仍然由 `OS scheduler` 来调度。

##### goroutine 状态

- **等待中**：Goroutine 正在等待某些条件满足，例如：系统调用结束等，包括 `_Gwaiting`、`_Gsyscall` 和 `_Gpreempted` 几个状态；
- **可运行**：Goroutine 已经准备就绪，可以在线程运行，如果当前程序中有非常多的 Goroutine，每个 Goroutine 就可能会等待更多的时间，即 `_Grunnable；`
- **运行中**：Goroutine 正在某个线程上运行，即 `_Grunning`；

### 调度器启动

**系统加载可执行文件大概都会经过这几个阶段：**

1. 从磁盘上读取可执行文件，加载到内存
2. 创建进程和主线程
3. 为主线程分配栈空间
4. 把由用户在命令行输入的参数拷贝到主线程的栈
5. 把主线程放入操作系统的运行队列等待被调度

大致过程：

1. **运行时类型检查**，主要是校验编译器的翻译工作是否正确，是否有 “坑”。
2. **系统参数传递**，主要是将系统参数转换传递给程序使用。
3. `runtime.osinit`：**系统基本参数设置**，主要是获取 CPU 核心数和内存物理页大小。
4. **`runtime.schedinit`：进行各种运行时组件的初始化**，包含调度器、内存分配器、堆、栈、GC 等一大堆初始化工作。会进行 p 的初始化，并将 m0 和某一个 p 进行绑定。
   - 先通过汇编初始化 g0
   - 然后主线程 tls 绑定 m0
   - 初始化 m0 并挂到 allm 中
   - 接下来是 procresize 函数运行
 
    通过 `runtime.schedinit` 初始化调度器，在调度器初始函数执行的过程中会将 `maxmcount` 设置成 10000，这也就是一个 Go 语言程序能够创建的最大线程数，虽然最多可以创建 10000 个线程，但是可以同时运行的线程还是由 `GOMAXPROCS` 变量控制。

    **调用 `runtime.procresize` 的执行过程如下**

    1. 如果全局变量 `allp` 切片中的处理器数量少于期望数量，会对切片进行扩容；
    2. 使用 `new` 创建新的处理器结构体并调用 `runtime.p.init` 初始化刚刚扩容的处理器；
    3. 通过指针将线程 `m0` 和处理器 `allp[0]` 绑定到一起；**也就是绑定 `m0` 和 `p0`**
    4. 调用 `runtime.p.destroy` 释放不再使用的处理器结构；
    5. 通过截断改变全局变量 `allp` 的长度保证与期望处理器数量相等；
    6. 将除 `allp[0]` 之外的处理器 `P` 全部设置成 `_Pidle` 并加入到全局的空闲队列中；
5. `runtime.main`：主要工作是运行 main goroutine，虽然在 `runtime·rt0_go` 中指向的是$runtime·mainPC，但实质指向的是 runtime.main。
6. runtime.newproc：创建一个新的 goroutine，且绑定 runtime.main 方法（也就是应用程序中的入口 main 方法）。并将其放入 m0 绑定的p的本地队列中去，以便后续调度。
7. runtime.mstart：启动 m，调度器开始进行循环调度。

**注意要点：**

- 因为 m0 是全局变量，而 m0 又要绑定到工作线程才能执行。runtime 会启动多个工作线程，每个线程都会绑定一个 `m0`。而且，代码里还得保持一致，都是用 m0 来表示。这就要用到线程本地存储的知识了，也就是常说的 TLS（Thread Local Storage）。简单来说，TLS 就是线程本地的私有的全局变量。

### 3.x 创建 goroutine

使用 Go 语言的 `go` 关键字，编译器会通过 `cmd/compile/internal/gc.state.stmt` 和 `cmd/compile/internal/gc.state.call` 两个方法**将该关键字转换成 `runtime.newproc` 函数调用**, 入参是参数大小和表示函数的指针 `funcval`，它会获取 Goroutine 以及调用方的程序计数器，然后调用 `runtime.newproc1` 函数获取新的 Goroutine 结构体、将其加入处理器的运行队列并在满足条件时调用 `runtime.wakep` 唤醒新的处理执行 Goroutine.

`runtime.newproc1` 会根据传入参数初始化一个 `g` 结构体，我们可以将该函数分成以下几个部分介绍它的实现：

1. 获取或者创建新的 Goroutine 结构体；
2. 将传入的参数移到 Goroutine 的栈上；
3. 更新 Goroutine 调度相关的属性；

#### 初始化结构体

`runtime.gfget` 通过两种不同的方式获取新的 `runtime.g`：

1. 从 Goroutine 所在处理器的 `gFree` 列表或者调度器的 `sched.gFree` 列表中获取 `runtime.g`；
   1. 当处理器的 Goroutine 列表为空时，会将调度器持有的空闲 Goroutine 转移到当前处理器上，直到 `gFree` 列表中的 Goroutine 数量达到 32；
   2. 当处理器的 Goroutine 数量充足时，会从列表头部返回一个新的 Goroutine；
2. 调用 `runtime.malg` 生成一个新的 `runtime.g` 并将结构体追加到全局的 Goroutine 列表 `allgs` 中。

#### 运行队列

**`runtime.runqput` 会将 Goroutine 放到运行队列上，这既可能是全局的运行队列，也可能是处理器本地的运行队列。**

1. 当 `next` 为 `true` 时，将 Goroutine 设置到处理器的 `runnext` 作为下一个处理器执行的任务；
2. 当 `next` 为 `false` 并且本地运行队列还有剩余空间时，将 Goroutine 加入处理器持有的本地运行队列；
3. **当处理器的本地运行队列已经没有剩余空间时就会把本地队列中的一半分 Goroutine 和待加入的 Goroutine 通过 `runtime.runqputslow` 添加到调度器持有的全局运行队列上**；

**处理器本地的运行队列是一个使用数组构成的环形链表，它最多可以存储 256 个待执行任务。**

### 调度循环

调度器启动之后，Go 语言运行时会调用 `runtime.mstart` 以及 `runtime.mstart1`，前者会初始化 `g0` 的 `stackguard0` 和 `stackguard1` 字段，后者会初始化线程并调用 `runtime.schedule` 进入调度循环。

**`runtime.schedule` 函数会从下面几个地方查找待执行的 Goroutine：**

1. 为了保证公平，当全局运行队列中有待执行的 Goroutine 时，通过 `schedtick` 保证有一定几率会从全局的运行队列中查找对应的 Goroutine；
2. 从处理器本地的运行队列中查找待执行的 Goroutine；
3. 如果前两种方法都没有找到 Goroutine，会通过 `runtime.findrunnable` 进行阻塞地查找 Goroutine；

`runtime.findrunnable` 的实现非常复杂，通过以下的过程获取可运行的 Goroutine：

1. 从本地运行队列、全局运行队列中查找；
2. 从网络轮询器中查找是否有 Goroutine 等待运行；
3. 通过 `runtime.runqsteal` 尝试从其他随机的处理器中窃取待运行的 Goroutine，该函数还可能窃取处理器的计时器；

**main goroutine 和普通 goroutine 的退出过程：**

- 对于 main goroutine，在执行完用户定义的 main 函数的所有代码后，直接调用 `exit(0)` 退出整个进程，非常霸道。

- 对于普通 goroutine 需要经历一系列的过程。先是跳转到提前设置好的 `goexit` 函数的第二条指令，然后调用 `runtime.goexit1`，接着调用 mcall(goexit0)，而 mcall 函数会切换到 `g0` 栈，运行 `goexit0` 函数，清理 goroutine 的一些字段，并将其添加到 goroutine 缓存池里，然后进入 `schedule` 调度循环。到这里，普通 goroutine 才算完成使命。

**循环运转**
栈空间在调用函数时会自动“增大”，而函数返回时，会自动“减小”，这里的增大和减小是指栈顶指针 SP 的变化。上述这些函数都没有返回，说明调用者不需要用到被调用者的返回值，有点像“尾递归”。

因为 g0 一直没有动过，所有它之前保存的 sp 还能继续使用。每一次调度循环都会覆盖上一次调度循环的栈数据，完美！

### 3.x 触发调度

![image](https://mail.wangkekai.cn/1641112475515.jpg)

除了上图中可能触发调度的时间点，运行时还会在线程启动 `runtime.mstart` 和 Goroutine 执行结束 `runtime.goexit0` 触发调度。我们在这里会重点介绍运行时触发调度的几个路径：

1. **主动挂起** — `runtime.gopark` -> `runtime.park_m`
2. **系统调用** — `runtime.exitsyscall` -> `runtime.exitsyscall0`
3. **协作式调度** — `runtime.Gosched` -> `runtime.gosched_m` -> `runtime.goschedImpl`
4. **系统监控** — `runtime.sysmon` -> `runtime.retake` -> `runtime.preemptone`

#### 主动挂起

1. `runtime.gopark` 是触发调度最常见的方法，该函数会将当前 Goroutine 暂停，被暂停的任务不会放回运行队列
2. 然后通过 r`untime.mcal`l 切换到 `g0` 的栈上调用 `runtime.park_m`：
    `runtime.park_m` 会将当前 Goroutine 的状态从 `_Grunning` 切换至 `_Gwaiting`，调用 `runtime.dropg` 移除线程和 Goroutine 之间的关联，在这之后就可以调用 `runtime.schedule` 触发新一轮的调度了。
3. 当 Goroutine 等待的特定条件满足后，运行时会调用 `runtime.goready` 将因为调用 `runtime.gopark` 而陷入休眠的 Goroutine 唤醒。
4. `runtime.ready` 会将准备就绪的 Goroutine 的状态切换至 `_Grunnable` 并将其加入处理器的运行队列中，等待调度器的调度。

#### 系统调用

第一步在通过汇编指令 `INVOKE_SYSCALL` 执行系统调用前后，上述函数会调用运行时的 `runtime.entersyscall` 和 `runtime.exitsyscall`，**正是这一层包装能够在陷入系统调用前触发运行时的准备和清理工作**。

1. `runtime.entersyscall` 会在获取当前程序计数器和栈位置之后调用 `runtime.reentersyscall`，它会完成 Goroutine 进入系统调用前的准备工作
   1. 禁止线程上发生的抢占，防止出现内存不一致的问题；
   2. 保证当前函数不会触发栈分裂或者增长；
   3. 保存当前的程序计数器 PC 和栈指针 SP 中的内容；
   4. 将 Goroutine 的状态更新至 `_Gsyscall`
   5. 将 Goroutine 的处理器和线程暂时分离并更新处理器的状态到 `_Psyscall`；
   6. 释放当前线程上的锁；

   不过出于性能的考虑，如果这次系统调用不需要运行时参与，就会使用 `syscall.RawSyscall` 简化这一过程，不再调用运行时函数
2. 当系统调用结束后，会调用退出系统调用的函数 `runtime.exitsyscall` 为当前 Goroutine 重新分配资源
   1. 调用 `runtime.exitsyscallfast`；
       - 如果 Goroutine 的原处理器处于 `_Psyscall` 状态，会直接调用 `wirep` 将 Goroutine 与处理器进行关联；
       - 如果调度器中存在闲置的处理器，会调用 `runtime.acquirep` 使用闲置的处理器处理当前 Goroutine
   2. 切换至调度器的 Goroutine 并调用 `runtime.exitsyscall0`；
        另一个相对较慢的路径 `runtime.exitsyscall0` 会将当前 Goroutine 切换至 `_Grunnable` 状态，并移除线程 `M` 和当前 Goroutine 的关联：

      - 当我们通过 `runtime.pidleget` 获取到闲置的处理器时就会在该处理器上执行 Goroutine；
      - 在其它情况下，我们会将当前 Goroutine 放到全局的运行队列中，等待调度器的调度

#### sysmon

- sysmon 执行一个无限循环，一开始每次循环休眠 20us，之后（1 ms 后）每次休眠时间倍增，最终每一轮都会休眠 10ms。

- sysmon 中会进行 netpool（获取 fd 事件）、retake（抢占）、forcegc（按时间强制执行 gc），scavenge heap（释放自由列表中多余的项减少内存占用）等处理。

#### 协作式调度

`runtime.Gosched` 函数会主动让出处理器，允许其他 Goroutine 运行。该函数无法挂起 Goroutine，调度器可能会将当前 Goroutine 调度到其他线程上：

经过连续几次跳转，我们最终在 `g0` 的栈上调用 `runtime.goschedImpl`，运行时会更新 Goroutine 的状态到 `_Grunnable`，让出当前的处理器并将 Goroutine 重新放回全局队列，在最后，该函数会调用 `runtime.schedule` 触发调度。

### 线程管理

Go 语言的运行时会通过调度器改变线程的所有权，它也提供了 `runtime.LockOSThread` 和 `runtime.UnlockOSThread` 让我们有能力绑定 Goroutine 和线程完成一些比较特殊的操作。

- Go 语言的运行时会通过 `runtime.startm` 启动线程来执行处理器 `P`，如果我们在该函数中没能从闲置列表中获取到线程 `M` 就会调用 `runtime.newm` 创建新的线程;
- 创建新的线程需要使用如下所示的 `runtime.newosproc`，该函数在 Linux 平台上会通过系统调用 clone 创建新的操作系统线程，它也是创建线程链路上距离操作系统最近的 Go 语言函数
- 使用系统调用 `clone` 创建的线程会在线程主动调用 `exit`、或者传入的函数 `runtime.mstart` 返回会主动退出，`runtime.mstart` 会执行调用 `runtime.newm` 时传入的匿名函数 `fn`，到这里也就完成了从线程创建到销毁的整个闭环。

### 调度概览

- 为了保证调度的公平性，**每个工作线程每进行 61 次调度就需要优先从全局运行队列中获取 goroutine 出来运行**
    因为如果只调度本地运行队列中的 goroutine，则全局运行队列中的 goroutine 有可能得不到运行
    **所有工作线程都能访问全局队列，所以需要加锁获取 goroutine**
- _如果从全局队列没有获取到 goroutine_，从与 `m` 关联的 `p` 的本地运行队列中获取 goroutine
- **如果从本地运行队列和全局运行队列都没有找到需要运行的goroutine， 则调用 `findrunnable` 函数从其它工作线程的运行队列中偷取**，如果偷取不到，则当前工作线程进入睡眠，直到获取到需要运行的 goroutine 之后 findrunnable 函数才会返回。

#### 3.2 从全局队列获取 goroutine

- ⾸先 `globrunqget` 函数会根据全局运⾏队列中 goroutine 的数量，函数参数 `max` 以及 `_p_`的本地队列的容量计算出到底应该拿多少个 goroutine
- 然后把第⼀个 `g` 结构体对象通过返回值的⽅式返回给调⽤函数，其它的则通过 `runqput` 函数放⼊当前⼯作线程的本地运⾏队列。
*计算应该从全局运⾏队列中拿⾛多少个 goroutine 时根据 p 的数量（gomaxprocs）做了负载均衡。*
- 如果没有从全局运⾏队列中获取到 goroutine，那么接下来就在⼯作线程的本地运⾏队列中寻找需要运⾏的goroutine。

**单核线程当本地 P 中的对列满了再添加第 257 个 g 的时候会分配一半的 g 到全局队列中去**.

#### 3.3 从⼯作线程本地运⾏队列中获取 goroutine

```go
func runqget(_p_ *p) (gp *g, inheritTime bool) {
    // If there's a runnext, it's the next G to run.
    for {
        next := _p_.runnext
        if next == 0 {
            break
        }
        if _p_.runnext.cas(next, 0) {
            return next.ptr(), true
        }
    }

    for {
        h := atomic.LoadAcq(&_p_.runqhead) // load-acquire, synchronize with other consumers
        t := _p_.runqtail
        if t == h {
            return nil, false
        }
        gp := _p_.runq[h%uint32(len(_p_.runq))].ptr()
        if atomic.CasRel(&_p_.runqhead, h, h+1) { // cas-release, commits consume
            return gp, false
        }
    }
}
```

⼯作线程的本地运⾏队列分为两个部分

- ⼀部分是由 `p` 的 `runq`、`runqhead` 和 `runqtail` 这三个成员组成的⼀个⽆锁循环队列，**该队列最多可包含 256 个 goroutine**；
- 另⼀部分是 `p` 的 `runnext` 成员，它是⼀个指向 `g` 结构体对象的指针，**它最多只包含⼀个 goroutine**。

从本地运⾏队列中寻找 goroutine 是通过 `runqget` 函数完成的，寻找时代码⾸先查看 `runnext` 成员是否为空，如果不为空则返回 `runnext` 所指的 goroutine，并把 `runnext` 成员清零，如果 `runnext` 为空，则继续从循环队列中查找 goroutine。

1. **⾸先需要注意的是不管是从 `runnext` 还是从循环队列中拿取 goroutine 都使⽤了 `cas` 操作**，这⾥的  `cas` 操作是必需的，因为可能有其他⼯作线程此时此刻也正在访问这两个成员，从这⾥偷取可运⾏的 goroutine。
2. 其次，代码中对 `runqhead` 的操作使⽤了 `atomic.LoadAcq` 和 `atomic.CasRel`，它们分别提供了 `load-acquire` 和 `cas-release` 语义。

对于 `atomic.LoadAcq` 来说，其语义主要包含如下⼏条：

- **原⼦读取**，也就是说不管代码运⾏在哪种平台，保证在读取过程中不会有其它线程对该变量进⾏写⼊；
- 位于 `atomic.LoadAcq` 之后的代码，对内存的读取和写⼊必须在 `atomic.LoadAcq` 读取完成后才能执⾏，编译器和 CPU 都不能打乱这个顺序；
- 当前线程执⾏ `atomic.LoadAcq` 时可以读取到其它线程最近⼀次通过 `atomic.CasRel` 对同⼀个变量写⼊的值，与此同时，位于 `atomic.LoadAcq` 之后的代码，不管读取哪个内存地址中的值，都可以读取到其它线程中位于 `atomic.CasRel`（对同⼀个变量操作）之前的代码最近⼀次对内存的写⼊。

对于 `atomic.CasRel` 来说，其语义主要包含如下⼏条：

- 原⼦的执⾏⽐较并交换的操作；
- 位于 `atomic.CasRel` 之前的代码，对内存的读取和写⼊必须在 `atomic.CasRel` 对内存的写⼊之前完成，编译器和 CPU 都不能打乱这个顺序；
- 线程执⾏ `atomic.CasRel` 完成后其它线程通过 `atomic.LoadAcq` 读取同⼀个变量可以读到最新的值，与此同时，位于 `atomic.CasRel` 之前的代码对内存写⼊的值，可以被其它线程中位于 `atomic.LoadAcq` （对同⼀个变量操作）之后的代码读取到。

*因为可能有多个线程会并发的修改和读取 `runqhead`，以及需要依靠 `runqhead` 的值来读取 `runq` 数组的元素，所以需要使⽤ `atomic.LoadAcq` 和 `atomic.CasRel` 来保证上述语义。*

**为什么读取 p 的 `runqtail` 成员不需要使⽤ `atomic.LoadAcq` 或 `atomic.load`？**
因为 `runqtail` 不会被其它线程修改，只会被当前⼯作线程修改，此时没有⼈修改它，所以也就不需要使⽤原⼦相关的操作。

### 3.x 经典例题

#### 正在执行的 goroutine 什么情况下让出执行权

**正在被执⾏的 goroutine 发⽣以下情况时让出当前 goroutine 的执⾏权，并调度后⾯的goroutine 执⾏**：

- **IO 操作**
- **Channel 阻塞**
- **system call**
- **运⾏较⻓时间**

如果⼀个 goroutine 执⾏时间太⻓，`scheduler` 会在其 `G` 对象上打上⼀个标志（preempt），当这个 goroutine 内部发⽣函数调⽤的时候，会先主动检查这个标志，如果为 `true` 则会让出执⾏权。

**main 函数⾥启动的 goroutine 其实是⼀个没有 IO 阻塞、没有 Channel 阻塞、没有system call、没有函数调⽤的死循环**。

也就是，它⽆法主动让出⾃⼰的执⾏权，即使已经执⾏很⻓时间，`scheduler` 已经标志了 `preempt`。⽽ golang 的 GC 动作是需要所有正在运⾏ goroutine 都停⽌后进⾏的。因此，程序会卡在 `runtime.GC()` 等待所有协程退出。

#### 优先调度

```go
func main() {
    runtime.GOMAXPROCS(1)
    wg := sync.WaitGroup{}
    wg.Add(20)
    for i := 0; i < 10; i++ {
        go func() {
            fmt.Println("i: ", i)
            wg.Done()
        }()
    }
    for i := 0; i < 10; i++ {
        go func(i int) {
            fmt.Println("i: ", i)
            wg.Done()
        }(i)
    }
    wg.Wait()
}
```

这个输出结果决定来⾃于调度器优先调度哪个 `G`。从 `runtime` 的源码可以看到，**当创建⼀个 `G` 时，会优先放⼊到下⼀个调度的 `runnext` 字段上作为下⼀次优先调度的 `G`。因此，最先输出的是最后创建的 `G`**，也就是9.
