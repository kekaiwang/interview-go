package main

import (
	"fmt"
	"sort"
	"time"
)

func getA() (i int) {
	// var i int
	defer func() {
		i = 3
	}()
	i = 1
	return i

}

// search. 二分查找
func search(nums []int, target int) int {
	high := len(nums) - 1
	low := 0

	for high >= low {
		middle := low + (high-low)/2

		if nums[middle] == target {
			return middle
		} else if nums[middle] > target {
			high = middle - 1
		} else {
			low = middle + 1
		}
	}

	return -1
}

// searchInsert. 二分查找插入
func searchInsert(nums []int, target int) int {
	high := len(nums) - 1
	low := 0
	res := len(nums)

	for high >= low {
		middle := (high-low)>>1 + low

		if nums[middle] >= target {
			res = middle
			high = middle - 1
		} else {
			low = middle + 1
		}
	}

	return res
}

func sortedSquares(nums []int) []int {
	n := len(nums)
	i, j, k := 0, n-1, n-1
	res := make([]int, n)

	for i <= j {
		left, right := nums[i]*nums[i], nums[j]*nums[j]

		if left > right {
			res[k] = left
			i++
		} else {
			res[k] = right
			j--
		}
		k--
	}

	return res
}

func reverse(nums []int) {
	for i, n := 0, len(nums); i < n/2; i++ {
		nums[i], nums[n-i-1] = nums[n-i-1], nums[i]
	}
}

func rotate(nums []int, k int) {
	k %= len(nums)
	reverse(nums)
	reverse(nums[:k])
	reverse(nums[k:])
	// n := len(nums) - 1
	// for k > 0 {

	// 	temp := nums[n]
	// 	for i := n; i > 0; i-- {
	// 		nums[i] = nums[i-1]
	// 	}
	// 	nums[0] = temp

	// 	k--
	// }
}

func moveZeroes(nums []int) {
	left, right, n := 0, 0, len(nums)
	for right < n {
		if nums[right] != 0 {
			nums[left], nums[right] = nums[right], nums[left]
			left++
		}
		right++
	}
}

func twoSum(numbers []int, target int) []int {
	// for i := 0; i < len(numbers); i++ {
	// 	low, high := i+1, len(numbers)-1
	// 	for low <= high {
	// 		middle := (high-low)>>1 + low

	// 		if numbers[middle] == target-numbers[i] {
	// 			return []int{i + 1, middle + 1}
	// 		} else if numbers[middle] > target-numbers[i] {
	// 			high = middle - 1
	// 		} else {
	// 			low = middle + 1
	// 		}
	// 	}
	// }

	left, right := 0, len(numbers)-1

	for left <= right {
		sum := numbers[left] + numbers[right]

		if sum == target {
			return []int{left + 1, right + 1}
		} else if sum < target {
			left++
		} else {
			right--
		}
	}

	return []int{-1, -1}
}

func reverseWords(s string) string {
	b := []byte(s)
	l := 0

	for i, v := range s {
		if v == ' ' || i == len(s)-1 {
			r := i - 1

			if i == len(s)-1 {
				r = i
			}

			for l < r {
				b[l], b[r] = b[r], b[l]
				l++
				r--
			}

			l = i + 1
		}
	}

	return string(b)
}

func lengthOfLongestSubstring(s string) int {
	res := map[byte]int{}
	n := len(s)
	right, ans := -1, 0

	for i := 0; i < n; i++ {
		if i != 0 {
			delete(res, s[i-1])
		}

		for right+1 < n && res[s[right+1]] == 0 {
			res[s[right+1]]++
			right++
		}

		ans = max(ans, right-i+1)
	}

	return ans
}

func max(x, y int) int {
	if x > y {
		return x
	}

	return y
}

func checkInclusion(s1, s2 string) bool {
	n, m := len(s1), len(s2)

	if n > m {
		return false
	}

	var res1, res2 [26]int

	for i, v := range s1 {
		res1[v-'a']++
		res2[s2[i]-'a']++
	}

	if res1 == res2 {
		return true
	}

	for i := n; i < m; i++ {
		res2[s2[i]-'a']++
		res2[s2[i-n]-'a']--

		if res1 == res2 {
			return true
		}
	}

	return false
}

var (
	dx = []int{1, 0, 0, -1}
	dy = []int{0, 1, -1, 0}
)

