---
title: Memoization
quiz:
  - question: What is memoization?
    options:
      - text: Writing comments that explain what a function does
      - text: Caching the results of function calls so repeated inputs are answered from the cache instead of recomputed
        correct: true
      - text: Storing a program's memory on disk
    explanation: |
      A memoized function checks a cache (often a map) before doing any work, and
      stores each new result in it. It only pays off when the same inputs come up
      repeatedly, as they do in recursive Fibonacci.
  - question: What's the time complexity of memoized Fibonacci?
    options:
      - text: O(2ⁿ)
      - text: O(n)
        correct: true
      - text: O(n²)
      - text: O(1)
    explanation: |
      Each value `fib(0)` through `fib(n)` is computed once and then looked up in
      O(1). That's `n + 1` real computations, so O(n) time, plus O(n) space for
      the cache.
  - question: |
      What's wrong with this memo?

      ```go
      var memo map[int]int

      func fib(n int) int {
      	if v, ok := memo[n]; ok {
      		return v
      	}
      	// ... compute v ...
      	memo[n] = v
      	return v
      }
      ```
    options:
      - text: Reading from a nil map panics
      - text: Writing to a nil map panics
        correct: true
      - text: Maps can't have `int` keys
    explanation: |
      `memo` is declared but never initialised, so it's `nil`. Reading from a nil
      map is fine (it returns the zero value), but writing to one panics with
      `assignment to entry in nil map`. Use `make(map[int]int)`.
---

Naive Fibonacci is slow because it solves the same subproblems millions of
times. The fix is almost embarrassingly simple: **remember the answers**.

## A memo

**Memoization** means caching a function's results, keyed by its inputs.
Before computing, check the cache. After computing, store the result.

```go
package main

import "fmt"

func fibMemo(n int, memo map[int]int) int {
	if n < 2 {
		return n
	}
	if v, ok := memo[n]; ok {
		return v
	}
	v := fibMemo(n-1, memo) + fibMemo(n-2, memo)
	memo[n] = v
	return v
}

func main() {
	memo := make(map[int]int)
	fmt.Println(fibMemo(40, memo))
	fmt.Println(fibMemo(90, memo))
}
```

```
102334155
2880067194370816120
```

Both answers come back instantly. Earlier, `fib(40)` took a third of a billion
calls. Now each `fib(k)` is computed exactly once, and every repeat is a map
lookup. That's **O(n)** time and **O(n)** space for the memo. The exponential
tree has collapsed into a straight line.

Remember the Go gotcha: `make` the map before using it. A `nil` map can be read
from but **panics on write**.

## Wrapping it up nicely

Passing `memo` around is a bit clunky. With a closure the cache becomes an
implementation detail:

```go
func newFib() func(int) int {
	memo := map[int]int{}
	var fib func(int) int
	fib = func(n int) int {
		if n < 2 {
			return n
		}
		if v, ok := memo[n]; ok {
			return v
		}
		memo[n] = fib(n-1) + fib(n-2)
		return memo[n]
	}
	return fib
}
```

The `var fib func(int) int` line comes first so the closure can refer to itself
recursively.

## Even better: bottom up

Memoization is *top-down*: start from `n`, recurse down, cache on the way back.
Often you can go **bottom-up** instead, building answers from the smallest
subproblem upwards. That's called **dynamic programming**, and for Fibonacci
you only ever need the previous two values:

```go
func fibIter(n int) int {
	a, b := 0, 1
	for range n {
		a, b = b, a+b
	}
	return a
}
```

O(n) time, O(1) space, no recursion, no map. That's the version you'd ship.

## When does memoization help?

Only when a problem has **overlapping subproblems**, meaning the same inputs
recur. Fibonacci is the textbook example. Merge sort, by contrast, never sorts
the same half twice, so caching would just waste memory.

Watch the limits too: `fib(92)` is the largest that fits in an `int64`. From
`fib(93)` the addition silently overflows. Fast algorithms can't save you from
running out of bits!
