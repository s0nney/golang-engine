---
title: The Red-Black Rules
quiz:
  - question: Which of these is **not** one of the red-black tree rules?
    options:
      - text: The root is black
      - text: A red node can't have a red child
      - text: Every path from a node down to a nil leaf has the same number of black nodes
      - text: A black node can't have a black child
        correct: true
    explanation: |
      Black nodes can have black children; a perfectly balanced tree where every
      node is black is a valid red-black tree. The restriction is only on *red*
      nodes, which can't be stacked on top of each other.
  - question: |
      Is this a valid red-black tree? (`B` = black, `R` = red, nil leaves not drawn.)

      ```
              20B
             /   \
           10B    30R
           /
         5R
      ```
    options:
      - text: Yes
      - text: No, because 5 and 30 are both red
      - text: No, because the paths down to nil leaves don't all pass through the same number of black nodes
        correct: true
      - text: No, because 10 has only one child
    explanation: |
      Count the black nodes on each path from the root to a nil leaf, including the
      nil. 20 → 10 → 5 → nil passes 20, 10 and the nil: 3. But 20 → 30 → nil passes
      only 20 and the nil: 2. That breaks rule 5. Two red nodes are fine as long as
      neither is the other's child, and nodes may have just one child.
  - question: What does the black-height rule guarantee about the longest root-to-leaf path?
    options:
      - text: It's at most twice as long as the shortest one
        correct: true
      - text: It's exactly the same length as the shortest one
      - text: It's at most one node longer than the shortest one
      - text: Nothing; only the red rule limits path length
    explanation: |
      The shortest possible path is all black. The longest alternates black and red,
      because red can't follow red, and it has the same number of blacks. So it's at
      most twice as long. Both rules working together give that guarantee.
---

Last chapter ended on a sour note: feed sorted scores into a BST and you get a linked
list. A **self-balancing** BST fixes this by quietly restructuring itself after
every insert or delete so the height stays O(log n).

The **red-black tree** is the most widely deployed self-balancing tree. It's
behind Java's `TreeMap`, C++'s `std::map`, and the Linux kernel's process scheduler
and memory maps. Go's standard library doesn't ship one (it favours maps and sorted
slices), which is a great excuse for us to build one.

## The five rules

A red-black tree is a BST where every node also has a **colour**, red or black,
and the colours must obey these rules:

1. Every node is either red or black.
2. The **root is black**.
3. Every nil child pointer counts as a **black leaf**.
4. A **red node never has a red child**. (No two reds in a row.)
5. For every node, **every path from it down to a nil leaf passes through the same
   number of black nodes.** This count is the node's *black-height*.

```
              40B
            /     \
         20R       60B
        /   \      /
      10B   30B  50R
```

Check rule 5 from the root: 40 → 20 → 10 → nil passes three blacks (40, 10, the nil),
and so do 40 → 60 → nil and 40 → 60 → 50 → nil. Red nodes are "free": they don't count.

## Why these rules balance the tree

Rule 5 says every root-to-leaf path has the same number of black nodes, call it b.
Rule 4 says reds can't be adjacent, so at most every other node on a path is red.
So the shortest path is b nodes (all black) and the longest is about 2b (alternating).
**No path is more than twice as long as any other.** A lopsided, list-like tree is
simply impossible. In the last lesson of this chapter we'll turn that into the
height bound h ≤ 2·log₂(n + 1).

## Checking the rules in Go

We'll store the colour as a `bool`. A nil pointer counts as black (rule 3), so a
tiny helper saves a lot of nil checks later:

```go
type rbNode[T cmp.Ordered] struct {
	val                 T
	red                 bool
	left, right, parent *rbNode[T]
}

func isRed[T cmp.Ordered](n *rbNode[T]) bool { return n != nil && n.red }
```

The `parent` pointer will be needed for rotations. A post-order walk can verify rules
4 and 5 at once by returning each subtree's black-height:

```go
package main

import (
	"cmp"
	"fmt"
)

type rbNode[T cmp.Ordered] struct {
	val                 T
	red                 bool
	left, right, parent *rbNode[T]
}

func isRed[T cmp.Ordered](n *rbNode[T]) bool { return n != nil && n.red }

// blackHeight returns the black-height of n, or an error if a rule is broken.
func blackHeight[T cmp.Ordered](n *rbNode[T]) (int, error) {
	if n == nil {
		return 1, nil // nil leaves are black
	}
	if n.red && (isRed(n.left) || isRed(n.right)) {
		return 0, fmt.Errorf("red node %v has a red child", n.val)
	}
	lh, err := blackHeight(n.left)
	if err != nil {
		return 0, err
	}
	rh, err := blackHeight(n.right)
	if err != nil {
		return 0, err
	}
	if lh != rh {
		return 0, fmt.Errorf("black-heights differ under %v", n.val)
	}
	if !n.red {
		lh++
	}
	return lh, nil
}

func main() {
	good := &rbNode[int]{val: 40,
		left:  &rbNode[int]{val: 20, red: true, left: &rbNode[int]{val: 10}, right: &rbNode[int]{val: 30}},
		right: &rbNode[int]{val: 60, left: &rbNode[int]{val: 50, red: true}},
	}
	fmt.Println(blackHeight(good))

	bad := &rbNode[int]{val: 40, right: &rbNode[int]{val: 60, red: true,
		right: &rbNode[int]{val: 70, red: true}}}
	fmt.Println(blackHeight(bad))
}
```

Output:

```
3 <nil>
0 red node 60 has a red child
```

(This hand-built tree skips the `parent` pointers because the checker doesn't use
them.) Notice `fmt.Println(blackHeight(good))`: when a function returns several
values, you can pass them straight into a variadic call like `Println`.

Next up: the one move a red-black tree uses to reshape itself, the **rotation**.
