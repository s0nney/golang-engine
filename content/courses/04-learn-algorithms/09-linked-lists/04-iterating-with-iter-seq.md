---
title: Iterating with iter.Seq
quiz:
  - question: |
      What does this print?

      ```go
      for v := range l.All() { // l holds 10, 20, 30, 40
      	if v > 20 {
      		break
      	}
      	fmt.Print(v, " ")
      }
      ```
    options:
      - text: '`10 20 `'
        correct: true
      - text: '`10 20 30 40 `'
      - text: '`30 40 `'
      - text: It panics because `yield` is called after `break`
    explanation: |
      When the loop body breaks, `yield` returns `false`, and `All` stops
      walking the list. The body printed 10 and 20, then saw 30 and broke out.
      A correct iterator never calls `yield` again after it returns `false`.
  - question: 'Why does `All` check `if !yield(n.Value) { return }`?'
    options:
      - text: '`yield` returns `false` when the caller''s loop has stopped (via `break` or `return`), and the iterator must stop too'
        correct: true
      - text: '`yield` returns `false` when the value is the zero value'
      - text: To skip nil nodes
    explanation: |
      Range-over-func turns the loop body into the `yield` function. A `false`
      result means "no more values, please". Calling `yield` again after that
      makes the runtime panic.
---

Our lists print, but callers can't loop over them without reaching into
`head` and `Next` themselves. That leaks the implementation. Go's answer, since
Go 1.23, is **range-over-func iterators**: give the list an `All` method that
returns an `iter.Seq[T]`, and callers can use a plain `for ... range`.

## The All method

`iter.Seq[T]` is just a named function type:

```go
type Seq[V any] func(yield func(V) bool)
```

An iterator is a function that calls `yield` once per value, and stops early if
`yield` returns `false`. For a linked list that's the traversal loop you
already know, with `yield` in the middle:

```go
func (l *LinkedList[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for n := l.head; n != nil; n = n.Next {
			if !yield(n.Value) {
				return
			}
		}
	}
}
```

When a caller writes `for v := range l.All() { ... }`, Go turns the loop body
into the `yield` function. If the body `break`s or `return`s, `yield` returns
`false`, and we must stop. Keep calling `yield` after that and the runtime
panics.

## Using it

```go
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

func (l *LinkedList[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for n := l.head; n != nil; n = n.Next {
			if !yield(n.Value) {
				return
			}
		}
	}
}

type Influencer struct {
	Handle    string
	Followers int
}

func main() {
	feed := &LinkedList[Influencer]{}
	feed.PushBack(Influencer{"ava", 90_000})
	feed.PushBack(Influencer{"bo", 1_200})
	feed.PushBack(Influencer{"cy", 450_000})
	feed.PushBack(Influencer{"dee", 30_000})

	total := 0
	for inf := range feed.All() {
		total += inf.Followers
	}
	fmt.Println("total followers:", total)

	// Stop at the first account over 100k.
	for inf := range feed.All() {
		if inf.Followers > 100_000 {
			fmt.Println("first big account:", inf.Handle)
			break
		}
	}

	// Iterators plug straight into the standard library.
	all := slices.Collect(feed.All())
	fmt.Println(len(all), all[len(all)-1].Handle)
}
```

```
total followers: 571200
first big account: cy
4 dee
```

## Why this is a big deal

- **Encapsulation.** Callers never see `Node`, `head` or `Next`. You could swap
  the linked list for a slice tomorrow and every `range l.All()` loop would still
  compile.
- **Early exit is free.** `break` stops the traversal immediately. We don't walk
  the rest of the list or build a temporary slice.
- **It composes.** Anything that accepts an `iter.Seq` works with your list:
  `slices.Collect`, `slices.Sorted`, `maps.Collect` with an `iter.Seq2`, and
  your own filter and map helpers from the FP course.

`slices.Sorted(feed.All())` would even collect and sort in one go, for types
that satisfy `cmp.Ordered`.

## An indexed variant with iter.Seq2

Sometimes callers want the position too, like `range` over a slice gives
`i, v`. That's an `iter.Seq2[int, T]`:

```go
func (l *LinkedList[T]) Enumerate() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		i := 0
		for n := l.head; n != nil; n = n.Next {
			if !yield(i, n.Value) {
				return
			}
			i++
		}
	}
}
```

Then `for rank, inf := range feed.Enumerate() { ... }` reads exactly like a
slice loop.

## A caution about mutation

Don't remove nodes from the list *while* ranging over it. `All` holds a pointer
to the current node, and our `RemoveFirst` sets a removed node's `Next` to `nil`,
which would end the loop early. The standard library's iterators carry the same
warning. If you need to filter a list, collect what to remove first, or build a
new list.
