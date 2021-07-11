package main

import "fmt"

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
	res := make([]int, n)
	i, j, k := 0, n-1, n-1

	for i <= j {
		leftMul, rightMul := nums[i]*nums[i], nums[j]*nums[j]

		if leftMul > rightMul {
			res[k] = leftMul
			i++
		} else {
			res[k] = rightMul
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

	for i, v := range b {
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

func main() {

	str := "Let's take LeetCode contest"
	res := reverseWords(str)
	fmt.Println(res)

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
