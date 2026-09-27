---
title: Adjacency List
quiz:
  - question: How much memory does an adjacency list use for a graph with V vertices and E edges?
    options:
      - text: O(V²)
      - text: O(V + E)
        correct: true
      - text: O(E²)
      - text: O(V · E)
    explanation: |
      There's one entry per vertex, plus one slice element per edge (two for an
      undirected edge). Nothing is stored for pairs that *aren't* connected, which is
      what makes lists so good for sparse graphs.
  - question: |
      What does this print?

      ```go
      adj := map[string][]string{}
      adj["Village"] = append(adj["Village"], "Forest")
      adj["Village"] = append(adj["Village"], "Lake")
      fmt.Println(len(adj), len(adj["Village"]), len(adj["Forest"]))
      ```
    options:
      - text: '`1 2 0`'
        correct: true
      - text: '`2 2 1`'
      - text: '`3 2 0`'
      - text: It panics, because `adj["Village"]` starts as nil
    explanation: |
      Appending to a nil slice is fine, so the first `append` creates the slice. Only
      Village was ever assigned, so the map has 1 key. Forest was never added in the
      other direction, so its slice is nil with length 0. That's the classic bug the
      exercise asks you to avoid.
exercise:
  starter: |
    package main

    import "fmt"

    // World is the game map as an adjacency list: zone -> neighbouring zones.
    type World struct {
    	adj map[string][]string
    }

    func NewWorld() *World {
    	return &World{adj: map[string][]string{}}
    }

    // AddPath adds a two-way path between zones a and b.
    func (w *World) AddPath(a, b string) {
    	// ?
    }

    // Neighbors returns the zones directly connected to zone, in the order
    // their paths were added. Unknown zones have no neighbours.
    func (w *World) Neighbors(zone string) []string {
    	// ?
    	return nil
    }

    func main() {
    	w := NewWorld()
    	w.AddPath("Village", "Forest")
    	w.AddPath("Village", "Lake")
    	w.AddPath("Lake", "Forest")
    	fmt.Println("Village:", w.Neighbors("Village")) // want [Forest Lake]
    	fmt.Println("Forest:", w.Neighbors("Forest"))   // want [Village Lake]
    	fmt.Println("Castle:", w.Neighbors("Castle"))   // want []
    }
  solution: |
    package main

    import "fmt"

    // World is the game map as an adjacency list: zone -> neighbouring zones.
    type World struct {
    	adj map[string][]string
    }

    func NewWorld() *World {
    	return &World{adj: map[string][]string{}}
    }

    func (w *World) AddPath(a, b string) {
    	w.adj[a] = append(w.adj[a], b)
    	w.adj[b] = append(w.adj[b], a)
    }

    func (w *World) Neighbors(zone string) []string {
    	return w.adj[zone]
    }

    func main() {
    	w := NewWorld()
    	w.AddPath("Village", "Forest")
    	w.AddPath("Village", "Lake")
    	w.AddPath("Lake", "Forest")
    	fmt.Println("Village:", w.Neighbors("Village")) // want [Forest Lake]
    	fmt.Println("Forest:", w.Neighbors("Forest"))   // want [Village Lake]
    	fmt.Println("Castle:", w.Neighbors("Castle"))   // want []
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func buildWorld() *World {
    	w := NewWorld()
    	w.AddPath("Village", "Forest")
    	w.AddPath("Village", "Lake")
    	w.AddPath("Lake", "Forest")
    	w.AddPath("Forest", "Caves")
    	w.AddPath("Caves", "Castle")
    	w.AddPath("Lake", "Mountain")
    	w.AddPath("Mountain", "Castle")
    	return w
    }

    func TestNeighbors(t *testing.T) {
    	w := buildWorld()
    	for zone, want := range map[string][]string{
    		"Village":  {"Forest", "Lake"},
    		"Forest":   {"Village", "Lake", "Caves"},
    		"Lake":     {"Village", "Forest", "Mountain"},
    		"Caves":    {"Forest", "Castle"},
    		"Mountain": {"Lake", "Castle"},
    		"Castle":   {"Caves", "Mountain"},
    	} {
    		if got := w.Neighbors(zone); !slices.Equal(got, want) {
    			t.Errorf("Neighbors(%q) = %v, want %v", zone, got, want)
    		}
    	}
    }

    func TestPathsGoBothWays(t *testing.T) {
    	w := NewWorld()
    	w.AddPath("Village", "Forest")
    	if !slices.Contains(w.Neighbors("Forest"), "Village") {
    		t.Errorf(`after AddPath("Village", "Forest"), the Forest can't get back to the Village`)
    	}
    }

    func TestUnknownZone(t *testing.T) {
    	w := buildWorld()
    	if got := w.Neighbors("Atlantis"); len(got) != 0 {
    		t.Errorf(`Neighbors("Atlantis") = %v, want no neighbours`, got)
    	}
    }

    func TestDegreeSum(t *testing.T) {
    	w := buildWorld()
    	total := 0
    	for _, zone := range []string{"Village", "Forest", "Lake", "Caves", "Mountain", "Castle"} {
    		total += len(w.Neighbors(zone))
    	}
    	if total != 14 {
    		t.Errorf("7 two-way paths should give 14 neighbour entries in total, got %d", total)
    	}
    }
---

The adjacency **list** is the representation you'll use most. Each vertex keeps a
list of just the vertices it connects to. No row of zeros, no wasted space.

```
Village  → Forest, Lake
Forest   → Village, Lake, Caves
Lake     → Village, Forest, Mountain
Caves    → Forest, Castle
Mountain → Lake, Castle
Castle   → Caves, Mountain
```

## In Go

A map from vertex to a slice of neighbours is all you need:

```go
package main

import "fmt"

func main() {
	adj := map[string][]string{}
	addPath := func(a, b string) {
		adj[a] = append(adj[a], b)
		adj[b] = append(adj[b], a) // undirected: store both directions
	}
	addPath("Village", "Forest")
	addPath("Village", "Lake")
	addPath("Lake", "Forest")
	addPath("Forest", "Caves")
	addPath("Caves", "Castle")
	addPath("Lake", "Mountain")
	addPath("Mountain", "Castle")

	fmt.Println(adj["Lake"])
	fmt.Println(len(adj), "zones")
	for _, n := range adj["Castle"] {
		fmt.Println("Castle ->", n)
	}
}
```

Output:

```
[Village Forest Mountain]
6 zones
Castle -> Caves
Castle -> Mountain
```

A few Go details make this pleasantly short:

- `append` to a nil slice creates it, and a missing map key reads as a nil slice. So
  `adj[a] = append(adj[a], b)` works whether or not `a` has been seen before.
- You must **assign** the result back to `adj[a]`. A bare `append(adj[a], b)`
  statement doesn't even compile ("is not used"), and if you stored the result in
  some other variable, the map would still hold the old slice header.
- For an undirected graph, forgetting the second `append` is the #1 bug: you can walk
  from the Village to the Forest but not back.
- Ranging over a *slice* of neighbours gives them in insertion order, which keeps
  graph algorithms deterministic. Ranging over the *map* `adj` would give zones in a
  random order.

## Trade-offs vs the matrix

| Operation | Adjacency list | Adjacency matrix |
|---|---|---|
| Is there an edge a → b? | O(degree of a) | **O(1)** |
| List neighbours of a | **O(degree of a)** | O(V) |
| Visit every edge | **O(V + E)** | O(V²) |
| Memory | **O(V + E)** | O(V²) |

Most graph algorithms, including BFS and DFS in the next chapter, repeatedly ask "who
are my neighbours?" and eventually visit every edge. With a list, that totals
O(V + E). With a matrix it's O(V²), even if the graph has hardly any edges.

If you also need fast "is a connected to b?" checks, you can store the neighbours in
a set instead of a slice: `map[string]map[string]bool`. You lose the stable order,
but edge checks become O(1) on average.

## Your turn

The exercise gives you a `World` type backed by an adjacency list. Complete
`AddPath` so it adds an undirected path between two zones (both directions!), and
`Neighbors` so it returns a zone's neighbours in the order their paths were added.
An unknown zone has no neighbours, so return an empty result rather than panicking.
