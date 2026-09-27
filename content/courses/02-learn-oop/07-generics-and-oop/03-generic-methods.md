---
title: Generic Methods (Go 1.27)
quiz:
  - question: What did Go 1.27 add to methods?
    options:
      - text: Methods can now be declared inside the struct body
      - text: Methods can declare their own type parameters, in addition to any from the receiver's type
        correct: true
      - text: Methods can now be overloaded by parameter type
      - text: Methods can now be declared on types from other packages
    explanation: |
      Before Go 1.27, only functions and types could have type parameters. A method
      could use its receiver's type parameters but not introduce new ones. Now
      `func (b *Bag[T]) Map[U any](f func(T) U) *Bag[U]` is legal.
  - question: |
      Given

      ```go
      func (inv *Inventory) First[T Item]() (T, bool)
      ```

      which call compiles?
    options:
      - text: '`inv.First()`'
      - text: '`inv.First[*Potion]()`'
        correct: true
      - text: '`First[*Potion](inv)`'
      - text: '`inv[*Potion].First()`'
    explanation: |
      `T` only appears in the results, so Go can't infer it from arguments. You
      instantiate it explicitly after the method name: `inv.First[*Potion]()`.
exercise:
  starter: |
    package main

    import "fmt"

    type Item interface{ Name() string }

    type Potion struct{ Heal int }

    func (p *Potion) Name() string { return "potion" }

    type Sword struct{ Damage int }

    func (s *Sword) Name() string { return "sword" }

    type Inventory struct {
    	items []Item
    }

    func (inv *Inventory) Add(it Item) { inv.items = append(inv.items, it) }

    // Count returns how many items in the inventory have the concrete type T.
    func (inv *Inventory) Count[T Item]() int {
    	// ?
    	return 0
    }

    // All returns every item whose concrete type is T, in order, already typed.
    func (inv *Inventory) All[T Item]() []T {
    	// ?
    	return nil
    }

    func main() {
    	var inv Inventory
    	inv.Add(&Potion{Heal: 20})
    	inv.Add(&Sword{Damage: 12})
    	inv.Add(&Potion{Heal: 50})

    	fmt.Println("potions:", inv.Count[*Potion]()) // want 2
    	fmt.Println("swords:", inv.Count[*Sword]())   // want 1
    	for _, p := range inv.All[*Potion]() {
    		fmt.Println("potion heals", p.Heal)
    	}
    }
  solution: |
    package main

    import "fmt"

    type Item interface{ Name() string }

    type Potion struct{ Heal int }

    func (p *Potion) Name() string { return "potion" }

    type Sword struct{ Damage int }

    func (s *Sword) Name() string { return "sword" }

    type Inventory struct {
    	items []Item
    }

    func (inv *Inventory) Add(it Item) { inv.items = append(inv.items, it) }

    func (inv *Inventory) Count[T Item]() int {
    	n := 0
    	for _, it := range inv.items {
    		if _, ok := it.(T); ok {
    			n++
    		}
    	}
    	return n
    }

    func (inv *Inventory) All[T Item]() []T {
    	var out []T
    	for _, it := range inv.items {
    		if t, ok := it.(T); ok {
    			out = append(out, t)
    		}
    	}
    	return out
    }

    func main() {
    	var inv Inventory
    	inv.Add(&Potion{Heal: 20})
    	inv.Add(&Sword{Damage: 12})
    	inv.Add(&Potion{Heal: 50})

    	fmt.Println("potions:", inv.Count[*Potion]())
    	fmt.Println("swords:", inv.Count[*Sword]())
    	for _, p := range inv.All[*Potion]() {
    		fmt.Println("potion heals", p.Heal)
    	}
    }
  tests: |
    package main

    import "testing"

    type Shield struct{}

    func (Shield) Name() string { return "shield" }

    func TestCount(t *testing.T) {
    	var inv Inventory
    	if got := inv.Count[*Potion](); got != 0 {
    		t.Errorf("empty inventory: Count[*Potion]() = %d, want 0", got)
    	}
    	inv.Add(&Potion{Heal: 20})
    	inv.Add(&Sword{Damage: 12})
    	inv.Add(&Potion{Heal: 50})
    	inv.Add(Shield{})
    	if got := inv.Count[*Potion](); got != 2 {
    		t.Errorf("Count[*Potion]() = %d, want 2", got)
    	}
    	if got := inv.Count[*Sword](); got != 1 {
    		t.Errorf("Count[*Sword]() = %d, want 1", got)
    	}
    	if got := inv.Count[Shield](); got != 1 {
    		t.Errorf("Count[Shield]() = %d, want 1", got)
    	}
    }

    func TestAll(t *testing.T) {
    	var inv Inventory
    	inv.Add(&Potion{Heal: 20})
    	inv.Add(&Sword{Damage: 12})
    	inv.Add(&Potion{Heal: 50})
    	potions := inv.All[*Potion]()
    	if len(potions) != 2 {
    		t.Fatalf("All[*Potion]() returned %d items, want 2", len(potions))
    	}
    	if potions[0].Heal != 20 || potions[1].Heal != 50 {
    		t.Errorf("All[*Potion]() heals = [%d %d], want [20 50] (keep inventory order)", potions[0].Heal, potions[1].Heal)
    	}
    	if got := inv.All[Shield](); len(got) != 0 {
    		t.Errorf("All[Shield]() returned %d items, want 0", len(got))
    	}
    }
