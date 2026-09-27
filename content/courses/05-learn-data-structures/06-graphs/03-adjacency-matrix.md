---
title: Adjacency Matrix
quiz:
  - question: A graph has 10,000 zones and each zone connects to about 4 others. Roughly how many cells does its adjacency matrix have, and how many are non-zero?
    options:
      - text: 10,000 cells, about 4,000 non-zero
      - text: 100,000,000 cells, about 40,000 non-zero
        correct: true
      - text: 40,000 cells, all non-zero
      - text: 100,000,000 cells, about half non-zero
    explanation: |
      A matrix always has V × V cells: 10,000² = 100 million. With ~4 neighbours per
      zone, only about 40,000 are filled in. That's 99.96% wasted space, which is why
      sparse graphs use adjacency lists.
  - question: For an **undirected** graph, what's special about the adjacency matrix?
    options:
      - text: It's symmetric, `m[i][j] == m[j][i]` for all i and j
        correct: true
      - text: The diagonal is always 1
      - text: It has only one row
      - text: It's always completely filled
    explanation: |
      Every undirected edge is stored in both directions, so the matrix mirrors across
      its diagonal. A directed graph's matrix usually isn't symmetric.
  - question: What's the cost of answering "is there a path directly from zone i to zone j?" with an adjacency matrix?
    options:
      - text: O(1)
        correct: true
      - text: O(log V)
      - text: O(V)
      - text: O(E)
    explanation: |
      It's a single index lookup, `m[i][j]`. That's the matrix's big strength. Listing
      all neighbours of a zone, however, takes a scan of the whole row: O(V).
---

An edge list makes "who are my neighbours?" slow. The first classic fix is to lay out
every possible pair of vertices in a grid: an **adjacency matrix**.

## The idea

Number the vertices 0 to V−1. Make a V × V grid of numbers. Cell `[i][j]` holds the
weight of the edge from vertex i to vertex j, or 0 if there's no edge. (For an
unweighted graph, use `bool` or 1s and 0s.)

Our world, with walking times:

```
             Village Forest Lake Caves Mountain Castle
  Village       0      5     3     0      0       0
  Forest        5      0     4    15      0       0
  Lake          3      4     0     0      8       0
  Caves         0     15     0     0      0       6
  Mountain      0      0     8     0      0      12
  Castle        0      0     0     6     12       0
```

The matrix is symmetric across its diagonal because the graph is undirected: each
path is stored as `[i][j]` *and* `[j][i]`. The portal from the Castle to the Village
would set only `[Castle][Village]`, breaking the symmetry.

Using 0 for "no edge" only works if 0 is never a real weight. Our portal takes 0
minutes, so a real game would store `-1`, use a separate `bool` grid, or use
`math.MaxInt` for "unreachable".

## In Go

Go has no built-in 2D array type of dynamic size, so we use a slice of slices. The
zone names live in a slice, and a map translates names to indexes:

```go
package main

import "fmt"

type MatrixGraph struct {
	names []string
	index map[string]int
	w     [][]int // w[i][j] = minutes from i to j, 0 = no path
}

func NewMatrixGraph(names ...string) *MatrixGraph {
	g := &MatrixGraph{names: names, index: map[string]int{}}
	g.w = make([][]int, len(names))
	for i, n := range names {
		g.index[n] = i
		g.w[i] = make([]int, len(names))
	}
	return g
}

func (g *MatrixGraph) AddPath(a, b string, minutes int) {
	i, j := g.index[a], g.index[b]
	g.w[i][j] = minutes
	g.w[j][i] = minutes
}

func (g *MatrixGraph) Connected(a, b string) bool {
	return g.w[g.index[a]][g.index[b]] != 0
}

func (g *MatrixGraph) Neighbors(a string) []string {
	var out []string
	for j, minutes := range g.w[g.index[a]] { // scan the whole row: O(V)
		if minutes != 0 {
			out = append(out, g.names[j])
		}
	}
	return out
}

func main() {
	g := NewMatrixGraph("Village", "Forest", "Lake", "Caves", "Mountain", "Castle")
	g.AddPath("Village", "Forest", 5)
	g.AddPath("Village", "Lake", 3)
	g.AddPath("Lake", "Forest", 4)
	g.AddPath("Forest", "Caves", 15)
	g.AddPath("Caves", "Castle", 6)
	g.AddPath("Lake", "Mountain", 8)
	g.AddPath("Mountain", "Castle", 12)

	fmt.Println(g.Connected("Lake", "Forest"), g.Connected("Village", "Castle"))
	fmt.Println(g.Neighbors("Lake"))
	fmt.Println(g.w[g.index["Castle"]])
}
```

Output:

```
true false
[Village Forest Mountain]
[0 0 0 6 12 0]
```

Each row gets its own `make`. Writing `make([][]int, n)` alone gives you n *nil* rows,
and `g.w[i][j] = ...` would panic with an index out of range.

Watch out for a subtle bug: `g.index[a]` returns 0 for an unknown name, which is the
Village's index! A typo like `g.Connected("Vilage", "Forest")` silently checks the
Village instead. Real code should use the comma-ok form and report unknown zones.

## Trade-offs

| Operation | Adjacency matrix |
|---|---|
| Is there an edge i → j? | **O(1)** |
| List neighbours of i | O(V) |
| Add or remove an edge | O(1) |
| Add a vertex | O(V²), rebuild the grid |
| Memory | **O(V²)** |

The matrix shines for **dense** graphs, where most pairs are connected, and for small
graphs where O(1) edge checks matter. But most real graphs are **sparse**: a zone
connects to a handful of neighbours, not to every other zone in the world. For those,
O(V²) memory is a disaster, and the next representation wins.
