---
title: Dijkstra's Algorithm
quiz:
  - question: |
      Why does the loop skip an item when `it.dist > dist[it.zone]`?
    options:
      - text: The zone is unreachable
      - text: The heap holds an outdated entry; a shorter route to that zone was already found and processed
        correct: true
      - text: The edge has a negative weight
      - text: To stop the algorithm early
    explanation: |
      Instead of updating an item's priority inside the heap, this version pushes a new
      item every time it finds a better distance. The old, longer entries are still in
      the heap, so when they eventually pop out, they're ignored. It's simpler than
      `heap.Fix` and just as correct.
  - question: When can Dijkstra's algorithm give a **wrong** answer?
    options:
      - text: When the graph has cycles
      - text: When some edges have negative weights
        correct: true
      - text: When two routes have the same length
      - text: When the graph is undirected
    explanation: |
      Dijkstra assumes that once a vertex is settled, nothing later can make its route
      shorter. A negative edge could do exactly that. Cycles and ties are fine. For
      negative weights, use an algorithm like Bellman-Ford.
---

BFS finds the route with the fewest paths. But players care about **time**, and the
paths have different walking times. We want the route with the lowest *total* weight.
That's the job of **Dijkstra's algorithm**, invented by Edsger Dijkstra in 1956
(supposedly in about 20 minutes, while having coffee).

## The idea

Dijkstra is BFS with a **priority queue** instead of a plain queue:

1. Set the start's distance to 0 and push it onto a min-heap.
2. Pop the zone with the **smallest** distance so far. Its distance is now final.
3. For each neighbour, compute `dist[v] + weight`. If that beats the neighbour's best
   known distance, record it (and the parent) and push the neighbour.
4. Repeat until the heap is empty, or until you pop the goal.

Why is the popped distance final? Every other route would have to leave through some
zone still in the heap, and all of those are at least as far away already. Adding more
non-negative edges can only make that route longer.

## In Go with container/heap

```go
package main

import (
	"container/heap"
	"fmt"
	"slices"
)

type Edge struct {
	To      string
	Minutes int
}

type item struct {
	zone string
	dist int
}

type minHeap []item

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i].dist < h[j].dist }
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x any)        { *h = append(*h, x.(item)) }
func (h *minHeap) Pop() any {
	old := *h
	it := old[len(old)-1]
	*h = old[:len(old)-1]
	return it
}

// Fastest returns the quickest route from start to goal and its total time.
func Fastest(adj map[string][]Edge, start, goal string) ([]string, int, bool) {
	dist := map[string]int{start: 0}
	parent := map[string]string{}
	h := &minHeap{{start, 0}}
	for h.Len() > 0 {
		it := heap.Pop(h).(item)
		if it.dist > dist[it.zone] {
			continue // stale entry: we already found a shorter route
		}
		if it.zone == goal {
			path := []string{goal}
			for v := goal; v != start; {
				v = parent[v]
				path = append(path, v)
			}
			slices.Reverse(path)
			return path, it.dist, true
		}
		for _, e := range adj[it.zone] {
			d := it.dist + e.Minutes
			if best, ok := dist[e.To]; !ok || d < best {
				dist[e.To] = d
				parent[e.To] = it.zone
				heap.Push(h, item{e.To, d})
			}
		}
	}
	return nil, 0, false
}

func main() {
	adj := map[string][]Edge{}
	addPath := func(a, b string, minutes int) {
		adj[a] = append(adj[a], Edge{b, minutes})
		adj[b] = append(adj[b], Edge{a, minutes})
	}
	addPath("Village", "Forest", 5)
	addPath("Village", "Lake", 3)
	addPath("Lake", "Forest", 4)
	addPath("Forest", "Caves", 15)
	addPath("Caves", "Castle", 6)
	addPath("Lake", "Mountain", 8)
	addPath("Mountain", "Castle", 12)

	fmt.Println(Fastest(adj, "Village", "Castle"))
	fmt.Println(Fastest(adj, "Lake", "Caves"))
	fmt.Println(Fastest(adj, "Village", "Atlantis"))
}
```

Output:

```
[Village Lake Mountain Castle] 23 true
[Lake Forest Caves] 19 true
[] 0 false
```

BFS picked Village → Forest → Caves → Castle (26 minutes). Dijkstra finds the 23-minute
route through the Lake and the Mountain. (`fmt` prints a nil slice as `[]`.)

## Implementation notes

- **Lazy deletion.** When we find a better route to a zone that's already in the heap,
  we don't update it; we push a new item. The old one pops out later with a larger
  `dist` and is skipped by the `it.dist > dist[it.zone]` check. This avoids tracking
  heap indexes for `heap.Fix`.
- **`dist` doubles as "seen"**. A missing key means "no route found yet", which is why
  the check is `!ok || d < best`, not just `d < best` (a missing key reads as 0).
- **Early exit.** We stop as soon as the goal is popped. Remove that check and you get
  the fastest time to *every* zone.
- **`heap.Pop` returns `any`**, so we need the `.(item)` type assertion. And the heap
  starts non-empty via a composite literal, `&minHeap{{start, 0}}`: a single item is
  already a valid heap, so there's no need for `heap.Init`.

## Cost and limits

With a binary heap, Dijkstra runs in **O((V + E) log V)**: each edge can push one item,
and each push or pop is logarithmic.

Dijkstra requires **non-negative weights**. Our 0-minute portal is fine, but an edge
with −10 minutes could make a settled zone's route shorter later, and the algorithm
wouldn't notice. Games that need it use Bellman-Ford (slower, but handles negatives).
And when you know roughly where the goal is, as with a map with coordinates,
**A\*** adds a distance estimate to the priority to steer the search toward the goal.
That's what most game pathfinding actually uses.
