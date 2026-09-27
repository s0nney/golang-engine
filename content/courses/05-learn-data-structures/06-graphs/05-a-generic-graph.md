---
title: A Generic Graph
quiz:
  - question: Why does `Graph[T comparable]` use the `comparable` constraint rather than `cmp.Ordered`?
    options:
      - text: Vertices are used as map keys, which only needs `==`; they never need to be ordered
        correct: true
      - text: '`cmp.Ordered` doesn''t work with maps'
      - text: '`comparable` is faster'
      - text: So that vertices can be slices
    explanation: |
      The graph stores vertices as map keys, and map keys must be comparable. Nothing in
      the graph sorts vertices, so there's no reason to demand `<`. That lets vertices be
      structs like `Zone{X, Y}`, which aren't ordered. (Slices aren't comparable at all.)
  - question: |
      What does this print?

      ```go
      g := NewGraph[string](false)
      g.AddEdge("Village", "Forest", 5)
      g.AddEdge("Village", "Forest", 7)
      fmt.Println(len(g.adj["Village"]))
      ```
    options:
      - text: '1'
      - text: '2'
        correct: true
      - text: '0'
      - text: It panics
    explanation: |
      `AddEdge` appends without checking for an existing edge, so Village ends up with
      two edges to Forest. That's a *multigraph*. Whether that's a bug or a feature
      (two different paths between the same zones) is a design decision; this version
      allows it.
---

Time to package everything into a reusable type. Zones are strings today, but the
engine team wants the same graph for grid cells (`Zone{X, Y}`), NPC IDs (`int`) and
quest names. Generics to the rescue.

## Design

- **Vertices** can be any `comparable` type, because they're map keys.
- **Edges** are weighted. Unweighted graphs just use weight 1.
- **Directed or not** is a flag chosen at construction.
- Neighbours are stored in **slices**, so iteration order is predictable.
- An `order` slice remembers vertices in the order they were added, since ranging
  over a map would be random.

```go
package main

import (
	"fmt"
	"iter"
)

type Edge[T comparable] struct {
	To     T
	Weight int
}

type Graph[T comparable] struct {
	directed bool
	adj      map[T][]Edge[T]
	order    []T // vertices in insertion order
}

func NewGraph[T comparable](directed bool) *Graph[T] {
	return &Graph[T]{directed: directed, adj: map[T][]Edge[T]{}}
}

func (g *Graph[T]) AddVertex(v T) {
	if _, ok := g.adj[v]; !ok {
		g.adj[v] = nil
		g.order = append(g.order, v)
	}
}

func (g *Graph[T]) AddEdge(from, to T, weight int) {
	g.AddVertex(from)
	g.AddVertex(to)
	g.adj[from] = append(g.adj[from], Edge[T]{to, weight})
	if !g.directed {
		g.adj[to] = append(g.adj[to], Edge[T]{from, weight})
	}
}

// Vertices yields every vertex in the order it was added.
func (g *Graph[T]) Vertices() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range g.order {
			if !yield(v) {
				return
			}
		}
	}
}

// Neighbors yields each neighbour of v with the edge's weight.
func (g *Graph[T]) Neighbors(v T) iter.Seq2[T, int] {
	return func(yield func(T, int) bool) {
		for _, e := range g.adj[v] {
			if !yield(e.To, e.Weight) {
				return
			}
		}
	}
}

type Zone struct{ X, Y int }

func main() {
	world := NewGraph[string](false)
	world.AddEdge("Village", "Forest", 5)
	world.AddEdge("Village", "Lake", 3)
	world.AddEdge("Lake", "Mountain", 8)
	for to, minutes := range world.Neighbors("Village") {
		fmt.Printf("Village -> %s (%d min)\n", to, minutes)
	}

	quests := NewGraph[string](true) // prerequisites are one-way
	quests.AddEdge("Goblin Hunt", "Dragon Slayer", 1)
	fmt.Println(len(quests.adj["Goblin Hunt"]), len(quests.adj["Dragon Slayer"]))

	grid := NewGraph[Zone](false)
	grid.AddEdge(Zone{0, 0}, Zone{0, 1}, 1)
	grid.AddEdge(Zone{0, 1}, Zone{1, 1}, 1)
	for v := range grid.Vertices() {
		fmt.Print(v, " ")
	}
	fmt.Println()
}
```

Output:

```
Village -> Forest (5 min)
Village -> Lake (3 min)
1 0
{0 0} {0 1} {1 1} 
```

## Walking through it

- **`AddVertex` uses comma-ok** to add each vertex only once. Storing `nil` as its
  edge list still creates the map key, so an isolated zone with no paths is
  remembered.
- **`AddEdge` adds missing vertices automatically**, so callers don't have to.
- **The directed quest graph** stored `Goblin Hunt → Dragon Slayer` only. Dragon
  Slayer is a vertex with zero outgoing edges.
- **`Neighbors` returns an `iter.Seq2[T, int]`**, so a `range` loop gets two values per
  step, just like ranging over a map, but in a stable order. The `if !yield(...)
  { return }` check honours `break` in the caller's loop.
- **Struct vertices** such as `Zone{0, 1}` work because every field is comparable.

## Cost

| Operation | Cost |
|---|---|
| `AddVertex` | O(1) on average |
| `AddEdge` | O(1) amortized |
| `Neighbors(v)` | O(degree of v) |
| Visiting every vertex and edge | O(V + E) |

This `Graph[T]` is the foundation for the next chapter. Breadth-first search,
depth-first search and Dijkstra's algorithm are all just different ways of walking it.

## Further reading

- [Go by Example: Generics](https://gobyexample.com/generics)
