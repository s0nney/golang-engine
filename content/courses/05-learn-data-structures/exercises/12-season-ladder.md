---
title: Season Ladder
difficulty: hard
after: red-black-trees
hints:
  - 'A sorted slice makes `Rank` and `Top` easy, but every `Add` shifts up to n scores: O(n) each, O(n²) in total. A plain BST is O(log n) only while it''s balanced, and a season''s scores arrive in sorted runs, which turn it into a linked list. You need a tree that **stays balanced**: the red-black tree from this chapter.'
  - 'Augment each node: store its `score`, a `count` of players with that score, and a `size` = total count in its whole subtree. Then "how many scores are greater than s?" is one walk down: whenever you go **left** from node n, everything in n''s right subtree, plus n itself, is greater. `Top(k)` walks down the same way, comparing k with `size(n.right)`.'
  - 'Keep `size` correct: on `Add`, every node you pass on the way down gains 1 (even if the score already exists and you only bump its `count`). In `rotateLeft` / `rotateRight`, only the two rotated nodes change subtrees, so recompute them bottom-up: first the node that moved down, then the one that moved up. Recolouring doesn''t change sizes.'
exercise:
  starter: |
    package main

    import "fmt"

    // Ladder holds every player's season score, duplicates included.
    type Ladder struct {
    	// your fields here
    }

    func NewLadder() *Ladder {
    	return &Ladder{}
    }

    // Add records one more player with this score.
    func (l *Ladder) Add(score int) {
    }

    // Len returns how many scores have been added.
    func (l *Ladder) Len() int {
    	return 0
    }

    // Rank returns 1 + the number of stored scores strictly greater than score.
    func (l *Ladder) Rank(score int) int {
    	return 0
    }

    // Top returns the k-th highest stored score (k = 1 is the best), or
    // 0, false if k is out of range.
    func (l *Ladder) Top(k int) (int, bool) {
    	return 0, false
    }

    func main() {
    	l := NewLadder()
    	for _, s := range []int{1200, 1500, 900, 1500, 2100} {
    		l.Add(s)
    	}
    	fmt.Println(l.Len(), l.Rank(1500), l.Rank(1000)) // want 5 2 5
    	fmt.Println(l.Top(1))                            // want 2100 true
    	fmt.Println(l.Top(3))                            // want 1500 true
    }
  solution: |
    package main

    import "fmt"

    type node struct {
    	score, count, size  int
    	red                 bool
    	left, right, parent *node
    }

    // Ladder is a red-black tree of distinct scores. Each node also stores how
    // many players have that score (count) and the total count in its subtree
    // (size), which is what makes Rank and Top O(log n).
    type Ladder struct {
    	root *node
    }

    func NewLadder() *Ladder { return &Ladder{} }

    func size(n *node) int {
    	if n == nil {
    		return 0
    	}
    	return n.size
    }

    func isRed(n *node) bool { return n != nil && n.red }

    func (n *node) resize() { n.size = size(n.left) + size(n.right) + n.count }

    func (l *Ladder) replaceChild(parent, old, child *node) {
    	switch {
    	case parent == nil:
    		l.root = child
    	case parent.left == old:
    		parent.left = child
    	default:
    		parent.right = child
    	}
    	if child != nil {
    		child.parent = parent
    	}
    }

    func (l *Ladder) rotateLeft(x *node) {
    	y := x.right
    	x.right = y.left
    	if y.left != nil {
    		y.left.parent = x
    	}
    	l.replaceChild(x.parent, x, y)
    	y.left = x
    	x.parent = y
    	x.resize()
    	y.resize()
    }

    func (l *Ladder) rotateRight(x *node) {
    	y := x.left
    	x.left = y.right
    	if y.right != nil {
    		y.right.parent = x
    	}
    	l.replaceChild(x.parent, x, y)
    	y.right = x
    	x.parent = y
    	x.resize()
    	y.resize()
    }

    func (l *Ladder) Add(score int) {
    	var parent *node
    	for n := l.root; n != nil; {
    		n.size++
    		if score == n.score {
    			n.count++
    			return
    		}
    		parent = n
    		if score < n.score {
    			n = n.left
    		} else {
    			n = n.right
    		}
    	}
    	z := &node{score: score, count: 1, size: 1, red: true, parent: parent}
    	switch {
    	case parent == nil:
    		l.root = z
    	case score < parent.score:
    		parent.left = z
    	default:
    		parent.right = z
    	}
    	l.fixInsert(z)
    }

    func (l *Ladder) fixInsert(z *node) {
    	for isRed(z.parent) {
    		p, g := z.parent, z.parent.parent
    		if p == g.left {
    			if u := g.right; isRed(u) {
    				p.red, u.red, g.red = false, false, true
    				z = g
    				continue
    			}
    			if z == p.right {
    				l.rotateLeft(p)
    				z, p = p, z
    			}
    			p.red, g.red = false, true
    			l.rotateRight(g)
    		} else {
    			if u := g.left; isRed(u) {
    				p.red, u.red, g.red = false, false, true
    				z = g
    				continue
    			}
    			if z == p.left {
    				l.rotateRight(p)
    				z, p = p, z
    			}
    			p.red, g.red = false, true
    			l.rotateLeft(g)
    		}
    	}
    	l.root.red = false
    }

    func (l *Ladder) Len() int { return size(l.root) }

    func (l *Ladder) Rank(score int) int {
    	greater := 0
    	for n := l.root; n != nil; {
    		switch {
    		case score < n.score:
    			greater += n.count + size(n.right)
    			n = n.left
    		case score > n.score:
    			n = n.right
    		default:
    			return greater + size(n.right) + 1
    		}
    	}
    	return greater + 1
    }

    func (l *Ladder) Top(k int) (int, bool) {
    	if k < 1 || k > l.Len() {
    		return 0, false
    	}
    	n := l.root
    	for {
    		r := size(n.right)
    		switch {
    		case k <= r:
    			n = n.right
    		case k <= r+n.count:
    			return n.score, true
    		default:
    			k -= r + n.count
    			n = n.left
    		}
    	}
    }

    func main() {
    	l := NewLadder()
    	for _, s := range []int{1200, 1500, 900, 1500, 2100} {
    		l.Add(s)
    	}
    	fmt.Println(l.Len(), l.Rank(1500), l.Rank(1000)) // 5 2 5
    	fmt.Println(l.Top(1))
    	fmt.Println(l.Top(3))
    }
  tests: |
    package main

    import (
    	"slices"
    	"sort"
    	"testing"
    	"time"
    )

    func TestLadderExample(t *testing.T) {
    	l := NewLadder()
    	if n, r := l.Len(), l.Rank(100); n != 0 || r != 1 {
    		t.Errorf("empty ladder: Len() = %d, Rank(100) = %d, want 0 and 1", n, r)
    	}
    	if v, ok := l.Top(1); ok {
    		t.Errorf("empty ladder: Top(1) = %d, true, want 0, false", v)
    	}
    	for _, s := range []int{1200, 1500, 900, 1500, 2100} {
    		l.Add(s)
    	}
    	if n := l.Len(); n != 5 {
    		t.Errorf("after adding 1200, 1500, 900, 1500, 2100: Len() = %d, want 5", n)
    	}
    	for _, tt := range []struct{ score, want int }{
    		{2100, 1}, {1500, 2}, {1200, 4}, {900, 5},
    		{5000, 1}, {2000, 2}, {1000, 5}, {0, 6},
    	} {
    		if got := l.Rank(tt.score); got != tt.want {
    			t.Errorf("scores 900 1200 1500 1500 2100: Rank(%d) = %d, want %d", tt.score, got, tt.want)
    		}
    	}
    	for k, want := range []int{2100, 1500, 1500, 1200, 900} {
    		if got, ok := l.Top(k + 1); got != want || !ok {
    			t.Errorf("scores 900 1200 1500 1500 2100: Top(%d) = %d, %v, want %d, true", k+1, got, ok, want)
    		}
    	}
    	for _, k := range []int{0, -1, 6} {
    		if got, ok := l.Top(k); ok {
    			t.Errorf("Top(%d) with 5 scores = %d, true, want 0, false", k, got)
    		}
    	}
    }

    func TestLadderNegativeAndTwoLadders(t *testing.T) {
    	a, b := NewLadder(), NewLadder()
    	for _, s := range []int{-5, 0, -5, -20} {
    		a.Add(s)
    	}
    	if r := a.Rank(-5); r != 2 {
    		t.Errorf("scores -20 -5 -5 0: Rank(-5) = %d, want 2", r)
    	}
    	if v, ok := a.Top(4); v != -20 || !ok {
    		t.Errorf("scores -20 -5 -5 0: Top(4) = %d, %v, want -20, true", v, ok)
    	}
    	if b.Len() != 0 {
    		t.Errorf("two ladders share state: a fresh ladder has Len() = %d", b.Len())
    	}
    }

    // TestLadderAgainstSortedSlice compares every answer with a simple (slow)
    // sorted-slice ladder, over a mix of rising, falling and repeated scores.
    func TestLadderAgainstSortedSlice(t *testing.T) {
    	l := NewLadder()
    	var ref []int // ascending
    	var x uint64 = 99
    	next := func(n int) int {
    		x = x*6364136223846793005 + 1442695040888963407
    		return int(x>>33) % n
    	}
    	for i := range 3000 {
    		var s int
    		switch i % 3 {
    		case 0:
    			s = i // rising
    		case 1:
    			s = -i // falling
    		default:
    			s = next(200) // lots of repeats
    		}
    		l.Add(s)
    		ref = slices.Insert(ref, sort.SearchInts(ref, s), s)
    		q := next(6001) - 3000
    		want := len(ref) - sort.SearchInts(ref, q+1) + 1
    		if got := l.Rank(q); got != want {
    			t.Fatalf("after %d adds (last: %d): Rank(%d) = %d, want %d", i+1, s, q, got, want)
    		}
    		k := 1 + next(len(ref))
    		if got, ok := l.Top(k); got != ref[len(ref)-k] || !ok {
    			t.Fatalf("after %d adds (last: %d): Top(%d) = %d, %v, want %d, true", i+1, s, k, got, ok, ref[len(ref)-k])
    		}
    	}
    	if l.Len() != 3000 {
    		t.Errorf("after 3000 adds: Len() = %d, want 3000", l.Len())
    	}
    }

    func TestLadderLarge(t *testing.T) {
    	done := make(chan int, 1)
    	go func() {
    		l := NewLadder()
    		sum := 0
    		var x uint64 = 7
    		next := func() int {
    			x = x*6364136223846793005 + 1442695040888963407
    			return int(x >> 33)
    		}
    		for i := range 200_000 {
    			var s int
    			switch {
    			case i%7 == 6:
    				s = 1_000_000 + (i/2)%50_000 // a repeat of an earlier score
    			case i < 100_000:
    				s = 1_000_000 + i // the season's first half: scores climb
    			default:
    				s = 1_000_000 - (i - 100_000) // second half: late joiners, lower scores
    			}
    			l.Add(s)
    			sum += l.Rank(850_000 + next()%300_000)
    			if i%10 == 0 {
    				v, ok := l.Top(1 + next()%l.Len())
    				if !ok {
    					done <- -1
    					return
    				}
    				sum += v
    			}
    		}
    		done <- sum*31 + l.Len()
    	}()
    	select {
    	case got := <-done:
    		if got != 995513249146 {
    			t.Errorf("200,000 Adds with Rank and Top queries: checksum = %d, want 995513249146 (some Rank or Top answer was wrong)", got)
    		}
    	case <-time.After(time.Second):
    		t.Fatalf("200,000 Adds (in sorted runs) with Rank and Top queries took over a second: use a balanced tree so every operation is O(log n)")
    	}
    }
