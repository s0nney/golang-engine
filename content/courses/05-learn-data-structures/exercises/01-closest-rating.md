---
title: Closest Rating
difficulty: easy
after: binary-trees
hints:
  - 'You don''t need to visit every node. At each node, compare `target` with its score and go left or right, exactly as `contains` does. The closest score is somewhere on that one path.'
  - 'Keep a `best` score as you walk. At every node, replace `best` if this node is strictly closer to `target`, or equally close but lower. If you land on `target` exactly, you can return it straight away.'
exercise:
  starter: |
    package main

    import "fmt"

    type Node struct {
    	Rating      int
    	Left, Right *Node
    }

    func insert(n *Node, rating int) *Node {
    	if n == nil {
    		return &Node{Rating: rating}
    	}
    	if rating < n.Rating {
    		n.Left = insert(n.Left, rating)
    	} else if rating > n.Rating {
    		n.Right = insert(n.Right, rating)
    	}
    	return n
    }

    // closestRating returns the rating in the BST rooted at root that is
    // closest to target. If two ratings are equally close, it returns the
    // lower one. For an empty tree it returns 0, false.
    func closestRating(root *Node, target int) (int, bool) {
    	// 1. Handle the empty tree.
    	// 2. Walk down from the root like a search, remembering the closest
    	//    rating seen so far.
    	return 0, false
    }

    func main() {
    	var root *Node
    	for _, r := range []int{1500, 1200, 1800, 1000, 1400, 1600, 2000} {
    		root = insert(root, r)
    	}
    	fmt.Println(closestRating(root, 1450)) // want 1400 true
    	fmt.Println(closestRating(root, 1690)) // want 1600 true
    }
  solution: |
    package main

    import "fmt"

    type Node struct {
    	Rating      int
    	Left, Right *Node
    }

    func insert(n *Node, rating int) *Node {
    	if n == nil {
    		return &Node{Rating: rating}
    	}
    	if rating < n.Rating {
    		n.Left = insert(n.Left, rating)
    	} else if rating > n.Rating {
    		n.Right = insert(n.Right, rating)
    	}
    	return n
    }

    func abs(x int) int { return max(x, -x) }

    func closestRating(root *Node, target int) (int, bool) {
    	if root == nil {
    		return 0, false
    	}
    	best := root.Rating
    	for n := root; n != nil; {
    		d, bd := abs(n.Rating-target), abs(best-target)
    		if d < bd || d == bd && n.Rating < best {
    			best = n.Rating
    		}
    		switch {
    		case target < n.Rating:
    			n = n.Left
    		case target > n.Rating:
    			n = n.Right
    		default:
    			return n.Rating, true
    		}
    	}
    	return best, true
    }

    func main() {
    	var root *Node
    	for _, r := range []int{1500, 1200, 1800, 1000, 1400, 1600, 2000} {
    		root = insert(root, r)
    	}
    	fmt.Println(closestRating(root, 1450))
    	fmt.Println(closestRating(root, 1690))
    }
  tests: |
    package main

    import (
    	"testing"
    	"time"
    )

    func build(ratings ...int) *Node {
    	var root *Node
    	for _, r := range ratings {
    		root = insert(root, r)
    	}
    	return root
    }

    func TestClosestRating(t *testing.T) {
    	ladder := []int{1500, 1200, 1800, 1000, 1400, 1600, 2000}
    	tests := []struct {
    		ratings []int
    		target  int
    		want    int
    		wantOK  bool
    	}{
    		{nil, 1500, 0, false},
    		{[]int{900}, 5000, 900, true},
    		{[]int{900}, -5000, 900, true},
    		{ladder, 1450, 1400, true},
    		{ladder, 1690, 1600, true},
    		{ladder, 1700, 1600, true}, // tie: the lower rating wins
    		{ladder, 1300, 1200, true}, // tie deep in the tree
    		{ladder, 1800, 1800, true},
    		{ladder, 1801, 1800, true},
    		{ladder, 0, 1000, true},
    		{ladder, 9999, 2000, true},
    		{ladder, 1549, 1500, true},
    		{ladder, 1551, 1600, true},
    		{[]int{0, -300, 300, -100}, -180, -100, true},
    	}
    	for _, tt := range tests {
    		got, ok := closestRating(build(tt.ratings...), tt.target)
    		if got != tt.want || ok != tt.wantOK {
    			t.Errorf("closestRating(tree of %v, %d) = %d, %v, want %d, %v", tt.ratings, tt.target, got, ok, tt.want, tt.wantOK)
    		}
    	}
    }

    func TestClosestRatingFollowsOnePath(t *testing.T) {
    	// A balanced tree with 2^20 - 1 ratings, built directly (no insert calls).
    	var build func(lo, hi int) *Node
    	build = func(lo, hi int) *Node {
    		if lo > hi {
    			return nil
    		}
    		mid := (lo + hi) / 2
    		return &Node{Rating: mid * 10, Left: build(lo, mid-1), Right: build(mid+1, hi)}
    	}
    	root := build(1, 1<<20-1)
    	start := time.Now()
    	for i := range 20_000 {
    		target := i*523 + 7
    		want := (target + 4) / 10 * 10 // nearest multiple of 10, lower on ties
    		want = max(10, want)
    		if got, ok := closestRating(root, target); got != want || !ok {
    			t.Fatalf("closestRating(big tree, %d) = %d, %v, want %d, true", target, got, ok, want)
    		}
    	}
    	if d := time.Since(start); d > time.Second {
    		t.Errorf("20,000 lookups in a million-node tree took %v: follow one root-to-leaf path instead of visiting every node", d)
    	}
    }
---

The matchmaker wants an opponent whose rating is as close as possible to yours.
Every online player's rating is stored in a binary search tree.

Complete `closestRating(root, target)`. It returns the rating in the tree that is
closest to `target`, and `true`. If two ratings are equally close, return the
**lower** one. For an empty tree (`root == nil`), return `0, false`.

## Examples

```
tree: 1500, 1200, 1800, 1000, 1400, 1600, 2000

closestRating(root, 1450) // 1400, true
closestRating(root, 1700) // 1600, true (1600 and 1800 tie; lower wins)
closestRating(root, 9999) // 2000, true
closestRating(nil, 1500)  // 0, false
```

## Constraints

- Ratings in the tree are unique and may be negative.
- Aim for **O(h)**, where h is the tree's height: follow a single path from the root,
  like a search, instead of visiting every node. One test runs 20,000 lookups in a
  tree with a million ratings.
