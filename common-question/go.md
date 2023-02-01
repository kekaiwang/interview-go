[toc]

## 常见基础题

### 高并发

高并发就是提高系统的瞬时并发量，并且不随着时间的积累不降低，不积压并发请求。最核心就是提高单机性能或者单接口的性能。最后是分布式部署应该是用资源换取吞吐量。

1. 代码层面
    减少冗余无用代码的调用，多线程操作，mysql 性能的调优、返回值的大小，避免占用带宽
2. 存储层面
    引入缓存、消息队列
3. 第三方服务调用
    串行调用可以改为并行调用避免时间消耗
4. 最后分布式部署，数据库读写分离、缓存集群等

### 高可用

最大限度保证服务正常健康运行，分布式和集群也提高了高可用

1. 负载均衡
2. 隔离
3. 限流
    nginx 做限流，redis 或 lua 限流，防止单用户恶意的攻击，保证单接口的安全。应用层面隔离
    nginx 做为整体请求数的限流，连接数的限制。
4. 降级
    读服务 写服务降级，

### 1.1 make 和 new 的区别

- **`make`只用于内建类型 `map、slice 和channel` 的内存分配**, 返回初始化后的（非零）值。
    make只能创建 `slice`、`map` 和 `channel`，并且返回一个有初始值(非零)的T类型，而不是 `*T`。对于 `slice`、`map` 和 `channel` 来说，`make` 初始化了内部的数据结构，填充适当的值。

  - 向一个为 `nil` 的 `map` 读写数据，将会导致 `panic`，使用 `make` 可以指定 `map` 的初始空间大小，以容纳元素。如果未指定，则初始空间比较小。
  - 向一个为 `nil` 的 `chan` 读写数据，会导致 `deadlock!`，使用 `make` 可以初始化`chan`，并指定 `chan` 是缓存 `chan（make(chan T,size)）`，还是非缓存 `chan（make(chan T)）`
  - 使用 `make` 可以在创建 `slice` 时，定义切片的 `len`、`cap` 的大小
    使用 `make` 可以创建一个非零值的引用对象

- **`new` 用于各种类型的内存分配**，new 返回指针。
  - `new` 无法初始化对应字段或者结构体的零值
    分配一片零值的内存空间，并返回指向这片内存空间的指针 `value *T`

#### 变量在哪？

- **初始化的全局变量或静态变量**，会被分配在 Data 段。
- 未初始化的全局变量或静态变量，会被分配在 BSS 段。
- 在函数中定义的局部变量，会被分配在堆（Heap 段）或栈（Stack 段）

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
- `slice` 不能直接进行比较，可以使用 `reflect.DeepEqual(x, y interface) bool` 进行深度比较
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

- **消息编码**，消息编码这块，gRPC 使用 `protobuf` 进行消息编码，而 `Restful` 一般使用 `JSON` 进行编码

- **传输协议**，gRPC 使用 `HTTP/2` 作为底层传输协议，而 RestFul 则使用 HTTP 或则其他协议 `https/tcp` 等。

- **传输性能**，由于 gRPC 使用 protobuf 进行消息编码（即序列化），而经 protobuf 序列化后的消息体积很小（传输内容少，传输相对就快）；再加上HTTP/2协议的加持（HTTP1.1的进一步优化），使得gRPC的传输性能要优于Restful。

- **传输形式**，gRPC 最大的优势就是支持流式传输，传输形式具体可以分为四种（unary、client stream、server stream、bidirectional stream），这个后面我们会讲到；而 Restful 是不支持流式传输的。

- **浏览器的支持度**
目前浏览器对gRPC的支持度并不是很好，而对 Restful 的支持可谓是密不可分，这也是 gRPC 的一个劣势，

- **消息的可读性和安全性**，由于gRPC序列化的数据是二进制，且如果你不知道定义的Request和Response是什么，你几乎是没办法解密的，所以gRPC的安全性也非常高，但随着带来的就是可读性的降低，调试会比较麻烦；而Restful则相反（现在有HTTPS，安全性其实也很高）

- **代码的编写**，由于gRPC调用的函数，以及字段名，都是使用stub文件的，所以从某种角度看，代码更不容易出错，联调成本也会比较低，不会出现低级错误，比如字段名写错、写漏。

总的来说：

1. **gRPC 主要用于公司内部的服务调用，性能消耗低，传输效率高，服务治理方便。**
2. **Restful 主要用于对外，比如提供接口给前端调用，提供外部服务给其他人调用等，**

### 1.8 go 内存泄漏

**go 中的内存泄露一般都是 `goroutine` 泄露，就是 `goroutine` 没有被关闭，或者没有添加超时控制，让 `goroutine` 一只处于阻塞状态，不能被 `GC`**。

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
`O_RDWR、O_CREATE、O_TRUNC` 是file.go文件中定义好的一些常量，标识文件以什么模式打开，常见的模式有读写，只写，只读，权限依次降低。

看到 syscall 对于文件的操作进行了封装，继续进入

```go
func Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno)
func Syscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err Errno)
func RawSyscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno)
func RawSyscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err Errno)
```

这里，看到相似的函数有四个，后面没数字的，就是4个参数，后面为6的，就是6个参数。调用的是操作系统封装好的API。

### runtime 包中的常用方法

1. `NumCPU`: 返回当前系统的 CPU 核数量
2. `GOMAXPROCS`: 设置最大的可同时使用的 CPU 核数

    > 通过 `runtime.GOMAXPROCS` 函数，应用程序何以在运行期间设置运行时系统中的 P 最大数量。但这会引起 “Stop the World”。所以，应在应用程序最早的调用。并且最好是在运行Go程序之前设置好操作程序的环境变量 GOMAXPROCS，而不是在程序中调用 runtime.GOMAXPROCS 函数。

    无论我们传递给函数的整数值是什么值，运行时系统的P最大值总会在1~256之间。

3. `Gosched`: 让当前线程让出 cpu 以让其它线程运行,它不会挂起当前线程，因此当前线程未来会继续执行  
    这个函数的作用是让当前 goroutine 让出 CPU，当一个 goroutine 发生阻塞，Go 会自动地把与该 goroutine 处于同一系统线程的其他 goroutine 转移到另一个系统线程上去，以使这些 goroutine 不阻塞。
4. `Goexit`: 退出当前 goroutine(但是 `defer` 语句会照常执行)
5. `NumGoroutine`: 返回正在执行和排队的任务总数  
  runtime.NumGoroutine 函数在被调用后，会返回系统中的处于特定状态的 Goroutine 的数量。这里的特指是指 Grunnable\Gruning\Gsyscall\Gwaition。处于这些状态的 Groutine 即被看做是活跃的或者说正在被调度。
