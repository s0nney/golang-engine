---
title: Tree Audit
difficulty: medium
after: binary-trees
hints:
  - 'Checking only that `left.val < n.val < right.val` at every node isn''t enough: a node deep in the **left** subtree must still be smaller than the root. Try your idea on the "grandchild" example before coding it.'
  - 'Pass bounds down: every node must lie strictly between a lower and an upper bound. Going left, the current value becomes the new upper bound; going right, it becomes the new lower bound. Since `T` is generic you have no "minus infinity", so use pointers (`lo, hi *T`, where `nil` means no bound).'
  - 'Another way: an in-order traversal of a valid BST visits values in strictly increasing order. Walk the tree in order and compare each value with the previous one (and remember whether there *is* a previous one yet).'
exercise:
  starter: |
    package main

    import (
    	"cmp"
    	"fmt"
    )

    type node[T cmp.Ordered] struct {
    	val         T
    	left, right *node[T]
    }

    func validBST[T cmp.Ordered](root *node[T]) bool {
    	return false
    }

    func main() {
    	good := &node[int]{val: 50, left: &node[int]{val: 30}, right: &node[int]{val: 70}}
    	sneaky := &node[int]{val: 50,
    		left:  &node[int]{val: 30, right: &node[int]{val: 60}},
    		right: &node[int]{val: 70},
    	}
    	fmt.Println(validBST(good))   // want true
    	fmt.Println(validBST(sneaky)) // want false
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    )

    type node[T cmp.Ordered] struct {
    	val         T
    	left, right *node[T]
    }

    func validBST[T cmp.Ordered](root *node[T]) bool {
    	return within(root, nil, nil)
    }

    // within reports whether every value under n is strictly between
    // *lo and *hi (a nil bound means unbounded) and n is a valid BST.
    func within[T cmp.Ordered](n *node[T], lo, hi *T) bool {
    	if n == nil {
    		return true
    	}
    	if lo != nil && n.val <= *lo || hi != nil && n.val >= *hi {
    		return false
    	}
    	return within(n.left, lo, &n.val) && within(n.right, &n.val, hi)
    }

    func main() {
    	good := &node[int]{val: 50, left: &node[int]{val: 30}, right: &node[int]{val: 70}}
    	sneaky := &node[int]{val: 50,
    		left:  &node[int]{val: 30, right: &node[int]{val: 60}},
    		right: &node[int]{val: 70},
    	}
    	fmt.Println(validBST(good))
    	fmt.Println(validBST(sneaky))
    }
  tests: |
    package main

    import (
    	"math"
    	"testing"
    )

    type n = node[int]

    func TestValidBSTInts(t *testing.T) {
    	tests := []struct {
    		name string
    		root *n
    		want bool
    	}{
    		{"empty tree", nil, true},
    		{"single node", &n{val: 7}, true},
    		{"three nodes", &n{val: 50, left: &n{val: 30}, right: &n{val: 70}}, true},
    		{"left child too big", &n{val: 50, left: &n{val: 55}}, false},
    		{"right child too small", &n{val: 50, right: &n{val: 45}}, false},
    		{"duplicate on the left", &n{val: 50, left: &n{val: 50}}, false},
    		{"duplicate on the right", &n{val: 50, right: &n{val: 50}}, false},
    		{"grandchild bigger than root", &n{val: 50, left: &n{val: 30, right: &n{val: 60}}, right: &n{val: 70}}, false},
    		{"grandchild smaller than root", &n{val: 50, left: &n{val: 30}, right: &n{val: 70, left: &n{val: 40}}}, false},
    		{"grandchild equal to root", &n{val: 50, left: &n{val: 30, right: &n{val: 50}}}, false},
    		{
    			"deep valid tree",
    			&n{val: 50,
    				left:  &n{val: 30, left: &n{val: 10, right: &n{val: 20}}, right: &n{val: 40, left: &n{val: 35}}},
    				right: &n{val: 70, left: &n{val: 60, right: &n{val: 65}}, right: &n{val: 90, left: &n{val: 80}}},
    			},
    			true,
    		},
    		{
    			"deep, one bad leaf",
    			&n{val: 50,
    				left:  &n{val: 30, left: &n{val: 10, right: &n{val: 20}}, right: &n{val: 40, left: &n{val: 35}}},
    				right: &n{val: 70, left: &n{val: 60, right: &n{val: 71}}, right: &n{val: 90, left: &n{val: 80}}},
    			},
    			false,
    		},
    		{"extreme ints", &n{val: 0, left: &n{val: math.MinInt}, right: &n{val: math.MaxInt}}, true},
    		{"left chain", &n{val: 3, left: &n{val: 2, left: &n{val: 1}}}, true},
    		{"zigzag", &n{val: 10, left: &n{val: 5, right: &n{val: 8, left: &n{val: 6, right: &n{val: 7}}}}}, true},
    		{"zigzag, bad", &n{val: 10, left: &n{val: 5, right: &n{val: 8, left: &n{val: 6, right: &n{val: 9, right: &n{val: 11}}}}}}, false},
    	}
    	for _, tt := range tests {
    		if got := validBST(tt.root); got != tt.want {
    			t.Errorf("%s: validBST = %v, want %v", tt.name, got, tt.want)
    		}
    	}
    }

    func TestValidBSTStrings(t *testing.T) {
    	type s = node[string]
    	good := &s{val: "mira", left: &s{val: "ava", right: &s{val: "kai"}}, right: &s{val: "zed"}}
    	if !validBST(good) {
    		t.Errorf("validBST(mira(ava(-, kai), zed)) = false, want true")
    	}
    	bad := &s{val: "mira", left: &s{val: "ava", right: &s{val: "nova"}}, right: &s{val: "zed"}}
    	if validBST(bad) {
    		t.Errorf("validBST(mira(ava(-, nova), zed)) = true, want false: nova is in mira's left subtree")
    	}
    	empty := &s{val: "", right: &s{val: "a"}}
    	if !validBST(empty) {
    		t.Errorf(`validBST(""(-, a)) = false, want true: the empty string is a real value`)
    	}
    }
---

A tree-editing tool let a designer rearrange the leaderboard's binary search tree
by hand, and now lookups are failing. Before the tree goes live again, the server
must **audit** it.

Write `validBST(root)`. It reports whether the tree is a valid binary search tree:
for **every** node, all values in its left subtree are **smaller** than its value,
and all values in its right subtree are **larger**. Duplicates are not allowed. An
empty tree is valid.

`validBST` is generic, so it must work for any `cmp.Ordered` type: scores,
usernames, and so on.

## Examples

```
    50               50
   /  \             /  \
  30   70         30    70
                    \
                     60
```

The left tree is valid. The right one is **not**: 60 sits in 50's left subtree
but is bigger than 50, even though it's correctly to the right of its parent 30.

## Constraints

- Up to 100,000 nodes. Visit each node once: O(n).
- Values can be anything `T` can hold, including `math.MinInt`, `math.MaxInt` and
  the empty string, so don't rely on "impossible" sentinel values.
