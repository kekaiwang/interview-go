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

## 3.0 数据结构

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

Go 语言的数组有两种不同的创建方式，**一种是显式的指定数组大小，另一种是使用 `[...]T` 声明数组**，Go 语言会在编译期间通过源代码推导数组的大小：

```go
arr1 := [3]int{1, 2, 3}
arr2 := [...]int{1, 2, 3}
```

上述两种声明方式在运行期间得到的结果是完全相同的，后一种声明方式在编译期间就会被转换成前一种，这也就是编译器对数组大小的推导，下面我们来介绍编译器的推导过程。

#### 上限推导

两种不同的声明方式会导致编译器做出完全不同的处理，如果我们使用第一种方式 `[10]T`，那么变量的类型在编译进行到类型检查阶段就会被提取出来，随后使用 `cmd/compile/internal/types.NewArray` 创建包含数组大小的 `cmd/compile/internal/types.Array` 结构体。

当我们使用 `[...]T` 的方式声明数组时，编译器会在的 `cmd/compile/internal/gc.typecheckcomplit` 函数中对该数组的大小进行推导：

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

对于一个由字面量组成的数组，根据数组元素数量的不同，编译器会在负责初始化字面量的 `cmd/compile/internal/gc.anylit` 函数中做两种不同的优化：

1. 当元素数量小于或者等于 4 个时，会<font color=red>直接将数组中的元素放置在栈上</font>；
2. 当元素数量大于 4 个时，会<font color=red>将数组中的元素放置到静态区并在运行时取出</font>；

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

当数组的元素小于或者等于四个时，`cmd/compile/internal/gc.fixedlit` 会负责在函数编译之前将 `[3]{1, 2, 3}` 转换成更加原始的语句：

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

当数组中元素的个数小于或者等于四个并且 cmd/compile/internal/gc.fixedlit 函数接收的 kind 是 `initKindLocalCode` 时，上述代码会将原有的初始化语句 `[3]int{1, 2, 3}` 拆分成一个声明变量的表达式和几个赋值表达式，这些表达式会完成对数组的初始化：

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

总结起来，<font color=red>在不考虑逃逸分析的情况下，如果数组中元素的个数小于或者等于 4 个，那么所有的变量会直接在栈上初始化，</font>如果数组元素大于 4 个，变量就会在静态存储区初始化然后拷贝到栈上，这些转换后的代码才会继续进入中间代码生成和机器码生成两个阶段，最后生成可以执行的二进制文件。

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

**切片引入了一个抽象层，提供了对数组中部分连续片段的引用，而作为数组的引用，我们在运行区间可以修改它的长度和范围**。当切片底层的数组长度不足时就会触发扩容，切片指向的数组可能会发生变化，不过在上层看来切片是没有变化的，上层只需要与切片打交道不需要关心数组的变化。

介绍过编译器在编译期间简化了获取数组大小、读写数组中的元素等操作：因为数组的内存固定且连续，多数操作都会直接读写内存的特定位置。但是切片是运行时才会确定内容的结构，所有操作还需要依赖 Go 语言的运行时，下面的内容会结合运行时介绍切片常见操作的实现原理。

### 3.2.2 初始化

Go 语言中包含三种初始化切片的方式：

1. 通过下标的方式获得数组或者切片的一部分；
2. 使用字面量初始化新的切片；
3. 使用关键字 make 创建切片：

#### 使用下标

使用下标创建切片是最原始也最接近汇编语言的方式，它是所有方法中最为底层的一种，编译器会将 `arr[0:3]` 或者 `slice[0:3]` 等语句转换成 OpSliceMake 操作，我们可以通过下面的代码来验证一下：

通过 `GOSSAFUNC` 变量编译上述代码可以得到一系列 SSA 中间代码，其中 `slice := arr[0:1]` 语句在 “decompose builtin” 阶段对应的代码如下所示：

```go
v27 (+5) = SliceMake <[]int> v11 v14 v17

name &arr[*[3]int]: v11
name slice.ptr[*int]: v11
name slice.len[int]: v14
name slice.cap[int]: v17
```

`SliceMake` 操作会接受四个参数创建新的切片，**元素类型、数组指针、切片大小和容量**，这也是我们在数据结构一节中提到的切片的几个字段 ，<font color=red>需要注意的是使用下标初始化切片不会拷贝原数组或者原切片中的数据，它只会创建一个指向原数组的切片结构体，所以修改新切片的数据也会修改原切片。</font>

#### 字面量

当我们使用字面量 `[]int{1, 2, 3}` 创建新的切片时，`cmd/compile/internal/gc.slicelit` 函数会在编译期间将它展开成如下所示的代码片段：

```go
var vstat [3]int
vstat[0] = 1
vstat[1] = 2
vstat[2] = 3
var vauto *[3]int = new([3]int)
*vauto = vstat
slice := vauto[:]
```

1. 根据切片中的元素数量对底层数组的大小进行推断并创建一个数组；
2. 将这些字面量元素存储到初始化的数组中；
3. 创建一个同样指向 `[3]int` 类型的数组指针；
4. 将静态存储区的数组 `vstat` 赋值给 `vauto` 指针所在的地址；
5. 通过 `[:]` 操作获取一个底层使用 `vauto` 的切片；

> 第 5 步中的 `[:]` 就是使用下标创建切片的方法，从这一点我们也能看出 `[:]` 操作是创建切片最底层的一种方法。

#### 关键字

**如果使用字面量的方式创建切片，大部分的工作都会在编译期间完成**。<u>但是当我们使用 make 关键字创建切片时，很多工作都需要运行时的参与；</u>调用方必须向 make 函数传入切片的大小以及可选的容量，类型检查期间的 cmd/compile/internal/gc.typecheck1 函数会校验入参：

```go
func typecheck1(n *Node, top int) (res *Node) {
    switch n.Op {
    ...
    case OMAKE:
        args := n.List.Slice()

        i := 1
        switch t.Etype {
        case TSLICE:
            if i >= len(args) {
                yyerror("missing len argument to make(%v)", t)
                return n
            }

            l = args[i]
            i++
            var r *Node
            if i < len(args) {
                r = args[i]
            }
            ...
            if Isconst(l, CTINT) && r != nil && Isconst(r, CTINT) && l.Val().U.(*Mpint).Cmp(r.Val().U.(*Mpint)) > 0 {
                yyerror("len larger than cap in make(%v)", t)
                return n
            }

            n.Left = l
            n.Right = r
            n.Op = OMAKESLICE
        }
    ...
    }
}
```

上述函数不仅会检查 `len` 是否传入，还会保证传入的容量 `cap` 一定大于或者等于 `len`。除了校验参数之外，当前函数会将 `OMAKE` 节点转换成 `OMAKESLICE`，中间代码生成的 `cmd/compile/internal/gc.walkexpr` 函数会依据下面两个条件转换 `OMAKESLICE` 类型的节点：

1. 切片的大小和容量是否足够小；
2. 切片是否发生了逃逸，最终在堆上初始化

**当切片发生逃逸或者非常大时，运行时需要 runtime.makeslice 在堆上初始化切片**，如果当前的切片不会发生逃逸并且切片非常小的时候，make([]int, 3, 4) 会被直接转换成如下所示的代码：

```go
var arr [4]int
n := arr[:3]
```

上述代码会初始化数组并通过下标 `[:3]` 得到数组对应的切片，这两部分操作都会在编译阶段完成，编译器会在栈上或者静态存储区创建数组并将 `[:3]` 转换成上一节提到的 OpSliceMake 操作。

分析了主要由编译器处理的分支之后，我们回到用于创建切片的运行时函数 `runtime.makeslice`，这个函数的实现很简单：

```go
func makeslice(et *_type, len, cap int) unsafe.Pointer {
    mem, overflow := math.MulUintptr(et.size, uintptr(cap))
    if overflow || mem > maxAlloc || len < 0 || len > cap {
        mem, overflow := math.MulUintptr(et.size, uintptr(len))
        if overflow || mem > maxAlloc || len < 0 {
            panicmakeslicelen()
        }
        panicmakeslicecap()
    }

    return mallocgc(mem, et, true)
}
```

上述函数的主要工作是计算切片占用的内存空间并在堆上申请一片连续的内存，它使用如下的方式计算占用的内存：
<center>内存空间=切片中元素大小×切片容量</center>  

虽然编译期间可以检查出很多错误，但是在创建切片的过程中如果发生了以下错误会直接触发运行时错误并崩溃：

1. 存空间的大小发生了溢出；
2. 申请的内存大于最大可分配的内存；
3. 传入的长度小于 0 或者长度大于容量；

