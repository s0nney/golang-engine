---
title: Self-Referential Constraints
quiz:
  - question: |
      In `type Adder[A Adder[A]] interface { Add(A) A }`, what does the constraint `A Adder[A]` require?
    options:
      - text: That `A` is an integer type
      - text: That `A` has a method `Add` which takes an `A` and returns an `A`, i.e. it can be added to its own type
        correct: true
      - text: That `A` embeds the `Adder` interface
      - text: Nothing; it's a compile error
    explanation: |
      The constraint refers back to the type being constrained. A type satisfies
      it when it can combine with values of *its own* type, such as `Stats.Add(Stats) Stats`.
  - question: |
      What does this print?

      ```go
      type Stats struct{ Str, Dex int }

      func (s Stats) Add(o Stats) Stats { return Stats{s.Str + o.Str, s.Dex + o.Dex} }

      func Sum[A Adder[A]](xs ...A) A {
      	var total A
      	for _, x := range xs {
      		total = total.Add(x)
      	}
      	return total
      }

      fmt.Println(Sum(Stats{1, 2}, Stats{3, 4}, Stats{5, 0}))
      ```
    options:
      - text: '`{9 6}`'
        correct: true
      - text: '`{5 0}`'
      - text: '`15`'
      - text: It panics because `total` starts as nil
    explanation: |
      `total` starts as the zero `Stats{0, 0}`, which is a usable value, and each
      call to `Add` combines field by field: Str 1+3+5 = 9, Dex 2+4+0 = 6.
---

Here's a design problem from our RPG. Heroes gain bonuses from equipment, buffs and potions, each described as a `Stats` value. Gold amounts add up. Resistances add up. We'd like **one** `Sum` function for all of them.

Each type knows how to add *itself* to *another value of the same type*:

```go
func (s Stats) Add(o Stats) Stats
func (g Gold) Add(o Gold) Gold
```

How do we write a constraint for "a type with an `Add` method that takes and returns *its own type*"?

## A regular interface isn't enough

This doesn't work:

```go
type Adder interface {
	Add(Adder) Adder
}
```

`Stats.Add` takes a `Stats`, not an `Adder`, so `Stats` doesn't satisfy it. And even if it did, `Sum` would return an `Adder` interface, not a `Stats`.

## A constraint that refers to itself

What we mean is: "`A` must have a method `Add(A) A`". The constraint needs to mention the very type parameter it constrains. Since **Go 1.26**, a generic type can refer to itself in its own type parameter list, so we can say exactly that:

```go
type Adder[A Adder[A]] interface {
	Add(A) A
}
```

Read `A Adder[A]` as "`A` is a type that is an adder *of `A`*". This pattern is often called **F-bounded** polymorphism in other languages.

```go
package main

import "fmt"

type Adder[A Adder[A]] interface {
	Add(A) A
}

func Sum[A Adder[A]](xs ...A) A {
	var total A
	for _, x := range xs {
		total = total.Add(x)
	}
	return total
}

type Stats struct {
	Str, Dex, Int int
}

func (s Stats) Add(o Stats) Stats {
	return Stats{s.Str + o.Str, s.Dex + o.Dex, s.Int + o.Int}
}

type Gold int

func (g Gold) Add(o Gold) Gold { return g + o }

func main() {
	base := Stats{Str: 10, Dex: 8, Int: 5}
	sword := Stats{Str: 3}
	ring := Stats{Dex: 2, Int: 2}
	fmt.Printf("%+v\n", Sum(base, sword, ring))

	fmt.Println(Sum[Gold](120, 80, 45))
}
```

```
{Str:13 Dex:10 Int:7}
245
```

`Sum(base, sword, ring)` returns a `Stats`, with no interfaces or assertions anywhere. `Sum[Gold]` works too, because `Gold` also knows how to add itself.

## Another classic: comparing to yourself

Tournaments need a champion. Any type that can say whether it "beats" another of its own kind can take part:

```go
type Ranked[T Ranked[T]] interface {
	Beats(other T) bool
}

func Champion[T Ranked[T]](contenders ...T) T {
	best := contenders[0]
	for _, c := range contenders[1:] {
		if c.Beats(best) {
			best = c
		}
	}
	return best
}
```

Give `Hero` a `Beats(o Hero) bool` method comparing levels, and `Champion(aria, borin, cyra)` returns the highest-level `Hero`. Give `Dragon` one comparing hoard size, and dragons can hold their own tournament. Heroes can't accidentally be compared with dragons, because `Hero.Beats` only accepts a `Hero`.

## Before Go 1.26

You could get the same effect with an inline constraint in each function, `func Sum[A interface{ Add(A) A }](xs ...A) A`, because the function's own type parameter list can refer to `A`. What you couldn't do was give that constraint a reusable *name*. Go 1.26 lifted that restriction.

## Use it sparingly

Self-referential constraints are powerful, but they're one of the harder things in Go to read. Reach for them when you have several types sharing an algorithm that must return *the same concrete type* it was given. For most game code, a plain interface or a concrete function is clearer.
