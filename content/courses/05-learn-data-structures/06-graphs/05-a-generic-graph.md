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
exercise:
  starter: |
    package main

    import "fmt"

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

    // HasEdge reports whether there's an edge from -> to.
    func (g *Graph[T]) HasEdge(from, to T) bool {
    	// ?
    	return false
    }

    // RemoveEdge deletes the edge from -> to (and to -> from, if the graph is
    // undirected). Both vertices stay in the graph. Removing a missing edge
    // does nothing.
    func (g *Graph[T]) RemoveEdge(from, to T) {
    	// ?
    }

    // Reversed returns a NEW graph with every edge flipped: a -> b becomes
    // b -> a with the same weight. It has the same vertices in the same order,
    // and it must not change g. (Flipping an undirected graph gives an equal copy.)
    func (g *Graph[T]) Reversed() *Graph[T] {
    	// ?
    	return NewGraph[T](g.directed)
    }

    func main() {
    	quests := NewGraph[string](true) // "finish A before B"
    	quests.AddEdge("Goblin Hunt", "Dragon Slayer", 1)
    	quests.AddEdge("Fetch Herbs", "Dragon Slayer", 1)
    	quests.AddEdge("Goblin Hunt", "Orc Siege", 1)

    	fmt.Println(quests.HasEdge("Goblin Hunt", "Dragon Slayer"), quests.HasEdge("Dragon Slayer", "Goblin Hunt"))
    	// want: true false

    	needs := quests.Reversed() // "B needs A"
    	fmt.Println(needs.adj["Dragon Slayer"], needs.order)
    	// want: [{Goblin Hunt 1} {Fetch Herbs 1}] [Goblin Hunt Dragon Slayer Fetch Herbs Orc Siege]

    	quests.RemoveEdge("Goblin Hunt", "Dragon Slayer")
    	fmt.Println(quests.HasEdge("Goblin Hunt", "Dragon Slayer"), len(quests.adj["Goblin Hunt"]))
    	// want: false 1
    }
  solution: |
    package main

    import (
    	"fmt"
    	"slices"
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

    func (g *Graph[T]) HasEdge(from, to T) bool {
    	return slices.ContainsFunc(g.adj[from], func(e Edge[T]) bool { return e.To == to })
    }

    func (g *Graph[T]) RemoveEdge(from, to T) {
    	if _, ok := g.adj[from]; ok {
    		g.adj[from] = slices.DeleteFunc(g.adj[from], func(e Edge[T]) bool { return e.To == to })
    	}
    	if !g.directed {
    		if _, ok := g.adj[to]; ok {
    			g.adj[to] = slices.DeleteFunc(g.adj[to], func(e Edge[T]) bool { return e.To == from })
    		}
    	}
    }

    func (g *Graph[T]) Reversed() *Graph[T] {
    	r := NewGraph[T](g.directed)
    	for _, v := range g.order {
    		r.AddVertex(v)
    	}
    	if !g.directed {
    		for _, v := range g.order {
    			r.adj[v] = slices.Clone(g.adj[v])
    		}
    		return r
    	}
    	for _, from := range g.order {
    		for _, e := range g.adj[from] {
    			r.adj[e.To] = append(r.adj[e.To], Edge[T]{from, e.Weight})
    		}
    	}
    	return r
    }

    func main() {
    	quests := NewGraph[string](true) // "finish A before B"
    	quests.AddEdge("Goblin Hunt", "Dragon Slayer", 1)
    	quests.AddEdge("Fetch Herbs", "Dragon Slayer", 1)
    	quests.AddEdge("Goblin Hunt", "Orc Siege", 1)

    	fmt.Println(quests.HasEdge("Goblin Hunt", "Dragon Slayer"), quests.HasEdge("Dragon Slayer", "Goblin Hunt"))
    	// want: true false

    	needs := quests.Reversed() // "B needs A"
    	fmt.Println(needs.adj["Dragon Slayer"], needs.order)
    	// want: [{Goblin Hunt 1} {Fetch Herbs 1}] [Goblin Hunt Dragon Slayer Fetch Herbs Orc Siege]

    	quests.RemoveEdge("Goblin Hunt", "Dragon Slayer")
    	fmt.Println(quests.HasEdge("Goblin Hunt", "Dragon Slayer"), len(quests.adj["Goblin Hunt"]))
    	// want: false 1
    }
  tests: |
    package main

    import (
    	"maps"
    	"slices"
    	"testing"
    )

    func quests() *Graph[string] {
    	g := NewGraph[string](true)
    	g.AddEdge("hunt", "dragon", 5)
    	g.AddEdge("herbs", "dragon", 2)
    	g.AddEdge("hunt", "siege", 3)
    	g.AddVertex("tutorial")
    	return g
    }

    func snapshot[T comparable](g *Graph[T]) (map[T][]Edge[T], []T) {
    	adj := map[T][]Edge[T]{}
    	for k, v := range g.adj {
    		adj[k] = slices.Clone(v)
    	}
    	return adj, slices.Clone(g.order)
    }

    func sameAdj[T comparable](a, b map[T][]Edge[T]) bool {
    	return maps.EqualFunc(a, b, func(x, y []Edge[T]) bool { return slices.Equal(x, y) })
    }

    func TestHasEdge(t *testing.T) {
    	g := quests()
    	for _, tt := range []struct {
    		from, to string
    		want     bool
    	}{{"hunt", "dragon", true}, {"dragon", "hunt", false}, {"herbs", "siege", false}, {"tutorial", "hunt", false}, {"nowhere", "hunt", false}} {
    		if got := g.HasEdge(tt.from, tt.to); got != tt.want {
    			t.Errorf("directed quest graph: HasEdge(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
    		}
    	}
    	u := NewGraph[int](false)
    	u.AddEdge(1, 2, 7)
    	if !u.HasEdge(1, 2) || !u.HasEdge(2, 1) {
    		t.Error("undirected graph with edge 1-2: HasEdge should be true both ways")
    	}
    }

    func TestRemoveEdge(t *testing.T) {
    	g := quests()
    	g.RemoveEdge("hunt", "dragon")
    	if g.HasEdge("hunt", "dragon") {
    		t.Error("after RemoveEdge(hunt, dragon), HasEdge(hunt, dragon) is still true")
    	}
    	if want := []Edge[string]{{"siege", 3}}; !slices.Equal(g.adj["hunt"], want) {
    		t.Errorf("after RemoveEdge(hunt, dragon), hunt's edges = %v, want %v", g.adj["hunt"], want)
    	}
    	if want := []string{"hunt", "dragon", "herbs", "siege", "tutorial"}; !slices.Equal(g.order, want) {
    		t.Errorf("RemoveEdge changed the vertices to %v, want %v (vertices stay)", g.order, want)
    	}
    	g.RemoveEdge("herbs", "siege") // no such edge
    	g.RemoveEdge("ghost", "hunt")  // no such vertex
    	if _, ok := g.adj["ghost"]; ok {
    		t.Error(`RemoveEdge("ghost", "hunt") added "ghost" as a vertex`)
    	}
    	if len(g.adj["herbs"]) != 1 {
    		t.Errorf("removing a missing edge changed herbs' edges to %v", g.adj["herbs"])
    	}

    	u := NewGraph[string](false)
    	u.AddEdge("village", "lake", 3)
    	u.AddEdge("village", "forest", 5)
    	u.RemoveEdge("lake", "village")
    	if u.HasEdge("village", "lake") || u.HasEdge("lake", "village") {
    		t.Error("undirected: RemoveEdge(lake, village) must remove both directions")
    	}
    	if !u.HasEdge("forest", "village") {
    		t.Error("undirected: RemoveEdge(lake, village) removed the wrong edge")
    	}
    }

    func TestReversed(t *testing.T) {
    	g := quests()
    	adj, order := snapshot(g)
    	r := g.Reversed()
    	if !sameAdj(g.adj, adj) || !slices.Equal(g.order, order) {
    		t.Fatalf("Reversed() changed the original graph: adj %v, order %v", g.adj, g.order)
    	}
    	if r == g {
    		t.Fatal("Reversed() returned the same graph; it must build a new one")
    	}
    	want := map[string][]Edge[string]{
    		"hunt":     nil,
    		"dragon":   {{"hunt", 5}, {"herbs", 2}},
    		"herbs":    nil,
    		"siege":    {{"hunt", 3}},
    		"tutorial": nil,
    	}
    	if !sameAdj(r.adj, want) {
    		t.Errorf("Reversed().adj = %v, want %v", r.adj, want)
    	}
    	if !slices.Equal(r.order, order) {
    		t.Errorf("Reversed().order = %v, want %v", r.order, order)
    	}
    	if !r.directed {
    		t.Error("the reverse of a directed graph should be directed")
    	}
    	r.AddEdge("tutorial", "hunt", 1)
    	if g.HasEdge("tutorial", "hunt") || len(g.adj["hunt"]) != 2 {
    		t.Error("adding an edge to the reversed graph changed the original")
    	}

    	u := NewGraph[int](false)
    	u.AddEdge(1, 2, 4)
    	u.AddEdge(2, 3, 6)
    	ru := u.Reversed()
    	if !sameAdj(ru.adj, u.adj) || ru.directed {
    		t.Errorf("reversing an undirected graph should give an equal undirected copy, got %v", ru.adj)
    	}
    	ru.AddEdge(1, 3, 9)
    	if u.HasEdge(1, 3) || len(u.adj[1]) != 1 {
    		t.Error("changing the reversed undirected graph changed the original: copy the edge slices")
    	}
    }
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

## Your turn: more graph methods

The quest designers want three more methods on `Graph[T]`:

- **`HasEdge(from, to)`** reports whether there's an edge from `from` to `to`.
  `slices.ContainsFunc` makes it a one-liner.
- **`RemoveEdge(from, to)`** deletes that edge, and also `to → from` if the
  graph is undirected. Vertices stay put, even if they're left with no edges,
  and removing an edge that isn't there does nothing (in particular, it must
  not add a missing vertex to `adj`). `slices.DeleteFunc` is handy.
- **`Reversed()`** returns a **new** graph with every edge flipped, so
  "Goblin Hunt unlocks Dragon Slayer" becomes "Dragon Slayer needs Goblin Hunt".
  Keep the same vertices in the same `order`, the same weights and the same
  `directed` flag, and don't touch `g`. Reversing an undirected graph gives an
  equal, independent copy (clone the edge slices so changes to one graph don't
  leak into the other).

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
