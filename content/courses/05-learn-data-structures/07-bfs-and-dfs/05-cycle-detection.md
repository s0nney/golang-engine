---
title: Cycle Detection
quiz:
  - question: In the three-colour DFS for directed graphs, what does reaching a **grey** vertex mean?
    options:
      - text: The vertex was already fully explored, so skip it
      - text: The vertex is on the current path, so you've found a cycle
        correct: true
      - text: The vertex has never been seen
      - text: The graph is undirected
    explanation: |
      Grey means "entered but not finished", so it's an ancestor on the current DFS
      path. An edge back to an ancestor closes a loop. Black vertices are finished, and
      reaching one just means two paths meet, which isn't a cycle.
  - question: |
      Quests: `A → C`, `B → C`, where `X → Y` means X must be finished before Y.
      Does `HasCycle` report a cycle?
    options:
      - text: Yes, because C is reached twice
      - text: No; C is black by the time it's reached the second time, and two routes meeting isn't a cycle
        correct: true
      - text: Yes, because there are two edges into C
      - text: It depends on map iteration order
    explanation: |
      DFS from A turns C grey then black. DFS from B finds C already black and moves on.
      A plain `seen` set would wrongly flag this "diamond" shape, which is why directed
      graphs need three colours rather than two.
exercise:
  starter: |
    package main

    import "fmt"

    type color int

    const (
    	white color = iota // unvisited
    	grey               // on the current DFS path
    	black              // finished
    )

    // QuestOrder returns every quest ordered so that each one comes after all of
    // its prerequisites (an edge X -> Y means X must be done before Y), and true.
    // If the prerequisites contain a cycle it returns nil, false.
    func QuestOrder(prereqs map[string][]string) ([]string, bool) {
    	// ?
    	// Tip: start a DFS from each quest in slices.Sorted(maps.Keys(prereqs))
    	// (import "maps" and "slices") so the result is the same every run.
    	return nil, true
    }

    func main() {
    	quests := map[string][]string{
    		"Goblin Hunt":   {"Dragon Slayer"},
    		"Wolf Pack":     {"Dragon Slayer"},
    		"Dragon Slayer": {"Crown the King"},
    	}
    	order, ok := QuestOrder(quests)
    	fmt.Println(order, ok) // want a valid order ending in Crown the King, true

    	quests["Crown the King"] = []string{"Goblin Hunt"} // oops, a cycle
    	order, ok = QuestOrder(quests)
    	fmt.Println(order, ok) // want [] false
    }
  solution: |
    package main

    import (
    	"fmt"
    	"maps"
    	"slices"
    )

    type color int

    const (
    	white color = iota
    	grey
    	black
    )

    func QuestOrder(prereqs map[string][]string) ([]string, bool) {
    	colors := map[string]color{}
    	var order []string
    	var visit func(q string) bool
    	visit = func(q string) bool {
    		colors[q] = grey
    		for _, next := range prereqs[q] {
    			switch colors[next] {
    			case grey:
    				return false
    			case white:
    				if !visit(next) {
    					return false
    				}
    			}
    		}
    		colors[q] = black
    		order = append(order, q)
    		return true
    	}
    	for _, q := range slices.Sorted(maps.Keys(prereqs)) {
    		if colors[q] == white && !visit(q) {
    			return nil, false
    		}
    	}
    	slices.Reverse(order)
    	return order, true
    }

    func main() {
    	quests := map[string][]string{
    		"Goblin Hunt":   {"Dragon Slayer"},
    		"Wolf Pack":     {"Dragon Slayer"},
    		"Dragon Slayer": {"Crown the King"},
    	}
    	order, ok := QuestOrder(quests)
    	fmt.Println(order, ok) // want a valid order ending in Crown the King, true

    	quests["Crown the King"] = []string{"Goblin Hunt"} // oops, a cycle
    	order, ok = QuestOrder(quests)
    	fmt.Println(order, ok) // want [] false
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    // checkOrder verifies order contains every quest once and respects every edge.
    func checkOrder(t *testing.T, prereqs map[string][]string, order []string) {
    	t.Helper()
    	pos := map[string]int{}
    	for i, q := range order {
    		if _, dup := pos[q]; dup {
    			t.Errorf("quest %q appears twice in %v", q, order)
    		}
    		pos[q] = i
    	}
    	for q, nexts := range prereqs {
    		if _, ok := pos[q]; !ok {
    			t.Errorf("quest %q is missing from %v", q, order)
    		}
    		for _, n := range nexts {
    			pn, ok := pos[n]
    			if !ok {
    				t.Errorf("quest %q is missing from %v", n, order)
    				continue
    			}
    			if pos[q] > pn {
    				t.Errorf("%q must come before %q, but got %v", q, n, order)
    			}
    		}
    	}
    }

    func TestSimpleChain(t *testing.T) {
    	prereqs := map[string][]string{
    		"Goblin Hunt":   {"Dragon Slayer"},
    		"Dragon Slayer": {"Crown the King"},
    	}
    	order, ok := QuestOrder(prereqs)
    	if !ok {
    		t.Fatalf("QuestOrder reported a cycle in a simple chain")
    	}
    	if want := []string{"Goblin Hunt", "Dragon Slayer", "Crown the King"}; !slices.Equal(order, want) {
    		t.Errorf("QuestOrder = %v, want %v", order, want)
    	}
    }

    func TestDiamondIsNotACycle(t *testing.T) {
    	prereqs := map[string][]string{
    		"Goblin Hunt":   {"Dragon Slayer"},
    		"Wolf Pack":     {"Dragon Slayer"},
    		"Dragon Slayer": {"Crown the King"},
    		"Tutorial":      {"Goblin Hunt", "Wolf Pack"},
    	}
    	order, ok := QuestOrder(prereqs)
    	if !ok {
    		t.Fatalf("QuestOrder reported a cycle, but two quests sharing a follow-up is not a cycle")
    	}
    	if len(order) != 5 {
    		t.Errorf("QuestOrder returned %d quests, want 5: %v", len(order), order)
    	}
    	checkOrder(t, prereqs, order)
    }

    func TestCycle(t *testing.T) {
    	prereqs := map[string][]string{
    		"Goblin Hunt":    {"Dragon Slayer"},
    		"Dragon Slayer":  {"Crown the King"},
    		"Crown the King": {"Goblin Hunt"},
    	}
    	if order, ok := QuestOrder(prereqs); ok || order != nil {
    		t.Errorf("QuestOrder = %v, %v; want nil, false for a cycle", order, ok)
    	}
    }

    func TestSelfLoop(t *testing.T) {
    	prereqs := map[string][]string{"Fishing": {"Fishing"}}
    	if order, ok := QuestOrder(prereqs); ok || order != nil {
    		t.Errorf("QuestOrder = %v, %v; want nil, false when a quest requires itself", order, ok)
    	}
    }

    func TestCycleInSeparatePart(t *testing.T) {
    	prereqs := map[string][]string{
    		"Alpha":   {"Beta"},
    		"Fishing": {"Cooking"},
    		"Cooking": {"Fishing"},
    	}
    	if order, ok := QuestOrder(prereqs); ok || order != nil {
    		t.Errorf("QuestOrder = %v, %v; want nil, false (Fishing and Cooking require each other)", order, ok)
    	}
    }

    func TestEmpty(t *testing.T) {
    	order, ok := QuestOrder(map[string][]string{})
    	if !ok || len(order) != 0 {
    		t.Errorf("QuestOrder(empty) = %v, %v; want an empty order, true", order, ok)
    	}
    }
---

The quest designers have been busy. Quest prerequisites form a directed graph: an edge
`Goblin Hunt → Dragon Slayer` means you must finish Goblin Hunt before Dragon Slayer
unlocks. If someone accidentally makes Dragon Slayer a prerequisite of Goblin Hunt too,
neither quest can ever be started, and players will be furious. Before each release, we
want to **detect cycles** automatically. DFS is the tool.

## Directed graphs: three colours

Give every vertex a colour:

- **White**: not visited yet.
- **Grey**: visit started, but not finished. It's on the current DFS path.
- **Black**: fully explored, along with everything reachable from it.

Run DFS. When you follow an edge `v → n`:

- `n` is **white**: recurse into it.
- `n` is **grey**: `n` is an ancestor on the current path, so this edge loops back. **Cycle!**
- `n` is **black**: `n` was finished earlier by another route. No cycle through here.

Why not just a `seen` set? Consider `Goblin Hunt → Dragon Slayer` and `Wolf Pack →
Dragon Slayer`. DFS from Goblin Hunt finishes Dragon Slayer. Later, DFS from Wolf Pack
reaches Dragon Slayer again. It's been seen, but that's not a cycle, just two routes
meeting. Grey versus black tells the two cases apart.

```go
package main

import "fmt"

type color int

const (
	white color = iota // zero value: unvisited
	grey               // on the current path
	black              // finished
)

// HasCycle reports whether the directed graph contains a cycle.
func HasCycle[T comparable](adj map[T][]T) bool {
	colors := map[T]color{}
	var visit func(v T) bool
	visit = func(v T) bool {
		colors[v] = grey
		for _, n := range adj[v] {
			switch colors[n] {
			case grey:
				return true // back edge: cycle
			case white:
				if visit(n) {
					return true
				}
			}
		}
		colors[v] = black
		return false
	}
	for v := range adj { // start from every vertex: the graph may be disconnected
		if colors[v] == white && visit(v) {
			return true
		}
	}
	return false
}

func main() {
	quests := map[string][]string{
		"Goblin Hunt":   {"Dragon Slayer"},
		"Wolf Pack":     {"Dragon Slayer"},
		"Dragon Slayer": {"Crown the King"},
	}
	fmt.Println(HasCycle(quests))

	quests["Crown the King"] = []string{"Goblin Hunt"} // oops
	fmt.Println(HasCycle(quests))
}
```

Output:

```
false
true
```

Notes:

- `white` is `0`, the zero value of `color`, so `colors[n]` for a never-seen vertex is
  automatically white. `iota` numbers the constants 0, 1, 2.
- The outer loop starts a DFS from every white vertex, because a cycle might live in a
  part of the graph unreachable from wherever we'd start. Ranging over the map gives a
  random start order, but the *answer* doesn't depend on it.
- It's O(V + E): each vertex is coloured grey and black once, and each edge is checked once.

## Undirected graphs: ignore your parent

In an undirected graph, every edge is stored both ways, so from the Forest you can
"see" the Village you just came from. That isn't a cycle. Pass along the parent and
skip it; any *other* already-seen neighbour means a cycle:

```go
func hasCycleUndirected(adj map[string][]string) bool {
	seen := map[string]bool{}
	var visit func(v, parent string) bool
	visit = func(v, parent string) bool {
		seen[v] = true
		for _, n := range adj[v] {
			if n == parent {
				continue
			}
			if seen[n] || visit(n, v) {
				return true
			}
		}
		return false
	}
	for v := range adj {
		if !seen[v] && visit(v, "") {
			return true
		}
	}
	return false
}
```

Our world map has cycles (Village ─ Forest ─ Lake ─ Village), which is good game design:
players can loop around instead of backtracking. (This simple version assumes no two zones
are joined by duplicate paths, and that no zone is named `""`.)

## Beyond yes or no

Once you know the quest graph is a DAG, the **finishing order** of this same DFS is
useful: if you record each vertex when it turns black and reverse the list, you get a
**topological order**, a sequence where every quest appears after all its prerequisites.
That's how build systems (including `go build`) decide what to compile first.

## Your turn

The release tool needs more than a yes or no. Complete `QuestOrder` in the exercise.
Given the prerequisite graph (`X → Y` means X must be finished before Y), it returns
every quest in an order where each quest comes after all of its prerequisites, plus
`true`. If the graph has a cycle, it returns `nil, false`.

Use the three-colour DFS: a grey neighbour means a cycle. Append each quest when it
turns **black**, then reverse the whole list at the end. Make sure quests that only
appear as someone's follow-up (with no key of their own in the map) are included, and
that a quest listed as its own prerequisite counts as a cycle.
