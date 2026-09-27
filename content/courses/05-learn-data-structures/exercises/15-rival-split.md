---
title: Rival Split
difficulty: hard
after: p-vs-np
hints:
  - 'Two teams means two colours, and "no rivals on the same team" means no edge joins two same-coloured players: a **2-colouring**. Colour greedily with BFS: the first player of each component gets team 0, and every newly discovered rival gets the opposite team of the player who discovered them. Run it from every not-yet-coloured player, because the rivalry graph can have many components.'
  - 'If BFS ever finds a rival already on the **same** team, no split exists. For the proof, record each player''s BFS `parent` and `depth`. In BFS, rivals'' depths differ by at most 1, so same-team rivals `u` and `v` are at the same depth. Walk `u` and `v` up the parents **in lockstep** until they meet at a common ancestor `a`. Then `u → … → a` followed by the path from `a` back down to `v` is a cycle of `2·(depth[u] − depth[a]) + 1` players: odd.'
  - 'Mind the edge cases: a player who is their own rival (`{a, a}`) is an odd cycle of length 1 (`[a]`). Build the cycle with `append` and one `slices.Reverse`, not by inserting at the front of a slice, because the only odd cycle may contain 100,000 players.'
exercise:
  starter: |
    package main

    import "fmt"

    // SplitTeams puts every player 0..n-1 on team 0 or 1 so that no two
    // rivals are on the same team, and returns team, nil. If that's
    // impossible, it returns nil and an odd cycle of rivals as proof.
    func SplitTeams(n int, rivalries [][2]int) (team []int, oddCycle []int) {
    	return nil, nil
    }

    func main() {
    	fmt.Println(SplitTeams(4, [][2]int{{0, 1}, {1, 2}, {2, 3}, {3, 0}})) // e.g. [0 1 0 1] []
    	fmt.Println(SplitTeams(5, [][2]int{{0, 1}, {1, 2}, {2, 0}, {3, 4}})) // e.g. [] [0 1 2]
    }
  solution: |
    package main

    import (
    	"fmt"
    	"slices"
    )

    // SplitTeams puts every player on team 0 or 1 so that no two rivals are on
    // the same team. If that's impossible, it returns an odd cycle of rivals
    // instead, as proof.
    func SplitTeams(n int, rivalries [][2]int) (team []int, oddCycle []int) {
    	adj := make([][]int, n)
    	for _, r := range rivalries {
    		a, b := r[0], r[1]
    		if a == b {
    			return nil, []int{a} // rivals with themselves: a cycle of length 1
    		}
    		adj[a] = append(adj[a], b)
    		adj[b] = append(adj[b], a)
    	}
    	depth := make([]int, n)
    	parent := make([]int, n)
    	for i := range depth {
    		depth[i] = -1
    	}
    	for s := range n {
    		if depth[s] != -1 {
    			continue // already placed via an earlier component
    		}
    		depth[s], parent[s] = 0, -1
    		queue := []int{s}
    		for len(queue) > 0 {
    			u := queue[0]
    			queue = queue[1:]
    			for _, v := range adj[u] {
    				if depth[v] == -1 {
    					depth[v], parent[v] = depth[u]+1, u
    					queue = append(queue, v)
    				} else if depth[v]%2 == depth[u]%2 {
    					// Same team on both ends. In BFS, rivals' depths differ by at
    					// most 1, so here depth[u] == depth[v]: climb together to the
    					// lowest common ancestor.
    					left, right := []int{u}, []int{v}
    					for u != v {
    						u, v = parent[u], parent[v]
    						left, right = append(left, u), append(right, v)
    					}
    					right = right[:len(right)-1] // the ancestor is already in left
    					slices.Reverse(right)
    					return nil, append(left, right...)
    				}
    			}
    		}
    	}
    	team = make([]int, n)
    	for i, d := range depth {
    		team[i] = d % 2
    	}
    	return team, nil
    }

    func main() {
    	fmt.Println(SplitTeams(4, [][2]int{{0, 1}, {1, 2}, {2, 3}, {3, 0}}))
    	fmt.Println(SplitTeams(5, [][2]int{{0, 1}, {1, 2}, {2, 0}, {3, 4}}))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"testing"
    	"time"
    )

    // checkSplit returns "" if team is a valid split, or what's wrong with it.
    func checkSplit(n int, rivalries [][2]int, team []int) string {
    	if len(team) != n {
    		return fmt.Sprintf("team has %d entries, want %d", len(team), n)
    	}
    	for i, t := range team {
    		if t != 0 && t != 1 {
    			return fmt.Sprintf("team[%d] = %d, want 0 or 1", i, t)
    		}
    	}
    	for _, r := range rivalries {
    		if team[r[0]] == team[r[1]] {
    			return fmt.Sprintf("rivals %d and %d are both on team %d", r[0], r[1], team[r[0]])
    		}
    	}
    	return ""
    }

    // checkCycle returns "" if cycle is an odd cycle of distinct rivals.
    func checkCycle(n int, rivalries [][2]int, cycle []int) string {
    	if len(cycle)%2 == 0 {
    		return fmt.Sprintf("the cycle has %d players, want an odd number", len(cycle))
    	}
    	rivals := map[[2]int]bool{}
    	for _, r := range rivalries {
    		rivals[r] = true
    		rivals[[2]int{r[1], r[0]}] = true
    	}
    	seen := map[int]bool{}
    	for i, p := range cycle {
    		if p < 0 || p >= n || seen[p] {
    			return fmt.Sprintf("player %d is out of range or appears twice", p)
    		}
    		seen[p] = true
    		if q := cycle[(i+1)%len(cycle)]; !rivals[[2]int{p, q}] {
    			return fmt.Sprintf("%d and %d are next to each other in the cycle but aren't rivals", p, q)
    		}
    	}
    	return ""
    }

    func expectSplit(t *testing.T, name string, n int, rivalries [][2]int) {
    	t.Helper()
    	team, cycle := SplitTeams(n, rivalries)
    	if cycle != nil {
    		t.Errorf("%s: SplitTeams(%d, %v) returned odd cycle %v, but a split exists", name, n, rivalries, cycle)
    	} else if msg := checkSplit(n, rivalries, team); msg != "" {
    		t.Errorf("%s: SplitTeams(%d, %v) = %v: %s", name, n, rivalries, team, msg)
    	}
    }

    func expectCycle(t *testing.T, name string, n int, rivalries [][2]int) {
    	t.Helper()
    	team, cycle := SplitTeams(n, rivalries)
    	if team != nil {
    		t.Errorf("%s: SplitTeams(%d, %v) returned team %v, but no valid split exists (want team == nil)", name, n, rivalries, team)
    	}
    	if len(cycle) == 0 {
    		t.Errorf("%s: SplitTeams(%d, %v) returned no odd cycle, want one as proof", name, n, rivalries)
    	} else if msg := checkCycle(n, rivalries, cycle); msg != "" {
    		t.Errorf("%s: SplitTeams(%d, %v) returned cycle %v: %s", name, n, rivalries, cycle, msg)
    	}
    }

    func TestSplitPossible(t *testing.T) {
    	expectSplit(t, "nobody", 0, nil)
    	expectSplit(t, "one player", 1, nil)
    	expectSplit(t, "no rivalries", 3, nil)
    	expectSplit(t, "one rivalry", 2, [][2]int{{1, 0}})
    	expectSplit(t, "path", 4, [][2]int{{0, 1}, {1, 2}, {2, 3}})
    	expectSplit(t, "square", 4, [][2]int{{0, 1}, {1, 2}, {2, 3}, {3, 0}})
    	expectSplit(t, "same rivalry twice", 2, [][2]int{{0, 1}, {1, 0}})
    	expectSplit(t, "two components", 6, [][2]int{{0, 1}, {4, 5}, {5, 3}})
    	expectSplit(t, "hexagon with a long chord", 6, [][2]int{{0, 1}, {1, 2}, {2, 3}, {3, 4}, {4, 5}, {5, 0}, {0, 3}})
    }

    func TestSplitImpossible(t *testing.T) {
    	expectCycle(t, "triangle", 3, [][2]int{{0, 1}, {1, 2}, {2, 0}})
    	expectCycle(t, "triangle plus a separate pair", 5, [][2]int{{0, 1}, {1, 2}, {2, 0}, {3, 4}})
    	expectCycle(t, "pentagon", 5, [][2]int{{0, 1}, {1, 2}, {2, 3}, {3, 4}, {4, 0}})
    	expectCycle(t, "own rival", 3, [][2]int{{0, 1}, {2, 2}})
    	expectCycle(t, "odd cycle in the second component", 7, [][2]int{{0, 1}, {1, 2}, {3, 4}, {4, 5}, {5, 6}, {6, 3}, {4, 6}})
    	expectCycle(t, "hexagon with a short chord", 6, [][2]int{{0, 1}, {1, 2}, {2, 3}, {3, 4}, {4, 5}, {5, 0}, {0, 2}})
    	expectCycle(t, "triangle far from player 0", 7, [][2]int{{0, 1}, {1, 2}, {2, 3}, {3, 4}, {4, 5}, {5, 6}, {6, 4}})
    }

    // TestAgainstBruteForce decides every small case by trying all 2^n splits
    // (the exponential way), then checks SplitTeams' answer and its proof.
    func TestAgainstBruteForce(t *testing.T) {
    	var x uint64 = 3
    	next := func(k int) int {
    		x = x*6364136223846793005 + 1442695040888963407
    		return int(x>>33) % k
    	}
    	for range 500 {
    		n := 1 + next(9)
    		var rivalries [][2]int
    		for range next(12) {
    			rivalries = append(rivalries, [2]int{next(n), next(n)})
    		}
    		possible := false
    		for mask := range 1 << n {
    			ok := true
    			for _, r := range rivalries {
    				if mask>>r[0]&1 == mask>>r[1]&1 {
    					ok = false
    					break
    				}
    			}
    			if ok {
    				possible = true
    				break
    			}
    		}
    		if possible {
    			expectSplit(t, "random", n, rivalries)
    		} else {
    			expectCycle(t, "random", n, rivalries)
    		}
    		if t.Failed() {
    			return
    		}
    	}
    }

    func TestBigLeagues(t *testing.T) {
    	const n = 200_000
    	var x uint64 = 11
    	next := func(k int) int {
    		x = x*6364136223846793005 + 1442695040888963407
    		return int(x>>33) % k
    	}
    	shuffle := func(s [][2]int) {
    		for i := len(s) - 1; i > 0; i-- {
    			j := next(i + 1)
    			s[i], s[j] = s[j], s[i]
    		}
    	}
    	// League 1: 300,000 rivalries that respect a hidden split.
    	side := make([]int, n)
    	for i := range side {
    		side[i] = next(2)
    	}
    	var league1 [][2]int
    	for len(league1) < 300_000 {
    		if a, b := next(n), next(n); side[a] != side[b] {
    			league1 = append(league1, [2]int{a, b})
    		}
    	}
    	// League 2: a ring of 100,001 players (the only odd cycle), plus
    	// rivalries among the other players that respect a hidden split.
    	perm := make([]int, n)
    	for i := range perm {
    		perm[i] = i
    	}
    	for i := n - 1; i > 0; i-- {
    		j := next(i + 1)
    		perm[i], perm[j] = perm[j], perm[i]
    	}
    	const ring = 100_001
    	var league2 [][2]int
    	for i := range ring {
    		league2 = append(league2, [2]int{perm[i], perm[(i+1)%ring]})
    	}
    	rest := perm[ring:]
    	for len(league2) < 300_000 {
    		if a, b := next(len(rest)), next(len(rest)); a%2 != b%2 {
    			league2 = append(league2, [2]int{rest[a], rest[b]})
    		}
    	}
    	shuffle(league1)
    	shuffle(league2)

    	type answer struct{ team, cycle []int }
    	done := make(chan [2]answer, 1)
    	go func() {
    		var res [2]answer
    		res[0].team, res[0].cycle = SplitTeams(n, league1)
    		res[1].team, res[1].cycle = SplitTeams(n, league2)
    		done <- res
    	}()
    	select {
    	case res := <-done:
    		if res[0].cycle != nil {
    			t.Errorf("league 1 (%d players, %d rivalries): returned an odd cycle of %d players, but a split exists", n, len(league1), len(res[0].cycle))
    		} else if msg := checkSplit(n, league1, res[0].team); msg != "" {
    			t.Errorf("league 1 (%d players, %d rivalries): %s", n, len(league1), msg)
    		}
    		if res[1].team != nil || len(res[1].cycle) == 0 {
    			t.Errorf("league 2 (%d players, %d rivalries): want team == nil and an odd cycle, got a team of %d and a cycle of %d", n, len(league2), len(res[1].team), len(res[1].cycle))
    		} else if msg := checkCycle(n, league2, res[1].cycle); msg != "" {
    			t.Errorf("league 2 (%d players, %d rivalries): %s", n, len(league2), msg)
    		}
    	case <-time.After(time.Second):
    		t.Fatalf("SplitTeams on two leagues of %d players and 300,000 rivalries took over a second: use BFS (O(V + E)), and build the cycle in O(its length)", n)
    	}
    }
