---
title: container/list
quiz:
  - question: |
      What does this print?

      ```go
      l := list.New()
      l.PushBack("ava")
      bo := l.PushBack("bo")
      l.PushBack("cy")
      l.MoveToFront(bo)
      fmt.Println(l.Front().Value, l.Back().Value)
      ```
    options:
      - text: '`ava cy`'
      - text: '`bo cy`'
        correct: true
      - text: '`bo ava`'
    explanation: |
      `MoveToFront` relinks the `bo` element at the head in O(1). The list is now
      bo, ava, cy, so the front is `bo` and the back is still `cy`.
  - question: Why do values from `container/list` usually need a type assertion like `e.Value.(string)`?
    options:
      - text: '`container/list` predates generics, so `Element.Value` has type `any`'
        correct: true
      - text: Strings are stored as bytes
      - text: Type assertions make lookups O(1)
    explanation: |
      The package was written long before Go 1.18, so it stores values as `any`.
      You get them back with a type assertion, and a wrong assertion panics
      unless you use the comma-ok form.
---

You've built a singly linked list from scratch. Go's standard library ships a
**doubly linked list** in `container/list`, and it's worth knowing both what
it does well and why you won't reach for it often.

## Doubly linked

Each node (called an `*list.Element`) has **two** pointers, `next` and `prev`:

```
nil ◀──[ava]◀──▶[bo]◀──▶[cy]──▶ nil
       front                 back
```

That `prev` pointer fixes the singly linked list's weak spots. Removing the tail
is O(1), because the tail knows its predecessor. And if you hold an `*Element`,
you can remove it, or move it anywhere, in **O(1)** without searching.

## The API

```go
package main

import (
	"container/list"
	"fmt"
)

func main() {
	l := list.New()
	l.PushBack("bo")
	l.PushBack("cy")
	ava := l.PushFront("ava")
	l.InsertAfter("bea", ava)

	for e := l.Front(); e != nil; e = e.Next() {
		fmt.Print(e.Value, " ")
	}
	fmt.Println("| len", l.Len())

	l.Remove(ava)
	for e := l.Back(); e != nil; e = e.Prev() {
		fmt.Print(e.Value.(string), " ")
	}
	fmt.Println()
}
```

```
ava bea bo cy | len 4
cy bo bea 
```

- `PushFront`, `PushBack`, `InsertBefore` and `InsertAfter` return the new
  `*Element`. Hold on to it if you'll want to remove or move it later.
- `Remove(e)`, `MoveToFront(e)`, `MoveToBack(e)`, `MoveBefore` and `MoveAfter` are
  all O(1).
- `Front()`/`Back()` and `e.Next()`/`e.Prev()` walk the list in either direction.
  They return `nil` at the ends, so the traversal loop looks just like ours.
- `list.New()` or a zero `list.List` value both work.

## The catch: it isn't generic

`container/list` predates generics, so `Element.Value` is `any`. Getting a
string back means a type assertion, `e.Value.(string)`, which panics if you're
wrong. Nothing stops someone pushing an `int` into your list of handles. It
also has no `All()` iterator, so you're back to the manual `for e := l.Front()`
loop. Your own `LinkedList[T]` with `iter.Seq` is type-safe and friendlier.

## Where it shines: recently viewed

Clout's app shows each brand manager the **5 most recently viewed** influencer
profiles. Viewing a profile moves it to the front; if the list grows past 5,
the oldest falls off the back. Combine a `container/list` with a map from
handle to `*list.Element`, and every operation is O(1):

```go
package main

import (
	"container/list"
	"fmt"
)

type Recent struct {
	limit int
	order *list.List               // front = most recent
	index map[string]*list.Element // handle → its element
}

func NewRecent(limit int) *Recent {
	return &Recent{limit: limit, order: list.New(), index: map[string]*list.Element{}}
}

func (r *Recent) View(handle string) {
	if e, ok := r.index[handle]; ok {
		r.order.MoveToFront(e)
		return
	}
	r.index[handle] = r.order.PushFront(handle)
	if r.order.Len() > r.limit {
		oldest := r.order.Back()
		r.order.Remove(oldest)
		delete(r.index, oldest.Value.(string))
	}
}

func (r *Recent) List() []string {
	var out []string
	for e := r.order.Front(); e != nil; e = e.Next() {
		out = append(out, e.Value.(string))
	}
	return out
}

func main() {
	r := NewRecent(3)
	for _, h := range []string{"ava", "bo", "cy", "ava", "dee"} {
		r.View(h)
	}
	fmt.Println(r.List())
}
```

```
[dee ava cy]
```

`bo` was the least recently viewed, so it fell off when `dee` arrived. This
map-plus-doubly-linked-list pairing is the classic **LRU cache** design, used
everywhere from databases to CDNs. The map finds an element in O(1); the list
reorders it in O(1). Neither structure could do both alone.

## Wrapping up

That's the end of Data Structures and Algorithms 1. You can now:

- measure algorithms with **Big O** and confirm it with **benchmarks**,
- write and compare **five sorting algorithms**, and know what `slices.Sort` does,
- recognise **exponential** blow-ups and tame some with **memoization**,
- and build **stacks**, **queues**, **ring buffers** and **linked lists**,
  generic and iterable, from scratch.

Next up is [Data Structures and Algorithms 2](/courses/learn-data-structures), which builds
on these foundations with trees, hashmaps, tries, heaps and graphs.

## Further reading

- [`container/list` package documentation](https://pkg.go.dev/container/list)
