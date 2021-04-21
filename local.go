package main

import "fmt"

func strStr(haystack string, needle string) int {
	m, n := len(haystack), len(needle)
counter:
	for i := 0; i+n <= m; i++ {
		for j := range needle {
			if haystack[i+j] != needle[j] {
				continue counter
			}
		}

		return i
	}

	return -1
}

const (
	a0 = iota
	a1
	a2 = iota
	a3 = 4
	a4
	a5
	a6 = iota
)

func testFunc() []func() {
	var funs []func()
	for i := 0; i < 2; i++ {
		// x := i -- 可以解决延迟求值的问题
		funs = append(funs, func() {
			println(&i, i) // 两次打印值相等
		})
	}
	return funs
}

type student struct {
	Name string
	Age  int
}

func main() {
	for i := 0; i < 10; i++ {
		go func() {
			fmt.Println(i)
		}()
	}

	// m := make(map[string]*student)
	// stus := []student{
	// 	{Name: "zhou", Age: 24},
	// 	{Name: "li", Age: 23},
	// 	{Name: "wang", Age: 22},
	// }
	// for k, stu := range stus {
	// 	m[stu.Name] = &stus[k]
	// 	println(stu.Name, "=>", &stus[k])
	// }

	// for k, v := range m {
	// 	println(k, "=>", v.Name)
	// }

	// res := testFunc()
	// for _, v := range res {
	// 	fmt.Println(v)
	// 	v()
	// }
	// fmt.Println(res)
	// res := strStr("hello", "lo")
	// fmt.Println(res)

	// fmt.Println(a0, a1, a2, a3, a4, a5, a6)
}
