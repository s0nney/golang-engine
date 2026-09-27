---
title: Errors Are Interfaces Too
quiz:
  - question: Why does `error` count as polymorphism?
    options:
      - text: Because errors are always strings underneath
      - text: Because Go has an `Exception` class hierarchy
      - text: Because `error` is an interface with one method, `Error() string`, so any type with that method can be returned as an error
        correct: true
      - text: Because `errors.New` returns different types on different platforms
    explanation: |
      `error` is just `interface { Error() string }`. Your own struct types can be
      errors and carry extra data. Code that only logs calls `Error()`; code that
      cares can look at the concrete type or at extra methods.
  - question: |
      `react` uses `switch e := err.(type)` with a `case *OutOfManaError:`.
      What happens when it receives `fmt.Errorf("turn 3: %w", &OutOfManaError{...})`?
    options:
      - text: The `*OutOfManaError` case matches, because `%w` keeps the original
      - text: 'The `default` case runs: the dynamic type is now the wrapper''s type, not `*OutOfManaError`'
        correct: true
      - text: It panics, because a wrapped error can't be used in a type switch
      - text: It doesn't compile
    explanation: |
      A type switch looks only at the *outermost* dynamic type. `fmt.Errorf` with `%w`
      returns a wrapper type, so none of your cases match. To search the whole wrap
      chain, use `errors.AsType`, which works with both concrete types and interfaces.
  - question: What is the advantage of checking for a `Retryable() bool` *method* rather than for the `*CooldownError` *type*?
    options:
      - text: It's faster at runtime
      - text: Method checks don't need the `errors` package
      - text: Any current or future error type that has the method is handled, without the caller knowing its name
        correct: true
      - text: There is no advantage; they are equivalent
    explanation: |
      Asking "what can you do?" instead of "what are you?" is polymorphism again.
      New error types opt in just by adding the method, and the handling code never
      changes.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    type OutOfManaError struct {
    	Spell      string
    	Need, Have int
    }

    // Error should read like "fireball needs 30 mana, you have 12".
    func (e *OutOfManaError) Error() string {
    	// ?
    	return ""
    }

    type CooldownError struct {
    	Spell string
    	Turns int
    }

    func (e *CooldownError) Error() string {
    	return fmt.Sprintf("%s is on cooldown for %d turns", e.Spell, e.Turns)
    }

    // Retryable reports that waiting will fix this error.
    func (e *CooldownError) Retryable() bool { return true }

    var ErrSilenced = errors.New("you are silenced")

    // advice tells the player what to do about the result of a cast:
    //
    //	nil                                        -> "spell cast!"
    //	an *OutOfManaError anywhere in the chain   -> "drink a potion: <Need-Have> more mana needed"
    //	any error with a Retryable() bool method
    //	that returns true, anywhere in the chain   -> "try again later"
    //	anything else                              -> "cannot cast: <err>"
    func advice(err error) string {
    	// ?
    	return "?"
    }

    func main() {
    	results := []error{
    		nil,
    		fmt.Errorf("turn 3: %w", &OutOfManaError{Spell: "fireball", Need: 30, Have: 12}),
    		fmt.Errorf("turn 4: %w", &CooldownError{Spell: "blink", Turns: 2}),
    		fmt.Errorf("turn 5: %w", ErrSilenced),
    	}
    	for _, err := range results {
    		fmt.Println(advice(err))
    	}
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    type OutOfManaError struct {
    	Spell      string
    	Need, Have int
    }

    func (e *OutOfManaError) Error() string {
    	return fmt.Sprintf("%s needs %d mana, you have %d", e.Spell, e.Need, e.Have)
    }

    type CooldownError struct {
    	Spell string
    	Turns int
    }

    func (e *CooldownError) Error() string {
    	return fmt.Sprintf("%s is on cooldown for %d turns", e.Spell, e.Turns)
    }

    // Retryable reports that waiting will fix this error.
    func (e *CooldownError) Retryable() bool { return true }

    var ErrSilenced = errors.New("you are silenced")

    type retryable interface {
    	error
    	Retryable() bool
    }

    func advice(err error) string {
    	if err == nil {
    		return "spell cast!"
    	}
    	if oom, ok := errors.AsType[*OutOfManaError](err); ok {
    		return fmt.Sprintf("drink a potion: %d more mana needed", oom.Need-oom.Have)
    	}
    	if r, ok := errors.AsType[retryable](err); ok && r.Retryable() {
    		return "try again later"
    	}
    	return "cannot cast: " + err.Error()
    }

    func main() {
    	results := []error{
    		nil,
    		fmt.Errorf("turn 3: %w", &OutOfManaError{Spell: "fireball", Need: 30, Have: 12}),
    		fmt.Errorf("turn 4: %w", &CooldownError{Spell: "blink", Turns: 2}),
    		fmt.Errorf("turn 5: %w", ErrSilenced),
    	}
    	for _, err := range results {
    		fmt.Println(advice(err))
    	}
    }
  tests: |
    package main

    import (
    	"errors"
    	"fmt"
    	"testing"
    )

    // stormError is a brand-new error type that advice has never heard of.
    // It can only be handled by its behaviour, not by its type.
    type stormError struct{ passing bool }

    func (e stormError) Error() string   { return "a storm blocks the spell" }
    func (e stormError) Retryable() bool { return e.passing }

    func TestOutOfManaErrorMessage(t *testing.T) {
    	var err error = &OutOfManaError{Spell: "fireball", Need: 30, Have: 12}
    	if got, want := err.Error(), "fireball needs 30 mana, you have 12"; got != want {
    		t.Errorf("Error() = %q, want %q", got, want)
    	}
    }

    func TestAdvice(t *testing.T) {
    	for _, tt := range []struct {
    		desc string
    		err  error
    		want string
    	}{
    		{"nil", nil, "spell cast!"},
    		{"bare *OutOfManaError", &OutOfManaError{Spell: "spark", Need: 5, Have: 0}, "drink a potion: 5 more mana needed"},
    		{"wrapped *OutOfManaError", fmt.Errorf("turn 3: %w", &OutOfManaError{Spell: "fireball", Need: 30, Have: 12}), "drink a potion: 18 more mana needed"},
    		{"doubly wrapped *OutOfManaError", fmt.Errorf("battle: %w", fmt.Errorf("turn 9: %w", &OutOfManaError{Need: 50, Have: 49})), "drink a potion: 1 more mana needed"},
    		{"wrapped *CooldownError", fmt.Errorf("turn 4: %w", &CooldownError{Spell: "blink", Turns: 2}), "try again later"},
    		{"retryable storm (a type advice doesn't know)", fmt.Errorf("turn 6: %w", stormError{passing: true}), "try again later"},
    		{"non-retryable storm", stormError{passing: false}, "cannot cast: a storm blocks the spell"},
    		{"wrapped ErrSilenced", fmt.Errorf("turn 5: %w", ErrSilenced), "cannot cast: turn 5: you are silenced"},
    		{"plain error", errors.New("the scroll is wet"), "cannot cast: the scroll is wet"},
    	} {
    		if got := advice(tt.err); got != tt.want {
    			t.Errorf("advice(%s) = %q, want %q", tt.desc, got, tt.want)
    		}
    	}
    }
---

You've been using polymorphism since your first `if err != nil`. The built-in `error` type is simply an interface:

```go
type error interface {
	Error() string
}
```

Any type with an `Error() string` method *is* an error. In Learn Go you saw the mechanics of [wrapping and `errors.AsType`](/courses/learn-go/errors/errors-astype). This lesson looks at errors through the OOP lens: custom error types are just more types satisfying an interface, and everything from this chapter (dynamic dispatch, type switches, small interfaces) applies to them.

## Errors as objects

Casting a spell can fail for different reasons, and the game wants to react differently to each. So each reason gets its own type, carrying its own data:

```go
type OutOfManaError struct {
	Spell      string
	Need, Have int
}

func (e *OutOfManaError) Error() string {
	return fmt.Sprintf("%s needs %d mana, you have %d", e.Spell, e.Need, e.Have)
}

type CooldownError struct {
	Spell string
	Turns int
}

func (e *CooldownError) Error() string {
	return fmt.Sprintf("%s is on cooldown for %d turns", e.Spell, e.Turns)
}

// Retryable reports that waiting will fix this error.
func (e *CooldownError) Retryable() bool { return true }
```

Code that just logs never needs to know which one it got: `fmt.Println(err)` calls `Error()` through the interface, which is dynamic dispatch. Code that *does* care can ask, using the tools from the last lesson.

## Asking "what are you?" and "what can you do?"

A type switch works on errors like on any other interface value. Notice the third case: it isn't a concrete type but a tiny interface. It matches *any* error that has a `Retryable() bool` method:

```go
package main

import (
	"errors"
	"fmt"
)

type OutOfManaError struct {
	Spell      string
	Need, Have int
}

func (e *OutOfManaError) Error() string {
	return fmt.Sprintf("%s needs %d mana, you have %d", e.Spell, e.Need, e.Have)
}

type CooldownError struct {
	Spell string
	Turns int
}

func (e *CooldownError) Error() string {
	return fmt.Sprintf("%s is on cooldown for %d turns", e.Spell, e.Turns)
}

func (e *CooldownError) Retryable() bool { return true }

type retryable interface {
	Retryable() bool
}

func react(err error) {
	switch e := err.(type) {
	case nil:
		fmt.Println("spell cast!")
	case *OutOfManaError:
		fmt.Printf("drink a potion: %d more mana needed\n", e.Need-e.Have)
	case retryable:
		fmt.Println("retryable?", e.Retryable())
	default:
		fmt.Println("cannot cast:", err)
	}
}

func main() {
	react(nil)
	react(&OutOfManaError{Spell: "fireball", Need: 30, Have: 12})
	react(&CooldownError{Spell: "blink", Turns: 2})
	react(errors.New("the scroll is wet"))
	react(fmt.Errorf("turn 3: %w", &CooldownError{Spell: "blink", Turns: 2}))
}
```

```
spell cast!
drink a potion: 18 more mana needed
retryable? true
cannot cast: the scroll is wet
cannot cast: turn 3: blink is on cooldown for 2 turns
```

Checking for **behaviour** (`retryable`) rather than a **type** (`*CooldownError`) is the same idea as small interfaces in chapter 4: if a new `*StormError` gains a `Retryable` method next month, `react` handles it without being edited.

## The catch: wrapping

Look at the last line. The cooldown error was wrapped with `fmt.Errorf("turn 3: %w", ...)` to add context, and the type switch missed it. A type switch (or `err.(T)` assertion) only sees the *outermost* dynamic type, which is now `fmt`'s wrapper type.

That's why real code uses the `errors` package, which walks the whole wrap chain. As a quick recap from Learn Go:

- **`errors.Is(err, ErrSilenced)`** asks "is this particular *value* anywhere in the chain?". Use it for sentinel values made with `errors.New`.
- **`errors.AsType[T](err)`** (Go 1.26+) asks "is there an error of *type* `T` anywhere in the chain? Give it to me". `T` can be a concrete type *or an interface*:

```go
if oom, ok := errors.AsType[*OutOfManaError](err); ok {
	// oom is a *OutOfManaError, even if err was wrapped
}
if r, ok := errors.AsType[retryable](err); ok && r.Retryable() {
	// some error in the chain knows it can be retried
}
```

For `AsType` the type argument must satisfy `error`, so the behaviour interface used with it embeds `error`: `type retryable interface { error; Retryable() bool }`. That's interface embedding from chapter 5, earning its keep.

One more receiver gotcha: `Error` has a pointer receiver and we return `&OutOfManaError{...}`, so you must ask for `*OutOfManaError`. Asking for `OutOfManaError` (no star) never matches. Pick one style per error type; pointer receivers are the common choice.

## Your turn

Write the spell advisor:

1. Complete `(*OutOfManaError).Error` so it reads like `fireball needs 30 mana, you have 12`.
2. Complete `advice(err)`. Return `"spell cast!"` for `nil`; `"drink a potion: N more mana needed"` if an `*OutOfManaError` is anywhere in the chain (N is `Need - Have`); `"try again later"` if *any* error in the chain has a `Retryable() bool` method returning `true`; otherwise `"cannot cast: "` followed by the error's message.

The tests use an error type your code has never seen, so check for the behaviour, not for `*CooldownError`.

## Further reading

- [Go by Example: Custom Errors](https://gobyexample.com/custom-errors)
- [Errors are values](https://go.dev/blog/errors-are-values)
