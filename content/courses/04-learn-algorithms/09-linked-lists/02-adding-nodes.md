---
title: Adding to Head and Tail
quiz:
  - question: |
      What does this print?

      ```go
      var l LinkedList[int]
      l.PushBack(2)
      l.PushFront(1)
      l.PushBack(3)
      l.PushFront(0)
      fmt.Println(l.String())
      ```
    options:
      - text: '`[2 1 3 0]`'
      - text: '`[0 1 2 3]`'
        correct: true
      - text: '`[3 2 1 0]`'
      - text: '`[0 1 3 2]`'
    explanation: |
      `PushFront` adds before the current head and `PushBack` after the current
      tail. The list goes `[2]`, `[1 2]`, `[1 2 3]`, `[0 1 2 3]`.
  - question: Without a `tail` pointer, what would `PushBack` cost?
    options:
      - text: O(1)
      - text: O(log n)
      - text: O(n)
        correct: true
    explanation: |
      You'd have to walk from the head to find the last node before linking the
      new one on. Keeping a `tail` pointer turns that walk into a single step.
---

Loose nodes are awkward to work with. Let's wrap them in a `LinkedList[T]`
type that tracks the **head** (first node), the **tail** (last node) and the
**size**.

```
      head                       tail
       │                          │
       ▼                          ▼
     [ava | •]──▶[bo | •]──▶[cy | nil]      size = 3
```

## The type

```go
type Node[T any] struct {
	Value T
	Next  *Node[T]
}

type LinkedList[T any] struct {
	head, tail *Node[T]
	size       int
}
```

The zero value is an empty list: both pointers `nil` and `size` 0. As with the
stack, no constructor is needed.

## PushFront: O(1)

To add at the head, make a node that points at the current head, then make it
the new head. If the list was empty, the new node is the tail as well.

```go
func (l *LinkedList[T]) PushFront(v T) {
	n := &Node[T]{Value: v, Next: l.head}
	l.head = n
	if l.tail == nil {
		l.tail = n
	}
	l.size++
}
```

No matter how long the list is, that's a handful of assignments: **O(1)**.
Compare prepending to a slice, `append([]T{v}, s...)`, which copies every
element.

## PushBack: O(1) thanks to the tail

To add at the end, link the current tail to the new node, then move the tail.
The empty list is a special case, since there's no tail to link from.

```go
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
```

Without `tail`, we'd have to walk the whole list to find the last node every
time: O(n). One extra pointer turns it into **O(1)**. That's a pattern you'll
see again and again in data structures: store a little extra information to
make a common operation cheap.

## Printing it

A `String` method makes the list print nicely with `fmt`:

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

func (l *LinkedList[T]) Len() int { return l.size }

func (l *LinkedList[T]) String() string {
	var b strings.Builder
	b.WriteString("[")
	for n := l.head; n != nil; n = n.Next {
		if n != l.head {
			b.WriteString(" ")
		}
		fmt.Fprint(&b, n.Value)
	}
	b.WriteString("]")
	return b.String()
}

func main() {
	var trending LinkedList[string]
	trending.PushBack("bo")
	trending.PushBack("cy")
	trending.PushFront("ava") // new #1 trending creator
	fmt.Println(trending.String(), trending.Len())
}
```

```
[ava bo cy] 3
```

## A pointer-receiver gotcha

`String` has a pointer receiver, so `fmt.Println(trending)` with a plain
`LinkedList` *value* won't find it, and prints Go's default struct format
(something like `{0xc000010030 0xc000010060 3}`) instead. Either call
`.String()` explicitly as above, or pass a pointer: `fmt.Println(&trending)`.
We'll return `*LinkedList[T]` from functions in later lessons, so this becomes
automatic.

## Keeping the invariants

Every method must keep these statements true:

- `head == nil` exactly when `tail == nil` exactly when `size == 0`.
- Following `Next` from `head` reaches `tail` after `size - 1` hops.
- `tail.Next == nil`.

When a linked-list method has a bug, it's almost always because one of these
invariants was broken in an edge case, usually the empty or one-element list.
Removal, next, has the most edge cases of all.
