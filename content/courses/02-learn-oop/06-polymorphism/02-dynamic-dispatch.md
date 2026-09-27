---
title: Dynamic Dispatch
quiz:
  - question: |
      What does this print?

      ```go
      var c Combatant = &Hero{Name: "Aria"}
      fmt.Printf("%T\n", c)
      c = &Dragon{Name: "Ember"}
      fmt.Printf("%T\n", c)
      ```
    options:
      - text: '`main.Combatant` twice'
      - text: '`*main.Hero` then `*main.Dragon`'
        correct: true
      - text: '`interface` twice'
      - text: '`*main.Hero` twice, because the variable''s type never changes'
    explanation: |
      `%T` prints the *dynamic* type: the concrete type currently stored inside the
      interface. The variable's *static* type is always `Combatant`, but what it
      holds can change.
  - question: What are the two parts of a non-nil interface value?
    options:
      - text: A name and a method list
      - text: A concrete type and a value of that type
        correct: true
      - text: A pointer and a length
      - text: A package path and a function
    explanation: |
      An interface value is a pair: (dynamic type, dynamic value). Method calls
      look up the method on the dynamic type and call it with the dynamic value.
      That lookup, at runtime, is dynamic dispatch.
---

When you call `c.Attack()` on an interface value, which code runs? The compiler can't know, because `c` might hold a hero one moment and a dragon the next. The decision happens at **runtime**, and that's called **dynamic dispatch**.

## Inside an interface value

An interface value is a little pair:

```
┌────────────────────┐
│ type:  *main.Hero  │   which concrete type is stored
│ value: → Hero{...} │   the actual data (here, a pointer)
└────────────────────┘
```

When you call a method, Go looks at the **type** half, finds that type's method, and calls it with the **value** half as the receiver. Change what's stored, and the same call site runs different code.

## A turn-based battle

Let's put that to work. Every creature in the battle implements one interface, and the battle loop never checks what anything *is*:

```go
package main

import "fmt"

type Combatant interface {
	Name() string
	Attack() int
	TakeDamage(n int)
	Alive() bool
}

type Hero struct {
	name string
	hp   int
}

func (h *Hero) Name() string     { return h.name }
func (h *Hero) Attack() int      { return 15 }
func (h *Hero) TakeDamage(n int) { h.hp -= n }
func (h *Hero) Alive() bool      { return h.hp > 0 }

type Dragon struct {
	name  string
	hp    int
	turns int
}

func (d *Dragon) Name() string { return d.name }

// Dragons charge up: every third attack is a fire blast.
func (d *Dragon) Attack() int {
	d.turns++
	if d.turns%3 == 0 {
		return 40
	}
	return 10
}

func (d *Dragon) TakeDamage(n int) { d.hp -= n / 2 } // thick scales
func (d *Dragon) Alive() bool      { return d.hp > 0 }

func fight(a, b Combatant) {
	for a.Alive() && b.Alive() {
		b.TakeDamage(a.Attack())
		if b.Alive() {
			a.TakeDamage(b.Attack())
		}
	}
	for _, c := range []Combatant{a, b} {
		if c.Alive() {
			fmt.Printf("%s wins! (%T)\n", c.Name(), c)
		}
	}
}

func main() {
	fight(&Hero{name: "Aria", hp: 100}, &Dragon{name: "Ember", hp: 30})
	fight(&Hero{name: "Borin", hp: 30}, &Dragon{name: "Smaug", hp: 80})
}
```

```
Aria wins! (*main.Hero)
Smaug wins! (*main.Dragon)
```

`fight` has no `if` statements about heroes or dragons. The line `b.TakeDamage(a.Attack())` means "whatever `a` is, attack in *its* way; whatever `b` is, take damage in *its* way". Scales, fire blasts, charging up: all of that lives in the concrete types.

Add a `Goblin` or a `Wizard` tomorrow and `fight` handles them unchanged. That's the promise of polymorphism: **new behaviour by adding types, not by editing old code.**

## Static type vs dynamic type

Every interface variable has two types to keep in mind:

- The **static type** is what you declared: `Combatant`. It decides which methods you're *allowed* to call, and it never changes.
- The **dynamic type** is what's stored inside right now, such as `*main.Hero`. It decides which *code* runs. The `%T` verb prints it.

That's why `c.Name()` compiles for any `Combatant`, but `c.hp` doesn't: the static type promises methods, not fields.

## Is it slow?

Dynamic dispatch costs an extra lookup per call, a few nanoseconds. For a game loop running a few thousand calls per frame, it's nothing. Don't avoid interfaces for speed unless a profiler tells you to.

## Further reading

- [A Tour of Go: Interface values](https://go.dev/tour/methods/11)
- [The Laws of Reflection](https://go.dev/blog/laws-of-reflection) (its first half explains interface values beautifully)
