# 基础知识

## iota

- 每当 `const` 出现，都会使 `iota` 初始化为 0
- `const` 中++每新增一行常量声明将使 `iota` 计数一次++.

```go
const a0 = iota // a0 = 0  // const出现, iota初始化为0
const (
    a1 = iota   // a1 = 0   // 又一个const出现, iota初始化为0
    a2 = iota   // a1 = 1   // const新增一行, iota 加1
    a3 = 6      // a3 = 6   // 自定义一个常量
    a4          // a4 = 6   // 不赋值就和上一行相同
    a5 = iota   // a5 = 4   // const已经新增了4行, 所以这里是4
)
```

## 变量声明、常量定义

- 定义变量同时显式初始化
- 不能提供数据类型
- 短声明只能在函数内部使用

```go
var (
    size     = 1024
    max_size = size * 2
)
func main() {
    size := 1024
}
```

- `const`，**常量不同于变量的在运行期分配内存，常量通常会被编译器在预处理阶段直接展开，作为指令数据使用**。`cannot take the address of cl`

```go
const cl = 100

var bl = 123

func main() {
    println(&bl, bl)
    println(&cl, cl)
}
```

- `type alias` `Go1.9` 新特性
  - 基于一个类型创建一个新类型，称之为 `defintion`。  
    `MyInt1` 为称之为  `defintion`，虽然底层类型为 `int` 类型，但是不能直接赋值，需要强转；
  - 基于一个类型创建一个别名，称之为 `alias`。  
    `MyInt2` 称之为 `alias`，可以直接赋值。
  - 结果不限于方法，字段也一样；也不限于 `type alias`，`type defintion` 也是一样的，只要有重复的方法、字段，就会有这种提示，因为不知道该选择哪个。`ambiguous selector my.m1`

```go
type T1 struct {
}
func (t T1) m1() {
    fmt.Println("T1.m1")
}
type T2 = T1
type MyStruct struct {
    T1
    T2
}
func main() {
    type MyInt1 int
    type MyInt2 = int
    var i int = 9
    var i1 MyInt1 = i // cannot use i (type int) as type MyInt1 in assignment
    var i2 MyInt2 = i
    fmt.Println(i1, i2)

    // 第三点
    my:=MyStruct{}
    my.m1()
}
```

## func

- <font color=red>闭包延迟求值</font>，`for` 循环复用局部变量 `i`，每一次放入匿名函数的应用都是相同一个变量

```go
func test() []func() {
    var funs []func()
    for i := 0; i < 2; i++ {
        // x := i -- 可以解决延迟求值的问题
        funs = append(funs, func() {
            println(&i, i) // 两次打印值相等
        })
    }
    return funs
}
```

- 闭包引用相同变量

```go
func test(x int) (func(), func()) {
    return func() {
            println(x)
            x += 10
        }, func() {
            println(x)
        }
}
func main() {
    a, b := test(100)
    a()
    b() // 100, 110
}
```

## map

### 1、for range

- 使用 `for range` 迭代 `map` 时每次迭代的顺序可能不一样，<font color=red>因为map的迭代是随机的</font>
- 与 Java 的 `foreach` 一样，都是使用副本的方式。所以 `m[stu.Name]=&stu` 实际上一致指向同一个指针， 最终该指针的值为遍历的最后一个 `struct` 的值拷贝。 如下：

```go
func pase_student() {
    m := make(map[string]*student)
    stus := []student{
        {Name: "zhou", Age: 24},
        {Name: "li", Age: 23},
        {Name: "wang", Age: 22},
    }
    for _, stu := range stus {
        m[stu.Name] = &stu
        println(stu.Name, "=>", &stu)
    }

    for k, v := range m {
        println(k, "=>", v.Name)
    }
}
```

## slice

- make  
**make初始化会有默认值**；如下：

```go
func main() {
    s := make([]int, 5)
    s = append(s, 1, 2, 3)
    fmt.Println(s) // [0 0 0 0 0 1 2 3]
}
```

