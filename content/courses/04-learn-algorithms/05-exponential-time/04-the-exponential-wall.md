---
title: The Exponential Wall
quiz:
  - question: Your O(2ⁿ) algorithm handles 40 influencers in an hour. You buy a computer 1,000 times faster. Roughly how many influencers can you now handle in an hour?
    options:
      - text: About 40,000
      - text: About 4,000
      - text: About 50
        correct: true
    explanation: |
      1,000 is about 2¹⁰, so a 1,000× speed-up buys you only about 10 more items.
      With exponential algorithms, faster hardware barely moves the wall.
  - question: |
      How many iterations does this loop run for `n = 20`?

      ```go
      for mask := range 1 << n {
      	// check subset
      }
      ```
    options:
      - text: '20'
      - text: '400'
      - text: '1,048,576'
        correct: true
    explanation: |
      `1 << 20` is 2²⁰ = 1,048,576: one iteration per subset of 20 items. Each
      bit of `mask` says whether one item is in that subset.
---

You've now seen two kinds of runaway algorithm: O(2ⁿ) from exploring every
subset or every branch, and O(n!) from exploring every ordering. Let's see why
they're a **wall** rather than a slope, and what to do when you hit one.

## Every subset

Clout's campaign planner has a budget and a list of influencers, each with a
price and an expected reach. Which **group** of influencers gives the most
reach without going over budget? (This is the famous *knapsack problem*.)

The brute-force approach tries every subset. With `n` influencers, each one is
either in or out, so there are 2ⁿ subsets. A neat trick is to count from 0 to
2ⁿ - 1 and treat each bit of the counter as "in" or "out":

```go
package main

import "fmt"

type Offer struct {
	Handle string
	Price  int
	Reach  int
}

func bestCampaign(offers []Offer, budget int) (best int, picked []string) {
	n := len(offers)
	for mask := range 1 << n {
		price, reach := 0, 0
		for i := range n {
			if mask&(1<<i) != 0 {
				price += offers[i].Price
				reach += offers[i].Reach
			}
		}
		if price <= budget && reach > best {
			best = reach
			picked = picked[:0]
			for i := range n {
				if mask&(1<<i) != 0 {
					picked = append(picked, offers[i].Handle)
				}
			}
		}
	}
	return best, picked
}

func main() {
	offers := []Offer{
		{"ava", 500, 90_000},
		{"bo", 300, 40_000},
		{"cy", 200, 35_000},
		{"dee", 400, 60_000},
	}
	fmt.Println(bestCampaign(offers, 900))
}
```

```
150000 [ava dee]
```

It's correct, and for 4 influencers it checks just 16 subsets. The time is
O(n × 2ⁿ): 2ⁿ subsets, O(n) work to total each one.

## Why hardware can't save you

Say this runs in one second for 25 influencers. Watch what extra items do:

| influencers | subsets | time |
|------------:|--------:|-----:|
| 25 | 33 million | 1 second |
| 35 | 34 billion | 17 minutes |
| 45 | 35 trillion | 12 days |
| 60 | 1.2 × 10¹⁸ | about 1,100 years |

Now suppose you buy a computer 1,000 times faster. Since 1,000 ≈ 2¹⁰, it lets
you handle about **10 more** influencers in the same time. A million times
faster? About 20 more. For a polynomial O(n²) algorithm, a 1,000× speed-up
lets you handle about 31 times as many items. For exponential, it's a rounding
error.

This is the practical meaning of the wall: with exponential algorithms, the
problem size you can handle is basically fixed, no matter what you spend.

## What to do about it

Engineers hit exponential problems all the time. Scheduling, routing, packing,
and choosing campaign line-ups are all in this family. The strategies:

1. **Look for overlapping subproblems.** Memoization and dynamic programming
   turned Fibonacci from O(2ⁿ) into O(n). Knapsack with whole-number prices
   has a dynamic programming solution that's O(n × budget), which is fast when
   the budget is small.
2. **Settle for "good enough".** A *greedy heuristic* such as "keep picking the
   influencer with the best reach per dollar that still fits" runs in
   O(n log n) and is usually close to optimal, though not always optimal.
3. **Prune.** Skip branches that can't possibly beat the best answer so far.
   Still exponential in the worst case, but often much faster in practice.
4. **Keep n small.** If a brand only ever books 15 influencers, 2¹⁵ = 32,768
   subsets is nothing. Exponential is fine when `n` is tiny and guaranteed to
   stay tiny.

## P vs NP, in one paragraph

Problems like knapsack and travelling salesperson belong to a class called
**NP-complete**. Nobody has found a polynomial-time algorithm for any of them,
and nobody has proven one can't exist. Whether one does is the **P vs NP**
question, one of the great unsolved problems in mathematics, with a million-
dollar prize attached. Until someone claims it, when you meet one of these
problems in the wild, reach for the strategies above rather than trying to
find a perfect fast algorithm.
