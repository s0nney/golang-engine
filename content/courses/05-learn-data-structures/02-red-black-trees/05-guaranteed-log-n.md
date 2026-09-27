---
title: Why O(log n) Is Guaranteed
quiz:
  - question: A red-black tree's root has black-height 5 (not counting nil leaves). What's the fewest nodes the tree can have?
    options:
      - text: '5'
      - text: '10'
      - text: '31'
        correct: true
      - text: '63'
    explanation: |
      A subtree whose black-height is b has at least 2^b − 1 nodes. That minimum is the
      all-black perfect tree of height b − 1. With b = 5, that's 2⁵ − 1 = 31.
  - question: Why does the guarantee "height ≤ 2·log₂(n + 1)" matter more than "height is usually small"?
    options:
      - text: It doesn't; average case is all that matters in practice
      - text: It means no input order, including sorted or adversarial data, can make operations slow
        correct: true
      - text: It makes the tree use less memory
      - text: It makes inserts O(1)
    explanation: |
      A plain BST is fast on *random* input but degrades to O(n) on sorted input, which
      is common in the real world, and an attacker can feed you bad input on purpose.
      A worst-case bound means every search, insert and delete is O(log n), whatever happens.
  - question: |
      The leaderboard needs "all players with scores between 1,000 and 2,000, in order".
      Why is a balanced BST a better fit than a Go `map[int]string`?
    options:
      - text: Maps can't store integers as keys
      - text: A BST keeps keys sorted, so you can find the range start in O(log n) and walk forward; a map has no order and you'd have to scan every entry
        correct: true
      - text: Maps are always slower than trees
      - text: A BST uses less memory than a map
    explanation: |
      Maps give O(1) average lookups by exact key, but their iteration order is random.
      Ordered structures (balanced trees, or a sorted slice if updates are rare) make
      range queries, min/max and "next larger" cheap.
---

We've claimed a red-black tree's height is O(log n). Let's see why that's a
guarantee rather than a hope.

## The proof, in three steps

Let **bh(x)** be the black-height of node x: the number of black nodes on any path
from x down to (but not including) a nil leaf. Rule 5 says this is the same for
every path, so it's well defined.

**Step 1: a subtree with black-height b has at least 2ᵇ − 1 nodes.**
Think of the smallest tree that can have black-height b: every node black, and
every path the same length b. That's a perfect binary tree with b levels, which has
2ᵇ − 1 nodes. Adding red nodes only adds more nodes. (The formal version is a short
induction on the height.)

**Step 2: the root's black-height is at least h / 2.**
Take the longest path from the root, which has h nodes below the root. Rule 4
says no two reds in a row, so at least half of them are black. So bh(root) ≥ h/2.

**Step 3: combine them.**

```
n ≥ 2^(h/2) − 1
n + 1 ≥ 2^(h/2)
log₂(n + 1) ≥ h/2
h ≤ 2·log₂(n + 1)
```

So the height is O(log n), and search, insert and delete, which each walk a path and
do O(1) work per level, are all **O(log n) in the worst case**. No input order can
break it.

## What that bound looks like

```go
package main

import (
	"fmt"
	"math"
)

func main() {
	for _, n := range []int{1_000, 1_000_000, 1_000_000_000} {
		perfect := math.Ceil(math.Log2(float64(n + 1)))
		worstRB := math.Floor(2 * math.Log2(float64(n+1)))
		fmt.Printf("n=%-13d perfect≈%2.0f  red-black≤%2.0f  unbalanced≤%d\n",
			n, perfect, worstRB, n-1)
	}
}
```

Output:

```
n=1000          perfect≈10  red-black≤19  unbalanced≤999
n=1000000       perfect≈20  red-black≤39  unbalanced≤999999
n=1000000000    perfect≈30  red-black≤59  unbalanced≤999999999
```

(These columns count *levels* for the perfect tree and *edges* for the others, which
is why the numbers are off by one here and there. Only the growth rate matters.)
A billion players, and every lookup takes at most about 60 steps.

## Red-black vs other balanced trees

- **AVL trees** keep the two subtrees' heights within 1 of each other. They're more
  rigidly balanced, so lookups are slightly faster, but inserts and deletes do more
  rotations. Red-black trees make the opposite trade.
- **B-trees** store many keys per node, which suits disks and databases (fewer, larger
  reads). Most SQL database indexes are B-trees. A red-black tree is in fact equivalent
  to a B-tree with 2 to 4 children per node, if you glue each black node together with
  its red children.

## When to reach for which structure in Go

Go's standard library has no balanced tree, and that's on purpose: most of the time
something simpler wins.

- **Exact lookups only?** Use a `map`. O(1) on average (chapter 3).
- **Sorted data that rarely changes?** Use a sorted slice with `slices.BinarySearch`.
  It's O(log n) to search, cache-friendly, and has no pointers to chase.
- **Sorted data with lots of inserts and deletes, plus range queries, min/max or
  "next score above mine"?** That's when a balanced tree earns its keep. In real
  projects you'd reach for a well-tested third-party package, but now you know
  exactly what's inside it.