func floodFill(image [][]int, sr int, sc int, newColor int) [][]int {

	currColor := image[sr][sc]
	if currColor == newColor {
		return image
	}

	n, m := len(image), len(image[0])
	queue := [][]int{}

	queue = append(queue, []int{sr, sc})
	image[sr][sc] = newColor

	for i := 0; i < len(queue); i++ {
		cell := queue[i]

		for j := 0; j < 4; j++ {
			mx, my := cell[0]+dx[j], cell[1]+dy[j]
			if mx >= 0 && mx < n && my >= 0 && my < m && image[mx][my] == currColor {
				queue = append(queue, []int{mx, my})
				image[mx][my] = newColor
			}
		}
	}

	return image

	// currColor := image[sr][sc]
	// if currColor == newColor {
	// 	return image
	// }
	// n, m := len(image), len(image[0])
	// queue := [][]int{}
	// queue = append(queue, []int{sr, sc})
	// image[sr][sc] = newColor
	// for i := 0; i < len(queue); i++ {
	// 	cell := queue[i]
	// 	for j := 0; j < 4; j++ {
	// 		mx, my := cell[0]+dx[j], cell[1]+dy[j]
	// 		if mx >= 0 && mx < n && my >= 0 && my < m && image[mx][my] == currColor {
	// 			queue = append(queue, []int{mx, my})
	// 			image[mx][my] = newColor
	// 		}
	// 	}
	// }
	// return image
}

func updateMatrix(matrix [][]int) [][]int {

	n, m := len(matrix), len(matrix[0])
	queue := make([][]int, 0)

	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			if matrix[i][j] == 0 {
				queue = append(queue, []int{i, j})
			} else {
				matrix[i][j] = -1
			}
		}
	}

	direction := [][]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
	for len(queue) > 0 {
		point := queue[0]
		queue = queue[1:]

		for _, v := range direction {
			x := point[0] + v[0]
			y := point[1] + v[1]
			if x >= 0 && x < n && y >= 0 && y < m && matrix[x][y] == -1 {
				matrix[x][y] = matrix[point[0]][point[1]] + 1
				queue = append(queue, []int{x, y})
			}
		}
	}

	return matrix

	// n, m := len(matrix), len(matrix[0])
	// queue := make([][]int, 0)
	// for i := 0; i < n; i++ { // 把0全部存进队列，后面从队列中取出来，判断每个访问过的节点的上下左右，直到所有的节点都被访问过为止。
	// 	for j := 0; j < m; j++ {
	// 		if matrix[i][j] == 0 {
	// 			point := []int{i, j}
	// 			queue = append(queue, point)
	// 		} else {
	// 			matrix[i][j] = -1
	// 		}
	// 	}
	// }
	// direction := [][]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}

	// for len(queue) > 0 { // 这里就是 BFS 模板操作了。
	// 	point := queue[0]
	// 	queue = queue[1:]
	// 	for _, v := range direction {
	// 		x := point[0] + v[0]
	// 		y := point[1] + v[1]
	// 		if x >= 0 && x < n && y >= 0 && y < m && matrix[x][y] == -1 {
	// 			matrix[x][y] = matrix[point[0]][point[1]] + 1
	// 			queue = append(queue, []int{x, y})
	// 		}
	// 	}
	// }

	// return matrix
}

func combine(n int, k int) [][]int {
	var (
		res [][]int
		tmp []int
		dfs func(int)
	)

	dfs = func(i int) {

		if len(tmp)+(n-i+1) < k {
			return
		}

		if len(tmp) == k {
			comb := make([]int, k)
			copy(comb, tmp)
			res = append(res, comb)
			return
		}

		tmp = append(tmp, i)
		dfs(i + 1)
		tmp = tmp[:len(tmp)-1]
		dfs(i + 1)
	}

	dfs(1)

	return res
}

func rob(nums []int) int {
	if len(nums) == 0 {
		return 0
	}

	if len(nums) == 1 {
		return nums[0]
	}

	dp := make([]int, len(nums))
	dp[0] = nums[0]
	dp[1] = max(nums[0], nums[1])

	for i := 2; i < len(nums); i++ {
		dp[i] = max(nums[i-2]+nums[i], dp[i-1])
	}

	return dp[len(nums)-1]
}

func increase(d int) (ret int) {
	defer func() {
		ret++
	}()

	fmt.Println("-------", d, ret)

	return d
}

