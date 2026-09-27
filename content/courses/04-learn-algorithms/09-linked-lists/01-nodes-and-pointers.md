---
title: Nodes and Pointers
quiz:
  - question: |
      What does this print?

      ```go
      c := &Node[string]{Value: "cy"}
      b := &Node[string]{Value: "bo", Next: c}
      a := &Node[string]{Value: "ava", Next: b}
      fmt.Println(a.Next.Next.Value)
      ```
    options:
      - text: '`bo`'
      - text: '`cy`'
        correct: true
      - text: '`ava`'
      - text: It panics with a nil pointer dereference
    explanation: |
      `a.Next` is `b`, and `b.Next` is `c`, so `a.Next.Next.Value` is `"cy"`. One
      more `.Next` would be `nil`, and reading `.Value` through it would panic.
  - question: Why is getting the 1,000th element of a linked list O(n) when it's O(1) for a slice?
    options:
      - text: Linked lists store their elements in sorted order
      - text: Nodes are scattered in memory, so you must follow 999 `Next` pointers to reach it
        correct: true
      - text: Go doesn't allow indexing into structs
    explanation: |
      A slice computes an element's address directly from its index. A linked
      list has no such formula, because each node could be anywhere. The only
      way to find node 1,000 is to walk there, one pointer at a time.
---

Slices store elements side by side in one block of memory. That's what makes
indexing O(1), and it's also what makes inserting or removing at the front
O(n): everything has to shift. A **linked list** makes the opposite trade.

## A chain of nodes

A linked list is built from **nodes**. Each node holds a value and a
**pointer** to the next node. The last node's pointer is `nil`, marking the
end:

```
head
 │
 ▼
[ava | •]──▶[bo | •]──▶[cy | nil]
```

The nodes don't need to be next to each other in memory. Each one can live
anywhere, as long as the previous node knows where to find it.

In Go, a generic node is a struct that refers to its own type through a pointer:

```go
type Node[T any] struct {
	Value T
	Next  *Node[T]
}
```

It *has* to be a pointer. A `Node` containing a `Node` directly would be
infinitely large, and the compiler rejects it with `invalid recursive type`.
A pointer is just an address, a fixed 8 bytes on a 64-bit machine.

## Building and walking a chain

```go
package main

import "fmt"

type Node[T any] struct {
	Value T
	Next  *Node[T]
}

func main() {
	cy := &Node[string]{Value: "cy"}
	bo := &Node[string]{Value: "bo", Next: cy}
	head := &Node[string]{Value: "ava", Next: bo}

	for n := head; n != nil; n = n.Next {
		fmt.Println(n.Value)
	}

	// Insert "bea" between ava and bo: no shifting required.
	head.Next = &Node[string]{Value: "bea", Next: head.Next}

	for n := head; n != nil; n = n.Next {
		fmt.Print(n.Value, " → ")
	}
	fmt.Println("nil")
}
```

```
ava
bo
cy
ava → bea → bo → cy → nil
```

Two things to notice:

- **The traversal loop** `for n := head; n != nil; n = n.Next` is the heartbeat
  of every linked-list algorithm. Start at the head, stop at `nil`, hop along.
- **Inserting in the middle** took one allocation and two pointer assignments.
  Once you're *at* the right spot, inserting is O(1). No shifting, unlike a
  slice.

## The nil pointer gotcha

Follow a `nil` pointer and Go panics: `invalid memory address or nil pointer
dereference`. Every linked-list method must think about the empty list
(`head == nil`) and the last node (`Next == nil`). Most linked-list bugs are
really "forgot the nil case" bugs.

## Linked lists vs slices

| operation | slice | linked list |
|-----------|-------|-------------|
| get element `i` | O(1) | O(n), walk from the head |
| insert or remove at the front | O(n), shift everything | O(1) |
| insert or remove after a node you hold | O(n) | O(1) |
| append at the back | amortized O(1) | O(1) with a tail pointer |
| memory per item | just the value | value + a pointer, and a separate allocation |
| iteration speed | very fast, cache friendly | slower, pointer chasing |

That last row matters more than Big O suggests. Slices sit in contiguous
memory that the CPU cache loves; linked-list nodes are scattered, so every
`n.Next` may be a cache miss. In practice, Go programmers reach for slices far
more often. Linked lists shine when you do lots of inserting and removing at
the ends or in the middle, and they're the foundation of many other structures,
so they're essential to understand.

Next we'll wrap our nodes in a proper `LinkedList[T]` type.
