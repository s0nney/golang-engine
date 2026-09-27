---
title: Find the Minimum
quiz:
  - question: |
      What does this program print?

      ```go
      func findMin(nums []int) int {
      	lowest := 0
      	for _, n := range nums {
      		if n < lowest {
      			lowest = n
      		}
      	}
      	return lowest
      }

      fmt.Println(findMin([]int{40, 12, 99}))
      ```
    options:
      - text: '`12`'
      - text: '`40`'
      - text: '`0`'
        correct: true
      - text: '`99`'
    explanation: |
      `lowest` starts at `0`, and no follower count is smaller than zero, so it
      never changes. Starting from an arbitrary value is a classic bug. Start from
      the first element instead.
  - question: What should `findMin` do with an empty slice?
    options:
      - text: Return `0`, since that's the zero value
      - text: Index `nums[0]` anyway, since Go handles it
      - text: Report that there is no minimum, for example by returning `(0, false)`
        correct: true
    explanation: |
      An empty slice has no minimum. Indexing `nums[0]` would panic, and returning
      `0` silently lies to the caller. Go's idiom is to return a second value (a
      `bool` or an `error`) that says whether the result is valid.
---

Your first Clout ticket: *"Brands want to see the smallest account in each
campaign so they can decide who to cut. Write `findMin`."*

## Designing the algorithm

Before touching code, think about how *you* would do it with a list of numbers
on paper. You'd probably:

1. Look at the first number and remember it as "the smallest so far".
2. Look at each remaining number. If it's smaller than the one you remembered,
   remember it instead.
3. When you run out of numbers, the one you remember is the minimum.

That's an algorithm! Now translate it to Go.

```go
package main

import "fmt"

func findMin(nums []int) (int, bool) {
	if len(nums) == 0 {
		return 0, false
	}
	lowest := nums[0]
	for _, n := range nums[1:] {
		if n < lowest {
			lowest = n
		}
	}
	return lowest, true
}

func main() {
	followers := []int{8200, 312, 99000, 45, 7100}
	if m, ok := findMin(followers); ok {
		fmt.Println("smallest account:", m)
	}
	_, ok := findMin(nil)
	fmt.Println("empty ok?", ok)
}
```

Output:

```
smallest account: 45
empty ok? false
```

## The tricky bits

**Where to start.** It's tempting to write `lowest := 0`. That breaks as soon
as every number is positive, because nothing is ever smaller than zero. Some
people reach for `math.MaxInt` instead, which works, but starting with the first
element is clearer and works for any comparable type.

**Empty input.** `nums[0]` on an empty slice panics with *index out of range*.
Always ask "what happens with zero items?" It's the most common edge case in
algorithm work. Returning `(value, ok)` is the Go way to say "there might not
be an answer".

**Slicing `nums[1:]`.** This doesn't copy anything. It's a new slice header
pointing at the same backing array, so it's cheap. We skip index 0 because we've
already used it.

## Make it generic

Clout doesn't only track integer follower counts. Engagement rates are
`float64`s, and handles are `string`s. With generics and the `cmp.Ordered`
constraint, one function covers them all:

```go
package main

import (
	"cmp"
	"fmt"
)

func findMin[T cmp.Ordered](items []T) (T, bool) {
	var zero T
	if len(items) == 0 {
		return zero, false
	}
	lowest := items[0]
	for _, it := range items[1:] {
		lowest = min(lowest, it)
	}
	return lowest, true
}

func main() {
	fmt.Println(findMin([]float64{0.042, 0.017, 0.09}))
	fmt.Println(findMin([]string{"zed", "ava", "mo"}))
}
```

```
0.017 true
ava true
```

We swapped the `if` for Go's built-in `min`. In real code you'd simply call
`slices.Min`, which panics on an empty slice, but writing it yourself once is
how you learn to think in steps.
