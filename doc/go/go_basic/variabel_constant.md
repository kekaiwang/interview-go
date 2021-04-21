# 变量、常量

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
  - 结果不限于方法，字段也也一样；也不限于 `type alias`，`type defintion` 也是一样的，只要有重复的方法、字段，就会有这种提示，因为不知道该选择哪个。`ambiguous selector my.m1`

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

### 内存四区概念

A.数据类型本质：
固定内存大小的别名

B. 数据类型的作用：
编译器预算对象(变量)分配的内存空间大小。

![image](https://mail.wangkekai.cn/38B68C1D-284E-4566-82A1-1727D53A73E4.png)

C. 四区内存

![image](https://mail.wangkekai.cn/8BE69355-DD24-4CD2-B0C6-DA566A3B3174.png)

流程说明

1. 操作系统把物理硬盘代码 load 到内存
2. 操作系统把 c 代码分成四个区
3. 操作系统找到 main 函数入口执行

#### 栈区(Stack)

**空间较小，要求数据读写性能高，数据存放时间较短暂**。由编译器自动分配和释放，存放函数的参数值、函数的调用流程方法地址、局部变量等(局部变量如果产生逃逸现象，可能会挂在在堆区)。

#### 堆区(Heap)

​**空间充裕，数据存放时间较久**。一般由开发者分配及释放(**但是 Golang 中会根据变量的逃逸现象来选择是否分配到栈上或堆上**)，启动 Golang 的 GC 由 GC 清除机制自动回收。

#### 全局区-静态全局变量区

全局变量的开辟是在程序在 `main` 之前就已经放在内存中。而且对外完全可见。即作用域在全部代码中，任何同包代码均可随时使用，在变量会搞混淆，而且在局部函数中如果同名称变量使用 `:=` 赋值会出现编译错误。

> 全局变量最终在进程退出时，由操作系统回收。

常量区也归属于全局区，常量为存放数值字面值单位，即不可修改。或者说的有的常量是直接挂钩字面值的。

所以在 golang 中，常量是无法取出地址的，因为字面量符号并没有地址而言。
