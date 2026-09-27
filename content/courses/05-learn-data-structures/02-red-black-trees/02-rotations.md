---
title: Rotations
quiz:
  - question: |
      You rotate this tree **left** at `10`. Which value becomes the new root?

      ```
        10
          \
           20
          /  \
        15    30
      ```
    options:
      - text: '10'
      - text: '15'
      - text: '20'
        correct: true
      - text: '30'
    explanation: |
      A left rotation lifts the right child into the node's place. 20 becomes the root,
      10 becomes 20's left child, and 20's old left child (15) moves over to become
      10's right child.
  - question: What happens to a BST's in-order traversal after a rotation?
    options:
      - text: It stays exactly the same
        correct: true
      - text: It gets reversed
      - text: The rotated node moves to the end
      - text: It's no longer sorted until you rebalance
    explanation: |
      Rotations only change the *shape* of the tree, never the left-to-right order of
      the values. That's what makes them safe to use for rebalancing a BST.
exercise:
  starter: |
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
    }

    // replaceChild makes parent point at newChild where it used to point at old.
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

    // rotateRight is the mirror image of rotateLeft: x's left child y takes
    // x's place, and x becomes y's right child.
    func (t *RBTree[T]) rotateRight(x *rbNode[T]) {
    	// ?
    }

    // link builds a node and sets its children's parent pointers.
    func link[T cmp.Ordered](val T, left, right *rbNode[T]) *rbNode[T] {
    	n := &rbNode[T]{val: val, left: left, right: right}
    	if left != nil {
    		left.parent = n
    	}
    	if right != nil {
    		right.parent = n
    	}
    	return n
    }

    func show[T cmp.Ordered](n *rbNode[T]) string {
    	if n == nil {
    		return "."
    	}
    	return fmt.Sprintf("(%s %v %s)", show(n.left), n.val, show(n.right))
    }

    func main() {
    	// A left-leaning chain of player IDs: 30 -> 20 -> 10.
    	var t RBTree[int]
    	t.root = link(30, link(20, link(10, nil, nil), nil), nil)
    	fmt.Println("before:", show(t.root))
    	t.rotateRight(t.root)
    	fmt.Println("after: ", show(t.root)) // want ((. 10 .) 20 (. 30 .))
    }
  solution: |
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
    }

    // replaceChild makes parent point at newChild where it used to point at old.
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

    // link builds a node and sets its children's parent pointers.
    func link[T cmp.Ordered](val T, left, right *rbNode[T]) *rbNode[T] {
    	n := &rbNode[T]{val: val, left: left, right: right}
    	if left != nil {
    		left.parent = n
    	}
    	if right != nil {
    		right.parent = n
    	}
    	return n
    }

    func show[T cmp.Ordered](n *rbNode[T]) string {
    	if n == nil {
    		return "."
    	}
    	return fmt.Sprintf("(%s %v %s)", show(n.left), n.val, show(n.right))
    }

    func main() {
    	// A left-leaning chain of player IDs: 30 -> 20 -> 10.
    	var t RBTree[int]
    	t.root = link(30, link(20, link(10, nil, nil), nil), nil)
    	fmt.Println("before:", show(t.root))
    	t.rotateRight(t.root)
    	fmt.Println("after: ", show(t.root)) // want ((. 10 .) 20 (. 30 .))
    }
  tests: |
    package main

    import "testing"

    // checkParents verifies every child's parent pointer points back at its parent.
    func checkParents(t *testing.T, n, parent *rbNode[int]) {
    	t.Helper()
    	if n == nil {
    		return
    	}
    	if n.parent != parent {
    		want := "nil"
    		if parent != nil {
    			want = show(parent)
    		}
    		t.Errorf("node %d has the wrong parent pointer, want parent %s", n.val, want)
    	}
    	checkParents(t, n.left, n)
    	checkParents(t, n.right, n)
    }

    func TestRotateRightAtRoot(t *testing.T) {
    	var tree RBTree[int]
    	tree.root = link(30, link(20, link(10, nil, nil), nil), nil)
    	tree.rotateRight(tree.root)
    	if got, want := show(tree.root), "((. 10 .) 20 (. 30 .))"; got != want {
    		t.Fatalf("after rotateRight at 30, tree = %s, want %s", got, want)
    	}
    	checkParents(t, tree.root, nil)
    }

    func TestRotateRightMovesInnerSubtree(t *testing.T) {
    	//        50               30
    	//       /  \             /  \
    	//     30    60   ->    20    50
    	//    /  \                   /  \
    	//  20    40               40    60
    	var tree RBTree[int]
    	tree.root = link(50, link(30, link(20, nil, nil), link(40, nil, nil)), link(60, nil, nil))
    	tree.rotateRight(tree.root)
    	if got, want := show(tree.root), "((. 20 .) 30 ((. 40 .) 50 (. 60 .)))"; got != want {
    		t.Fatalf("after rotateRight at 50, tree = %s, want %s", got, want)
    	}
    	checkParents(t, tree.root, nil)
    }

    func TestRotateRightBelowRoot(t *testing.T) {
    	// Rotate at 80, the right child of the root 40.
    	var tree RBTree[int]
    	tree.root = link(40, link(20, nil, nil), link(80, link(60, link(50, nil, nil), link(70, nil, nil)), link(90, nil, nil)))
    	tree.rotateRight(tree.root.right)
    	if tree.root.val != 40 {
    		t.Fatalf("the root changed to %d; rotating below the root must leave it alone", tree.root.val)
    	}
    	if got, want := show(tree.root), "((. 20 .) 40 ((. 50 .) 60 ((. 70 .) 80 (. 90 .))))"; got != want {
    		t.Fatalf("after rotateRight at 80, tree = %s, want %s", got, want)
    	}
    	checkParents(t, tree.root, nil)
    }

    func TestRotateRightThenLeftUndoes(t *testing.T) {
    	var tree RBTree[int]
    	tree.root = link(50, link(30, link(20, nil, nil), link(40, nil, nil)), link(60, nil, nil))
    	before := show(tree.root)
    	tree.rotateRight(tree.root)
    	tree.rotateLeft(tree.root)
    	if got := show(tree.root); got != before {
    		t.Errorf("rotateRight then rotateLeft gave %s, want the original %s", got, before)
    	}
    	checkParents(t, tree.root, nil)
    }
