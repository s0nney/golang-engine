---
title: When Not to Use OOP
quiz:
  - question: |
      You need to compute experience points needed for a level: `100 * level * level`. Which is the most idiomatic Go?
    options:
      - text: An `XPCalculator` interface, an `xpCalculatorImpl` struct, and a `NewXPCalculator` constructor
      - text: '`func XPForLevel(level int) int { return 100 * level * level }`'
        correct: true
      - text: A method on a global singleton `var Calc = &Calculator{}`
      - text: A generic `Calculator[T]` type with a self-referential constraint
    explanation: |
      It's a pure calculation with no state and one implementation. A plain
      function is the clearest option. Wrapping it in objects adds ceremony and
      no value.
  - question: When is introducing an interface most clearly justified?
    options:
      - text: Always, for every struct, just in case
      - text: When there are (or you need to substitute, e.g. in tests) at least two implementations that code must treat uniformly
        correct: true
      - text: Only when the interface has at least five methods
      - text: Never; Go discourages interfaces
    explanation: |
      Interfaces pay for themselves when they let code work with more than one
      concrete type. With only one implementation and no need to fake it, the
      interface is just an extra layer to read through.
---

You've now got a full toolbox: methods, encapsulation, interfaces, embedding, generics, dependency injection, functional options. The last lesson is about the most underrated skill of all: **knowing when to leave the tools in the box**.

## Plain functions are fine

Go is not an "everything must be an object" language. Many things are best as ordinary functions:

```go
package main

import "fmt"

func xpForLevel(level int) int {
	return 100 * level * level
}

func levelForXP(xp int) int {
	level := 1
	for xpForLevel(level+1) <= xp {
		level++
	}
	return level
}

func main() {
	fmt.Println(xpForLevel(5))
	fmt.Println(levelForXP(2600))
}
```

```
2500
5
```

No `XPCalculator` interface. No `XPService` struct. No constructor. It's a calculation, so it's a function. The standard library is full of these: `strings.ToUpper`, `slices.Sort`, `math.Max`.

## Warning signs of over-engineering

If you catch yourself doing any of these, stop and ask whether the abstraction is earning its keep:

- **An interface with exactly one implementation**, and no test that needs to fake it. Go's advice: don't define the interface until a second implementation, or a consumer that needs one, shows up. Implicit satisfaction means you can add it later without touching the type.
- **Getters and setters for every field** with no rules to protect. Export the field.
- **A `Manager`, `Helper`, `Service` or `Impl` type** whose methods don't use any of its fields. Those are functions wearing a costume.
- **Deep embedding chains** that recreate a class hierarchy (`Entity` → `Creature` → `Monster` → `Dragon`). Readers have to walk four types to find a method.
- **Factories for factories.** If a constructor needs a constructor, something has gone wrong.

## A before and after

The over-engineered version:

```go
type DamageCalculator interface {
	Calculate(attack, defense int) int
}

type damageCalculatorImpl struct{}

func NewDamageCalculator() DamageCalculator {
	return &damageCalculatorImpl{}
}

func (c *damageCalculatorImpl) Calculate(attack, defense int) int {
	return max(1, attack-defense)
}

// usage
calc := NewDamageCalculator()
dmg := calc.Calculate(30, 12)
```

The Go version:

```go
func damage(attack, defense int) int {
	return max(1, attack-defense)
}

// usage
dmg := damage(30, 12)
```

Same behaviour, a quarter of the code, nothing to navigate through. If one day there really are several damage formulas chosen at runtime, *then* introduce a small interface (or just a `func(int, int) int` field), in the package that needs it.

## So when is OOP the right call?

Use the tools from this course when they solve a real problem:

| Tool | Worth it when... |
|------|------------------|
| Methods on a type | behaviour clearly belongs to that data (`hero.TakeDamage`) |
| Unexported fields + constructor | the type has invariants to protect |
| Interfaces | code must work with **several** concrete types, or you need a fake in tests |
| Embedding | you're genuinely reusing a component's methods |
| Generics | the same algorithm or container is needed for several types |
| Functional options | many optional settings, most callers use few |

## The Go mindset

Go's creators designed the language to push back against complexity. The course started with "clear is better than clever", and it ends there too. The best Go code usually looks almost boring: concrete types, small interfaces discovered from real needs, plain functions for plain calculations, and exactly as much abstraction as the problem calls for. No more.

Now go slay some dragons. Clearly.

## Further reading

- [Go FAQ: Is Go an object-oriented language?](https://go.dev/doc/faq#Is_Go_an_object-oriented_language)
- [Go Proverbs](https://go-proverbs.github.io/)
