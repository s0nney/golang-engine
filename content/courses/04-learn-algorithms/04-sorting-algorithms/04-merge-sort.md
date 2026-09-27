---
title: Merge Sort
quiz:
  - question: Why is merge sort O(n log n)?
    options:
      - text: It compares every element with log n others
      - text: The slice is halved about log n times, and each level of halving does O(n) merging work
        correct: true
      - text: It uses binary search to place each element
    explanation: |
      Splitting in half repeatedly gives about log₂ n levels. At each level, all
      the merges together touch every element once, which is O(n). Multiply:
      O(n log n).
  - question: What's the main downside of this merge sort compared with insertion sort?
    options:
      - text: It's O(n²) in the worst case
      - text: It needs O(n) extra memory for the merged slices
        correct: true
      - text: It isn't stable
    explanation: |
      Merging needs somewhere to put the result, so merge sort allocates O(n)
      extra space. Its time is O(n log n) in every case, and it is stable as long
      as ties are taken from the left half first.
---

The O(n²) sorts all work by moving elements a little at a time. **Merge sort**
takes a completely different approach called **divide and conquer**:

1. **Divide**: split the slice in half.
2. **Conquer**: sort each half (recursively, with merge sort!).
3. **Combine**: merge the two sorted halves into one sorted slice.

The recursion stops at slices of length 0 or 1, which are already sorted.

## Merging is the clever bit

Merging two *sorted* slices is easy and fast. Look at the front of each, take
the smaller, repeat:

```
left  [3 27 38]   right [9 10 82]   result []
take 3            → [3]
take 9            → [3 9]
take 10           → [3 9 10]
take 27           → [3 9 10 27]
take 38           → [3 9 10 27 38]
left is empty, append the rest of right → [3 9 10 27 38 82]
```

Each element is looked at once, so merging is O(n).

## In Go

```go
package main

import (
	"cmp"
	"fmt"
)

func MergeSort[T cmp.Ordered](s []T) []T {
	if len(s) <= 1 {
		return s
	}
	mid := len(s) / 2
	left := MergeSort(s[:mid])
	right := MergeSort(s[mid:])
	return merge(left, right)
}

func merge[T cmp.Ordered](left, right []T) []T {
	result := make([]T, 0, len(left)+len(right))
	i, j := 0, 0
	for i < len(left) && j < len(right) {
		if left[i] <= right[j] {
			result = append(result, left[i])
			i++
		} else {
			result = append(result, right[j])
			j++
		}
	}
	result = append(result, left[i:]...)
	result = append(result, right[j:]...)
	return result
}

func main() {
	followers := []int{38, 27, 43, 3, 9, 82, 10}
	sorted := MergeSort(followers)
	fmt.Println(sorted)
}
```

```
[3 9 10 27 38 43 82]
```

Some details:

- This version **returns a new slice** rather than sorting in place. Don't
  assume `followers` is sorted afterwards. Use the return value.
- `make([]T, 0, len(left)+len(right))` preallocates the exact capacity, so the
  `append` calls never have to grow the slice.
- `left[i] <= right[j]` uses `<=`, not `<`. On a tie it takes from the left
  half first, which keeps equal elements in their original order. That makes
  merge sort **stable**.
- After the loop, one side is empty. Appending both leftovers is fine, since
  appending an empty slice does nothing.

## Complexity

Picture the recursion as a tree. The top level is one slice of `n`. The next
has two slices of `n/2`, then four of `n/4`, and so on down to single
elements. That's log₂ n levels.

At every level, the merges together touch all `n` elements once: O(n) per level.
So the total is **O(n log n)**, and that's true for the best, worst and
average case, because merge sort always splits the same way regardless of
the data.

The cost is **O(n) extra space** for the merged slices (plus O(log n) stack
depth for the recursion). For Clout's 10 million influencers, n log n is about
10 million × 23 ≈ 230 million steps: well under a second, compared with
over a day for the O(n²) sorts.
