---
title: O(log n) and O(n log n)
quiz:
  - question: Binary search needs at most how many comparisons to search 1,000,000 sorted follower counts?
    options:
      - text: About 20
        correct: true
      - text: About 1,000
      - text: About 500,000
      - text: About 1,000,000
    explanation: |
      Each comparison halves the remaining range, and a million can be halved
      about 20 times (2²⁰ ≈ 1,048,576). That's O(log n).
  - question: Why can't you binary-search an unsorted slice?
    options:
      - text: Go's compiler rejects it
      - text: It's too slow on unsorted data
      - text: Throwing away half the slice is only safe when the order tells you which half can't contain the target
        correct: true
    explanation: |
      Binary search decides "the target must be to the left" by comparing with
      the middle. That conclusion only holds if everything to the left is
      smaller. On unsorted data it will confidently return wrong answers.
---

Logarithmic algorithms are the superstars of this course. An **O(log n)**
algorithm does a constant amount of work and then *throws away a fixed
fraction* (usually half) of what's left.

## Binary search

Clout keeps a slice of follower counts sorted in ascending order. A brand asks,
"Is there an influencer with exactly 48,210 followers?" Instead of scanning all
of them, look at the middle one:

- If it's the target, done.
- If it's smaller than the target, the target can only be in the right half.
- If it's bigger, the target can only be in the left half.

Repeat on the remaining half until the range is empty.

```go
package main

import (
	"cmp"
	"fmt"
)

func binarySearch[T cmp.Ordered](sorted []T, target T) (int, bool) {
	lo, hi := 0, len(sorted) // search the half-open range [lo, hi)
	for lo < hi {
		mid := lo + (hi-lo)/2
		switch {
		case sorted[mid] == target:
			return mid, true
		case sorted[mid] < target:
			lo = mid + 1
		default:
			hi = mid
		}
	}
	return lo, false
}

func main() {
	counts := []int{45, 312, 7100, 8200, 48210, 99000, 250000}
	fmt.Println(binarySearch(counts, 48210))
	fmt.Println(binarySearch(counts, 9000))
}
```

```
4 true
4 false
```

When the target is missing, `lo` ends up where it *would* go, which is handy
for inserting. Two details worth noticing:

- `mid := lo + (hi-lo)/2` rather than `(lo+hi)/2`. With huge indices, `lo+hi`
  could overflow. It's a famous bug that sat in Java's standard library for
  nine years.
- Using a half-open range `[lo, hi)` avoids the off-by-one headaches of `hi = mid - 1`.

Each iteration halves the range, so the loop runs at most about log₂(n) + 1
times. **O(log n)**. A billion sorted records? About 30 comparisons.

In real code, use the standard library: `slices.BinarySearch(counts, 48210)`
returns exactly the same `(index, found)` pair, and `slices.BinarySearchFunc`
works with a custom comparison.

## O(n log n)

**O(n log n)** means doing O(log n) work for each of `n` items, or splitting
the input in half log n times and doing O(n) work at each level. It's the
sweet spot for sorting: merge sort, quick sort (on average) and Go's own
`slices.Sort` are all O(n log n), and it's been proven that no sort based
on comparing elements can do better in the worst case.

For example, looking up each of `n` handles in a sorted slice of `n` handles is
`n` binary searches of O(log n) each:

```go
found := 0
for _, h := range queries { // n times
	if _, ok := slices.BinarySearch(sortedHandles, h); ok { // O(log n)
		found++
	}
}
```

That's O(n log n). In practice O(n log n) behaves almost like O(n): at a million
items, log n is only about 20. It's the difference between sorting a million
influencers in a fraction of a second and waiting minutes with an O(n²) sort,
as you'll see in the sorting chapter.
