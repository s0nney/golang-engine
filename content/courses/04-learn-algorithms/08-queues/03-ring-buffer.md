---
title: A Ring Buffer Queue
quiz:
  - question: |
      A ring buffer has a backing array of length 4, `head = 3` and `size = 2`.
      At which index does the next enqueued item go?
    options:
      - text: '`5`'
      - text: '`1`'
        correct: true
      - text: '`0`'
      - text: '`2`'
    explanation: |
      The next free slot is `(head + size) % len(buf)` = `(3 + 2) % 4` = `1`. The
      two live items sit at indexes 3 and 0, having wrapped around the end.
  - question: What's the cost of `Enqueue` and `Dequeue` in a growable ring buffer?
    options:
      - text: Both O(n)
      - text: Enqueue amortized O(1), dequeue O(1)
        correct: true
      - text: Enqueue O(1), dequeue O(n)
    explanation: |
      Dequeue just reads a slot and moves `head`. Enqueue writes a slot, except
      when the buffer is full and must be copied into one twice the size. Since
      that doubling is rare, enqueue is amortized O(1), just like `append`.
---

A **ring buffer** (also called a *circular buffer*) is a fixed-size array
where the end wraps around to the beginning. It lets a queue reuse the slots
that dequeued items leave behind, so nothing ever has to shift.

## Thinking in circles

Keep two numbers alongside the array:

- `head`: the index of the front item.
- `size`: how many items are in the queue.

The back of the queue is `size` places after `head`, **wrapping around** with
the modulo operator `%`:

```
next free slot = (head + size) % len(buf)
```

Watch a 4-slot buffer handle Clout's refresh jobs:

```
enqueue ava, bo, cy   [ava][bo ][cy ][   ]  head=0 size=3
dequeue → ava         [   ][bo ][cy ][   ]  head=1 size=2
enqueue dee           [   ][bo ][cy ][dee]  head=1 size=3
enqueue eve           [eve][bo ][cy ][dee]  head=1 size=4  ← wrapped to slot 0!
dequeue → bo          [eve][   ][cy ][dee]  head=2 size=3
```

`eve` went into slot 0, which `ava` had vacated. No shifting, no waste. The
queue's logical order is cy, dee, eve, even though eve sits at the lowest
index.

## When it's full

A fixed-size ring buffer can simply refuse new items when `size == len(buf)`.
That's often what you want for bounded buffers (a rate limiter that drops
excess requests, or an audio buffer). For a general-purpose queue, we'll
**grow** instead: allocate an array twice as big and copy the items across *in
queue order*, so `head` starts at 0 again. It's the same doubling trick that
makes `append` amortized O(1).

## In Go

```go
package main

import "fmt"

type Ring[T any] struct {
	buf  []T
	head int
	size int
}

func (r *Ring[T]) Len() int { return r.size }

func (r *Ring[T]) Enqueue(v T) {
	if r.size == len(r.buf) {
		r.grow()
	}
	r.buf[(r.head+r.size)%len(r.buf)] = v
	r.size++
}

func (r *Ring[T]) Dequeue() (T, bool) {
	var zero T
	if r.size == 0 {
		return zero, false
	}
	v := r.buf[r.head]
	r.buf[r.head] = zero
	r.head = (r.head + 1) % len(r.buf)
	r.size--
	return v, true
}

func (r *Ring[T]) grow() {
	newBuf := make([]T, max(4, 2*len(r.buf)))
	for i := range r.size {
		newBuf[i] = r.buf[(r.head+i)%len(r.buf)]
	}
	r.buf = newBuf
	r.head = 0
}

func main() {
	var q Ring[string]
	for _, h := range []string{"ava", "bo", "cy"} {
		q.Enqueue(h)
	}
	first, _ := q.Dequeue()
	fmt.Println("served", first)

	for _, h := range []string{"dee", "eve", "fay"} {
		q.Enqueue(h)
	}
	fmt.Println("len", q.Len(), "cap", len(q.buf))
	for q.Len() > 0 {
		v, _ := q.Dequeue()
		fmt.Print(v, " ")
	}
	fmt.Println()
}
```

```
served ava
len 5 cap 8
bo cy dee eve fay
```

The first four enqueues fit in the initial 4 slots; `eve` wrapped into slot 0;
`fay` found the buffer full, so it grew to 8 and copied everything back into
order.

## Details

- **The zero value works.** `buf` starts nil with length 0, so the first
  `Enqueue` sees `size == len(buf)` (both 0) and grows to 4. That `max(4, ...)`
  (the built-in `max`, Go 1.21+) avoids growing to `2 × 0 = 0`.
- **Don't take `%` of zero.** Every `%` happens only when `len(r.buf) > 0`. Take
  the modulo of an empty buffer and you get a *divide by zero* panic.
- **Zeroing on dequeue** lets the garbage collector reclaim whatever the item
  pointed to, just like in the stack.
- **Complexity**: `Dequeue`, `Len` and peeking are **O(1)**. `Enqueue` is
  **amortized O(1)**. Space is **O(n)**, and freed slots are reused instead of
  leaked.

This is how serious queue implementations work. Go's own buffered channels are
built on a ring buffer, which brings us to the last lesson of this chapter.
