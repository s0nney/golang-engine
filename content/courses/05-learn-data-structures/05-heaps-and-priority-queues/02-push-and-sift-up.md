---
title: Push and Sift Up
quiz:
  - question: |
      You push `92` onto the max-heap `[95, 80, 90, 40, 75, 60]`. What's the slice afterwards?
    options:
      - text: '`[95, 80, 90, 40, 75, 60, 92]`'
      - text: '`[95, 80, 92, 40, 75, 60, 90]`'
        correct: true
      - text: '`[92, 80, 95, 40, 75, 60, 90]`'
      - text: '`[95, 92, 90, 80, 75, 60, 40]`'
    explanation: |
      92 is appended at index 6. Its parent is index (6 − 1) / 2 = 2, holding 90.
      92 > 90, so they swap and 92 moves to index 2. Its new parent is index 0, holding
      95. 92 < 95, so it stops.
  - question: Why is `Push` O(log n)?
    options:
      - text: '`append` is O(log n)'
      - text: The new item climbs at most one level per swap, and a complete tree with n nodes has about log₂ n levels
        correct: true
      - text: It does a binary search to find the right position
      - text: It re-sorts half the slice
    explanation: |
      Because a heap is always a *complete* tree, its height is ⌊log₂ n⌋. Sift-up
      moves up one level per step, so it does at most that many swaps. (`append` is
      O(1) amortized.)
  - question: |
      With `h := NewHeap(func(a, b int) bool { return a < b })`, which item does
      `Peek` return after pushing 30, 10 and 20?
    options:
      - text: '30'
      - text: '20'
      - text: '10'
        correct: true
      - text: The first item pushed, 30
    explanation: |
      `before(a, b)` returns true when `a` should come out first. With `a < b`,
      smaller values rise to the top, so this is a **min**-heap and the root is 10.
---

Adding an item to a heap has to keep two promises: the tree stays **complete**, and
every parent still beats its children.

## The algorithm

1. **Append** the new item to the end of the slice. That's the next free spot in the
   bottom level, so the tree stays complete.
2. **Sift up**: while the item beats its parent, swap them. Stop when it doesn't, or
   when it reaches the root.

Pushing 92 into a max-heap:

```
          95                   95                    95
        /    \               /    \                /    \
      80      90    ->     80      90     ->     80      92
     /  \    /  \         /  \    /  \          /  \    /  \
   40   75  60  [92]    40   75  60   92      40   75  60   90

   append at the end     92 > 90: swap          92 < 95: stop
```

Only the path from the new leaf to the root is touched. Everything else was already a
valid heap and stays that way.

## A generic heap

The same code should work as a max-heap of scores, a min-heap of wait times, or a heap
of `Player` structs. So instead of requiring `cmp.Ordered`, we take a function that
says which of two items should come out **first**:

```go
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

// Peek returns the top item without removing it.
func (h *Heap[T]) Peek() (T, bool) {
	if len(h.items) == 0 {
		var zero T
		return zero, false
	}
	return h.items[0], true
}

func (h *Heap[T]) Push(v T) {
	h.items = append(h.items, v)
	i := len(h.items) - 1
	for i > 0 {
		parent := (i - 1) / 2
		if !h.before(h.items[i], h.items[parent]) {
			break // parent already comes first: heap property holds
		}
		h.items[i], h.items[parent] = h.items[parent], h.items[i]
		i = parent
	}
}

func main() {
	scores := NewHeap(func(a, b int) bool { return a > b }) // max-heap
	for _, s := range []int{40, 75, 90, 95, 80, 60} {
		scores.Push(s)
		top, _ := scores.Peek()
		fmt.Println("pushed", s, "top", top, scores.items)
	}
}
```

Output:

```
pushed 40 top 40 [40]
pushed 75 top 75 [75 40]
pushed 90 top 90 [90 40 75]
pushed 95 top 95 [95 90 75 40]
pushed 80 top 95 [95 90 75 40 80]
pushed 60 top 95 [95 90 75 40 80 60]
```

Notes on the Go:

- `NewHeap(func(a, b int) bool { return a > b })` doesn't spell out `[int]`. Go
  infers `T` from the function argument's type.
- Flip `>` to `<` and the same type becomes a min-heap. For other types you can use
  `cmp.Compare`, or compare a field: `func(a, b Player) bool { return a.Score > b.Score }`.
- `h.items[i], h.items[parent] = h.items[parent], h.items[i]` swaps two elements in
  one statement. Go evaluates the right-hand side fully before assigning.
- `Peek` returns `(T, bool)` because an empty heap has no top. Same "comma ok" idea
  as `BST.Min`.

## Cost

A complete tree with n nodes has height ⌊log₂ n⌋, and sift-up climbs at most one level
per swap. So `Push` is **O(log n)**, and in practice often less: a random new item
usually belongs near the bottom, where most of the nodes are.
