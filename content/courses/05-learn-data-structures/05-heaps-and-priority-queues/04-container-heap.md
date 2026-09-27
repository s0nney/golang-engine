---
title: container/heap
quiz:
  - question: |
      You've implemented `heap.Interface` on `type Queue []*Player`. What's wrong with
      this code?

      ```go
      q := &Queue{}
      q.Push(&Player{Name: "mira", Score: 50})
      q.Push(&Player{Name: "kai", Score: 90})
      top := heap.Pop(q).(*Player)
      ```
    options:
      - text: Nothing, it works
      - text: 'It calls the method `q.Push`, which only appends, instead of `heap.Push(q, ...)`, so the items are never sifted into heap order'
        correct: true
      - text: '`heap.Pop` returns a `*Player`, so the type assertion won''t compile'
      - text: '`Queue` must be a struct, not a slice'
    explanation: |
      Your `Push` and `Pop` methods are low-level hooks for the `heap` package to call.
      Callers must use the package functions `heap.Push(q, x)` and `heap.Pop(q)`, which
      call your methods *and* restore the heap property. Here the slice is just
      `[mira, kai]` in insertion order, and `heap.Pop` trusts that index 0 is the
      top, so it returns mira (50) instead of kai (90).
  - question: Out of the box, is a `container/heap` a min-heap or a max-heap?
    options:
      - text: A min-heap, ordered by your `Less`; flip `Less` to get a max-heap
        correct: true
      - text: A max-heap
      - text: Whichever you pass to `heap.Init`
      - text: Neither; it's a sorted slice
    explanation: |
      The package always puts the element for which `Less` says "smallest" at index 0.
      To make the *highest* score come out first, write `Less` as `q[i].Score > q[j].Score`.
  - question: A player's score changes while they're in the heap. What's the right way to update it?
    options:
      - text: Pop everything, change the score, and push everything back
      - text: Change the score, then call `heap.Fix(q, p.index)`
        correct: true
      - text: Call `heap.Init` after every change
      - text: Just change the score; the heap reorders itself automatically
    explanation: |
      `heap.Fix` re-establishes order after one element's priority changes, sifting it
      up or down in O(log n). It needs the element's current index, which is why the
      `Swap` method keeps each player's `index` field up to date.
---

Writing sift-up and sift-down yourself is great for understanding. In production Go,
reach for the standard library's `container/heap`. It works on any type that
implements `heap.Interface`:

```go
type Interface interface {
	sort.Interface // Len() int, Less(i, j int) bool, Swap(i, j int)
	Push(x any)    // add x as element Len()
	Pop() any      // remove and return element Len() - 1
}
```

You write five small methods on your own slice type. The package provides the
algorithms: `heap.Init`, `heap.Push`, `heap.Pop`, `heap.Fix` and `heap.Remove`.

## A live leaderboard

Let's keep the current top player available while scores change during a match.

```go
package main

import (
	"container/heap"
	"fmt"
)

type Player struct {
	Name  string
	Score int
	index int // position in the heap, maintained by Swap
}

type Leaderboard []*Player

func (lb Leaderboard) Len() int { return len(lb) }

// Less puts the higher score first, making this a max-heap.
func (lb Leaderboard) Less(i, j int) bool { return lb[i].Score > lb[j].Score }

func (lb Leaderboard) Swap(i, j int) {
	lb[i], lb[j] = lb[j], lb[i]
	lb[i].index = i
	lb[j].index = j
}

func (lb *Leaderboard) Push(x any) {
	p := x.(*Player)
	p.index = len(*lb)
	*lb = append(*lb, p)
}

func (lb *Leaderboard) Pop() any {
	old := *lb
	n := len(old)
	p := old[n-1]
	old[n-1] = nil // let the GC reclaim it later
	p.index = -1   // no longer in the heap
	*lb = old[:n-1]
	return p
}

func main() {
	mira := &Player{Name: "mira", Score: 50}
	lb := &Leaderboard{
		mira,
		{Name: "kai", Score: 90},
		{Name: "bo", Score: 70},
	}
	for i, p := range *lb {
		p.index = i
	}
	heap.Init(lb) // O(n) heapify
	fmt.Println("leader:", (*lb)[0].Name)

	heap.Push(lb, &Player{Name: "zed", Score: 95})
	fmt.Println("leader:", (*lb)[0].Name)

	mira.Score = 120 // mira scores big
	heap.Fix(lb, mira.index)
	fmt.Println("leader:", (*lb)[0].Name)

	for lb.Len() > 0 {
		p := heap.Pop(lb).(*Player)
		fmt.Print(p.Name, "=", p.Score, " ")
	}
	fmt.Println()
}
```

Output:

```
leader: kai
leader: zed
leader: mira
mira=120 zed=95 kai=90 bo=70 
```

## The rules of the road

- **Call the package functions, not your methods.** `heap.Push(lb, p)` appends via
  your `Push` and then sifts up. Calling `lb.Push(p)` directly only appends, and the
  heap is silently broken. Same for `heap.Pop` vs `lb.Pop`.
- **`Less` decides the order.** The package always keeps the "least" element at index
  0, so it's a min-heap by default. We made a max-heap by writing `>`.
- **`Push` and `Pop` take pointer receivers** because they change the slice's length.
  `Len`, `Less` and `Swap` work fine on a value receiver.
- **`Pop` returns `any`**, so you need a type assertion: `heap.Pop(lb).(*Player)`. The
  package predates generics, which is why it uses `any`.
- **`heap.Fix` needs an index.** Tracking `index` inside `Swap` lets you update any
  player's priority in O(log n). Without it, you'd have to search the slice for them first.

## Further reading

- The priority queue example in `go doc container/heap`, and its source in
  `example_pq_test.go`, which this leaderboard follows closely.
