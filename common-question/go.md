# go

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

### 1.4 select

golang 的 select 就是监听 IO 操作，当 IO 操作发生时，触发相应的动作每个case语句里必须是一个IO操作，确切的说，应该是一个面向 channel 的IO操作。

1. select 语句只能用于信道的读写操作
2. select 中的 case 条件(非阻塞)是并发执行的，select 会选择先操作成功的那个 case 条件去执行，**如果多个同时返回，则随机选择一个执行**，此时将无法保证执行顺序。对于阻塞的 case 语句会直到其中有信道可以操作，如果有多个信道可操作，会随机选择其中一个 case 执行
3. 对于 case 条件语句中，如果存在通道值为 `nil` 的读写操作（也就是 `var chan int`），则该分支将被忽略，可以理解为从 `select` 语句中删除了这个 `case` 语句; **并且会报 deadlock**
4. 如果有超时条件语句，`case <-time.After(2 * time.Second)`，判断逻辑为如果在这个时间段内一直没有满足条件的 case,则执行这个超时 case。如果此段时间内出现了可操作的 case,则直接执行这个 case。一般用超时语句代替了 default 语句
5. 对于空的 select{}，会引起死锁
6. 对于 for 中的 select{}, 也有可能会引起cpu占用过高的问题，比如增加一个监听退出信号的case 当前是处于阻塞状态，又加一个 default 分支什么都不做，这时候就会莫名拉高 cpu

#### 1.4.1 直接阻塞

**空的 select 语句会直接阻塞当前 Goroutine，导致 Goroutine 进入无法被唤醒的永久休眠状态。**
当 select 结构中不包含任何 case。它直接将类似 `select {}` 的语句转换成调用 `runtime.block` 函数：
`runtime.block` 的实现非常简单，它会调用 `runtime.gopark` 让出当前 Goroutine 对处理器的使用权并传入等待原因 `waitReasonSelectNoCases`。

#### 1.4.2 单一管道

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

#### 1.4.3 非阻塞操作

当 select 中仅包含两个 `case`，并且其中一个是 `default` 时，Go 语言的编译器就会认为这是一次非阻塞的收发操作

##### 发送

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

##### 接收

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

#### 1.4.4 常见流程

在默认的情况下，编译器会使用如下的流程处理 select 语句：

1. 将所有的 case 转换成包含 Channel 以及类型等信息的 `runtime.scase` 结构体；
2. 调用运行时函数 `runtime.selectgo` 从多个准备就绪的 Channel 中选择一个可执行的 `runtime.scase` 结构体；
3. 通过 `for` 循环生成一组 `if` 语句，在语句中判断自己是不是被选中的 case；

##### 初始化

`runtime.selectgo` 函数首先会进行执行必要的初始化操作并决定处理 `case` 的两个顺序 — **轮询顺序 `pollOrder` 和加锁顺序 `lockOrder`**：

- **轮询顺序**：通过 `runtime.fastrandn` 函数引入随机性；
- **加锁顺序**：按照 `Channel` 的地址排序后确定加锁顺序；

**随机的轮询顺序可以避免 Channel 的饥饿问题，保证公平性；而根据 Channel 的地址顺序确定加锁顺序能够避免死锁的发生**。这段代码最后调用的 `runtime.sellock` 会按照之前生成的加锁顺序锁定 `select` 语句中包含所有的 Channel。

##### 循环

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

### 1.5 context

#### 1.5.1 context 引入

多并发情况下：

- 使用 waitgroup 等待协程结束
  - 优点是使用等待组的并发控制模型，尤其适用于好多个goroutine协同做一件事情的时候，因为每个goroutine做的都是这件事情的一部分，只有全部的goroutine都完成，这件事情才算完成；
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

### 1.7 接口

### 1.8 channel

#### 1.8.1 向通道发送数据

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

##### 1.8.1.1 发送数据

