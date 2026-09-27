---
title: A Generic BST
quiz:
  - question: |
      What does this print?

      ```go
      var t BST[int]
      for _, s := range []int{40, 20, 60, 20, 40} {
          t.Insert(s)
      }
      fmt.Println(t.Len())
      ```
    options:
      - text: '5'
      - text: '3'
        correct: true
      - text: '2'
      - text: '0'
    explanation: |
      Our `insert` ignores values that are already in the tree, and only bumps `size`
      when it creates a new node. The distinct values are 40, 20 and 60.
  - question: |
      What's wrong with this version of `insert`?

      ```go
      func insert[T cmp.Ordered](n *node[T], v T) {
          if n == nil {
              n = &node[T]{val: v}
              return
          }
          if v < n.val {
              insert(n.left, v)
          } else if v > n.val {
              insert(n.right, v)
          }
      }
      ```
    options:
      - text: Nothing, it works
      - text: It recurses forever
      - text: Assigning to `n` only changes the local copy of the pointer, so the new node is never attached to the tree
        correct: true
      - text: It panics with a nil pointer dereference
    explanation: |
      Go passes the pointer by value. `n = &node[T]{...}` changes the function's own
      variable, not the parent's `left` or `right` field. That's why our version
      *returns* the (possibly new) subtree root and the caller stores it.
exercise:
  starter: |
    package main

    import (
    	"cmp"
    	"fmt"
    )

    type node[T cmp.Ordered] struct {
    	val         T
    	count       int // how many times val was inserted
    	left, right *node[T]
    }

    type BST[T cmp.Ordered] struct {
    	root *node[T]
    }

    // Insert adds v to the tree. If v is already there, bump its count instead.
    func (t *BST[T]) Insert(v T) {
    	// ?
    }

    // Count reports how many times v was inserted (0 if never).
    func (t *BST[T]) Count(v T) int {
    	// ?
    	return 0
    }

    func main() {
    	var scores BST[int]
    	for _, s := range []int{500, 250, 800, 250, 600, 250} {
    		scores.Insert(s)
    	}
    	fmt.Println("players on 250:", scores.Count(250)) // want 3
    	fmt.Println("players on 800:", scores.Count(800)) // want 1
    	fmt.Println("players on 999:", scores.Count(999)) // want 0
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    )

    type node[T cmp.Ordered] struct {
    	val         T
    	count       int
    	left, right *node[T]
    }

    type BST[T cmp.Ordered] struct {
    	root *node[T]
    }

    func insert[T cmp.Ordered](n *node[T], v T) *node[T] {
    	if n == nil {
    		return &node[T]{val: v, count: 1}
    	}
    	switch {
    	case v < n.val:
    		n.left = insert(n.left, v)
    	case v > n.val:
    		n.right = insert(n.right, v)
    	default:
    		n.count++
    	}
    	return n
    }

    func (t *BST[T]) Insert(v T) {
    	t.root = insert(t.root, v)
    }

    func (t *BST[T]) Count(v T) int {
    	n := t.root
    	for n != nil {
    		switch {
    		case v < n.val:
    			n = n.left
    		case v > n.val:
    			n = n.right
    		default:
    			return n.count
    		}
    	}
    	return 0
    }

    func main() {
    	var scores BST[int]
    	for _, s := range []int{500, 250, 800, 250, 600, 250} {
    		scores.Insert(s)
    	}
    	fmt.Println("players on 250:", scores.Count(250))
    	fmt.Println("players on 800:", scores.Count(800))
    	fmt.Println("players on 999:", scores.Count(999))
    }
  tests: |
    package main

    import "testing"

    func TestCountScores(t *testing.T) {
    	var scores BST[int]
    	for _, s := range []int{500, 250, 800, 250, 600, 250, 100, 800} {
    		scores.Insert(s)
    	}
    	for _, tt := range []struct{ score, want int }{
    		{500, 1}, {250, 3}, {800, 2}, {600, 1}, {100, 1}, {999, 0}, {0, 0}, {300, 0},
    	} {
    		if got := scores.Count(tt.score); got != tt.want {
    			t.Errorf("Count(%d) = %d, want %d", tt.score, got, tt.want)
    		}
    	}
    }

    func TestKeepsBSTShape(t *testing.T) {
    	var scores BST[int]
    	for _, s := range []int{50, 30, 70, 30, 60} {
    		scores.Insert(s)
    	}
    	r := scores.root
    	if r == nil || r.val != 50 {
    		t.Fatalf("root should be 50 (the first value inserted)")
    	}
    	if r.left == nil || r.left.val != 30 || r.left.count != 2 {
    		t.Errorf("root.left should be 30 with count 2")
    	}
    	if r.right == nil || r.right.val != 70 || r.right.left == nil || r.right.left.val != 60 {
    		t.Errorf("60 should be the left child of 70")
    	}
    }

    func TestStrings(t *testing.T) {
    	var names BST[string]
    	for _, n := range []string{"mira", "ash", "mira", "zed"} {
    		names.Insert(n)
    	}
    	if got := names.Count("mira"); got != 2 {
    		t.Errorf(`Count("mira") = %d, want 2`, got)
    	}
    	if got := names.Count("bob"); got != 0 {
    		t.Errorf(`Count("bob") = %d, want 0`, got)
    	}
    }

    func TestEmptyTree(t *testing.T) {
    	var empty BST[float64]
    	if got := empty.Count(1.5); got != 0 {
    		t.Errorf("Count on an empty tree = %d, want 0", got)
    	}
    }
