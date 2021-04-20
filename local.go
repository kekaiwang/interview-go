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

func main() {
	res := strStr("hello", "lo")
	fmt.Println(res)
}
