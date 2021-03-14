# Go基础

## 2.0 编译原理

### 2.1.2 编译原理

Go 的编译器在逻辑上可以被分成四个阶段：词法与语法分析、类型检查和 AST 转换、通用 SSA 生成和最后的机器代码生成。

## 5.0 常用关键字

### 5.3 defer

Go 语言的 `defer` 会在当前函数返回前执行传入的函数，它会经常++被用于关闭文件描述符、关闭数据库连接以及解锁资源++。

#### 5.3.1 现象

我们在 Go 语言中使用 `defer` 时会遇到两个常见问题:

1. 作用域
2. 预计算参数

这里会介绍具体的场景并分析这两个现象背后的设计原理：

1. `defer` 关键字的调用时机以及多次调用 `defer` 时执行顺序是如何确定的；
2. `defer` 关键字使用传值的方式传递参数时会进行预计算，导致不符合预期的结果；

#### 作用域

假设我们在 `for` 循环中多次调用 `defer` 关键字：

```golang
func main() {
    for i := 0; i < 5; i++ {
        defer fmt.Println(i)
    }
}
// go run main.go
// 4 3 2 1 0
```

运行上述代码会倒序执行传入 `defer` 关键字的所有表达式，因为最后一次调用 `defer` 时传入了 `fmt.Println(4)`，所以这段代码会优先打印 4。我们可以通过下面这个简单例子强化对 defer 执行时机的理解：

```golang
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

从上述代码的输出我们会发现，`defer` 传入的函数不是在退出代码块的作用域时执行的，它只会在当前函数和方法返回之前被调用。

#### 预计算参数

> Go 语言中所有的函数调用都是传值的，虽然 `defer` 是关键字，但是也继承了这个特性。

```golang
func main() {
    startedAt := time.Now()
    defer fmt.Println(time.Since(startedAt))

    time.Sleep(time.Second)
}

$ go run main.go
0s
```

> 调用 `defer` 关键字会立刻拷贝函数中引用的外部参数，**所以 `time.Since(startedAt)` 的结果不是在 `main` 函数退出之前计算的，而是在 `defer` 关键字调用时计算的，最终导致上述代码输出 0s**。

向 defer 关键字传入匿名函数：

```go
func main() {
    startedAt := time.Now()
    defer func() { fmt.Println(time.Since(startedAt)) }()

    time.Sleep(time.Second)
}

$ go run main.go
1s
```

<font color=red>虽然调用 `defer` 关键字时也使用值传递，但是因为拷贝的是函数指针，所以 `time.Since(startedAt)` 会在 `main` 函数返回前调用并打印出符合预期的结果</font>。

### 5.3.2 数据结构
