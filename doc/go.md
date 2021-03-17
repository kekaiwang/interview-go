# Go基础

## 2.0 编译原理

### 2.1.2 编译原理

Go 的编译器在逻辑上可以被分成四个阶段：词法与语法分析、类型检查和 AST 转换、通用 SSA 生成和最后的机器代码生成。

#### 静态单赋值

静态单赋值（Static Single Assignment、SSA）是中间代码的特性，如果中间代码具有静态单赋值的特性，那么每个变量就只会被赋值一次。在实践中，我们通常会用下标实现静态单赋值，这里以下面的代码举个例子：

```go
x := 1
x := 2
y := x
```

经过简单的分析，我们就能够发现上述的代码第一行的赋值语句 x := 1 不会起到任何作用。下面是具有 SSA 特性的中间代码，我们可以清晰地发现变量 y_1 和 x_1 是没有任何关系的，所以在机器码生成时就可以省去 x := 1 的赋值，通过减少需要执行的指令优化这段代码。

```go
x_1 := 1
x_2 := 2
y_1 := x_2
```

因为 SSA 的主要作用是对代码进行优化，所以它是编译器后端的一部分；当然代码编译领域除了 SSA 还有很多中间代码的优化方法，编译器生成代码的优化也是一个古老并且复杂的领域，这里就不会展开介绍了。

## 2.2 词法和语法分析

### 2.2.2 语法分析

顶层声明有五大类型，分别是常量、类型、变量、函数和方法

## 2.3 类型检查

# 3.0 数据结构

## 3.1 数组

### 3.1.1 概述

数组是由相同类型元素的集合组成的数据结构，计算机会为数组分配一块连续的内存来保存其中的元素，我们可以利用数组中元素的索引快速访问特定元素，常见的数组大多都是一维的线性数组，而多维数组在数值和图形计算领域却有比较常见的应用。

