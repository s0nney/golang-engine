---
title: O(2ⁿ) and Recursive Fibonacci
quiz:
  - question: |
      Roughly how many calls does this make for `fib(30)`?

      ```go
      func fib(n int) int {
      	if n < 2 {
      		return n
      	}
      	return fib(n-1) + fib(n-2)
      }
      ```
    options:
      - text: '30'
      - text: About 900
      - text: About 2.7 million
        correct: true
    explanation: |
      Each call spawns two more, and the same values get recomputed over and
      over. `fib(30)` makes 2,692,537 calls. The count grows exponentially with
      `n` (roughly by a factor of 1.6 each step), which is bounded by O(2ⁿ).
  - question: An O(2ⁿ) algorithm takes 1 second for n = 40. About how long for n = 50?
    options:
      - text: About 1.25 seconds
      - text: About 10 seconds
      - text: About 17 minutes
        correct: true
    explanation: |
      Adding 10 to `n` multiplies the work by 2¹⁰ = 1024. So 1 second becomes
      about 1024 seconds, roughly 17 minutes. Add 10 more and it's 12 days.
---

Everything so far has been *polynomial*: n, n², n log n. Now we step off a
cliff. An algorithm is **exponential**, O(2ⁿ), when adding **one** item to the
input roughly **doubles** the work.

## Clout's growth forecast

Clout's data team has a toy model of follower growth: an account's followers
this month are last month's plus the month before's. Starting from 0 and 1,
that's the famous **Fibonacci sequence**:

```
0, 1, 1, 2, 3, 5, 8, 13, 21, 34, 55, ...
```

The definition is naturally recursive: `fib(n) = fib(n-1) + fib(n-2)`, with
`fib(0) = 0` and `fib(1) = 1`. And it translates straight into Go:

```go
package main

import "fmt"

var calls int

func fib(n int) int {
	calls++
	if n < 2 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func main() {
	for _, n := range []int{10, 20, 30, 40} {
		calls = 0
		fmt.Printf("fib(%d) = %d in %d calls\n", n, fib(n), calls)
	}
}
```

```
fib(10) = 55 in 177 calls
fib(20) = 6765 in 21891 calls
fib(30) = 832040 in 2692537 calls
fib(40) = 102334155 in 331160281 calls
```

Beautiful code, horrifying numbers. `fib(40)` takes a third of a *billion*
calls, a few hundred milliseconds on a fast laptop. `fib(50)` needs over 120
times as many calls (most of a minute), and `fib(90)` would outlast you.

## Why so many calls?

Draw the call tree for `fib(5)`:

```
                    fib(5)
              /              \
          fib(4)            fib(3)
         /      \           /     \
     fib(3)    fib(2)    fib(2)  fib(1)
     /    \     /   \     /   \
 fib(2) fib(1) f(1) f(0) f(1) f(0)
  /  \
f(1) f(0)
```

Every call branches into two more, so the tree roughly doubles in size with each
level. Worse, it keeps solving the **same subproblems**: `fib(3)` is computed
twice, `fib(2)` three times. For `fib(40)`, `fib(2)` is computed over 60
million times.

Strictly, the growth factor is about 1.618 (the golden ratio!) rather than 2,
so the tight bound is O(1.618ⁿ). But it's still exponential, and O(2ⁿ) is the
usual way to describe it.

## Feeling exponential growth

With O(2ⁿ), each extra item doubles the work. So each extra *ten* items
multiplies it by 1,024:

| n  | 2ⁿ steps          | at a billion steps/sec |
|---:|------------------:|-----------------------:|
| 20 | ~1 million        | 1 ms                   |
| 30 | ~1 billion        | 1 second               |
| 40 | ~1 trillion       | 18 minutes             |
| 50 | ~10¹⁵             | 13 days                |
| 60 | ~10¹⁸             | 36 years               |

Compare O(n²): going from 50 to 60 items takes you from 2,500 steps to 3,600.
Exponential algorithms don't slow down gracefully. They hit a wall.

Also notice the space: recursion depth is only `n`, so naive Fibonacci uses just
O(n) stack space. It isn't memory that kills it, it's the repeated work. And
repeated work is something we can fix.
