---
title: Breadth-First Search
quiz:
  - question: |
      Using the world map and `BFS` from this lesson, what does
      `BFS(world, "Castle")` return?
    options:
      - text: '`[Castle Caves Mountain Forest Lake Village]`'
        correct: true
      - text: '`[Castle Caves Forest Village Lake Mountain]`'
      - text: '`[Castle Mountain Caves Lake Forest Village]`'
      - text: '`[Castle Caves Mountain Village Forest Lake]`'
    explanation: |
      Castle's neighbours, in order, are Caves then Mountain. Next Caves adds Forest
      (Castle is already seen), then Mountain adds Lake. Finally Forest adds Village.
      Distance 0, then everything at distance 1, then 2, then 3.
  - question: Why does BFS mark a vertex as *seen* when it's **added to the queue**, not when it's taken out?
    options:
      - text: It's required by Go's map semantics
      - text: Otherwise the same vertex could be enqueued several times, once by each neighbour that finds it before it's processed
        correct: true
      - text: So that the queue stays sorted
      - text: It doesn't matter; both give identical work
    explanation: |
      In our map, both Village and Lake discover Forest. If Forest were only marked when
      dequeued, it would be enqueued twice and processed twice. On big, dense graphs that
      duplication can blow up badly. Marking on enqueue keeps the work at O(V + E).
  - question: What's the time complexity of BFS on an adjacency list?
    options:
      - text: O(V)
      - text: O(E)
      - text: O(V + E)
        correct: true
      - text: O(V · E)
    explanation: |
      Each vertex is enqueued and dequeued once, which is O(V), and each vertex's
      neighbour list is scanned once, which touches every edge (twice for undirected
      edges), which is O(E).
---

A new player spawns in the Village. The quest-hint system wants to list zones in order
of **how many paths away** they are: nearby zones first, far-off ones later. That's
exactly the order **breadth-first search** (BFS) visits a graph.

## The idea

BFS explores in rings, like a ripple in a pond:

1. Start at the source (distance 0).
2. Visit all its neighbours (distance 1).
3. Then all *their* unvisited neighbours (distance 2).
4. And so on, until nothing new is reachable.

The tool that makes this work is a **queue**, first in, first out, which you built in
DSA 1. Vertices discovered earlier are processed earlier, so every distance-1 zone is
handled before any distance-2 zone.

Graphs have cycles, so BFS also needs a **seen** set. Without it, the Village would
queue the Forest, the Forest would queue the Village again, and you'd go round forever.

## In Go

We'll keep adjacency lists in a `map[T][]T` and write BFS as a generic function.

```go
package main

import "fmt"

// BFS returns every vertex reachable from start, nearest first.
func BFS[T comparable](adj map[T][]T, start T) []T {
	seen := map[T]bool{start: true}
	queue := []T{start}
	var order []T
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:] // dequeue
		order = append(order, v)
		for _, n := range adj[v] {
			if !seen[n] {
				seen[n] = true // mark when enqueued, not when dequeued
				queue = append(queue, n)
			}
		}
	}
	return order
}

func main() {
	world := map[string][]string{
		"Village":  {"Forest", "Lake"},
		"Forest":   {"Village", "Lake", "Caves"},
		"Lake":     {"Village", "Forest", "Mountain"},
		"Caves":    {"Forest", "Castle"},
		"Mountain": {"Lake", "Castle"},
		"Castle":   {"Caves", "Mountain"},
		"Island":   {}, // no paths: unreachable
	}
	fmt.Println(BFS(world, "Village"))
}
```

Output:

```
[Village Forest Lake Caves Mountain Castle]
```

Step by step, with the queue after each dequeue:

| Dequeued | Newly queued | Queue |
|---|---|---|
| Village | Forest, Lake | Forest, Lake |
| Forest | Caves (Village, Lake already seen) | Lake, Caves |
| Lake | Mountain | Caves, Mountain |
| Caves | Castle | Mountain, Castle |
| Mountain | (nothing new) | Castle |
| Castle | (nothing new) | empty |

The Island never shows up, since no path leads there. BFS only finds what's
**reachable** from the start. Run it from every unseen vertex in turn and you can
count a graph's *connected components*.

## Go details

- `seen := map[T]bool{start: true}` relies on missing keys reading as `false`, so
  `!seen[n]` works for vertices we've never touched.
- `queue = queue[1:]` is a simple slice-based queue. It never frees the front of the
  backing array until the whole slice is dropped, which is fine for a single search.
  For a long-running queue, use a ring buffer or `container/list`.
- The loop over `adj[v]` visits neighbours in slice order, so the result is
  deterministic. If adjacency were a set (`map[T]bool`), the order among zones at the
  same distance would change from run to run.

## Cost

Every reachable vertex is enqueued once and dequeued once: O(V). Every adjacency list
is scanned once, touching each edge: O(E). Total: **O(V + E)**, the best you can hope
for, since you have to at least look at the graph.
