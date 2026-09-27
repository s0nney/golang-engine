---
title: Directed, Undirected and Weighted
quiz:
  - question: The Castle has a one-way portal that teleports players back to the Village. How do you model it?
    options:
      - text: An undirected edge between Castle and Village
      - text: A directed edge from Castle to Village
        correct: true
      - text: A directed edge from Village to Castle
      - text: Two directed edges, one each way
    explanation: |
      A directed edge only goes one way: from its tail (Castle) to its head (Village).
      Two directed edges, one each way, would be the same as an ordinary two-way path.
  - question: |
      With the travel times from this lesson, how long does the route
      Village → Lake → Forest → Caves take?
    options:
      - text: 3 minutes, one per path
      - text: '22 minutes'
        correct: true
      - text: '20 minutes'
      - text: '26 minutes'
    explanation: |
      Add the weights: Village─Lake is 3, Lake─Forest is 4 and Forest─Caves is 15, which
      makes 22. In a weighted graph the length of a path is the sum of its weights, not
      the number of edges.
  - question: 'Which of these is naturally a **directed** graph?'
    options:
      - text: Two-way hiking paths between zones
      - text: Friendships where both players must accept
      - text: Quest prerequisites ("finish Goblin Hunt before Dragon Slayer")
        correct: true
      - text: Which zones share a border
    explanation: |
      Prerequisites have a direction: Goblin Hunt must come before Dragon Slayer, not
      the other way round. The other three relationships are symmetric, so undirected
      edges fit them.
---

Not every connection is a two-way street. Graphs come in a few flavours, and picking
the right one is the first step in modelling a problem.

## Undirected graphs

In an **undirected** graph, an edge `{A, B}` goes both ways: if the Village connects
to the Forest, the Forest connects to the Village. The world map from the last lesson
is undirected, because players can walk along a path in either direction.

## Directed graphs

In a **directed** graph (a *digraph*), each edge has a direction, `A → B`. You can
follow it from A to B but not back. Examples from our game:

- The Castle has a **one-way portal** back to the Village: `Castle → Village`.
- **Quest prerequisites**: `Goblin Hunt → Dragon Slayer` means you must finish Goblin
  Hunt first.
- **Followers**: mira following kai doesn't mean kai follows mira.

For a vertex in a directed graph, the **out-degree** counts edges leaving it and the
**in-degree** counts edges arriving. An undirected graph is really just a directed
graph where every edge has a twin going the other way, and that's exactly how we'll
store it in code.

## Weighted graphs

In a **weighted** graph, each edge carries a number: a distance, a cost, a time.
Here's our map with walking times in minutes:

```
   Village ───5─── Forest ───15─── Caves
      │          /                  │
      3        4                    6
      │      /                      │
     Lake ──┘                     Castle
      │                             │
      └────8──── Mountain ────12────┘
```

Now "shortest route" has two meanings. The Village → Forest → Caves → Castle route
has the fewest **edges** (3), and so does Village → Lake → Mountain → Castle. But by
**time**, the first takes 5 + 15 + 6 = 26 minutes and the second 3 + 8 + 12 = 23. The
next chapter has algorithms for both: BFS for fewest edges, Dijkstra for lowest total
weight.

## Representing edges in Go

A small struct covers all three flavours:

```go
package main

import "fmt"

type Edge struct {
	From, To string
	Minutes  int // weight; 1 for an unweighted graph
}

func main() {
	paths := []Edge{
		{"Village", "Forest", 5}, {"Village", "Lake", 3}, {"Lake", "Forest", 4},
		{"Forest", "Caves", 15}, {"Caves", "Castle", 6}, {"Lake", "Mountain", 8},
		{"Mountain", "Castle", 12},
	}
	portal := Edge{"Castle", "Village", 0} // directed: one way only

	// An undirected path is two directed edges.
	var directed []Edge
	for _, p := range paths {
		directed = append(directed, p, Edge{p.To, p.From, p.Minutes})
	}
	directed = append(directed, portal)

	out := map[string]int{}
	for _, e := range directed {
		out[e.From]++
	}
	fmt.Println(len(directed), "directed edges")
	fmt.Println("out-degree of Castle:", out["Castle"])
}
```

Output:

```
15 directed edges
out-degree of Castle: 3
```

The seven two-way paths became 14 directed edges, plus the portal makes 15. The Castle
has three ways out (Caves, Mountain, and the portal), but only two ways in.

## A quick classification guide

| Question | If yes |
|---|---|
| Can the relationship go one way only? | Directed |
| Does each connection have a cost, distance or time? | Weighted |
| Can you get back to where you started? | The graph has cycles |
| Is it directed with no cycles? | A **DAG** (directed acyclic graph) |

DAGs deserve a special mention. Quest prerequisites *must* form a DAG: if Quest A
needs B, and B needs A, nobody can ever finish either. Detecting that kind of cycle
is one of the algorithms in the next chapter.
