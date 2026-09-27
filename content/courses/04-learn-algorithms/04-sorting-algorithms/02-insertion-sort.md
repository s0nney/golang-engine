---
title: Insertion Sort
quiz:
  - question: |
      Insertion sort is partway through `[3, 8, 12, 5, 1]` and has sorted the
      first three elements. What does the slice look like after it inserts `5`?
    options:
      - text: '`[3, 5, 8, 12, 1]`'
        correct: true
      - text: '`[1, 3, 5, 8, 12]`'
      - text: '`[3, 8, 5, 12, 1]`'
    explanation: |
      `5` is shifted left past `12` and `8`, stopping after `3`. The part to the
      left of the current position is always sorted; the rest (`1`) hasn't been
      looked at yet.
  - question: When is insertion sort a genuinely good choice?
    options:
      - text: For sorting millions of random values
      - text: For small or nearly-sorted slices
        correct: true
      - text: Never, it's always O(n²)
    explanation: |
      On nearly-sorted data each element only moves a little, so it runs close to
      O(n). It also has very low overhead, which is why real sort
      implementations (including Go's) switch to insertion sort for small
      chunks.
---

**Insertion sort** is how most people sort a hand of playing cards. You pick up
cards one at a time and slide each new one left until it's in the right spot
among the cards you're already holding.

## The idea

Split the slice into a sorted left part and an unsorted right part. Initially,
the sorted part is just the first element. Then, for each next element:

1. Take it out.
2. Shift every bigger element in the sorted part one place to the right.
3. Drop it into the gap.

```
[8 | 3 12 5 1]   take 3, shift 8 right  → [3 8 | 12 5 1]
[3 8 | 12 5 1]   take 12, nothing bigger → [3 8 12 | 5 1]
[3 8 12 | 5 1]   take 5, shift 12 and 8 → [3 5 8 12 | 1]
[3 5 8 12 | 1]   take 1, shift all four  → [1 3 5 8 12]
```

## In Go

```go
package main

import (
	"cmp"
	"fmt"
)

func InsertionSort[T cmp.Ordered](s []T) {
	for i := 1; i < len(s); i++ {
		current := s[i]
		j := i - 1
		for j >= 0 && s[j] > current {
			s[j+1] = s[j] // shift right
			j--
		}
		s[j+1] = current
	}
}

func main() {
	rates := []float64{0.031, 0.007, 0.12, 0.045}
	InsertionSort(rates)
	fmt.Println(rates)
}
```

```
[0.007 0.031 0.045 0.12]
```

Note the order of the conditions in `j >= 0 && s[j] > current`. Go's `&&`
short-circuits, so when `j` reaches `-1` we never evaluate `s[-1]`, which would
panic. Swap them and you've written a crash.

## Complexity

- **Worst case**: reverse-sorted input. Element `i` shifts `i` places, for
  1 + 2 + ... + (n-1) ≈ n²/2 moves. **O(n²)**.
- **Best case**: already sorted. The inner loop never runs. **O(n)**.
- **Space**: **O(1)**, it sorts in place.

## Why it matters

Insertion sort's worst case matches bubble sort's, but it's much better in
practice:

- It does **shifts** (one write each) instead of **swaps** (two writes each).
- On **nearly sorted** data, like Clout's leaderboard after a few follower
  counts change overnight, each element moves only a few places, so it runs
  in close to O(n).
- It's **online**: it can sort items as they arrive, since each new item just
  gets inserted into the sorted part.
- It's **stable**: equal elements keep their original order (more on that at
  the end of this chapter).

That's why serious sort implementations use insertion sort for small pieces.
Go's `slices.Sort` switches to insertion sort for chunks of 12 or fewer
elements, where its low overhead beats the fancy algorithms.
