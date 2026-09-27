---
title: O(n!) and Permutations
quiz:
  - question: How many different posting orders are there for 6 influencers?
    options:
      - text: '36'
      - text: '64'
      - text: '720'
        correct: true
      - text: '46,656'
    explanation: |
      6 choices for the first slot, 5 for the second, and so on:
      6! = 6 × 5 × 4 × 3 × 2 × 1 = 720. (64 is 2⁶, the number of *subsets*;
      46,656 is 6⁶, orders where repeats are allowed.)
  - question: |
      Why does the `permutations` iterator yield `slices.Clone(p)` instead of `p`?
    options:
      - text: '`yield` only accepts newly allocated slices'
      - text: '`p` is swapped in place as the algorithm continues, so a caller who kept `p` would see it change underneath them'
        correct: true
      - text: Cloning makes the algorithm O(n) instead of O(n!)
    explanation: |
      Every yielded permutation would otherwise share one backing array. A caller
      that stored them (say with `slices.Collect`) would end up with a list of
      identical slices. Cloning gives each permutation its own copy.
---

If O(2ⁿ) is a wall, O(n!) is a wall with a moat in front of it. **Factorial
time** algorithms usually come from trying every possible **ordering** of the
input.

## The best posting order

A brand has booked several Clout influencers for a product launch and asks:
"In what order should they post to maximise total reach?" Each ordering scores
differently (audiences overlap, some creators boost the next one), and there's
no shortcut formula. The brute-force answer is to score **every ordering** and
keep the best.

There are n! orderings of n items (n choices for the first slot, n-1 for the
next, ...). Let's generate them.

## Generating permutations

The classic recursive algorithm: for position `k`, try swapping each remaining
item into that slot, recurse on position `k+1`, then swap back. We'll wrap it
in an `iter.Seq`, so callers can `range` over permutations and stop early:

```go
package main

import (
	"fmt"
	"iter"
	"slices"
)

func permutations[T any](items []T) iter.Seq[[]T] {
	return func(yield func([]T) bool) {
		p := slices.Clone(items)
		var permute func(k int) bool
		permute = func(k int) bool {
			if k == len(p) {
				return yield(slices.Clone(p))
			}
			for i := k; i < len(p); i++ {
				p[k], p[i] = p[i], p[k]
				if !permute(k + 1) {
					return false
				}
				p[k], p[i] = p[i], p[k]
			}
			return true
		}
		permute(0)
	}
}

func main() {
	for order := range permutations([]string{"ava", "bo", "cy"}) {
		fmt.Println(order)
	}

	for n := range 11 {
		count := 0
		for range permutations(make([]int, n)) {
			count++
		}
		if n >= 8 {
			fmt.Printf("%d influencers: %d orders\n", n, count)
		}
	}
}
```

```
[ava bo cy]
[ava cy bo]
[bo ava cy]
[bo cy ava]
[cy bo ava]
[cy ava bo]
8 influencers: 40320 orders
9 influencers: 362880 orders
10 influencers: 3628800 orders
```

A few details:

- `p` is a clone, so we never scramble the caller's slice.
- Each permutation is yielded as `slices.Clone(p)`, because `p` keeps being
  swapped. Yield `p` itself and a caller who stored the results would find
  them all pointing at the same, constantly changing array.
- If the loop body `break`s, `yield` returns `false`, and we propagate that
  all the way up the recursion to stop immediately.

## How bad is n!?

Generating each permutation costs O(n) (for the clone), and there are n! of them,
so the whole thing is **O(n × n!)**, which we usually just call O(n!).

| influencers | orderings | at a billion per second |
|------------:|----------:|------------------------:|
| 10 | 3,628,800 | 4 ms |
| 13 | 6.2 billion | 6 seconds |
| 15 | 1.3 trillion | 22 minutes |
| 18 | 6.4 × 10¹⁵ | 74 days |
| 20 | 2.4 × 10¹⁸ | 77 years |
| 25 | 1.6 × 10²⁵ | 490 million years |

Going from 10 influencers to 20 doesn't make the job twice as hard. It makes
it about 670 billion times harder. Factorial grows even faster than
exponential: at n = 20, 2ⁿ is about a million, while n! is 2.4 quintillion.

## A famous example

The **Travelling Salesperson Problem** asks: given a list of cities, what's the
shortest route that visits each one and returns home? Imagine a Clout brand
ambassador touring 20 influencers' cities. The brute-force solution checks
every ordering of cities: O(n!). Smarter exact algorithms get it down to about
O(n² × 2ⁿ), which is still exponential. Nobody knows a polynomial-time
algorithm for it, and most computer scientists believe none exists.
