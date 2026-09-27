---
title: Stability and slices.Sort
quiz:
  - question: |
      `infs` is sorted alphabetically by handle. What's guaranteed after this call?

      ```go
      slices.SortStableFunc(infs, func(a, b Influencer) int {
      	return cmp.Compare(b.Followers, a.Followers)
      })
      ```
    options:
      - text: Influencers are sorted by followers, biggest first, and ties stay in alphabetical order
        correct: true
      - text: Influencers are sorted by followers, smallest first, and ties stay in alphabetical order
      - text: Influencers are sorted by followers, biggest first, and ties are in an unspecified order
    explanation: |
      Comparing `b` to `a` (instead of `a` to `b`) reverses the order, so the
      biggest account comes first. Because the sort is stable, influencers with
      equal follower counts keep the order they had before: alphabetical.
  - question: What algorithm does `slices.Sort` use, and what's its worst case?
    options:
      - text: Bubble sort, O(n²)
      - text: Pattern-defeating quicksort (pdqsort), O(n log n)
        correct: true
      - text: Merge sort, O(n)
      - text: Quick sort, O(n²)
    explanation: |
      Go's `slices.Sort` and `slices.SortFunc` use pdqsort: quick sort with
      smart pivots, insertion sort for small chunks, and a fallback to heapsort
      if the pivots keep going badly. That fallback guarantees O(n log n) even
      in the worst case.
  - question: |
      What does this return?

      ```go
      cmp.Compare(3, 7)
      ```
    options:
      - text: '`-1`'
        correct: true
      - text: '`4`'
      - text: '`true`'
      - text: '`1`'
    explanation: |
      `cmp.Compare(a, b)` returns a negative number if `a < b`, zero if they're
      equal and a positive number if `a > b`. Specifically it returns `-1`, `0` or
      `+1`, which is exactly the shape `slices.SortFunc` expects.
---

You've written five sorts by hand. In real Go code you'll almost always use the
standard library instead, but now you understand what it's doing and what
the trade-offs are.

## What is stability?

A sort is **stable** if elements that compare as *equal* keep their original
relative order. With plain integers you can't tell the difference: one `5` looks
like another. With structs, you can.

Clout's leaderboard is first sorted alphabetically by handle, then by follower
count. Two influencers tie at 5,000 followers:

```
before:            stable by followers:   unstable might give:
ava   5000         cy    1200             cy    1200
bo    9000         ava   5000             dee   5000   ← dee before ava!
cy    1200         dee   5000             ava   5000
dee   5000         bo    9000             bo    9000
```

A stable sort guarantees `ava` stays ahead of `dee`, so ties remain
alphabetical. That lets you sort by several keys by sorting several times,
least important key first.

From this chapter: **insertion sort and merge sort are stable** (as we wrote
them), **bubble sort is stable**, and **selection sort and quick sort are
not**.

## The standard library

The `slices` package has everything you need:

- `slices.Sort(s)` sorts any slice of a `cmp.Ordered` type in ascending order.
- `slices.SortFunc(s, cmpFn)` sorts using your comparison function, which must
  return a negative number, zero or a positive number (like `cmp.Compare`).
- `slices.SortStableFunc(s, cmpFn)` does the same, but is stable.
- `slices.IsSorted` and `slices.BinarySearch` round out the family.

```go
package main

import (
	"cmp"
	"fmt"
	"slices"
)

type Influencer struct {
	Handle    string
	Followers int
}

func main() {
	counts := []int{5000, 12, 88000, 950}
	slices.Sort(counts)
	fmt.Println(counts)

	infs := []Influencer{
		{"ava", 5000}, {"bo", 9000}, {"cy", 1200}, {"dee", 5000},
	}

	// Biggest first, ties broken by handle: one sort, two keys.
	slices.SortFunc(infs, func(a, b Influencer) int {
		return cmp.Or(
			cmp.Compare(b.Followers, a.Followers),
			cmp.Compare(a.Handle, b.Handle),
		)
	})
	fmt.Println(infs)

	// Or rely on stability: the input order (alphabetical) breaks ties.
	slices.SortFunc(infs, func(a, b Influencer) int { return cmp.Compare(a.Handle, b.Handle) })
	slices.SortStableFunc(infs, func(a, b Influencer) int {
		return cmp.Compare(b.Followers, a.Followers)
	})
	fmt.Println(infs)
}
```

```
[12 950 5000 88000]
[{bo 9000} {ava 5000} {dee 5000} {cy 1200}]
[{bo 9000} {ava 5000} {dee 5000} {cy 1200}]
```

Two tricks to remember:

- **Descending order**: swap the arguments, `cmp.Compare(b.X, a.X)`. And don't
  write comparators as `a.X - b.X`: with large values the subtraction can
  overflow and flip the sign. `cmp.Compare` is always safe.
- **Multiple keys**: `cmp.Or` returns its first non-zero argument, so it falls
  through to the next key only on a tie.

## pdqsort under the hood

`slices.Sort` and `slices.SortFunc` use **pattern-defeating quicksort**
(pdqsort), which is everything you learned in this chapter stitched together:

- **Quick sort** for the main work, with carefully chosen pivots.
- **Insertion sort** for chunks of 12 elements or fewer, where it's fastest.
- **Heapsort** as a fallback if the pivots keep coming out badly, which
  caps the worst case at **O(n log n)**.
- Pattern detection, so already-sorted or reversed input takes close to O(n).

It is *not* stable. When you need stability, `slices.SortStableFunc` uses
insertion sort on small blocks and then merges them in place. It uses O(1) extra
memory, at the cost of being somewhat slower (it can do O(n log² n) element
moves in the worst case).

## Which should you use?

Use `slices.Sort` or `slices.SortFunc` by default. Reach for
`slices.SortStableFunc` when ties must keep their order. Writing your own sort
is for learning, or for very specialised cases. And now that you've written
five of them, you'll never look at a one-line `slices.Sort` the same way again.

## Further reading

- [Go by Example: Sorting](https://gobyexample.com/sorting)
- [Go by Example: Sorting by Functions](https://gobyexample.com/sorting-by-functions)
