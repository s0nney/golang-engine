---
title: Space Complexity
quiz:
  - question: |
      What's the extra space complexity of this function?

      ```go
      func reversed(counts []int) []int {
      	out := make([]int, len(counts))
      	for i, c := range counts {
      		out[len(counts)-1-i] = c
      	}
      	return out
      }
      ```
    options:
      - text: O(1)
      - text: O(n)
        correct: true
      - text: O(n²)
    explanation: |
      It allocates a brand-new slice the same length as the input, so the extra
      memory grows linearly with `n`. An in-place reverse that swaps elements
      would use O(1) extra space.
  - question: A recursive function calls itself `n` levels deep before returning. Even if it allocates nothing, what's its space complexity?
    options:
      - text: O(1), since it creates no slices or maps
      - text: O(n), because each pending call keeps a stack frame alive
        correct: true
      - text: O(log n)
    explanation: |
      Every call that hasn't returned yet has a frame on the goroutine's stack
      holding its parameters and locals. `n` nested calls means `n` frames, so
      O(n) space.
---

Big O isn't just about time. **Space complexity** describes how much *extra
memory* an algorithm needs as its input grows. We usually don't count the
input itself, only what the algorithm allocates on top of it (sometimes called
*auxiliary space*).

## O(1) space

An algorithm that only uses a fixed number of variables, however large the
input, is O(1) space. `findMin` keeps one `lowest` variable: O(1). So does this
in-place reverse:

```go
package main

import "fmt"

func reverseInPlace(counts []int) {
	for i, j := 0, len(counts)-1; i < j; i, j = i+1, j-1 {
		counts[i], counts[j] = counts[j], counts[i]
	}
}

func main() {
	counts := []int{10, 20, 30, 40, 50}
	reverseInPlace(counts)
	fmt.Println(counts)
}
```

This prints `[50 40 30 20 10]`. Two index variables, no matter how long the slice.
(The standard library's `slices.Reverse` does exactly this.)

Note that it *mutates* its argument. Because a slice shares its backing array
with the caller, the caller sees the change. That's often exactly what you want
for performance, but make it obvious in the function's name and docs.

## O(n) space

Anything that builds a collection proportional to the input is O(n) space:
copying a slice, building a map of every handle, or collecting results.

```go
func uniqueHandles(handles []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, h := range handles {
		if _, ok := seen[h]; !ok {
			seen[h] = struct{}{}
			out = append(out, h)
		}
	}
	return out
}
```

In the worst case every handle is unique, so both `seen` and `out` hold `n`
entries. O(n) + O(n) = O(n) space.

## The time-space trade-off

Remember the duplicate-detection problem from chapter 1?

| approach        | time  | extra space |
|-----------------|-------|-------------|
| compare every pair | O(n²) | O(1)     |
| map of seen counts | O(n)  | O(n)     |
| sort, then check neighbours | O(n log n) | O(1) if you may sort in place |

None is "best". The map is fastest but costs memory. The pair check needs no
memory but is slow. Sorting sits in between but reorders the caller's data.
Picking between them is the daily work of algorithm design, and Big O is how
you talk about the options.

## The hidden cost of recursion

Function calls use memory too. Each call that hasn't returned yet keeps a
**stack frame** with its parameters and local variables. A recursive function
that goes `n` levels deep uses O(n) stack space, even if it never calls `make`.

```go
func sum(counts []int) int {
	if len(counts) == 0 {
		return 0
	}
	return counts[0] + sum(counts[1:])
}
```

This is O(n) time *and* O(n) space, while a simple loop is O(n) time and O(1)
space. Go's goroutine stacks grow automatically, so you won't hit a stack
overflow at a few thousand levels, but at hundreds of millions you will: the
runtime caps a goroutine's stack at 1 GB by default and crashes with
`goroutine stack exceeds 1000000000-byte limit`. Prefer loops when recursion
doesn't make the code clearer.
