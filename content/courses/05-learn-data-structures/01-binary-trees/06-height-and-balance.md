---
title: Height and Balance
quiz:
  - question: |
      You insert the scores `10, 20, 30, 40, 50` into an empty BST, in that order.
      What's the tree's height (counting edges)?
    options:
      - text: '2'
      - text: '3'
      - text: '4'
        correct: true
      - text: '5'
    explanation: |
      Each value is larger than everything before it, so each new node becomes the
      right child of the previous one. The result is a straight line of 5 nodes, which
      has 4 edges from top to bottom.
  - question: What's the worst-case time complexity of searching a plain (unbalanced) BST with n nodes?
    options:
      - text: O(1)
      - text: O(log n)
      - text: O(n)
        correct: true
      - text: O(n log n)
    explanation: |
      A search costs O(h). In the worst case, such as inserting sorted data, the tree
      degenerates into a linked list with h = n - 1, so the search is O(n).
  - question: |
      With the `height` function from this lesson, what does `height(nil)` return, and why?
    options:
      - text: '`0`, because an empty tree has no edges'
      - text: '`-1`, so that a single node (two nil children) gets height `1 + max(-1, -1) = 0`'
        correct: true
      - text: '`1`, because every tree has at least one level'
      - text: It panics, because you can't take the height of `nil`
    explanation: |
      Returning `-1` for an empty tree is the trick that makes the recursive formula
      give 0 for a single leaf, matching the definition "number of edges on the
      longest root-to-leaf path".
---

Every BST operation you've written so far costs O(h), where h is the tree's height.
That number decides whether your leaderboard is snappy or sluggish.

## Measuring height

Height is the number of edges on the longest path from the root down to a leaf. It's
a textbook post-order computation, since you need the children's heights before
you can compute the node's:

```go
func height[T cmp.Ordered](n *node[T]) int {
	if n == nil {
		return -1
	}
	return 1 + max(height(n.left), height(n.right))
}
```

An empty tree has height -1, so a single node has height 0. (Some books count
nodes instead of edges and get 1 more everywhere. Either works if you're consistent.)
`max` is a built-in since Go 1.21, so no import needed.

## Same data, different shapes

Insert the same seven scores in two different orders:

```go
package main

import (
	"cmp"
	"fmt"
)

type node[T cmp.Ordered] struct {
	val         T
	left, right *node[T]
}

func insert[T cmp.Ordered](n *node[T], v T) *node[T] {
	if n == nil {
		return &node[T]{val: v}
	}
	switch {
	case v < n.val:
		n.left = insert(n.left, v)
	case v > n.val:
		n.right = insert(n.right, v)
	}
	return n
}

func height[T cmp.Ordered](n *node[T]) int {
	if n == nil {
		return -1
	}
	return 1 + max(height(n.left), height(n.right))
}

func build(scores ...int) *node[int] {
	var root *node[int]
	for _, s := range scores {
		root = insert(root, s)
	}
	return root
}

func main() {
	balanced := build(400, 200, 600, 100, 300, 500, 700)
	sorted := build(100, 200, 300, 400, 500, 600, 700)
	fmt.Println(height(balanced), height(sorted))

	var big *node[int]
	for i := range 10_000 {
		big = insert(big, i) // scores arrive in increasing order
	}
	fmt.Println(height(big))
}
```

Output:

```
2 6
9999
```

The first order produces a bushy tree of height 2. The sorted order makes every new
node the right child of the last one:

```
100
   \
   200
      \
      300
         \
         ...
```

That's not really a tree any more. It's a **linked list** with an unused `left`
pointer on every node, and searching it is O(n). The 10,000-node version has height
9,999, where a balanced tree would have height 13. (Building it is also O(n²),
which you'll notice if you bump the count up.)

## This happens in real life

Sorted input isn't an edge case. Player IDs are handed out in increasing order.
Timestamps always go up. Data loaded from a database often comes back sorted. A
naive BST fed any of these turns into a linked list.

There's a second danger: recursion depth. Our recursive `insert` goes one call
deeper per level. Go's goroutine stacks grow dynamically, so 10,000 levels is fine,
but a tree with hundreds of millions of levels will eventually hit Go's maximum stack
size and crash. A balanced tree never gets anywhere close.

## Balanced trees

A tree is **balanced** when its height stays O(log n) no matter the insertion order.
You can shuffle data before inserting it, but you often don't have all the data up
front. The real fix is a tree that **rebalances itself** as it goes. There are
several kinds (AVL trees, B-trees, treaps, skip lists as a cousin), and the next
chapter builds one of the most widely used: the red-black tree.