---

Until Go 1.27, there was an awkward gap in Go's generics. A generic type's methods could *use* the type's parameters, but a method couldn't declare **new** ones. If you wanted a `Map` that turned a `Queue[Hero]` into a `Queue[string]`, you had to write it as a plain function:

```go
func MapQueue[T, U any](q *Queue[T], f func(T) U) *Queue[U]
```

It worked, but it broke the object-style reading of your code. You'd write `MapQueue(heroes, f)` instead of `heroes.Map(f)`.

**Go 1.27 fills the gap: methods can have their own type parameters.** The type parameter list goes after the method name, exactly as it does for a function.

## Mapping a generic container

```go
package main

import (
	"fmt"
	"strings"
)

type Bag[T any] struct {
	items []T
}

func (b *Bag[T]) Add(v T) { b.items = append(b.items, v) }

// Map is a generic method: T comes from the receiver, U is its own.
func (b *Bag[T]) Map[U any](f func(T) U) *Bag[U] {
	out := &Bag[U]{}
	for _, v := range b.items {
		out.Add(f(v))
	}
	return out
}

type Hero struct {
	Name  string
	Level int
}

func main() {
	var party Bag[Hero]
	party.Add(Hero{Name: "Aria", Level: 7})
	party.Add(Hero{Name: "Borin", Level: 12})

	names := party.Map(func(h Hero) string { return strings.ToUpper(h.Name) })
	levels := party.Map(func(h Hero) int { return h.Level })

	fmt.Println(names.items, levels.items)
}
```

```
[ARIA BORIN] [7 12]
```

`party.Map(...)` infers `U` from the function you pass (`string` in one call, `int` in the other), exactly like calling a generic function. You could also write it explicitly as `party.Map[string](...)`.

## Generic methods on ordinary types

The receiver doesn't have to be generic. Here's a very handy pattern for our game: an inventory that holds any `Item`, with a method that finds the first item of a *specific concrete type*:

```go
package main

import "fmt"

type Item interface{ Name() string }

type Potion struct{ Heal int }

func (p *Potion) Name() string { return "potion" }

type Sword struct{ Damage int }

func (s *Sword) Name() string { return "sword" }

type Inventory struct {
	items []Item
}

func (inv *Inventory) Add(it Item) { inv.items = append(inv.items, it) }

// First returns the first item whose concrete type is T.
func (inv *Inventory) First[T Item]() (T, bool) {
	for _, it := range inv.items {
		if t, ok := it.(T); ok {
			return t, true
		}
	}
	var zero T
	return zero, false
}

func main() {
	var inv Inventory
	inv.Add(&Sword{Damage: 12})
	inv.Add(&Potion{Heal: 30})

	if p, ok := inv.First[*Potion](); ok {
		fmt.Printf("drinking a potion: +%d HP\n", p.Heal)
	}
	if s, ok := inv.First[*Sword](); ok {
		fmt.Println("swinging a sword for", s.Damage)
	}
}
```

```
drinking a potion: +30 HP
swinging a sword for 12
```

`inv.First[*Potion]()` returns a real `*Potion`, so `p.Heal` just works. There's no `any`, and no type assertion at the call site. Because `T` appears only in the results, Go can't infer it, so you write the type argument explicitly after the method name.

Notice the type assertion `it.(T)` inside the method. Asserting to a type parameter is allowed, and it's what makes this "find by type" trick possible.

## Instantiate before use

A generic method must be instantiated before you use it as a value, just like a generic function:

```go
find := inv.First[*Sword] // OK: a method value of type func() (*Sword, bool)
bad := inv.First          // compile error: cannot use generic function inv.First without instantiation
```

## Why it took until 1.27

Generic methods were left out of Go 1.18 on purpose, and the reason is interfaces. An interface lists methods that a type's dynamic value must have at runtime, but a generic method isn't one method; it's a whole *family* of them, one per type argument. Go's answer was to allow generic methods but keep them **out of interfaces entirely**. The next lesson looks at exactly what that means.

## Assignment

Give `Inventory` two generic methods, each with its own type parameter `T Item`:

- `Count[T Item]() int` returns how many items have the concrete type `T`.
- `All[T Item]() []T` returns those items, typed as `T`, in the order they were added.

Use a type assertion to the type parameter, `it.(T)`, just like `First` above.
