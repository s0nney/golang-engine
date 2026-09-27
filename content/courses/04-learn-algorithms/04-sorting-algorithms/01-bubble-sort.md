---
title: Bubble Sort
quiz:
  - question: |
      After the **first** full pass of bubble sort over `[5, 1, 4, 2]`, what does
      the slice look like?
    options:
      - text: '`[1, 2, 4, 5]`'
      - text: '`[1, 4, 2, 5]`'
        correct: true
      - text: '`[1, 5, 4, 2]`'
      - text: '`[5, 4, 2, 1]`'
    explanation: |
      Compare 5,1: swap → `[1 5 4 2]`. Compare 5,4: swap → `[1 4 5 2]`. Compare
      5,2: swap → `[1 4 2 5]`. One pass carries the largest element to the end,
      but the rest isn't sorted yet.
  - question: What's bubble sort's time complexity in the worst case, and in the best case with the early-exit `swapped` check?
    options:
      - text: Worst O(n²), best O(n)
        correct: true
      - text: Worst O(n log n), best O(1)
      - text: Worst O(n²), best O(n²)
    explanation: |
      A reverse-sorted input needs about n²/2 comparisons. But if a pass makes no
      swaps, the slice is sorted and we stop, so an already-sorted slice takes a
      single O(n) pass.
---

Clout's dashboard has a leaderboard: influencers ranked by follower count. It's
time to **sort**. Go's standard library can do this in one line, and we'll get
there, but writing sorts yourself is the best way to learn how algorithms
trade simplicity for speed. First up: the simplest sort of all.

## The idea

**Bubble sort** walks through the slice comparing each pair of neighbours. If
they're in the wrong order, it swaps them. After one full pass, the largest
value has "bubbled up" to the end. Repeat until a pass makes no swaps.

Here's `[5, 1, 4, 2]` going through the first pass:

```
[5 1 4 2]  5 > 1, swap
[1 5 4 2]  5 > 4, swap
[1 4 5 2]  5 > 2, swap
[1 4 2 5]  end of pass: 5 is in its final place
```

The second pass moves 4 into place, and the third finds nothing to swap.

## In Go, generically

Using the `cmp.Ordered` constraint, one implementation sorts `int` follower
counts, `float64` engagement rates and `string` handles:

```go
package main

import (
	"cmp"
	"fmt"
)

func BubbleSort[T cmp.Ordered](s []T) {
	for end := len(s); end > 1; end-- {
		swapped := false
		for i := 1; i < end; i++ {
			if s[i-1] > s[i] {
				s[i-1], s[i] = s[i], s[i-1]
				swapped = true
			}
		}
		if !swapped {
			return
		}
	}
}

func main() {
	followers := []int{5000, 120, 88000, 950, 12}
	BubbleSort(followers)
	fmt.Println(followers)

	handles := []string{"zoe", "ava", "mo"}
	BubbleSort(handles)
	fmt.Println(handles)
}
```

```
[12 120 950 5000 88000]
[ava mo zoe]
```

A few things to notice:

- The function sorts **in place**. It doesn't return anything, because the
  slice shares its backing array with the caller. That's the same convention
  as `slices.Sort`.
- `end` shrinks each pass, because the last elements are already in place.
- Go's tuple assignment `a, b = b, a` swaps without a temporary variable.
- If a pass makes no swaps, the slice is sorted and we return early.

## How fast is it?

In the worst case (a reverse-sorted slice), the first pass does `n-1`
comparisons, the next `n-2`, and so on: about n²/2 in total. That's
**O(n²)** time. It uses **O(1)** extra space since it only swaps in place.

Thanks to the `swapped` flag, the best case (already sorted) is a single pass:
**O(n)**.

Bubble sort is almost never the right choice in production, because even
other O(n²) sorts beat it in practice. But it's a great first algorithm to
reason about, and its "compare neighbours and swap" idea shows up again
in insertion sort.
