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

    // blackHeight is the checker from the first lesson of this chapter.
    func blackHeight[T cmp.Ordered](n *rbNode[T]) (int, error) {
    	if n == nil {
    		return 1, nil
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

    // ---- Your code ----

    // check returns nil if t is a healthy red-black tree, or an error
    // describing the first problem it finds. It checks that:
    //   - the root (if any) is black and has a nil parent,
    //   - every child's parent pointer points back at its parent,
    //   - an in-order walk gives strictly increasing values (BST order),
    //   - the number of nodes equals t.size,
    //   - blackHeight(t.root) reports no error.
    //
    // It must not change the tree.
    func check[T cmp.Ordered](t *RBTree[T]) error {
    	// ?
    	return nil
    }

    func main() {
    	var ids RBTree[int]
    	for id := range 1000 {
    		ids.Insert((id * 37) % 1000)
    	}
    	fmt.Println(check(&ids)) // want: <nil>

    	ids.root.left.right.parent = ids.root // a broken parent pointer
    	fmt.Println(check(&ids) != nil)       // want: true
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

    // blackHeight is the checker from the first lesson of this chapter.
    func blackHeight[T cmp.Ordered](n *rbNode[T]) (int, error) {
    	if n == nil {
    		return 1, nil
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

    // ---- Your code ----

    // check returns nil if t is a healthy red-black tree, or an error
    // describing the first problem it finds. It checks that:
    //   - the root (if any) is black and has a nil parent,
    //   - every child's parent pointer points back at its parent,
    //   - an in-order walk gives strictly increasing values (BST order),
    //   - the number of nodes equals t.size,
    //   - blackHeight(t.root) reports no error.
    //
    // It must not change the tree.
    func check[T cmp.Ordered](t *RBTree[T]) error {
    	if t.root != nil && (t.root.red || t.root.parent != nil) {
    		return fmt.Errorf("root %v must be black with a nil parent", t.root.val)
    	}
    	var vals []T
    	var walk func(n *rbNode[T]) error
    	walk = func(n *rbNode[T]) error {
    		if n == nil {
    			return nil
    		}
    		for _, c := range []*rbNode[T]{n.left, n.right} {
    			if c != nil && c.parent != n {
    				return fmt.Errorf("child %v of %v has the wrong parent", c.val, n.val)
    			}
    		}
    		if err := walk(n.left); err != nil {
    			return err
    		}
    		if len(vals) > 0 && vals[len(vals)-1] >= n.val {
    			return fmt.Errorf("%v comes after %v in order: not a BST", n.val, vals[len(vals)-1])
    		}
    		vals = append(vals, n.val)
    		return walk(n.right)
    	}
    	if err := walk(t.root); err != nil {
    		return err
    	}
    	if len(vals) != t.size {
    		return fmt.Errorf("tree has %d nodes but size is %d", len(vals), t.size)
    	}
    	_, err := blackHeight(t.root)
    	return err
    }

    func main() {
    	var ids RBTree[int]
    	for id := range 1000 {
    		ids.Insert((id * 37) % 1000)
    	}
    	fmt.Println(check(&ids)) // want: <nil>

    	ids.root.left.right.parent = ids.root // a broken parent pointer
    	fmt.Println(check(&ids) != nil)       // want: true
    }
  tests: |
    package main

    import "testing"

    // node builds a node and wires up its children's parent pointers.
    func node(val int, red bool, left, right *rbNode[int]) *rbNode[int] {
    	n := &rbNode[int]{val: val, red: red, left: left, right: right}
    	if left != nil {
    		left.parent = n
    	}
    	if right != nil {
    		right.parent = n
    	}
    	return n
    }

    const R, B = true, false

    // good is black 20 with red children 10 and 30.
    func good() *RBTree[int] {
    	return &RBTree[int]{root: node(20, B, node(10, R, nil, nil), node(30, R, nil, nil)), size: 3}
    }

    func TestHealthyTrees(t *testing.T) {
    	if err := check(&RBTree[int]{}); err != nil {
    		t.Errorf("empty tree: check = %v, want nil", err)
    	}
    	if err := check(good()); err != nil {
    		t.Errorf("black 20 with red children 10 and 30: check = %v, want nil", err)
    	}
    	var tr RBTree[int]
    	for i := range 3000 {
    		tr.Insert((i * 7919) % 3001)
    		if i%101 == 0 {
    			if err := check(&tr); err != nil {
    				t.Fatalf("after %d inserts, check on a tree built by Insert = %v, want nil", i+1, err)
    			}
    		}
    	}
    	if err := check(&tr); err != nil {
    		t.Fatalf("check on a tree built by 3000 inserts = %v, want nil", err)
    	}
    	before := tr.root
    	check(&tr)
    	if tr.root != before || tr.size != 3000 {
    		t.Error("check changed the tree")
    	}
    }

    func TestBrokenTrees(t *testing.T) {
    	tests := []struct {
    		name  string
    		build func() *RBTree[int]
    	}{
    		{"red root", func() *RBTree[int] { tr := good(); tr.root.red = true; return tr }},
    		{"root with a parent", func() *RBTree[int] { tr := good(); tr.root.parent = tr.root.left; return tr }},
    		{"child with nil parent", func() *RBTree[int] { tr := good(); tr.root.right.parent = nil; return tr }},
    		{"child pointing at the wrong parent", func() *RBTree[int] {
    			tr := good()
    			tr.root.left.left = node(5, B, nil, nil)
    			tr.root.left.red = false
    			tr.root.right.red = false
    			tr.root.left.left.parent = tr.root // wrong: should be 10
    			tr.root.left.left.red = true
    			tr.size = 4
    			return tr
    		}},
    		{"out of order (BST broken)", func() *RBTree[int] {
    			return &RBTree[int]{root: node(20, B, node(30, R, nil, nil), node(10, R, nil, nil)), size: 3}
    		}},
    		{"duplicate value", func() *RBTree[int] {
    			return &RBTree[int]{root: node(20, B, node(20, R, nil, nil), node(30, R, nil, nil)), size: 3}
    		}},
    		{"order broken deeper down", func() *RBTree[int] {
    			// 25 is in 20's LEFT subtree, but it's bigger than 20.
    			return &RBTree[int]{root: node(20, B, node(10, B, nil, node(25, R, nil, nil)), node(30, B, nil, nil)), size: 4}
    		}},
    		{"size too big", func() *RBTree[int] { tr := good(); tr.size = 4; return tr }},
    		{"size too small", func() *RBTree[int] { tr := good(); tr.size = 2; return tr }},
    		{"red node with a red child", func() *RBTree[int] {
    			return &RBTree[int]{root: node(20, B, node(10, R, node(5, R, nil, nil), nil), node(30, B, nil, nil)), size: 4}
    		}},
    		{"black heights differ", func() *RBTree[int] {
    			return &RBTree[int]{root: node(20, B, node(10, B, nil, nil), nil), size: 2}
    		}},
    	}
    	for _, tt := range tests {
    		if err := check(tt.build()); err == nil {
    			t.Errorf("%s: check = nil, want an error", tt.name)
    		}
    	}
    }
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

## Your turn: a tree health check

The testing tip below says to check more than colours. Let's build that checker.
`blackHeight` from the first lesson is already in the editor. Complete
`check(t)` so it returns `nil` for a healthy tree, or an error saying what's
wrong. It must catch:

- a **red root**, or a root whose `parent` isn't `nil`,
- a child whose `parent` pointer doesn't point back at its actual parent
  (the classic rotation bug),
- values out of **BST order**, including duplicates: an in-order walk must be
  strictly increasing, and checking each node only against its direct children
  isn't enough,
- a node count that doesn't match `t.size`,
- anything `blackHeight` complains about.

Collect the in-order values in a slice as you walk (or remember just the last
one), and compare each new value with the previous. Don't modify the tree.

## Testing tip

Balanced trees are the perfect target for *property tests*. Insert thousands of
random values, and after every insert check that:

- `blackHeight(t.root)` returns no error,
- the root is black,
- the in-order traversal is sorted (`slices.IsSorted`),
- every child's `parent` points back to its parent.

A bug in a rotation usually trips one of these within a few dozen inserts.
