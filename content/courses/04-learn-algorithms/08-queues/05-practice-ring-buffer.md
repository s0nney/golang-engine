---
title: 'Practice: Ring Buffer Queue'
exercise:
  starter: |
    package main

    import "fmt"

    // Ring is a FIFO queue backed by a circular buffer.
    // The zero value is an empty queue.
    type Ring[T any] struct {
    	buf  []T
    	head int // index of the front item
    	size int // number of items in the queue
    }

    // Len returns the number of items in the queue.
    func (r *Ring[T]) Len() int { return r.size }

    // Enqueue adds v at the back. If the buffer is full, it grows first.
    func (r *Ring[T]) Enqueue(v T) {
    	if r.size == len(r.buf) {
    		r.grow()
    	}
    	// ? write v into the next free slot, wrapping around with %
    }

    // Dequeue removes and returns the front item, or the zero value and
    // false if the queue is empty.
    func (r *Ring[T]) Dequeue() (T, bool) {
    	var zero T
    	// ? read the front, zero its slot, move head forward (wrapping)
    	return zero, false
    }

    // grow copies the items, in queue order, into a buffer twice as big
    // (at least 4 slots) and resets head to 0.
    func (r *Ring[T]) grow() {
    	// ?
    }

    func main() {
    	var q Ring[string]
    	for _, h := range []string{"ava", "bo", "cy"} {
    		q.Enqueue(h)
    	}
    	first, ok := q.Dequeue()
    	fmt.Println("served", first, ok) // want: served ava true
    	for _, h := range []string{"dee", "eve", "fay"} {
    		q.Enqueue(h)
    	}
    	for q.Len() > 0 {
    		v, _ := q.Dequeue()
    		fmt.Print(v, " ") // want: bo cy dee eve fay
    	}
    	fmt.Println()
    }
  solution: |
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
    	first, ok := q.Dequeue()
    	fmt.Println("served", first, ok)
    	for _, h := range []string{"dee", "eve", "fay"} {
    		q.Enqueue(h)
    	}
    	for q.Len() > 0 {
    		v, _ := q.Dequeue()
    		fmt.Print(v, " ")
    	}
    	fmt.Println()
    }
  tests: |
    package main

    import (
    	"math/rand/v2"
    	"testing"
    )

    func TestRingBasics(t *testing.T) {
    	var q Ring[string]
    	if _, ok := q.Dequeue(); ok {
    		t.Fatal("Dequeue on an empty Ring returned ok = true, want false")
    	}
    	q.Enqueue("ava")
    	q.Enqueue("bo")
    	if q.Len() != 2 {
    		t.Fatalf("after 2 enqueues, Len() = %d, want 2", q.Len())
    	}
    	if v, ok := q.Dequeue(); v != "ava" || !ok {
    		t.Fatalf("Dequeue() = %q, %v; want \"ava\", true (first in, first out)", v, ok)
    	}
    }

    func TestRingReusesSlots(t *testing.T) {
    	var q Ring[int]
    	for i := range 4 {
    		q.Enqueue(i)
    	}
    	capBefore := len(q.buf)
    	for i := range 1000 {
    		v, ok := q.Dequeue()
    		if !ok || v != i {
    			t.Fatalf("Dequeue() = %d, %v; want %d, true", v, ok, i)
    		}
    		q.Enqueue(i + 4)
    	}
    	if len(q.buf) != capBefore {
    		t.Errorf("a queue that never held more than 4 items grew its buffer from %d to %d slots: dequeued slots should be reused", capBefore, len(q.buf))
    	}
    }

    func TestRingZeroesDequeuedSlots(t *testing.T) {
    	var q Ring[*string]
    	s := "ava"
    	q.Enqueue(&s)
    	q.Dequeue()
    	for i, p := range q.buf {
    		if p != nil {
    			t.Errorf("buf[%d] still points at a dequeued item: set the slot to the zero value", i)
    		}
    	}
    }

    func TestRingRandom(t *testing.T) {
    	r := rand.New(rand.NewPCG(8, 9))
    	var q Ring[int]
    	var want []int
    	next := 0
    	for step := range 5000 {
    		if r.IntN(3) > 0 {
    			q.Enqueue(next)
    			want = append(want, next)
    			next++
    		} else {
    			v, ok := q.Dequeue()
    			if len(want) == 0 {
    				if ok {
    					t.Fatalf("step %d: Dequeue on an empty queue returned %d, true", step, v)
    				}
    				continue
    			}
    			if !ok || v != want[0] {
    				t.Fatalf("step %d: Dequeue() = %d, %v; want %d, true", step, v, ok, want[0])
    			}
    			want = want[1:]
    		}
    		if q.Len() != len(want) {
    			t.Fatalf("step %d: Len() = %d, want %d", step, q.Len(), len(want))
    		}
    	}
    }
---

Clout's refresh queue is processing millions of jobs a night, and the naive
slice queue's O(n) dequeue is hurting. Replace it with a **ring buffer**.

## Your task

The `Ring[T]` struct keeps a backing slice `buf`, the index of the front item
`head`, and the number of items `size`. Complete three methods:

- `Enqueue(v)`: the grow-if-full check is already there. Write `v` into the
  next free slot, `(head + size) % len(buf)`, and increment `size`.
- `Dequeue()`: if empty, return the zero value and `false`. Otherwise read
  `buf[head]`, set that slot to the zero value, move `head` forward one place
  **with wrap-around**, decrement `size`, and return the item and `true`.
- `grow()`: allocate a new buffer of `max(4, 2*len(buf))` slots, copy the items
  across **in queue order** (item `i` lives at `(head + i) % len(buf)`), and
  reset `head` to 0.

```
[eve][bo ][cy ][dee]  head=1 size=4   → grow →
[bo ][cy ][dee][eve][   ][   ][   ][   ]  head=0 size=4
```

Careful: `x % 0` panics with *integer divide by zero*. Every `%` in your code
must run only when `len(buf) > 0`.

## How it's graded

The tests check FIFO order, that dequeued slots are cleared, and that a queue
which never holds more than 4 items **doesn't keep growing** (slots must be
reused). Then they run 5,000 random enqueues and dequeues (fixed seed) against
a simple reference queue.
