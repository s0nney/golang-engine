---
title: Exponents
quiz:
  - question: |
      What does this print?

      ```go
      fmt.Println(2 ^ 3)
      ```
    options:
      - text: '`8`'
      - text: '`6`'
      - text: '`1`'
        correct: true
      - text: It doesn't compile
    explanation: |
      In Go `^` is bitwise XOR, not "to the power of". `2` is `010` and `3` is
      `011`, so XOR gives `001`, which is `1`. Use `math.Pow` or a loop for powers,
      or `1 << n` for powers of two.
  - question: A post is shared by 3 people, each of whom shares it with 3 more, and so on. How many people see it at the 5th level of sharing?
    options:
      - text: '15'
      - text: '125'
      - text: '243'
        correct: true
    explanation: |
      Each level multiplies by 3, so level 5 is 3 to the power of 5:
      3 × 3 × 3 × 3 × 3 = 243. Repeated multiplication is exactly what an
      exponent is.
---

Algorithms are measured by how they *grow*, and the language of growth is
math. Don't panic: you only need three ideas, and the first is one you already
know.

## Repeated multiplication

An **exponent** says how many times to multiply a number (the *base*) by itself:

- 2³ = 2 × 2 × 2 = 8
- 10² = 10 × 10 = 100
- 5¹ = 5
- any number to the power 0 is 1

In plain text we often write 2³ as `2^3`. In math notation that's fine, but
**in Go `^` means XOR**, so `2 ^ 3` is `1`. That's a classic Go gotcha. Go has
no power operator at all.

## Powers in Go

For floating-point powers, use `math.Pow`. For integer powers, a loop is exact
and avoids float rounding. For powers of two, a bit shift is the fastest trick
there is: `1 << n` is 2ⁿ.

```go
package main

import (
	"fmt"
	"math"
)

func pow(base, exp int) int {
	result := 1
	for range exp {
		result *= base
	}
	return result
}

func main() {
	fmt.Println(pow(3, 5))
	fmt.Println(math.Pow(1.05, 10))
	fmt.Println(1 << 10)
}
```

```
243
1.6288946267774416
1024
```

## Going viral

Exponents describe things that *multiply* at each step. Say a Clout
influencer's post gets shared by 3 people, each of whom shares it with 3 more:

| level | new viewers |
|------:|------------:|
| 1     | 3           |
| 2     | 9           |
| 3     | 27          |
| 5     | 243         |
| 10    | 59,049      |
| 20    | 3,486,784,401 |

By level 20 that's more people than own smartphones. This is **exponential
growth**, and it's brilliant for a viral post and catastrophic for an
algorithm. If an algorithm's step count doubles every time you add *one* item
to the input, it's 2ⁿ, and n = 64 already means more steps than a computer could
finish in your lifetime.

## Watch for overflow

Exponential numbers get big fast, and Go's integers have fixed sizes. An `int`
is 64 bits on modern machines, so its maximum is 2⁶³ − 1 (about 9.2 × 10¹⁸).
Go doesn't panic on integer overflow: it silently wraps around.

```go
package main

import "fmt"

func main() {
	x := 1 << 62
	fmt.Println(x)
	x *= 2
	fmt.Println(x)
}
```

```
4611686018427387904
-9223372036854775808
```

Doubling a positive number gave a negative one. Remember that when your
algorithm starts computing powers.
