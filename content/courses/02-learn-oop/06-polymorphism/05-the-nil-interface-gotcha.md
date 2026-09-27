---
title: The Nil Interface Gotcha
quiz:
  - question: |
      What does this print?

      ```go
      type SpellError struct{}

      func (*SpellError) Error() string { return "fizzle" }

      func cast() error {
      	var e *SpellError // nil pointer
      	return e
      }

      func main() {
      	err := cast()
      	fmt.Println(err == nil)
      }
      ```
    options:
      - text: '`true`'
      - text: '`false`'
        correct: true
      - text: It panics
      - text: It doesn't compile
    explanation: |
      The returned interface holds (type: `*SpellError`, value: nil). An interface
      is only `nil` when *both* halves are empty, so `err == nil` is false even
      though the pointer inside is nil.
  - question: What's the simplest way to avoid the nil-interface bug in functions that return `error`?
    options:
      - text: Always return `nil` literally on success, never a typed nil pointer variable
        correct: true
      - text: Compare errors with `reflect.DeepEqual`
      - text: Use value receivers for all methods
      - text: Never create custom error types
    explanation: |
      If the success path does `return nil`, the interface really is empty. The bug
      only appears when a nil *pointer of a concrete type* gets converted into the
      interface.
---

This is one of Go's most famous gotchas, and it follows directly from what you now know about interface values. Let's trip over it on purpose.

## Recap: an interface is a pair

An interface value holds **(type, value)**. There are three situations:

| Interface holds              | `== nil`? |
|------------------------------|-----------|
| (no type, no value)          | **true**  |
| (`*Dragon`, pointer to a dragon) | false |
| (`*Dragon`, **nil pointer**)  | **false** |

The third row is the gotcha. The interface isn't empty: it knows it's holding a `*Dragon`. The dragon pointer just happens to be nil.

## The bug

Here's a very natural-looking piece of game code:

```go
package main

import "fmt"

type QuestError struct {
	Quest string
}

func (e *QuestError) Error() string { return "quest failed: " + e.Quest }

func completeQuest(name string, success bool) error {
	var qerr *QuestError // nil *QuestError
	if !success {
		qerr = &QuestError{Quest: name}
	}
	return qerr // BUG: converts a nil pointer into a non-nil interface
}

func main() {
	err := completeQuest("slay the dragon", true)
	if err != nil {
		fmt.Printf("oh no: %v (type %T)\n", err == nil, err)
	} else {
		fmt.Println("quest complete!")
	}
}
```

```
oh no: false (type *main.QuestError)
```

The quest *succeeded*, `qerr` is nil, and yet `err != nil`. The caller thinks the quest failed. Worse, if the caller printed `err.Error()`, the method would run with a nil receiver and `e.Quest` would panic.

## Why Go works this way

Go can't treat "interface holding a nil pointer" as `nil`, because a nil pointer can still be a perfectly useful value! Methods with pointer receivers can be called on nil pointers:

```go
type Party struct{ members []string }

func (p *Party) Size() int {
	if p == nil {
		return 0
	}
	return len(p.members)
}

var s interface{ Size() int } = (*Party)(nil)
fmt.Println(s.Size()) // 0: the method runs fine on a nil *Party
```

So the interface keeps the type information, and `== nil` only asks "is the interface itself empty?".

## The fix

Return a literal `nil` on the success path, and only convert to `error` when you actually have an error:

```go
func completeQuest(name string, success bool) error {
	if !success {
		return &QuestError{Quest: name}
	}
	return nil // a truly empty interface
}
```

Rules of thumb that keep you safe:

- Functions that return `error` should declare the result as `error`, **never** as `*MyError`.
- Don't store an error in a variable of a concrete pointer type and then return it.
- The same applies to any interface, not just `error`. A function returning `Combatant` that returns a nil `*Dragon` gives the caller a non-nil `Combatant` that panics when used.

## Spotting it

If you ever see output like `<nil>` printed for an error that `!= nil`, or a panic inside a method on a value you *checked* wasn't nil, suspect this gotcha. Printing with `%T` shows you the hidden type:

```go
fmt.Printf("%T %v\n", err, err == nil) // *main.QuestError false
```

## Further reading

- [Go FAQ: Why is my nil error value not equal to nil?](https://go.dev/doc/faq#nil_error)
- [A Tour of Go: Interface values with nil underlying values](https://go.dev/tour/methods/12)
