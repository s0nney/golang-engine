---
title: Undo and Redo
difficulty: medium
after: closures
hints:
  - 'Keep the state as local variables inside `NewEditor`: a `past []string` of earlier versions, the `current` string, and a `future []string` of undone versions. Each of the four closures captures the same variables, so they all see each other''s changes.'
  - '`Apply`: push `current` onto `past`, set `current = edit(current)`, and clear `future`. `Undo`: if `past` is empty return `false`; otherwise push `current` onto `future` and pop the last element of `past` into `current`. `Redo` is the mirror image.'
  - 'Store the **results** of edits, not the edit functions. Re-running an edit on undo or redo would call it more than once, and an edit like "append a timestamp" would give a different answer the second time.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    type Editor struct {
    	Apply func(edit func(string) string)
    	Undo  func() bool
    	Redo  func() bool
    	Text  func() string
    }

    func NewEditor(text string) Editor {
    	return Editor{
    		Apply: func(edit func(string) string) {},
    		Undo:  func() bool { return false },
    		Redo:  func() bool { return false },
    		Text:  func() string { return text },
    	}
    }

    func main() {
    	e := NewEditor("draft")
    	e.Apply(strings.ToUpper)
    	e.Apply(func(s string) string { return s + "!" })
    	fmt.Println(e.Text()) // want: DRAFT!
    	e.Undo()
    	fmt.Println(e.Text()) // want: DRAFT
    	e.Redo()
    	fmt.Println(e.Text()) // want: DRAFT!
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    type Editor struct {
    	Apply func(edit func(string) string)
    	Undo  func() bool
    	Redo  func() bool
    	Text  func() string
    }

    func NewEditor(text string) Editor {
    	var past, future []string
    	current := text
    	return Editor{
    		Apply: func(edit func(string) string) {
    			past = append(past, current)
    			current = edit(current)
    			future = nil
    		},
    		Undo: func() bool {
    			if len(past) == 0 {
    				return false
    			}
    			future = append(future, current)
    			current = past[len(past)-1]
    			past = past[:len(past)-1]
    			return true
    		},
    		Redo: func() bool {
    			if len(future) == 0 {
    				return false
    			}
    			past = append(past, current)
    			current = future[len(future)-1]
    			future = future[:len(future)-1]
    			return true
    		},
    		Text: func() string { return current },
    	}
    }

    func main() {
    	e := NewEditor("draft")
    	e.Apply(strings.ToUpper)
    	e.Apply(func(s string) string { return s + "!" })
    	fmt.Println(e.Text())
    	e.Undo()
    	fmt.Println(e.Text())
    	e.Redo()
    	fmt.Println(e.Text())
    }
  tests: |
    package main

    import (
    	"strings"
    	"testing"
    )

    func suffix(s string) func(string) string {
    	return func(doc string) string { return doc + s }
    }

    func expectText(t *testing.T, e Editor, step, want string) {
    	t.Helper()
    	if got := e.Text(); got != want {
    		t.Errorf("after %s: Text() = %q, want %q", step, got, want)
    	}
    }

    func TestApply(t *testing.T) {
    	e := NewEditor("doc")
    	expectText(t, e, "NewEditor(\"doc\")", "doc")
    	e.Apply(suffix(" v1"))
    	expectText(t, e, "Apply(+\" v1\")", "doc v1")
    	e.Apply(strings.ToUpper)
    	expectText(t, e, "Apply(ToUpper)", "DOC V1")
    }

    func TestUndoRedo(t *testing.T) {
    	e := NewEditor("a")
    	e.Apply(suffix("b"))
    	e.Apply(suffix("c"))
    	e.Apply(suffix("d"))
    	expectText(t, e, "three edits", "abcd")

    	steps := []struct {
    		name   string
    		do     func() bool
    		wantOK bool
    		want   string
    	}{
    		{"Undo", e.Undo, true, "abc"},
    		{"Undo", e.Undo, true, "ab"},
    		{"Redo", e.Redo, true, "abc"},
    		{"Undo", e.Undo, true, "ab"},
    		{"Undo", e.Undo, true, "a"},
    		{"Undo (nothing left)", e.Undo, false, "a"},
    		{"Redo", e.Redo, true, "ab"},
    		{"Redo", e.Redo, true, "abc"},
    		{"Redo", e.Redo, true, "abcd"},
    		{"Redo (nothing left)", e.Redo, false, "abcd"},
    	}
    	for i, s := range steps {
    		if ok := s.do(); ok != s.wantOK {
    			t.Errorf("step %d, %s: returned %v, want %v", i+1, s.name, ok, s.wantOK)
    		}
    		expectText(t, e, "step "+s.name, s.want)
    	}
    }

    func TestApplyClearsRedo(t *testing.T) {
    	e := NewEditor("x")
    	e.Apply(suffix("1"))
    	e.Apply(suffix("2"))
    	e.Undo()
    	e.Apply(suffix("3"))
    	expectText(t, e, "undo then a new edit", "x13")
    	if e.Redo() {
    		t.Errorf("Redo() = true after a new edit, want false: a new edit throws away the undone ones")
    	}
    	expectText(t, e, "Redo after a new edit", "x13")
    	e.Undo()
    	e.Undo()
    	expectText(t, e, "two undos", "x")
    }

    func TestNewEditorNothingToUndo(t *testing.T) {
    	e := NewEditor("fresh")
    	if e.Undo() || e.Redo() {
    		t.Errorf("Undo() or Redo() returned true on a brand new editor, want false")
    	}
    	expectText(t, e, "Undo and Redo on a new editor", "fresh")
    }

    func TestEditsRunOnce(t *testing.T) {
    	calls := 0
    	stamp := func(doc string) string {
    		calls++
    		return doc + "#" + strings.Repeat("|", calls)
    	}
    	e := NewEditor("log")
    	e.Apply(stamp)
    	e.Undo()
    	e.Redo()
    	e.Undo()
    	e.Redo()
    	if calls != 1 {
    		t.Errorf("an edit applied once was called %d times after undo/redo, want 1: store the edited text, don't re-run the edit", calls)
    	}
    	expectText(t, e, "Apply(stamp), then undo/redo twice", "log#|")
    }

    func TestEditorsAreIndependent(t *testing.T) {
    	a, b := NewEditor("A"), NewEditor("B")
    	a.Apply(suffix("1"))
    	b.Apply(suffix("2"))
    	b.Apply(suffix("3"))
    	a.Undo()
    	expectText(t, a, "a.Undo()", "A")
    	expectText(t, b, "a.Undo() (b should be untouched)", "B23")
    	if a.Undo() {
    		t.Errorf("a.Undo() = true with nothing left in a's history: editors must not share history")
    	}
    }
---

Every editor needs **undo** and **redo**. Doc2Doc's preview pane builds them
entirely out of closures: no methods, no global state, just four functions that
share the same private history.

Write `NewEditor(text)`. It returns an `Editor` whose four function fields
share one history:

- `Apply(edit)` replaces the text with `edit(text)` and records the old version.
  It also clears anything that could be redone.
- `Undo()` goes back one version and returns `true`, or returns `false` (and
  changes nothing) if there's nothing to undo.
- `Redo()` re-does the last undone edit and returns `true`, or returns `false`
  if there's nothing to redo.
- `Text()` returns the current text.

Each edit function must be called **exactly once**, when it's applied. Undo and
redo move between stored versions; they never re-run edits.

## Example

```go
e := NewEditor("draft")
e.Apply(strings.ToUpper)                          // "DRAFT"
e.Apply(func(s string) string { return s + "!" }) // "DRAFT!"
e.Undo()                                          // true, "DRAFT"
e.Undo()                                          // true, "draft"
e.Undo()                                          // false, still "draft"
e.Redo()                                          // true, "DRAFT"
e.Apply(strings.ToLower)                          // "draft", redo history cleared
e.Redo()                                          // false
```

## Constraints

- Two editors never share history.
- No package-level variables: all state lives in the closures.
