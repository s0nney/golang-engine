---
title: Growth Rates
quiz:
  - question: 'Which of these grows fastest as `n` gets large?'
    options:
      - text: n²
      - text: 2ⁿ
      - text: n!
        correct: true
      - text: n log n
    explanation: |
      For large `n` the order from slowest to fastest is log n, n, n log n, n²,
      2ⁿ, n!. At n = 20, n² is 400, 2ⁿ is about a million, and n! is about 2.4
      quintillion.
  - question: At `n = 1,000,000`, roughly how big is log₂ n?
    options:
      - text: About 20
        correct: true
      - text: About 1,000
      - text: About 500,000
    explanation: |
      2²⁰ is 1,048,576, so log₂ of a million is just under 20. Logarithmic growth
      is so slow that it's nearly constant in practice.
---

You now have three pieces of math: exponents, logarithms and factorials. Add
plain old multiplication (n, n²) and you can describe how nearly every
algorithm in this course grows.

## The lineup

Here's how much "work" each growth rate implies for a few input sizes. Think of
each number as the steps an algorithm needs to process `n` Clout influencers.

| n     | log₂ n | n     | n log₂ n | n²        | 2ⁿ         | n!          |
|------:|-------:|------:|---------:|----------:|-----------:|------------:|
| 10    | 3      | 10    | 33       | 100       | 1,024      | 3,628,800   |
| 20    | 4      | 20    | 86       | 400       | 1,048,576  | 2.4 × 10¹⁸  |
| 100   | 7      | 100   | 664      | 10,000    | 1.3 × 10³⁰ | 9.3 × 10¹⁵⁷ |
| 1,000 | 10     | 1,000 | 9,966    | 1,000,000 | 1.1 × 10³⁰¹ | way too big |

A modern CPU does very roughly a billion simple steps per second. So:

- **log n, n, n log n**: a million influencers? Easy. A billion? Still fine.
- **n²**: fine for a few thousand, painful for a million (a trillion steps
  is over 15 minutes of pure CPU, even at best).
- **2ⁿ and n!**: hopeless beyond a few dozen items, no matter how fast your
  computer is.

## Print your own table

Let's have Go compute it. Using `float64` avoids overflow for the big ones:

```go
package main

import (
	"fmt"
	"math"
)

func main() {
	fmt.Printf("%6s %8s %10s %12s %10s\n", "n", "log n", "n log n", "n^2", "2^n")
	for _, n := range []float64{8, 16, 32, 64} {
		fmt.Printf("%6.0f %8.0f %10.0f %12.0f %10.2g\n",
			n, math.Log2(n), n*math.Log2(n), n*n, math.Pow(2, n))
	}
}
```

```
     n    log n    n log n          n^2        2^n
     8        3         24           64    2.6e+02
    16        4         64          256    6.6e+04
    32        5        160         1024    4.3e+09
    64        6        384         4096    1.8e+19
```

Every time `n` doubles, log n goes up by just 1, n² quadruples, and 2ⁿ
*squares*. That last column is the one to be scared of.

## Why constants don't matter (much)

Suppose one algorithm takes `100n` steps and another takes `n²` steps. For
n = 10 the quadratic one wins (100 vs 1,000). But at n = 100 they tie, and at
n = 1,000,000 the linear one is 10,000 times faster. Eventually, the *shape*
of the growth always beats the constant in front of it.

That idea is the foundation of **Big O notation**, which is the next chapter:
a shorthand for "which column of this table does my algorithm live in?"