- 创建 `slice`
    1. 直接声明： `var slice []int`
    1. `new: slice := *new([]int)`
    1. 字面量： `slice := []int{1,2,3,4,5}`
    1. `make: slice :=  make([]int, 5, 10)`
    1. 从切片或数组“截取”: `slice := array[1:5]` 或 `slice := sourceSlice[1:5]`

```go
func main() { // 编译不通过
    list := new([]int) // 此处需要改为 *new([]int)
    list = append(list, 1)
    fmt.Println(list)
}
```

- `append` 切片 <font color=red>...</font>

```go
// 切记 ...
func main() {
    s1 := []int{1, 2, 3}
    s2 := []int{4, 5}
    s1 = append(s1, s2...)
}
```

## struct 结构体

- 进行结构体比较时候，只有相同类型的结构体才可以比较，<font color=red>结构体是否相同不但与属性类型个数有关，还与属性顺序相关</font>。
- 结构体属性中有**不可以比较的类型**，如 `map`, `slice`, 则结构体不能用 `==` 比较。

```go
func main() {
    sn1 := struct {
        age  int
        name string
    }{age: 11, name: "qq"}
    sn2 := struct {
        age  int
        name string
    }{age: 11, name: "qq"}

    if sn1 == sn2 {
        fmt.Println("sn1== sn2")
    }
    // 以下编译不过，因为map不可以比较
    sm1 := struct {
        age int
        m   map[string]string
    }{age: 11, m: map[string]string{"a": "1"}}
    sm2 := struct {
        age int
        m   map[string]string
    }{age: 11, m: map[string]string{"a": "1"}}
    // 可以使用 reflect.DeepEqual()
    if sm1 == sm2 {
        fmt.Println("sm1== sm2")
    }
}
```

## 指针

- 可以通过 “&” 取指针的地址，如：变量 a 的地址是 &a
- 可以通过“ * ”对指针取值，`*&a`，就是 a 变量所在地址的值，当然也就是 a 的值了
- 不可以对指针进行自增或自减运算，不可以对指针进行下标运算

## import

go语言 import 的最后一个元素是 `目录名` 而不是包名。
> 只不过是习惯性包名使用目录名

## defer、panic

- `defer` 是<font color=red>后进先出</font>。  
    `panic` 需要等 `defer` 结束后才会向上传递。 出现 `panic` 恐慌时候，会先按照 `defer` 的后入先出的顺序执行，最后才会执行 `panic`。
- 执行顺序例子  
    需要注意到 `defer` 执行顺序和值传递 `index:1` 肯定是最后执行的，但是 `index:1` 的第三个参数是一个函数，所以最先被调用 `calc("10",1,2)==>10,1,2,3`。  
    执行 `index:2` 时,与之前一样，需要先调用 `calc("20",1,2)==>20,1,2,3` 执行到 `b=1` 时候开始调用，`index:2==>calc("2",0,2)==>2,0,2,2` 最后执行 `index:1==>calc("1",1,3)==>1,1,3,4`。

```go
func calc(index string, a, b int) int {
    ret := a + b
    fmt.Println(index, a, b, ret)
    return ret
}

func main() {
    a := 1
    b := 2
    defer calc("1", a, calc("10", a, b))
    defer calc("2", a, calc("20", a, b))
}
```

- 需要明确一点是 `defer` 需要在函数结束前执行。 函数返回值名字会在函数起始处被初始化为对应类型的零值并且作用域为整个函数  `DeferFunc1` 有函数返回值 `t` 作用域为整个函数，在 `return` 之前 `defer` 会被执行，所以 `t` 会被修改，返回4;  `DeferFunc2` 函数中 `t` 的作用域为函数，返回1; `DeferFunc3` 返回3
- <font color=red> `panic` 仅有最后一个可以被 `revover` 捕获</font>。
触发 `panic("panic")` 后顺序执行 `defer` ，但是 `defer` 中还有一个 `panic` ，所以覆盖了之前的 `panic("panic")`

