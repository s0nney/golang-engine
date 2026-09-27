---
title: Depth-First Search
quiz:
  - question: |
      Using the world map and recursive `DFS` from this lesson, what does
      `DFS(world, "Castle")` return?
    options:
      - text: '`[Castle Caves Mountain Forest Lake Village]`'
      - text: '`[Castle Caves Forest Village Lake Mountain]`'
        correct: true
      - text: '`[Castle Mountain Lake Village Forest Caves]`'
      - text: '`[Castle Caves Forest Lake Village Mountain]`'
    explanation: |
      Castle → its first neighbour, Caves → Caves' first unseen neighbour, Forest →
      Forest's first unseen neighbour, Village → Village's next unseen neighbour, Lake →
      Lake's unseen neighbour, Mountain. The first option is the BFS order.
  - question: What plays the role of BFS's queue in recursive DFS?
    options:
      - text: A priority queue
      - text: The call stack; each recursive call is a frame waiting to resume
        correct: true
      - text: The `seen` map
      - text: Nothing; DFS doesn't need to remember anything
    explanation: |
      When `visit` recurses into a neighbour, the current call pauses on the stack.
      Once the neighbour's exploration finishes, the paused call resumes with its next
      neighbour. That's backtracking, and the call stack does the bookkeeping for free.
  - question: What's the time complexity of DFS on an adjacency list?
    options:
      - text: O(V + E), the same as BFS
        correct: true
      - text: O(V²)
      - text: O(E log V)
      - text: O(2^V), because it explores every path
    explanation: |
      DFS visits each vertex once (thanks to `seen`) and scans each adjacency list
      once. It does *not* explore every path; it explores every vertex and edge.
---

BFS explores in rings. **Depth-first search** (DFS) does the opposite: it charges
down one path as far as it can go, and only when it hits a dead end does it
**backtrack** to the last junction and try the next branch. It's how a player explores
a dungeon: keep going until you hit a wall, then go back to the last fork.

## Recursive DFS

Recursion makes DFS almost trivially short, because the call stack remembers where to
backtrack to:

```go
package main

import "fmt"

// DFS returns the vertices reachable from start, in depth-first order.
func DFS[T comparable](adj map[T][]T, start T) []T {
	seen := map[T]bool{}
	var order []T
	var visit func(v T)
	visit = func(v T) {
		seen[v] = true
		order = append(order, v)
		for _, n := range adj[v] {
			if !seen[n] {
				visit(n)
			}
		}
	}
	visit(start)
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
	}
	fmt.Println(DFS(world, "Village"))
}
```

Output:

```
[Village Forest Lake Mountain Castle Caves]
```

Trace it: Village → Forest (its first neighbour) → Lake (Forest's first unseen) →
Mountain → Castle → Caves. Caves' neighbours (Forest, Castle) are all seen, so the
calls unwind. Every paused call finds its remaining neighbours already seen, and the
search ends. Compare BFS from the Village: `[Village Forest Lake Caves Mountain Castle]`.
Same zones, very different order.

## A Go gotcha: recursive closures

`visit` is a closure, so it can use `seen`, `order` and `adj` without passing them
around. But notice the two-step declaration:

```go
var visit func(v T)
visit = func(v T) { ... visit(n) ... }
```

Writing `visit := func(v T) { ... visit(n) ... }` doesn't compile, because inside the
function literal, `visit` isn't declared yet. Declaring the variable first lets the
closure refer to itself.

## BFS vs DFS

| | BFS | DFS |
|---|---|---|
| Data structure | Queue | Stack (or recursion) |
| Visit order | By distance from start | Deep first, then backtrack |
| Shortest path (unweighted) | Yes | No |
| Memory | The widest "ring" | The longest path |
| Cost | O(V + E) | O(V + E) |

DFS doesn't find shortest paths: its first route from the Village to the Caves went
through the Forest, Lake, Mountain and Castle! But DFS is the natural fit whenever you
care about the *structure* of a graph rather than distances:

- **Cycle detection**: are there circular quest prerequisites? (Two lessons from now.)
- **Topological sort**: in what order can a player complete quests?
- **Connected components** and **maze generation**.
- **Exhaustive search**: exploring every possibility in a puzzle, with backtracking.

You've actually used DFS already. The pre-order traversal of a tree in chapter 1, and
walking a trie for autocomplete, were both DFS. Trees just don't need a `seen` set,
because they have no cycles.

## Deep recursion

Each level of recursion is a stack frame. Go's goroutine stacks start small and grow
automatically, so recursing tens of thousands of levels deep is fine. On a graph with
millions of vertices in a long chain, though, you'd rather not rely on that. The next
lesson removes the recursion.
