---
title: 'Practice: Build a Linked List'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"iter"
    	"slices"
    )

    type Node[T any] struct {
    	Value T
    	Next  *Node[T]
    }

    // LinkedList is a singly linked list. The zero value is an empty list.
    type LinkedList[T any] struct {
    	head, tail *Node[T]
    	size       int
    }

    // Len returns the number of values in the list.
    func (l *LinkedList[T]) Len() int { return l.size }

    // PushFront adds v at the head in O(1).
    func (l *LinkedList[T]) PushFront(v T) {
    	// ?
    }

    // PushBack adds v at the tail in O(1).
    func (l *LinkedList[T]) PushBack(v T) {
    	// ?
    }

    // PopFront removes and returns the head's value, or the zero value
    // and false if the list is empty.
    func (l *LinkedList[T]) PopFront() (T, bool) {
    	var zero T
    	// ?
    	return zero, false
    }

    // RemoveFirst removes the first value for which match returns true,
    // and reports whether it removed anything. Keep head, tail and size right!
    func (l *LinkedList[T]) RemoveFirst(match func(T) bool) bool {
    	// ?
    	return false
    }

    // All yields every value from head to tail, stopping early if the
    // caller's loop stops.
    func (l *LinkedList[T]) All() iter.Seq[T] {
    	return func(yield func(T) bool) {
    		// ?
    	}
    }

    func main() {
    	var feed LinkedList[string]
    	feed.PushBack("bo")
    	feed.PushBack("cy")
    	feed.PushFront("ava")
    	feed.RemoveFirst(func(h string) bool { return h == "bo" })
    	fmt.Println(slices.Collect(feed.All()), feed.Len()) // want: [ava cy] 2
    }
  solution: |
    package main

    import (
    	"fmt"
    	"iter"
    	"slices"
    )

    type Node[T any] struct {
    	Value T
    	Next  *Node[T]
    }

    type LinkedList[T any] struct {
    	head, tail *Node[T]
    	size       int
    }

    func (l *LinkedList[T]) Len() int { return l.size }

    func (l *LinkedList[T]) PushFront(v T) {
    	n := &Node[T]{Value: v, Next: l.head}
    	l.head = n
    	if l.tail == nil {
    		l.tail = n
    	}
    	l.size++
    }

    func (l *LinkedList[T]) PushBack(v T) {
    	n := &Node[T]{Value: v}
    	if l.tail == nil {
    		l.head, l.tail = n, n
    	} else {
    		l.tail.Next = n
    		l.tail = n
    	}
    	l.size++
    }

    func (l *LinkedList[T]) PopFront() (T, bool) {
    	if l.head == nil {
    		var zero T
    		return zero, false
    	}
    	n := l.head
    	l.head = n.Next
    	if l.head == nil {
    		l.tail = nil
    	}
    	n.Next = nil
    	l.size--
    	return n.Value, true
    }

    func (l *LinkedList[T]) RemoveFirst(match func(T) bool) bool {
    	var prev *Node[T]
    	for n := l.head; n != nil; prev, n = n, n.Next {
    		if !match(n.Value) {
    			continue
    		}
    		if prev == nil {
    			l.head = n.Next
    		} else {
    			prev.Next = n.Next
    		}
    		if n == l.tail {
    			l.tail = prev
    		}
    		n.Next = nil
    		l.size--
    		return true
    	}
    	return false
    }

    func (l *LinkedList[T]) All() iter.Seq[T] {
    	return func(yield func(T) bool) {
    		for n := l.head; n != nil; n = n.Next {
    			if !yield(n.Value) {
    				return
    			}
    		}
    	}
    }

    func main() {
    	var feed LinkedList[string]
    	feed.PushBack("bo")
    	feed.PushBack("cy")
    	feed.PushFront("ava")
    	feed.RemoveFirst(func(h string) bool { return h == "bo" })
    	fmt.Println(slices.Collect(feed.All()), feed.Len())
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func values(l *LinkedList[int]) []int {
    	var out []int
    	for n := l.head; n != nil; n = n.Next {
    		out = append(out, n.Value)
    		if len(out) > 10_000 {
    			break // a cycle: don't loop forever
    		}
    	}
    	return out
    }

    func checkList(t *testing.T, what string, l *LinkedList[int], want []int) {
    	t.Helper()
    	got := values(l)
    	if !slices.Equal(got, want) {
    		t.Fatalf("after %s the list is %v, want %v", what, got, want)
    	}
    	if l.Len() != len(want) {
    		t.Fatalf("after %s Len() = %d, want %d", what, l.Len(), len(want))
    	}
    	if len(want) == 0 {
    		if l.head != nil || l.tail != nil {
    			t.Fatalf("after %s the list is empty but head or tail is not nil", what)
    		}
    		return
    	}
    	if l.tail == nil || l.tail.Value != want[len(want)-1] || l.tail.Next != nil {
    		t.Fatalf("after %s tail doesn't point at the last node (%d)", what, want[len(want)-1])
    	}
    }

    func TestPush(t *testing.T) {
    	var l LinkedList[int]
    	l.PushBack(2)
    	checkList(t, "PushBack(2)", &l, []int{2})
    	l.PushFront(1)
    	checkList(t, "PushFront(1)", &l, []int{1, 2})
    	l.PushBack(3)
    	checkList(t, "PushBack(3)", &l, []int{1, 2, 3})
    	l.PushFront(0)
    	checkList(t, "PushFront(0)", &l, []int{0, 1, 2, 3})

    	var m LinkedList[int]
    	m.PushFront(7)
    	m.PushBack(8)
    	checkList(t, "PushFront(7) then PushBack(8) on an empty list", &m, []int{7, 8})
    }

    func TestPopFront(t *testing.T) {
    	var l LinkedList[int]
    	if _, ok := l.PopFront(); ok {
    		t.Fatal("PopFront on an empty list returned ok = true, want false")
    	}
    	l.PushBack(1)
    	l.PushBack(2)
    	if v, ok := l.PopFront(); v != 1 || !ok {
    		t.Fatalf("PopFront() = %d, %v; want 1, true", v, ok)
    	}
    	checkList(t, "one PopFront", &l, []int{2})
    	l.PopFront()
    	checkList(t, "popping the last value", &l, nil)
    	l.PushBack(9)
    	checkList(t, "PushBack(9) on a list emptied by PopFront", &l, []int{9})
    }

    func TestRemoveFirst(t *testing.T) {
    	is := func(x int) func(int) bool { return func(v int) bool { return v == x } }
    	var l LinkedList[int]
    	for i := range 5 {
    		l.PushBack(i)
    	}
    	if !l.RemoveFirst(is(2)) {
    		t.Fatal("RemoveFirst(== 2) returned false, but 2 is in the list")
    	}
    	checkList(t, "removing 2 (middle)", &l, []int{0, 1, 3, 4})
    	l.RemoveFirst(is(4))
    	checkList(t, "removing 4 (tail)", &l, []int{0, 1, 3})
    	l.PushBack(5)
    	checkList(t, "PushBack(5) after removing the tail", &l, []int{0, 1, 3, 5})
    	l.RemoveFirst(is(0))
    	checkList(t, "removing 0 (head)", &l, []int{1, 3, 5})
    	if l.RemoveFirst(is(42)) {
    		t.Fatal("RemoveFirst(== 42) returned true, but 42 isn't in the list")
    	}
    	checkList(t, "removing a missing value", &l, []int{1, 3, 5})
    	l.RemoveFirst(is(1))
    	l.RemoveFirst(is(3))
    	l.RemoveFirst(is(5))
    	checkList(t, "removing everything", &l, nil)
    }

    func TestAll(t *testing.T) {
    	var l LinkedList[int]
    	if got := slices.Collect(l.All()); len(got) != 0 {
    		t.Fatalf("All() on an empty list yielded %v, want nothing", got)
    	}
    	for i := range 5 {
    		l.PushBack(i * 10)
    	}
    	if got := slices.Collect(l.All()); !slices.Equal(got, []int{0, 10, 20, 30, 40}) {
    		t.Fatalf("slices.Collect(All()) = %v, want [0 10 20 30 40]", got)
    	}
    	var seen []int
    	for v := range l.All() {
    		if v > 10 {
    			break
    		}
    		seen = append(seen, v)
    	}
    	if !slices.Equal(seen, []int{0, 10}) {
    		t.Fatalf("breaking out of range All() at 20 saw %v, want [0 10]", seen)
    	}
    }
---

Build it! Clout's trending feed is a linked list of handles, and you're
writing the list itself: adding at both ends, removing, and iterating with
`iter.Seq`.

## Your task

`LinkedList[T]` tracks `head`, `tail` and `size`. Complete:

- `PushFront(v)`: new node pointing at the old head. If the list was empty,
  it's also the tail. O(1).
- `PushBack(v)`: link the old tail to the new node and move `tail`. If the list
  was empty, the new node is both head and tail. O(1).
- `PopFront()`: remove the head and return its value and `true`, or the zero
  value and `false`. If that empties the list, set `tail` to `nil` too.
- `RemoveFirst(match)`: walk the list keeping a `prev` pointer. Unlink the
  first node whose value matches and return `true`. Handle removing the
  **head** (`prev == nil`) and the **tail** (move `tail` back to `prev`).
- `All()`: yield each value from head to tail, and stop as soon as `yield`
  returns `false`:

```go
for n := l.head; n != nil; n = n.Next {
	if !yield(n.Value) {
		return
	}
}
```

After every operation these invariants must hold, and the tests check them:
`size` matches the number of nodes; `head` and `tail` are both `nil` exactly
when the list is empty; `tail` is the last node and its `Next` is `nil`.

Most bugs will be in edge cases, so test yourself: remove the tail, then
`PushBack`. Pop the only element, then push again.
