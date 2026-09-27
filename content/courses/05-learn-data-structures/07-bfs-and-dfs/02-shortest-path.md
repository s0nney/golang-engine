---
title: Shortest Paths with BFS
quiz:
  - question: 'Does BFS find the fastest route in the **weighted** world map (walking minutes)?'
    options:
      - text: Yes, BFS always finds the lowest total weight
      - text: Only if you reverse the edges
      - text: No; it minimises the number of paths taken, ignoring minutes
        correct: true
      - text: No; BFS doesn't work on weighted graphs at all
    explanation: |
      BFS counts edges. Village → Forest → Caves → Castle is 3 paths but 26 minutes,
      while Village → Lake → Mountain → Castle is also 3 paths but only 23 minutes.
      For weights you need Dijkstra's algorithm, at the end of this chapter.
  - question: |
      The `parent` map ends up as `{Forest: Village, Lake: Village, Caves: Forest,
      Mountain: Lake, Castle: Caves}`. What path does following it back from the
      Castle give?
    options:
      - text: '`[Castle Caves Forest Village]`, which you reverse to `[Village Forest Caves Castle]`'
        correct: true
      - text: '`[Village Lake Mountain Castle]`'
      - text: '`[Castle Mountain Lake Village]`'
      - text: '`[Village Castle]`'
    explanation: |
      Castle's parent is Caves, Caves' parent is Forest, Forest's parent is Village,
      and Village is the start. Walking back produces the path in reverse, so reverse
      it with `slices.Reverse`.
exercise:
  starter: |
    package main

    import "fmt"

    // Distances returns the fewest paths needed to reach each zone from start.
    // Unreachable zones are left out of the map.
    func Distances(adj map[string][]string, start string) map[string]int {
    	dist := map[string]int{}
    	// ?
    	return dist
    }

    func main() {
    	world := map[string][]string{
    		"Village":  {"Forest", "Lake"},
    		"Forest":   {"Village", "Lake", "Caves"},
    		"Lake":     {"Village", "Forest", "Mountain"},
    		"Caves":    {"Forest", "Castle"},
    		"Mountain": {"Lake", "Castle"},
    		"Castle":   {"Caves", "Mountain"},
    		"Island":   {},
    	}
    	d := Distances(world, "Village")
    	for _, zone := range []string{"Village", "Forest", "Caves", "Castle", "Island"} {
    		hops, ok := d[zone]
    		fmt.Printf("%-8s %d %v\n", zone, hops, ok)
    	}
    	// want: Village 0 true, Forest 1 true, Caves 2 true, Castle 3 true, Island 0 false
    }
  solution: |
    package main

    import "fmt"

    func Distances(adj map[string][]string, start string) map[string]int {
    	dist := map[string]int{start: 0}
    	queue := []string{start}
    	for len(queue) > 0 {
    		v := queue[0]
    		queue = queue[1:]
    		for _, n := range adj[v] {
    			if _, seen := dist[n]; !seen {
    				dist[n] = dist[v] + 1
    				queue = append(queue, n)
    			}
    		}
    	}
    	return dist
    }

    func main() {
    	world := map[string][]string{
    		"Village":  {"Forest", "Lake"},
    		"Forest":   {"Village", "Lake", "Caves"},
    		"Lake":     {"Village", "Forest", "Mountain"},
    		"Caves":    {"Forest", "Castle"},
    		"Mountain": {"Lake", "Castle"},
    		"Castle":   {"Caves", "Mountain"},
    		"Island":   {},
    	}
    	d := Distances(world, "Village")
    	for _, zone := range []string{"Village", "Forest", "Caves", "Castle", "Island"} {
    		hops, ok := d[zone]
    		fmt.Printf("%-8s %d %v\n", zone, hops, ok)
    	}
    	// want: Village 0 true, Forest 1 true, Caves 2 true, Castle 3 true, Island 0 false
    }
  tests: |
    package main

    import (
    	"maps"
    	"testing"
    )

    var world = map[string][]string{
    	"Village":    {"Forest", "Lake"},
    	"Forest":     {"Village", "Lake", "Caves"},
    	"Lake":       {"Village", "Forest", "Mountain"},
    	"Caves":      {"Forest", "Castle"},
    	"Mountain":   {"Lake", "Castle"},
    	"Castle":     {"Caves", "Mountain"},
    	"Island":     {"Lighthouse"},
    	"Lighthouse": {"Island"},
    }

    func TestDistancesFromVillage(t *testing.T) {
    	got := Distances(world, "Village")
    	want := map[string]int{"Village": 0, "Forest": 1, "Lake": 1, "Caves": 2, "Mountain": 2, "Castle": 3}
    	if !maps.Equal(got, want) {
    		t.Errorf("Distances(world, \"Village\") = %v, want %v", got, want)
    	}
    }

    func TestDistancesFromCastle(t *testing.T) {
    	got := Distances(world, "Castle")
    	want := map[string]int{"Castle": 0, "Caves": 1, "Mountain": 1, "Forest": 2, "Lake": 2, "Village": 3}
    	if !maps.Equal(got, want) {
    		t.Errorf("Distances(world, \"Castle\") = %v, want %v", got, want)
    	}
    }

    func TestUnreachableZonesLeftOut(t *testing.T) {
    	got := Distances(world, "Island")
    	want := map[string]int{"Island": 0, "Lighthouse": 1}
    	if !maps.Equal(got, want) {
    		t.Errorf("Distances(world, \"Island\") = %v, want %v", got, want)
    	}
    }

    func TestLongChain(t *testing.T) {
    	// A ring of zones: going the short way round must win.
    	adj := map[string][]string{
    		"z0": {"z1", "z5"}, "z1": {"z0", "z2"}, "z2": {"z1", "z3"},
    		"z3": {"z2", "z4"}, "z4": {"z3", "z5"}, "z5": {"z4", "z0"},
    	}
    	got := Distances(adj, "z0")
    	want := map[string]int{"z0": 0, "z1": 1, "z5": 1, "z2": 2, "z4": 2, "z3": 3}
    	if !maps.Equal(got, want) {
    		t.Errorf("Distances around a ring = %v, want %v", got, want)
    	}
    }