- 如果当前 Channel 的 `recvq` 上存在已经被阻塞的 Goroutine，那么会直接将数据发送给当前 Goroutine 并将其设置成下一个运行的 Goroutine,**也就是将接收方的 Goroutine 放到了处理器的 `runnext` 中，程序没有立刻执行该 Goroutine**
    > 向一个非缓冲型的 channel 发送数据、从一个无元素的（非缓冲型或缓冲型但空）的 channel 接收数据，都会导致一个 goroutine 直接操作另一个 goroutine 的栈, 由于 GC 假设对栈的写操作只能发生在 goroutine 正在运行中并且由当前 goroutine 来写, 所以这里实际上违反了这个假设。可能会造成一些问题，**所以需要用到写屏障来规避**
- 如果 Channel 存在缓冲区并且其中还有空闲的容量，我们会直接将数据存储到缓冲区 `sendx` 所在的位置上；
- 如果不满足上面的两种情况，会创建一个 `runtime.sudog` 结构并将其加入 Channel 的 `sendq` 队列中，当前 Goroutine 也会陷入阻塞等待其他的协程从 Channel 接收数据；

#### 1.8.2 接收数据

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

##### 1.8.2.1 从通道接收数据

1. 如果 Channel 为空，那么会直接调用 `runtime.gopark` 挂起当前 Goroutine；
2. 如果 Channel 已经关闭并且缓冲区没有任何数据，`runtime.chanrecv` 会直接返回；
3. 如果 Channel 的 `sendq` 队列中存在挂起的 Goroutine，会将 `recvx` 索引所在的数据拷贝到接收变量所在的内存空间上并将 `sendq` 队列中 Goroutine 的数据拷贝到缓冲区；
4. 如果 Channel 的缓冲区中包含数据，那么直接读取 `recvx` 索引对应的数据；
5. 在默认情况下会挂起当前的 Goroutine，将 `runtime.sudog` 结构加入 `recvq` 队列并陷入休眠等待调度器的唤醒；

#### 1.8.4、3 触发调度时机

1. **发送数据时**
   1. 发送数据时发现 Channel 上存在等待接收数据的 Goroutine，立刻设置处理器的 `runnext` 属性，但是并不会立刻触发调度；
   2. 发送数据时并没有找到接收方并且缓冲区已经满了，这时会将自己加入 Channel 的 `sendq` 队列并调用 `runtime.goparkunlock` 触发 Goroutine 的调度让出处理器的使用权；
2. **从 Channel 接收数据时，会触发 Goroutine 调度的两个时机**：

   1. 当 Channel 为空时；
   2. 当缓冲区中不存在数据并且也不存在数据的发送者时；

#### 1.8.3 最佳实践

- **读已经关闭的 `chan` 能⼀直读到东⻄，但是读到的内容根据通道内关闭前是否有元素⽽不同。**
   1. 如果 `chan` 关闭前，`buffer` 内有元素还未读, 会正确读到 `chan` 内的值，且返回的第⼆个 `bool` 值（是否读成功）为 `true`。
   2. 如果 `chan` 关闭前，`buffer` 内有元素已经被读完，`chan` 内⽆值，接下来所有接收的值都会⾮阻塞直接成功，返回 `channel` 元素的零值，但是第⼆个 `bool` 值⼀直为 `false`。
- **触发 `panic` 的情况**
   1. **向已经关闭的 `chan` 发送数据会 `panic`**
   2. 关闭一个 `nil` 的 channel；
   3. 重复关闭一个 channel。

- 向 `nil` 的通道发送或接收数据会调用 `gopark` 挂起，并陷入永久阻塞

## 2. GC 垃圾回收

### 2.1 GC 策略

