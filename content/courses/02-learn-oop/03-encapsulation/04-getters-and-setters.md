---
title: Getters and Setters
quiz:
  - question: A `Hero` has an unexported field `level`. What should its getter be called in idiomatic Go?
    options:
      - text: '`GetLevel()`'
      - text: '`Level()`'
        correct: true
      - text: '`level()`'
      - text: '`LevelGetter()`'
    explanation: |
      Go getters drop the `Get` prefix: the field `level` gets a getter `Level()`.
      A setter, if you need one, is `SetLevel(n)`. A lower-case `level()` would be
      unexported, so other packages couldn't call it.
  - question: Why might you NOT write a `SetHP` setter for a hero?
    options:
      - text: Go doesn't allow methods that start with `Set`
      - text: Setters are slower than direct field access
      - text: A setter that accepts any value just re-exposes the field; behaviour methods like `TakeDamage` and `Heal` express and protect the rules better
        correct: true
      - text: Setters can only be used on value receivers
    explanation: |
      Encapsulation isn't about wrapping every field in `Get`/`Set` pairs. Offer
      the operations that make sense for the type and enforce its rules. A raw
      `SetHP(9999)` would let callers bypass everything.
exercise:
  starter: |
    package main

    import "fmt"

    // Inventory holds a hero's items and gold. Gold is never negative.
    type Inventory struct {
    	items []string
    	gold  int
    }

    // Add puts an item in the inventory.
    func (inv *Inventory) Add(item string) {
    	inv.items = append(inv.items, item)
    }

    // AddGold adds n gold. Negative amounts are ignored.
    func (inv *Inventory) AddGold(n int) {
    	inv.gold += max(0, n)
    }

    // GetItems returns the items in the order they were added.
    func (inv *Inventory) GetItems() []string {
    	return inv.items
    }

    // SetGold sets the gold to n.
    func (inv *Inventory) SetGold(n int) {
    	inv.gold = n
    }

    func main() {
    	var inv Inventory
    	inv.Add("potion")
    	inv.Add("rope")
    	inv.AddGold(30)

    	items := inv.GetItems()
    	items[0] = "cursed sword" // should NOT change the inventory
    	fmt.Println(inv.GetItems())

    	inv.SetGold(-9999) // should not even be possible
    	fmt.Println(inv.gold)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"slices"
    )

    // Inventory holds a hero's items and gold. Gold is never negative.
    type Inventory struct {
    	items []string
    	gold  int
    }

    // Add puts an item in the inventory.
    func (inv *Inventory) Add(item string) {
    	inv.items = append(inv.items, item)
    }

    // AddGold adds n gold. Negative amounts are ignored.
    func (inv *Inventory) AddGold(n int) {
    	inv.gold += max(0, n)
    }

    // Items returns a copy of the items in the order they were added.
    func (inv *Inventory) Items() []string {
    	return slices.Clone(inv.items)
    }

    // Gold returns the current amount of gold.
    func (inv *Inventory) Gold() int {
    	return inv.gold
    }

    func main() {
    	var inv Inventory
    	inv.Add("potion")
    	inv.Add("rope")
    	inv.AddGold(30)

    	items := inv.Items()
    	items[0] = "cursed sword"
    	fmt.Println(inv.Items())
    	fmt.Println(inv.Gold())
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    type itemser interface{ Items() []string }
    type golder interface{ Gold() int }

    func TestItemsReturnsACopy(t *testing.T) {
    	inv := &Inventory{}
    	inv.Add("potion")
    	inv.Add("rope")
    	g, ok := any(inv).(itemser)
    	if !ok {
    		t.Fatal("*Inventory has no Items() []string method: Go getters drop the Get prefix")
    	}
    	items := g.Items()
    	if want := []string{"potion", "rope"}; !slices.Equal(items, want) {
    		t.Fatalf("Items() = %q, want %q", items, want)
    	}
    	items[0] = "cursed sword"
    	if got := g.Items(); got[0] != "potion" {
    		t.Errorf("changing the slice returned by Items() changed the inventory to %q: return a copy", got)
    	}
    }

    func TestGold(t *testing.T) {
    	inv := &Inventory{}
    	g, ok := any(inv).(golder)
    	if !ok {
    		t.Fatal("*Inventory has no Gold() int getter")
    	}
    	inv.AddGold(30)
    	inv.AddGold(-100)
    	if got := g.Gold(); got != 30 {
    		t.Errorf("after AddGold(30), AddGold(-100): Gold() = %d, want 30", got)
    	}
    }

    func TestNoLeakyAccessors(t *testing.T) {
    	var v any = &Inventory{}
    	if _, ok := v.(interface{ SetGold(int) }); ok {
    		t.Error("SetGold still exists: it lets anyone set negative gold, so remove it (AddGold is the meaningful operation)")
    	}
    	if _, ok := v.(interface{ GetItems() []string }); ok {
    		t.Error("GetItems still exists: rename it to Items")
    	}
    }
---

Once fields are unexported, other packages need *some* way to read them. The usual tool is a **getter**: a method that returns the field's value.

## Go's getter naming

In Java you'd write `getName()`. Go's convention, straight from *Effective Go*, is to **drop the `Get`**:

```go
package hero

type Hero struct {
	name  string
	hp    int
	maxHP int
}

func (h *Hero) Name() string { return h.name }
func (h *Hero) HP() int      { return h.hp }
func (h *Hero) MaxHP() int   { return h.maxHP }
```

Callers write `aria.Name()` and `aria.HP()`, which read like plain English. `aria.GetName()` compiles, but it marks your code as written by someone who hasn't read *Effective Go* yet.

This works neatly with Go's capitalisation rule: the field is `name` (unexported), and the getter is `Name` (exported). No naming clash, no underscores.

Setters, when you need them, keep the `Set` prefix: `SetName(n string)`.

## Don't write setters by reflex

Here's a trap. You make fields private, then immediately write:

```go
func (h *Hero) SetHP(hp int) { h.hp = hp }
```

Congratulations, you've built an exported field with extra steps. Anyone can still `SetHP(-9399)`.

Encapsulation is not "put `Get` and `Set` on every field". It's "offer **meaningful operations** that keep the object valid". For HP, those operations are things that actually happen in the game:

```go
package main

import "fmt"

type Hero struct {
	name  string
	hp    int
	maxHP int
}

func NewHero(name string, maxHP int) *Hero {
	return &Hero{name: name, hp: maxHP, maxHP: maxHP}
}

func (h *Hero) Name() string { return h.name }
func (h *Hero) HP() int      { return h.hp }

func (h *Hero) TakeDamage(n int) {
	h.hp = max(0, h.hp-max(0, n))
}

func (h *Hero) Heal(n int) {
	if h.hp == 0 {
		return // the dead can't be healed, only revived
	}
	h.hp = min(h.maxHP, h.hp+max(0, n))
}

func main() {
	aria := NewHero("Aria", 100)
	aria.TakeDamage(70)
	aria.Heal(500)
	fmt.Println(aria.Name(), aria.HP())

	aria.TakeDamage(-50) // negative damage is ignored, not a sneaky heal
	aria.TakeDamage(150)
	aria.Heal(20)
	fmt.Println(aria.Name(), aria.HP())
}
```

```
Aria 100
Aria 0
```

Every rule is in one place: damage can't be negative, HP can't go below 0 or above max, and fallen heroes stay fallen. There's no way to call these methods and end up with an invalid hero.

## When a plain exported field is fine

Go is pragmatic. If a field has **no rules** attached, just export it:

```go
type Position struct {
	X, Y int
}
```

Wrapping `X` and `Y` in `X()`/`SetX()` adds nothing. Plenty of standard library types, like `http.Request` and `image.Point`, are full of exported fields.

A handy test: *if a caller could set this field to any value of its type and nothing would break*, export it. If some values are invalid, hide it and expose operations instead.

## Returning slices and maps: watch out

A getter that returns an internal slice hands out a way to modify it:

```go
func (inv *Inventory) Items() []string { return inv.items }

items := inv.Items()
items[0] = "cursed sword" // modifies the inventory's own slice!
```

To keep control, return a copy with `slices.Clone(inv.items)`, or return an iterator (`slices.Values(inv.items)`) so callers can loop but not write.

## Assignment

The `Inventory` below has unexported fields, but its accessors undo all the good work. Fix them:

- Rename `GetItems` to `Items`, Go style, and make it return a **copy**, so editing the returned slice can't change the inventory.
- Add a `Gold() int` getter.
- Delete `SetGold`. It lets anyone set the gold to `-9999`, and `AddGold` is already the meaningful operation.

Then update `main` to use the new methods (it won't compile until you do).

## Further reading

- [Effective Go: Getters](https://go.dev/doc/effective_go#Getters)
