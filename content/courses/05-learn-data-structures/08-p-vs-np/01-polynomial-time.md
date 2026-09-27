---
title: Polynomial Time and P
quiz:
  - question: Which of these running times is **not** polynomial?
    options:
      - text: O(n³)
      - text: O(n log n)
      - text: O(2ⁿ)
        correct: true
      - text: O(V + E)
    explanation: |
      Polynomial means bounded by n^k for some fixed k. n log n is below n², and V + E
      is linear in the input size. 2ⁿ has n in the *exponent*, which eventually beats
      every polynomial.
  - question: What does it mean for a problem to be in **P**?
    options:
      - text: It can be solved by *some* algorithm in polynomial time
        correct: true
      - text: Every algorithm for it runs in polynomial time
      - text: It can be solved in O(1)
      - text: It can be solved on a parallel computer
    explanation: |
      P is a class of *problems*, not algorithms. Sorting is in P because merge sort
      exists, even though bogosort is hopelessly slow. One fast algorithm is enough.
  - question: |
      A computer does a billion steps per second. Roughly how long do 2ⁿ steps take for n = 60?
    options:
      - text: About a minute
      - text: About a day
      - text: About 36 years
        correct: true
      - text: About a millisecond
    explanation: |
      2⁶⁰ is about 1.15 × 10¹⁸. At 10⁹ steps per second, that's about 1.15 × 10⁹
      seconds, which is roughly 36 years. Add 10 more items and it's over 37,000 years.
---

You've spent two courses making algorithms faster: O(n²) sorts became O(n log n), O(n)
lookups became O(log n) and then O(1). This last chapter asks a bigger question: are
there problems that **no** clever algorithm can make fast? It's one of the deepest open
questions in computer science, with a million-dollar prize attached, and it has very
practical consequences for the code you write.

## Two kinds of slow

Look at how these grow as n doubles:

```go
package main

import (
	"fmt"
	"math"
)

func main() {
	fmt.Printf("%4s %10s %12s %24s\n", "n", "n log n", "n³", "2ⁿ")
	for _, n := range []float64{10, 20, 40, 80} {
		fmt.Printf("%4.0f %10.0f %12.0f %24.0f\n",
			n, n*math.Log2(n), math.Pow(n, 3), math.Pow(2, n))
	}
}
```

Output:

```
   n    n log n           n³                       2ⁿ
  10         33         1000                     1024
  20         86         8000                  1048576
  40        213        64000            1099511627776
  80        506       512000 1208925819614629174706176
```

Doubling n multiplies n³ by 8. That hurts, but a faster computer or a bit of patience
keeps up. Doubling n **squares** 2ⁿ. At n = 80, a billion operations per second would
take about 38 million years. No hardware upgrade saves you.

(`%24.0f` prints `2⁸⁰` exactly here because powers of two fit perfectly in a
`float64`. Most huge floats would print with rounding noise at the end.)

## Polynomial time

An algorithm runs in **polynomial time** if its running time is O(nᵏ) for some constant
k: O(n), O(n log n), O(n²), O(n³), even O(n¹⁰⁰). Anything with n in the exponent, like
O(2ⁿ) or O(n!), is **exponential** (or worse) and isn't polynomial.

Computer scientists draw the line between "tractable" and "intractable" right there.
It's a slightly blunt line (an O(n¹⁰⁰) algorithm is useless in practice), but it
turns out to be a remarkably good one: once a problem has *any* polynomial algorithm,
people usually find a practical one soon after.

## The class P

**P** is the set of all **decision problems** (yes-or-no questions) that some algorithm
can solve in polynomial time. Everything you've built in this course is in P:

| Problem | Algorithm | Time |
|---|---|---|
| Is score 600 in the leaderboard? | BST search | O(log n) |
| Is this username taken? | Hashmap / trie | O(1) / O(L) |
| Can you walk from the Village to the Castle? | BFS | O(V + E) |
| Is there a route of at most 25 minutes? | Dijkstra | O((V + E) log V) |
| Do the quest prerequisites contain a cycle? | DFS | O(V + E) |

Why yes-or-no questions? It keeps the theory clean, and it loses very little: "what's
the shortest route?" is essentially as hard as "is there a route of length at most k?",
since you can ask the second question for different k.

Notice that P is about **problems**, not algorithms. Sorting is in P because merge sort
exists. It doesn't matter that bogosort also exists.

The next lesson introduces problems where nobody knows whether they're in P, even
though checking a proposed answer is easy.
