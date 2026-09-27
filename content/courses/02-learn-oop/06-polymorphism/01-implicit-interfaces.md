---
title: Implicit Interfaces
quiz:
  - question: What does "polymorphism" mean in programming?
    options:
      - text: One type that can change its fields at runtime
      - text: One piece of code that works with values of many different types through a shared interface
        correct: true
      - text: Having several functions with the same name and different parameters
      - text: Copying a struct into another struct
    explanation: |
      "Poly" means many and "morph" means form. A function that takes an `Attacker`
      works with heroes, archers and dragons alike, each behaving in its own way.
      (Several functions with the same name is *overloading*, which Go doesn't have.)
  - question: |
      Package `combat` declares `type Healer interface{ Heal() int }`. Package `potion`, which does NOT import `combat`, declares `type Potion struct{}` with `func (Potion) Heal() int`. Can a `potion.Potion` be passed where a `combat.Healer` is expected?
    options:
      - text: No, `potion` must import `combat` and declare that `Potion` implements `Healer`
      - text: Yes, because it has the right method set; interface satisfaction is implicit
        correct: true
      - text: Only after a type conversion `combat.Healer(p)` inside package `potion`
      - text: Only if both packages are in the same module
    explanation: |
      Go checks the method set, nothing else. Neither package has to know about the
      other, which is what lets you declare small interfaces next to the code that
      uses them.
---

The fourth pillar of OOP is **polymorphism**: writing code once that works with many different types. You've already seen it in action with `Attacker`. This chapter looks at how it really works in Go.

## "Is-a" without inheritance

In Python, polymorphism usually comes from inheritance: an `Archer` can be used where a `Hero` is expected because `Archer` extends `Hero`. (Python also has duck typing: if it has a `.attack()` method, you can just call it and hope.)

Go gets polymorphism entirely from **interfaces**, and those interfaces are satisfied **implicitly**. If a type has the methods, it's in. No `implements`, no registration, no shared ancestor.

It's often described as *compile-time checked duck typing*: "if it quacks like a duck, it's a duck", but the compiler checks the quacking before your program runs.

## Retrofitting an interface

Here's where implicit satisfaction shines. Imagine these types already exist, written by three different teammates who never talked to each other:

```go
package main

import "fmt"

// Written by the hero team.
type Cleric struct{ Faith int }

func (c Cleric) Heal() int { return c.Faith * 3 }

// Written by the items team.
type Potion struct{ Size string }

func (p Potion) Heal() int {
	if p.Size == "large" {
		return 50
	}
	return 20
}

// Written by the world team.
type Fountain struct{}

func (Fountain) Heal() int { return 100 }
func (Fountain) Drink()    {}

// Now YOU need to total up healing. Declare the interface you need, right here.
type Healer interface {
	Heal() int
}

func totalHealing(sources ...Healer) int {
	total := 0
	for _, s := range sources {
		total += s.Heal()
	}
	return total
}

func main() {
	fmt.Println(totalHealing(Cleric{Faith: 4}, Potion{Size: "large"}, Fountain{}))
}
```

```
162
```

`Healer` was declared *after* all three types, and none of them mention it. They satisfy it anyway, because each has `Heal() int`. `Fountain` has an extra `Drink` method, which doesn't matter: an interface only asks for a *minimum*.

In Java, you'd have to go back and edit all three classes to add `implements Healer`, which might not even be possible if they came from a library.

## Consumers own the interface

That leads to one of Go's most important design habits:

> Define interfaces in the package that **uses** them, not the package that implements them.

The `combat` package declares `Healer` because `combat` is what needs healing. The `items` package just writes a `Potion` with a good `Heal` method. `items` doesn't import `combat`, so the two packages stay independent, and there's no risk of an import cycle.

## Method sets still apply

Remember from chapter 2: if a method has a **pointer receiver**, only the *pointer* type satisfies the interface.

```go
type Priest struct{ mana int }

func (p *Priest) Heal() int { p.mana -= 10; return 40 }

totalHealing(&Priest{mana: 100}) // OK
totalHealing(Priest{mana: 100})  // compile error: Priest does not implement Healer (method Heal has pointer receiver)
```

That makes sense here: `Heal` spends mana, so it needs the real priest, not a copy.

## Accidental satisfaction?

A common worry: "what if a type satisfies my interface by accident?" In practice it's rare, because method names *and* full signatures must match. And when it does happen, it's usually because the type genuinely does the thing. If you want the compiler to confirm that a type satisfies an interface on purpose, there's a one-line trick for that in the last chapter.

## Further reading

- [A Tour of Go: Interfaces are implemented implicitly](https://go.dev/tour/methods/10)
- [Go FAQ: Why doesn't Go have "implements" declarations?](https://go.dev/doc/faq#implements_interface)