---

The weekend tournament splits every signed-up player into two teams, **Red (0)**
and **Blue (1)**, and players have reported their rivals. Rivals must never be
on the same team: they'd sabotage each other.

Splitting into **three** teams under the same rule is *graph 3-colouring*, which is
NP-complete: nobody knows a fast algorithm. Two teams is different. It's in P, and
you can find the split (or prove there isn't one) in O(V + E).

Write `SplitTeams(n, rivalries)`. Players are numbered `0` to `n-1`, and each
`{a, b}` in `rivalries` says that `a` and `b` are rivals.

- If a split exists, return `team, nil`, where `team[i]` is `0` or `1` and every
  pair of rivals is on different teams. Any valid split is accepted.
- Otherwise return `nil, oddCycle`: a **proof** that no split exists. `oddCycle`
  lists distinct players `p0, p1, …, pk-1` where **k is odd**, each player is a
  rival of the next, and the last is a rival of the first. (Around an odd cycle the
  teams would have to alternate Red, Blue, Red, … and end on the same colour they
  started with, so no split can work.) Any odd cycle is accepted.

Like a certificate in NP, the proof is quick to check even when it was hard to
find. Here it's easy to find too.

## Examples

```go
SplitTeams(4, [][2]int{{0, 1}, {1, 2}, {2, 3}, {3, 0}})
// [0 1 0 1] []      (or [1 0 1 0] [])

SplitTeams(5, [][2]int{{0, 1}, {1, 2}, {2, 0}, {3, 4}})
// [] [0 1 2]        (any order around the triangle, e.g. [2 1 0], is fine)
```

## Constraints

- Up to 200,000 players and 300,000 rivalries. The rivalry graph can have many
  components, repeated rivalries, and a player can be their own rival (`{a, a}`,
  an odd cycle of length 1).
- O(V + E) is expected. The performance test solves one league that can be split
  and one whose only odd cycle has 100,001 players, within one second in total.
  Trying all 2ⁿ splits is hopeless, and so is re-scanning the whole rivalry list
  for every player.
