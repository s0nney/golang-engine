---
title: Pop and Sift Down
quiz:
  - question: |
      You pop from the max-heap `[95, 80, 90, 40, 75, 60]`. What's the slice afterwards?
    options:
      - text: '`[90, 80, 60, 40, 75]`'
        correct: true
      - text: '`[80, 90, 40, 75, 60]`'
      - text: '`[90, 80, 40, 75, 60]`'
      - text: '`[60, 80, 90, 40, 75]`'
    explanation: |
      Move the last item (60) to the root and shrink: `[60, 80, 90, 40, 75]`. 60's
      children are 80 and 90; the bigger is 90, and 60 < 90, so swap:
      `[90, 80, 60, 40, 75]`. 60 is now at index 2, whose children would be at
      indexes 5 and 6, which don't exist. Done.
  - question: During sift-down in a max-heap, why swap with the **larger** child rather than either one?
    options:
      - text: It's faster to compare with the larger one
      - text: The item that moves up becomes the parent of the other child, so it has to be the larger of the two
        correct: true
      - text: The larger child is always on the left
      - text: It doesn't matter; either child works
    explanation: |
      If you swapped with the smaller child, it would become the parent of the larger
      one, breaking the heap property right there. Promoting the larger child keeps it
      above its sibling.
exercise:
  starter: |
    package main

    import "fmt"

    type Heap[T any] struct {
    	items  []T
    	before func(a, b T) bool // true if a should come out before b
    }

    func NewHeap[T any](before func(a, b T) bool) *Heap[T] {
    	return &Heap[T]{before: before}
    }

    func (h *Heap[T]) Len() int { return len(h.items) }

    // Push adds v, then sifts it up until its parent comes before it.
    func (h *Heap[T]) Push(v T) {
    	h.items = append(h.items, v)
    	// ?
    }

    // Pop removes and returns the item that comes first, or false if empty.
    // Move the last item to the root, shrink the slice, then sift down.
    func (h *Heap[T]) Pop() (T, bool) {
    	var zero T
    	// ?
    	return zero, false
    }

    type Waiting struct {
    	Name     string
    	JoinedAt int // seconds since the server started
    }

    func main() {
    	// Matchmaking: whoever joined the queue earliest gets matched first.
    	queue := NewHeap(func(a, b Waiting) bool { return a.JoinedAt < b.JoinedAt })
    	queue.Push(Waiting{"mira", 30})
    	queue.Push(Waiting{"kai", 10})
    	queue.Push(Waiting{"bo", 20})
    	fmt.Println("waiting:", queue.Len())
    	for {
    		p, ok := queue.Pop()
    		if !ok {
    			break
    		}
    		fmt.Println("matching", p.Name) // want kai, then bo, then mira
    	}
    	fmt.Println("still waiting:", queue.Len())
    }
  solution: |
    package main

    import "fmt"

    type Heap[T any] struct {
    	items  []T
    	before func(a, b T) bool // true if a should come out before b
    }

    func NewHeap[T any](before func(a, b T) bool) *Heap[T] {
    	return &Heap[T]{before: before}
    }

    func (h *Heap[T]) Len() int { return len(h.items) }

    func (h *Heap[T]) Push(v T) {
    	h.items = append(h.items, v)
    	i := len(h.items) - 1
    	for i > 0 {
    		parent := (i - 1) / 2
    		if !h.before(h.items[i], h.items[parent]) {
    			break
    		}
    		h.items[i], h.items[parent] = h.items[parent], h.items[i]
    		i = parent
    	}
    }

    func (h *Heap[T]) Pop() (T, bool) {
    	var zero T
    	if len(h.items) == 0 {
    		return zero, false
    	}
    	top := h.items[0]
    	last := len(h.items) - 1
    	h.items[0] = h.items[last]
    	h.items[last] = zero
    	h.items = h.items[:last]
    	i := 0
    	for {
    		best := i
    		for _, c := range []int{2*i + 1, 2*i + 2} {
    			if c < len(h.items) && h.before(h.items[c], h.items[best]) {
    				best = c
    			}
    		}
    		if best == i {
    			break
    		}
    		h.items[i], h.items[best] = h.items[best], h.items[i]
    		i = best
    	}
    	return top, true
    }

    type Waiting struct {
    	Name     string
    	JoinedAt int // seconds since the server started
    }

    func main() {
    	// Matchmaking: whoever joined the queue earliest gets matched first.
    	queue := NewHeap(func(a, b Waiting) bool { return a.JoinedAt < b.JoinedAt })
    	queue.Push(Waiting{"mira", 30})
    	queue.Push(Waiting{"kai", 10})
    	queue.Push(Waiting{"bo", 20})
    	fmt.Println("waiting:", queue.Len())
    	for {
    		p, ok := queue.Pop()
    		if !ok {
    			break
    		}
    		fmt.Println("matching", p.Name)
    	}
    	fmt.Println("still waiting:", queue.Len())
    }
  tests: |
    package main

    import (
    	"math/rand/v2"
    	"slices"
    	"testing"
    )

    func less(a, b int) bool { return a < b }

    func TestPopEmpty(t *testing.T) {
    	h := NewHeap(less)
    	if v, ok := h.Pop(); ok {
    		t.Errorf("Pop() on an empty heap = %d, true; want 0, false", v)
    	}
    }

    func TestPushKeepsMinOnTop(t *testing.T) {
    	h := NewHeap(less)
    	for _, v := range []int{30, 10, 20, 5, 40} {
    		h.Push(v)
    		if !slices.Contains(h.items, v) {
    			t.Fatalf("after Push(%d), items = %v; the value is missing", v, h.items)
    		}
    	}
    	if h.items[0] != 5 {
    		t.Errorf("after pushing 30, 10, 20, 5, 40 the root is %d, want 5 (items = %v)", h.items[0], h.items)
    	}
    	for i := 1; i < len(h.items); i++ {
    		if h.items[(i-1)/2] > h.items[i] {
    			t.Errorf("heap property broken: parent %d at index %d is bigger than child %d at index %d (items = %v)",
    				h.items[(i-1)/2], (i-1)/2, h.items[i], i, h.items)
    		}
    	}
    }

    func TestPopReturnsSortedOrder(t *testing.T) {
    	r := rand.New(rand.NewPCG(7, 7))
    	var want []int
    	h := NewHeap(less)
    	for range 200 {
    		v := r.IntN(1000)
    		want = append(want, v)
    		h.Push(v)
    	}
    	slices.Sort(want)
    	var got []int
    	for range 200 {
    		v, ok := h.Pop()
    		if !ok {
    			t.Fatalf("Pop() returned false while Len() = %d", h.Len())
    		}
    		got = append(got, v)
    	}
    	if h.Len() != 0 {
    		t.Errorf("after popping all 200 items, Len() = %d, want 0", h.Len())
    	}
    	if !slices.Equal(got, want) {
    		t.Errorf("popping everything did not give ascending order.\n got: %v\nwant: %v", got, want)
    	}
    }

    func TestMaxHeapOfPlayers(t *testing.T) {
    	type player struct {
    		name  string
    		score int
    	}
    	h := NewHeap(func(a, b player) bool { return a.score > b.score })
    	for _, p := range []player{{"mira", 50}, {"kai", 90}, {"bo", 70}, {"zed", 95}} {
    		h.Push(p)
    	}
    	var order []string
    	for range 4 {
    		p, _ := h.Pop()
    		order = append(order, p.name)
    	}
    	if want := []string{"zed", "kai", "bo", "mira"}; !slices.Equal(order, want) {
    		t.Errorf("max-heap popped %v, want %v", order, want)
    	}
    }
