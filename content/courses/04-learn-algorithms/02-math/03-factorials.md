---
title: Factorials
quiz:
  - question: What is 5! (5 factorial)?
    options:
      - text: '25'
      - text: '120'
        correct: true
      - text: '15'
      - text: '3125'
    explanation: |
      5! = 5 × 4 × 3 × 2 × 1 = 120. It's the number of different orders you can
      put 5 distinct things in.
  - question: |
      A brand wants to try every possible posting order for 21 influencers. Why
      does computing `factorial(21)` with a 64-bit `int` give a wrong answer?
    options:
      - text: Go's `int` can't hold negative numbers
      - text: Recursion in Go is limited to 20 levels
      - text: 21! is larger than the biggest `int64`, so the multiplication silently overflows
        correct: true
    explanation: |
      20! is about 2.4 × 10¹⁸, just under the `int64` limit of about 9.2 × 10¹⁸.
      21! is about 5.1 × 10¹⁹, which doesn't fit, and Go integer overflow wraps
      around silently instead of panicking.
---

The **factorial** of a number `n`, written `n!`, is the product of every whole
number from `n` down to 1:

- 3! = 3 × 2 × 1 = 6
- 5! = 5 × 4 × 3 × 2 × 1 = 120
- 0! = 1 (by definition, and it makes the math work out)

## Counting orderings

Factorials show up whenever you count **orderings**. Suppose a brand hires
three Clout influencers (Ava, Bo and Cy) and asks "in what order should they
post?" The first slot has 3 choices, the second has 2 left, the last has 1:
3 × 2 × 1 = 6 orders.

```
Ava Bo Cy    Bo Ava Cy    Cy Ava Bo
Ava Cy Bo    Bo Cy Ava    Cy Bo Ava
```

With 10 influencers there are 10! = 3,628,800 orders. With 20 there are about
2.4 quintillion. An algorithm that tries every ordering is doomed very quickly,
and we'll meet exactly that kind of algorithm in the exponential-time chapter.

## Factorials in Go

A loop is the simplest version:

```go
package main

import "fmt"

func factorial(n int) int {
	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result
}

func main() {
	for _, n := range []int{0, 3, 5, 10, 20, 21} {
		fmt.Printf("%d! = %d\n", n, factorial(n))
	}
}
```

```
0! = 1
3! = 6
5! = 120
10! = 3628800
20! = 2432902008176640000
21! = -4249290049419214848
```

The last line is garbage. 21! is about 5.1 × 10¹⁹, too big for a 64-bit `int`,
so it overflowed and wrapped around to a negative number without any error. When
you really need huge values, the `math/big` package has arbitrary-precision
integers:

```go
package main

import (
	"fmt"
	"math/big"
)

func main() {
	f := new(big.Int).MulRange(1, 25)
	fmt.Println(f)
}
```

`MulRange(1, 25)` multiplies every integer from 1 to 25, so this prints
`15511210043330985984000000`.

## Recursive definition

Factorials also have a neat recursive definition: `n! = n × (n-1)!`, with
`0! = 1` as the base case.

```go
func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * factorial(n-1)
}
```

It reads just like the math. Recursion will matter a lot in the sorting
and exponential-time chapters, where merge sort, quick sort and Fibonacci all
call themselves.
