---
title: Tree Traversals
quiz:
  - question: |
      What's the **pre-order** traversal of this tree?

      ```
            8
           / \
          3   10
         / \    \
        1   6    14
      ```
    options:
      - text: 1, 3, 6, 8, 10, 14
      - text: 8, 3, 1, 6, 10, 14
        correct: true
      - text: 1, 6, 3, 14, 10, 8
      - text: 8, 3, 10, 1, 6, 14
    explanation: |
      Pre-order visits the node *before* its subtrees: 8, then all of the left
      subtree (3, 1, 6), then all of the right subtree (10, 14). The first option is
      in-order, the third is post-order, and the last is level-by-level, which is
      breadth-first order and a topic for a later chapter.
  - question: |
      In the `walk` method, what happens if the `range` loop's body executes `break`?

      ```go
      func (n *node[T]) walk(yield func(T) bool) bool {
          if n == nil {
              return true
          }
          return n.left.walk(yield) && yield(n.val) && n.right.walk(yield)
      }
      ```
    options:
      - text: '`yield` returns `false`, the `&&` short-circuits, and `false` propagates up so no more nodes are visited'
        correct: true
      - text: The walk continues through the rest of the tree, but the values are discarded
      - text: The program panics because the iterator kept going
      - text: '`break` isn''t allowed inside a range-over-func loop'
    explanation: |
      `break` makes `yield` return `false`. Because every call is chained with `&&`,
      the `false` stops each level from doing any more work and bubbles all the way
      up. Calling `yield` again after it returned `false` would panic, which is
      exactly what this pattern avoids.