---

The hint system knows how *far* each zone is. Now players want directions: "what's
the quickest way from the Village to the Castle, counting paths taken?" BFS can
answer that with one extra map.

## Remember where you came from

When BFS discovers a vertex `n` while processing `v`, record `parent[n] = v`. Because
BFS reaches each vertex first along a shortest route, those parent links form a
**shortest-path tree** rooted at the start. To get the route to any target, follow
parents backwards from the target to the start, then reverse.

```go
package main

import (
	"fmt"
	"slices"
)

// ShortestPath returns a path from start to goal with the fewest edges,
// or nil if goal can't be reached.
func ShortestPath[T comparable](adj map[T][]T, start, goal T) []T {
	parent := map[T]T{}
	seen := map[T]bool{start: true}
	queue := []T{start}
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]
		if v == goal {
			path := []T{goal}
			for v != start {
				v = parent[v]
				path = append(path, v)
			}
			slices.Reverse(path)
			return path
		}
		for _, n := range adj[v] {
			if !seen[n] {
				seen[n] = true
				parent[n] = v
				queue = append(queue, n)
			}
		}
	}
	return nil
}

func main() {
	world := map[string][]string{
		"Village":  {"Forest", "Lake"},
		"Forest":   {"Village", "Lake", "Caves"},
		"Lake":     {"Village", "Forest", "Mountain"},
		"Caves":    {"Forest", "Castle"},
		"Mountain": {"Lake", "Castle"},
		"Castle":   {"Caves", "Mountain"},
		"Island":   {},
	}
	fmt.Println(ShortestPath(world, "Village", "Castle"))
	fmt.Println(ShortestPath(world, "Mountain", "Forest"))
	fmt.Println(ShortestPath(world, "Village", "Village"))
	fmt.Println(ShortestPath(world, "Village", "Island") == nil)
}
```

Output:

```
[Village Forest Caves Castle]
[Mountain Lake Forest]
[Village]
true
```

## Details that matter

- **Stop early.** Once the goal comes out of the queue, we're done. There's no need to
  explore the rest of the world.
- **Start equals goal** returns `[start]`, a path with zero edges. The loop
  `for v != start` doesn't run at all.
- **Unreachable goals** return `nil`. Callers can check `path == nil`, or `len(path) == 0`,
  which also covers an empty slice.
- **Ties.** Two routes to the Castle take 3 paths each. BFS returns whichever it found
  first, which depends on neighbour order. Both are correct answers.
- **Reversing.** Walking parents gives the path backwards. `slices.Reverse` flips it in
  place.

The path has `len(path) - 1` edges. For a distance-only answer, you can keep a
`dist map[T]int` instead, with `dist[n] = dist[v] + 1` on discovery.

## Unweighted only

BFS minimises the **number of edges**. It knows nothing about walking times. In the
weighted map, it returns Village → Forest → Caves → Castle (26 minutes) even though
Village → Lake → Mountain → Castle takes 23. For weighted graphs with non-negative
weights, you need Dijkstra's algorithm, coming at the end of this chapter. (A useful
trick: if every edge has the *same* weight, BFS is still optimal.)

## Your turn

The quest-hint system needs the distance to *every* zone at once, not a single route.
Complete `Distances` in the exercise. It returns a map from each zone reachable from
`start` to the fewest paths needed to reach it (`start` itself is 0). Zones that can't
be reached must not appear in the map at all.
