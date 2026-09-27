---
title: A Naive Slice Queue
quiz:
  - question: |
      Why is this `Dequeue` O(n)?

      ```go
      func (q *Queue[T]) Dequeue() (T, bool) {
      	var zero T
      	if len(q.items) == 0 {
      		return zero, false
      	}
      	v := q.items[0]
      	q.items = slices.Delete(q.items, 0, 1)
      	return v, true
      }
      ```
    options:
      - text: '`slices.Delete` sorts the slice'
      - text: Deleting index 0 shifts every remaining element one place to the left
        correct: true
      - text: Reading `q.items[0]` scans the slice
    explanation: |
      Slices are contiguous, so removing the first element means moving all the
      others down to fill the gap. That's `n - 1` copies for a queue of `n`
      items.
  - question: 'Dequeuing with `q.items = q.items[1:]` is O(1). What''s the catch?'
    options:
      - text: It panics once the slice has been re-sliced 100 times
      - text: The consumed slots at the front are never reused, so memory is only reclaimed when append eventually copies the live items to a new array
        correct: true
      - text: It changes the order of the items
    explanation: |
      Re-slicing moves the start pointer forward, but the backing array still
      holds the dead front slots, and new items keep being appended at the back.
      The space ahead of the pointer is wasted until a reallocation happens.
---

Let's build the obvious queue first: a slice, where the **front of the queue
is index 0** and new items are appended at the end.

## Version 1: shift everything

```go
package main

import (
	"fmt"
	"slices"
)

type Queue[T any] struct {
	items []T
}

func (q *Queue[T]) Enqueue(v T) {
	q.items = append(q.items, v)
}

func (q *Queue[T]) Dequeue() (T, bool) {
	var zero T
	if len(q.items) == 0 {
		return zero, false
	}
	v := q.items[0]
	q.items = slices.Delete(q.items, 0, 1)
	return v, true
}

func (q *Queue[T]) Len() int { return len(q.items) }

func main() {
	var jobs Queue[string]
	jobs.Enqueue("refresh:ava")
	jobs.Enqueue("refresh:bo")
	jobs.Enqueue("refresh:cy")

	for jobs.Len() > 0 {
		j, _ := jobs.Dequeue()
		fmt.Println("processing", j)
	}
}
```

```
processing refresh:ava
processing refresh:bo
processing refresh:cy
```

It's correct and it's FIFO. `Enqueue` is amortized O(1), just like the stack's
`Push`. But look at what `slices.Delete(q.items, 0, 1)` has to do:

```
before: [ava][bo][cy][dee]
shift:   bo → slot 0, cy → slot 1, dee → slot 2
after:  [bo][cy][dee]
```

Removing the front of a contiguous array means **moving every other element**.
That's **O(n) per dequeue**. Processing a queue of `n` jobs one at a time costs
n + (n-1) + ... + 1, which is O(n²) in total. With a million refresh jobs
queued overnight, that's hundreds of billions of copies. Ouch.

## Version 2: just move the start

A slice header can point into the *middle* of a backing array. So instead of
shifting the elements, shift the slice:

```go
func (q *Queue[T]) Dequeue() (T, bool) {
	var zero T
	if len(q.items) == 0 {
		return zero, false
	}
	v := q.items[0]
	q.items[0] = zero // don't keep a reference to it
	q.items = q.items[1:]
	return v, true
}
```

Now dequeue is **O(1)**, because re-slicing just bumps a pointer and decrements
the length. Many Go programs use exactly this, and for short-lived queues it's
perfectly fine.

But the dead slots at the front are **never reused**:

```
after 3 dequeues:  [ _ ][ _ ][ _ ][dee][eve]
                                   ↑ q.items starts here
```

The slice's capacity only counts from its start to the end of the array, so
the dead slots in front are unreachable. As you keep enqueuing, `append` runs
out of room at the back and allocates a new array, copying only the live items.
So memory does eventually get reclaimed, but a long-running queue churns
through allocations and holds wasted space in the meantime.

## What we actually want

The ideal queue would:

- dequeue in **O(1)** without shifting anything,
- **reuse** the slots that dequeued items leave behind,
- only allocate when it's genuinely full.

The trick is to stop thinking of the array as a line with a start and an end,
and start thinking of it as a **circle**. That's the ring buffer, next.
