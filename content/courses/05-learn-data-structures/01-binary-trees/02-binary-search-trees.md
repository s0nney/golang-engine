---
title: Binary Search Trees
quiz:
  - question: |
      Which of these is a valid binary search tree?

      ```
      A:     50          B:     50          C:     50
            /  \               /  \               /  \
          30    70           60    70           30    70
          /                  /                   \
        20                 20                    55
      ```
    options:
      - text: A only
        correct: true
      - text: B only
      - text: A and C
      - text: All three
    explanation: |
      In B, 60 sits in the *left* subtree of 50 but is bigger than 50. In C, 55 is
      the right child of 30, which is fine for 30, but it's also in the left subtree
      of 50 while being greater than 50. The ordering rule applies to *whole subtrees*,
      not just to a node's direct children. Only A follows it everywhere.
  - question: |
      You search this BST for `65`. Which nodes do you visit, in order?

      ```
              50
             /  \
           30    70
                /  \
              60    80
      ```
    options:
      - text: 50, 30, 70, 60, 80
      - text: 50, 70, 60
        correct: true
      - text: 50, 70, 80
      - text: 50, 70, 60, 80
    explanation: |
      65 > 50, go right to 70. 65 < 70, go left to 60. 65 > 60, go right, and
      that child is `nil`, so 65 isn't in the tree. You never look at 30 or 80.
---

A **binary tree** is a tree where every node has *at most two* children, called
`left` and `right`. A **binary search tree** (BST) adds one ordering rule:

> For every node, everything in its **left** subtree is **smaller**, and everything
> in its **right** subtree is **larger**.

Here's a BST of leaderboard scores:

```
              500
            /     \
         250       800
        /   \     /   \
      100   300 600   950
```

Every score left of 500 is below 500. Every score right of 500 is above it. And the
rule holds again at 250, at 800, and at every other node.

## Why the rule is so useful

Say you want to know whether anyone scored exactly 600. Start at the root:

1. 600 > 500, so it can't be on the left. Go right.
2. 600 < 800, so go left.
3. Found it.

At every step you throw away one entire subtree. It's binary search, but in a
structure you can also insert into and delete from cheaply. A sorted slice gives you
O(log n) search, but inserting into its middle means shifting every later element,
which is O(n).

## Searching in Go

A binary tree node has a value and two child pointers. `nil` means "no child here".

```go
package main

import "fmt"

type Node struct {
	Score       int
	Left, Right *Node
}

func contains(n *Node, score int) bool {
	for n != nil {
		switch {
		case score < n.Score:
			n = n.Left
		case score > n.Score:
			n = n.Right
		default:
			return true
		}
	}
	return false
}

func main() {
	root := &Node{Score: 500,
		Left: &Node{Score: 250,
			Left:  &Node{Score: 100},
			Right: &Node{Score: 300},
		},
		Right: &Node{Score: 800,
			Left:  &Node{Score: 600},
			Right: &Node{Score: 950},
		},
	}
	fmt.Println(contains(root, 600))
	fmt.Println(contains(root, 700))
}
```

Output:

```
true
false
```

Searching for 700 goes 500 → 800 → 600 → right child, which is `nil`. Falling off the
tree means the value isn't there. Each step moves one level down, so a search costs
at most the tree's **height** plus one comparisons.

## The catch

For a nicely *balanced* tree like the one above, the height is about log₂(n), so
search is O(log n). With a million players, that's roughly 20 steps. The shape of
a BST depends on the order you insert things, though, and a bad order can make the
height as large as n. We'll see exactly how in the last lesson of this chapter, and
fix it in the next chapter.

## Duplicates

The rule as written says "smaller" and "larger", so what about equal values? You
have to choose a policy: ignore duplicates, keep a count on the node, or always
send equal values to one side. In this chapter our BST acts like a *set* and
ignores duplicates.
