---
title: Logarithms
quiz:
  - question: What is log₂(64)?
    options:
      - text: '6'
        correct: true
      - text: '8'
      - text: '32'
      - text: '128'
    explanation: |
      A logarithm asks "what power do I raise the base to?" 2 to the power of 6
      is 64, so log₂(64) = 6. Equivalently, you can halve 64 six times before
      reaching 1.
  - question: |
      What does this print?

      ```go
      n, steps := 1000, 0
      for n > 1 {
      	n /= 2
      	steps++
      }
      fmt.Println(steps)
      ```
    options:
      - text: '`500`'
      - text: '`10`'
      - text: '`9`'
        correct: true
      - text: '`1000`'
    explanation: |
      The loop halves `n` using integer division: 1000, 500, 250, 125, 62, 31, 15,
      7, 3, 1. That's 9 halvings. log₂(1000) is about 9.97, and integer
      division rounds each step down, so we get 9. A loop that halves its input
      runs about log₂(n) times.
---

A **logarithm** is the opposite of an exponent. Where an exponent asks "what
do I get if I multiply 2 by itself 6 times?" (64), a logarithm asks "how many
times do I multiply 2 by itself to get 64?" (6).

- 2⁶ = 64, so log₂(64) = 6
- 10³ = 1000, so log₁₀(1000) = 3
- 2¹⁰ = 1024, so log₂(1024) = 10

The little number is the **base**. In computer science we almost always use
base 2, because computers love halving things, and people often just write
`log n` to mean log₂(n).

## Logs as "how many halvings"

Here's the intuition that matters for algorithms: **log₂(n) is how many times
you can cut `n` in half before you reach 1.**

```go
package main

import (
	"fmt"
	"math"
)

func halvings(n int) int {
	steps := 0
	for n > 1 {
		n /= 2
		steps++
	}
	return steps
}

func main() {
	for _, n := range []int{8, 1024, 1_000_000, 8_000_000_000} {
		fmt.Printf("%d -> %d halvings (log2 = %.2f)\n", n, halvings(n), math.Log2(float64(n)))
	}
}
```

```
8 -> 3 halvings (log2 = 3.00)
1024 -> 10 halvings (log2 = 10.00)
1000000 -> 19 halvings (log2 = 19.93)
8000000000 -> 32 halvings (log2 = 32.90)
```

Look at the last line. Eight *billion* (roughly everyone on Earth) halves down
to 1 in just 32 steps. Logarithms grow incredibly slowly.

## Why Clout cares

Imagine Clout's influencers are sorted by follower count, and a brand asks
"is there anyone with exactly 48,210 followers?" You could check all of them.
Or you could look at the middle one: if it's too big, throw away the top half;
if too small, throw away the bottom half. Each look halves the problem.

That's **binary search**, and it finds any account among 8 billion sorted
records in at most about 33 looks. Algorithms whose work is proportional to
log n are among the best you'll ever meet. We'll write binary search in the
Big O chapter.

## Logs in Go

The `math` package has `math.Log2`, `math.Log10` and `math.Log` (natural log,
base *e*). For an exact integer log₂ there's a neat trick in `math/bits`:

```go
package main

import (
	"fmt"
	"math/bits"
)

func main() {
	fmt.Println(bits.Len(1024) - 1)
}
```

`bits.Len(x)` is the number of bits needed to represent `x`, so `bits.Len(x) - 1`
is ⌊log₂ x⌋ for any `x > 0`. This prints `10`.
