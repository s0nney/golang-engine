---
title: Iterative DFS
quiz:
  - question: What's the only structural difference between the iterative DFS in this lesson and BFS?
    options:
      - text: DFS uses a stack (take from the end) where BFS uses a queue (take from the front)
        correct: true
      - text: DFS doesn't use a `seen` set
      - text: DFS visits each vertex twice
      - text: DFS needs the graph to be a tree
    explanation: |
      Swap `queue[0]` for `stack[len(stack)-1]` and you've turned BFS's first-in,
      first-out order into DFS's last-in, first-out order. The rest of the loop has the
      same shape.
  - question: |
      Why does the iterative version push neighbours in **reverse** order?
    options:
      - text: To make it faster
      - text: A stack pops the last item pushed first, so pushing in reverse makes the first neighbour come out first, matching the recursive order
        correct: true
      - text: To avoid visiting a vertex twice
      - text: It's required for correctness; without it DFS gives wrong results
    explanation: |
      Without the reversal you still get a perfectly valid depth-first order, just a
      different one: it would explore the *last* neighbour first. Reversing only makes
      it match the recursive version, which is handy for tests.
  - question: In this iterative DFS, why is the `seen` check done when a vertex is **popped**, not when it's pushed?
    options:
      - text: Checking on push would make the stack overflow
      - text: A vertex can be pushed several times; marking on pop means it's visited when it's reached by the deepest route, like the recursive version
        correct: true
      - text: Maps can't be read during a push
      - text: It's arbitrary; both give exactly the same order
    explanation: |
      If you marked on push, a vertex seen early from a shallow neighbour would be
      locked in there, and the order would no longer be a true depth-first order. The
      cost is some duplicate entries on the stack, which we skip with `continue`.
---

Recursive DFS is elegant, but sometimes you want an explicit loop: to avoid very deep
recursion, to pause and resume a search, or to stop early without unwinding a stack of
calls. Replace the call stack with a **stack** of your own.

## From recursion to a stack

```go
package main

import (
	"fmt"
	"slices"
)

// DFS returns the vertices reachable from start, in depth-first order.
func DFS[T comparable](adj map[T][]T, start T) []T {
	seen := map[T]bool{}
	stack := []T{start}
	var order []T
	for len(stack) > 0 {
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1] // pop
		if seen[v] {
			continue // already reached by another route
		}
		seen[v] = true
		order = append(order, v)
		// Push in reverse so the first neighbour is popped first.
		for _, n := range slices.Backward(adj[v]) {
			if !seen[n] {
				stack = append(stack, n)
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
	}
	fmt.Println(DFS(world, "Village"))
}
```

Output:

```
[Village Forest Lake Mountain Castle Caves]
```

Same order as the recursive version. Put this next to BFS and the resemblance is
striking: a `seen` set, a slice of pending vertices, a loop. The only real difference
is which end of the slice you take from:

- **BFS**: `v := queue[0]`, the oldest item. First in, first out.
- **DFS**: `v := stack[len(stack)-1]`, the newest item. Last in, first out.

## The details

- **`slices.Backward`** (Go 1.23) returns an iterator that walks a slice from the end.
  Pushing neighbours in reverse means the first neighbour ends up on top, which
  reproduces the recursive visiting order. Without it you still get a valid DFS, just
  exploring the last neighbour first.
- **Mark on pop, not on push.** A vertex can land on the stack more than once, from
  different neighbours, before it's visited. The `if seen[v] { continue }` skips the
  stale copies. This is the opposite of BFS, where marking on push is correct and
  important. Marking on push here would still visit everything, but not in true
  depth-first order.
- **Memory.** The stack can hold up to O(E) entries, because of those duplicates.
  That's still O(V + E) overall.

## Early exit

A loop makes stopping easy. Say the player wants to know whether *any* zone within
reach sells health potions:

```go
func findShop(adj map[string][]string, start string, hasShop map[string]bool) (string, bool) {
	seen := map[string]bool{}
	stack := []string{start}
	for len(stack) > 0 {
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[v] {
			continue
		}
		seen[v] = true
		if hasShop[v] {
			return v, true // stop the whole search right here
		}
		for _, n := range slices.Backward(adj[v]) {
			if !seen[n] {
				stack = append(stack, n)
			}
		}
	}
	return "", false
}
```

With recursion you'd have to thread a "found" flag back up through every call. Here
it's a plain `return`.

## As an iterator

You can also wrap either search as an `iter.Seq[T]`, yielding each vertex as it's
visited. The loop structure makes that natural: replace `order = append(order, v)`
with `if !yield(v) { return }`, and callers can `range` over the search and `break`
whenever they like.
