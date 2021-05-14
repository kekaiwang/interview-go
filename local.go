package main

import (
	"fmt"
	"time"
)

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

type Student struct {
	Age  int
	Name string
}

var list map[string]Student

func add(a int, b int) int {
	A := a ^ b
	B := a & b
	fmt.Println(A, B)
	if B == 0 {
		return A
	} else {
		B = B << 1
		return add(A, B)
	}
}

func sortArrayByParityII(nums []int) []int {
	for i, j := 0, 0; i < len(nums); i += 2 {
		if nums[i]%2 == 1 {
			for j%2 == 1 {
				j += 2
			}

			fmt.Println(i, j, nums[i])
		}
	}

	return nums
}

func replaceDigts(s string) string {
	var byts = []byte(s)

	for i := 1; i < len(byts); i = i + 2 {
		byts[i] = shift(byts[i-1], byts[i])
	}

	return string(byts)
}

func shift(a, b byte) byte {
	return a + (b - 48)
}

type TestStruct struct {
	Age     int
	Created time.Time
}

func bubbleSort(arr []int) {
	len := len(arr)

	if len <= 1 {
		return
	}

	for i := 0; i < len; i++ {
		flag := false

		for j := 0; j < len-i-1; j++ {
			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
				flag = true
			}
		}

		if !flag {
			break
		}
	}
}

func selectPrime() []int {
	var res []int
	var i int

	for i < 100 {
		nums := 0
		for j := 1; j < i; j++ {
			if i%j == 0 {
				nums++
			}
		}

		if nums == 1 {
			res = append(res, i)
		}
		i++
	}

	return res
}

// 将未排序区间的数字放到已排序区间的合适位置
func insertSort(arr []int) {
	len := len(arr)

	if len <= 1 {
		return
	}

	for i := 1; i < len; i++ {
		value := arr[i]

		j := i - 1

		for ; j >= 0; j-- {
			if arr[j] > value {
				arr[j+1] = arr[j]
			} else {
				break
			}
		}

		arr[j+1] = value
	}

	return
}

// 从未排序区间选择最小的插入到合适位置
func selectionSort(arr []int) {
	len := len(arr)
	if len <= 1 {
		return
	}

	for i := 0; i < len; i++ {
		min := i

		for j := i + 1; j < len; j++ {
			if arr[min] > arr[j] {
				min = j
			}
		}

		if min != i {
			arr[min], arr[i] = arr[i], arr[min]
		}
	}

	return
}

func main() {

	arr := []int{4, 5, 3, 7, 8, 2, 1, 6}

	// bubbleSort(arr)
	// insertSort(arr)

	selectionSort(arr)
	fmt.Println(arr)

	// res := selectPrime()

	// fmt.Println(res)

	// var ts TestStruct

	// ts.Age = 1

	// fmt.Println(ts)

	// res := replaceDigts("a1b2c3d4e")
	// fmt.Println(res)

	// var ans float64 = 15 + 25 + 5.2
	// fmt.Println(ans)

	//创建trace文件
	// f, err := os.Create("trace.out")
	// if err != nil {
	// 	panic(err)
	// }

	// defer f.Close()

	// //启动trace goroutine
	// err = trace.Start(f)
	// if err != nil {
	// 	panic(err)
	// }
	// defer trace.Stop()

	// //main
	// fmt.Println("Hello World")

	// res := sortArrayByParityII([]int{4, 2, 5, 7, 6, 5, 8, 9})
	// fmt.Println(res)

	// fmt.Println(2 & 4)

	// add(4, 2)
	// task_cnt := math.MaxInt64

	// fmt.Println(task_cnt)

	// for i := 0; i < task_cnt; i++ {
	// 	go func(i int) {
	// 		//... do some busi...

	// 		fmt.Println("go func ", i, " goroutine count = ", runtime.NumGoroutine())
	// 	}(i)
	// }
	// var str = "thequickbrownfoxjumpsoverthelazydog"

	// for _, v := range str {
	// 	fmt.Printf("%T", v)
	// 	break
	// }

	// list = make(map[string]Student)

	// student := Student{"Aceld", 1}
	// fmt.Println(student)

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
