---
title: Compile-Time Interface Checks
quiz:
  - question: What does `var _ Combatant = (*Dragon)(nil)` do?
    options:
      - text: Creates a global nil dragon that the game uses at startup
      - text: Makes the build fail if `*Dragon` doesn't implement `Combatant`, without allocating anything or creating a usable variable
        correct: true
      - text: Registers `Dragon` so that Go knows it implements `Combatant`
      - text: Causes a nil pointer panic at startup
    explanation: |
      The blank identifier `_` throws the value away, and `(*Dragon)(nil)` is just a
      typed nil pointer, so nothing runs or allocates. The only effect is the
      assignment check at compile time.
  - question: |
      `Dragon`'s methods all have pointer receivers. Which check is correct?
    options:
      - text: '`var _ Combatant = Dragon{}`'
      - text: '`var _ Combatant = (*Dragon)(nil)`'
        correct: true
      - text: '`var _ = Combatant(Dragon)`'
      - text: '`var _ Combatant = nil`'
    explanation: |
      With pointer receivers, only `*Dragon` has the methods. `Dragon{}` would fail
      the check, and `Combatant(Dragon)` isn't valid syntax. `var _ Combatant = nil`
      compiles but checks nothing.
---

Implicit interface satisfaction is lovely, but it has a quiet downside. Nothing in `dragon.go` says "a `Dragon` is meant to be a `Combatant`". If someone renames a method, the mistake only shows up where a `*Dragon` is actually *used* as a `Combatant`, which might be in a different package, or nowhere at all yet.

## The bug

```go
type Combatant interface {
	Name() string
	Attack() int
	TakeDamage(n int)
}

type Dragon struct {
	name string
	hp   int
}

func (d *Dragon) Name() string    { return d.name }
func (d *Dragon) Attack() int     { return 40 }
func (d *Dragon) TakeDamge(n int) { d.hp -= n } // typo!
```

This file compiles happily. `*Dragon` simply isn't a `Combatant`, and nobody finds out until the arena code (maybe written by another team, maybe weeks later) tries to put a dragon in a fight.

## The one-line check

Add this line next to the type:

```go
var _ Combatant = (*Dragon)(nil)
```

Now the dragon's own package fails to build immediately:

```
cannot use (*Dragon)(nil) (value of type *Dragon) as Combatant value in variable declaration: *Dragon does not implement Combatant (missing method TakeDamage)
```

The error points straight at the dragon file and names the missing method, so the typo is a ten-second fix.

## How it works

Let's take it apart:

- `var _ Combatant = ...` declares a package-level variable of type `Combatant`. The name `_` (the blank identifier) means "throw it away", so no variable actually exists and nothing can use it.
- `(*Dragon)(nil)` converts `nil` into a nil `*Dragon`. It's a typed value with no allocation.
- Assigning it to a `Combatant` forces the compiler to check that `*Dragon` has every method in `Combatant`.

So the line costs nothing at runtime and has exactly one effect: a compile error if the relationship ever breaks. For value receivers you can write `var _ Combatant = Dragon{}` instead.

Here's a complete program with the check in place:

```go
package main

import (
	"fmt"
	"io"
	"strings"
)

type Combatant interface {
	Name() string
	Attack() int
	TakeDamage(n int)
}

type Dragon struct {
	name string
	hp   int
}

// Compile-time checks: *Dragon must be a Combatant and a fmt.Stringer.
var (
	_ Combatant    = (*Dragon)(nil)
	_ fmt.Stringer = (*Dragon)(nil)
)

func (d *Dragon) Name() string     { return d.name }
func (d *Dragon) Attack() int      { return 40 }
func (d *Dragon) TakeDamage(n int) { d.hp -= n }
func (d *Dragon) String() string   { return fmt.Sprintf("%s (%d HP)", d.name, d.hp) }

// Standard library types can be checked too.
var _ io.Reader = (*strings.Reader)(nil)

func main() {
	var c Combatant = &Dragon{name: "Ember", hp: 300}
	c.TakeDamage(75)
	fmt.Println(c)
}
```

```
Ember (225 HP)
```

## When to use it

This check is a way of writing down **intent**. Use it when:

- a type **exists to implement** an interface, especially one from another package (say, your type is meant to be an `io.Writer` or an `http.Handler`),
- the interface is used far away from the type, so a break wouldn't be noticed nearby,
- the type is part of a library other people depend on.

You don't need it everywhere. If the only place a type is used as the interface is three lines below, the compiler already checks it there. The standard library and big Go projects use this idiom exactly where it documents an important promise.

## Further reading

- [Effective Go: Interface checks](https://go.dev/doc/effective_go#blank_implements)
- [Go FAQ: How can I guarantee my type satisfies an interface?](https://go.dev/doc/faq#guarantee_satisfies_interface)
