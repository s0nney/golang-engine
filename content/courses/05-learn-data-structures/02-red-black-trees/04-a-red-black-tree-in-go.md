---
title: A Red-Black Tree in Go
quiz:
  - question: In the `Insert` method, why does the search loop return early when `v == n.val`?
    options:
      - text: To avoid an infinite loop
      - text: The tree stores each value once, so an equal value means there's nothing to insert
        correct: true
      - text: Equal values must be inserted by the fix-up instead
      - text: Because `cmp.Ordered` can't compare equal values
    explanation: |
      Like our BST from chapter 1, this tree behaves like a set. If the value is already
      there, the tree doesn't change, so there's no new node, no size increment and no
      fix-up.
  - question: |
      You insert 1,000,000 player IDs **in increasing order** into the tree from this
      lesson. Roughly how tall can it get?
    options:
      - text: About 20, the same as a perfect tree
      - text: At most about 40
        correct: true
      - text: About 1,000
      - text: About 1,000,000
    explanation: |
      A red-black tree's height is at most 2·log₂(n + 1). For a million nodes, log₂ is
      about 20, so the height is at most about 40. The program in this lesson measures
      36 for a million sorted inserts. A plain BST would be 999,999 tall.
---

Let's assemble the pieces into a working tree and throw the worst possible input at
it: player IDs arriving in sorted order.

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

type RBTree[T cmp.Ordered] struct {
	root *rbNode[T]
	size int
}

func isRed[T cmp.Ordered](n *rbNode[T]) bool { return n != nil && n.red }

func (t *RBTree[T]) replaceChild(parent, old, newChild *rbNode[T]) {
	switch {
	case parent == nil:
		t.root = newChild
	case parent.left == old:
		parent.left = newChild
	default:
		parent.right = newChild
	}
	if newChild != nil {
		newChild.parent = parent
	}
}

func (t *RBTree[T]) rotateLeft(x *rbNode[T]) {
	y := x.right
	x.right = y.left
	if y.left != nil {
		y.left.parent = x
	}
	t.replaceChild(x.parent, x, y)
	y.left = x
	x.parent = y
}

func (t *RBTree[T]) rotateRight(x *rbNode[T]) {
	y := x.left
	x.left = y.right
	if y.right != nil {
		y.right.parent = x
	}
	t.replaceChild(x.parent, x, y)
	y.right = x
	x.parent = y
}

func (t *RBTree[T]) Insert(v T) {
	var parent *rbNode[T]
	for n := t.root; n != nil; {
		parent = n
		switch {
		case v < n.val:
			n = n.left
		case v > n.val:
			n = n.right
		default:
			return // already present
		}
	}
	z := &rbNode[T]{val: v, red: true, parent: parent}
	switch {
	case parent == nil:
		t.root = z
	case v < parent.val:
		parent.left = z
	default:
		parent.right = z
	}
	t.size++
	t.fixInsert(z)
}

func (t *RBTree[T]) fixInsert(z *rbNode[T]) {
	for isRed(z.parent) {
		p := z.parent
		g := p.parent
		if p == g.left {
			uncle := g.right
			if isRed(uncle) {
				p.red, uncle.red, g.red = false, false, true
				z = g
				continue
			}
			if z == p.right {
				t.rotateLeft(p)
				z, p = p, z
			}
			p.red, g.red = false, true
			t.rotateRight(g)
		} else {
			uncle := g.left
			if isRed(uncle) {
				p.red, uncle.red, g.red = false, false, true
				z = g
				continue
			}
			if z == p.left {
				t.rotateRight(p)
				z, p = p, z
			}
			p.red, g.red = false, true
			t.rotateLeft(g)
		}
	}
	t.root.red = false
}

func (t *RBTree[T]) Contains(v T) bool {
	for n := t.root; n != nil; {
		switch {
		case v < n.val:
			n = n.left
		case v > n.val:
			n = n.right
		default:
			return true
		}
	}
	return false
}

func height[T cmp.Ordered](n *rbNode[T]) int {
	if n == nil {
		return -1
	}
	return 1 + max(height(n.left), height(n.right))
}

func main() {
	var small RBTree[int]
	for _, id := range []int{10, 20, 30} {
		small.Insert(id)
	}
	r := small.root
	fmt.Println(r.val, r.red, r.left.val, r.left.red, r.right.val, r.right.red)

	var ids RBTree[int]
	for id := range 1_000_000 {
		ids.Insert(id) // sorted input: a plain BST's worst case
	}
	fmt.Println(ids.size, height(ids.root), ids.Contains(424_242))
}
```

Output:

```
20 false 10 true 30 true
1000000 36 true
```

Inserting 10, 20, 30 into a plain BST makes a chain. Here the third insert triggered
case 3, and the tree rotated into a black 20 with two red children.

The million sorted inserts produced a tree of height **36**. A plain BST fed the same
data would be 999,999 levels tall, and building it would take around 500 billion
steps. This one finishes in well under a second.

## What we left out

**Delete.** Deleting from a red-black tree follows the same pattern (BST delete, then a
fix-up), but when you remove a black node, a path loses a black and the fix-up has
more cases (four, plus mirrors). It's a great exercise if you want a challenge, and
the `blackHeight` checker from the first lesson will tell you if you got it wrong.

**Iteration.** The `All() iter.Seq[T]` method from chapter 1 works unchanged, since
the node still has `left` and `right`. Colours and parents don't matter for
traversal.

## Testing tip

Balanced trees are the perfect target for *property tests*. Insert thousands of
random values, and after every insert check that:

- `blackHeight(t.root)` returns no error,
- the root is black,
- the in-order traversal is sorted (`slices.IsSorted`),
- every child's `parent` points back to its parent.

A bug in a rotation usually trips one of these within a few dozen inserts.
