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
exercise:
  starter: |
    package main

    import "fmt"

    type Player struct {
    	Name  string
    	Score int
    }

    // better reports whether a ranks above b: higher score first, and for
    // equal scores, the alphabetically earlier name.
    func better(a, b Player) bool {
    	if a.Score != b.Score {
    		return a.Score > b.Score
    	}
    	return a.Name < b.Name
    }

    // TopK returns the k best players, best first, using a heap that never
    // holds more than k players. It must not modify players.
    func TopK(players []Player, k int) []Player {
    	// ?
    	return nil
    }

    func main() {
    	players := []Player{{"kai", 40}, {"mira", 95}, {"bo", 95}, {"ada", 12}, {"zed", 70}}
    	fmt.Println(TopK(players, 3))                                          // want: [{bo 95} {mira 95} {zed 70}]
    	fmt.Println(players[0], len(TopK(players, 10)), len(TopK(players, 0))) // want: {kai 40} 5 0
    }
  solution: |
    package main

    import (
    	"cmp"
    	"container/heap"
    	"fmt"
    	"slices"
    )

    type Player struct {
    	Name  string
    	Score int
    }

    // better reports whether a ranks above b: higher score first, and for
    // equal scores, the alphabetically earlier name.
    func better(a, b Player) bool {
    	if a.Score != b.Score {
    		return a.Score > b.Score
    	}
    	return a.Name < b.Name
    }

    // worstFirst is a min-heap: the WEAKEST of the current top k sits at index 0.
    type worstFirst []Player

    func (h worstFirst) Len() int           { return len(h) }
    func (h worstFirst) Less(i, j int) bool { return better(h[j], h[i]) }
    func (h worstFirst) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
    func (h *worstFirst) Push(x any)        { *h = append(*h, x.(Player)) }
    func (h *worstFirst) Pop() any {
    	old := *h
    	p := old[len(old)-1]
    	*h = old[:len(old)-1]
    	return p
    }

    // TopK returns the k best players, best first, using a heap that never
    // holds more than k players. It must not modify players.
    func TopK(players []Player, k int) []Player {
    	if k <= 0 {
    		return []Player{}
    	}
    	h := &worstFirst{}
    	for _, p := range players {
    		switch {
    		case h.Len() < k:
    			heap.Push(h, p)
    		case better(p, (*h)[0]):
    			(*h)[0] = p
    			heap.Fix(h, 0)
    		}
    	}
    	top := []Player(*h)
    	slices.SortFunc(top, func(a, b Player) int {
    		return cmp.Or(cmp.Compare(b.Score, a.Score), cmp.Compare(a.Name, b.Name))
    	})
    	return top
    }

    func main() {
    	players := []Player{{"kai", 40}, {"mira", 95}, {"bo", 95}, {"ada", 12}, {"zed", 70}}
    	fmt.Println(TopK(players, 3))
    	fmt.Println(players[0], len(TopK(players, 10)), len(TopK(players, 0)))
    }
  tests: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"math/rand/v2"
    	"slices"
    	"testing"
    )

    func sortedTop(players []Player, k int) []Player {
    	all := slices.Clone(players)
    	slices.SortFunc(all, func(a, b Player) int {
    		return cmp.Or(cmp.Compare(b.Score, a.Score), cmp.Compare(a.Name, b.Name))
    	})
    	return all[:max(0, min(k, len(all)))]
    }

    func TestTopKSmall(t *testing.T) {
    	players := []Player{{"kai", 40}, {"mira", 95}, {"bo", 95}, {"ada", 12}, {"zed", 70}}
    	orig := slices.Clone(players)
    	for _, k := range []int{1, 2, 3, 5, 10, 0, -1} {
    		got := TopK(players, k)
    		want := sortedTop(orig, k)
    		if !slices.Equal(got, want) {
    			t.Errorf("TopK(%v, %d) = %v, want %v", orig, k, got, want)
    		}
    		if !slices.Equal(players, orig) {
    			t.Fatalf("TopK(players, %d) modified players: now %v, was %v", k, players, orig)
    		}
    	}
    	if got := TopK(nil, 3); len(got) != 0 {
    		t.Errorf("TopK(nil, 3) = %v, want empty", got)
    	}
    }

    func TestTopKTies(t *testing.T) {
    	players := []Player{{"eve", 50}, {"dan", 50}, {"cat", 50}, {"bob", 50}, {"amy", 50}}
    	want := []Player{{"amy", 50}, {"bob", 50}}
    	if got := TopK(players, 2); !slices.Equal(got, want) {
    		t.Errorf("TopK(%v, 2) = %v, want %v (equal scores: earlier name wins)", players, got, want)
    	}
    }

    func TestTopKLarge(t *testing.T) {
    	r := rand.New(rand.NewPCG(7, 11))
    	players := make([]Player, 200_000)
    	for i := range players {
    		players[i] = Player{fmt.Sprintf("p%06d", i), r.IntN(50_000)}
    	}
    	orig := slices.Clone(players)
    	for _, k := range []int{1, 10, 250} {
    		got := TopK(players, k)
    		if want := sortedTop(orig, k); !slices.Equal(got, want) {
    			t.Errorf("TopK over 200000 random players, k=%d: got %d players starting %v, want starting %v", k, len(got), got[:min(3, len(got))], want[:3])
    		}
    	}
    	if !slices.Equal(players, orig) {
    		t.Error("TopK modified its input slice")
    	}
    }
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

## Your turn: fair tie-breaks

The prize committee found a problem with `TopK`: when two players tie, which
one wins depends on the heap's internal order. Write your own `TopK` where the
ranking is **higher score first, then alphabetical name**, as given by the
`better` function.

Keep the heap to at most `k` players, with the **weakest** of them at the root.
You can use `container/heap` like the lesson (its `Less(i, j)` should report
whether `h[i]` is *worse* than `h[j]`, which is `better(h[j], h[i])`) or your
own heap. Then sort the winners best first.

`k` of 0 or less gives an empty result, a `k` bigger than the number of players
returns everyone, and `players` must not be modified: the tests check.

## Where else this shows up

- "Top 10 most-searched usernames" (count with a map, then top-k over the counts).
- Ranked autocomplete from the tries chapter: top-k over all the matching names.
- Merging k sorted lists (keep one candidate from each list in a heap).
- Dijkstra's shortest-path algorithm, which you'll meet at the end of the graphs
  chapters, is built around a priority queue.