```go
func DeferFunc1(i int) (t int) {
    t = i
    defer func() {
        t += 3
    }()
    return t
}
func DeferFunc2(i int) int {
    t := i
    defer func() {
        t += 3
        fmt.Println("---", t)
    }()
    fmt.Println("---oo", t)
    return t
}
func DeferFunc3(i int) (t int) {
    defer func() {
        t += i
    }()
    return 2
}
```

## goto

- `goto` 不能跳转到其他函数或者内层代码

```go
func main() {
// loop: -- 放这就不会报错
    for i := 0; i < 10; i++ {
    loop:
        fmt.Println(i)
    }
    goto loop
}
```

## goroutine

- go 随机性
  - 谁也不知道执行后打印的顺序是什么样的，所以只能说是随机数字。但是 A: 均为输出10，B: 从 0~9 输出(顺序不定)。 第一个 `go func` 中i是外部 for 的一个变量，地址不变化。遍历完成后，最终 i=10。 故 `go func` 执行时，i的值始终是10。
  - 第二个 `go func` 中 i 是函数参数，与外部 for 中的 i 完全是两个变量。尾部(i)将发生值拷贝，`go func` 内部指向值拷贝地址。

```go
func main() {
    runtime.GOMAXPROCS(1)
    wg := sync.WaitGroup{}
    wg.Add(20)
    for i := 0; i < 10; i++ {
        go func() {
            fmt.Println("A: ", i)
            wg.Done()
        }()
    }
    for i := 0; i < 10; i++ {
        go func(i int) {
            fmt.Println("B: ", i)
            wg.Done()
        }(i)
    }
    wg.Wait()
}
```

- 以通信的方式共享内存.

## channel

- 给一个 `nil channel` 发送数据，造成永远阻塞
- 从一个 `nil channel` 接收数据，造成永远阻塞
- 给一个已经关闭的 `channel` 发送数据，引起 `panic`
- 从一个已经关闭的 `channel` 接收数据，如果缓冲区中为空，则返回一个零值
- 无缓冲的 `channel` 是同步的，而有缓冲的 `channel` 是非同步的

> 空读写阻塞，写关闭异常，读关闭空零

### 组合继承

- 这是 `Golang` 的组合模式，可以实现 `OOP` 的继承。 被组合的类型 `People` 所包含的方法虽然升级成了外部类型 `Teacher` 这个组合类型的方法（一定要是匿名字段），但它们的方法 `(ShowA())` 调用时接受者并没有发生变化。 此时 `People` 类型并不知道自己会被什么类型组合，当然也就无法调用方法时去使用未知的组合者 `Teacher` 类型的功能。

```go
type People struct{}

func (p *People) ShowA() {
    fmt.Println("showA")
    p.ShowB()
}
func (p *People) ShowB() {
    fmt.Println("showB")
}

type Teacher struct {
    People
}
func (t *Teacher) ShowB() {
    fmt.Println("teachershowB")
}

func main() {
    t := Teacher{}
    t.ShowA()
}
// showA
// showB
```

## select

- 随机性
     `select` 会随机选择一个可用通道做收发操作。 所以代码是有可能触发异常，也有可能不会。 单个 `chan` 如果无缓冲时，将会阻塞。但结合  `select` 可以在多个 `chan` 间等待执行。有三点原则：
  - `select`  中只要有一个 `case` 能 `return` ，则立刻执行。
  - 当如果同一时间有多个 `case` 均能 `return` <font color=red>则伪随机方式抽取任意一个执行</font>。
  - 如果没有一个 `case` 能 `return` 则可以执行”default”块。

```go
func main() {
    runtime.GOMAXPROCS(1)
    int_chan := make(chan int, 1)
    string_chan := make(chan string, 1)
    int_chan <- 1
    string_chan <- "hello"
    select {
    case value := <-int_chan:
        fmt.Println(value)
    case value := <-string_chan:
        panic(value)
    }
}
```

