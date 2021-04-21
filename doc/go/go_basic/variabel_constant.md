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
