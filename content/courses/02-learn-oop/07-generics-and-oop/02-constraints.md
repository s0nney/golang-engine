---
title: Constraints
quiz:
  - question: |
      What does `~int` mean in a constraint?
    options:
      - text: Any type except `int`
      - text: Exactly `int` and nothing else
      - text: Any type whose underlying type is `int`, such as `type Gold int`
        correct: true
      - text: An approximate integer, like a float
    explanation: |
      The tilde means "underlying type". Without it, `int` would match only `int`
      itself, and your `type Gold int` wouldn't be allowed.
  - question: |
      Consider these two functions:

      ```go
      func StrongestA(cs []Combatant) Combatant
      func StrongestB[T Combatant](cs []T) T
      ```

      You call both with a `[]*Dragon`. What's the difference?
    options:
      - text: There is no difference at all
      - text: '`StrongestA` won''t accept a `[]*Dragon` directly (you must build a `[]Combatant`), and returns a `Combatant`; `StrongestB` accepts `[]*Dragon` and returns a `*Dragon`'
        correct: true
      - text: '`StrongestB` doesn''t compile because interfaces can''t be constraints'
      - text: '`StrongestA` is faster because it avoids generics'
    explanation: |
      A `[]*Dragon` isn't a `[]Combatant`, so the non-generic version forces a
      conversion loop and gives you back an interface you'd have to assert. The
      generic version keeps the concrete type all the way through.
---

A type parameter's **constraint** says which types are allowed and, just as importantly, what you can *do* with a value of that type inside the generic code.

`any` allows everything, but lets you do almost nothing: you can store and pass `T` values around, but not add them, compare them or call methods on them.

## Constraints are interfaces

In Go, every constraint is an interface. There are two kinds of things an interface can list.

**Methods**, just like the interfaces you know:

```go
type Combatant interface {
	Name() string
	Power() int
}
```

**Type sets**, a union of types, which is only allowed in constraints:

```go
type Number interface {
	~int | ~int64 | ~float64
}
```

The `~` means "any type whose *underlying* type is this". So `~int` includes `int` and also our `type Gold int`. Without the tilde, `Gold` wouldn't qualify.

## Built-in and standard constraints

- **`any`**: every type.
- **`comparable`**: types that support `==` and `!=`. Required for map keys.
- **`cmp.Ordered`** (package `cmp`): types that support `<`, `>` and friends: integers, floats and strings.

## Method constraints: generics meets OOP

This is where generics and interfaces work together. Using a method interface as a constraint lets generic code call methods, while *keeping the concrete type*:

```go
package main

import (
	"cmp"
	"fmt"
	"slices"
)

type Combatant interface {
	Name() string
	Power() int
}

type Hero struct {
	name  string
	level int
}

func (h *Hero) Name() string { return h.name }
func (h *Hero) Power() int   { return h.level * 10 }
func (h *Hero) LevelUp()     { h.level++ }

type Dragon struct {
	name string
	age  int
}

func (d *Dragon) Name() string { return d.name }
func (d *Dragon) Power() int   { return 100 + d.age }

// Strongest returns the most powerful combatant, as its concrete type.
func Strongest[T Combatant](cs []T) T {
	return slices.MaxFunc(cs, func(a, b T) int {
		return cmp.Compare(a.Power(), b.Power())
	})
}

type Gold int

func Total[N ~int | ~float64](amounts ...N) N {
	var sum N
	for _, a := range amounts {
		sum += a
	}
	return sum
}

func main() {
	party := []*Hero{{name: "Aria", level: 7}, {name: "Borin", level: 12}}
	best := Strongest(party) // best is a *Hero, not a Combatant
	best.LevelUp()           // so Hero-only methods are available
	fmt.Println(best.Name(), best.Power())

	lair := []*Dragon{{name: "Ember", age: 40}, {name: "Smaug", age: 170}}
	fmt.Println(Strongest(lair).Name())

	fmt.Println(Total(Gold(100), Gold(250)), Total(1.5, 2.25))
}
```

```
Borin 130
Smaug
350 3.75
```

Look at `best.LevelUp()`. `LevelUp` isn't in `Combatant`, but because `Strongest` returns `T`, and `T` is `*Hero` here, we get a real `*Hero` back. With a plain `func Strongest(cs []Combatant) Combatant` we'd get an interface and need a type assertion. Worse, we couldn't even pass `party` directly, because a `[]*Hero` is not a `[]Combatant` in Go.

`Total` shows a type-set constraint: `sum += a` compiles because every type in the set supports `+`. And the result for `Gold` inputs is a `Gold`.

## Inline constraints

For one-off constraints you can write them in place, as `Total` did with `~int | ~float64`. `[T interface{ Power() int }]` works too. Name the constraint when it's reused or meaningful.

## Rule of thumb

- Use an **interface as a value type** when a collection holds **mixed** types: a battle with heroes *and* dragons.
- Use an **interface as a constraint** when each call deals with **one** type at a time and you want that type back.

## Further reading

- [Go by Example: Generics](https://gobyexample.com/generics)
- [Learn Go with Tests: Generics](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/generics)