## 方法集

- 对于 `T` 类型，它的方法集只包含接收者类型是T的方法；而对于 `*T` 类型，它的方法集则包含接收者为 `T` 和 `*T` 类型的方法，也就是全部方法。

## reflect

- `type` 只能用在 `interface` 类型上

```go
func GetValue() int {
    return 1
}
// 编译不通过
func main() {
    i := GetValue()
    switch i.(type) {
    case int:
        println("int")
    case string:
        println("string")
    case interface{}:
        println("interface")
    default:
        println("unknown")
    }
}
```

## nil

- `nil` 可以用作 `interface、function、pointer、map、slice` 和 `channel` 的“空值”。但是如果不特别指定的话，Go 语言不能识别类型，所以会报错。报: `cannot use nil as type string in return argument.`

## interface

### 在golang中对多态的特点体现从语法上并不是很明显，我们知道发生多态的几个要素

1. 有interface接口，并且有接口定义的方法。
2. 有子类去重写interface的接口。
3. 有父类指针指向子类的具体对象

### interface在使用的过程中，共有两种表现形式

- 一种为空接口(empty interface)

```go
var MyInterface interface{}
```

- 另一种为非空接口(non-empty interface)

```go
type MyInterface interface {
    function()
}
```

这两种 `interface` 类型分别用两种 `struct` 表示，空接口为 `eface`, 非空接口为 `iface`.