1. 内存达到上限触发 GC
就像上面的 `GOGC` 配置那样，当程序达到驻留内存的相应倍数时候，触发 GC， 默认值就是两倍于当前内存。
--
    > GOGC 参数
    `GOGC` 默认值是100，举个例子：你程序的上一次 GC 完，驻留内存是100MB，由于你 GOGC 设置的是100，所以下次你的内存达到 200MB 的时候就会触发一次 GC，如果你 GOGC 设置的是200，那么下次你的内存达到300MB的时候就会触发 GC。

2. 时间达到触发 GC 的阈值
垃圾收集器关注的第二个指标是两个垃圾收集器之间的延迟。如果它没有被触发超过两分钟，一个循环将被强制。
3. **主动触发**，上面两种是**被动触发**，通过调用 `runtime.GC` 来触发 GC，此调用阻塞式地等待当前 GC 运行完毕。

### 2.2 GC 如何调优

- 通过 `go tool pprof` 和 `go tool trace` 等工具
- 控制内存分配的速度，限制 goroutine 的数量，从而提高赋值器对 CPU 的利用率。
- 减少并复用内存，例如使用 `sync.Pool` 来复用需要频繁创建临时对象，例如提前分配足够的内存来降低多余的拷贝。
- 需要时，增大 `GOGC` 的值，降低 GC 的运行频率。

## 3. GMP

### 3.1 schedule 调度

- 为了保证调度的公平性，**每个工作线程每进行 61 次调度就需要优先从全局运行队列中获取 goroutine 出来运行**
    因为如果只调度本地运行队列中的 goroutine，则全局运行队列中的 goroutine 有可能得不到运行
    **所有工作线程都能访问全局队列，所以需要加锁获取 goroutine**
- _如果从全局队列没有获取到 goroutine_，从与 `m` 关联的 `p` 的本地运行队列中获取 goroutine
- **如果从本地运行队列和全局运行队列都没有找到需要运行的goroutine， 则调用 `findrunnable` 函数从其它工作线程的运行队列中偷取**，如果偷取不到，则当前工作线程进入睡眠，直到获取到需要运行的 goroutine 之后 findrunnable 函数才会返回。

### 3.x 经典例题

1. **例1**:

    **正在被执⾏的 goroutine 发⽣以下情况时让出当前 goroutine 的执⾏权，并调度后⾯的goroutine 执⾏**：

    - **IO 操作**
    - **Channel 阻塞**
    - **system call**
    - **运⾏较⻓时间**

    如果⼀个 goroutine 执⾏时间太⻓，`scheduler` 会在其 `G` 对象上打上⼀个标志（preempt），当这个 goroutine 内部发⽣函数调⽤的时候，会先主动检查这个标志，如果为 `true` 则会让出执⾏权。
    **main 函数⾥启动的 goroutine 其实是⼀个没有 IO 阻塞、没有 Channel 阻塞、没有system call、没有函数调⽤的死循环**。
    也就是，它⽆法主动让出⾃⼰的执⾏权，即使已经执⾏很⻓时间，`scheduler` 已经标志了 `preempt`。
    ⽽ golang 的 GC 动作是需要所有正在运⾏ goroutine 都停⽌后进⾏的。因此，程序会卡在 `runtime.GC()` 等待所有协程退出。

2. **例2:**

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

## 4. defer

`defer` 关键字的实现跟 `go` 关键字很类似，不同的是它调⽤的是 `runtime.deferproc` ⽽不是 `runtime.newproc`。 在 `defer` 出现的地⽅，插⼊了指令 `call runtime.deferproc`，然后在函数返回之前的地⽅，插⼊指令 `call runtime.deferreturn`。 goroutine 的控制结构中，有⼀张表记录 `defer`，调⽤ `runtime.deferproc` 时会将需要 `defer` 的表达式记录在表中，⽽在调⽤ `runtime.deferreturn` 的时候，则会依次从 `defer` 表中出栈并执⾏。
因此，题⽬最后输出顺序应该是 defer 定义顺序的倒序。 panic 错误并不能终⽌ defer 的执⾏。
