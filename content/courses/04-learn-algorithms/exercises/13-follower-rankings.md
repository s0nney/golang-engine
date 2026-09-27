---
title: Follower Rankings
difficulty: hard
after: sorting-algorithms
hints:
  - 'Counting with a loop over every account in each query is O(n) per query: 200,000 × 200,000 = 40 billion steps. Do the expensive work **once**, in `NewRankings`.'
  - 'Store a **sorted copy** of the counts (`slices.Clone`, then `slices.Sort`). In a sorted slice, "how many are below x" is just the position where `x` would be inserted, and binary search finds that in O(log n).'
  - '`slices.BinarySearch(sorted, x)` returns the index of the first element `>= x`. The number of accounts with more than `f` followers is `len(sorted)` minus the index of the first element `>= f+1`. `CountBetween(lo, hi)` is (first index `>= hi+1`) − (first index `>= lo`).'
exercise:
  starter: |
    package main

    import "fmt"

    // Rankings answers ranking questions about a fixed set of follower counts.
    type Rankings struct {
    	// your fields here
    }

    // NewRankings prepares rankings for the given follower counts.
    // It must not modify followers.
    func NewRankings(followers []int) *Rankings {
    	return &Rankings{}
    }

    // Rank returns the leaderboard position an account with f followers
    // would have: 1 + the number of accounts with strictly more than f.
    func (r *Rankings) Rank(f int) int {
    	return 0
    }

    // CountBetween returns how many accounts have between lo and hi
    // followers, inclusive. It returns 0 if lo > hi.
    func (r *Rankings) CountBetween(lo, hi int) int {
    	return 0
    }

    func main() {
    	r := NewRankings([]int{500, 20, 9000, 500, 75})
    	fmt.Println(r.Rank(9000), r.Rank(500), r.Rank(100), r.Rank(0)) // want 1 2 4 6
    	fmt.Println(r.CountBetween(50, 500))                           // want 3
    }
  solution: |
    package main

    import (
    	"fmt"
    	"slices"
    )

    // Rankings answers ranking questions about a fixed set of follower counts.
    type Rankings struct {
    	sorted []int
    }

    // NewRankings prepares rankings for the given follower counts.
    // It must not modify followers.
    func NewRankings(followers []int) *Rankings {
    	sorted := slices.Clone(followers)
    	slices.Sort(sorted)
    	return &Rankings{sorted: sorted}
    }

    // atLeast returns the index of the first count >= x.
    func (r *Rankings) atLeast(x int) int {
    	i, _ := slices.BinarySearch(r.sorted, x)
    	return i
    }

    // Rank returns the leaderboard position an account with f followers
    // would have: 1 + the number of accounts with strictly more than f.
    func (r *Rankings) Rank(f int) int {
    	return 1 + len(r.sorted) - r.atLeast(f+1)
    }

    // CountBetween returns how many accounts have between lo and hi
    // followers, inclusive. It returns 0 if lo > hi.
    func (r *Rankings) CountBetween(lo, hi int) int {
    	if lo > hi {
    		return 0
    	}
    	return r.atLeast(hi+1) - r.atLeast(lo)
    }

    func main() {
    	r := NewRankings([]int{500, 20, 9000, 500, 75})
    	fmt.Println(r.Rank(9000), r.Rank(500), r.Rank(100), r.Rank(0))
    	fmt.Println(r.CountBetween(50, 500))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    	"time"
    )

    func TestRank(t *testing.T) {
    	counts := []int{500, 20, 9000, 500, 75}
    	orig := slices.Clone(counts)
    	r := NewRankings(counts)
    	if !slices.Equal(counts, orig) {
    		t.Fatalf("NewRankings(%v) modified its input to %v", orig, counts)
    	}
    	for _, tt := range []struct{ f, want int }{
    		{9000, 1}, {10_000, 1}, {8999, 2}, {500, 2}, {499, 4}, {100, 4}, {75, 4}, {74, 5}, {20, 5}, {0, 6}, {-10, 6},
    	} {
    		if got := r.Rank(tt.f); got != tt.want {
    			t.Errorf("NewRankings(%v).Rank(%d) = %d, want %d", orig, tt.f, got, tt.want)
    		}
    	}
    }

    func TestCountBetween(t *testing.T) {
    	counts := []int{500, 20, 9000, 500, 75, -3}
    	r := NewRankings(counts)
    	for _, tt := range []struct{ lo, hi, want int }{
    		{50, 500, 3}, {500, 500, 2}, {501, 8999, 0}, {0, 10_000, 5}, {-100, 10_000, 6},
    		{-3, -3, 1}, {21, 74, 0}, {20, 20, 1}, {9000, 9000, 1}, {600, 100, 0},
    	} {
    		if got := r.CountBetween(tt.lo, tt.hi); got != tt.want {
    			t.Errorf("NewRankings(%v).CountBetween(%d, %d) = %d, want %d", counts, tt.lo, tt.hi, got, tt.want)
    		}
    	}
    }

    func TestEmptyRankings(t *testing.T) {
    	r := NewRankings(nil)
    	if got := r.Rank(42); got != 1 {
    		t.Errorf("NewRankings(nil).Rank(42) = %d, want 1", got)
    	}
    	if got := r.CountBetween(0, 100); got != 0 {
    		t.Errorf("NewRankings(nil).CountBetween(0, 100) = %d, want 0", got)
    	}
    }

    func TestRankingsLarge(t *testing.T) {
    	n := 200_000
    	counts := make([]int, n)
    	for i := range counts {
    		counts[i] = (i * 7919) % n // every value 0..n-1 exactly once, shuffled
    	}
    	type result struct{ rankSum, betweenSum int }
    	done := make(chan result, 1)
    	go func() {
    		r := NewRankings(counts)
    		var res result
    		for q := range n {
    			res.rankSum += r.Rank(q)
    			res.betweenSum += r.CountBetween(q-10, q+10)
    		}
    		done <- res
    	}()
    	select {
    	case got := <-done:
    		// Rank(q) = n - q, and CountBetween(q-10, q+10) counts the values
    		// of 0..n-1 inside that range.
    		want := result{}
    		for q := range n {
    			want.rankSum += n - q
    			want.betweenSum += min(n-1, q+10) - max(0, q-10) + 1
    		}
    		if got != want {
    			t.Errorf("%d Rank and CountBetween queries over %d accounts: sums = %+v, want %+v", n, n, got, want)
    		}
    	case <-time.After(time.Second):
    		t.Fatalf("%d queries over %d accounts took over a second: sort once, then binary search each query", 2*n, n)
    	}
    }
