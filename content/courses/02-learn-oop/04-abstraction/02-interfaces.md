---
title: Interfaces as Contracts
quiz:
  - question: |
      Given:

      ```go
      type Attacker interface {
      	Attack() int
      }

      type Archer struct{ Arrows int }

      func (a Archer) Attack() int { return 8 * min(a.Arrows, 3) }
      ```

      Does `Archer` implement `Attacker`?
    options:
      - text: No, it needs to write `type Archer struct implements Attacker`
      - text: Yes, because it has an `Attack() int` method, which is everything the interface asks for
        correct: true
      - text: Only if it's declared in the same file as `Attacker`
      - text: Only if you convert it with `Attacker(a)` first
    explanation: |
      Go interfaces are satisfied implicitly. There's no `implements` keyword.
      Any type whose method set includes every method in the interface satisfies it.
  - question: |
      What does this print?

      ```go
      func totalDamage(team []Attacker) int {
      	sum := 0
      	for _, a := range team {
      		sum += a.Attack()
      	}
      	return sum
      }

      // Hero.Attack returns 20; Archer{Arrows: 5}.Attack returns 24
      fmt.Println(totalDamage([]Attacker{Hero{}, Archer{Arrows: 5}}))
      ```
    options:
      - text: '`20`'
      - text: '`44`'
        correct: true
      - text: '`45`'
      - text: It doesn't compile because the slice holds two different types
    explanation: |
      A `[]Attacker` can hold any mix of types that satisfy `Attacker`.
      `totalDamage` calls each one's own `Attack` method: 20 + 24 = 44.
exercise:
  starter: |
    package main

    import "fmt"

    type Attacker interface {
    	Attack() int
    }

    type Archer struct {
    	Name   string
    	Arrows int
    }

    // Attack: 8 damage per arrow, firing at most 3 arrows per turn.
    // TODO: add the Attack method for Archer.

    type Mage struct {
    	Name string
    	Mana int
    }

    // Attack: 25 damage if the mage has at least 10 mana, otherwise 2 (a staff bonk).
    // TODO: add the Attack method for Mage.

    // totalDamage adds up the attacks of the whole party.
    func totalDamage(party []Attacker) int {
    	// ?
    	return 0
    }

    func main() {
    	party := []any{Archer{Name: "Lyra", Arrows: 5}, Mage{Name: "Zed", Mana: 30}}
    	var attackers []Attacker
    	for _, p := range party {
    		if a, ok := p.(Attacker); ok {
    			attackers = append(attackers, a)
    		} else {
    			fmt.Printf("%T can't attack yet\n", p)
    		}
    	}
    	fmt.Println("total damage:", totalDamage(attackers)) // want 49
    }
  solution: |
    package main

    import "fmt"

    type Attacker interface {
    	Attack() int
    }

    type Archer struct {
    	Name   string
    	Arrows int
    }

    func (a Archer) Attack() int { return 8 * min(a.Arrows, 3) }

    type Mage struct {
    	Name string
    	Mana int
    }

    func (m Mage) Attack() int {
    	if m.Mana >= 10 {
    		return 25
    	}
    	return 2
    }

    func totalDamage(party []Attacker) int {
    	sum := 0
    	for _, a := range party {
    		sum += a.Attack()
    	}
    	return sum
    }

    func main() {
    	party := []any{Archer{Name: "Lyra", Arrows: 5}, Mage{Name: "Zed", Mana: 30}}
    	var attackers []Attacker
    	for _, p := range party {
    		if a, ok := p.(Attacker); ok {
    			attackers = append(attackers, a)
    		} else {
    			fmt.Printf("%T can't attack yet\n", p)
    		}
    	}
    	fmt.Println("total damage:", totalDamage(attackers)) // want 49
    }
  tests: |
    package main

    import "testing"

    // attackOf calls Attack through the interface, so the test compiles even
    // before Archer and Mage have their methods.
    func attackOf(v any) (int, bool) {
    	a, ok := v.(Attacker)
    	if !ok {
    		return 0, false
    	}
    	return a.Attack(), true
    }

    func TestArcherAttack(t *testing.T) {
    	for _, tt := range []struct{ arrows, want int }{{0, 0}, {2, 16}, {3, 24}, {10, 24}} {
    		got, ok := attackOf(Archer{Name: "Lyra", Arrows: tt.arrows})
    		if !ok {
    			t.Fatal("Archer does not implement Attacker: add an Attack() int method")
    		}
    		if got != tt.want {
    			t.Errorf("Archer{Arrows: %d}.Attack() = %d, want %d", tt.arrows, got, tt.want)
    		}
    	}
    }

    func TestMageAttack(t *testing.T) {
    	for _, tt := range []struct{ mana, want int }{{30, 25}, {10, 25}, {9, 2}, {0, 2}} {
    		got, ok := attackOf(Mage{Name: "Zed", Mana: tt.mana})
    		if !ok {
    			t.Fatal("Mage does not implement Attacker: add an Attack() int method")
    		}
    		if got != tt.want {
    			t.Errorf("Mage{Mana: %d}.Attack() = %d, want %d", tt.mana, got, tt.want)
    		}
    	}
    }

    type fixedAttacker int

    func (f fixedAttacker) Attack() int { return int(f) }

    func TestTotalDamage(t *testing.T) {
    	party := []Attacker{fixedAttacker(10), fixedAttacker(7), fixedAttacker(25)}
    	if got := totalDamage(party); got != 42 {
    		t.Errorf("totalDamage(attackers doing 10, 7, 25) = %d, want 42", got)
    	}
    	if got := totalDamage(nil); got != 0 {
    		t.Errorf("totalDamage(nil) = %d, want 0", got)
    	}
    }
