---
title: The Travelling Salesman Problem
quiz:
  - question: 'Which version of the travelling-salesman problem is **NP-complete**?'
    options:
      - text: '"Find the shortest tour"'
      - text: '"Is there a tour of at most k minutes?"'
        correct: true
      - text: '"Is the tour I''m holding the shortest one?"'
      - text: Both the first and second
    explanation: |
      The yes/no version is in NP (a tour is a certificate you can add up in O(n)) and
      NP-hard, so it's NP-complete. The optimisation version is NP-hard, but it isn't
      a decision problem, so it isn't in NP.
  - question: The greedy nearest-neighbour tour was 57 minutes and the optimum was 49. What does that show?
    options:
      - text: Greedy algorithms are always wrong
      - text: Greedy is fast but can miss the optimum; always walking to the closest zone can leave an expensive trip for later
        correct: true
      - text: The brute-force code has a bug
      - text: TSP is actually in P
    explanation: |
      Nearest neighbour is O(n²) and often decent, but it has no guarantee. Here the
      cheap early steps (Lake, Forest) forced a long walk home from the Caves at the end.
exercise:
  starter: |
    package main

    import "fmt"

    // minutes[i][j] is the walking time between zones i and j.
    var minutes = [][]int{
    	{0, 5, 3, 20, 11, 23},
    	{5, 0, 4, 15, 12, 21},
    	{3, 4, 0, 19, 8, 20},
    	{20, 15, 19, 0, 18, 6},
    	{11, 12, 8, 18, 0, 12},
    	{23, 21, 20, 6, 12, 0},
    }

    // tourLength returns the total time of walking tour in order and then back
    // to tour[0]. An empty tour takes 0 minutes.
    func tourLength(minutes [][]int, tour []int) int {
    	// ?
    	return 0
    }

    // checkTour verifies a certificate for the decision version of TSP: "is
    // there a World Tour of at most limit minutes?". tour must start at zone 0
    // and visit every zone exactly once (the walk back to 0 is implied), and
    // its length must be at most limit. It must not modify tour.
    func checkTour(minutes [][]int, tour []int, limit int) bool {
    	// ?
    	return false
    }

    func main() {
    	fmt.Println(tourLength(minutes, []int{0, 1, 3, 5, 4, 2}))    // want: 49
    	fmt.Println(checkTour(minutes, []int{0, 1, 3, 5, 4, 2}, 50)) // want: true
    	fmt.Println(checkTour(minutes, []int{0, 1, 3, 5, 4, 2}, 48)) // want: false (too slow)
    	fmt.Println(checkTour(minutes, []int{0, 1, 3, 5, 4, 1}, 99)) // want: false (Forest twice, Lake never)
    }
  solution: |
    package main

    import "fmt"

    // minutes[i][j] is the walking time between zones i and j.
    var minutes = [][]int{
    	{0, 5, 3, 20, 11, 23},
    	{5, 0, 4, 15, 12, 21},
    	{3, 4, 0, 19, 8, 20},
    	{20, 15, 19, 0, 18, 6},
    	{11, 12, 8, 18, 0, 12},
    	{23, 21, 20, 6, 12, 0},
    }

    func tourLength(minutes [][]int, tour []int) int {
    	total := 0
    	for i, zone := range tour {
    		next := tour[(i+1)%len(tour)]
    		total += minutes[zone][next]
    	}
    	return total
    }

    func checkTour(minutes [][]int, tour []int, limit int) bool {
    	n := len(minutes)
    	if len(tour) != n || n == 0 || tour[0] != 0 {
    		return false
    	}
    	seen := make([]bool, n)
    	for _, zone := range tour {
    		if zone < 0 || zone >= n || seen[zone] {
    			return false
    		}
    		seen[zone] = true
    	}
    	return tourLength(minutes, tour) <= limit
    }

    func main() {
    	fmt.Println(tourLength(minutes, []int{0, 1, 3, 5, 4, 2}))    // want: 49
    	fmt.Println(checkTour(minutes, []int{0, 1, 3, 5, 4, 2}, 50)) // want: true
    	fmt.Println(checkTour(minutes, []int{0, 1, 3, 5, 4, 2}, 48)) // want: false (too slow)
    	fmt.Println(checkTour(minutes, []int{0, 1, 3, 5, 4, 1}, 99)) // want: false (Forest twice, Lake never)
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func TestTourLength(t *testing.T) {
    	for _, tt := range []struct {
    		tour []int
    		want int
    	}{
    		{[]int{0, 1, 3, 5, 4, 2}, 49},
    		{[]int{0, 2, 1, 4, 5, 3}, 3 + 4 + 12 + 12 + 6 + 20},
    		{[]int{0, 1}, 10},
    		{[]int{3}, 0},
    		{nil, 0},
    	} {
    		if got := tourLength(minutes, tt.tour); got != tt.want {
    			t.Errorf("tourLength(%v) = %d, want %d (don't forget the walk back to the start)", tt.tour, got, tt.want)
    		}
    	}
    }

    func TestCheckTour(t *testing.T) {
    	for _, tt := range []struct {
    		why   string
    		tour  []int
    		limit int
    		want  bool
    	}{
    		{"optimal tour, generous limit", []int{0, 1, 3, 5, 4, 2}, 60, true},
    		{"optimal tour, exact limit", []int{0, 1, 3, 5, 4, 2}, 49, true},
    		{"optimal tour, limit too low", []int{0, 1, 3, 5, 4, 2}, 48, false},
    		{"reverse direction is just as long", []int{0, 2, 4, 5, 3, 1}, 49, true},
    		{"doesn't start at the Village", []int{1, 0, 3, 5, 4, 2}, 99, false},
    		{"visits a zone twice", []int{0, 1, 3, 5, 4, 1}, 99, false},
    		{"misses a zone", []int{0, 1, 3, 5, 4}, 99, false},
    		{"too many stops", []int{0, 1, 3, 5, 4, 2, 0}, 99, false},
    		{"zone out of range", []int{0, 1, 3, 5, 4, 6}, 99, false},
    		{"negative zone", []int{0, 1, 3, 5, 4, -2}, 99, false},
    		{"empty tour", []int{}, 99, false},
    	} {
    		in := slices.Clone(tt.tour)
    		if got := checkTour(minutes, tt.tour, tt.limit); got != tt.want {
    			t.Errorf("%s: checkTour(minutes, %v, %d) = %v, want %v", tt.why, tt.tour, tt.limit, got, tt.want)
    		}
    		if !slices.Equal(in, tt.tour) {
    			t.Errorf("%s: checkTour modified the tour to %v", tt.why, tt.tour)
    		}
    	}
    }
---

The live-ops team wants a daily **World Tour** event: start in the Village, visit every
zone exactly once, and return to the Village. Players who finish fastest win a prize.
The designers need the fastest possible tour so they can set a fair target time. This is
the famous **travelling-salesman problem** (TSP).

## The input

For each pair of zones, we need the walking time between them. Running Dijkstra from
every zone of our weighted world map gives this table:

```
            Village Forest Lake Caves Mountain Castle
Village        0       5     3    20     11      23
Forest         5       0     4    15     12      21
Lake           3       4     0    19      8      20
Caves         20      15    19     0     18       6
Mountain      11      12     8    18      0      12
Castle        23      21    20     6     12       0
```

## Brute force: try every order

With the Village fixed as the start, the other 5 zones can be visited in 5! = 120
orders. Try them all and keep the best. The code builds tours with the same recursive
backtracking shape as DFS: choose a next zone, recurse, then undo the choice.

```go
package main

import (
	"fmt"
	"math"
)

var zones = []string{"Village", "Forest", "Lake", "Caves", "Mountain", "Castle"}

// minutes[i][j] = walking time between zones i and j (via the fastest route).
var minutes = [][]int{
	{0, 5, 3, 20, 11, 23},
	{5, 0, 4, 15, 12, 21},
	{3, 4, 0, 19, 8, 20},
	{20, 15, 19, 0, 18, 6},
	{11, 12, 8, 18, 0, 12},
	{23, 21, 20, 6, 12, 0},
}

// bestTour tries every order of visiting zones 1..n-1, starting and ending at zone 0.
func bestTour() ([]int, int, int) {
	n := len(zones)
	best, bestLen, tried := []int(nil), math.MaxInt, 0
	tour := []int{0}
	used := make([]bool, n)
	used[0] = true
	var extend func(length int)
	extend = func(length int) {
		if len(tour) == n {
			tried++
			total := length + minutes[tour[n-1]][0] // walk home
			if total < bestLen {
				best, bestLen = append([]int(nil), tour...), total
			}
			return
		}
		for next := range n {
			if !used[next] {
				used[next] = true
				tour = append(tour, next)
				extend(length + minutes[tour[len(tour)-2]][next])
				tour = tour[:len(tour)-1]
				used[next] = false
			}
		}
	}
	extend(0)
	return best, bestLen, tried
}

func nearestNeighbour() ([]int, int) {
	n := len(zones)
	tour, used, total := []int{0}, make([]bool, n), 0
	used[0] = true
	for len(tour) < n {
		cur, next := tour[len(tour)-1], -1
		for j := range n {
			if !used[j] && (next == -1 || minutes[cur][j] < minutes[cur][next]) {
				next = j
			}
		}
		used[next] = true
		total += minutes[cur][next]
		tour = append(tour, next)
	}
	return tour, total + minutes[tour[n-1]][0]
}

func names(tour []int) []string {
	var out []string
	for _, i := range tour {
		out = append(out, zones[i])
	}
	return out
}

func main() {
	best, length, tried := bestTour()
	fmt.Println("brute force:", names(best), length, "min, tried", tried, "tours")
	greedy, glen := nearestNeighbour()
	fmt.Println("greedy:     ", names(greedy), glen, "min")
}
```

Output:

```
brute force: [Village Forest Caves Castle Mountain Lake] 49 min, tried 120 tours
greedy:      [Village Lake Forest Mountain Castle Caves] 57 min
```

The best World Tour takes 49 minutes: Village → Forest → Caves → Castle → Mountain →
Lake → Village.

Some Go details:

- `append([]int(nil), tour...)` copies the tour. Storing `tour` itself would be a bug,
  because the backtracking keeps modifying the same backing array. Same slice aliasing
  trap as in the tries chapter.
- `math.MaxInt` is a handy "infinity" for the first comparison.
- The `used` slice and the `tour = tour[:len(tour)-1]` line undo each choice after
  exploring it. That's what "backtracking" means.

## Why it doesn't scale

| Zones | Tours to check (start fixed) | At a billion tours per second |
|---|---|---|
| 6 | 120 | instant |
| 11 | 3,628,800 | ~4 ms |
| 16 | 1.3 × 10¹² | ~22 minutes |
| 21 | 2.4 × 10¹⁸ | ~77 years |
| 31 | 2.7 × 10³² | far longer than the age of the universe |

(n − 1)! grows even faster than 2ⁿ. A smarter exact algorithm, Held-Karp, uses dynamic
programming to get down to O(n² · 2ⁿ), which handles a couple of dozen zones. But no one
has found a polynomial-time algorithm, and if P ≠ NP, no one ever will.

## Where TSP sits

- **Decision version** ("is there a tour of at most 50 minutes?"): **NP-complete**.
  Given a tour, you can add up its minutes in O(n) and compare, so it's in NP, and it's
  as hard as any NP problem.
- **Optimisation version** ("find the shortest tour"): **NP-hard**. It's at least as hard
  as the decision version, since if you had the shortest tour you could answer "at most
  50?" instantly.

## Your turn: verify a certificate

Before paying out World Tour prizes, the live-ops team needs to check each
player's claimed route. That's exactly the **verifier** that puts the decision
version of TSP in NP.

- **`tourLength(minutes, tour)`** adds up the walk from each zone to the next,
  **plus the walk from the last zone back to the first**. Using
  `tour[(i+1)%len(tour)]` for the next zone handles the wrap-around.
- **`checkTour(minutes, tour, limit)`** returns `true` only if the tour starts
  at zone 0 (the Village), visits every zone **exactly once** (so it has exactly
  `len(minutes)` stops, none out of range, none repeated) and takes at most
  `limit` minutes. Don't modify `tour`.

The check runs in O(n), even though *finding* a tour that passes may take O(n!).

## A fast, imperfect alternative

The `nearestNeighbour` function in the program is **greedy**: from wherever you are,
walk to the closest zone you haven't visited. It runs in O(n²), fast for thousands of
zones. Here it found a 57-minute tour, 16% worse than optimal. Grabbing the cheap early
steps to the Lake and Forest left a long walk home from the Caves at the end.

That's the typical trade-off with NP-hard problems: **exact and slow, or fast and
approximate**. The final lesson looks at how engineers make that choice in practice.