`runtime.makeslice` 在最后调用的 `runtime.mallocgc` 是用于申请内存的函数，这个函数的实现还是比较复杂，**如果遇到了比较小的对象会直接初始化在 Go 语言调度器里面的 P 结构中，而大于 32KB 的对象会在堆上初始化。**

在之前版本的 Go 语言中，数组指针、长度和容量会被合成一个 `runtime.slice` 结构，但是从 cmd/compile: move slice construction to callers of makeslice 提交之后，构建结构体 reflect.SliceHeader 的工作就都交给了 runtime.makeslice 的调用方，该函数仅会返回指向底层数组的指针，调用方会在编译期间构建切片结构体：

```go
func typecheck1(n *Node, top int) (res *Node) {
    switch n.Op {
    ...
    case OSLICEHEADER:
    switch 
        t := n.Type
        n.Left = typecheck(n.Left, ctxExpr)
        l := typecheck(n.List.First(), ctxExpr)
        c := typecheck(n.List.Second(), ctxExpr)
        l = defaultlit(l, types.Types[TINT])
        c = defaultlit(c, types.Types[TINT])

        n.List.SetFirst(l)
        n.List.SetSecond(c)
    ...
    }
}
```

`OSLICEHEADER` 操作会创建我们在上面介绍过的结构体 `reflect.SliceHeader`，其中包含数组指针、切片长度和容量，它是切片在运行时的表示：

```go
type SliceHeader struct {
    Data uintptr
    Len  int
    Cap  int
}
```

> 正是因为大多数对切片类型的操作并不需要直接操作原来的 runtime.slice 结构体，所以 reflect.SliceHeader 的引入能够减少切片初始化时的少量开销，**该改动不仅能够减少 ~0.2% 的 Go 语言包大小，还能够减少 92 个 runtime.panicIndex 的调用，占 Go 语言二进制的 ~3.5%1。**

### 3.2.3 访问元素

使用 `len` 和 `cap` 获取长度或者容量是切片最常见的操作，编译器将它们看成两种特殊操作，即 `OLEN` 和 `OCAP，cmd`/compile/internal/gc.state.expr 函数会在 SSA 生成阶段阶段将它们分别转换成 `OpSliceLen` 和 `OpSliceCap：`

```go
func (s *state) expr(n *Node) *ssa.Value {
    switch n.Op {
    case OLEN, OCAP:
        switch {
        case n.Left.Type.IsSlice():
            op := ssa.OpSliceLen
            if n.Op == OCAP {
                op = ssa.OpSliceCap
            }
            return s.newValue1(op, types.Types[TINT], s.expr(n.Left))
        ...
        }
    ...
    }
}
```

访问切片中的字段可能会触发 “decompose builtin” 阶段的优化，`len(slice)` 或者 `cap(slice)` 在一些情况下会直接替换成切片的长度或者容量，不需要在运行时获取：

```go
(SlicePtr (SliceMake ptr _ _ )) -> ptr
(SliceLen (SliceMake _ len _)) -> len
(SliceCap (SliceMake _ _ cap)) -> cap
```

**除了获取切片的长度和容量之外，访问切片中元素使用的 `OINDEX` 操作也会在中间代码生成期间转换成对地址的直接访问：**

```go
func (s *state) expr(n *Node) *ssa.Value {
    switch n.Op {
    case OINDEX:
        switch {
        case n.Left.Type.IsSlice():
            p := s.addr(n, false)
            return s.load(n.Left.Type.Elem(), p)
        ...
        }
    ...
    }
}
```

切片的操作基本都是在编译期间完成的，除了访问切片的长度、容量或者其中的元素之外，编译期间也会将包含 `range` 关键字的遍历转换成形式更简单的循环。

### 3.2.4 追加和扩容

使用 append 关键字向切片中追加元素也是常见的切片操作，中间代码生成阶段的 cmd/compile/internal/gc.state.append 方法会根据返回值是否会覆盖原变量，选择进入两种流程，**如果 append 返回的新切片不需要赋值回原有的变量**，就会进入如下的处理流程：

```go
// append(slice, 1, 2, 3)
ptr, len, cap := slice
newlen := len + 3
if newlen > cap {
    ptr, len, cap = growslice(slice, newlen)
    newlen = len + 3
}
*(ptr+len) = 1
*(ptr+len+1) = 2
*(ptr+len+2) = 3
return makeslice(ptr, newlen, cap)
```

我们会先结构切片结构体获取它的数组指针、大小和容量，如果在追加元素后切片的大小大于容量，那么就会调用 runtime.growslice 对切片进行扩容并将新的元素依次加入切片。

**如果使用 `slice = append(slice, 1, 2, 3)` 语句，那么 append 后的切片会覆盖原切片**，这时 cmd/compile/internal/gc.state.append 方法会使用另一种方式展开关键字：

```go
// slice = append(slice, 1, 2, 3)
a := &slice
ptr, len, cap := slice
newlen := len + 3
if uint(newlen) > uint(cap) {
   newptr, len, newcap = growslice(slice, newlen)
   vardef(a)
   *a.cap = newcap
   *a.ptr = newptr
}
newlen = len + 3
*a.len = newlen
*(ptr+len) = 1
*(ptr+len+1) = 2
*(ptr+len+2) = 3
```

<font color=red>是否覆盖原变量的逻辑其实差不多，最大的区别在于得到的新切片是否会赋值回原变量。</font>如果我们选择覆盖原有的变量，就不需要担心切片发生拷贝影响性能，因为 Go 语言编译器已经对这种常见的情况做出了优化。

