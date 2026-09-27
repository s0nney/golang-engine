---
title: 'Stash: A Generic Heap'
quiz:
  - question: What does `container/heap` require of the type you give it?
    options:
      - text: A `Less(a, b T) bool` method only
      - text: '`heap.Interface`: `Len`, `Less(i, j int)`, `Swap`, plus `Push(x any)` and `Pop() any`'
        correct: true
      - text: That it's a slice of a `cmp.Ordered` type
      - text: Nothing; it works on any slice
    explanation: |
      `heap.Interface` embeds `sort.Interface` (`Len`, `Less`, `Swap`) and adds `Push` and
      `Pop`, which take and return `any` because the package predates generics.
  - question: |
      Why is the adapter type `items[T]` unexported in Stash's `Heap`?
    options:
      - text: Unexported types are faster
      - text: 'So users only ever see the typed `Push(v T)` and `Pop() (T, bool)`; the `any`-based `heap.Interface` methods stay an implementation detail'
        correct: true
      - text: Generic types must be unexported
      - text: '`container/heap` only accepts unexported types'
    explanation: |
      If `Heap` itself implemented `heap.Interface`, its public `Push` would have to take
      an `any`, and users could call `Swap` or break the heap invariant. Wrapping keeps the
      public API typed and safe.
---

You built binary heaps by hand in [Learn Data Structures](/courses/learn-data-structures/heaps-and-priority-queues/container-heap), and met `container/heap` there too. Stash wants a priority queue for things like "expire the oldest cache entry" or "run the most urgent job", so this lesson shows how to put a clean generic face on the old, `any`-based standard library package.

## The problem with container/heap

`container/heap` was written long before generics. It works on anything implementing:

```go
type Interface interface {
	sort.Interface // Len() int; Less(i, j int) bool; Swap(i, j int)
	Push(x any)    // add x as element Len()
	Pop() any      // remove and return element Len() - 1
}
```

So a heap of `Job`s means writing five methods on a `JobHeap` type, and every `heap.Pop(h)` returns an `any` you have to assert back to `Job`. And you do it again for every element type.

## Wrap it once

The generic version writes the adapter **once**, keeps it unexported, and exposes only typed methods:

```go
package main

import (
	"cmp"
	"container/heap"
	"fmt"
)

// items adapts a slice to heap.Interface. It's unexported: users never see any.
type items[T any] struct {
	s    []T
	less func(a, b T) bool
}

func (h *items[T]) Len() int           { return len(h.s) }
func (h *items[T]) Less(i, j int) bool { return h.less(h.s[i], h.s[j]) }
func (h *items[T]) Swap(i, j int)      { h.s[i], h.s[j] = h.s[j], h.s[i] }
func (h *items[T]) Push(x any)         { h.s = append(h.s, x.(T)) }
func (h *items[T]) Pop() any {
	last := h.s[len(h.s)-1]
	var zero T
	h.s[len(h.s)-1] = zero
	h.s = h.s[:len(h.s)-1]
	return last
}

// Heap is a priority queue: Pop always returns the smallest value by less.
type Heap[T any] struct {
	items items[T]
}

func NewHeap[T any](less func(a, b T) bool) *Heap[T] {
	return &Heap[T]{items: items[T]{less: less}}
}

// NewMinHeap orders values by their natural order.
func NewMinHeap[T cmp.Ordered]() *Heap[T] { return NewHeap(cmp.Less[T]) }

func (h *Heap[T]) Push(v T) { heap.Push(&h.items, v) }

func (h *Heap[T]) Pop() (T, bool) {
	if h.items.Len() == 0 {
		var zero T
		return zero, false
	}
	return heap.Pop(&h.items).(T), true
}

func (h *Heap[T]) Len() int { return h.items.Len() }

type Job struct {
	Name     string
	Priority int
}

func main() {
	jobs := NewHeap(func(a, b Job) bool { return a.Priority > b.Priority }) // max-heap
	jobs.Push(Job{"backup", 2})
	jobs.Push(Job{"alert", 9})
	jobs.Push(Job{"report", 5})
	for jobs.Len() > 0 {
		j, _ := jobs.Pop()
		fmt.Println(j.Priority, j.Name)
	}

	sizes := NewMinHeap[int]()
	for _, n := range []int{40, 10, 30} {
		sizes.Push(n)
	}
	smallest, _ := sizes.Pop()
	fmt.Println("smallest:", smallest)
}
```

```
9 alert
5 report
2 backup
smallest: 10
```

What to notice:

- **The `any` never escapes.** `items[T].Push(x any)` asserts `x.(T)`, which can't fail because only `Heap[T].Push(v T)` ever calls it. Users see `Push(v T)` and `Pop() (T, bool)`.
- **Order is a function, not a constraint.** `NewHeap` takes `less func(a, b T) bool`, so it works for structs, max-heaps and multi-key orders. `NewMinHeap[T cmp.Ordered]` is the convenience version, passing the instantiated generic function `cmp.Less[T]` as a value (chapter 3). That's the Func-variant pattern from chapter 4, in reverse.
- **`Pop` zeroes the slot** it removes, so a heap of pointers doesn't keep popped values alive.
- **`Pop` on an empty heap returns `(zero, false)`** instead of panicking the way `heap.Pop` does.

## The cost of any

Every `heap.Push(&h.items, v)` converts `v` to `any`. For pointer-shaped values (pointers, maps, channels) that's free, but for a `Job` struct it usually means allocating a copy on the heap, and `Pop` does the same in reverse. For most programs that's irrelevant. For a hot path, a hand-written generic heap that sifts `[]T` directly (like the one you built in Learn Data Structures) avoids the boxing entirely. Chapter 8 shows how to measure the difference instead of guessing.

## A heap has no iterator

Unlike `Set` and `OrderedMap`, `Heap` doesn't get an `All()` method. Its slice is only *partially* ordered, so iterating it would expose an order that means nothing. The meaningful traversal is destructive: `for h.Len() > 0 { v, _ := h.Pop() }`. When an API would have to explain "the order is weird", it's often better not to offer it.
