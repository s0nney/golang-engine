---
title: Top K
quiz:
  - question: To find the **top 10 highest** scores in a stream, which heap do you keep?
    options:
      - text: A max-heap of all the scores
      - text: A min-heap holding at most 10 scores
        correct: true
      - text: A max-heap holding at most 10 scores
      - text: Two heaps, one min and one max
    explanation: |
      The min-heap's root is the *weakest* of your current top 10, the one to kick out.
      Each new score only needs comparing with that root. A max-heap of everything
      works too, but it has to hold all n scores in memory at once.
  - question: |
      With `k = 3`, the heap holds `[70 90 80]` (root 70). The next score is `60`.
      What happens?
    options:
      - text: 60 is pushed, and the heap grows to 4
      - text: 60 replaces 70 at the root
      - text: Nothing; 60 is lower than the root, so it can't be in the top 3
        correct: true
      - text: 90 is removed
    explanation: |
      The root, 70, is the lowest score still in the top 3. Anything that doesn't beat
      it can't make the cut, so it's skipped with a single comparison.
---

End of season: the studio wants to award prizes to the **top 3** of a million players.
The obvious approach is to sort everyone and take the first three:

```go
slices.SortFunc(players, func(a, b Player) int { return cmp.Compare(b.Score, a.Score) })
top := players[:3]
```

That's O(n log n) and it reorders (or copies) the entire slice. We're doing a lot of
work to put player 700,000 in exactly the right spot, when we only care about the
first three.

## The trick: a small min-heap

Keep a **min-heap of at most k players**. It holds the best k seen so far, and its
root is the *weakest* of them, the one on the bubble.

For each player:

1. If the heap has fewer than k players, push them.
2. Otherwise, if they beat the root, replace the root with them and sift down.
3. Otherwise, skip them. They can't be in the top k.

It feels backwards to use a min-heap to find the maximums. The point is that the heap
must make it cheap to find and evict the *worst* member of the current top k.

```go
package main

import (
	"cmp"
	"container/heap"
	"fmt"
	"math/rand/v2"
	"slices"
)

type Player struct {
	Name  string
	Score int
}

// minHeap keeps the lowest score at index 0.
type minHeap []Player

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i].Score < h[j].Score }
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x any)        { *h = append(*h, x.(Player)) }
func (h *minHeap) Pop() any {
	old := *h
	p := old[len(old)-1]
	*h = old[:len(old)-1]
	return p
}

// TopK returns the k highest-scoring players, best first.
func TopK(players []Player, k int) []Player {
	h := &minHeap{}
	for _, p := range players {
		switch {
		case h.Len() < k:
			heap.Push(h, p)
		case p.Score > (*h)[0].Score:
			(*h)[0] = p    // evict the weakest...
			heap.Fix(h, 0) // ...and sift the newcomer into place
		}
	}
	top := []Player(*h)
	slices.SortFunc(top, func(a, b Player) int { return cmp.Compare(b.Score, a.Score) })
	return top
}

func main() {
	r := rand.New(rand.NewPCG(1, 2)) // fixed seed: same "random" scores every run
	players := make([]Player, 1_000_000)
	for i := range players {
		players[i] = Player{Name: fmt.Sprintf("player%d", i), Score: r.IntN(1_000_000)}
	}
	players[123_456].Score = 5_000_000 // a legend

	for _, p := range TopK(players, 3) {
		fmt.Println(p.Name, p.Score)
	}
	fmt.Println(len(TopK(players[:2], 3)))
}
```

Output:

```
player123456 5000000
player736903 999998
player61481 999998
2
```

The legend comes first, followed by the two best random scores. The last line shows
that asking for the top 3 of only two players just returns both.

Some details:

- `(*h)[0] = p` followed by `heap.Fix(h, 0)` does a "replace the root" in one sift-down.
  That's cheaper than `heap.Pop` then `heap.Push`, which would do two sifts.
- The heap's internal order isn't sorted, so `TopK` sorts the k winners at the end.
  Sorting k items is O(k log k), which is tiny.
- `rand.New(rand.NewPCG(1, 2))` from `math/rand/v2` gives a seeded generator, so the
  "random" scores are the same every run. Handy for tests and benchmarks.

## Why it's faster

| Approach | Time | Extra memory |
|---|---|---|
| Sort everything | O(n log n) | O(n) if you copy first |
| Max-heap of everything, pop k | O(n + k log n) | O(n) |
| Min-heap of size k | O(n log k) | O(k) |

With n = 1,000,000 and k = 3, log k is under 2. And in practice most players lose to
the root on the first comparison and cost almost nothing. The size-k heap also works on
a **stream**: scores arriving from game servers one at a time, too many to keep in
memory. You never need more than k of them at once.

## Where else this shows up

- "Top 10 most-searched usernames" (count with a map, then top-k over the counts).
- Ranked autocomplete from the tries chapter: top-k over all the matching names.
- Merging k sorted lists (keep one candidate from each list in a heap).
- Dijkstra's shortest-path algorithm, which you'll meet at the end of the graphs
  chapters, is built around a priority queue.