![image](https://mail.wangkekai.cn/1615989405465.jpg)

到这里我们已经清除了 Go 语言如何在切片容量足够时向切片中追加元素，不过仍然需要研究切片容量不足时的处理流程。当切片的容量不足时，我们会调用 `runtime.growslice` 函数为切片扩容，扩容是为切片分配新的内存空间并拷贝原切片中元素的过程，我们先来看新切片的容量是如何确定的：

```go
func growslice(et *_type, old slice, cap int) slice {
    newcap := old.cap
    doublecap := newcap + newcap
    if cap > doublecap {
        newcap = cap
    } else {
        if old.len < 1024 {
            newcap = doublecap
        } else {
            for 0 < newcap && newcap < cap {
                newcap += newcap / 4
            }
            if newcap <= 0 {
                newcap = cap
            }
        }
    }
}
```

在分配内存空间之前需要先确定新的切片容量，运行时根据切片的当前容量选择不同的策略进行扩容：

1. 如果期望容量大于当前容量的两倍就会使用期望容量；
2. 如果当前切片的长度小于 1024 就会将容量翻倍；
3. 如果当前切片的长度大于 1024 就会每次增加 25% 的容量，直到新容量大于期望容量；

上述代码片段仅会确定切片的大致容量，下面还需要根据切片中的元素大小对齐内存，当数组中元素所占的字节大小为 1、8 或者 2 的倍数时，运行时会使用如下所示的代码对齐内存：

```go
    var overflow bool
    var lenmem, newlenmem, capmem uintptr
    switch {
    case et.size == 1:
        lenmem = uintptr(old.len)
        newlenmem = uintptr(cap)
        capmem = roundupsize(uintptr(newcap))
        overflow = uintptr(newcap) > maxAlloc
        newcap = int(capmem)
    case et.size == sys.PtrSize:
        lenmem = uintptr(old.len) * sys.PtrSize
        newlenmem = uintptr(cap) * sys.PtrSize
        capmem = roundupsize(uintptr(newcap) * sys.PtrSize)
        overflow = uintptr(newcap) > maxAlloc/sys.PtrSize
        newcap = int(capmem / sys.PtrSize)
    case isPowerOfTwo(et.size):
        ...
    default:
        ...
    }
```

`runtime.roundupsize` 函数会将待申请的内存向上取整，取整时会使用 `runtime.class_to_size` 数组，使用该数组中的整数可以提高内存的分配效率并减少碎片，我们会在内存分配一节详细介绍该数组的作用：

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

在默认情况下，我们会将目标容量和元素大小相乘得到占用的内存。如果计算新容量时发生了内存溢出或者请求内存超过上限，就会直接崩溃退出程序，不过这里为了减少理解的成本，将相关的代码省略了。

```go
    var overflow bool
    var newlenmem, capmem uintptr
    switch {
    ...
    default:
        lenmem = uintptr(old.len) * et.size
        newlenmem = uintptr(cap) * et.size
        capmem, _ = math.MulUintptr(et.size, uintptr(newcap))
        capmem = roundupsize(capmem)
        newcap = int(capmem / et.size)
    }
    ...
    var p unsafe.Pointer
    if et.kind&kindNoPointers != 0 {
        p = mallocgc(capmem, nil, false)
        memclrNoHeapPointers(add(p, newlenmem), capmem-newlenmem)
    } else {
        p = mallocgc(capmem, et, true)
        if writeBarrier.enabled {
            bulkBarrierPreWriteSrcOnly(uintptr(p), uintptr(old.array), lenmem)
        }
    }
    memmove(p, old.array, lenmem)
    return slice{p, old.len, newcap}
}
```

如果切片中元素不是指针类型，那么会调用 `runtime.memclrNoHeapPointers` 将超出切片当前长度的位置清空并在最后使用 `runtime.memmove` 将原数组内存中的内容拷贝到新申请的内存中。这两个方法都是用目标机器上的汇编指令实现的，这里就不展开介绍了。

`runtime.growslice` 函数最终会返回一个新的切片，其中包含了新的数组指针、大小和容量，这个返回的三元组最终会覆盖原切片。

```go
var arr []int64
arr = append(arr, 1, 2, 3, 4, 5)
```

简单总结一下扩容的过程，当我们执行上述代码时，会触发 `runtime.growslice` 函数扩容 arr 切片并传入期望的新容量 5，这时期望分配的内存大小为 40 字节；不过因为切片中的元素大小等于 sys.PtrSize，所以运行时会调用 runtime.roundupsize 向上取整内存的大小到 48 字节，所以新切片的容量为 48 / 8 = 6。

### 3.2.5 拷贝切片

切片的拷贝虽然不是常见的操作，但是却是我们学习切片实现原理必须要涉及的。当我们使用 `copy(a, b)` 的形式对切片进行拷贝时，编译期间的 cmd/compile/internal/gc.copyany 也会分两种情况进行处理拷贝操作，如果当前 copy 不是在运行时调用的，copy(a, b) 会被直接转换成下面的代码：

```go
n := len(a)
if n > len(b) {
    n = len(b)
}
if a.ptr != b.ptr {
    memmove(a.ptr, b.ptr, n*sizeof(elem(a))) 
}
```

上述代码中的 `runtime.memmove` 会负责拷贝内存。而如果拷贝是在运行时发生的，例如：`go copy(a, b)`，编译器会使用 runtime.slicecopy 替换运行期间调用的 copy，该函数的实现

```go
func slicecopy(to, fm slice, width uintptr) int {
    if fm.len == 0 || to.len == 0 {
        return 0
    }
    n := fm.len
    if to.len < n {
        n = to.len
    }
    if width == 0 {
        return n
    }
    ...

    size := uintptr(n) * width
    if size == 1 {
        *(*byte)(to.array) = *(*byte)(fm.array)
    } else {
        memmove(to.array, fm.array, size)
    }
    return n
}
```

无论是编译期间拷贝还是运行时拷贝，两种拷贝方式都会通过 runtime.memmove 将整块内存的内容拷贝到目标的内存区域中：

![image](https://mail.wangkekai.cn/1615989853481.jpg)

相比于依次拷贝元素，`runtime.memmove` 能够提供更好的性能。需要注意的是，整块拷贝内存仍然会占用非常多的资源，在大切片上执行拷贝操作时一定要注意对性能的影响。

> 切片的很多功能都是由运行时实现的，无论是初始化切片，还是对切片进行追加或扩容都需要运行时的支持，需要注意的是在遇到大切片扩容或者复制时可能会发生大规模的内存拷贝，一定要减少类似操作避免影响程序的性能。

## 3.3 哈希表

数组和哈希是两种设计集合元素的思路，**数组用于表示元素的序列，而哈希表示的是键值对之间映射关系。**

### 3.3.1 设计原理

哈希表是计算机科学中重要的数据结构，不仅是因为它 O(1) 的读写性能，还因为它提供了键值之间的映射。想要实现一个性能优异的哈希表，需要注意两个关键点 -- 哈希函数和哈希冲突解决办法。

#### 哈希函数

实现哈希表的关键点在于哈希函数的选择，哈希函数的选择在很大程度上能够决定哈希表的读写性能。在理想情况下，哈希函数应该能够将不同键映射到不同的索引上，这要求哈希函数的输出范围大于输入范围，但是由于键的数量会远远大于映射的范围，所以在实际使用时，这个理想的效果是不可能实现的。

比较实际的方式是让哈希函数的结果能够尽可能的均匀分布，然后通过工程上的手段解决哈希碰撞的问题。**哈希函数映射的结果一定要尽可能均匀，结果不均匀的哈希函数会带来更多的哈希冲突以及更差的读写性能**。

![image](https://mail.wangkekai.cn/1616077557775.jpg)

如果使用结果分布较为均匀的哈希函数，那么哈希的增删改查的时间复杂度为 𝑂(1)；但是如果哈希函数的结果分布不均匀，那么所有操作的时间复杂度可能会达到 𝑂(𝑛)，由此看来，使用好的哈希函数是至关重要的。

#### 冲突解决

就像我们之前所提到的，在通常情况下，哈希函数输入的范围一定会远远大于输出的范围，所以在使用哈希表时一定会遇到冲突，哪怕我们使用了完美的哈希函数，当输入的键足够多也会产生冲突。然而多数的哈希函数都是不够完美的，所以仍然存在发生哈希碰撞的可能，这时就需要一些方法来解决哈希碰撞的问题，**常见方法的就是开放寻址法和拉链法**。

> 需要注意的是，这里提到的哈希碰撞不是多个键对应的哈希完全相等，可能是多个哈希的部分相等，例如：两个键对应哈希的前四个字节相同。

##### 开放寻址法

核心思想是**依次探测和比较数组中的元素以判断目标键值对是否存在于哈希表中**，如果我们使用开放寻址法来实现哈希表，那么实现哈希表底层的数据结构就是数组，不过因为数组的长度有限，向哈希表写入 (author, draven) 这个键值对时会从如下的索引开始遍历：

```go
index := hash("author") % array.len
```

当我们向当前哈希表写入新的数据时，如果发生了冲突，就会将键值对写入到下一个索引不为空的位置：
![image](https://mail.wangkekai.cn/1616162749340.jpg)

如上图所示（开放寻址法写入数据），当 Key3 与已经存入哈希表中的两个键值对 Key1 和 Key2 发生冲突时，Key3 会被写入 Key2 后面的空闲位置。当我们再去读取 Key3 对应的值时就会先获取键的哈希并取模，这会先帮助我们找到 Key1，找到 Key1 后发现它与 Key 3 不相等，所以会继续查找后面的元素，直到内存为空或者找到目标元素。

![image](https://mail.wangkekai.cn/1616162852151.jpg)
开放寻址法读取数据

当需要查找某个键对应的值时，会从索引的位置开始线性探测数组，找到目标键值对或者空内存就意味着这一次查询操作的结束。

开放寻址法中对性能影响最大的是装载因子，它是数组中元素的数量与数组大小的比值。随着装载因子的增加，线性探测的平均用时就会逐渐增加，这会影响哈希表的读写性能。当装载率超过 70% 之后，哈希表的性能就会急剧下降，而一旦装载率达到 100%，整个哈希表就会完全失效，这时查找和插入任意元素的时间复杂度都是 𝑂(𝑛) 的，这时需要遍历数组中的全部元素，所以在实现哈希表时一定要关注装载因子的变化。

##### 拉链法

与开放地址法相比，拉链法是哈希表最常见的实现方法，大多数的编程语言都用拉链法实现哈希表，它的实现比较开放地址法稍微复杂一些，但是平均查找的长度也比较短，各个用于存储节点的内存都是动态申请的，可以节省比较多的存储空间。

实现拉链法一般会使用数组加上链表，不过一些编程语言会在拉链法的哈希中引入红黑树以优化性能，拉链法会使用链表数组作为哈希底层的数据结构，我们可以将它看成可以扩展的二维数组：

![image](https://mail.wangkekai.cn/1616163025556.jpg)

如上图所示（拉链法写入数据），当我们需要将一个键值对 (Key6, Value6) 写入哈希表时，键值对中的键 Key6 都会先经过一个哈希函数，哈希函数返回的哈希会帮助我们选择一个桶，和开放地址法一样，选择桶的方式是直接对哈希返回的结果取模：

```go
index := hash("Key6") % array.len
```

选择了 2 号桶后就可以遍历当前桶中的链表了，在遍历链表的过程中会遇到以下两种情况：

1. 找到键相同的键值对 — 更新键对应的值；
2. 没有找到键相同的键值对 — 在链表的末尾追加新的键值对；


如果要在哈希表中获取某个键对应的值，会经历如下的过程：
![image](https://mail.wangkekai.cn/1616163141176.jpg)

Key11 展示了一个键在哈希表中不存在的例子，当哈希表发现它命中 4 号桶时，它会依次遍历桶中的链表，然而遍历到链表的末尾也没有找到期望的键，所以哈希表中没有该键对应的值。

在一个性能比较好的哈希表中，每一个桶中都应该有 0~1 个元素，有时会有 2~3 个，很少会超过这个数量。计算哈希、定位桶和遍历链表三个过程是哈希表读写操作的主要开销，使用拉链法实现的哈希也有装载因子这一概念：

<center>装载因子:=元素数量÷桶数量</center>

与开放地址法一样，拉链法的装载因子越大，哈希的读写性能就越差。在一般情况下使用拉链法的哈希表装载因子都不会超过 1，当哈希表的装载因子较大时会触发哈希的扩容，创建更多的桶来存储哈希中的元素，保证性能不会出现严重的下降。如果有 1000 个桶的哈希表存储了 10000 个键值对，它的性能是保存 1000 个键值对的 1/10，但是仍然比在链表中直接读写好 1000 倍。

### 3.3.2 数据结构

Go 语言运行时同时使用了多个数据结构组合表示哈希表，其中 runtime.hmap 是最核心的结构体，我们先来了解一下该结构体的内部字段：

```go
type hmap struct {
    count     int
    flags     uint8
    B         uint8
    noverflow uint16
    hash0     uint32

    buckets    unsafe.Pointer
    oldbuckets unsafe.Pointer
    nevacuate  uintptr

    extra *mapextra
}

type mapextra struct {
    overflow    *[]*bmap
    oldoverflow *[]*bmap
    nextOverflow *bmap
}
```

1. `count` 表示当前哈希表中的元素数量；
2. B 表示当前哈希表持有的 `buckets` 数量，但是因为哈希表中桶的数量都 2 的倍数，所以该字段会存储对数，也就是 `len(buckets) == 2^B`；
3. `hash0` 是哈希的种子，它能为哈希函数的结果引入随机性，这个值在创建哈希表时确定，并在调用哈希函数时作为参数传入；
4. `oldbuckets` 是哈希在扩容时用于保存之前 `buckets` 的字段，它的大小是当前 `buckets` 的一半；

![image](https://mail.wangkekai.cn/1616163355154.jpg)

哈希表 `runtime.hmap` 的桶是 `runtime.bmap`。每一个 `runtime.bmap` 都能存储 8 个键值对，当哈希表中存储的数据过多，单个桶已经装满时就会使用 `extra.nextOverflow` 中桶存储溢出的数据。

上述两种不同的桶在内存中是连续存储的，我们在这里将它们分别称为正常桶和溢出桶，上图中黄色的 runtime.bmap 就是正常桶，绿色的 runtime.bmap 是溢出桶，溢出桶是在 Go 语言还使用 C 语言实现时使用的设计，由于它能够减少扩容的频率所以一直使用至今。

桶的结构体 runtime.bmap 在 Go 语言源代码中的定义只包含一个简单的 tophash 字段，tophash 存储了键的哈希的高 8 位，通过比较不同键的哈希的高 8 位可以减少访问键值对次数以提高性能：

```go
type bmap struct {
    tophash [bucketCnt]uint8
}
```

在运行期间，runtime.bmap 结构体其实不止包含 tophash 字段，因为哈希表中可能存储不同类型的键值对，而且 Go 语言也不支持泛型，所以键值对占据的内存空间大小只能在编译时进行推导。runtime.bmap 中的其他字段在运行时也都是通过计算内存地址的方式访问的，所以它的定义中就不包含这些字段，不过我们能根据编译期间的 cmd/compile/internal/gc.bmap 函数重建它的结构：

```go
type bmap struct {
    topbits  [8]uint8
    keys     [8]keytype
    values   [8]valuetype
    pad      uintptr
    overflow uintptr
}
```

随着哈希表存储的数据逐渐增多，我们会扩容哈希表或者使用额外的桶存储溢出的数据，不会让单个桶中的数据超过 8 个，不过溢出桶只是临时的解决方案，创建过多的溢出桶最终也会导致哈希的扩容。

从 Go 语言哈希的定义中可以发现，改进元素比数组和切片复杂得多，它的结构体中不仅包含大量字段，还使用复杂的嵌套结构，后面的小节会详细介绍不同字段的作用。

### 3.3.3 初始化

Go 语言初始化哈希的两种方法 — **通过字面量和运行时**.

#### 字面量

目前的现代编程语言基本都支持使用字面量的方式初始化哈希，一般都会使用 key: value 的语法来表示键值对，Go 语言中也不例外：

```go
hash := map[string]int{
    "1": 2,
    "3": 4,
    "5": 6,
}
```

我们需要在初始化哈希时声明键值对的类型，这种使用字面量初始化的方式最终都会通过 cmd/compile/internal/gc.maplit 初始化，我们来分析一下该函数初始化哈希的过程：

```go
func maplit(n *Node, m *Node, init *Nodes) {
    a := nod(OMAKE, nil, nil)
    a.Esc = n.Esc
    a.List.Set2(typenod(n.Type), nodintconst(int64(n.List.Len())))
    litas(m, a, init)

    entries := n.List.Slice()
    if len(entries) > 25 {
        ...
        return
    }

    // Build list of var[c] = expr.
    // Use temporaries so that mapassign1 can have addressable key, elem.
    ...
}
```

当哈希表中的元素数量少于或者等于 25 个时，编译器会将字面量初始化的结构体转换成以下的代码，将所有的键值对一次加入到哈希表中：

```go
hash := make(map[string]int, 3)
hash["1"] = 2
hash["3"] = 4
hash["5"] = 6
```

这种初始化的方式与的数组和切片几乎完全相同，由此看来集合类型的初始化在 Go 语言中有着相同的处理逻辑。

<font color=red>一旦哈希表中元素的数量超过了 25 个，编译器会创建两个数组分别存储键和值</font>，这些键值对会通过如下所示的 for 循环加入哈希：

```go
hash := make(map[string]int, 26)
vstatk := []string{"1", "2", "3", ... ， "26"}
vstatv := []int{1, 2, 3, ... , 26}
for i := 0; i < len(vstak); i++ {
    hash[vstatk[i]] = vstatv[i]
}
```

这里展开的两个切片 `vstatk` 和 `vstatv` 还会被编辑器继续展开，具体的展开方式可以阅读上一节了解切片的初始化，不过无论使用哪种方法，使用字面量初始化的过程都会使用 Go 语言中的关键字 make 来创建新的哈希并通过最原始的 [] 语法向哈希追加元素。

#### 运行时

当创建的哈希被分配到栈上并且其容量小于 BUCKETSIZE = 8 时，Go 语言在编译阶段会使用如下方式快速初始化哈希，这也是编译器对小容量的哈希做的优化：

```go
var h *hmap
var hv hmap
var bv bmap
h := &hv
b := &bv
h.buckets = b
h.hash0 = fashtrand0()
```

除了上述特定的优化之外，**无论 make 是从哪里来的，只要我们使用 make 创建哈希，Go 语言编译器都会在类型检查期间将它们转换成 runtime.makemap，使用字面量初始化哈希也只是语言提供的辅助工具**，最后调用的都是 runtime.makemap：

```go
func makemap(t *maptype, hint int, h *hmap) *hmap {
    mem, overflow := math.MulUintptr(uintptr(hint), t.bucket.size)
    if overflow || mem > maxAlloc {
        hint = 0
    }

    if h == nil {
        h = new(hmap)
    }
    h.hash0 = fastrand()

    B := uint8(0)
    for overLoadFactor(hint, B) {
        B++
    }
    h.B = B

    if h.B != 0 {
        var nextOverflow *bmap
        h.buckets, nextOverflow = makeBucketArray(t, h.B, nil)
        if nextOverflow != nil {
            h.extra = new(mapextra)
            h.extra.nextOverflow = nextOverflow
        }
    }
    return h
}
```

这个函数会按照下面的步骤执行：

1. 计算哈希占用的内存是否溢出或者超出能分配的最大值；
2. 调用 `runtime.fastrand` 获取一个随机的哈希种子；
3. 根据传入的 `hint` 计算出需要的最小需要的桶的数量；
4. 使用 `runtime.makeBucketArray` 创建用于保存桶的数组；

`runtime.makeBucketArray` 会根据传入的 B 计算出的需要创建的桶数量并在内存中分配一片连续的空间用于存储数据：

```go
func makeBucketArray(t *maptype, b uint8, dirtyalloc unsafe.Pointer) (buckets unsafe.Pointer, nextOverflow *bmap) {
    base := bucketShift(b)
    nbuckets := base
    if b >= 4 {
        nbuckets += bucketShift(b - 4)
        sz := t.bucket.size * nbuckets
        up := roundupsize(sz)
        if up != sz {
            nbuckets = up / t.bucket.size
        }
    }

    buckets = newarray(t.bucket, int(nbuckets))
    if base != nbuckets {
        nextOverflow = (*bmap)(add(buckets, base*uintptr(t.bucketsize)))
        last := (*bmap)(add(buckets, (nbuckets-1)*uintptr(t.bucketsize)))
        last.setoverflow(t, (*bmap)(buckets))
    }
    return buckets, nextOverflow
}
```

- 当桶的数量小于 2^4 时，由于数据较少、使用溢出桶的可能性较低，会省略创建的过程以减少额外开销；
- 当桶的数量多于 2^4 时，会额外创建 2^𝐵−4 个溢出桶；

根据上述代码，我们能确定在正常情况下，正常桶和溢出桶在内存中的存储空间是连续的，只是被 `runtime.hmap` 中的不同字段引用，当溢出桶数量较多时会通过 `runtime.newobject` 创建新的溢出桶。

### 3.3.4 读写操作

哈希表的访问一般都是通过下标或者遍历进行的：

```go
_ = hash[key]

for k, v := range hash {
    // k, v
}
```

这两种方式虽然都能读取哈希表的数据，但是使用的函数和底层原理完全不同。前者需要知道哈希的键并且一次只能获取单个键对应的值，而后者可以遍历哈希中的全部键值对，访问数据时也不需要预先知道哈希的键。

数据结构的写一般指的都是增加、删除和修改，增加和修改字段都使用索引和赋值语句，而删除字典中的数据需要使用关键字 `delete`：

```go
hash[key] = value
hash[key] = newValue
delete(hash, key)
```

#### 访问

在编译的类型检查期间，`hash[key]` 以及类似的操作都会被转换成哈希的 `OINDEXMAP` 操作，中间代码生成阶段会在 cmd/compile/internal/gc.walkexpr 函数中将这些 `OINDEXMAP` 操作转换成如下的代码：

```go
v     := hash[key] // => v     := *mapaccess1(maptype, hash, &key)
v, ok := hash[key] // => v, ok := mapaccess2(maptype, hash, &key)
```

赋值语句左侧接受参数的个数会决定使用的运行时方法：

- 当接受一个参数时，会使用 runtime.mapaccess1，该函数仅会返回一个指向目标值的指针；
- 当接受两个参数时，会使用 runtime.mapaccess2，除了返回目标值之外，它还会返回一个用于表示当前键对应的值是否存在的 bool 值：

`runtime.mapaccess1` 会先通过哈希表设置的哈希函数、种子获取当前键对应的哈希，  
再通过 `runtime.bucketMask` 和 `runtime.add` 拿到该键值对所在的桶序号和哈希高位的 8 位数字。

```go
func mapaccess1(t *maptype, h *hmap, key unsafe.Pointer) unsafe.Pointer {
    alg := t.key.alg
    hash := alg.hash(key, uintptr(h.hash0))
    m := bucketMask(h.B)
    b := (*bmap)(add(h.buckets, (hash&m)*uintptr(t.bucketsize)))
    top := tophash(hash)
    bucketloop:
    for ; b != nil; b = b.overflow(t) {
        for i := uintptr(0); i < bucketCnt; i++ {
            if b.tophash[i] != top {
                if b.tophash[i] == emptyRest {
                    break bucketloop
                }
                continue
            }
            k := add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
            if alg.equal(key, k) {
                v := add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.valuesize))
                return v
            }
        }
    }
    return unsafe.Pointer(&zeroVal[0])
}
```

在 `bucketloop` 循环中，哈希会依次遍历正常桶和溢出桶中的数据，它会先比较哈希的高 8 位和桶中存储的 `tophash`，后比较传入的和桶中的值以加速数据的读写。用于选择桶序号的是哈希的最低几位，而用于加速访问的是哈希的高 8 位，这种设计能够减少同一个桶中有大量相等 `tophash` 的概率影响性能。

![image](https://mail.wangkekai.cn/1616420989023.jpg)

如上图所示，每一个桶都是一整片的内存空间，当发现桶中的 tophash 与传入键的 tophash 匹配之后，我们会通过指针和偏移量获取哈希中存储的键 keys[0] 并与 key 比较，如果两者相同就会获取目标值的指针 values[0] 并返回。

另一个同样用于访问哈希表中数据的 `runtime.mapaccess2` 只是在 `runtime.mapaccess1` 的基础上多返回了一个标识键值对是否存在的 bool 值：

```go
func mapaccess2(t *maptype, h *hmap, key unsafe.Pointer) (unsafe.Pointer, bool) {
    ...
    bucketloop:
    for ; b != nil; b = b.overflow(t) {
        for i := uintptr(0); i < bucketCnt; i++ {
            if b.tophash[i] != top {
                if b.tophash[i] == emptyRest {
                    break bucketloop
                }
                continue
            }
            k := add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
            if alg.equal(key, k) {
                v := add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.valuesize))
                return v, true
            }
        }
    }
    return unsafe.Pointer(&zeroVal[0]), false
}
```

使用 `v, ok := hash[k]` 的形式访问哈希表中元素时，我们能够通过这个布尔值更准确地知道当 `v == nil` 时，v 到底是哈希中存储的元素还是表示该键对应的元素不存在，**所以在访问哈希时，更推荐使用这种方式判断元素是否存在**。

上面的过程是在正常情况下，访问哈希表中元素时的表现，然而与数组一样，哈希表可能会在装载因子过高或者溢出桶过多时进行扩容，哈希表扩容并不是原子过程，在扩容的过程中保证哈希的访问是比较有意思的话题。

#### 写入

当形如 `hash[k]` 的表达式出现在赋值符号左侧时，该表达式也会在编译期间转换成 `runtime.mapassign` 函数的调用，该函数与 `runtime.mapaccess1` 比较相似，我们将其分成几个部分依次分析，首先是函数会根据传入的键拿到对应的哈希和桶：

```go
func mapassign(t *maptype, h *hmap, key unsafe.Pointer) unsafe.Pointer {
    alg := t.key.alg
    hash := alg.hash(key, uintptr(h.hash0))

    h.flags ^= hashWriting

    again:
    bucket := hash & bucketMask(h.B)
    b := (*bmap)(unsafe.Pointer(uintptr(h.buckets) + bucket*uintptr(t.bucketsize)))
    top := tophash(hash)
```

然后通过遍历比较桶中存储的 `tophash` 和键的哈希，如果找到了相同结果就会返回目标位置的地址。其中 `inserti` 表示目标元素在桶中的索引，`insertk` 和 val 分别表示键值对的地址，获得目标地址之后会通过算术计算寻址获得键值对 k 和 val：

```go
    var inserti *uint8
    var insertk unsafe.Pointer
    var val unsafe.Pointer
    bucketloop:
    for {
        for i := uintptr(0); i < bucketCnt; i++ {
            if b.tophash[i] != top {
                if isEmpty(b.tophash[i]) && inserti == nil {
                    inserti = &b.tophash[i]
                    insertk = add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
                    val = add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.valuesize))
                }
                if b.tophash[i] == emptyRest {
                    break bucketloop
                }
                continue
            }
            k := add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
            if !alg.equal(key, k) {
                continue
            }
            val = add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.valuesize))
            goto done
        }
        ovf := b.overflow(t)
        if ovf == nil {
            break
        }
        b = ovf
    }
```

上述的 for 循环会依次遍历正常桶和溢出桶中存储的数据，整个过程会分别判断 tophash 是否相等、key 是否相等，遍历结束后会从循环中跳出。

![image](https://mail.wangkekai.cn/1616421537342.jpg)

如果当前桶已经满了，哈希会调用 `runtime.hmap.newoverflow` 创建新桶或者使用 `runtime.hmap` 预先在 `noverflow` 中创建好的桶来保存数据，新创建的桶不仅会被追加到已有桶的末尾，还会增加哈希表的 `noverflow` 计数器。

```go
    if inserti == nil {
        newb := h.newoverflow(t, b)
        inserti = &newb.tophash[0]
        insertk = add(unsafe.Pointer(newb), dataOffset)
        val = add(insertk, bucketCnt*uintptr(t.keysize))
    }

    typedmemmove(t.key, insertk, key)
    *inserti = top
    h.count++

done:
    return val
}
```

如果当前键值对在哈希中不存在，哈希会为新键值对规划存储的内存地址，通过 `runtime.typedmemmove` 将键移动到对应的内存空间中并返回键对应值的地址 val。如果当前键值对在哈希中存在，那么就会直接返回目标区域的内存地址，哈希并不会在 `runtime.mapassign` 这个运行时函数中将值拷贝到桶中，该函数只会返回内存地址，真正的赋值操作是在编译期间插入的：

```go
00018 (+5) CALL runtime.mapassign_fast64(SB)
00020 (5) MOVQ 24(SP), DI               ;; DI = &value
00026 (5) LEAQ go.string."88"(SB), AX   ;; AX = &"88"
00027 (5) MOVQ AX, (DI)                 ;; *DI = AX
```

`runtime.mapassign_fast64` 与 `runtime.mapassign` 函数的逻辑差不多，我们需要关注的是后面的三行代码，其中 `24(SP)` 是该函数返回的值地址，我们通过 LEAQ 指令将字符串的地址存储到寄存器 AX 中，MOVQ 指令将字符串 "88" 存储到了目标地址上完成了这次哈希的写入。

#### 扩容

随着哈希表中元素的逐渐增加，哈希的性能会逐渐恶化，所以我们需要更多的桶和更大的内存保证哈希的读写性能：

```go
func mapassign(t *maptype, h *hmap, key unsafe.Pointer) unsafe.Pointer {
    ...
    if !h.growing() && (overLoadFactor(h.count+1, h.B) || tooManyOverflowBuckets(h.noverflow, h.B)) {
        hashGrow(t, h)
        goto again
    }
    ...
}
```

runtime.mapassign 函数会在以下两种情况发生时触发哈希的扩容：

1. 装载因子已经超过 6.5；
2. 哈希使用了太多溢出桶；

> 不过因为 Go 语言哈希的扩容不是一个原子的过程，所以 runtime.mapassign 还需要判断当前哈希是否已经处于扩容状态，避免二次扩容造成混乱。

根据触发的条件不同扩容的方式分成两种，**如果这次扩容是溢出的桶太多导致的，那么这次扩容就是等量扩容 sameSizeGrow**，sameSizeGrow 是一种特殊情况下发生的扩容，当我们持续向哈希中插入数据并将它们全部删除时，如果哈希表中的数据量没有超过阈值，就会不断积累溢出桶造成缓慢的内存泄漏。runtime: limit the number of map overflow buckets 引入了 sameSizeGrow 通过复用已有的哈希扩容机制解决该问题，一旦哈希中出现了过多的溢出桶，它会创建新桶保存数据，垃圾回收会清理老的溢出桶并释放内存。  
扩容的入口是 runtime.hashGrow：

```go
func hashGrow(t *maptype, h *hmap) {
    bigger := uint8(1)
    if !overLoadFactor(h.count+1, h.B) {
        bigger = 0
        h.flags |= sameSizeGrow
    }
    oldbuckets := h.buckets
    newbuckets, nextOverflow := makeBucketArray(t, h.B+bigger, nil)

    h.B += bigger
    h.flags = flags
    h.oldbuckets = oldbuckets
    h.buckets = newbuckets
    h.nevacuate = 0
    h.noverflow = 0

    h.extra.oldoverflow = h.extra.overflow
    h.extra.overflow = nil
    h.extra.nextOverflow = nextOverflow
}
```

哈希在扩容的过程中会通过 `runtime.makeBucketArray` 创建一组新桶和预创建的溢出桶，随后将原有的桶数组设置到 `oldbuckets` 上并将新的空桶设置到 `buckets` 上，溢出桶也使用了相同的逻辑更新，下图展示了触发扩容后的哈希：

![image](https://mail.wangkekai.cn/1616422032632.jpg)

我们在 `runtime.hashGrow` 中还看不出来等量扩容和翻倍扩容的太多区别，等量扩容创建的新桶数量只是和旧桶一样，该函数中只是创建了新的桶，并没有对数据进行拷贝和转移。哈希表的数据迁移的过程在是 `runtime.evacuate` 中完成的，它会对传入桶中的元素进行再分配。

```go
func evacuate(t *maptype, h *hmap, oldbucket uintptr) {
    b := (*bmap)(add(h.oldbuckets, oldbucket*uintptr(t.bucketsize)))
    newbit := h.noldbuckets()
    if !evacuated(b) {
        var xy [2]evacDst
        x := &xy[0]
        x.b = (*bmap)(add(h.buckets, oldbucket*uintptr(t.bucketsize)))
        x.k = add(unsafe.Pointer(x.b), dataOffset)
        x.v = add(x.k, bucketCnt*uintptr(t.keysize))

        y := &xy[1]
        y.b = (*bmap)(add(h.buckets, (oldbucket+newbit)*uintptr(t.bucketsize)))
        y.k = add(unsafe.Pointer(y.b), dataOffset)
        y.v = add(y.k, bucketCnt*uintptr(t.keysize))
```

`runtime.evacuate` 会将一个旧桶中的数据分流到两个新桶，所以它会创建两个用于保存分配上下文的 `runtime.evacDst` 结构体，这两个结构体分别指向了一个新桶：

![image](https://mail.wangkekai.cn/1616423468013.jpg)

**如果这是等量扩容，那么旧桶与新桶之间是一对一的关系，所以两个 runtime.evacDst 只会初始化一个**。而当哈希表的容量翻倍时，每个旧桶的元素会都分流到新创建的两个桶中，这里仔细分析一下分流元素的逻辑：

```go
        for ; b != nil; b = b.overflow(t) {
        k := add(unsafe.Pointer(b), dataOffset)
        v := add(k, bucketCnt*uintptr(t.keysize))
        for i := 0; i < bucketCnt; i, k, v = i+1, add(k, uintptr(t.keysize)), add(v, uintptr(t.valuesize)) {
            top := b.tophash[i]
            k2 := k
            var useY uint8
            hash := t.key.alg.hash(k2, uintptr(h.hash0))
            if hash&newbit != 0 {
                useY = 1
            }
            b.tophash[i] = evacuatedX + useY
            dst := &xy[useY]

            if dst.i == bucketCnt {
                dst.b = h.newoverflow(t, dst.b)
                dst.i = 0
                dst.k = add(unsafe.Pointer(dst.b), dataOffset)
                dst.v = add(dst.k, bucketCnt*uintptr(t.keysize))
            }
            dst.b.tophash[dst.i&(bucketCnt-1)] = top
            typedmemmove(t.key, dst.k, k)
            typedmemmove(t.elem, dst.v, v)
            dst.i++
            dst.k = add(dst.k, uintptr(t.keysize))
            dst.v = add(dst.v, uintptr(t.valuesize))
        }
    }
    ...
}
```

只使用哈希函数是不能定位到具体某一个桶的，哈希函数只会返回很长的哈希，例如：b72bfae3f3285244c4732ce457cca823bc189e0b，我们还需一些方法将哈希映射到具体的桶上。**我们一般都会使用取模或者位操作来获取桶的编号**，假如当前哈希中包含 4 个桶，那么它的桶掩码就是 0b11(3)，使用位操作就会得到 3， 我们就会在 3 号桶中存储该数据：

```go
0xb72bfae3f3285244c4732ce457cca823bc189e0b & 0b11 #=> 0
```

如果新的哈希表有 8 个桶，在大多数情况下，原来经过桶掩码 0b11 结果为 3 的数据会因为桶掩码增加了一位变成 0b111 而分流到新的 3 号和 7 号桶，所有数据也都会被 `runtime.typedmemmove` 拷贝到目标桶中：

![image](https://mail.wangkekai.cn/1616423620746.jpg)

`runtime.evacuate` 最后会调用 `runtime.advanceEvacuationMark` 增加哈希的 `nevacuate` 计数器并在所有的旧桶都被分流后清空哈希的 `oldbuckets` 和 `oldoverflow`：

```go
func advanceEvacuationMark(h *hmap, t *maptype, newbit uintptr) {
    h.nevacuate++
    stop := h.nevacuate + 1024
    if stop > newbit {
        stop = newbit
    }
    for h.nevacuate != stop && bucketEvacuated(t, h, h.nevacuate) {
        h.nevacuate++
    }
    if h.nevacuate == newbit { // newbit == # of oldbuckets
        h.oldbuckets = nil
        if h.extra != nil {
            h.extra.oldoverflow = nil
        }
        h.flags &^= sameSizeGrow
    }
}
```

之前在分析哈希表访问函数 runtime.mapaccess1 时其实省略了扩容期间获取键值对的逻辑，当哈希表的 oldbuckets 存在时，会先定位到旧桶并在该桶没有被分流时从中获取键值对。

```go
func mapaccess1(t *maptype, h *hmap, key unsafe.Pointer) unsafe.Pointer {
    ...
    alg := t.key.alg
    hash := alg.hash(key, uintptr(h.hash0))
    m := bucketMask(h.B)
    b := (*bmap)(add(h.buckets, (hash&m)*uintptr(t.bucketsize)))
    if c := h.oldbuckets; c != nil {
        if !h.sameSizeGrow() {
            m >>= 1
        }
        oldb := (*bmap)(add(c, (hash&m)*uintptr(t.bucketsize)))
        if !evacuated(oldb) {
            b = oldb
        }
    }
bucketloop:
    ...
}
```

因为旧桶中的元素还没有被 runtime.evacuate 函数分流，其中还保存着我们需要使用的数据，所以旧桶会替代新创建的空桶提供数据。

我们在 runtime.mapassign 函数中也省略了一段逻辑，**当哈希表正在处于扩容状态时，每次向哈希表写入值时都会触发 runtime.growWork 增量拷贝哈希表中的内容**：

```go
func mapassign(t *maptype, h *hmap, key unsafe.Pointer) unsafe.Pointer {
    ...
again:
    bucket := hash & bucketMask(h.B)
    if h.growing() {
        growWork(t, h, bucket)
    }
    ...
}
```

当然除了写入操作之外，删除操作也会在哈希表扩容期间触发 `runtime.growWork`，触发的方式和代码与这里的逻辑几乎完全相同，都是计算当前值所在的桶，然后拷贝桶中的元素。

我们简单总结一下哈希表扩容的设计和原理，哈希在存储元素过多时会触发扩容操作，每次都会将桶的数量翻倍，扩容过程不是原子的，而是通过 `runtime.growWork` 增量触发的，在扩容期间访问哈希表时会使用旧桶，向哈希表写入数据时会触发旧桶元素的分流。除了这种正常的扩容之外，为了解决大量写入、删除造成的内存泄漏问题，哈希引入了 `sameSizeGrow` 这一机制，在出现较多溢出桶时会整理哈希的内存减少空间的占用。

#### 删除

如果想要删除哈希中的元素，就需要使用 Go 语言中的 `delete` 关键字，这个关键字的唯一作用就是将某一个键对应的元素从哈希表中删除，无论是该键对应的值是否存在，这个内建的函数都不会返回任何的结果。

![image](https://mail.wangkekai.cn/1616423852268.jpg)

在编译期间，`delete` 关键字会被转换成操作为 `ODELETE` 的节点，而 `cmd/compile/internal/gc.walkexpr` 会将 `ODELETE` 节点转换成 `runtime.mapdelete` 函数簇中的一个，包括 `runtime.mapdelete`、`mapdelete_faststr`、`mapdelete_fast32` 和 `mapdelete_fast64`：

```go
func walkexpr(n *Node, init *Nodes) *Node {
    switch n.Op {
    case ODELETE:
        init.AppendNodes(&n.Ninit)
        map_ := n.List.First()
        key := n.List.Second()
        map_ = walkexpr(map_, init)
        key = walkexpr(key, init)

        t := map_.Type
        fast := mapfast(t)
        if fast == mapslow {
            key = nod(OADDR, key, nil)
        }
        n = mkcall1(mapfndel(mapdelete[fast], t), nil, init, typename(t), map_, key)
    }
}
```

这些函数的实现其实差不多，我们挑选其中的 `runtime.mapdelete` 分析一下。哈希表的删除逻辑与写入逻辑很相似，只是触发哈希的删除需要使用关键字，如果在删除期间遇到了哈希表的扩容，就会分流桶中的元素，分流结束之后会找到桶中的目标元素完成键值对的删除工作。

```go
func mapdelete(t *maptype, h *hmap, key unsafe.Pointer) {
    ...
    if h.growing() {
        growWork(t, h, bucket)
    }
    ...
search:
    for ; b != nil; b = b.overflow(t) {
        for i := uintptr(0); i < bucketCnt; i++ {
            if b.tophash[i] != top {
                if b.tophash[i] == emptyRest {
                    break search
                }
                continue
            }
            k := add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
            k2 := k
            if !alg.equal(key, k2) {
                continue
            }
            *(*unsafe.Pointer)(k) = nil
            v := add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.valuesize))
            *(*unsafe.Pointer)(v) = nil
            b.tophash[i] = emptyOne
            ...
        }
    }
}
```

我们其实只需要知道 delete 关键字在编译期间经过类型检查和中间代码生成阶段被转换成 `runtime.mapdelete` 函数簇中的一员，用于处理删除逻辑的函数与哈希表的 runtime.mapassign 几乎完全相同，不太需要刻意关注。

**Go 语言使用拉链法来解决哈希碰撞的问题实现了哈希表，它的访问、写入和删除等操作都在编译期间转换成了运行时的函数或者方法**。哈希在每一个桶中存储键对应哈希的前 8 位，当对哈希进行操作时，这些 tophash 就成为可以帮助哈希快速遍历桶中元素的缓存。

哈希表的每个桶都只能存储 8 个键值对，一旦当前哈希的某个桶超出 8 个，新的键值对就会存储到哈希的溢出桶中。随着键值对数量的增加，溢出桶的数量和哈希的装载因子也会逐渐升高，超过一定范围就会触发扩容，扩容会将桶的数量翻倍，元素再分配的过程也是在调用写操作时增量进行的，不会造成性能的瞬时巨大抖动。

## 3.4 字符串

虽然字符串往往被看做一个整体，但是它实际上是一片连续的内存空间，我们也可以将它理解成一个由字符组成的数组。

**数组会占用一片连续的内存空间，而内存空间存储的字节共同组成了字符串，Go 语言中的字符串只是一个只读的字节数组**，下图展示了 "hello" 字符串在内存中的存储方式：

![image](https://mail.wangkekai.cn/1617025688813.jpg)

如果是代码中存在的字符串，编译器会将其标记成只读数据 `SRODATA`.

只读只意味着字符串会分配到只读的内存空间，但是 Go 语言只是不支持直接修改 `string` 类型变量的内存空间，我们仍然可以通过在 `string` 和 `[]byte` 类型之间反复转换实现修改这一目的：

1. 先将这段内存拷贝到堆或者栈上；
2. 将变量的类型转换成 `[]byte` 后并修改字节数据；
3. 将修改后的字节数组转换回 `string`；

很多编程语言的字符串也都是不可变的，**这种不可变的特性可以保证我们不会引用到意外发生改变的值**，而因为 Go 语言的字符串可以作为哈希的键，所以如果哈希的键是可变的，不仅会增加哈希实现的复杂度，还可能会影响哈希的比较。

### 3.4.1 数据结构

字符串在 Go 语言中的接口其实非常简单，每一个字符串在运行时都会使用如下的 reflect.StringHeader 表示，其中包含指向字节数组的指针和数组的大小：

```go
type StringHeader struct {
    Data uintptr
    Len  int
}
```

与切片的结构体相比，字符串只少了一个表示容量的 Cap 字段，而正是因为切片在 Go 语言的运行时表示与字符串高度相似，所以我们经常会说字符串是一个只读的切片类型。

```go
type SliceHeader struct {
    Data uintptr
    Len  int
    Cap  int
}
```

> 因为字符串作为只读的类型，我们并不会直接向字符串直接追加元素改变其本身的内存空间，**所有在字符串上的写入操作都是通过拷贝实现的**。

### 4.3.2 远程解析

解析器会在`词法分析`阶段解析字符串，词法分析阶段会对源文件中的字符串进行切片和分组，将原有无意义的字符流转换成 Token 序列。我们可以使用两种字面量方式在 Go 语言中声明字符串，即<font color=red>双引号和反引号</font>:

```go
str1 := "this is a string"
str2 := `this is another
string`
```

**使用双引号声明的字符串和其他语言中的字符串没有太多的区别，它只能用于单行字符串的初始化**，如果字符串内部出现双引号，需要使用 \ 符号避免编译器的解析错误，  
**而反引号声明的字符串可以摆脱单行的限制**。
> 当使用反引号时，因为双引号不再负责标记字符串的开始和结束，我们可以在字符串内部直接使用 "，在遇到需要手写 JSON 或者其他复杂数据格式的场景下非常方便。

```go
json := `{"author": "draven", "tags": ["golang"]}`
```

两种不同的声明方式其实也意味着 Go 语言编译器需要能够区分并且正确解析两种不同的字符串格式。解析字符串使用的扫描器 `cmd/compile/internal/syntax.scanner` 会将输入的字符串转换成 Token 流，`cmd/compile/internal/syntax.scanner.stdString` 方法是它用来解析使用双引号的标准字符串：

```go
func (s *scanner) stdString() {
    s.startLit()
    for {
        r := s.getr()
        if r == '"' {
            break
        }
        if r == '\\' {
            s.escape('"')
            continue
        }
        if r == '\n' {
            s.ungetr()
            s.error("newline in string")
            break
        }
        if r < 0 {
            s.errh(s.line, s.col, "string not terminated")
            break
        }
    }
    s.nlsemi = true
    s.lit = string(s.stopLit())
    s.kind = StringLit
    s.tok = _Literal
}
```

从这个方法的实现我们能分析出 Go 语言处理标准字符串的逻辑：

1. 标准字符串使用双引号表示开头和结尾；
2. 标准字符串需要使用反斜杠 \ 来逃逸双引号；
3. 标准字符串不能出现如下所示的隐式换行 \n；

```go
str := "start
end"
```

使用反引号声明的原始字符串的解析规则就非常简单了，`cmd/compile/internal/syntax.scanner.rawString` 会将非反引号的所有字符都划分到当前字符串的范围中，所以我们可以使用它支持复杂的多行字符串：

```go
func (s *scanner) rawString() {
    s.startLit()
    for {
        r := s.getr()
        if r == '`' {
            break
        }
        if r < 0 {
            s.errh(s.line, s.col, "string not terminated")
            break
        }
    }
    s.nlsemi = true
    s.lit = string(s.stopLit())
    s.kind = StringLit
    s.tok = _Literal
}
```

无论是标准字符串还是原始字符串都会被标记成 `StringLit` 并传递到语法分析阶段。在语法分析阶段，与字符串相关的表达式都会由 `cmd/compile/internal/gc.noder.basicLit` 方法处理：

```go
func (p *noder) basicLit(lit *syntax.BasicLit) Val {
    switch s := lit.Value; lit.Kind {
    case syntax.StringLit:
        if len(s) > 0 && s[0] == '`' {
            s = strings.Replace(s, "\r", "", -1)
        }
        u, _ := strconv.Unquote(s)
        return Val{U: u}
    }
}
```

无论是 import 语句中包的路径、结构体中的字段标签还是表达式中的字符串都会使用这个方法将原生字符串中最后的换行符删除并对字符串 Token 进行 Unquote，也就是去掉字符串两边的引号等无关干扰，还原其本来的面目。

### 3.4.3 拼接

Go 语言拼接字符串会使用 `+` 符号，编译器会将该符号对应的 `OADD` 节点转换成 `OADDSTR` 类型的节点，随后在 `cmd/compile/internal/gc.walkexpr` 中调用 `cmd/compile/internal/gc.addstr` 函数生成用于拼接字符串的代码：

```go
func walkexpr(n *Node, init *Nodes) *Node {
    switch n.Op {
    ...
    case OADDSTR:
        n = addstr(n, init)
    }
}
```

`cmd/compile/internal/gc.addstr` 能帮助我们在编译期间选择合适的函数对字符串进行拼接，该函数会根据带拼接的字符串数量选择不同的逻辑：

- 如果小于或者等于 5 个，那么会调用 `concatstring{2,3,4,5}` 等一系列函数；
- 如果超过 5 个，那么会选择 `runtime.concatstrings` 传入一个数组切片；

```go
func addstr(n *Node, init *Nodes) *Node {
    c := n.List.Len()

    buf := nodnil()
    args := []*Node{buf}
    for _, n2 := range n.List.Slice() {
        args = append(args, conv(n2, types.Types[TSTRING]))
    }

    var fn string
    if c <= 5 {
        fn = fmt.Sprintf("concatstring%d", c)
    } else {
        fn = "concatstrings"

        t := types.NewSlice(types.Types[TSTRING])
        slice := nod(OCOMPLIT, nil, typenod(t))
        slice.List.Set(args[1:])
        args = []*Node{buf, slice}
    }

    cat := syslook(fn)
    r := nod(OCALL, cat, nil)
    r.List.Set(args)
    ...

    return r
}
```

其实无论使用 `concatstring{2,3,4,5}` 中的哪一个，最终都会调用 `runtime.concatstrings`，它会先对遍历传入的切片参数，再过滤空字符串并计算拼接后字符串的长度。

```go
func concatstrings(buf *tmpBuf, a []string) string {
    idx := 0
    l := 0
    count := 0
    for i, x := range a {
        n := len(x)
        if n == 0 {
            continue
        }
        l += n
        count++
        idx = i
    }
    if count == 0 {
        return ""
    }
    if count == 1 && (buf != nil || !stringDataOnStack(a[idx])) {
        return a[idx]
    }
    s, b := rawstringtmp(buf, l)
    for _, x := range a {
        copy(b, x)
        b = b[len(x):]
    }
    return s
}
```

如果非空字符串的数量为 1 并且当前的字符串不在栈上，就可以直接返回该字符串，不需要做出额外操作。

![image](https://mail.wangkekai.cn/1617026595602.jpg)

在正常情况下，运行时会调用 `copy` 将输入的多个字符串拷贝到目标字符串所在的内存空间。新的字符串是一片新的内存空间，与原来的字符串也没有任何关联，一旦需要拼接的字符串非常大，拷贝带来的性能损失是无法忽略的。

### 3.4.4 类型转换

当我们使用 Go 语言解析和序列化 JSON 等数据格式时，经常需要将数据在 `string` 和 `[]byte` 之间来回转换，类型转换的开销并没有想象的那么小，我们经常会看到 `runtime.slicebytetostring` 等函数出现在火焰图中，成为程序的性能热点。

从字节数组到字符串的转换需要使用 `runtime.slicebytetostring` 函数，例如：`string(bytes)`，该函数在函数体中会先处理两种比较常见的情况，也就是长度为 0 或者 1 的字节数组，这两个情况处理起来都非常简单：

```go
func slicebytetostring(buf *tmpBuf, b []byte) (str string) {
    l := len(b)
    if l == 0 {
        return ""
    }
    if l == 1 {
        stringStructOf(&str).str = unsafe.Pointer(&staticbytes[b[0]])
        stringStructOf(&str).len = 1
        return
    }
    var p unsafe.Pointer
    if buf != nil && len(b) <= len(buf) {
        p = unsafe.Pointer(buf)
    } else {
        p = mallocgc(uintptr(len(b)), nil, false)
    }
    stringStructOf(&str).str = p
    stringStructOf(&str).len = len(b)
    memmove(p, (*(*slice)(unsafe.Pointer(&b))).array, uintptr(len(b)))
    return
}
```

处理过后会根据传入的缓冲区大小决定是否需要为新字符串分配一片内存空间，`runtime.stringStructOf` 会将传入的字符串指针转换成 runtime.stringStruct 结构体指针，然后设置结构体持有的字符串指针 str 和长度 len，最后通过 `runtime.memmove` 将原 `[]byte` 中的字节全部复制到新的内存空间中。

当我们想要将字符串转换成 `[]byte` 类型时，需要使用 `runtime.stringtoslicebyte` 函数，该函数的实现非常容易理解：

```go
func stringtoslicebyte(buf *tmpBuf, s string) []byte {
    var b []byte
    if buf != nil && len(s) <= len(buf) {
        *buf = tmpBuf{}
        b = buf[:len(s)]
    } else {
        b = rawbyteslice(len(s))
    }
    copy(b, s)
    return b
}
```

上述函数会根据是否传入缓冲区做出不同的处理：

- 当传入缓冲区时，它会使用传入的缓冲区存储 []byte；
- 当没有传入缓冲区时，运行时会调用 runtime.rawbyteslice 创建新的字节切片并将字符串中的内容拷贝过去；

![image](https://mail.wangkekai.cn/1617026877997.jpg)

字符串和 `[]byte` 中的内容虽然一样，但是字符串的内容是只读的，我们不能通过下标或者其他形式改变其中的数据，而 `[]byte` 中的内容是可以读写的。不过无论从哪种类型转换到另一种都需要拷贝数据，而内存拷贝的性能损耗会随着字符串和 `[]byte` 长度的增长而增长。

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
