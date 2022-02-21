[toc]

从最基础的开始吧

### 常见的一些小知识

#### 异或运算

利用按位异或的性质，可以得到 mid 和相邻的数之间的如下关系，其中 ^ 是按位异或运算符：

- **当 mid 是偶数时，`mid + 1 = mid ^ 1`**
- **当 mid 是奇数时，`mid - 1 = mid ^ 1`**

### 十大排序

- **稳定排序**
冒泡排序（bubble sort） — O(n2)
插入排序 （insertion sort）— O(n2)
归并排序 （merge sort）— O(n log n)

- **非稳定排序**
选择排序 （selection sort）— O(n2)
希尔排序 （shell sort）— O(n log n)
堆排序 （heapsort）— O(n log n)
快速排序 （quicksort）— O(n log n)

### 冒泡排序

冒泡排序就是把小的元素往前调或者把大的元素往后调，比较是相邻的两个元素比较，交换也发生在这两个元素之间。

**稳定排序，时间复杂度：`O(n2)`**

```go
func bubbleSort(arr []int) {
    len := len(arr)

    for i := 0; i < len; i++ {
        flag := false
        for j := 0; j < len-i-1; j++ {
            if arr[j] > arr[j+1] {
                fmt.Println(arr[j], arr[j+1])
                arr[j], arr[j+1] = arr[j+1], arr[j]
                flag = true
            }
        }

        if !flag {
            break
        }
    }

    fmt.Println(arr)
}
```

### 插入排序

**稳定排序，时间复杂度：`O(n2)`**
插入排序是在一个已经有序的小序列的基础上，一次插入一个元素

```go
func insertSort(arr []int) {
    // [9, 5, 4, 1, 7, 5, 0]
    for i := 1; i < len(arr); i++ {
        if arr[i] < arr[i-1] {
            tmp := arr[i]
            j := i - 1

            for j >= 0 && arr[j] > tmp {
                arr[j+1] = arr[j]
                j--
            }
            arr[j+1] = tmp
        }
    }
}
```

### 归并排序

**稳定排序，时间复杂度 O(n log n)**
将一个大的无序数组有序，我们可以把大的数组分成两个，然后对这两个数组分别进行排序，之后在把这两个数组合并成一个有序的数组。由于两个小的数组都是有序的，所以在合并的时候是很快的。

```go
func mergeSort(arr []int) []int {
    if len(arr) <= 1 {
        return arr
    }

    p := len(arr) / 2

    left := mergeSort(arr[:p])
    right := mergeSort(arr[p:])

    return merge(left, right)
}

func merge(left, right []int) []int {
    i, j := 0, 0
    m, n := len(left), len(right)

    var result []int
    for {
        if i >= m || j >= n {
            break
        }

        if left[i] < right[j] {
            result = append(result, left[i])
            i++
        } else {
            result = append(result, right[j])
            j++
        }
    }

    if i != m {
        result = append(result, left[i:]...)
    }

    if j != n {
        result = append(result, right[j:]...)
    }

    return result
}
```

### 选择排序

**非稳定排序， O(n2)**。

选择排序是给每个位置选择当前元素最小的，比如给第一个位置选择最小的，在剩余元素里面给>二个元素选择第二小的，依次类推，直到第n-1个元素，第n个 元素不用选择了，因为只剩下它一个最大的元素了。

```go
func selectSort(arr []int) {
    len := len(arr)

    for i := 0; i < len; i++ {
        min := i
        for j := i + 1; j < len; j++ {
            if arr[min] > arr[j] {
                min = j
            }
        }

        arr[min], arr[i] = arr[i], arr[min]
    }
}
```

### 快速排序

**非稳定排序， O(n log n)**。

选取第一个数为基准，将比基准小的数交换到前面，比基准大的数交换到后面，对左右区间重复第二步，直到各区间只有一个数