---

The Clout leaderboard has hundreds of thousands of accounts, and the profile page
asks it questions constantly: *"If I had 5,000 followers, what rank would I be?"*
and *"How many accounts have between 1k and 10k followers?"*

The follower counts don't change while the page is open, so you can do some work
up front to make every question fast. Implement `Rankings`:

- `NewRankings(followers)` prepares the rankings. It must **not** modify the
  caller's slice.
- `Rank(f)` returns the position an account with `f` followers would have on the
  leaderboard: 1 + the number of accounts with **strictly more** than `f`.
  Accounts with the same count share a rank.
- `CountBetween(lo, hi)` returns how many accounts have between `lo` and `hi`
  followers, **inclusive**, or 0 if `lo > hi`.

## Example

```go
r := NewRankings([]int{500, 20, 9000, 500, 75})
r.Rank(9000)            // 1
r.Rank(500)             // 2: only the 9000 account has more
r.Rank(100)             // 4: 9000, 500 and 500 have more
r.Rank(0)               // 6: everyone has more
r.CountBetween(50, 500) // 3: 75, 500, 500
```

## Constraints

- Up to 200,000 accounts and 400,000 queries. Counts are any `int`, including
  negative (the tests throw in a test account at -3) and duplicates.
- Scanning every account for each query is O(n) per query: 80 billion steps for
  the performance test, which runs under a one-second limit. Aim for
  O(n log n) setup and **O(log n) per query**.