exercise:
  starter: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"iter"
    )

    type node[T cmp.Ordered] struct {
    	val         T
    	left, right *node[T]
    }

    type BST[T cmp.Ordered] struct{ root *node[T] }

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

    func (t *BST[T]) Insert(v T) { t.root = insert(t.root, v) }

    // Descending yields the tree's values from largest to smallest,
    // and stops as soon as the caller's loop breaks.
    func (t *BST[T]) Descending() iter.Seq[T] {
    	return func(yield func(T) bool) {
    		// ?
    	}
    }

    func main() {
    	var scores BST[int]
    	for _, s := range []int{500, 250, 800, 100, 300, 600, 950} {
    		scores.Insert(s)
    	}
    	fmt.Println("Podium:")
    	rank := 1
    	for s := range scores.Descending() {
    		fmt.Printf("#%d: %d\n", rank, s)
    		if rank == 3 {
    			break // podium only
    		}
    		rank++
    	}
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"iter"
    )

    type node[T cmp.Ordered] struct {
    	val         T
    	left, right *node[T]
    }

    type BST[T cmp.Ordered] struct{ root *node[T] }

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

    func (t *BST[T]) Insert(v T) { t.root = insert(t.root, v) }

    func (n *node[T]) walkDown(yield func(T) bool) bool {
    	if n == nil {
    		return true
    	}
    	return n.right.walkDown(yield) && yield(n.val) && n.left.walkDown(yield)
    }

    func (t *BST[T]) Descending() iter.Seq[T] {
    	return func(yield func(T) bool) {
    		t.root.walkDown(yield)
    	}
    }

    func main() {
    	var scores BST[int]
    	for _, s := range []int{500, 250, 800, 100, 300, 600, 950} {
    		scores.Insert(s)
    	}
    	fmt.Println("Podium:")
    	rank := 1
    	for s := range scores.Descending() {
    		fmt.Printf("#%d: %d\n", rank, s)
    		if rank == 3 {
    			break
    		}
    		rank++
    	}
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func build(vals ...int) *BST[int] {
    	t := &BST[int]{}
    	for _, v := range vals {
    		t.Insert(v)
    	}
    	return t
    }

    func TestDescendingOrder(t *testing.T) {
    	tree := build(500, 250, 800, 100, 300, 600, 950)
    	got := slices.Collect(tree.Descending())
    	want := []int{950, 800, 600, 500, 300, 250, 100}
    	if !slices.Equal(got, want) {
    		t.Errorf("Descending() = %v, want %v", got, want)
    	}
    }

    func TestDescendingSortedInput(t *testing.T) {
    	tree := build(1, 2, 3, 4, 5)
    	got := slices.Collect(tree.Descending())
    	want := []int{5, 4, 3, 2, 1}
    	if !slices.Equal(got, want) {
    		t.Errorf("Descending() on a lopsided tree = %v, want %v", got, want)
    	}
    }

    func TestDescendingEmpty(t *testing.T) {
    	var tree BST[int]
    	if got := slices.Collect(tree.Descending()); len(got) != 0 {
    		t.Errorf("Descending() on an empty tree = %v, want nothing", got)
    	}
    }

    func TestDescendingStopsEarly(t *testing.T) {
    	tree := build(500, 250, 800, 100, 300, 600, 950)
    	var got []int
    	defer func() {
    		if r := recover(); r != nil {
    			t.Fatalf("breaking out of the loop panicked: %v (did you keep calling yield after it returned false?)", r)
    		}
    	}()
    	for s := range tree.Descending() {
    		got = append(got, s)
    		if len(got) == 2 {
    			break
    		}
    	}
    	if want := []int{950, 800}; !slices.Equal(got, want) {
    		t.Errorf("first two values = %v, want %v", got, want)
    	}
    }

    func TestDescendingStrings(t *testing.T) {
    	var names BST[string]
    	for _, n := range []string{"mira", "ash", "zed", "kai"} {
    		names.Insert(n)
    	}
    	got := slices.Collect(names.Descending())
    	want := []string{"zed", "mira", "kai", "ash"}
    	if !slices.Equal(got, want) {
    		t.Errorf("Descending() = %v, want %v", got, want)
    	}
    }
---

Searching visits one path. Sometimes you need *every* node: printing the whole
leaderboard, saving the tree to disk, or freeing it. There are three classic
depth-first orders, named for when you visit the node relative to its children:

| Order | Visit order | Handy for |
|---|---|---|
| **Pre-order** | node, left, right | Copying or serialising a tree (the root comes first) |
| **In-order** | left, node, right | Getting a BST's values in **sorted** order |
| **Post-order** | left, right, node | Deleting a tree, or computing sizes (children first) |

For this tree:

```
              500
            /     \
         250       800
        /   \     /   \
      100   300 600   950
```

- Pre-order: 500 250 100 300 800 600 950
- In-order: 100 250 300 500 600 800 950
- Post-order: 100 300 250 600 950 800 500

## Collecting into a slice

The simplest version appends to a slice as it goes:

```go
func (n *node[T]) preOrder(out []T) []T {
	if n == nil {
		return out
	}
	out = append(out, n.val)
	out = n.left.preOrder(out)
	return n.right.preOrder(out)
}
```

Note that we call methods on `n.left` even when it's `nil`. In Go, calling a method
on a nil pointer is fine as long as the method checks for `nil` before touching any
fields. That makes recursive tree code very compact.

## In-order as an iterator

Building a slice of a million players just to print the top of the list is
wasteful. Since Go 1.23 you can return an `iter.Seq[T]` instead, and the caller
ranges over it and can stop early:

```go
package main

import (
	"cmp"
	"fmt"
	"iter"
)

type node[T cmp.Ordered] struct {
	val         T
	left, right *node[T]
}

type BST[T cmp.Ordered] struct{ root *node[T] }

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

func (t *BST[T]) Insert(v T) { t.root = insert(t.root, v) }

// walk visits the subtree in order and reports whether to keep going.
func (n *node[T]) walk(yield func(T) bool) bool {
	if n == nil {
		return true
	}
	return n.left.walk(yield) && yield(n.val) && n.right.walk(yield)
}

// All returns the values in ascending order.
func (t *BST[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		t.root.walk(yield)
	}
}

func (n *node[T]) postOrder(out []T) []T {
	if n == nil {
		return out
	}
	out = n.left.postOrder(out)
	out = n.right.postOrder(out)
	return append(out, n.val)
}

func main() {
	var scores BST[int]
	for _, s := range []int{500, 250, 800, 100, 300, 600, 950} {
		scores.Insert(s)
	}

	for s := range scores.All() {
		if s > 500 {
			break
		}
		fmt.Print(s, " ")
	}
	fmt.Println()

	fmt.Println(scores.root.postOrder(nil))
}
```

Output:

```
100 250 300 500 
[100 300 250 600 950 800 500]
```

(This version of the BST drops the size counter to keep the listing short.)

## How the early exit works

`yield` returns `false` when the loop body hits `break` (or `return`). The
expression `n.left.walk(yield) && yield(n.val) && n.right.walk(yield)` relies on
`&&` short-circuiting: the moment anything returns `false`, the rest is skipped
and `false` is passed back to the parent, which also stops. Calling `yield` again
after it returned `false` panics at runtime, so this propagation isn't optional.

Because `All` returns a plain `iter.Seq[int]`, it plugs into the standard library
too: `slices.Collect(scores.All())` gives you a sorted `[]int`.

## Cost

Every traversal touches each node once, so it's O(n) time. The recursion depth is
the tree's height, so the call stack uses O(h) space.

## Your turn

The leaderboard screen shows the **highest** scores first. In the exercise, complete
`Descending`, an `iter.Seq[T]` that yields the tree's values from largest to smallest.
Think about which subtree an in-order walk visits first, and flip it. Your iterator must
stop cleanly when the caller's loop breaks: never call `yield` again after it has
returned `false`.
