---
title: Efficiency
quiz:
  - question: |
      Algorithm A takes 2 seconds on 1,000 influencers and 4 seconds on 2,000.
      Algorithm B takes 1 second on 1,000 and 4 seconds on 2,000. Which is likely
      the better choice for 1,000,000 influencers?
    options:
      - text: B, because it's faster on 1,000
      - text: A, because its time grows in proportion to the input while B's grows much faster
        correct: true
      - text: They're equal, since both take 4 seconds on 2,000
    explanation: |
      Doubling the input doubles A's time (linear growth) but quadruples B's
      (quadratic growth). At a million items A takes about 2,000 seconds, while B
      takes about a million seconds, over 11 days. Growth rate beats raw speed on
      small inputs.
  - question: What do we usually measure when comparing the efficiency of algorithms?
    options:
      - text: The number of lines of code
      - text: The exact number of milliseconds on the author's laptop
      - text: How the number of steps (and memory) grows as the input size grows
        correct: true
    explanation: |
      Wall-clock time depends on the machine, the compiler and what else is
      running. Counting how the work *grows* with input size gives a
      machine-independent way to compare algorithms.
---

Once an algorithm is correct, the next question is: **how much does it cost?**
Cost usually means *time* (how many steps it takes) and *space* (how much
memory it uses). Clout's CEO has a simpler way to put it: "Will it still work
when we sign that social network with a billion users?"

## Two correct answers, two very different costs

Here's a ticket: *"Does any influencer in this list share a follower count with
another one? Duplicates usually mean bot farms."*

A first attempt compares every pair:

```go
func hasDuplicatePairs(counts []int) bool {
	for i := range counts {
		for j := i + 1; j < len(counts); j++ {
			if counts[i] == counts[j] {
				return true
			}
		}
	}
	return false
}
```

It's correct. But with `n` counts it makes roughly `n × n / 2` comparisons. For
1,000 influencers that's about 500,000 comparisons, which is nothing. For
1,000,000 it's about 500,000,000,000.

A second attempt remembers what it's already seen in a map:

```go
func hasDuplicateSet(counts []int) bool {
	seen := make(map[int]struct{}, len(counts))
	for _, c := range counts {
		if _, ok := seen[c]; ok {
			return true
		}
		seen[c] = struct{}{}
	}
	return false
}
```

Map lookups take roughly constant time, so this makes about `n` steps. For a
million influencers that's a million steps instead of half a trillion. Same
answer, wildly different cost. The trade-off is **memory**: the map can hold up
to `n` entries, while the pair version uses almost none.

## Growth matters more than speed

Here's the key insight of this whole course: for small inputs, almost any
algorithm is fast enough. What separates good algorithms from bad ones is how
their cost **grows** as the input grows.

| influencers | pairs version | map version |
|------------:|--------------:|------------:|
| 10          | 45            | 10          |
| 1,000       | ~500,000      | 1,000       |
| 1,000,000   | ~500,000,000,000 | 1,000,000 |

A faster CPU makes both columns quicker by a constant factor. It can't close
the gap between them, because the gap itself keeps growing.

## Why not just time it?

You can, and in chapter 3 we'll use Go's benchmarking tools to do exactly that.
But timings depend on your laptop, the Go version, the cache, and whether
Slack was open. To *compare algorithms* we count steps and ask how that count
grows with `n`. The math behind that is next: exponents, logarithms and
factorials.