---

Time to build a real, reusable tree. You know generics from the previous course,
so let's write one BST that works for scores (`int`), usernames (`string`), match
times (`float64`), or any other ordered type.

## The types

```go
type node[T cmp.Ordered] struct {
	val         T
	left, right *node[T]
}

type BST[T cmp.Ordered] struct {
	root *node[T]
	size int
}
```

`cmp.Ordered` is the standard constraint for "anything you can use `<` on": all the
integer and float types, plus strings. `node` is lowercase because callers never
need to see it; they only use `BST`. The zero value of `BST` is an empty tree ready
to use, just like the zero value of a `strings.Builder` or `sync.Mutex`.

## Insert

Recursion makes insert tidy. The helper takes a subtree and returns the new root of
that subtree:

```go
func insert[T cmp.Ordered](n *node[T], v T) (*node[T], bool) {
	if n == nil {
		return &node[T]{val: v}, true // empty spot: the new node goes here
	}
	var added bool
	switch {
	case v < n.val:
		n.left, added = insert(n.left, v)
	case v > n.val:
		n.right, added = insert(n.right, v)
	}
	// equal: already present, do nothing
	return n, added
}

func (t *BST[T]) Insert(v T) {
	var added bool
	t.root, added = insert(t.root, v)
	if added {
		t.size++
	}
}
```

Returning the subtree root is the key trick. When the recursion reaches a `nil`
child, it returns a brand-new node and the parent stores it with
`n.left, added = insert(n.left, v)`. Every other level just returns itself unchanged.

## Search

```go
func (t *BST[T]) Contains(v T) bool {
	n := t.root
	for n != nil {
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

func (t *BST[T]) Len() int { return t.size }
```

Search doesn't change the tree, so a loop is simpler than recursion here.

## Putting it together

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

type BST[T cmp.Ordered] struct {
	root *node[T]
	size int
}

func insert[T cmp.Ordered](n *node[T], v T) (*node[T], bool) {
	if n == nil {
		return &node[T]{val: v}, true
	}
	var added bool
	switch {
	case v < n.val:
		n.left, added = insert(n.left, v)
	case v > n.val:
		n.right, added = insert(n.right, v)
	}
	return n, added
}

func (t *BST[T]) Insert(v T) {
	var added bool
	t.root, added = insert(t.root, v)
	if added {
		t.size++
	}
}

func (t *BST[T]) Contains(v T) bool {
	n := t.root
	for n != nil {
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

func (t *BST[T]) Len() int { return t.size }

func main() {
	var names BST[string]
	for _, name := range []string{"mira", "ash", "zed", "kai", "ash"} {
		names.Insert(name)
	}
	fmt.Println(names.Len(), names.Contains("kai"), names.Contains("bob"))
}
```

This prints `4 true false`. The second `"ash"` was ignored, and strings compare
alphabetically, so `"ash"` went left of `"mira"` and `"zed"` went right.

## Cost

Both `Insert` and `Contains` walk one path from the root down, so both are O(h),
where h is the tree's height. For a balanced tree that's O(log n).

## A note on floats

`cmp.Ordered` includes `float64`, and floats have one weird value: `NaN`. Every
comparison with `NaN` is false, so `v < n.val` and `v > n.val` are both false and a
`NaN` looks "equal" to everything. If NaN can show up in your data, compare with
`cmp.Compare(a, b)`, which treats NaN as smaller than every other number, or
reject NaN before inserting.

## Your turn

Real leaderboards have ties: several players on exactly 250 points. In the exercise,
each node has a `count` field. Complete `Insert` so that a new score creates a node with
`count` 1, and a score that's already in the tree bumps that node's `count` instead of
being ignored. Then complete `Count`, which returns how many times a score was inserted,
or 0 if it never was. Keep the BST shape: smaller values go left, larger go right.