![image](https://mail.wangkekai.cn/1615900454489.jpg)

数组作为一种基础的数据类型，我们通常会从两个纬度去描述数组，也就是数组中存储的元素类型和数组最大能存储的元素个数

```go
[10]int
[100]interface{}
```

<font color=red>Go 语言数组在初始化之后大小就无法改变，</font>存储元素类型相同、但是大小不同的数组类型在 Go 语言看来也是完全不同的，只有两个条件都相同才是同一类型。

```go
func NewArray(elem *Type, bound int64) *Type {
    if bound < 0 {
        Fatalf("NewArray: invalid bound %v", bound)
    }
    t := New(TARRAY)
    t.Extra = &Array{Elem: elem, Bound: bound}
    t.SetNotInHeap(elem.NotInHeap())
return t
}
```

编译期间的数组类型是由上述的 cmd/compile/internal/types.NewArray 函数生成的，该类型包含两个字段，**分别是元素类型 Elem 和数组的大小 Bound，这两个字段共同构成了数组类型，**而当前数组是否应该在堆栈中初始化也在编译期就确定了。

### 3.1.2 初始化

Go 语言的数组有两种不同的创建方式，**一种是显式的指定数组大小，另一种是使用 [...]T 声明数组**，Go 语言会在编译期间通过源代码推导数组的大小：

```go
arr1 := [3]int{1, 2, 3}
arr2 := [...]int{1, 2, 3}
```

上述两种声明方式在运行期间得到的结果是完全相同的，后一种声明方式在编译期间就会被转换成前一种，这也就是编译器对数组大小的推导，下面我们来介绍编译器的推导过程。

#### 上限推导

两种不同的声明方式会导致编译器做出完全不同的处理，如果我们使用第一种方式 [10]T，那么变量的类型在编译进行到类型检查阶段就会被提取出来，随后使用 cmd/compile/internal/types.NewArray 创建包含数组大小的 cmd/compile/internal/types.Array 结构体。

当我们使用 [...]T 的方式声明数组时，编译器会在的 cmd/compile/internal/gc.typecheckcomplit 函数中对该数组的大小进行推导：

```go
func typecheckcomplit(n *Node) (res *Node) {
    ...
    if n.Right.Op == OTARRAY && n.Right.Left != nil && n.Right.Left.Op == ODDD {
        n.Right.Right = typecheck(n.Right.Right, ctxType)
        if n.Right.Right.Type == nil {
            n.Type = nil
            return n
        }
        elemType := n.Right.Right.Type

        length := typecheckarraylit(elemType, -1, n.List.Slice(), "array literal")

        n.Op = OARRAYLIT
        n.Type = types.NewArray(elemType, length)
        n.Right = nil
        return n
    }
    ...

    switch t.Etype {
    case TARRAY:
        typecheckarraylit(t.Elem(), t.NumElem(), n.List.Slice(), "array literal")
        n.Op = OARRAYLIT
        n.Right = nil
    }
}
```

#### 语句转换

对于一个由字面量组成的数组，根据数组元素数量的不同，编译器会在负责初始化字面量的 cmd/compile/internal/gc.anylit 函数中做两种不同的优化：

1. 当元素数量小于或者等于 4 个时，会直接将数组中的元素放置在栈上；
2. 当元素数量大于 4 个时，会将数组中的元素放置到静态区并在运行时取出；

```go
func anylit(n *Node, var_ *Node, init *Nodes) {
    t := n.Type
    switch n.Op {
    case OSTRUCTLIT, OARRAYLIT:
        if n.List.Len() > 4 {
            ...
        }

        fixedlit(inInitFunction, initKindLocalCode, n, var_, init)
    ...
    }
}
```

当数组的元素小于或者等于四个时，cmd/compile/internal/gc.fixedlit 会负责在函数编译之前将 [3]{1, 2, 3} 转换成更加原始的语句：

```go
func fixedlit(ctxt initContext, kind initKind, n *Node, var_ *Node, init *Nodes) {
    var splitnode func(*Node) (a *Node, value *Node)
    ...

    for _, r := range n.List.Slice() {
        a, value := splitnode(r)
        a = nod(OAS, a, value)
        a = typecheck(a, ctxStmt)
        switch kind {
        case initKindStatic:
            genAsStatic(a)
        case initKindLocalCode:
            a = orderStmtInPlace(a, map[string][]*Node{})
            a = walkstmt(a)
            init.Append(a)
        }
    }
}
```

当数组中元素的个数小于或者等于四个并且 cmd/compile/internal/gc.fixedlit 函数接收的 kind 是 `initKindLocalCode` 时，上述代码会将原有的初始化语句 [3]int{1, 2, 3} 拆分成一个声明变量的表达式和几个赋值表达式，这些表达式会完成对数组的初始化：

```go
var arr [3]int
arr[0] = 1
arr[1] = 2
arr[2] = 3
```

但是如果当前数组的元素大于四个，cmd/compile/internal/gc.anylit 会先获取一个唯一的 staticname，然后调用 cmd/compile/internal/gc.fixedlit 函数在静态存储区初始化数组中的元素并将临时变量赋值给数组：

```go
func anylit(n *Node, var_ *Node, init *Nodes) {
    t := n.Type
    switch n.Op {
    case OSTRUCTLIT, OARRAYLIT:
        if n.List.Len() > 4 {
            vstat := staticname(t)
            vstat.Name.SetReadonly(true)

            fixedlit(inNonInitFunction, initKindStatic, n, vstat, init)

            a := nod(OAS, var_, vstat)
            a = typecheck(a, ctxStmt)
            a = walkexpr(a, init)
            init.Append(a)
            break
        }

        ...
    }
}
```

假设代码需要初始化 [5]int{1, 2, 3, 4, 5}，那么我们可以将上述过程理解成以下的伪代码：

```go
var arr [5]int
statictmp_0[0] = 1
statictmp_0[1] = 2
statictmp_0[2] = 3
statictmp_0[3] = 4
statictmp_0[4] = 5
arr = statictmp_0
```

总结起来，在不考虑逃逸分析的情况下，如果数组中元素的个数小于或者等于 4 个，那么所有的变量会直接在栈上初始化，如果数组元素大于 4 个，变量就会在静态存储区初始化然后拷贝到栈上，这些转换后的代码才会继续进入中间代码生成和机器码生成两个阶段，最后生成可以执行的二进制文件。

#### 访问和赋值

无论是在栈上还是静态存储区，数组在内存中都是一连串的内存空间，我们通过指向数组开头的指针、元素的数量以及元素类型占的空间大小表示数组。如果我们不知道数组中元素的数量，访问时可能发生越界；而如果不知道数组中元素类型的大小，就没有办法知道应该一次取出多少字节的数据，无论丢失了那个信息，我们都无法知道这片连续的内存空间到底存储了什么数据:
![image](https://mail.wangkekai.cn/1615904201861.jpg)

**数组访问越界是非常严重的错误，Go 语言中可以在编译期间的静态类型检查判断数组越界**，cmd/compile/internal/gc.typecheck1 会验证访问数组的索引：

```go
func typecheck1(n *Node, top int) (res *Node) {
    switch n.Op {
    case OINDEX:
        ok |= ctxExpr
        l := n.Left  // array
        r := n.Right // index
        switch n.Left.Type.Etype {
        case TSTRING, TARRAY, TSLICE:
            ...
            if n.Right.Type != nil && !n.Right.Type.IsInteger() {
                yyerror("non-integer array index %v", n.Right)
                break
            }
            if !n.Bounded() && Isconst(n.Right, CTINT) {
                x := n.Right.Int64()
                if x < 0 {
                    yyerror("invalid array index %v (index must be non-negative)", n.Right)
                } else if n.Left.Type.IsArray() && x >= n.Left.Type.NumElem() {
                    yyerror("invalid array index %v (out of bounds for %d-element array)", n.Right, n.Left.Type.NumElem())
                }
            }
        }
    ...
    }
}
```

1. 访问数组的索引是非整数时，报错 “non-integer array index %v”；
2. 访问数组的索引是负数时，报错 “invalid array index %v (index must be non-negative)"；
3. 访问数组的索引越界时，报错 “invalid array index %v (out of bounds for %d-element array)"；

数组和字符串的一些简单越界错误都会在编译期间发现，例如：直接使用整数或者常量访问数组；但是如果使用变量去访问数组或者字符串时，编译器就无法提前发现错误，我们需要 Go 语言运行时阻止不合法的访问:

```shell
arr[4]: invalid array index 4 (out of bounds for 3-element array)
arr[i]: panic: runtime error: index out of range [4] with length 3
```

Go 语言运行时在发现数组、切片和字符串的越界操作会由运行时的 `runtime.panicIndex` 和 `runtime.goPanicIndex` 触发程序的运行时错误并导致崩溃退出：

```go
TEXT runtime·panicIndex(SB),NOSPLIT,$0-8
    MOVL	AX, x+0(FP)
    MOVL	CX, y+4(FP)
    JMP	runtime·goPanicIndex(SB)

func goPanicIndex(x int, y int) {
    panicCheck1(getcallerpc(), "index out of range")
    panic(boundsError{x: int64(x), signed: true, y: y, code: boundsIndex})
}
```

当数组的访问操作 `OINDEX` 成功通过编译器的检查后，会被转换成几个 SSA 指令，假设我们有如下所示的 Go 语言代码，通过如下的方式进行编译会得到 ssa.html 文件：

```go
package check

func outOfRange() int {
    arr := [3]int{1, 2, 3}
    i := 4
    elem := arr[i]
    return elem
}

$ GOSSAFUNC=outOfRange go build array.go
dumped SSA to ./ssa.html
```

`start` 阶段生成的 SSA 代码就是优化之前的第一版中间代码，下面展示的部分是 `elem := arr[i]` 对应的中间代码，在这段中间代码中我们发现 Go 语言为数组的访问操作生成了判断数组上限的指令 `IsInBounds` 以及当条件不满足时触发程序崩溃的 PanicBounds 指令：

```go
b1:
    ...
    v22 (6) = LocalAddr <*[3]int> {arr} v2 v20
    v23 (6) = IsInBounds <bool> v21 v11
If v23 → b2 b3 (likely) (6)

b2: ← b1-
    v26 (6) = PtrIndex <*int> v22 v21
    v27 (6) = Copy <mem> v20
    v28 (6) = Load <int> v26 v27 (elem[int])
    ...
Ret v30 (+7)

b3: ← b1-
    v24 (6) = Copy <mem> v20
    v25 (6) = PanicBounds <mem> [0] v21 v11 v24
Exit v25 (6)
```

编译器会将 `PanicBounds` 指令转换成上面提到的 `runtime.panicIndex` 函数，当数组下标没有越界时，编译器会先获取数组的内存地址和访问的下标、利用 PtrIndex 计算出目标元素的地址，最后使用 Load 操作将指针中的元素加载到内存中。

当然只有当编译器无法对数组下标是否越界无法做出判断时才会加入 PanicBounds 指令交给运行时进行判断，在使用字面量整数访问数组下标时会生成非常简单的中间代码，当我们将上述代码中的 arr[i] 改成 arr[2] 时，就会得到如下所示的代码：

```go
b1:
    ...
    v21 (5) = LocalAddr <*[3]int> {arr} v2 v20
    v22 (5) = PtrIndex <*int> v21 v14
    v23 (5) = Load <int> v22 v20 (elem[int])
    ...
```

数组的赋值和更新操作 a[i] = 2 也会生成 SSA 生成期间计算出数组当前元素的内存地址，然后修改当前内存地址的内容，这些赋值语句会被转换成如下所示的 SSA 代码：

```go
b1:
    ...
    v21 (5) = LocalAddr <*[3]int> {arr} v2 v19
    v22 (5) = PtrIndex <*int> v21 v13
    v23 (5) = Store <mem> {int} v22 v20 v19
    ...
```

赋值的过程中会先确定目标数组的地址，再通过 PtrIndex 获取目标元素的地址，最后使用 Store 指令将数据存入地址中，从上面的这些 SSA 代码中我们可以看出 上述数组寻址和赋值都是在编译阶段完成的，没有运行时的参与。

## 3.2 切片

切片即动态数组，其长度并不固定，我们可以向切片中追加元素，它会在容量不足时自动扩容。

```go
[]int
[]interface{}
```

从切片的定义我们能推测出，切片在编译期间的生成的类型只会包含切片中的元素类型，即 int 或者 interface{} 等。`cmd/compile/internal/types.NewSlice` 就是编译期间用于创建切片类型的函数：

```go
func NewSlice(elem *Type) *Type {
    if t := elem.Cache.slice; t != nil {
        if t.Elem() != elem {
            Fatalf("elem mismatch")
        }
        return t
    }

    t := New(TSLICE)
    t.Extra = Slice{Elem: elem}
    elem.Cache.slice = t
    return t
}
```

上述方法返回结构体中的 Extra 字段是一个只包含切片内元素类型的结构，也就是说切片内元素的类型都是在编译期间确定的，编译器确定了类型之后，会将类型存储在 Extra 字段中帮助程序在运行时动态获取。

### 3.2.1 数据结构

编译期间的切片是 `cmd/compile/internal/types.Slice` 类型的，但是在运行时切片可以由如下的 reflect.SliceHeader 结构体表示，其中:

- Data 是指向数组的指针;
- Len 是当前切片的长度；
- Cap 是当前切片的容量，即 Data 数组的大小：

```go
type SliceHeader struct {
    Data uintptr
    Len  int
    Cap  int
}
```

Data 是一片连续的内存空间，这片内存空间可以用于存储切片中的全部元素，数组中的元素只是逻辑上的概念，底层存储其实都是连续的，所以我们可以将切片理解成一片连续的内存空间加上长度与容量的标识。

![image](https://mail.wangkekai.cn/EBE589F0-689B-410D-A5C2-8C06A5144096.png)

**切片引入了一个抽象层，提供了对数组中部分连续片段的引用，而作为数组的引用，我们在运行区间可以修改它的长度和范围。**当切片底层的数组长度不足时就会触发扩容，切片指向的数组可能会发生变化，不过在上层看来切片是没有变化的，上层只需要与切片打交道不需要关心数组的变化。

介绍过编译器在编译期间简化了获取数组大小、读写数组中的元素等操作：因为数组的内存固定且连续，多数操作都会直接读写内存的特定位置。但是切片是运行时才会确定内容的结构，所有操作还需要依赖 Go 语言的运行时，下面的内容会结合运行时介绍切片常见操作的实现原理。

### 3.2.2 初始化

Go 语言中包含三种初始化切片的方式：

1. 通过下标的方式获得数组或者切片的一部分；
2. 使用字面量初始化新的切片；
3. 使用关键字 make 创建切片：

#### 使用下标

使用下标创建切片是最原始也最接近汇编语言的方式，它是所有方法中最为底层的一种，编译器会将 `arr[0:3]` 或者 `slice[0:3]` 等语句转换成 OpSliceMake 操作，我们可以通过下面的代码来验证一下：



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
