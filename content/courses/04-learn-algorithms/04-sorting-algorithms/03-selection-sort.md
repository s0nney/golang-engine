---
title: Selection Sort
quiz:
  - question: How many swaps does selection sort make on a slice of `n` elements, at most?
    options:
      - text: About n²/2
      - text: n - 1
        correct: true
      - text: n log n
    explanation: |
      Each pass finds the minimum of the unsorted part and does at most one swap
      to put it in place. There are `n - 1` passes, so at most `n - 1` swaps, even
      though it still does about n²/2 comparisons.
  - question: What's selection sort's time complexity on an already-sorted slice?
    options:
      - text: O(n), because nothing needs swapping
      - text: O(n²), because it still scans the whole unsorted part on every pass
        correct: true
      - text: O(1)
    explanation: |
      Selection sort can't tell the data is already sorted. To be sure it has
      the minimum, each pass must look at every remaining element. Best, worst
      and average cases are all O(n²).
---

You already wrote the core of **selection sort** in chapter 1: find the
minimum. Selection sort just does that over and over.

## The idea

1. Find the smallest element in the whole slice and swap it into position 0.
2. Find the smallest in the rest (positions 1 onward) and swap it into position 1.
3. Keep going until the unsorted part is empty.

```
[64 25 12 22 11]   min of all is 11, swap with 64  → [11 25 12 22 64]
[11 | 25 12 22 64] min of rest is 12, swap with 25 → [11 12 25 22 64]
[11 12 | 25 22 64] min of rest is 22, swap with 25 → [11 12 22 25 64]
[11 12 22 | 25 64] min of rest is 25, already there
```

## In Go

```go
package main

import (
	"cmp"
	"fmt"
)

func SelectionSort[T cmp.Ordered](s []T) {
	for i := range len(s) - 1 {
		minIdx := i
		for j := i + 1; j < len(s); j++ {
			if s[j] < s[minIdx] {
				minIdx = j
			}
		}
		s[i], s[minIdx] = s[minIdx], s[i]
	}
}

func main() {
	followers := []int{64, 25, 12, 22, 11}
	SelectionSort(followers)
	fmt.Println(followers)
}
```

```
[11 12 22 25 64]
```

`for i := range len(s) - 1` uses range-over-int (Go 1.22+). One gotcha: if `s`
is empty, `len(s) - 1` is `-1`, and ranging over a negative number simply runs
zero times, so there's no panic. Handy!

## Complexity

The inner loop does `n-1`, then `n-2`, ... comparisons: about n²/2, so
**O(n²)** time. Unlike bubble and insertion sort, there's no early exit.
It *always* scans everything, so its best case is O(n²) too. Space is **O(1)**.

## So why would anyone use it?

Selection sort makes **at most n − 1 swaps**. Bubble sort can make about n²/2.
If writing to memory is much more expensive than reading (think old flash
storage, or sorting huge structs by value), minimising writes can matter.

It's also **not stable**. That long-distance swap can jump an element over an
equal one. Sort `[{"ava", 5}, {"bo", 5}, {"cy", 1}]` by count and the first
swap sends `ava` to the end, *behind* `bo`, even though they're tied. We'll see
why that matters at the end of this chapter.

## The O(n²) club

Bubble, insertion and selection sort are all O(n²). For Clout's 50-person
campaign lists they're fine. For the full database of 10 million influencers,
n² is 10¹⁴ operations, which is over a day of CPU time. We need something
fundamentally better, and that means *divide and conquer*.
