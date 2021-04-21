# 数组&切片

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

- `append` 切片 <font color=red>...</font> , 记住需要进行将第二个 slice 进行 `...` 打散再拼接。

```go
// 切记 ...
func main() {
    s1 := []int{1, 2, 3}
    s2 := []int{4, 5}
    s1 = append(s1, s2...)
}
```

### new和make的区别

- ​二者都是内存的分配（堆上）
  - 但是 `make` 只用于 `slice`、`map` 以及 `channel` 的初始化（非零值）；
  - 而 `new` 用于类型的内存分配，并且内存置为零。  
所以在我们编写程序的时候，就可以根据自己的需要很好的选择了。
- `make` 返回的还是这三个引用类型本身；而 `new` 返回的是指向类型的指针。