func f2() (r int) {
	t := 5
	defer func() {
		t = t + 5
		fmt.Println("----", t)
	}()
	return t
}

type Person struct {
	age int
}

func (p Person) getAge() int {
	return p.age
}

func (p Person) addAge() {
	p.age += 1
}

func searchInts(n int, f func(int) bool) int {
	i, j := 0, n

	for i < j {
		h := int(uint(i+j) >> 1)

		if !f(h) {
			i = h + 1
		} else {
			j = h
		}
	}

	return i
}

func searchRevole(nums []int, target int) int {
	left, right := 0, len(nums)-1

	for left <= right {
		middle := (left + right) >> 1
		if nums[middle] == target {
			return middle
		}

		if nums[middle] >= nums[left] {
			if nums[middle] > target && target >= nums[left] {
				right = middle - 1
			} else {
				left = middle + 1
			}
		} else {
			if nums[middle] < target && target <= nums[right] {
				left = middle + 1
			} else {
				right = middle - 1
			}
		}
	}

	return -1
}

func searchMatrix(matrix [][]int, target int) bool {
	row := sort.Search(len(matrix), func(i int) bool {
		return matrix[i][0] > target
	}) - 1

	if row < 0 {
		return false
	}

	col := sort.SearchInts(matrix[row], target)

	return col < len(matrix[row]) && matrix[row][col] == target
}

type Back struct {
	A int
	B int
}

func setA() (*Back, error) {
	return &Back{A: 2}, nil
}

func lengthOfStr(s string) int {
	res := map[byte]int{}
	n := len(s)
	right := -1
	ans := 0

	for i := 0; i < n; i++ {
		if i != 0 {
			delete(res, s[i-1])
		}
		for right+1 < n && res[s[right+1]] == 0 {
			res[s[right+1]]++
			right++
		}

		ans = max(ans, right-i+1)
	}

	return ans
}

func main() {

	fmt.Println(time.Now().Unix())

	// ch := make(chan int, 10)

	// for i := 0; i < 10; i++ {
	// 	ch <- i
	// }
	// close(ch)
	// go func() {
	// 	i := 0
	// 	for i < 15 {
	// 		x := <-ch

	// 		fmt.Println("go routine get ch:", x)

	// 		i++
	// 	}

	// }()

	// ch <- 1

	// time.Sleep(1 * time.Second)
	// fmt.Println("done")

	// fmt.Println(lengthOfStr("abcabcdab"))

	// var res *Back
	// var err error
	// res = &Back{
	// 	A: 1,
	// 	B: 2,
	// }

	// fmt.Println(res)

	// res, err = setA()

	// fmt.Println(res, err)

	// var p Person = Person{age: 10}

	// fmt.Println(p.getAge())

	// p.addAge()
	// fmt.Println(p.getAge())

	// fmt.Println(p)

	// fmt.Println(f2())
	// nums := []int{1, 2, 4, 5, 6, 7, 8}
	// x := 5
	// res := searchInts(len(nums), func(i int) bool { return nums[i] >= x })

	// // res := combine(3, 2)
	// fmt.Println(res)

	// const (
	// 	mutexLocked      = 1 << iota
	// 	mutexWoken       = 1 << 1
	// 	mutexStarving    = 1 << 2
	// 	mutexWaiterShift = iota
	// )

	// // fmt.Println(mutexLocked, mutexWoken, mutexStarving, mutexWaiterShift)

	// // locker := sync.Mutex{}
	// var a int32 = 1

	// locker := sync.Mutex{}

	// locker.Lock()
	// fmt.Println(res, a)

	// fmt.Println(mutexLocked|mutexStarving, a&(mutexLocked|mutexStarving))

	// str := "Let's take LeetCode contest"
	// res := reverseWords(str)
	// fmt.Println(res)

	// res := twoSum([]int{2, 3, 4, 5, 6, 9}, 6)
	// fmt.Println(res)

	// arr := []int{0, 1, 0, 3, 12}
	// moveZeroes(arr)
	// fmt.Println(arr)

	// arr := []int{-1, 0, 3, 5, 9, 12}

	// rotate(arr, 2)
	// fmt.Println(arr)

	// res := sortedSquares(arr)

	// res := searchInsert(arr, 13)
	// fmt.Println(arr)

	// res := search(arr, 2)
	// fmt.Println(res)
	// fmt.Println("----")

	// i := getA()
	// fmt.Println(i)
}
