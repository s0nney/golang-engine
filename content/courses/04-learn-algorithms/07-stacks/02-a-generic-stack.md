---
title: A Generic Stack
quiz:
  - question: |
      What does this print?

      ```go
      s := &Stack[int]{}
      s.Push(1)
      s.Push(2)
      s.Push(3)
      s.Pop()
      top, _ := s.Peek()
      fmt.Println(top, s.Len())
      ```
    options:
      - text: '`3 3`'
      - text: '`2 2`'
        correct: true
      - text: '`1 2`'
      - text: '`2 3`'
    explanation: |
      After pushing 1, 2, 3 and popping once, `3` is gone. `Peek` looks at the new
      top, `2`, without removing it, so the length is 2.
  - question: Why does `Pop` return `(T, bool)` instead of just `T`?
    options:
      - text: Go methods must return two values
      - text: So the caller can tell "popped a zero value" apart from "the stack was empty"
        correct: true
      - text: The `bool` reports whether the slice had to be reallocated
    explanation: |
      For a `Stack[int]`, popping a real `0` and popping from an empty stack
      would look identical if we only returned `T`. The comma-ok `bool` removes
      the ambiguity without panicking.
  - question: Why does `Pop` set `s.items[last] = zero` before shrinking the slice?
    options:
      - text: It's required for the code to compile
      - text: So the backing array doesn't keep a reference to the popped item, which would stop the garbage collector freeing it
        correct: true
      - text: It makes `Pop` O(1) instead of O(n)
    explanation: |
      Shrinking `len` doesn't erase the old slot. If `T` holds pointers (like
      `*Campaign` or strings), the hidden slot keeps that memory alive. Zeroing
      it lets the garbage collector reclaim it.
---

Let's build a stack that works for any type: undo actions, strings, ints,
whatever Clout throws at it. We'll back it with a slice, where the **end of the
slice is the top of the stack**.

## The code

```go
package main

import "fmt"

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	last := len(s.items) - 1
	v := s.items[last]
	s.items[last] = zero // let the GC reclaim what it pointed to
	s.items = s.items[:last]
	return v, true
}

func (s *Stack[T]) Peek() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

func (s *Stack[T]) Len() int { return len(s.items) }

type Action struct {
	Kind  string
	Value string
}

func main() {
	var undo Stack[Action]
	undo.Push(Action{"add", "ava"})
	undo.Push(Action{"budget", "500"})
	undo.Push(Action{"add", "bo"})

	for undo.Len() > 0 {
		a, _ := undo.Pop()
		fmt.Println("undoing", a.Kind, a.Value)
	}
	_, ok := undo.Pop()
	fmt.Println("anything left?", ok)
}
```

```
undoing add bo
undoing budget 500
undoing add ava
anything left? false
```

## Design choices

**`any` constraint.** A stack never compares its items, so `T any` is enough.
We'll only need `cmp.Ordered` or `comparable` when an algorithm has to compare.

**Zero value is ready to use.** `var undo Stack[Action]` works immediately,
because a nil slice can be appended to. No constructor needed, which is a nice
Go idiom to aim for.

**Pointer receivers.** `Push` and `Pop` change `s.items`, so they must use
`*Stack[T]`. With a value receiver, `Push` would append to a *copy* of the
struct and the caller's stack would never change. Go automatically takes the
address when you call `undo.Push(...)` on an addressable variable.

**Comma-ok instead of panics.** Popping an empty stack isn't really
exceptional, so returning `(T, bool)` lets callers handle it cleanly. For a
`Stack[int]`, it also distinguishes "popped a 0" from "nothing there".

**Zeroing the popped slot.** `s.items[:last]` shrinks the *length*, but the
backing array still holds the old value in the slot just past the end. If `T`
contains pointers, that stale reference keeps the memory alive. Writing `zero`
there lets the garbage collector do its job. (`slices.Delete` does the same
clearing for you, which is one reason to prefer it over hand-rolled
re-slicing.)

## Complexity

- `Push`: **amortized O(1)**. Usually it writes into spare capacity; once in a
  while `append` reallocates and copies (the growth lesson in chapter 6).
- `Pop`, `Peek`, `Len`: **O(1)**. Shrinking a slice never allocates.
- Space: **O(n)** for `n` items.

One quirk: popping never gives memory back. If a stack briefly holds a million
items and then shrinks to ten, the backing array stays big. For most programs
that's fine. If it isn't, copy the survivors into a fresh slice once the stack
gets small, for example `s.items = slices.Clone(s.items)`. (Note that
`slices.Clip` wouldn't help: it only lowers the capacity of the slice header,
and the big backing array stays exactly where it was.)

## Further reading

- [Learn Go with Tests: Generics](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/generics), which test-drives a generic stack of its own.