注意：垃圾回收所在 Groutine 的状态也处于这个范围内的话，也会被纳入该计数器。
6. `GOOS`: 目标操作系统
7. `GC`: 会让运行时系统进行一次强制性的垃圾收集
  1.强制的垃圾回收：不管怎样，都要进行的垃圾回收。2.非强制的垃圾回收：只会在一定条件下进行的垃圾回收（即运行时，系统自上次垃圾回收之后新申请的堆内存的单元（也成为单元增量）达到指定的数值）。
8. `GOROOT`: 获取 goroot 目录
9. `GOOS`: 查看目标操作系统

### golang中值类型和引用类型

**值类型分别有**：int系列、float系列、bool、string、数组和结构体
**引用类型有**：指针、slice切片、管道channel、接口interface、map、函数等

值类型的特点是：变量直接存储值，内存通常在栈中分配
引用类型的特点是：变量存储的是一个地址，这个地址对应的空间里才是真正存储的值，内存通常在堆中分配

### json.Marshal

`channel、function、complex`、循环引用的数据类型不能被 `json.Marshal` 序列化

### atomic

```go
for {
    v := value
    if atomic.CompareAndSwapInt32(&value, v, (v+delta)) {
        break
    }
}
```

- `atomic.CompareAndSwapInt32` 不需要 `for` 循环调用

### byte

```go
var i byte
go func() {
    for i = 0; i <= 255; i++ {

    }
}()
```

`byte` 其实被 `alias` 到 `uint8` 上了。所以上⾯的 `for` 循环会始终成⽴，因为 `i++` 到 `i=255` 的时候会溢出，i <= 255 ⼀定成⽴。

### 接口返回的数据量特别大怎么处理？不能分页

1. 按照对源码的理解，可以得知在使用 `io.pipe()` 方法进行流式传输时
2. gzip 压缩返回二进制数据

### goroutine 和协程区别

本质上，goroutine 就是协程。 **不同的是，Golang 在 runtime、系统调用等多方面对 goroutine 调度进行了封装和处理，当遇到长时间执行或者进行系统调用时，会主动把当前 goroutine 的CPU (P) 转让出去，让其他 goroutine 能被调度并执行，也就是 Golang 从语言层面支持了协程**。
Golang 的一大特色就是从语言层面原生支持协程，在函数或者方法前面加 go 关键字就可创建一个协程。

其他方面的比较

1. 内存消耗方面
    - 每个 goroutine (协程) 默认占用内存远比 Java 、C 的线程少。
    - goroutine：2KB
    - 线程：8MB

2. 线程和 goroutine 切换调度开销方面
    - 线程/goroutine 切换开销方面，goroutine 远比线程小
    - 线程：涉及模式切换(从用户态切换到内核态)、16个寄存器、PC、SP...等寄存器的刷新等。
    - goroutine：只有三个寄存器的值修改 - PC / SP / DX.

pc：也就是 x86 下的 ip 指令寄存器
SP：栈指针寄存器

### gopark 挂起的过程

park_m:

- gopark 通过 mcall 将当前线程的堆栈切换到 g0 的堆栈
- 保存当前 goroutine 的上下文（pc、sp寄存器->g.sched）
- 在 g0 栈上，调用 park_m
- 将当前的 g 从 running 状态设置成 waiting 状态
- 通过 dropg 来解除 m 和 g 的关系

### 并发安全的类型

- 由一条机器指令完成赋值的类型并发赋值是安全的，这些类型有：**字节型，布尔型、整型、浮点型、字符型、指针、函数**
- 数组由一个或多个元素组成，大部分情况并发不安全。注意：当位宽不大于 64 位且是 2 的整数次幂（8，16，32，64），那么其并发赋值是安全的
- struct 或底层是 struct 的类型并发赋值大部分情况并发不安全，这些类型有：**复数、字符串、 数组、切片、映射、通道、接口**。
注意：当 struct 赋值时退化为单个字段由一个机器指令完成赋值时，并发赋值又是安全的

    ```go
    // 此时可以实现结构体的并发安全
    var v atomic.Value
    v.Store(Test{1,2})
    ```

### 产生 panic 的情况

1. slice 访问越界
2. 重复关闭 chnnel 、关闭 nil 的 channel、向已关闭的 channel 发送数据
3. 死锁产生的 panic
4. map 并发读写产生 panic
5. 类型断言不捕获 bool 类型产生 panic
6. 空指针的属性访问产生 panic

## string

字符串拼接最好使用下面缓冲的形式拼接

```go
var buffer bytes.Buffer

for i := 0; i < 500; i++ {
    buffer.WriteString("hello,world")
}
fmt.Println(buffer.String())
```

## 类型系统

**Go语言中每种类型都有对应的类型元数据，类型元数据都有一个相同的Header，就是 `runtime._type`**。

内置类型、自定义类型
不能给内置类型定义方法，接口类型是无效的方法接收者。

### 类型元数据

类型元数据：内置类型和自定义类型的类型描述信息，每种类型元数据都是全局唯一的。这些类型元数据构成了 Go 语言的 “类型系统”。

`runtime._type` 包含数据的基本 `header` 信息，后面跟随其他描述信息，比如 `slice` 类型的类型元数据等，**如果是自定义类型后面还跟随 `uncommontype` 结构体**。