---

Popping removes the top item, the one that should come out first. The tricky part
is filling the hole it leaves at the root without breaking the heap.

## The algorithm

1. Save the root; it's the answer.
2. Move the **last** item into the root's place and shrink the slice by one. The tree
   stays complete, but the root is now probably in the wrong place.
3. **Sift down**: while the item has a child that should come before it, swap it with
   the **best** child (the larger one, in a max-heap). Stop when neither child beats it,
   or when it reaches the bottom.

```
          95                  60                  90
        /    \              /    \              /    \
      80      90    ->    80      90    ->    80      60
     /  \    /           /  \                /  \
   40   75  60         40   75             40   75

   remove 95           60 moved to root    60 < 90 (the bigger child): swap, done
```

Why the *best* child? Whichever child moves up becomes the parent of the other one, so
it has to beat its sibling.

## The code

Continuing the `Heap[T]` from last lesson:

```go
package main

import "fmt"

type Heap[T any] struct {
	items  []T
	before func(a, b T) bool
}

func NewHeap[T any](before func(a, b T) bool) *Heap[T] {
	return &Heap[T]{before: before}
}

func (h *Heap[T]) Len() int { return len(h.items) }

func (h *Heap[T]) Push(v T) {
	h.items = append(h.items, v)
	i := len(h.items) - 1
	for i > 0 {
		parent := (i - 1) / 2
		if !h.before(h.items[i], h.items[parent]) {
			break
		}
		h.items[i], h.items[parent] = h.items[parent], h.items[i]
		i = parent
	}
}

func (h *Heap[T]) Pop() (T, bool) {
	var zero T
	if len(h.items) == 0 {
		return zero, false
	}
	top := h.items[0]
	last := len(h.items) - 1
	h.items[0] = h.items[last]
	h.items[last] = zero // don't keep a stale reference alive
	h.items = h.items[:last]

	i := 0
	for {
		best := i
		for _, c := range []int{2*i + 1, 2*i + 2} {
			if c < len(h.items) && h.before(h.items[c], h.items[best]) {
				best = c
			}
		}
		if best == i {
			break // no child beats it: heap property holds
		}
		h.items[i], h.items[best] = h.items[best], h.items[i]
		i = best
	}
	return top, true
}

func main() {
	scores := NewHeap(func(a, b int) bool { return a > b })
	for _, s := range []int{95, 80, 90, 40, 75, 60} {
		scores.Push(s)
	}
	scores.Pop()
	fmt.Println(scores.items)

	for scores.Len() > 0 {
		s, _ := scores.Pop()
		fmt.Print(s, " ")
	}
	fmt.Println()
}
```

