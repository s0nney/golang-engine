---
title: Removing Nodes
quiz:
  - question: |
      `RemoveFirst` deletes the **last** node of a three-node list. Which extra
      step is required so the list stays valid?
    options:
      - text: Set `l.head = nil`
      - text: Move `l.tail` back to the previous node
        correct: true
      - text: Nothing, unlinking the node is enough
    explanation: |
      If the removed node was the tail, `l.tail` would still point at it, and the
      next `PushBack` would link a new node onto a node that's no longer in the
      list. The fix is `if n == l.tail { l.tail = prev }`.
  - question: Why does `RemoveFirst` take a `match func(T) bool` instead of a value to compare with `==`?
    options:
      - text: Functions are faster than `==`
      - text: '`LinkedList[T any]` allows types that can''t be compared with `==`, and a method can''t tighten its receiver''s constraint to `comparable`'
        correct: true
      - text: Go doesn't allow `==` inside generic code
    explanation: |
      `any` includes slices, maps and funcs, which don't support `==`. Even Go
      1.27's generic methods can only add *new* type parameters, not constrain
      the receiver's `T`. A predicate works for every `T`, like `slices.IndexFunc`,
      and a top-level `func Remove[T comparable]` can wrap it.
---

Adding nodes was mostly about pointers. Removing them is where the edge cases
pile up: the empty list, the head, the tail, and a list with just one node,
which is both head *and* tail.

## PopFront: O(1)

Removing the head is the easy one. Move `head` to the second node. If that
leaves the list empty, `tail` must become `nil` too.

```go
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
	n.Next = nil // don't leave the removed node pointing into the list
	l.size--
	return n.Value, true
}
```

No shifting, no copying: **O(1)**. This is the operation where linked lists
beat slices most clearly.

What about removing the *tail*? In a **singly** linked list, the tail doesn't
know its predecessor, so you'd have to walk from the head to find the second-last
node: O(n). **Doubly** linked lists fix that with a `Prev` pointer, as you'll see
with `container/list` at the end of this chapter.

## Removing from the middle

To remove a node, you need its **predecessor**, because it's the predecessor's
`Next` that has to change:

```
before:  [ava]──▶[bo]──▶[cy]
                 remove bo
after:   [ava]───────────▶[cy]      (ava.Next = bo.Next)
```

So we walk the list keeping track of `prev`:

```go
package main

import (
	"fmt"
	"strings"
)

type Node[T any] struct {
	Value T
	Next  *Node[T]
}

type LinkedList[T any] struct {
	head, tail *Node[T]
	size       int
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

// RemoveFirst removes the first value for which match returns true.
func (l *LinkedList[T]) RemoveFirst(match func(T) bool) bool {
	var prev *Node[T]
	for n := l.head; n != nil; prev, n = n, n.Next {
		if !match(n.Value) {
			continue
		}
		if prev == nil {
			l.head = n.Next // removing the head
		} else {
			prev.Next = n.Next
		}
		if n == l.tail {
			l.tail = prev // removing the tail
		}
		n.Next = nil
		l.size--
		return true
	}
	return false
}

// Remove is a convenience for comparable types.
func Remove[T comparable](l *LinkedList[T], v T) bool {
	return l.RemoveFirst(func(x T) bool { return x == v })
}

func (l *LinkedList[T]) String() string {
	var parts []string
	for n := l.head; n != nil; n = n.Next {
		parts = append(parts, fmt.Sprint(n.Value))
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func main() {
	l := &LinkedList[string]{}
	for _, h := range []string{"ava", "bo", "cy", "dee"} {
		l.PushBack(h)
	}
	Remove(l, "bo")  // middle
	Remove(l, "dee") // tail
	Remove(l, "ava") // head
	fmt.Println(l, l.size, l.head == l.tail)

	l.PushBack("eve")
	fmt.Println(l, "found zed?", Remove(l, "zed"))
}
```

```
[cy] 1 true
[cy eve] found zed? false
```

After three removals only `cy` is left, and it's both head and tail. Then
`PushBack("eve")` works correctly, which proves the tail was fixed up when we
removed `dee`. Without the `if n == l.tail` line, `eve` would have been linked
onto the removed `dee` node and silently lost.

## Why a predicate?

`LinkedList[T any]` accepts any type, including slices and maps, which can't be
compared with `==`. A method can't narrow its receiver's constraint. Even
Go 1.27's **generic methods** can only declare *extra* type parameters, not
tighten `T`. So the method takes a `match` function (the same design as
`slices.IndexFunc` and `slices.DeleteFunc`), and a top-level generic function
`Remove[T comparable]` offers the convenient `==` version for types that
support it.

The predicate is also just more flexible. Removing every influencer under 100
followers is `l.RemoveFirst(func(i Influencer) bool { return i.Followers < 100 })`
in a loop.

## Complexity

- `PopFront`: **O(1)**.
- `RemoveFirst` / `Remove`: **O(n)** to *find* the node, then O(1) to unlink it.
- Removing the tail of a singly linked list: **O(n)**, since we need its predecessor.

That split between "finding" and "unlinking" is key. Linked lists make
structural changes cheap, but they don't make searching any faster than a slice.
