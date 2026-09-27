---
title: Vertices and Edges
quiz:
  - question: In the world map from this lesson, what's the **degree** of `Lake`?
    options:
      - text: '1'
      - text: '2'
      - text: '3'
        correct: true
      - text: '4'
    explanation: |
      Lake has paths to Village, Forest and Mountain, so three edges touch it.
      Degree is the number of edges connected to a vertex.
  - question: How is a graph different from a tree?
    options:
      - text: Graphs can't contain cycles
      - text: 'A tree is a special kind of graph: connected, with no cycles and exactly one path between any two nodes. A general graph can have cycles, many paths, or disconnected parts'
        correct: true
      - text: Graphs must have a root vertex
      - text: Trees can have weights on their edges but graphs can't
    explanation: |
      Every tree is a graph, but not every graph is a tree. Graphs drop the rules
      about roots, parents and a single path, which is exactly why they can model
      maps, social networks and road systems.
  - question: |
      What does this print?

      ```go
      edges := []Edge{{"Village", "Forest"}, {"Village", "Lake"}, {"Lake", "Forest"}}
      deg := map[string]int{}
      for _, e := range edges {
          deg[e.A]++
          deg[e.B]++
      }
      fmt.Println(deg["Forest"], len(deg))
      ```
    options:
      - text: '`1 3`'
      - text: '`2 3`'
        correct: true
      - text: '`2 6`'
      - text: '`3 3`'
    explanation: |
      Forest appears in two edges, so its degree is 2. The map has three distinct keys:
      Village, Forest and Lake. Incrementing a missing key works because a missing
      `int` reads as 0.
---

The game world is a set of **zones**: a starting village, a forest, some caves, a
lake, a mountain pass and a castle. Some zones are joined by paths. Players ask
questions like "can I get from the Village to the Castle?" and "what's the shortest
route?" Trees can't model this, because the paths form loops and there's no natural
root. It's a **graph**.

## Vertices and edges

A graph is a set of **vertices** (also called nodes) and a set of **edges**
connecting pairs of vertices. Here's the world map:

```
   Village ─────── Forest ─────── Caves
      │          /                  │
      │        /                    │
     Lake ───┘                    Castle
      │                             │
      └──────── Mountain ───────────┘
```

Six vertices (zones) and seven edges (paths):

| Edge | |
|---|---|
| Village ─ Forest | Village ─ Lake |
| Lake ─ Forest | Forest ─ Caves |
| Caves ─ Castle | Lake ─ Mountain |
| Mountain ─ Castle | |

## Vocabulary

- **Adjacent** / **neighbours**: two vertices joined by an edge. Forest's neighbours
  are Village, Lake and Caves.
- **Degree**: the number of edges touching a vertex. Forest has degree 3.
- **Path**: a sequence of vertices where each consecutive pair is joined by an edge.
  Village → Lake → Mountain → Castle is a path of length 3 (three edges).
- **Cycle**: a path that starts and ends at the same vertex without reusing an edge.
  Village → Forest → Lake → Village is a cycle.
- **Connected**: every vertex can reach every other vertex by some path. Our world is
  connected. If a zone had no paths at all, it would be isolated.

A tree is a graph that's connected and has **no cycles**. That's why there's exactly
one path between any two nodes in a tree. Graphs drop that restriction, which makes
them more powerful and their algorithms a little more careful: you have to avoid
going round in circles.

## The simplest representation: a list of edges

You can store a graph as a plain slice of edges:

```go
package main

import "fmt"

type Edge struct{ A, B string }

func main() {
	world := []Edge{
		{"Village", "Forest"}, {"Village", "Lake"}, {"Lake", "Forest"},
		{"Forest", "Caves"}, {"Caves", "Castle"}, {"Lake", "Mountain"},
		{"Mountain", "Castle"},
	}

	degree := map[string]int{}
	for _, e := range world {
		degree[e.A]++
		degree[e.B]++
	}
	fmt.Println(len(degree), "zones,", len(world), "paths")
	fmt.Println("Forest has degree", degree["Forest"])

	// Who are Forest's neighbours? We have to scan every edge.
	for _, e := range world {
		switch "Forest" {
		case e.A:
			fmt.Println(" ", e.B)
		case e.B:
			fmt.Println(" ", e.A)
		}
	}
}
```

Output:

```
6 zones, 7 paths
Forest has degree 3
  Village
  Lake
  Caves
```

The `switch "Forest"` form compares one value against several cases, which reads
nicely here. Also notice that adding up all the degrees gives 14, exactly twice the
number of edges, because each edge touches two vertices. That's true of every graph.

An edge list is compact and easy to build, but "who are my neighbours?" means scanning
**every** edge, which is O(E). Graph algorithms ask that question constantly. The next
lessons cover two representations built to answer it fast: the adjacency matrix and
the adjacency list. First, a look at the different kinds of graphs you'll meet.

## Notation

Graph algorithms are usually measured in two numbers: **V**, the number of vertices,
and **E**, the number of edges. You'll see complexities like O(V + E) and O(V²) a lot
in the rest of the course.
