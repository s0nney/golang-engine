---
title: Undo and Redo
quiz:
  - question: |
      A manager makes changes A, B and C, then presses Undo twice and Redo once.
      What's on the undo stack now (bottom to top)?
    options:
      - text: A, B
        correct: true
      - text: A
      - text: A, B, C
      - text: C, B
    explanation: |
      After A, B, C the undo stack is [A B C]. Undo twice moves C then B to the
      redo stack, leaving [A]. Redo pops B from the redo stack and pushes it back:
      [A B]. C is still waiting on the redo stack.
  - question: Why does making a *new* change clear the redo stack?
    options:
      - text: To save memory
      - text: The redo history described a different future; after a new change those actions may no longer make sense
        correct: true
      - text: Go's slices can't hold more than one stack
    explanation: |
      Redo replays changes you undid. Once you branch off in a new direction, the
      old undone changes were built on a state that no longer exists, so every
      editor throws them away.
---

Undo on its own is a single stack. Real editors, including Clout's campaign
builder, also have **Redo**, which re-applies what you just undid. That takes
**two** stacks working together.

## Two stacks

- **Undo stack**: changes that have been applied, most recent on top.
- **Redo stack**: changes that were undone, most recently undone on top.

The rules:

1. **Do** a new change: apply it, push it onto *undo*, and **clear** *redo*.
2. **Undo**: pop from *undo*, reverse it, push it onto *redo*.
3. **Redo**: pop from *redo*, re-apply it, push it onto *undo*.

Items shuttle back and forth between the two stacks, always taking the most
recent one first.

## In Go

We'll track a campaign's budget. Each change remembers the value before and
after, so it can be reversed.

```go
package main

import "fmt"

type Stack[T any] struct{ items []T }

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }
func (s *Stack[T]) Len() int { return len(s.items) }
func (s *Stack[T]) Clear()   { clear(s.items); s.items = s.items[:0] }
func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items[len(s.items)-1] = zero
	s.items = s.items[:len(s.items)-1]
	return v, true
}

type Change struct{ From, To int }

type Editor struct {
	Budget     int
	undo, redo Stack[Change]
}

func (e *Editor) SetBudget(v int) {
	e.undo.Push(Change{From: e.Budget, To: v})
	e.redo.Clear()
	e.Budget = v
}

func (e *Editor) Undo() bool {
	c, ok := e.undo.Pop()
	if !ok {
		return false
	}
	e.Budget = c.From
	e.redo.Push(c)
	return true
}

func (e *Editor) Redo() bool {
	c, ok := e.redo.Pop()
	if !ok {
		return false
	}
	e.Budget = c.To
	e.undo.Push(c)
	return true
}

func main() {
	var e Editor
	e.SetBudget(500)
	e.SetBudget(800)
	e.SetBudget(1200)

	e.Undo()
	e.Undo()
	fmt.Println("after 2 undos:", e.Budget)

	e.Redo()
	fmt.Println("after redo:", e.Budget)

	e.SetBudget(650) // new change: redo history is gone
	fmt.Println("redo possible?", e.Redo(), "budget:", e.Budget)
}
```

```
after 2 undos: 500
after redo: 800
redo possible? false budget: 650
```

## Things to notice

- **The zero-value `Editor` works.** Both stacks start as nil slices, and
  `Budget` starts at 0. `var e Editor` is all the setup needed.
- **`Clear` uses the `clear` built-in** (Go 1.21+) to zero the elements before
  truncating, for the same garbage-collection reason as `Pop`, and keeps the
  capacity for reuse.
- **Storing both `From` and `To`** makes every change reversible. This is
  sometimes called the *command pattern*: each action is a value that knows how
  to apply and un-apply itself.
- **Undo and Redo are O(1)**, and a new change is O(1) plus the cost of clearing
  the redo stack.

## Stacks instead of recursion

One more use worth knowing. Recursion uses the call stack implicitly, and any
recursive algorithm can be rewritten with an explicit `Stack` of "work still to
do". That's useful when the recursion could get too deep, or when you want to
pause and resume the work. For example, exploring who-referred-whom among
Clout's influencers depth-first could push each influencer's referrals onto a
stack and pop them one at a time, instead of calling itself recursively.

Stacks give you the *most recent* item. Next up is the structure that gives you
the *oldest*: the queue.
