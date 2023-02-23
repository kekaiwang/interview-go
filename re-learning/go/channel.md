# channel

## 设计原理

先进先出的 FIFO 队列

设计模式是：**不要通过共享内存的方式进行通信，而是通过通信的方式进行共享内存**。

**某种程度来说是用于通信和同步的有锁队列**。

### 介绍 channel

Go 语言中，**不要通过共享内存来通信，而要通过通信来实现内存共享**。Go 的 `CSP`(Communicating Sequential Process)并发模型，中文叫做**通信顺序进程**，是通过 goroutine 和 channel 来实现的。

**channel 收发遵循先进先出 FIFO，分为有缓存和无缓存**，channel 中大致有 `buffer`(当缓冲区大小部位 0 时，是个 `ring buffer`)、`sendx` 和 `recvx` 收发的位置(`ring buffer` 记录实现)、`sendq`、`recvq` 当前 channel 因为缓冲区不足而阻塞的队列、使用双向链表存储、还有一个 `mutex` 锁控制并发、其他原属等。

```go
type hchan struct {
    qcount   uint           // 数组长度，即已有元素个数
    dataqsiz uint           // 数组容量，即可容纳元素个数
    buf      unsafe.Pointer // 数组地址
    elemsize uint16         // 元素大小
    closed   uint32
    elemtype *_type // 元素类型
    sendx    uint   // 下一次写下标位置
    recvx    uint   // 下一次读下标位置
    recvq    waitq  // 读等待队列
    sendq    waitq  // 写等待队列
    lock     mutex
}
```

- `qcount` — Channel 中的元素个数；
- `dataqsiz` — Channel 中的循环队列的长度；
- `buf` — Channel 的缓冲区数据指针；
- `sendx` — Channel 的发送操作处理到的位置；
- `recvx` — Channel 的接收操作处理到的位置；

除此之外，`elemsize` 和 `elemtype` 分别表示当前 Channel 能够收发的元素类型和大小；`sendq` 和 `recvq` 存储了当前 Channel 由于缓冲区空间不足而阻塞的 Goroutine 列表，这些等待队列使用双向链表 `runtime.waitq` 表示，链表中所有的元素都是 `runtime.sudog` 结构：

```go
type waitq struct {
    first *sudog
    last  *sudog
}
```

`runtime.sudog` 表示一个在等待列表中的 Goroutine，该结构中存储了两个分别指向前后 `runtime.sudog` 的指针以构成链表。

**`struct{}` 类型的数据并不占空间，所以缓冲区并不用实际分配。**

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