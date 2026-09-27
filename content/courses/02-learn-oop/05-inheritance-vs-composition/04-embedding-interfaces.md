---
title: Embedding Interfaces
quiz:
  - question: |
      What does this print?

      ```go
      type Attacker interface{ Attack() int }

      type Sword struct{}

      func (Sword) Attack() int { return 10 }

      type Enchanted struct {
      	Attacker
      	Bonus int
      }

      func (e Enchanted) Attack() int { return e.Attacker.Attack() + e.Bonus }

      func main() {
      	var a Attacker = Enchanted{Attacker: Enchanted{Attacker: Sword{}, Bonus: 5}, Bonus: 3}
      	fmt.Println(a.Attack())
      }
      ```
    options:
      - text: '`10`'
      - text: '`13`'
      - text: '`15`'
      - text: '`18`'
        correct: true
    explanation: |
      Enchantments wrap each other. The outer one adds 3 to the inner one,
      which adds 5 to the sword's 10. 10 + 5 + 3 = 18. This is the decorator
      pattern, built with an embedded interface.
  - question: |
      A struct embeds an interface field `Attacker` but you create it with `Enchanted{Bonus: 3}` and call a method that isn't overridden. What happens?
    options:
      - text: The method returns the zero value
      - text: It fails to compile
      - text: It panics with a nil pointer dereference, because the embedded interface is nil
        correct: true
      - text: Go picks a default implementation
    explanation: |
      An embedded interface field is just a field holding an interface value, and
      its zero value is nil. Promoted calls go straight to that nil interface and
      panic at runtime. The compiler can't catch it.
exercise:
  starter: |
    package main

    import "fmt"

    type Weapon interface {
    	Attack() int
    	Name() string
    }

    type Sword struct{}

    func (Sword) Attack() int  { return 10 }
    func (Sword) Name() string { return "sword" }

    type Axe struct{}

    func (Axe) Attack() int  { return 13 }
    func (Axe) Name() string { return "axe" }

    // Flaming wraps any Weapon: it deals 5 extra damage and its name
    // gets a "flaming " prefix.
    type Flaming struct {
    	Weapon
    }

    // Blunted wraps any Weapon: it deals half damage (rounding down) and
    // keeps the wrapped weapon's name.
    type Blunted struct {
    	Weapon
    }

    func main() {
    	arsenal := []Weapon{
    		Sword{},
    		Flaming{Sword{}},
    		Blunted{Axe{}},
    		Flaming{Blunted{Axe{}}},
    	}
    	for _, w := range arsenal {
    		fmt.Printf("%-14s %d\n", w.Name(), w.Attack())
    	}
    }
  solution: |
    package main

    import "fmt"

    type Weapon interface {
    	Attack() int
    	Name() string
    }

    type Sword struct{}

    func (Sword) Attack() int  { return 10 }
    func (Sword) Name() string { return "sword" }

    type Axe struct{}

    func (Axe) Attack() int  { return 13 }
    func (Axe) Name() string { return "axe" }

    // Flaming wraps any Weapon: it deals 5 extra damage and its name
    // gets a "flaming " prefix.
    type Flaming struct {
    	Weapon
    }

    func (f Flaming) Attack() int  { return f.Weapon.Attack() + 5 }
    func (f Flaming) Name() string { return "flaming " + f.Weapon.Name() }

    // Blunted wraps any Weapon: it deals half damage (rounding down) and
    // keeps the wrapped weapon's name.
    type Blunted struct {
    	Weapon
    }

    func (b Blunted) Attack() int { return b.Weapon.Attack() / 2 }

    func main() {
    	arsenal := []Weapon{
    		Sword{},
    		Flaming{Sword{}},
    		Blunted{Axe{}},
    		Flaming{Blunted{Axe{}}},
    	}
    	for _, w := range arsenal {
    		fmt.Printf("%-14s %d\n", w.Name(), w.Attack())
    	}
    }
  tests: |
    package main

    import "testing"

    type club struct{ dmg int }

    func (c club) Attack() int  { return c.dmg }
    func (c club) Name() string { return "club" }

    func TestEnchantments(t *testing.T) {
    	for _, tt := range []struct {
    		desc string
    		w    Weapon
    		name string
    		dmg  int
    	}{
    		{"Flaming{Sword{}}", Flaming{Sword{}}, "flaming sword", 15},
    		{"Flaming{Axe{}}", Flaming{Axe{}}, "flaming axe", 18},
    		{"Blunted{Sword{}}", Blunted{Sword{}}, "sword", 5},
    		{"Blunted{Axe{}}", Blunted{Axe{}}, "axe", 6},
    		{"Blunted{club{7}}", Blunted{club{7}}, "club", 3},
    		{"Flaming{Blunted{Axe{}}}", Flaming{Blunted{Axe{}}}, "flaming axe", 11},
    		{"Blunted{Flaming{Axe{}}}", Blunted{Flaming{Axe{}}}, "flaming axe", 9},
    		{"Flaming{Flaming{Sword{}}}", Flaming{Flaming{Sword{}}}, "flaming flaming sword", 20},
    	} {
    		if got := tt.w.Name(); got != tt.name {
    			t.Errorf("%s.Name() = %q, want %q", tt.desc, got, tt.name)
    		}
    		if got := tt.w.Attack(); got != tt.dmg {
    			t.Errorf("%s.Attack() = %d, want %d", tt.desc, got, tt.dmg)
    		}
    	}
    }
