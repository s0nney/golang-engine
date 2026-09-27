---
title: O(n²) and Analysing Loops
quiz:
  - question: |
      What's the Big O of this function, where `n` is `len(counts)`?

      ```go
      func f(counts []int) int {
      	total := 0
      	for i := range counts {
      		for j := range 5 {
      			total += counts[i] * j
      		}
      	}
      	return total
      }
      ```
    options:
      - text: O(n²)
      - text: O(n)
        correct: true
      - text: O(5n²)
      - text: O(1)
    explanation: |
      The inner loop always runs exactly 5 times, whatever `n` is. That's `5n`
      steps in total, which simplifies to O(n). Nested loops are only O(n²) when
      *both* loops depend on `n`.
  - question: |
      What's the Big O of this function, where `a` and `b` are two different lists?

      ```go
      func sharedFollowers(a, b []string) int {
      	shared := 0
      	for _, x := range a {
      		for _, y := range b {
      			if x == y {
      				shared++
      			}
      		}
      	}
      	return shared
      }
      ```
    options:
      - text: O(n)
      - text: O(a + b)
      - text: O(a × b), where a and b are the two lengths
        correct: true
    explanation: |
      For each of the `len(a)` items, the inner loop runs `len(b)` times, so the
      total is `len(a) × len(b)` comparisons. When the inputs have independent
      sizes, keep both variables in the Big O.
---

Clout's growth team wants to find **follower overlap**: for a list of
influencers, which pairs have exactly the same follower count? (Identical
counts are a strong sign of bought followers from the same bot farm.)

## Every pair

The straightforward algorithm compares every influencer against every other:

```go
package main

import "fmt"

func suspiciousPairs(counts []int) [][2]int {
	var pairs [][2]int
	for i := range counts {
		for j := i + 1; j < len(counts); j++ {
			if counts[i] == counts[j] {
				pairs = append(pairs, [2]int{i, j})
			}
		}
	}
	return pairs
}

func main() {
	counts := []int{5000, 812, 5000, 91, 812, 5000}
	fmt.Println(suspiciousPairs(counts))
}
```

```
[[0 2] [0 5] [1 4] [2 5]]
```

How many comparisons is that? For `i = 0` the inner loop runs `n-1` times, for
`i = 1` it runs `n-2` times, and so on down to 0. The sum is
`(n-1) + (n-2) + ... + 1 = n(n-1)/2`, which is `n²/2 - n/2`. Drop the constant
and lower term: **O(n²)**.

Starting `j` at `i + 1` halves the work compared with starting at `0`, which
is a nice optimization, but it's still O(n²). Halving is a constant factor.

## Rules for analysing loops

You don't need to sum series every time. These rules cover most code:

**Sequential blocks add.** A loop over `n` followed by another loop over `n` is
O(n + n) = O(n). A loop over `n` followed by a nested pair of loops is
O(n + n²) = O(n²). Keep the biggest.

**Nested loops multiply.** An O(n) loop whose body does O(n) work is O(n × n) =
O(n²). Three nested loops over `n` are O(n³).

**Look at what the loop counts, not how many loops there are.** A nested loop
where the inner one runs a fixed 5 times is still O(n). A single loop that
halves `n` each time is O(log n).

**Different inputs get different letters.** If you loop over `influencers` and,
inside, over `posts`, that's O(i × p), not O(n²).

## How bad is O(n²)?

Quadratic algorithms are fine for small inputs and terrible for big ones. Every
time the input doubles, the work *quadruples*:

| influencers | comparisons (n²/2) |
|------------:|-------------------:|
| 1,000       | 500,000            |
| 10,000      | 50,000,000         |
| 1,000,000   | 500,000,000,000    |

At a million influencers, even at a billion comparisons a second, that's over
8 minutes, for a job the map-based duplicate check from chapter 1 does in a
few milliseconds. When you see nested loops over the same big collection,
ask whether a map, a sort, or a smarter data structure could flatten them.
