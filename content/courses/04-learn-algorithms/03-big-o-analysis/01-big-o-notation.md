---
title: Big O Notation
quiz:
  - question: An algorithm takes `3n² + 50n + 1000` steps. What is its Big O?
    options:
      - text: O(3n² + 50n + 1000)
      - text: O(n)
      - text: O(n²)
        correct: true
      - text: O(1000)
    explanation: |
      Keep only the fastest-growing term (n²) and drop its constant factor (3).
      For large `n`, the 50n and 1000 are rounding errors next to n².
  - question: 'What does Big O usually describe?'
    options:
      - text: The exact running time in nanoseconds
      - text: An upper bound on how the cost grows as the input grows, usually in the worst case
        correct: true
      - text: The best-case running time
    explanation: |
      Big O ignores machines and constants. It tells you the *shape* of the
      growth, and when people quote a single Big O for an algorithm they usually
      mean the worst case.
---

In the last chapter we built a table of growth rates. **Big O notation** is
the shorthand for saying which column of that table an algorithm belongs in.

When we say an algorithm is **O(n)** (read "big oh of n" or "order n"), we mean
its cost grows *at most* in proportion to `n`, the size of its input, once `n`
gets large enough. O(n²) means it grows at most in proportion to n², and so on.

## The common classes

From best to worst, these are the ones you'll meet over and over:

| Big O      | name          | Clout example                                  |
|------------|---------------|-----------------------------------------------|
| O(1)       | constant      | read one influencer's follower count by index |
| O(log n)   | logarithmic   | binary-search a sorted list of accounts       |
| O(n)       | linear        | find the smallest account                     |
| O(n log n) | linearithmic  | sort accounts with merge sort                 |
| O(n²)      | quadratic     | compare every pair of accounts                |
| O(2ⁿ)      | exponential   | try every subset of influencers for a campaign |
| O(n!)      | factorial     | try every posting order                       |

## Counting steps

To find an algorithm's Big O, count roughly how many basic steps it does as a
function of `n`, then simplify. Take this function, which reports total
followers and the biggest account:

```go
func stats(counts []int) (total, biggest int) {
	for _, c := range counts { // n times
		total += c
	}
	for _, c := range counts { // n times
		biggest = max(biggest, c)
	}
	return total, biggest
}
```

Two separate loops over `n` items is about `2n` steps, plus a couple for the
setup and return: `2n + 2`. Big O says: that's **O(n)**.

## The two simplification rules

**1. Drop constant factors.** O(2n) is O(n). O(n²/2) is O(n²). Constants
depend on the machine, the compiler and how you count a "step", so they don't
tell you anything about the *shape* of the growth. Doubling the input doubles
the time of an O(n) algorithm whether the constant is 2 or 200.

**2. Drop lower-order terms.** O(n² + n) is O(n²). O(n + log n) is O(n). As `n`
grows, the biggest term swamps the rest. At n = 1,000,000, n² is a trillion
and n is a million, so n is 0.0001% of the total.

Combine them and `3n² + 50n + 1000` becomes **O(n²)**.

## Worst, best and average case

Many algorithms do different amounts of work depending on the input, not just
its size. `slices.Contains(handles, "ava")` stops as soon as it finds `"ava"`:

- **Best case**: `"ava"` is first. One step.
- **Worst case**: `"ava"` is last or missing. `n` steps.
- **Average case**: about `n/2` steps if she's equally likely to be anywhere.

When engineers say "`Contains` is O(n)" they mean the worst case, because
that's the guarantee you can plan around. Strictly speaking, Big O is an upper
bound, and computer scientists have other symbols (Ω for lower bounds, Θ for
tight bounds), but in everyday engineering "Big O" means "the tightest upper
bound on the worst case", and that's how we'll use it.