![image](https://mail.wangkekai.cn/1644829502867.jpg)

`moff` 记录的是这些方法的元数据组成的数组，相对于这个 `uncommontype` 结构体偏移了多少字节。
方法元数据结构如下：

```go
type method struct {
    name nameOff
    mtyp typeOff
    ifn  textOff
    tfn  textOff
}
```

- `type U int32` 基于 `int` 定义的新类型，有属于自己的类型元数据。
- `type U2 = int32` 是 `int32` 的别名，等价于 int；U2 和 `int32` 会关联到同一个类型元数据属于同一种类型。
**`rune` 和 `int32`、`byte` 和 `uint8` 就是这样的关系**。

下面所示非空接口 `r` 的静态类型是 `io.Reader`，动态类型是 `*os.File`。

```go
var f *os.File
var r io.Reader = f
```

#### type.alg

在 `_type.alg` 这里记录了该类型的两个函数：`hash` 和 `equal`。

```go
type typeAlg struct {
    hash  func(unsafe.Pointer, uintptr) uintptr
    equal func(unsafe.Pointer, unsafe.Pointer) bool
}
```

**而不可比较的类型，例如 `slice`，它的类型元数据里是没有提供可用的 equal 方法的。**

## 接口

### 自动检测类型是否实现接口

```go
var _ io.Writer = (*myWriter)(nil)
var _ io.Writer = myWriter{}
```

上述赋值语句会发生隐式地类型转换，在转换的过程中，编译器会检测等号右边的类型是否实现了等号左边接口所规定的函数。

### 装箱

1. `interface{}` 被设计成一个容器，但它本质上是指针，可以直接装载地址，用来实现装载值的话，实际的内存要分配在别的地方，并把内存地址存储在这里。（convT64的作用就是分配这个存储值的内存空间，实际上 `runtime` 中有一系列的这类函数，如convT32、convTstring和convTslice等。）

2. 通过 `staticuint64s` 这种优化方式，能够反向推断出：被 `convT64` 分配的这个 `uint64`，**它的值在语义层面是不可修改的，是个类似 `const` 的常量，这样设计主要是为了跟 `interface{}` 配合来模拟“装载值”**。

3. **至于为什么这个值不可修改，因为interface{}只是一个容器，它支持把数据装入和取出，但是不支持直接在容器里修改**。这有些类似于Java和C#里的自动装箱，只不过interface{}是个万能包装类。

**因为接口装载的动态类型是可以变化的，所以通过接口调用它的方法时，需要根据它背后的动态类型来确定究竟调用哪一种实现**，这也是面向对象编程中，接口的一个核心功能：**接口的一个核心功能：实现“多态”，也就是实现方法的“动态派发”**.

### 动态派发

对于动态派发来讲，编译阶段能够确定的有：
（1）**要调用的方法的名字**；
（2）**方法的原型（参数与返回值列表）**。

实现动态派发要使用的函数地址表就存储在 `itab` 中。

```go
type itab struct {
    inter *interfacetype
    _type *_type
    hash  uint32
    _     [4]byte
    fun   [1]uintptr
}
```

（1）`itab.inter` 指向当前接口的类型元数据，记录着接口要求实现的方法列表；
（2）`itab._type` 指向动态类型元数据，从 _type 到 `uncommontype`，再到 `[mcount]method`，可以找到该动态类型的方法集。

**接口要求实现的方法列表与动态类型的方法集都是有序的，所以经过一次循环对比就可以确定该动态类型是否实现了特定接口**：
（1）如果没有实现，那么 `itab.fun[0]` 就等于0；
（2）如果实现了，就把动态类型实现的方法地址存储到itab.fun数组中。
有了这样的itab，再通过接口调用方法时，就可以像C++的虚函数那样直接按数组下标读取地址了。

### 包装方法

编译器会为接收者为值类型的方法生成接收者为指针类型的方法，也就是所谓的“**包装方法**”。

**接口是不能直接使用值接收者方法的，这就是编译器生成包装方法的根本原因。**

（1）无论是嵌入值还是嵌入指针，值接收者方法始终能够被继承；
（2）只有在能够拿到嵌入对象的地址时，才能继承指针接收者方法。

## 数组

Go 语言数组在初始化之后大小就无法改变，存储元素类型相同、但是大小不同的数组类型在 Go 语言看来也是完全不同的，只有两个条件都相同才是同一类型。

## slice

```go
s2 := s1[2:6:7]
```

长度 s2 从 s1 的索引2（闭区间）到索引6（开区间，元素真正取到索引5），容量到索引7（开区间，真正到索引6），为5。

### append

- append函数返回的是一个切片，append在原切片的末尾添加新元素，这个末尾是切片长度的末尾，不是切片容量的末尾。

    ```go
    func test() {
        a := make([]int, 0, 4)
        b := append(a, 1) // b=[1], a指向的底层数组的首元素为1，但是a的长度和容量不变
        c := append(a, 2) // a的长度还是0，c=[2], a指向的底层数组的首元素变为2
        fmt.Println(a, b, c) // [] [2] [2]
    }
    ```

- **如果原切片的容量足以包含新增加的元素，那 append 函数返回的切片结构里3个字段的值是**：

  - array 指针字段的值不变，和原切片的 array 指针的值相同，也就是 append 是在原切片的底层数组返回的切片还是指向原切片的底层数组
  - len 长度字段的值做相应增加，增加了 N 个元素，长度就增加 N
  - cap 容量不变

- **如果原切片的容量不够存储 append 新增加的元素，Go 会先分配一块容量更大的新内存**，然后把原切片里的所有元素拷贝过来，最后在新的内存里添加新元素。append 函数返回的切片结构里的3个字段的值是：

  - array 指针字段的值变了，不再指向原切片的底层数组了，会指向一块新的内存空间
  - len 长度字段的值做相应增加，增加了 N 个元素，长度就增加 N
  - cap 容量会增加

**append 不会改变原切片的值，原切片的长度和容量都不变，除非把 append 的返回值赋值给原切片**。

slice 是否传引用根据底层数组是否变化决定，比如

```go
a := make([]int, 0, 5)

func appendSlice(a []int) {
    a = append(a, 5)
}
```

上面这种情况外面的 slice 不受函数的影响！
切记跟随地层数组是否变动决定的。

**`var s []int` 此时切片为 `nil`**，指向的底层数组为 nil，空切片指向的底层数组有分配的地址。

### 常见经典题目

```go
func main() {
    s := make([]int, 5)
    s = append(s, 1, 2, 3)
    fmt.Println(s)
}

// 0 0 0 0 0 1 2 3
```

**`make` 在初始化切⽚时指定了⻓度**，所以追加数据时会从 `len(s)` 位置开始填充数据

### 数据结构

```go
type SliceHeader struct {
    Data uintptr
    Len  int
    Cap  int
}
```

- Data 是指向数组的指针;
- Len 是当前切片的长度；
- Cap 是当前切片的容量，即 Data 数组的大小可以存多少元素

#### 初始化切片

```go
    arr := arr[1:3]             // 下标
    arr := []int{1, 2, 3}       // 字面量
    arr := make([]int, 2, 3)    // 关键字
```

1. **使用下标初始化切片** 不会拷贝原数组或者原切片中的数据，它只会创建一个指向原数组的切片结构体，所以修改新切片的数据也会修改原切片。
2. **字面量初始化切片** 是在编译时完成的。
    1. 根据切片中的元素数量对底层数组的大小进行推断并创建一个数组
    2. 将这些字面量元素存储到初始化的数组中
    3. 创建一个同样指向 [3]int 类型的数组指针
    4. 将静态存储区的数组 vstat 赋值给 vauto 指针所在的地址
    5. 通过 [:] 操作获取一个底层使用 vauto 的切片
3. **关键字** 创建切片时，很多工作都需要运行时的参与。不仅会检查 len 是否传入，还会保证传入的容量 cap 一定大于或者等于 len。

当切片发生逃逸或者非常大时，运行时需要 runtime.makeslice 在堆上初始化切片，如果当前的切片不会发生逃逸并且切片非常小的时候，会直接使用下标得到得到数组对应的切片。

##### 整形切片

```go
var ints []int
```

slice 的元素要存在连续的内存中，也就是连续数组。 `data` 是底层数组的起始地址,这里只分配了切片结构没有分配底层数组，此时 `data = nil`，`Len 和 Cap` 都为零。

```go
var arr []int = make([]int, 2, 5)
arr = append(arr, 1)
arr[0] = 1
```

**通过 `make` 的方式定义变量，不仅会分配结构还会开辟一段内存作为它的底层数组**。此时分配的值都为 0。通过 `append` 之后此时索引 2 的位置被修改为 1，通过索引下标 0 修改后第一个元素为 1， 其他位置还是默认值 0。

##### 字符串类型切片

```go
arr := new([]string)

*arr = append(*arr, "kevin")
// append(arr, "kevin") 会报错 
// invalid argument: arr (variable of type *[]string) is not a slicecompiler
// 因为 new 返回的是指针起始地址
```

上面 `new` 一个 slice 对象同样会分配切片的三部分，它不负责底层数组的分配，new 的返回值是 slice 的指针地址，如果这时候 `(*arr)[0] = "kevin"` 通过下标修改切片内容是不允许的，此时可以通过 append 进行分配底层数组。

##### 和字符串相关的底层数组

底层数组是相同类型的元素一个挨一个的存储，不同的slice 可以关联到同一个数组。slice 的 data 起始指针并不一定指向数组的开头，如下例：

```go
arr := [10]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

var s1 = arr[1:4]
var s2 = arr[7:]
```

s1 的元素是 arr 索引 1到4 的元素，**左闭又开**，长度是 3，但是容量 `Cap` 是从索引 1 开始到末尾 为 9，s2 的元素是从 索引 7 到末尾，总共三个元素容量 Cap 也是 3。slice 访问和修改的都是底层数组的元素。

```go
s1[3] = 5
```

上面 s1 就会越界产生 panic，只能扩大读写区间范围。
此时如果给 s2 添加元素 `s2 = append(s2, 10)` 会开辟新数组，原来的元素要拷过来同时添加新元素，元素个数改为 4，容量扩到 6。

### 扩容规则

如果 `append` 返回的*新切片不需要赋值回原有的变量*，就会 `makeslice` 创建一个新的 slice；如果使用 `slice = append(slice, 1, 2, 3)` 语句，那么 `append` 后的切片会覆盖原切片。

```go
// append(slice, 1, 2, 3)
ptr, len, cap := slice
newlen := len + 3
if newlen > cap {
    ...
}
*(ptr+len) = 1
*(ptr+len+1) = 2
*(ptr+len+2) = 3
return makeslice(ptr, newlen, cap)

// slice = append(slice, 1, 2, 3)
a := &slice
... // 同上
if uint(newlen) > uint(cap) {
   *a.cap = newcap
   *a.ptr = newptr
}
newlen = len + 3
*a.len = newlen
*(ptr+len) = 1
*(ptr+len+1) = 2
*(ptr+len+2) = 3
```

#### 预估容量

```go
arr := []int{1, 2}
arr = append(arr, 3, 4, 5)
```

上面扩容后容量到 5，因为整形元素占有 8 字节，根据内存规格匹配到 48 。下面分析为什么？

在分配内存空间之前需要先确定新的切片容量，运行时根据切片的当前容量选择不同的策略进行扩容：

1. 如果期望容量大于当前容量的两倍就会使用期望容量（也就是当前容量翻倍）
2. 如果当前切片的长度小于 1024 就会将容量翻倍
3. 如果当前切片的长度大于 1024 就会每次增加 25% 的容量，直到新容量大于期望容量

#### 内存分配

**内存空间=切片中元素大小×目标容量**。

语言的内存管理模块会提前向操作系统申请一批内存，分成常用规格管理起来。当申请内存时会匹配合适的规格进行分配。

```go
var class_to_size = [_NumSizeClasses]uint16{
    0,
    8,
    16,
    32,
    48,
    64,
    80,
    ...,
}
```

例子如下：

```go
a := []string{"my", "name", "is"}
a = append(a, "kevin")
```

1. 字符串在 64 位机器上每个元素占 16 字节，扩容前容量是 3，添加一个最少扩容到4，原容量翻倍等于 6 大于 4，小于1024，直接翻倍预估容量为 6。
2. 预估容量\*元素大小（6*16=96byte）
3. 匹配到内存规格 96 所以最终扩容后容量为 6

### 切片 copy

无论是编译期间拷贝还是运行时拷贝，两种拷贝方式都会通过 `runtime.memmove` 将整块内存的内容拷贝到目标的内存区域中。

## map

- **删除 map 不存在的键值对时，不会报错，相当于没有任何作用**
- **获取不存在的键值对时，返回值类型对应的零值，所以返回 0**

### map 结构 hmap

`map` 类型的变量本质上是一个 `hmap` 类型的指针：

```go
type hmap struct {
    count     int    //已经存储的键值对个数
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
    overflow    *[]*bmap //把已经用到的溢出桶链起来
    oldoverflow *[]*bmap //渐进式扩容时，保存旧桶用到的溢出桶
    nextOverflow *bmap   //下一个尚未使用的溢出桶
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
**每个键值对的tophash、key和value的索引顺序一一对应**。这就是map使用的桶的内存布局。

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

## for range

```go
    ha := a
    hv1 := 0
    hn := len(ha)
    v1 := hv1
    v2 := nil
    for ; hv1 < hn; hv1++ {
        tmp := ha[hv1]
        v1, v2 = hv1, tmp
        ...
    }
```

对于所有的 `range` 循环，Go 语言都会在编译期将原切片或者数组赋值给一个新变量 `ha`，在赋值的过程中就发生了拷贝，而我们又***通过 `len` 关键字预先获取了切片的长度*，所以在循环中追加新的元素也不会改变循环执行的次数**，这也就解释了循环永动机一节提到的现象。

而遇到这种同时遍历索引和元素的 `range` 循环时，Go 语言会额外创建一个新的 v2 变量存储切片中的元素，**循环中使用的这个变量 v2 会在每一次迭代被重新赋值而覆盖，赋值时也会触发拷贝**。

### 遍历 hash 表

```go
func mapiterinit(t *maptype, h *hmap, it *hiter) {
    it.t = t
    it.h = h
    it.B = h.B
    it.buckets = h.buckets

    r := uintptr(fastrand())
    it.startBucket = r & bucketMask(h.B)
    it.offset = uint8(r >> h.B & (bucketCnt - 1))
    it.bucket = it.startBucket
    mapiternext(it)
}
```

**该函数会初始化 `runtime.hiter` 结构体中的字段，并通过 `runtime.fastrand` 生成一个随机数帮助我们随机选择一个遍历桶的起始位置**。Go 团队在设计哈希表的遍历时就不想让使用者依赖固定的遍历顺序，所以引入了随机数保证遍历的随机性。

## context

一个接口，四种具体实现，六个函数。

### context 引入

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

### Context的继承衍生

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

### context 最佳实践

- 不要将 Context 塞到结构体里。直接将 Context 类型作为函数的第一参数，而且一般都命名为 `ctx`。
- 不要向函数传入一个 `nil` 的 context，如果你实在不知道传什么就用 `context：todo`。
- 不要把本应该作为函数参数的类型塞到 context 中，context 存储的应该是一些共同的数据。例如：登陆的 session、cookie 等。
- 同一个 context 可能会被传递到多个 goroutine，别担心，context 是并发安全的。

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
`runtime.block` 的实现非常简单，它会**调用 `runtime.gopark` 让出当前 Goroutine 对处理器的使用权并传入等待原因 `waitReasonSelectNoCases`**。

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

- **轮询顺序**：**通过 `runtime.fastrand` 函数引入随机性**；
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

### atomic 与 sync

**atomic 包的原子操作是通过 CPU 指令，也就是在硬件层次去实现的，性能较好**，不需要像 mutex 那样记录很多状态。

**atomic的add操作（硬件层面的原子操作），同一时间点只让一个cpu不可中断的执行这些指令．**

### 2.1 mutex

- Mutex 是互斥锁同一线程内不能重复加锁

- `sync/rwmutex.go` 中注释可以知道，读写锁当有⼀个协程在等待写锁时，其他协程是不能获得读锁的

- **加锁后复制变量，会将锁的状态也复制，所以 mu1 其实是已经加锁状态，再加锁会死锁 `panic`**。

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

- 如果该函数返回的新状态等于 0，当前 Goroutine 就成功解锁了互斥锁；
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

- **`WaitGroup` 在调⽤ `Wait` 之后是不能再调⽤ `Add` ⽅法的**。

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

## 3. defer

Go 语言的 `defer` 是一个很方便的机制，能够把某些函数调用推迟到当前函数返回前才实际执行。**我们可以很方便的用defer关闭一个打开的文件、释放一个Redis连接，或者解锁一个Mutex**。而且Go语言在设计上保证，即使发生panic，所有的defer调用也能够被执行。不过多个defer函数是按照定义顺序倒序执行的。

`defer` 关键字的实现跟 `go` 关键字很类似，不同的是它调⽤的是 `runtime.deferproc` ⽽不是 `runtime.newproc`。
在 `defer` 出现的地⽅，插⼊了指令 `call runtime.deferproc`，然后在函数返回之前的地⽅，插⼊指令 `call runtime.deferreturn`。
goroutine 的控制结构中，有⼀张表记录 `defer`，调⽤ `runtime.deferproc` 时会将需要 `defer` 的表达式记录在表中，⽽在调⽤ `runtime.deferreturn` 的时候，则会依次从 `defer` 表中出栈并执⾏。
因此，题⽬最后输出顺序应该是 `defer` 定义顺序的倒序。 `panic` 错误并不能终⽌ `defer` 的执⾏。

```go
type _defer struct {
    siz       int32
    started   bool
    sp        uintptr // sp at time of defer
    pc        uintptr
    fn        *funcval
    _panic    *_panic // panic that is running defer
    link      *_defer

    heap      bool       // 1.13 中新加 标识是否为堆分配

    openDefer bool           //1
}
```

`runtime._defer` 结构体是延迟调用链表上的一个元素，所有的结构体都会通过 `link` 字段串联成链表。

- `siz` 是参数和结果的内存大小；
- `sp` 和 `pc` 分别代表栈指针和调用方的程序计数器；
- `fn` 是 `defer` 关键字中传入的函数；
- `_panic` 是触发延迟调用的结构体，可能为空；
- `openDefer` 表示当前 `defer` 是否经过开放编码的优化；

### 版本演进

- `defer1.12` 的性能问题主要缘于两个方面
    1. `_defer` 结构体堆分配，即使有预分配的 `deferpool`，也需要去堆上获取与释放。而且 `defer` 函数的参数还要在注册时从栈拷贝到堆，执行时又要从堆拷贝到栈。
    2. `defer` 信息保存到链表，而链表操作比较慢。

- 1.13 版本中并不是所有 `defer` 都能够在栈上分配。循环中的 defer，无论是显示的 `for` 循环，还是 `goto` 形成的隐式循环，都只能使用 1.12版本中的处理方式在堆上分配。
**即使只执行一次的 `for` 循环也是一样**。

- Go1.14 通过增加一个标识变量 **`df`** 来解决这类问题。用 df 中的每一位对应标识当前函数中的一个 `defer` 函数是否要执行。
函数A1要被执行，所以就通过 `df |= 1` 把 `df` 第一位置为 1；在函数返回前再通过 `df&1` 判断是否要调用函数 A1。
Go1.14把 defer 函数在当前函数内展开并直接调用，这种方式被称为 open coded defer。这种方式不仅不用创建 `_defer` 结构体，也脱离了 `defer` 链表的束缚。不过这种方式依然不适用于循环中的 defer，所以1.12版本 defer 的处理方式是一直保留的。

一旦发生 `panic` 或者调用了 `runtime.Goexit` 函数，在这之后的正常逻辑就都不会执行了，而是直接去执行 `defer` 链表。**那些使用open coded defer在函数内展开，因而没有被注册到链表的defer函数要通过栈扫描的方式来发现。**

## panic & recover

### 关键现象

- `panic` 只会触发当前 Goroutine 的 `defer`
- `recover` 只有在 `defer` 中调用才会生效
`recover` 只有在发生 `panic` 之后调用才会生效，所以我们需要在 `defer` 中使用 `recover` 关键字。
- `panic` 允许在 `defer` 中嵌套多次调用

```go
type _panic struct {
    argp      unsafe.Pointer
    arg       interface{}
    link      *_panic
    recovered bool
    aborted   bool
    pc        uintptr
    sp        unsafe.Pointer
    goexit    bool
}
```

- argp 是指向 defer 调用时参数的指针；
- arg 是调用 panic 时传入的参数；
- link 指向了更早调用的 runtime._panic 结构；
- recovered 表示当前 runtime._panic 是否被 recover 恢复；
- aborted 表示当前的 panic 是否被强行终止；

注意 panic 打印异常信息时，会打印此时 panic 链表中剩余的所有链表项。
**不过，并不是从链表头开始，而是从链表尾开始，按照链表项的插入顺序逐一输出**。

### 没有 recover 发生的 panic

**recover 只有在发生 panic 之后调用才会生效**

没有recover发生的panic处理逻辑就算梳理完了，理解这个过程的关键点有两个：

1. panic 执行 defer 函数的方式，先标记，后移除，目的是为了终止之前工作的 panic；
2. panic 异常信息：所有还在 panic 链表上的链表项都会被输出，顺序与 panic 发生的顺序一致。

### revoce



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

## 垃圾回收

追踪式标记清除

### GC 的根对象是什么？

1. **全局变量**：程序在编译期就能确定的那些存在于程序整个生命周期的变量。
2. **执行栈**：每个 goroutine 都包含自己的执行栈，这些执行栈上包含栈上的变量及指向分配的堆内存区块的指针。
3. **寄存器**：寄存器的值可能表示一个指针，参与计算的这些指针可能指向某些赋值器分配的堆内存区块。

**扫描根对象需要使用 `runtime.markroot`，该函数会扫描缓存、数据段、存放全局变量和静态变量的 `BSS` 段以及 Goroutine 的栈内存.**

### 三色标记

> 三色抽象只是一种描述追踪式回收器的方法，在实践中并没有实际含义，它的重要作用在于从逻辑上严密推导标记清理这种垃圾回收方法的正确性。

垃圾回收器的视角来看，三色抽象规定了三种不同类型的对象，并用不同的颜色相称：

- **白色对象**（可能死亡）：未被回收器访问到的对象。在回收开始阶段，所有对象均为白色，当回收结束后，白色对象均不可达。
- **灰色对象**（波面）：已被回收器访问到的对象，但回收器需要对其中的一个或多个指针进行扫描，因为他们可能还指向白色对象。
- **黑色对象**（确定存活）：已被回收器访问到的对象，其中所有字段都已被扫描，黑色对象中任何一个指针都不可能直接指向白色对象。

这样三种不变性所定义的回收过程其实是一个波面不断前进的过程，这个波面同时也是黑色对象和白色对象的边界，灰色对象就是这个波面。

**悬挂指针**，即指针没有指向特定类型的合法对象，影响了内存的安全性，想要并发或者增量地标记对象还是需要使用屏障技术。

### 屏障技术

- **强三色不变性** — 黑色对象不会指向白色对象，只会指向灰色对象或者黑色对象
- **弱三色不变性** — 黑色对象指向的白色对象必须包含一条从灰色对象经由多个白色对象的可达路径

#### 插入写屏障

```go
writePointer(slot, ptr):
    shade(ptr)
    *slot = ptr
```

Dijkstra 的插入写屏障是一种相对保守的屏障技术，**它会将有存活可能的对象都标记成灰色以满足强三色不变性**。

垃圾收集和用户程序交替运行时将黑色对象 A 指向白色对象 B，会将对象 B 标记成灰色，完成插入写屏障也就满足了强三色不变性。

插入式的 Dijkstra 写屏障虽然实现非常简单并且也能保证强三色不变性，但是它也有明显的缺点。因为栈上的对象在垃圾收集中也会被认为是根对象，所以为了保证内存的安全，Dijkstra 必须为栈上的对象增加写屏障或者在标记阶段完成重新对栈上的对象进行扫描，这两种方法各有各的缺点，前者会大幅度增加写入指针的额外开销，后者重新扫描栈对象时需要暂停程序，垃圾收集算法的设计者需要在这两者之间做出权衡。

#### 删除写屏障

会在老对象的引用被删除时，将白色的老对象涂成灰色，这样删除写屏障就可以保证弱三色不变性，老对象引用的下游对象一定可以被灰色对象引用。

垃圾收集和用户程序交替运行时黑色对象 A 指向灰色对象 B，B 指向白色对象 C，此时 A 指向 C，因为灰色对象 B 指向白色对象 C，此时不触发删除写屏障，如果将灰色对象 B 原本指向白色对象 C 的指针删除，会将对象 C 标记成灰色，完成删除写屏障也就满足了强/弱三色不变性。

#### 混合写屏障

在 Go 语言 v1.7 版本之前，运行时会使用 Dijkstra 插入写屏障保证强三色不变性，但是运行时并没有在所有的垃圾收集根对象上开启插入写屏障。因为应用程序可能包含成百上千的 Goroutine，而垃圾收集的根对象一般包括全局变量和栈对象，如果运行时需要在几百个 Goroutine 的栈上都开启写屏障，会带来巨大的额外开销，**所以 Go 团队在实现上选择了在标记阶段完成时暂停程序、将所有栈对象标记为灰色并重新扫描，在活跃 Goroutine 非常多的程序中，重新扫描的过程需要占用 10 ~ 100ms 的时间**。

**会将被覆盖的对象标记成灰色并在当前栈没有扫描时将新对象也标记成灰色**。

为了移除栈的重扫描过程，除了引入混合写屏障之外，**在垃圾收集的标记阶段，我们还需要将创建的所有新对象都标记成黑色，防止新分配的栈内存和堆内存中的对象被错误地回收**，因为栈内存在标记阶段最终都会变为黑色，所以不再需要重新扫描栈空间。

### 增量和并发

- **增量垃圾收集** — 增量地标记和清除垃圾，降低应用程序暂停的最长时间

    增量式（Incremental）的垃圾收集是减少程序最长暂停时间的一种方案，它可以将原本时间较长的暂停时间切分成多个更小的 GC 时间片，虽然从垃圾收集开始到结束的时间更长了，但是这也减少了应用程序暂停的最大时间；

    为了保证垃圾收集的正确性，我们需要在垃圾收集开始前打开写屏障，这样用户程序修改内存都会先经过写屏障的处理，保证了堆内存中对象关系的强三色不变性或者弱三色不变性。虽然增量式的垃圾收集能够减少最大的程序暂停时间，但是增量式收集也会增加一次 GC 循环的总时间，在垃圾收集期间，因为写屏障的影响用户程序也需要承担额外的计算开销，所以增量式的垃圾收集也不是只带来好处的，但是总体来说还是利大于弊。

- **并发垃圾收集** — 利用多核的计算资源，在用户程序执行时并发标记和清除垃圾

    并发（Concurrent）的垃圾收集不仅能够减少程序的最长暂停时间，还能减少整个垃圾收集阶段的时间，通过开启读写屏障、利用多核优势与用户程序并行执行，并发垃圾收集器确实能够减少垃圾收集对应用程序的影响

    虽然并发收集器能够与用户程序一起运行，但是并不是所有阶段都可以与用户程序一起运行，部分阶段还是需要暂停用户程序的，不过与传统的算法相比，并发的垃圾收集可以将能够并发执行的工作尽量并发执行；当然，因为读写屏障的引入，并发的垃圾收集器也一定会带来额外开销，不仅会增加垃圾收集的总时间，还会影响用户程序，这是我们在设计垃圾收集策略时必须要注意的。

### 垃圾收集的阶段

1. 清理终止阶段 **STW**
    1. 暂停程序，所有的处理器在这时会进入安全点（Safe point）；
    2. 如果当前垃圾收集循环是强制触发的，我们还需要处理还未被清理的内存管理单元；
2. 标记阶段
    1. 将状态切换至 `_GCmark`、开启写屏障、用户程序协助（Mutator Assists）并将根对象入队；
    2. 恢复执行程序，标记进程和用于协助的用户程序会开始并发标记内存中的对象，**写屏障会将被覆盖的指针和新指针都标记成灰色，而所有新创建的对象都会被直接标记成黑色**；
    3. 开始扫描根对象，包括所有 Goroutine 的栈、全局对象以及不在堆中的运行时数据结构，扫描 Goroutine 栈期间会暂停当前处理器；
    4. 依次处理灰色队列中的对象，将对象标记成黑色并将它们指向的对象标记成灰色；
    5. 使用分布式的终止算法检查剩余的工作，发现标记阶段完成后进入标记终止阶段；
3. 标记终止阶段 **STW**
    1. 暂停程序、将状态切换至 `_GCmarktermination` 并关闭辅助标记的用户程序；
    2. 清理处理器上的线程缓存；
4. 清理阶段
    1. 将状态切换至 `_GCoff` 开始清理阶段，初始化清理状态并关闭写屏障；
    2. 恢复用户程序，所有新创建的对象会标记成白色；
    3. 后台并发清理所有的内存管理单元，当 Goroutine 申请新的内存管理单元时就会触发清理

`runtime.work` 变量：该结构体中包含大量垃圾收集的相关字段，例如：表示完成的垃圾收集循环的次数、当前循环时间和 CPU 的利用率、垃圾收集的模式等等，我们会在后面的小节中见到该结构体中的更多字段。

### GC 触发时机

Go 语言中对 GC 的触发时机存在两种形式：

1. **主动触发**，通过调用 `runtime.GC` 来触发 GC，此调用阻塞式地等待当前 GC 运行完毕。

2. **被动触发**，分为两种方式：

    - 使用系统监控，如果一定时间内没有触发，就会触发新的循环，该触发条件由 runtime.forcegcperiod 变量控制，默认为 2 分钟；
    - 使用步调（Pacing）算法，其核心思想是控制内存增长的比例。

3. **申请内存**

    最后一个可能会触发垃圾收集的就是 `runtime.mallocgc` 了，我们在上一节内存分配器中曾经介绍过运行时会将堆上的对象按大小分成微对象、小对象和大对象三类，这三类对象的创建都可能会触发新的垃圾收集循环：

    1. 当前线程的内存管理单元中不存在空闲空间时，创建微对象和小对象需要调用 `runtime.mcache.nextFree` 从中心缓存或者页堆中获取新的管理单元，在这时就可能触发垃圾收集；
    2. 当用户程序申请分配 32KB 以上的大对象时，一定会构建 `runtime.gcTrigger` 结构体尝试触发垃圾收集；

    通过堆内存触发垃圾收集需要比较 `runtime.mstats` 中的两个字段 — 表示垃圾收集中存活对象字节数的 `heap_live` 和表示触发标记的堆内存大小的 `gc_trigger`；当内存中存活的对象字节数大于触发垃圾收集的堆大小时，新一轮的垃圾收集就会开始。在这里，我们将分别介绍这两个值的计算过程：

    - `heap_live` — 为了减少锁竞争，运行时只会在中心缓存分配或者释放内存管理单元以及在堆上分配大对象时才会更新；
    - `gc_trigger` — 在标记终止阶段调用 `runtime.gcSetTriggerRatio` 更新触发下一次垃圾收集的堆大小；

### GC 如何调优

- 通过 `go tool pprof` 和 `go tool trace` 等工具
- 控制内存分配的速度，限制 goroutine 的数量，从而提高赋值器对 CPU 的利用率。
- 减少并复用内存，例如使用 `sync.Pool` 来复用需要频繁创建临时对象，例如提前分配足够的内存来降低多余的拷贝。
- 需要时，增大 `GOGC` 的值，降低 GC 的运行频率。
    **GOGC 参数**
    `GOGC` 默认值是100，举个例子：你程序的上一次 GC 完，驻留内存是100MB，由于你 GOGC 设置的是100，所以下次你的内存达到 200MB 的时候就会触发一次 GC，如果你 GOGC 设置的是200，那么下次你的内存达到300MB的时候就会触发 GC。

### GC 跟不上分配的速度

目前的 Go 实现中，当 GC 触发后，会首先进入并发标记的阶段。并发标记会设置一个标志，并在 mallocgc 调用时进行检查。当存在新的内存分配时，会暂停分配内存过快的那些 goroutine，并将其转去执行一些辅助标记（Mark Assist）的工作，从而达到放缓继续分配、辅助 GC 的标记工作的目的。

### STW 的实现原理

1. 根据 `gomaxprocs` 的值来设置 `stopwait`，实际上就是P的个数。
2. 把 `gcwaiting` 置为1，并通过 preemptall 去抢占所有运行中的P。
**preemptall会遍历allp这个切片，调用 `preemptone` 逐个抢占处于 _Prunning 状态的P**。
接下来把当前M持有的P置为 `_Pgcstop` 状态，并把stopwait减去1，表示当前P已经被抢占了。
3. **遍历allp**，把所有处于 `_Psyscall` 状态的P置为 `_Pgcstop` 状态，并把 `stopwait` 减去对应的数量。
4. 再循环通过 `pidleget` 取得所有空闲的P，都置为 `_Pgcstop` 状态，从stopwait减去相应的数量。
5. 最后通过判断 `stopwait` 是否大于0，也就是是否还有没被抢占的P，来确定是否需要等待。如果需要等待，就以100微秒为超时时间，在sched.stopnote上等待，超时后再次通过preemptall抢占所有P。
因为preemptall不能保证一次就成功，所以需要循环。最后一个响应gcwaiting的工作线程在自我挂起之前，会通过stopnote唤醒当前线程，STW也就完成了。

而所谓的**抢占**，就是把 g 的 `preempt` 字段设置成true，并把 `stackguard0` `这个栈增长检测的下界设置成stackPreempt`


**1.13 通过栈增长检测代码实现goroutine抢占的原理。**
执行fmt.Println的goroutine需要执行GC，进而发起了 `STW`。而main函数中的空 `for` 循环因为没有调用任何函数，所以没有机会执行栈增长检测代码，也就不能被抢占。

1.14 实现了真正的抢占。
实际的抢占操作是由 `preemptM` 函数完成的。
`preemptM` 的主要逻辑，就是通过 `runtime.signalM` 函数向指定 M 发送 `sigPreempt` 信号。至于 `signalM` 函数，就是调用操作系统的信号相关系统调用，将指定信号发送给目标线程。至此，异步抢占工作的前一半就算完成了，信号已经发出去了。

以下几个方面来保证在当前位置进行异步抢占是安全的：

1. 可以挂起g并安全的扫描它的栈和寄存器，没有潜在的隐藏指针，而且当前并没有打断一个写屏障；
2. g还有足够的栈空间来注入一个对asyncPreempt的调用；
3. 可以安全的和runtime进行交互，例如未持有runtime相关的锁，因此在尝试获得锁时不会造成死锁。

## 3. GMP&调度器

### 从代码执行入口到 main.main 发生了什么？

程序入口是汇编实现的，主要任务是程序初始化，根据源码中的注释，程序启动的主要步骤如下：

1. 调用 `osinit()` 获取CPU核数与内存页大小；
2. 执行 `schedinit()` 初始化调度器，创建指定个数的 P，并建立 m0 与 P 的关联；
3. 以 `runtime.main` 为执行入口创建goroutine，也就是 main goroutine；
4. mstart 开启调度循环，此时等待队列里只有main goroutine等待执行；
5. main goroutine得到调度，开始执行 runtime.main；
6. runtime.main 会调用 main.main，开始执行我们编写的内容。main.main 返回后，会调用 exit 函数结束进程。

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

M 执行过程中，随时会发生上下文切换。当发生上线文切换时，需要对执行现场进行保护，以便下次被调度执行时进行现场恢复。**Go调度器M的栈保存在G对象上，只需要将M所需要的寄存器(SP、PC等)保存到G对象上就可以实现现场保护**。当这些寄存器数据被保护起来，就随时可以做上下文切换了，在中断之前把现场保存起来。如果此时G任务还没有执行完，M可以将任务重新丢到P的任务队列，等待下一次被调度执行。当再次被调度执行时，M通过访问G的vdsoSP、vdsoPC寄存器进行现场恢复(从上次中断位置继续执行)。

### M0

M0是启动程序后的编号为0的主线程，这个M对应的实例会在全局变量rutime.m0中，不需要在heap上分配，M0负责执行初始化操作和启动第一个G，在之后M0就和其他的M一样了

### G0

G0是每次启动一个M都会第一个创建的goroutine，G0仅用于负责调度G，G0不指向任何可执行的函数，每个M都会有一个自己的G0，在调度或系统调用时会使用G0的栈空间，全局变量的G0是M0的G0

**普通的 Goroutine 栈是在 Heap 分配的可增长的 stack,而 g0 的 stack 是 M 对应的线程栈**。

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

### 抢占式调度

现代操作系统调度线程都是抢占式的，我们不能依赖用户代码主动让出CPU，或者因为IO、锁等待而让出，这样会造成调度的不公平。
**基于经典的时间片算法，当线程的时间片用完之后，会被时钟中断给打断，调度器会将当前线程的执行上下文进行保存，然后恢复下一个线程的上下文，分配新的时间片令其开始执**行。这种抢占对于线程本身是无感知的，系统底层支持，不需要开发人员特殊处理。

基于时间片的抢占式调度有个明显的优点，**能够避免CPU资源持续被少数线程占用，从而使其他线程长时间处于饥饿状态**。

### 线程管理

Go 语言的运行时会通过调度器改变线程的所有权，它也提供了 `runtime.LockOSThread` 和 `runtime.UnlockOSThread` 让我们有能力绑定 Goroutine 和线程完成一些比较特殊的操作。

- Go 语言的运行时会通过 `runtime.startm` 启动线程来执行处理器 `P`，如果我们在该函数中没能从闲置列表中获取到线程 `M` 就会调用 `runtime.newm` 创建新的线程;
- 创建新的线程需要使用如下所示的 `runtime.newosproc`，该函数在 Linux 平台上会通过系统调用 clone 创建新的操作系统线程，它也是创建线程链路上距离操作系统最近的 Go 语言函数
- 使用系统调用 `clone` 创建的线程会在线程主动调用 `exit`、或者传入的函数 `runtime.mstart` 返回会主动退出，`runtime.mstart` 会执行调用 `runtime.newm` 时传入的匿名函数 `fn`，到这里也就完成了从线程创建到销毁的整个闭环。

### 调度概览

- 为了保证调度的公平性，**每个工作线程每进行 61 次调度就需要优先从全局运行队列中获取 goroutine 出来运行**
    因为如果只调度本地运行队列中的 goroutine，则全局运行队列中的 goroutine 有可能得不到运行
    **所有工作线程都能访问全局队列，所以需要加锁获取 goroutine**
- 如果从全局队列没有获取到 goroutine，从与 `m` 关联的 `p` 的本地运行队列中获取 goroutine
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

```go
go func() {
    for{}
}()
fmt.Println("Dropping mic")
// Yield execution to force executing other goroutines
runtime.Gosched()
runtime.GC()
fmt.Println("Done")
```

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

这个输出结果决定来⾃于调度器优先调度哪个 `G`。从 `runtime` 的源码可以看到，**当创建⼀个 `G` 时，会优先放⼊到下⼀个调度的 `runnext` 字段上作为下⼀次优先调度的 `G`。因此，最先输出的是最后创建的 `G`**，也就是9。然后就是按照顺序插入的 g 队列了，最后的结果是 `9 10 10  ... 0 1 2 .. 8`

## 继承与多态

```go
type People interface {
    Speak(string) string
}

type Stduent struct{}

func (stu *Stduent) Speak(think string) (talk string) {
    if think == "love" {
        talk = "You are a good boy"
    } else {
        talk = "hi"
    }
    return
}

func main() {
    var peo People = Stduent{}
    think := "love"
    fmt.Println(peo.Speak(think))
}
```

在 golang 中对多态的特点体现从语法上并不是很明显。

我们知道发生多态的几个要素：

1. 有 `interface` 接口，并且有接口定义的方法。
2. 有子类去重写 `interface` 的接口。
3. 有父类指针指向子类的具体对象

那么，满足上述 3 个条件，就可以产生多态效果，就是，父类指针可以调用子类的具体方法。

所以上述代码报错的地方在 `var peo People = Stduent{}` 这条语句， `Student{}` 已经重写了父类 `People{}` 中的 `Speak(string) string` 方法，那么只需要用父类指针指向子类对象即可。（Go 中不叫父类，这里是为了好理解）

所以应该改成 `var peo People = &Student{}` 即可编译通过。（People 为 `interface` 类型，就是指针类型）