---

An **interface** in Go is a set of method signatures. It describes **behaviour**: what something can *do*, not what it *is*.

```go
type Attacker interface {
	Attack() int
}
```

Read this as a contract: "an `Attacker` is anything that has an `Attack()` method returning an `int`". Heroes, archers, dragons, even a booby-trapped chest could all qualify.

## Satisfying an interface

In Java you'd write `class Archer implements Attacker`. Go has no such keyword. A type satisfies an interface **automatically**, just by having the right methods:

```go
package main

import "fmt"

type Attacker interface {
	Attack() int
}

type Hero struct {
	Name     string
	Strength int
}

func (h Hero) Attack() int { return h.Strength * 2 }

type Archer struct {
	Name   string
	Arrows int
}

func (a Archer) Attack() int { return 8 * min(a.Arrows, 3) }

type Dragon struct {
	Name string
	Age  int
}

func (d Dragon) Attack() int { return 50 + d.Age }

func totalDamage(team []Attacker) int {
	sum := 0
	for _, a := range team {
		sum += a.Attack()
	}
	return sum
}

func main() {
	team := []Attacker{
		Hero{Name: "Aria", Strength: 10},
		Archer{Name: "Lyra", Arrows: 5},
		Dragon{Name: "Ember", Age: 7}, // a tamed dragon joins the party!
	}
	fmt.Println(totalDamage(team))
}
```

```
101
```

`totalDamage` is the payoff. It works with **anything** that can attack, and it has no idea what a hero, archer or dragon is. Add a `Wizard` next month, give it an `Attack` method, and `totalDamage` handles it without a single change.

## The contract works both ways

An interface is a promise in two directions:

- **To the caller** of `totalDamage`: "give me anything with an `Attack() int` method and I'll handle it".
- **To the implementer** of `totalDamage`: "you may *only* call `Attack()` on these values; you don't get to see anything else".

That second part is the abstraction. Inside `totalDamage`, `a.Name` doesn't compile, because `Attacker` doesn't promise a name. The function can't accidentally depend on details of a particular type.

## Interface values

A variable of interface type holds two things: a **concrete value** and its **type**:

```go
var a Attacker = Archer{Name: "Lyra", Arrows: 2}
fmt.Println(a.Attack()) // 16, calls Archer.Attack
a = Hero{Name: "Aria", Strength: 10}
fmt.Println(a.Attack()) // 20, now calls Hero.Attack
```

The same line, `a.Attack()`, runs different code depending on what's inside. That's **polymorphism**, and we'll dive deep into it in chapter 6.

The zero value of an interface is `nil`: no type, no value. Calling a method on a nil interface panics, so don't leave them unset.

## `any` is the empty interface

`any` is an alias for `interface{}`, an interface with **zero** methods. Every type satisfies it, which is why `fmt.Println` accepts `...any`. The downside is that an `any` value promises nothing, so you can't call anything on it without checking its type first. Keep `any` for truly generic plumbing like printing and encoding.

## Where to declare interfaces

A Go habit that surprises newcomers: interfaces usually live **where they're used**, not where they're implemented. `totalDamage` needs an `Attacker`, so the combat package declares `Attacker`. The `hero` package just gives `Hero` an `Attack` method and doesn't import `combat` at all. Implicit satisfaction is what makes that possible.

## Assignment

Make two party members satisfy `Attacker`, then total up their damage:

- Give `Archer` an `Attack() int` method: 8 damage per arrow, firing at most 3 arrows.
- Give `Mage` an `Attack() int` method: 25 damage with at least 10 mana, otherwise a feeble 2.
- Complete `totalDamage` so it adds up `Attack()` for every member of the party.

No `implements` keyword needed: once the methods exist, `main` will find that both types are attackers.

## Further reading

- [A Tour of Go: Interfaces](https://go.dev/tour/methods/9) and [Interfaces are implemented implicitly](https://go.dev/tour/methods/10)
- [Go by Example: Interfaces](https://gobyexample.com/interfaces)
- [Learn Go with Tests: Structs, methods & interfaces](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/structs-methods-and-interfaces)
