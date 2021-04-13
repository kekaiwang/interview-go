# 5.0 常用关键字

## 5.1 for 和 range

### 5.1.1 现象

#### 循环永动机

#### 神奇的指针

#### 遍历清空数组

#### 随机遍历

### 5.1.2 经典循环

### 5.1.3 范围循环

与简单的经典循环相比，范围循环在 Go 语言中更常见、实现也更复杂。这种循环同时使用 for 和 range 两个关键字，**编译器会在编译期间将所有 for-range 循环变成经典循环**。

#### 数组和切片

对于数组和切片来说，Go 语言有三种不同的遍历方式，这三种不同的遍历方式分别对应着代码中的不同条件，它们会在 `cmd/compile/internal/gc.walkrange` 函数中转换成不同的控制逻辑，我们会分成几种情况分析该函数的逻辑：

1. 分析遍历数组和切片清空元素的情况；
2. 分析使用 for range a {} 遍历数组和切片，不关心索引和数据的情况；
3. 分析使用 for i := range a {} 遍历数组和切片，只关心索引的情况；
4. 分析使用 for i, elem := range a {} 遍历数组和切片，关心索引和数据的情况

对于所有的 range 循环，Go 语言都会在编译期将原切片或者数组赋值给一个新变量 ha，在赋值的过程中就发生了拷贝，而我们又通过 len 关键字预先获取了切片的长度，所以在循环中追加新的元素也不会改变循环执行的次数，这也就解释了循环永动机一节提到的现象。

**而遇到这种同时遍历索引和元素的 range 循环时，Go 语言会额外创建一个新的 v2 变量存储切片中的元素，循环中使用的这个变量 v2 会在每一次迭代被重新赋值而覆盖，赋值时也会触发拷贝**。

#### 哈希表

## 5.3 defer

### 5.3.1 现象

- `defer` 关键字的调用时机以及多次调用 `defer` 时执行顺序是如何确定的；
- `defer` 关键字使用传值的方式传递参数时会进行预计算，导致不符合预期的结果；

#### 作用域

```go
func main() {
    {
        defer fmt.Println("defer runs")
        fmt.Println("block ends")
    }
    
    fmt.Println("main ends")
}

$ go run main.go
block ends
main ends
defer runs
```

**`defer` 传入的函数不是在退出代码块的作用域时执行的，它只会在当前函数和方法返回之前被调用**。

#### 预计算参数

> Go 语言中所有的函数调用都是传值的，虽然 defer 是关键字，但是也继承了这个特性。

假设我们想计算 main 函数的运行时间

```golang
func main() {
    startedAt := time.Now()

    defer fmt.Println(time.Since(startedAt))

    time.Sleep(time.Second)
}

$ go run main.go
0s
```

<u>我们会发现调用 `defer` 关键字会立刻拷贝函数中引用的外部参数，所以 time.Since(startedAt) 的结果不是在 main 函数退出之前计算的，而是在 defer 关键字调用时计算的，最终导致上述代码输出 0s</u>。

要想解决这个问题，我们需要向 defer 传入匿名函数：

```go
func main() {
    startedAt := time.Now()
    defer func() { fmt.Println(time.Since(startedAt)) }()

    time.Sleep(time.Second)
}

$ go run main.go
1s
```

**虽然调用 defer 关键字时也使用值传递，但是因为拷贝的是函数指针**，所以 time.Since(startedAt) 会在 main 函数返回前调用并打印出符合预期的结果。

### 5.3.2 数据结构

```go
type _defer struct {
    siz       int32
    started   bool
    openDefer bool
    sp        uintptr
    pc        uintptr
    fn        *funcval
    _panic    *_panic
    link      *_defer
}
```

`runtime._defer` 结构体是延迟调用链表上的一个元素，所有的结构体都会通过 `link` 字段串联成链表。

## 5.4 panic 和 recover

- `panic` 能够改变程序的控制流，调用 `panic` 后会立刻停止执行当前函数的剩余代码，并在当前 `Goroutine` 中递归执行调用方的 `defer；`
- `recover` 可以中止 `panic` 造成的程序崩溃。它是一个只能在 `defer` 中发挥作用的函数，在其他作用域中调用不会发挥作用。

### 5.4.1 现象

- panic 只会触发当前 Goroutine 的 defer；
- recover 只有在 defer 中调用才会生效；
- panic 允许在 defer 中嵌套多次调用。

#### 跨协程失效

```go
func main() {
    defer println("in main")
    go func() {
        defer println("in goroutine")
        panic("")
    }()

    time.Sleep(1 * time.Second)
}

$ go run main.go
in goroutine
panic:
...
```

运行这段代码时发现 main 函数的 defer 并没有执行，执行的只是当前 goroutine 中的 defer 。

前面我们曾经介绍过 `defer` 关键字对应的 `runtime.deferproc` 会将延迟调用函数与调用方所在 `Goroutine` 进行关联。所以当程序发生崩溃时只会调用当前 `Goroutine` 的延迟调用函数也是非常合理的。

![image](https://mail.wangkekai.cn/B5A47EC2-1508-4930-B114-09C4AEF9F213.png)

如上图所示，多个 Goroutine 之间没有太多的关联，一个 Goroutine 在 panic 时也不应该执行其他 Goroutine 的延迟函数。

#### 失效的崩溃恢复