Output:

```
[90 80 60 40 75]
90 80 75 60 40 
```

Things to notice:

- The inner loop finds `best` among the node and its (up to two) children. If the node
  itself is still best, we're done. This handles the "only a left child" and "no
  children" cases without special code.
- `h.items[last] = zero` clears the slot we're about to slice off. The backing array
  still holds it, and if `T` were a pointer or a struct containing one, the garbage
  collector couldn't free that object. It's a habit worth having whenever you shrink a
  slice of pointers.
- Popping until empty gives the scores **in order**. That's **heap sort**: O(n log n),
  and it can even be done in place inside the original slice.

## Building a heap in O(n)

Pushing n items one at a time costs O(n log n). If you already have all the items,
there's a faster way called **heapify**: put them in a slice as-is, then sift *down*
every non-leaf node, starting from the last one (index `n/2 - 1`) back to the root.
Most nodes are near the bottom and barely move, so the total work adds up to O(n).
The standard library's `heap.Init` does exactly this.

## Cost summary

| Operation | Cost |
|---|---|
| `Peek` | O(1) |
| `Push` | O(log n) |
| `Pop` | O(log n) |
| Heapify n items | O(n) |

## Your turn

Matchmaking pairs whoever has waited longest, so the exercise uses the generic `Heap`
as a min-heap on each player's join time. Complete `Push` (sift up after the `append`)
and `Pop` (move the last item to the root, shrink, then sift down towards the child that
should come first). `Pop` on an empty heap returns the zero value and `false`. The tests
also use your heap as a max-heap of scores, so always compare with `h.before`, never
with `<`.
