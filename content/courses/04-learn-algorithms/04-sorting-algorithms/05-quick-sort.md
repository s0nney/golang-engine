---
title: Quick Sort
quiz:
  - question: 'With the "always pick the last element" pivot rule, which input triggers quick sort''s O(n²) worst case?'
    options:
      - text: A slice in random order
      - text: A slice that's already sorted
        correct: true
      - text: A slice with all-different values
    explanation: |
      On sorted input the last element is always the largest, so each partition
      puts everything on one side and shrinks the problem by just one element.
      That's n levels of O(n) work: O(n²). Picking a random pivot makes this
      vanishingly unlikely.
  - question: |
      After one call to `partition` on `[9, 2, 7, 4, 5]` (pivot `5`, the last
      element), the slice is `[2, 4, 5, 9, 7]`. What does partitioning guarantee?
    options:
      - text: The whole slice is sorted
      - text: '`5` is in its final position, with smaller values on its left and the rest on its right'
        correct: true
      - text: The smallest element is at index 0 and nothing else is known
    explanation: |
      Partitioning only puts the pivot where it belongs, with smaller values
      before it and bigger-or-equal values after it. `[9, 7]` on the right is
      still out of order, which is why quick sort recurses on both sides.
---

**Quick sort** is the other famous divide-and-conquer sort. Where merge sort
does the hard work *after* its recursive calls (merging), quick sort does it
*before* them (partitioning).

## The idea

1. Pick an element to be the **pivot**.
2. **Partition**: rearrange the slice so everything smaller than the pivot is
   on its left and everything else is on its right. The pivot is now in its
   final sorted position.
3. Recursively quick sort the left part and the right part.

No merging needed. Once both sides are sorted, the whole slice is.

## Partitioning in place

We'll use the *Lomuto* partition scheme: the last element is the pivot, and a
boundary `i` marks where the next smaller-than-pivot element should go.

```
pivot = 5          [9 2 7 4 5]  i=0
j=0: 9 < 5? no
j=1: 2 < 5? yes → swap s[0], s[1]  [2 9 7 4 5]  i=1
j=2: 7 < 5? no
j=3: 4 < 5? yes → swap s[1], s[3]  [2 4 7 9 5]  i=2
put the pivot at i → swap s[2], s[4]  [2 4 5 9 7]
```

`5` is now exactly where it'll be in the final sorted slice. `[2 4]` and
`[9 7]` get sorted by the recursive calls.

## In Go

```go
package main

import (
	"cmp"
	"fmt"
)

func QuickSort[T cmp.Ordered](s []T) {
	if len(s) <= 1 {
		return
	}
	p := partition(s)
	QuickSort(s[:p])
	QuickSort(s[p+1:])
}

func partition[T cmp.Ordered](s []T) int {
	last := len(s) - 1
	pivot := s[last]
	i := 0
	for j := range last {
		if s[j] < pivot {
			s[i], s[j] = s[j], s[i]
			i++
		}
	}
	s[i], s[last] = s[last], s[i]
	return i
}

func main() {
	followers := []int{9, 2, 7, 4, 5, 88, 1, 30}
	QuickSort(followers)
	fmt.Println(followers)
}
```

```
[1 2 4 5 7 9 30 88]
```

Notice how neatly slices fit here: `s[:p]` and `s[p+1:]` are views onto the
same backing array, so the recursive calls sort the original data **in place**
with no copying at all.

## Complexity

When the pivot splits the slice roughly in half, there are about log n levels
of recursion with O(n) partitioning work per level: **O(n log n)**. That's the
average case for random data.

But if the pivot is always the smallest or largest element, each partition
only peels off one element. That's n levels of O(n) work: **O(n²)**. With the
"last element" rule, that happens on data that is *already sorted*, which is
depressingly common in real life (think of Clout's leaderboard, re-sorted every
hour).

The fix is to choose the pivot more cleverly. A random pivot makes the worst
case astronomically unlikely:

```go
func partition[T cmp.Ordered](s []T) int {
	last := len(s) - 1
	r := rand.IntN(len(s)) // from math/rand/v2
	s[r], s[last] = s[last], s[r]
	// ... the rest is unchanged
}
```

Space is **O(log n)** on average for the recursion stack, with no extra slices.

## Merge sort vs quick sort

| | merge sort | quick sort |
|---|---|---|
| average time | O(n log n) | O(n log n) |
| worst time | O(n log n) | O(n²) (rare with a good pivot) |
| extra space | O(n) | O(log n) |
| stable? | yes | no |

In practice quick sort is often faster because it works in place and is kind to
the CPU cache. That's why Go's own sort is built on a quick sort variant, as
you're about to see.