---

The season ladder shows every player their **rank** live, and the "view the top"
button jumps to any position on the board. Scores are added all season long:
200,000 of them, often in long sorted runs as players climb (or as late joiners
arrive with lower scores).

Implement `Ladder`:

- `NewLadder()` returns an empty ladder.
- `Add(score)` records one more player with `score`. Several players can have the
  same score.
- `Len()` returns how many scores have been added.
- `Rank(score)` returns 1 + the number of stored scores **strictly greater** than
  `score`. Players with equal scores share a rank. `score` doesn't have to be
  stored: "what rank would 1750 get?" is a fair question.
- `Top(k)` returns the k-th highest stored score, counting duplicates (`k = 1` is
  the best), and `true`. If `k < 1` or `k > Len()`, return `0, false`.

## Example

```go
l := NewLadder()
for _, s := range []int{1200, 1500, 900, 1500, 2100} {
	l.Add(s)
}
l.Rank(1500) // 2: only 2100 is higher
l.Rank(1000) // 5: 2100, 1500, 1500 and 1200 are higher
l.Top(1)     // 2100, true
l.Top(3)     // 1500, true (the second 1500)
l.Top(6)     // 0, false
```

## Constraints

- Up to 200,000 `Add` calls, each followed by a `Rank` query, with a `Top` query
  every 10 adds.
- `Add`, `Rank` and `Top` must be **O(log n)** in the worst case. The performance
  test adds long sorted runs, so a plain BST degenerates into a list, and a sorted
  slice shifts up to 200,000 scores on each insert. Both take many seconds; the
  limit is one.
- Standard library only. `Len` should be O(1).
