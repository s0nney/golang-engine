---
title: Min, Max and Delete
quiz:
  - question: Where is the smallest value in a non-empty BST?
    options:
      - text: At the root
      - text: In the leftmost node, found by following `left` pointers until the next one is `nil`
        correct: true
      - text: In the leftmost *leaf*, which might require going right at some point
      - text: Anywhere; you must visit every node to find it
    explanation: |
      Everything smaller than a node lives to its left, so you keep going left. The
      node where `left` is `nil` is the minimum. It doesn't have to be a leaf: it
      can still have a right child.
  - question: |
      You delete `50` from this tree using the in-order successor. Which value
      ends up at the root?

      ```
              50
             /  \
           30    70
                /  \
              60    80
               \
                65
      ```
    options:
      - text: '30'
      - text: '60'
        correct: true
      - text: '65'
      - text: '70'
    explanation: |
      50 has two children, so we copy in its in-order successor, the minimum of the
      right subtree. From 70 go left to 60, and 60 has no left child, so 60 is the
      minimum. It moves to the root, and deleting the old 60 leaves 65 as 70's left child.
  - question: |
      When deleting a node that has **only a right child**, what does the parent's
      pointer end up pointing at?
    options:
      - text: '`nil`'
      - text: The deleted node's right child
        correct: true
      - text: The in-order successor, copied in
      - text: The root of the whole tree
    explanation: |
      With one child, you just splice the node out: return its only child, and the
      parent's pointer skips straight to it. Every value in that child subtree was
      already on the correct side of the parent, so the BST rule still holds.
---

A player gets banned and their score has to go. Deleting from a BST is the trickiest
basic operation, and it uses the minimum, so let's do min and max first.

## Min and max

The smallest value is as far **left** as you can go, and the largest is as far
**right** as you can go:

```go
func (t *BST[T]) Min() (T, bool) {
	if t.root == nil {
		var zero T
		return zero, false
	}
	n := t.root
	for n.left != nil {
		n = n.left
	}
	return n.val, true
}
```

`Max` is the mirror image, following `right`. Both return a second `bool` because
an empty tree has no minimum, and returning a zero value alone would be ambiguous: is
the lowest score 0, or is the tree empty? This is the same "comma ok" idea as
`v, ok := m[key]`.

## Delete: three cases

Find the node first (just like search). Then:

1. **No children**: remove it. The parent's pointer becomes `nil`.
2. **One child**: splice it out. The parent points at the node's only child.
3. **Two children**: you can't just remove it, because there are two subtrees to
   reattach. Instead, find the **in-order successor**, the smallest value in the
   right subtree. Copy it into this node, then delete the successor from the right
   subtree. The successor has no left child (or it wouldn't be the smallest), so
   that second delete is always case 1 or 2.

Why the successor? It's the next-larger value in the whole tree. Everything in the
left subtree is still smaller than it, and everything left in the right subtree is
still larger, so the BST rule survives.

Cases 1 and 2 collapse into one line of code: if `left` is `nil`, return `right`
(which might also be `nil`).

```go
package main

import (
	"cmp"
	"fmt"
	"slices"
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

func minNode[T cmp.Ordered](n *node[T]) *node[T] {
	for n.left != nil {
		n = n.left
	}
	return n
}

// remove deletes v from the subtree rooted at n and returns the new subtree root.
func remove[T cmp.Ordered](n *node[T], v T) *node[T] {
	if n == nil {
		return nil // not found
	}
	switch {
	case v < n.val:
		n.left = remove(n.left, v)
	case v > n.val:
		n.right = remove(n.right, v)
	default:
		if n.left == nil {
			return n.right // zero or one child
		}
		if n.right == nil {
			return n.left // one child
		}
		succ := minNode(n.right) // two children
		n.val = succ.val
		n.right = remove(n.right, succ.val)
	}
	return n
}

func inOrder[T cmp.Ordered](n *node[T], out []T) []T {
	if n == nil {
		return out
	}
	out = inOrder(n.left, out)
	out = append(out, n.val)
	return inOrder(n.right, out)
}

func main() {
	var root *node[int]
	for _, s := range []int{50, 30, 70, 60, 80, 65} {
		root = insert(root, s)
	}
	root = remove(root, 50)
	fmt.Println(root.val, inOrder(root, nil))
	root = remove(root, 999) // not there: nothing happens
	fmt.Println(slices.Equal(inOrder(root, nil), []int{30, 60, 65, 70, 80}))
}
```

Output:

```
60 [30 60 65 70 80]
true
```

Deleting 50 found its successor, 60, copied it into the root, then removed the old 60
node, whose only child 65 moved up to take its place.

## Wiring it into BST

In the full `BST` type you'd track the size too. A neat approach is to check
`Contains` first, or return a `bool` from `remove` the way `insert` did, and
decrement `size` only when something was actually removed.

## Cost

Delete walks down to the node, then possibly down again to the successor. Both
walks follow a single path, so it's O(h), where h is the height.
