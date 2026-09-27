---
title: Path Spread
difficulty: easy
after: red-black-trees
hints:
  - 'Think recursively, like `height`. An empty tree has no paths with nodes on them, so it returns `0, 0`. For any other node, ask both children for their spread first.'
  - 'A path from `n` goes through `n` (that''s the `1 +`) and then down **one** of the children. So the shortest is `1 + min(leftShortest, rightShortest)` and the longest is `1 + max(leftLongest, rightLongest)`. A node with a nil child has a shortest path of just 1.'
exercise:
  starter: |
    package main

    import (
    	"cmp"
    	"fmt"
    )

    type rbNode[T cmp.Ordered] struct {
    	val         T
    	red         bool
    	left, right *rbNode[T]
    }

    // pathSpread returns the number of nodes on the shortest and on the
    // longest path from n down to a nil child. An empty tree returns 0, 0.
    func pathSpread[T cmp.Ordered](n *rbNode[T]) (shortest, longest int) {
    	// Base case: nil. Otherwise combine the spreads of n.left and n.right.
    	return 0, 0
    }

    func main() {
    	tree := &rbNode[int]{val: 40,
    		left:  &rbNode[int]{val: 20, red: true, left: &rbNode[int]{val: 10}, right: &rbNode[int]{val: 30}},
    		right: &rbNode[int]{val: 60, left: &rbNode[int]{val: 50, red: true}},
    	}
    	fmt.Println(pathSpread(tree)) // want 2 3
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    )

    type rbNode[T cmp.Ordered] struct {
    	val         T
    	red         bool
    	left, right *rbNode[T]
    }

    func pathSpread[T cmp.Ordered](n *rbNode[T]) (shortest, longest int) {
    	if n == nil {
    		return 0, 0
    	}
    	ls, ll := pathSpread(n.left)
    	rs, rl := pathSpread(n.right)
    	return 1 + min(ls, rs), 1 + max(ll, rl)
    }

    func main() {
    	tree := &rbNode[int]{val: 40,
    		left:  &rbNode[int]{val: 20, red: true, left: &rbNode[int]{val: 10}, right: &rbNode[int]{val: 30}},
    		right: &rbNode[int]{val: 60, left: &rbNode[int]{val: 50, red: true}},
    	}
    	fmt.Println(pathSpread(tree))
    }
  tests: |
    package main

    import "testing"

    type node = rbNode[string]

    func TestPathSpread(t *testing.T) {
    	chain := &node{val: "a", right: &node{val: "b", red: true, right: &node{val: "c", right: &node{val: "d"}}}}
    	tests := []struct {
    		name              string
    		root              *node
    		shortest, longest int
    	}{
    		{"empty tree", nil, 0, 0},
    		{"single node", &node{val: "mira"}, 1, 1},
    		{"root with only a left child", &node{val: "b", left: &node{val: "a", red: true}}, 1, 2},
    		{"full tree of 3", &node{val: "b", left: &node{val: "a"}, right: &node{val: "c"}}, 2, 2},
    		{"lopsided chain", chain, 1, 4},
    		{
    			"valid red-black tree",
    			&node{val: "m",
    				left: &node{val: "f", left: &node{val: "c"}, right: &node{val: "h"}},
    				right: &node{val: "t", red: true,
    					left:  &node{val: "p", left: &node{val: "n"}, right: &node{val: "r"}},
    					right: &node{val: "x", left: &node{val: "w"}, right: &node{val: "y", right: &node{val: "z", red: true}}},
    				},
    			},
    			3, 5,
    		},
    		{
    			"deep on the left only",
    			&node{val: "e",
    				left:  &node{val: "c", left: &node{val: "b", left: &node{val: "a"}}, right: &node{val: "d"}},
    				right: &node{val: "g", left: &node{val: "f"}, right: &node{val: "h"}},
    			},
    			3, 4,
    		},
    	}
    	for _, tt := range tests {
    		s, l := pathSpread(tt.root)
    		if s != tt.shortest || l != tt.longest {
    			t.Errorf("%s: pathSpread = %d, %d, want %d, %d", tt.name, s, l, tt.shortest, tt.longest)
    		}
    	}
    }

    func TestPathSpreadInts(t *testing.T) {
    	root := &rbNode[int]{val: 40,
    		left:  &rbNode[int]{val: 20, red: true, left: &rbNode[int]{val: 10}, right: &rbNode[int]{val: 30}},
    		right: &rbNode[int]{val: 60, left: &rbNode[int]{val: 50, red: true}},
    	}
    	if s, l := pathSpread(root); s != 2 || l != 3 {
    		t.Errorf("pathSpread(the example tree) = %d, %d, want 2, 3", s, l)
    	}
    }
---

Chapter 2 proved that in a red-black tree the longest path from the root down to
a nil leaf is **at most twice** as long as the shortest one. The studio's
tree-debugging tool wants to show that spread for any tree, balanced or not.

Complete `pathSpread(n)`. It returns two numbers: how many nodes lie on the
**shortest** path from `n` down to a nil child, and how many lie on the
**longest** one. An empty tree returns `0, 0`. Colours don't affect the answer.
They're only there so you can check the 2× guarantee on real red-black trees.

## Example

```
        40
       /  \
    (20)   60        (x) = red
    /  \   /
   10  30 (50)
```

The shortest path is 40 → 60 (60 has no right child): 2 nodes. The longest is
40 → 20 → 10, or any other path of 3 nodes. So `pathSpread(tree)` returns `2, 3`.

A chain `a → b → c → d` going right returns `1, 4`: the root's missing left child
ends a path after just one node.

## Constraints

- O(n): visit each node once.
