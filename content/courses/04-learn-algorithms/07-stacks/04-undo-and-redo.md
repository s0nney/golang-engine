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
exercise:
  starter: |
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

    // Edit is one change to the end of the caption: some text was added,
    // or some text was removed.
    type Edit struct {
    	Added   string
    	Removed string
    }

    // Caption is a post caption with undo and redo.
    // The zero value is an empty caption.
    type Caption struct {
    	Text       string
    	undo, redo Stack[Edit]
    }

    // Type adds s to the end of the caption. It's a new change, so it must be
    // undoable, and it wipes out the redo history. Typing "" does nothing.
    func (c *Caption) Type(s string) {
    	// ?
    }

    // Backspace removes the last n bytes (or everything, if there are fewer
    // than n). It's undoable and clears redo. Removing nothing does nothing.
    func (c *Caption) Backspace(n int) {
    	// ?
    }

    // Undo reverses the most recent change and reports whether there was one.
    func (c *Caption) Undo() bool {
    	// ?
    	return false
    }

    // Redo re-applies the most recently undone change and reports whether
    // there was one.
    func (c *Caption) Redo() bool {
    	// ?
    	return false
    }

    func main() {
    	var c Caption
    	c.Type("New drop")
    	c.Type(" tmrw!!")
    	c.Backspace(3)
    	fmt.Printf("%q\n", c.Text) // want: "New drop tmr"
    	c.Undo()
    	fmt.Printf("%q\n", c.Text) // want: "New drop tmrw!!"
    	c.Undo()
    	fmt.Printf("%q\n", c.Text) // want: "New drop"
    	c.Redo()
    	fmt.Printf("%q\n", c.Text) // want: "New drop tmrw!!"
    	c.Type(" #ad")
    	fmt.Println(c.Redo(), c.Text) // want: false New drop tmrw!! #ad
    }
  solution: |
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

    // Edit is one change to the end of the caption: some text was added,
    // or some text was removed.
    type Edit struct {
    	Added   string
    	Removed string
    }

    // Caption is a post caption with undo and redo.
    // The zero value is an empty caption.
    type Caption struct {
    	Text       string
    	undo, redo Stack[Edit]
    }

    func (c *Caption) Type(s string) {
    	if s == "" {
    		return
    	}
    	c.Text += s
    	c.undo.Push(Edit{Added: s})
    	c.redo.Clear()
    }

    func (c *Caption) Backspace(n int) {
    	n = min(n, len(c.Text))
    	if n <= 0 {
    		return
    	}
    	cut := len(c.Text) - n
    	c.undo.Push(Edit{Removed: c.Text[cut:]})
    	c.Text = c.Text[:cut]
    	c.redo.Clear()
    }

    func (c *Caption) Undo() bool {
    	e, ok := c.undo.Pop()
    	if !ok {
    		return false
    	}
    	c.Text = c.Text[:len(c.Text)-len(e.Added)] + e.Removed
    	c.redo.Push(e)
    	return true
    }

    func (c *Caption) Redo() bool {
    	e, ok := c.redo.Pop()
    	if !ok {
    		return false
    	}
    	c.Text = c.Text[:len(c.Text)-len(e.Removed)] + e.Added
    	c.undo.Push(e)
    	return true
    }

    func main() {
    	var c Caption
    	c.Type("New drop")
    	c.Type(" tmrw!!")
    	c.Backspace(3)
    	fmt.Printf("%q\n", c.Text) // want: "New drop tmr"
    	c.Undo()
    	fmt.Printf("%q\n", c.Text) // want: "New drop tmrw!!"
    	c.Undo()
    	fmt.Printf("%q\n", c.Text) // want: "New drop"
    	c.Redo()
    	fmt.Printf("%q\n", c.Text) // want: "New drop tmrw!!"
    	c.Type(" #ad")
    	fmt.Println(c.Redo(), c.Text) // want: false New drop tmrw!! #ad
    }
  tests: |
    package main

    import "testing"

    func check(t *testing.T, step string, c *Caption, want string) {
    	t.Helper()
    	if c.Text != want {
    		t.Fatalf("after %s: Text = %q, want %q", step, c.Text, want)
    	}
    }

    func TestTypeAndBackspace(t *testing.T) {
    	var c Caption
    	c.Type("Hello")
    	c.Type(", Clout")
    	check(t, `Type("Hello"), Type(", Clout")`, &c, "Hello, Clout")
    	c.Backspace(7)
    	check(t, "Backspace(7)", &c, "Hello")
    	c.Backspace(50)
    	check(t, "Backspace(50) on a 5-byte caption", &c, "")
    	c.Backspace(3)
    	check(t, "Backspace(3) on an empty caption", &c, "")
    }

    func TestUndoRedo(t *testing.T) {
    	var c Caption
    	if c.Undo() || c.Redo() {
    		t.Fatal("Undo/Redo on a fresh Caption returned true, want false")
    	}
    	c.Type("Big")
    	c.Type(" sale")
    	c.Backspace(2)
    	check(t, "setup", &c, "Big sa")
    	if !c.Undo() {
    		t.Fatal("Undo() = false, want true")
    	}
    	check(t, "undoing Backspace(2)", &c, "Big sale")
    	c.Undo()
    	check(t, "undoing Type(\" sale\")", &c, "Big")
    	c.Undo()
    	check(t, "undoing Type(\"Big\")", &c, "")
    	if c.Undo() {
    		t.Error("Undo() with nothing left to undo returned true")
    	}
    	c.Redo()
    	c.Redo()
    	check(t, "two Redos", &c, "Big sale")
    	c.Redo()
    	check(t, "three Redos", &c, "Big sa")
    	if c.Redo() {
    		t.Error("Redo() with nothing left to redo returned true")
    	}
    }

    func TestNewChangeClearsRedo(t *testing.T) {
    	var c Caption
    	c.Type("abc")
    	c.Undo()
    	c.Type("xyz")
    	if c.Redo() {
    		t.Errorf("Redo() after a new Type returned true (Text %q): a new change must clear redo", c.Text)
    	}
    	check(t, "Type, Undo, Type", &c, "xyz")

    	c.Undo()
    	c.Type("q")
    	c.Backspace(1)
    	c.Undo()
    	check(t, "Backspace then Undo", &c, "q")
    	c.Undo()
    	c.Backspace(0) // removes nothing, so it must not count as a change
    	if !c.Redo() {
    		t.Error("Backspace(0) cleared the redo history, but it removed nothing and should do nothing")
    	}
    }

    func TestNoOpsAreNotRecorded(t *testing.T) {
    	var c Caption
    	c.Type("")
    	c.Backspace(4)
    	if c.Undo() {
    		t.Error(`Type("") and Backspace on an empty caption were recorded as changes; they should do nothing`)
    	}
    }
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

## Your turn: an undoable caption

Clout's caption box needs undo and redo too. This time an edit isn't a
before-and-after value but a change at the **end** of the text: either some
text was `Added`, or some text was `Removed`. Knowing exactly what changed is
enough to reverse it:

- Undoing `Edit{Added: "!!"}` chops `len("!!")` bytes off the end.
- Undoing `Edit{Removed: "tmrw"}` puts `"tmrw"` back on the end.
- Redoing does the opposite of undoing.

Complete `Type`, `Backspace`, `Undo` and `Redo` using the two stacks and the
three rules above. `Backspace(n)` removes at most `len(c.Text)` bytes. Changes
that do nothing (`Type("")`, or `Backspace` on an empty caption) must **not**
be recorded, and must not clear the redo history.

## Stacks instead of recursion

One more use worth knowing. Recursion uses the call stack implicitly, and any
recursive algorithm can be rewritten with an explicit `Stack` of "work still to
do". That's useful when the recursion could get too deep, or when you want to
pause and resume the work. For example, exploring who-referred-whom among
Clout's influencers depth-first could push each influencer's referrals onto a
stack and pop them one at a time, instead of calling itself recursively.

Stacks give you the *most recent* item. Next up is the structure that gives you
the *oldest*: the queue.
