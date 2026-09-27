---
title: Fair Teams
difficulty: medium
after: p-vs-np
hints:
  - '`teamDiff` is the **verifier**: check the certificate is well-formed first. It needs exactly `len(ratings)/2` indexes, each in range, none repeated (a `map[int]bool` or a `[]bool` of size n catches repeats). Then the other team is everyone else, so its total is `sum(ratings) - sum(team A)`.'
  - '`bestSplit` is the **search**. With n ≤ 20 you can afford to try every subset: loop `mask` from 0 to `1<<n - 1`, where bit `i` set means "player i is on team A". Skip masks whose `bits.OnesCount(uint(mask))` isn''t n/2, and keep the smallest difference.'
  - 'Every split is counted twice (team A and team B swapped), so you could fix player 0 on team A to halve the work. Not needed to pass, but a nice touch.'
exercise:
  starter: |
    package main

    import "fmt"

    func teamDiff(ratings []int, teamA []int) (int, bool) {
    	return 0, false
    }

    func bestSplit(ratings []int) int {
    	return -1
    }

    func main() {
    	ratings := []int{1500, 1300, 1200, 1400}
    	fmt.Println(teamDiff(ratings, []int{0, 1})) // want 200 true
    	fmt.Println(bestSplit(ratings))             // want 0
    }
  solution: |
    package main

    import (
    	"fmt"
    	"math/bits"
    )

    func abs(x int) int { return max(x, -x) }

    func teamDiff(ratings []int, teamA []int) (int, bool) {
    	if len(teamA) != len(ratings)/2 {
    		return 0, false
    	}
    	used := make([]bool, len(ratings))
    	total, sumA := 0, 0
    	for _, r := range ratings {
    		total += r
    	}
    	for _, i := range teamA {
    		if i < 0 || i >= len(ratings) || used[i] {
    			return 0, false
    		}
    		used[i] = true
    		sumA += ratings[i]
    	}
    	return abs(total - 2*sumA), true
    }

    func bestSplit(ratings []int) int {
    	n := len(ratings)
    	total := 0
    	for _, r := range ratings {
    		total += r
    	}
    	best := -1
    	for mask := range 1 << n {
    		if bits.OnesCount(uint(mask)) != n/2 {
    			continue
    		}
    		sumA := 0
    		for i, r := range ratings {
    			if mask&(1<<i) != 0 {
    				sumA += r
    			}
    		}
    		if d := abs(total - 2*sumA); best == -1 || d < best {
    			best = d
    		}
    	}
    	return best
    }

    func main() {
    	ratings := []int{1500, 1300, 1200, 1400}
    	fmt.Println(teamDiff(ratings, []int{0, 1}))
    	fmt.Println(bestSplit(ratings))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    	"time"
    )

    func TestTeamDiff(t *testing.T) {
    	ratings := []int{1500, 1300, 1200, 1400}
    	tests := []struct {
    		name    string
    		ratings []int
    		teamA   []int
    		want    int
    		ok      bool
    	}{
    		{"example", ratings, []int{0, 1}, 200, true},
    		{"perfect split", ratings, []int{0, 2}, 0, true},
    		{"order doesn't matter", ratings, []int{3, 1}, 0, true},
    		{"worst split", ratings, []int{2, 1}, 400, true},
    		{"too few players", ratings, []int{0}, 0, false},
    		{"too many players", ratings, []int{0, 1, 2}, 0, false},
    		{"same player twice", ratings, []int{2, 2}, 0, false},
    		{"index out of range", ratings, []int{0, 4}, 0, false},
    		{"negative index", ratings, []int{-1, 0}, 0, false},
    		{"nobody at all", nil, nil, 0, true},
    		{"two players", []int{900, 1400}, []int{1}, 500, true},
    	}
    	for _, tt := range tests {
    		in := slices.Clone(tt.teamA)
    		got, ok := teamDiff(tt.ratings, in)
    		if got != tt.want || ok != tt.ok {
    			t.Errorf("%s: teamDiff(%v, %v) = %d, %v, want %d, %v", tt.name, tt.ratings, tt.teamA, got, ok, tt.want, tt.ok)
    		}
    		if !slices.Equal(in, tt.teamA) {
    			t.Errorf("%s: teamDiff changed teamA to %v", tt.name, in)
    		}
    	}
    }

    func TestBestSplit(t *testing.T) {
    	tests := []struct {
    		ratings []int
    		want    int
    	}{
    		{nil, 0},
    		{[]int{900, 1400}, 500},
    		{[]int{1500, 1300, 1200, 1400}, 0},
    		{[]int{1500, 1200, 1800, 1100}, 200},
    		{[]int{1000, 1000, 1000, 1000, 1000, 1001}, 1},
    		{[]int{1, 2, 3, 4, 5, 100}, 91},                  // 100 + 1 + 2 vs 3 + 4 + 5
    		{[]int{2000, 1990, 1500, 1200, 1100, 1000}, 190}, // equal sizes force a gap
    		{[]int{50, 50, 50, 50}, 0},
    		{[]int{-10, 10, 20, -20}, 0},
    		{[]int{3, 1, 4, 1, 5, 9, 2, 6}, 1},
    	}
    	for _, tt := range tests {
    		if got := bestSplit(tt.ratings); got != tt.want {
    			t.Errorf("bestSplit(%v) = %d, want %d", tt.ratings, got, tt.want)
    		}
    	}
    }

    func TestBestSplitTwentyPlayers(t *testing.T) {
    	ratings := []int{
    		53562, 88075, 7304, 97953, 68710, 12807, 80756, 11645, 38354, 27075,
    		72000, 89753, 76734, 30783, 91756, 24789, 18314, 66107, 82456, 66929,
    	}
    	start := time.Now()
    	got := bestSplit(ratings)
    	if d := time.Since(start); d > 2*time.Second {
    		t.Errorf("bestSplit(20 players) took %v: 2^20 subsets should take well under a second", d)
    	}
    	if got != 2 {
    		t.Errorf("bestSplit(%v) = %d, want 2", ratings, got)
    	}
    }
---

The ranked 5v5 mode (and the 10v10 raid event) needs **fair teams**: split the
lobby into two teams of **equal size** so that the teams' total ratings are as
close as possible.

This is a version of the *partition* problem, which is NP-complete. Checking a
proposed split is easy; finding the best one seems to need a search. Write both
halves:

- **`teamDiff(ratings, teamA)`** is the verifier. `teamA` lists the indexes of
  the players on team A; everyone else is on team B. If `teamA` is a valid team
  (exactly `len(ratings)/2` indexes, all in range, none repeated), return the
  absolute difference between the two teams' total ratings and `true`.
  Otherwise return `0, false`. Don't modify `teamA`.
- **`bestSplit(ratings)`** returns the **smallest** difference any valid split
  can achieve. With no players, return 0.

## Examples

```go
ratings := []int{1500, 1300, 1200, 1400}
teamDiff(ratings, []int{0, 1}) // 200, true   (2800 vs 2600)
teamDiff(ratings, []int{2, 2}) // 0, false    (a player can't be picked twice)
bestSplit(ratings)             // 0           (1500 + 1200 vs 1300 + 1400)
```

## Constraints

- `len(ratings)` is even and at most 20. Ratings may be negative in tests.
- `teamDiff` should be **O(n)**.
- `bestSplit` may be exponential: there are only 2²⁰ ≈ 1 million subsets of 20
  players, which is fast. (At 60 players it would be hopeless, which is the whole
  point of NP-hardness.)
