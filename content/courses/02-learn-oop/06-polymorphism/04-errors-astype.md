---
title: Polymorphic Errors and errors.AsType
quiz:
  - question: Why does `error` count as polymorphism?
    options:
      - text: Because errors are always strings underneath
      - text: Because `error` is an interface with one method, `Error() string`, so any type with that method can be returned as an error
        correct: true
      - text: Because Go has an `Exception` class hierarchy
      - text: Because `errors.New` returns different types on different platforms
    explanation: |
      `error` is just `interface { Error() string }`. Your own struct types can be
      errors and carry extra data, and callers can recover that data with
      `errors.AsType` (or the older `errors.As`).
  - question: |
      What does this print?

      ```go
      type OutOfManaError struct{ Need, Have int }

      func (e *OutOfManaError) Error() string {
      	return fmt.Sprintf("need %d mana, have %d", e.Need, e.Have)
      }

      func main() {
      	err := fmt.Errorf("cast fireball: %w", &OutOfManaError{Need: 30, Have: 12})
      	if m, ok := errors.AsType[*OutOfManaError](err); ok {
      		fmt.Println(m.Need - m.Have)
      	} else {
      		fmt.Println("other error")
      	}
      }
      ```
    options:
      - text: '`other error`, because the error was wrapped'
      - text: '`18`'
        correct: true
      - text: '`cast fireball: need 30 mana, have 12`'
      - text: It doesn't compile, because `AsType` needs a pointer argument
    explanation: |
      `errors.AsType` walks the wrap chain made by `%w`, finds the
      `*OutOfManaError`, and returns it already typed. 30 - 12 = 18.
---

You've been using polymorphism since your first `if err != nil`. The built-in `error` type is simply an interface:

```go
type error interface {
	Error() string
}
```

Any type with an `Error() string` method *is* an error. That means errors can be rich objects carrying data, not just messages.

## Custom error types

Casting a spell can fail for different reasons, and the game wants to react differently to each:

```go
type OutOfManaError struct {
	Spell      string
	Need, Have int
}

func (e *OutOfManaError) Error() string {
	return fmt.Sprintf("%s needs %d mana, you have %d", e.Spell, e.Need, e.Have)
}
```

A `*OutOfManaError` can be returned anywhere an `error` is expected. Code that only wants to log it calls `err.Error()` (or prints it) without caring about the type. That's polymorphism.

## Getting the details back

Sometimes the caller *does* care: the UI wants to show exactly how much mana is missing. You could try a type assertion, `err.(*OutOfManaError)`, but that breaks as soon as someone wraps the error with extra context using `fmt.Errorf("...: %w", err)`. The outer error is a different type.

The `errors` package handles wrapping for you. Since Go 1.26 the best tool is the generic **`errors.AsType`**:

```go
func AsType[E error](err error) (E, bool)
```

It walks the whole chain of wrapped errors and returns the first one of type `E`, already typed:

```go
package main

import (
	"errors"
	"fmt"
)

var ErrSilenced = errors.New("you are silenced")

type OutOfManaError struct {
	Spell      string
	Need, Have int
}

func (e *OutOfManaError) Error() string {
	return fmt.Sprintf("%s needs %d mana, you have %d", e.Spell, e.Need, e.Have)
}

type Mage struct {
	mana     int
	silenced bool
}

func (m *Mage) Cast(spell string, cost int) error {
	if m.silenced {
		return ErrSilenced
	}
	if cost > m.mana {
		return &OutOfManaError{Spell: spell, Need: cost, Have: m.mana}
	}
	m.mana -= cost
	return nil
}

func castInBattle(m *Mage, spell string, cost int) error {
	if err := m.Cast(spell, cost); err != nil {
		return fmt.Errorf("turn 3: %w", err) // add context, keep the original
	}
	return nil
}

func report(err error) {
	if err == nil {
		fmt.Println("spell cast!")
		return
	}
	if errors.Is(err, ErrSilenced) {
		fmt.Println("can't speak, can't cast")
		return
	}
	if oom, ok := errors.AsType[*OutOfManaError](err); ok {
		fmt.Printf("drink a potion: %d more mana needed\n", oom.Need-oom.Have)
		return
	}
	fmt.Println("unexpected:", err)
}

func main() {
	m := &Mage{mana: 50}
	report(castInBattle(m, "fireball", 30))
	report(castInBattle(m, "meteor", 45))
	m.silenced = true
	report(castInBattle(m, "fireball", 30))
	fmt.Println(castInBattle(&Mage{}, "spark", 5))
}
```

```
spell cast!
drink a potion: 25 more mana needed
can't speak, can't cast
turn 3: spark needs 5 mana, you have 0
```

Two tools, two jobs:

- **`errors.Is(err, target)`** asks "is this particular *value* anywhere in the chain?". Use it for sentinel errors like `ErrSilenced`.
- **`errors.AsType[T](err)`** asks "is there an error of this *type* anywhere in the chain? Give it to me". Use it when you need the data inside.

## `errors.As`, the older way

Before Go 1.26 you'd write:

```go
var oom *OutOfManaError
if errors.As(err, &oom) {
	// use oom
}
```

It does the same job, but you must declare a variable first and pass a *pointer to it*, and passing the wrong thing panics at runtime rather than failing to compile. `errors.AsType` is shorter, scopes the variable to the `if`, and is checked by the compiler. You'll see `errors.As` in older code, and `go fix` in recent Go versions can modernise it for you.

## Pointer or value receivers for errors?

Notice `Error` has a pointer receiver and we return `&OutOfManaError{...}`. That means you must ask for `*OutOfManaError` in `AsType`. Asking for `OutOfManaError` (no star) would never match. Pick one style per error type and stick to it; pointer receivers are the common choice.

## Further reading

- [Go by Example: Custom Errors](https://gobyexample.com/custom-errors)
- [Errors are values](https://go.dev/blog/errors-are-values)