---

A **rotation** is a small, local rewiring of pointers that changes a tree's shape
while keeping it a valid BST. It's the only structural tool a red-black tree needs.

## Rotating left

A left rotation at node `x` lifts its right child `y` up into `x`'s place, and `x`
drops down to become `y`'s left child:

```
      x                      y
     / \      rotate        / \
    a   y     left at x    x   c
       / \    ------->    / \
      b   c              a   b
```

`a`, `b` and `c` are whole subtrees, possibly empty. Look at `b`: it was `y`'s left
subtree, so every value in it is between `x` and `y`. After the rotation it's `x`'s
right subtree, and it's still between `x` and `y`. In-order was `a x b y c` before and
it's `a x b y c` after. Same values, same order, different shape.

A **right rotation** is the exact mirror, and it undoes a left rotation.

## Rotations in Go (no parent pointers)

On a plain BST node, a rotation takes the subtree root and returns the new one,
just like `insert` did:

```go
package main

import "fmt"

type node struct {
	val         int
	left, right *node
}

func rotateLeft(x *node) *node {
	y := x.right
	x.right = y.left // b moves across
	y.left = x
	return y
}

func rotateRight(y *node) *node {
	x := y.left
	y.left = x.right
	x.right = y
	return x
}

func inOrder(n *node, out []int) []int {
	if n == nil {
		return out
	}
	out = inOrder(n.left, out)
	out = append(out, n.val)
	return inOrder(n.right, out)
}

func main() {
	// A right-leaning chain, like you get from sorted inserts.
	root := &node{val: 10, right: &node{val: 20, right: &node{val: 30}}}
	fmt.Println(root.val, inOrder(root, nil))

	root = rotateLeft(root)
	fmt.Println(root.val, root.left.val, root.right.val, inOrder(root, nil))
}
```

Output:

```
10 [10 20 30]
20 10 30 [10 20 30]
```

One rotation turned a height-2 chain into a height-1 tree. That's the whole idea
of rebalancing: spot a lopsided spot and rotate it level.

## With parent pointers

Our red-black tree stores a `parent` pointer in each node, because the insert fix-up
needs to walk *up* the tree. That means a rotation has more pointers to update: the
moved subtree's parent, and the link from `x`'s old parent (or the tree's root)
down to `y`. A small helper handles the second part:

```go
// replaceChild makes parent point at newChild where it used to point at old.
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
```

`rotateRight` is the mirror image: swap every `left` and `right`.

A classic bug here is forgetting one of the parent updates. The tree still *looks*
fine going downward, and in-order traversal still works, but the next fix-up walks up
through a stale `parent` pointer and corrupts things. The black-height checker from
the last lesson, plus a check that `child.parent == n` everywhere, is well worth
running in your tests.

Each rotation touches a constant number of pointers, so it's **O(1)**.

## Your turn

The exercise gives you an `RBTree` with parent pointers, `replaceChild` and a working
`rotateLeft`. Complete `rotateRight`: `x`'s left child `y` takes `x`'s place (under
`x`'s old parent, or as the root), `y`'s old right subtree becomes `x`'s left subtree,
and `x` becomes `y`'s right child. The tests check the shape of the tree **and** that
every node's `parent` pointer is correct afterwards.