---

You can embed **interfaces** as well as structs, in two different places. Both are everyday Go.

## Interfaces inside interfaces

An interface can embed other interfaces. The result has the union of their methods:

```go
type Attacker interface {
	Attack() int
}

type Defender interface {
	Defend(dmg int) int
}

type Combatant interface {
	Attacker
	Defender
}
```

`Combatant` requires both `Attack` and `Defend`. This is exactly how the standard library builds `io.ReadWriter` from `io.Reader` and `io.Writer`, and `io.ReadWriteCloser` from three small interfaces.

It's the right way to get a bigger interface when you truly need one: start small, and **compose**. Functions that only attack still ask for an `Attacker`, and anything that is a `Combatant` automatically fits.

## Interfaces inside structs

A struct can also embed an interface type. The struct then holds *some* value that implements the interface, and that value's methods are promoted.

That sounds abstract, so here's the classic use: **wrapping** a value to change one behaviour and pass the rest through. Let's add enchantments to weapons.

```go
package main

import "fmt"

type Weapon interface {
	Attack() int
	Name() string
}

type Sword struct{}

func (Sword) Attack() int  { return 10 }
func (Sword) Name() string { return "sword" }

type Bow struct{}

func (Bow) Attack() int  { return 7 }
func (Bow) Name() string { return "bow" }

// Flaming wraps any Weapon. It overrides Attack and Name,
// and would pass any other Weapon methods straight through.
type Flaming struct {
	Weapon
}

func (f Flaming) Attack() int  { return f.Weapon.Attack() + 5 }
func (f Flaming) Name() string { return "flaming " + f.Weapon.Name() }

// Sharp only overrides Attack. Name is promoted from the wrapped Weapon.
type Sharp struct {
	Weapon
}

func (s Sharp) Attack() int { return s.Weapon.Attack() * 2 }

func main() {
	arsenal := []Weapon{
		Sword{},
		Flaming{Sword{}},
		Sharp{Bow{}},
		Flaming{Sharp{Sword{}}},
	}
	for _, w := range arsenal {
		fmt.Printf("%-16s %d\n", w.Name(), w.Attack())
	}
}
```

```
sword            10
flaming sword    15
bow              14
flaming sword    25
```

Look at `Sharp{Bow{}}`. `Sharp` never declared a `Name` method, but it still satisfies `Weapon` because `Name` is promoted from the embedded `Weapon` field. That's why it prints plain `bow`. And enchantments stack: `Flaming{Sharp{Sword{}}}` does (10 × 2) + 5 = 25.

This is the **decorator** pattern. In class-based languages you'd build it with inheritance or a lot of boilerplate. In Go, embedding an interface gives you "forward everything I don't override" for free.

You'll find the same trick in real code: wrapping an `http.ResponseWriter` to record the status code, or wrapping an `io.Reader` to count bytes, overriding one method and letting the rest pass through.

## The nil gotcha

An embedded interface is still just a field, and an unset interface field is `nil`:

```go
f := Flaming{}   // forgot to wrap anything
fmt.Println(f.Name())
```

This compiles, since `Flaming` has all the methods, but at runtime `f.Weapon.Name()` is a call on a nil interface and the program panics with *invalid memory address or nil pointer dereference*. The compiler can't save you here, so always construct wrappers with the thing they wrap. A small constructor such as `func Enflame(w Weapon) Flaming` makes that hard to forget.

## Two kinds of embedding, one idea

| You embed...               | You get...                                                   |
|----------------------------|--------------------------------------------------------------|
| an interface in an interface | a bigger contract made of smaller ones                     |
| a struct in a struct       | reused fields and methods of a *specific* type              |
| an interface in a struct   | reused methods of *whatever* value you plug in at runtime   |

The last one is the most flexible: the struct doesn't care whether it's wrapping a `Sword`, a `Bow` or another enchantment.

## Assignment

The blacksmith has two new enchantments, `Flaming` and `Blunted`. Each already embeds a `Weapon`, so right now both simply pass everything through to the weapon inside. Give them their powers by overriding only what changes:

- `Flaming` deals **5 extra** damage and its name gets a `"flaming "` prefix: `Flaming{Sword{}}` is a `"flaming sword"` doing 15.
- `Blunted` deals **half** damage, rounding down, and keeps the wrapped weapon's name. Don't write a `Name` method for it: let the embedded one be promoted.

Enchantments must stack in any order. `Flaming{Blunted{Axe{}}}` is a `"flaming axe"` doing 13 / 2 + 5 = 11.
