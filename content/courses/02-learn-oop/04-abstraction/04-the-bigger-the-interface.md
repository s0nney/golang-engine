---
title: The Bigger the Interface, the Weaker the Abstraction
quiz:
  - question: |
      A function only calls `c.Name()` on its argument. Which parameter type is best?
    options:
      - text: '`c Character`, where `Character` has 12 methods'
      - text: '`c *Hero`'
      - text: '`c interface{ Name() string }` (or a named `Namer` interface)'
        correct: true
      - text: '`c any`'
    explanation: |
      Ask for exactly what you use. A one-method interface lets heroes, dragons,
      shopkeepers and test fakes all be passed in. `*Hero` and `Character` demand
      far more than the function needs; `any` promises nothing, so you couldn't
      call `Name()` at all without a type assertion.
  - question: |
      Your `Character` interface has 10 methods. You want a quick fake for a test that only exercises `Attack`. What's the problem?
    options:
      - text: There's no problem; Go fills in missing methods automatically
      - text: The fake must implement all 10 methods to satisfy the interface, even though the test uses only one
        correct: true
      - text: Test fakes can't implement interfaces
      - text: Interfaces with more than 8 methods are a compile error
    explanation: |
      To satisfy an interface a type needs *every* method in it. Big interfaces
      make every implementation, including tiny test fakes, expensive to write.
      That's one concrete way big interfaces weaken the abstraction.
---

Rob Pike has a proverb for this:

> The bigger the interface, the weaker the abstraction.

It sounds backwards at first. Doesn't a bigger interface *do* more? It does, and that's the problem.

## A tempting design

Coming from class-based languages, you might model every creature in the game with one big interface:

```go
type Character interface {
	Name() string
	HP() int
	MaxHP() int
	Attack() int
	Defend(dmg int) int
	Heal(n int)
	Move(x, y int)
	Inventory() []Item
	Level() int
	GainXP(n int)
	Say(line string)
	Save(w io.Writer) error
}
```

Now write a function that prints a health bar:

```go
func healthBar(c Character) string
```

It only needs `Name`, `HP` and `MaxHP`, but it demands all twelve methods. What does that cost?

- **Fewer things fit.** A destructible barrel has HP but no inventory, no XP and definitely no dialogue. It can't have a health bar unless you give it eight pointless stub methods.
- **Tests get painful.** A fake `Character` for testing `healthBar` needs all twelve methods.
- **It hides less.** `healthBar` could call `Save` or `GainXP`. Nothing in the signature tells the reader it won't.
- **Changes ripple.** Add a thirteenth method and every implementation in the codebase breaks until it's updated.

## The small version

Ask only for what you need:

```go
package main

import (
	"fmt"
	"strings"
)

type Health interface {
	Name() string
	HP() int
	MaxHP() int
}

func healthBar(h Health) string {
	const width = 10
	filled := h.HP() * width / h.MaxHP()
	return fmt.Sprintf("%-8s [%s%s]", h.Name(),
		strings.Repeat("#", filled), strings.Repeat(".", width-filled))
}

type Hero struct {
	name      string
	hp, maxHP int
	xp        int // plus lots of other hero things
}

func (h *Hero) Name() string { return h.name }
func (h *Hero) HP() int      { return h.hp }
func (h *Hero) MaxHP() int   { return h.maxHP }

type Barrel struct{ hp int }

func (b *Barrel) Name() string { return "barrel" }
func (b *Barrel) HP() int      { return b.hp }
func (b *Barrel) MaxHP() int   { return 20 }

func main() {
	things := []Health{
		&Hero{name: "Aria", hp: 70, maxHP: 100},
		&Barrel{hp: 5},
	}
	for _, t := range things {
		fmt.Println(healthBar(t))
	}
}
```

```
Aria     [#######...]
barrel   [##........]
```

The barrel gets a health bar with three honest methods. And anyone reading `healthBar(h Health)` knows immediately that it *only* looks at health.

## Discover interfaces, don't design them up front

Here's a very Go way of working, quite different from class-first design:

1. Write concrete types (`Hero`, `Dragon`, `Barrel`) with the methods they naturally need.
2. Write functions that use them.
3. When a function needs to accept **more than one** concrete type, declare a small interface *right there*, containing exactly the methods that function calls.

Because Go interfaces are satisfied implicitly, you can add `Health` long after `Hero` and `Barrel` were written, without touching them. In languages with `implements`, you'd have to go back and edit every class.

## When bigger is fine

This isn't a ban on multi-method interfaces. `sort.Interface` has three methods (`Len`, `Less`, `Swap`) because sorting genuinely needs all three. `Health` needs three too. The rule is "no *unnecessary* methods", not "one method or bust". And if you do need a bigger interface, build it by embedding small ones, the way `io.ReadWriter` embeds `io.Reader` and `io.Writer`.
