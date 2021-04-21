# map

## Map

```go
type Student struct {
    Name string
}

var list map[string]Student

func main() {
    list = make(map[string]Student)

    student := Student{"Aceld"}

    list["student"] = student
    // cannot assign to struct field list["student"].Name in map
    list["student"].Name = "LDB"

    fmt.Println(list["student"])
}
```

`map[string]Student` 的 value 是一个 Student 结构体，所以当 `list["student"] = student`,是一个值拷贝过程。而 `list["student"]` 则是一个值引用。那么值引用的特点是 `只读`。所以对 `list["student"].Name = "LDB"` 的修改是不允许的。

### 1.for range

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

![image](https://mail.wangkekai.cn/57937958-413D-45CA-B28E-0D97B56278DD.png)
