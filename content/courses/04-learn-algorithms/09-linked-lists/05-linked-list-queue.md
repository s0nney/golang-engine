---
title: A Linked-List Queue
quiz:
  - question: In a linked-list queue, which ends do `Enqueue` and `Dequeue` use?
    options:
      - text: Enqueue at the head, dequeue at the tail
      - text: Enqueue at the tail, dequeue at the head
        correct: true
      - text: Both at the head
    explanation: |
      In a singly linked list, adding at the tail and removing at the head are
      both O(1). Removing at the *tail* would be O(n), because we'd need the
      predecessor, so the queue's front must be the list's head.
  - question: Both the ring buffer and the linked-list queue have O(1) dequeue. Why is the ring buffer usually faster in practice?
    options:
      - text: The ring buffer has a better Big O for enqueue
      - text: Its items are contiguous and it allocates rarely, while the linked list allocates a node per item and chases pointers
        correct: true
      - text: Linked lists can't hold generic types
    explanation: |
      Big O ignores constants, and the constants differ a lot here. Each linked
      node is a separate heap allocation that the garbage collector has to
      track, and visiting scattered nodes causes cache misses.
---

In the queues chapter, the naive slice queue had an O(n) dequeue, and we fixed
it with a ring buffer. A linked list offers another fix, almost for free: it
already has O(1) `PushBack` and O(1) `PopFront`. That's exactly a queue.

## Which end is which?

In a singly linked list:

| operation | cost |
|-----------|------|
| add at head | O(1) |
| add at tail | O(1), thanks to the tail pointer |
| remove at head | O(1) |
| remove at tail | O(n), needs the predecessor |

A queue needs one end to add at and the *other* end to remove from. The only
combination with both operations O(1) is: **enqueue at the tail, dequeue at the
head**. The head is the front of the line.

## The code

We could embed a `LinkedList[T]`, but a queue is small enough to write
directly, and it shows how few moving parts it needs:

```go
package main

import (
	"fmt"
	"iter"
)

type node[T any] struct {
	value T
	next  *node[T]
}

type Queue[T any] struct {
	head, tail *node[T]
	size       int
}

func (q *Queue[T]) Enqueue(v T) {
	n := &node[T]{value: v}
	if q.tail == nil {
		q.head = n
	} else {
		q.tail.next = n
	}
	q.tail = n
	q.size++
}

func (q *Queue[T]) Dequeue() (T, bool) {
	if q.head == nil {
		var zero T
		return zero, false
	}
	n := q.head
	q.head = n.next
	if q.head == nil {
		q.tail = nil
	}
	q.size--
	return n.value, true
}

func (q *Queue[T]) Peek() (T, bool) {
	if q.head == nil {
		var zero T
		return zero, false
	}
	return q.head.value, true
}

func (q *Queue[T]) Len() int { return q.size }

// Drain yields and removes items, front first, until the queue is empty
// or the caller stops.
func (q *Queue[T]) Drain() iter.Seq[T] {
	return func(yield func(T) bool) {
		for q.size > 0 {
			v, _ := q.Dequeue()
			if !yield(v) {
				return
			}
		}
	}
}

func main() {
	var refresh Queue[string]
	refresh.Enqueue("ava")
	refresh.Enqueue("bo")
	next, _ := refresh.Peek()
	fmt.Println("next up:", next)

	first, _ := refresh.Dequeue()
	fmt.Println("refreshed", first)

	refresh.Enqueue("cy")
	refresh.Enqueue("dee")
	for h := range refresh.Drain() {
		fmt.Println("refreshed", h)
	}
	fmt.Println("left:", refresh.Len())
}
```

```
next up: ava
refreshed ava
refreshed bo
refreshed cy
refreshed dee
left: 0
```

Notice the lowercase `node` type and fields: callers of `Queue` never need to
know it's a linked list inside, so the node stays unexported. And the
`Drain` iterator, unlike `All` from the last lesson, *consumes* the queue as
it goes. Iterators can do that, as long as the name makes it obvious.

## Comparing our three queues

| | naive slice | ring buffer | linked list |
|---|---|---|---|
| enqueue | amortized O(1) | amortized O(1) | O(1), always |
| dequeue | O(n) | O(1) | O(1) |
| allocations | occasional (growth) | occasional (growth) | one per item |
| memory per item | the value | the value (plus spare slots) | value + pointer + allocation overhead |
| worst single operation | O(n) | O(n) during a grow | O(1) |

The linked list's selling point is that last row: **no operation is ever
O(n)**. There's never a big copy, because nothing is ever grown. That can
matter for latency-sensitive code.

But in raw throughput the ring buffer usually wins. Allocating a node per item
keeps the garbage collector busy, and hopping between scattered nodes wastes
CPU cache. Big O says they're equal; the benchmark would say otherwise. (Try
it with `testing.B` and `b.Loop()`, as you did in chapter 3.)

The lesson generalises: when two structures share a Big O, memory layout and
allocations decide the winner.