![image](https://mail.wangkekai.cn/96FDF748-A53A-4E19-B31C-17F8CE80BB54.png)

#### 空接口eface

空接口eface结构，由两个属性构成，一个是类型信息 `_type`，一个是数据信息。其数据结构声明如下：

```go
type eface struct {      //空接口
    _type *_type         //类型信息
    data  unsafe.Pointer //指向数据的指针(go语言中特殊的指针类型unsafe.Pointer类似于c语言中的void*)
}
```

`_type` 属性：是 GO 语言中所有类型的公共描述，Go 语言几乎所有的数据结构都可以抽象成 `_type`，是所有类型的公共描述，**type负责决定data应该如何解释和操作，** type 的结构代码如下:

```go
type _type struct {
    size       uintptr  //类型大小
    ptrdata    uintptr  //前缀持有所有指针的内存大小
    hash       uint32   //数据hash值
    tflag      tflag
    align      uint8    //对齐
    fieldalign uint8    //嵌入结构体时的对齐
    kind       uint8    //kind 有些枚举值kind等于0是无效的
    alg        *typeAlg //函数指针数组，类型实现的所有方法
    gcdata    *byte
    str       nameOff
    ptrToThis typeOff
}
```

`data` 属性: 表示指向具体的实例数据的指针，他是一个 `unsafe.Pointer` 类型，相当于一个 C 的万能指针 `void*`。

![image](https://mail.wangkekai.cn/356FF9BD-E988-4E8F-9DBF-D23FC3467DA9.png)

#### 非空接口iface

`iface` 表示 `non-empty interface` 的数据结构，非空接口初始化的过程就是初始化一个 `iface` 类型的结构，其中 `data` 的作用同 `eface` 的相同。

```go
type iface struct {
    tab  *itab
    data unsafe.Pointer
}
```

`iface` 结构中最重要的是 `itab` 结构（结构如下），每一个 `itab` 都占 32 字节的空间。`itab` 可以理解为 `pair<interface type, concrete type>` 。`itab` 里面包含了 `interface` 的一些关键信息，比如 `method` 的具体实现。

```go
type itab struct {
    inter  *interfacetype   // 接口自身的元信息
    _type  *_type           // 具体类型的元信息
    link   *itab
    bad    int32
    hash   int32            // _type里也有一个同样的hash，此处多放一个是为了方便运行接口断言
    fun    [1]uintptr       // 函数指针，指向具体类型所实现的方法
}
```

1. `interface type` 包含了一些关于 `interface` 本身的信息，比如 `package path`，包含的 `method` 。这里的 `interface type` 是定义 `interface` 的一种抽象表示。
2. `type` 表示具体化的类型，与 `eface` 的 `type` 类型相同。
3. `hash` 字段其实是对 `_type.hash` 的拷贝，它会在 `interface` 的实例化时，用于快速判断目标类型和接口中的类型是否一致。另，Go 的 `interface` 的 `Duck-typing` 机制也是依赖这个字段来实现。
4. `fun` 字段其实是一个动态大小的数组，虽然声明时是固定大小为 1，但在使用时会直接通过 fun 指针获取其中的数据，并且不会检查数组的边界，所以该数组中保存的元素数量是不确定的。

![image](https://mail.wangkekai.cn/BB52371E-F4B4-4C88-B1E7-C7E7151E4346.png)

## sizeof

```golang
slice := []int{1,2,3}
fmt.Println(unsafe.Sizeof(slice)) //24
```

- 如果 x 为一个切片，`sizeof` 返回的大小是切片的描述符，而不是切片所指向的内存的大小。

这里如果换成一个数组呢？而不是一个切片

```golang
arr := [...]int{1,2,3,4,5}
fmt.Println(unsafe.Sizeof(arr)) //40
arr2 := [...]int{1,2,3,4,5,6}
fmt.Println(unsafe.Sizeof(arr)) //48
```

- sizeof总是在编译期就进行求值，而不是在运行时，这意味着，sizeof的返回值可以赋值给常量

再来看一个字符串的例子

```go
str := "hello"
fmt.Println(unsafe.Sizeof(str)) //16
```

- 不论字符串的len有多大，sizeof 始终返回16。 实际上字符串类型对应一个结构体，该结构体有两个域，**第一个域是指向该字符串的指针，第二个域是字符串的长度，每个域占8个字节，但是并不包含指针指向的字符串的内容，这也就是为什么sizeof始终返回的是16**

## 并发编程

1. 通过channel通知实现并发控制
无缓冲的通道指的是通道的大小为0，也就是说，这种类型的通道在接收前没有能力保存任何值，它要求发送 `goroutine` 和接收 `goroutine` 同时准备好，才可以完成发送和接收操作。
从上面无缓冲的通道定义来看，发送 `goroutine` 和接收 `gouroutine` 必须是同步的，同时准备后，如果没有同时准备好的话，先执行的操作就会阻塞等待，直到另一个相对应的操作准备好为止。这种无缓冲的通道我们也称之为同步通道。

```go
func main() {
    ch := make(chan struct{})
    go func() {
        fmt.Println("start working")
        time.Sleep(time.Second * 1)
        ch <- struct{}{}
    }()

    <-ch

    fmt.Println("finished")
}
```

当主 `goroutine` 运行到 `<-ch` 接受 `channel` 的值的时候，如果该 `channel` 中没有数据，就会一直阻塞等待，直到有值。 这样就可以简单实现并发控制。

2. 通过sync包中的WaitGroup实现并发控制
`Goroutine` 是异步执行的，有的时候为了防止在结束main函数的时候结束掉 Goroutine，所以需要同步等待，这个时候就需要用 `WaitGroup` 了，在 `Sync` 包中，提供了 `WaitGroup`,它会等待它收集的所有 `goroutine` 任务全部完成。

在WaitGroup里主要有三个方法:

- `Add`, 可以添加或减少 `goroutine` 的数量.
- `Done`, 相当于 `Add(-1)`.
- `Wait`, 执行后会堵塞主线程，直到 `WaitGroup` 里的值减至0.

在主 `goroutine` 中 `Add(delta int)` 索要等待 `goroutine` 的数量。在每一个 `goroutine` 完成后 `Done()` 表示这一个 `goroutine` 已经完成，当所有的 `goroutine` 都完成后，在主 `goroutine` 中 `WaitGroup` 返回。

```go
func main(){
    var wg sync.WaitGroup
    var urls = []string{
        "http://www.golang.org/",
        "http://www.google.com/",
    }
    for _, url := range urls {
        wg.Add(1)
        go func(url string) {
            defer wg.Done()
            http.Get(url)
        }(url)
    }
    wg.Wait()
}
```

```go
func main(){
 wg := sync.WaitGroup{}
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(wg sync.WaitGroup, i int) {
            fmt.Printf("i:%d", i)
            wg.Done()
        }(wg, i)
    }
    wg.Wait()
    fmt.Println("exit")
}
```

运行：

```shell
i:1i:3i:2i:0i:4fatal error: all goroutines are asleep - deadlock!

goroutine 1 [semacquire]:
sync.runtime_Semacquire(0xc000094018)
        /home/keke/soft/go/src/runtime/sema.go:56 +0x39
sync.(*WaitGroup).Wait(0xc000094010)
        /home/keke/soft/go/src/sync/waitgroup.go:130 +0x64
main.main()
        /home/keke/go/Test/wait.go:17 +0xab
exit status 2
```

它提示所有的 `goroutine` 都已经睡眠了，出现了死锁。这是因为 wg 给拷贝传递到了 `goroutine` 中，导致只有 `Add` 操作，其实 `Done` 操作是在 `wg` 的副本执行的。

因此 Wait 就会死锁。

这个第一个修改方式: 将匿名函数中 `wg` 的传入类型改为 `*sync.WaitGroup`,这样就能引用到正确的 `WaitGroup` 了。

这个第二个修改方式: 将匿名函数中的 `wg` 的传入参数去掉，因为Go支持闭包类型，在匿名函数中可以直接使用外面的 `wg` 变量。

3. 在Go 1.7 以后引进的强大的Context上下文，实现并发控制.

通常,在一些简单场景下使用 `channel` 和 `WaitGroup` 已经足够了，但是当面临一些复杂多变的网络并发场景下 `channel` 和 `WaitGroup` 显得有些力不从心了。

比如一个网络请求 `Request`，每个 `Request` 都需要开启一个 `goroutine` 做一些事情，这些 goroutine 又可能会开启其他的 `goroutine`，比如数据库和RPC服务。

所以我们需要一种可以跟踪 `goroutine` 的方案，才可以达到控制他们的目的，这就是Go语言为我们提供的 `Context`，称之为上下文非常贴切，它就是 `goroutine` 的上下文。

context 包主要是用来处理多个 goroutine 之间共享数据，及多个 goroutine 的管理。

context 包的核心是 struct Context，接口声明如下：

```go
// A Context carries a deadline, cancelation signal, and request-scoped values
// across API boundaries. Its methods are safe for simultaneous use by multiple
// goroutines.
type Context interface {
    // Done returns a channel that is closed when this `Context` is canceled
    // or times out.
    // Done() 返回一个只能接受数据的channel类型，当该context关闭或者超时时间到了的时候，该channel就会有一个取消信号
    Done() <-chan struct{}

    // Err indicates why this Context was canceled, after the Done channel
    // is closed.
    // Err() 在Done() 之后，返回context 取消的原因。
    Err() error

    // Deadline returns the time when this Context will be canceled, if any.
    // Deadline() 设置该context cancel的时间点
    Deadline() (deadline time.Time, ok bool)

    // Value returns the value associated with key or nil if none.
    // Value() 方法允许 Context 对象携带request作用域的数据，该数据必须是线程安全的。
    Value(key interface{}) interface{}
}
```

`Context` 对象是线程安全的，你可以把一个 `Context` 对象传递给任意个数的 `gorotuine`，对它执行取消操作时，所有 `goroutine` 都会接收到取消信号。

一个 `Context` 不能拥有 `Cancel` 方法，同时我们也只能 `Done` `channel` 接收数据。其中的原因是一致的：接收取消信号的函数和发送信号的函数通常不是一个。

典型的场景是：父操作为子操作操作启动 `goroutine`，子操作也就不能取消父操作。