```go
func quickSort(arr []int) {
    recursionSort(arr, 0, len(arr)-1)
}

func recursionSort(arr []int, left, right int) {
    if left < right {
        pivot := partition(arr, left, right)

        recursionSort(arr, left, pivot-1)
        recursionSort(arr, pivot+1, right)
    }
}

func partition(arr []int, left, right int) int {

    for left < right {
        // 从后往前走，将比第一个小的移到前面
        for left < right && arr[left] <= arr[right] {
            right--
        }

        if left < right {
            arr[left], arr[right] = arr[right], arr[left]
            left++
        }

        // 从前往后走， 将比第一个大的移到后面
        for left < right && arr[left] <= arr[right] {
            left++
        }

        if left < right {
            arr[left], arr[right] = arr[right], arr[left]
            right--
        }
    }

    return left
}
```

### 希尔排序

**非稳定排序， O(n log n)**。

希尔排序可以说是插入排序的一种变种。无论是插入排序还是冒泡排序，如果数组的最大值刚好是在第一位，要将它挪到正确的位置就需要 n - 1 次移动
也就是说，原数组的一个元素如果距离它正确的位置很远的话，则需要与相邻元素交换很多次才能到达正确的位置，这样是相对比较花时间了。

```go
func shellSort(arr []int) {
    len := len(arr)

    for gap := len / 2; gap > 0; gap /= 2 {
        for i := gap; i < len; i++ {
            j := i
            for {
                if j-gap < 0 || arr[j] >= arr[j-gap] {
                    break
                }
                arr[j], arr[j-gap] = arr[j-gap], arr[j]
                j = j - gap
            }
        }
    }
}
```

### 堆排序

**非稳定排序， O(nlogn)**。

```go
func heapSort(nums []int) []int {
    n := len(nums)

    if n == 0 {
        return nil
    }

    for i := 0; i < n; i++ {
        minHeadp(nums[i:])
    }

    return nums
}

func minHeadp(arr []int) {
    n := len(arr)

    floor := n/2 - 1

    for i := floor; i >= 0; i-- {
        // fmt.Println(i)
        root := i
        left := floor*2 + 1
        right := floor*2 + 2

        if left < n && arr[root] > arr[left] {
            arr[root], arr[left] = arr[left], arr[root]
        }

        if right < n && arr[root] > arr[right] {
            arr[root], arr[right] = arr[right], arr[root]
        }
    }
}
```

### 整数反转 - 7

给你一个 32 位的有符号整数 x ，返回将 x 中的数字部分反转后的结果。

如果反转后整数超过 32 位的有符号整数的范围 [−231,  231 − 1] ，就返回 0。

```go
func reverse(x int) int {
    rev := 0

    for x != 0 {
        if rev < math.MinInt32/10 || rev > math.MaxInt32/10 {
            return 0
        }

        digit := x%10
        x /= 10

        rev = rev*10 + digit
    }

    return rev
}
```

### 三数之和等于0 - 15

给你一个包含 n 个整数的数组 nums，判断 nums 中是否存在三个元素 a，b，c ，使得 a + b + c = 0 ？请你找出所有和为 0 且不重复的三元组。

注意：答案中不可以包含重复的三元组。

```go
func threeSum(nums []int) [][]int {
    n := len(nums)
    sort.Ints(nums)
    ans := make([][]int, 0)
 
    // 枚举 a
    for first := 0; first < n; first++ {
        // 需要和上一次枚举的数不相同
        if first > 0 && nums[first] == nums[first - 1] {
            continue
        }
        // c 对应的指针初始指向数组的最右端
        third := n - 1
        target := -1 * nums[first]
        // 枚举 b
        for second := first + 1; second < n; second++ {
            // 需要和上一次枚举的数不相同
            if second > first + 1 && nums[second] == nums[second - 1] {
                continue
            }
            // 需要保证 b 的指针在 c 的指针的左侧
            for second < third && nums[second] + nums[third] > target {
                third--
            }
            // 如果指针重合，随着 b 后续的增加
            // 就不会有满足 a+b+c=0 并且 b<c 的 c 了，可以退出循环
            if second == third {
                break
            }
            if nums[second] + nums[third] == target {
                ans = append(ans, []int{nums[first], nums[second], nums[third]})
            }
        }
    }
    return ans
}
```
