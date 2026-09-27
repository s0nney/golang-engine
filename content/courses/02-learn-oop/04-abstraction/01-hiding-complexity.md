---
title: Hiding Complexity
quiz:
  - question: How is abstraction different from encapsulation?
    options:
      - text: They're two names for exactly the same thing
      - text: Encapsulation hides data to protect it; abstraction hides complexity so callers can think in simpler terms
        correct: true
      - text: Abstraction is about performance; encapsulation is about security
      - text: Abstraction only applies to interfaces; encapsulation only to structs
    explanation: |
      They overlap and usually appear together, but the focus differs.
      Encapsulation is about *who is allowed to touch* the internals. Abstraction
      is about *how much the caller needs to know*: `hero.Attack(dragon)` instead
      of forty lines of damage maths.
  - question: Which function signature is the best abstraction for a caller who just wants to know who won a fight?
    options:
      - text: '`func Fight(h *Hero, d *Dragon, rng *rand.Rand, turnLimit int, log io.Writer, crit float64, dodge float64) (int, int, []int, bool)`'
      - text: '`func Fight(h *Hero, d *Dragon) (winner string)`'
        correct: true
      - text: '`func Fight(state map[string]any) map[string]any`'
    explanation: |
      A good abstraction exposes what the caller cares about and hides the rest.
      The first leaks every internal knob. The third hides *too* much: the types
      no longer tell you anything, so you've lost clarity instead of gaining it.
---

The second pillar of OOP is **abstraction**: hiding *how* something works so the people using it can focus on *what* it does.

You use abstractions all day. You press the accelerator, and the car goes. You don't think about fuel injection timing. The pedal is an abstraction over an enormously complicated engine.

## Abstraction vs encapsulation

These two are easy to confuse, and they do overlap:

- **Encapsulation** hides the *data* so nobody can break it. It's about protection.
- **Abstraction** hides the *complexity* so nobody has to understand it. It's about simplicity.

The accelerator is an abstraction. The locked bonnet is encapsulation.

## An abstraction in our game

Here's the combat maths for a single hero attack:

```go
package main

import "fmt"

type Hero struct {
	Name   string
	Attack int
	Crit   bool
}

type Dragon struct {
	Name  string
	HP    int
	Armor int
	Weak  bool // weak to hero weapons this turn
}

// Strike is the abstraction: callers just say "hit the dragon".
func (h *Hero) Strike(d *Dragon) int {
	dmg := h.rawDamage()
	dmg = d.mitigate(dmg)
	d.HP = max(0, d.HP-dmg)
	return dmg
}

// Everything below is detail the caller never sees.
func (h *Hero) rawDamage() int {
	dmg := h.Attack
	if h.Crit {
		dmg *= 2
	}
	return dmg
}

func (d *Dragon) mitigate(dmg int) int {
	dmg -= d.Armor
	if d.Weak {
		dmg += dmg / 2
	}
	return max(1, dmg) // every hit does at least 1
}

func main() {
	aria := &Hero{Name: "Aria", Attack: 30, Crit: true}
	smolder := &Dragon{Name: "Smolder", HP: 200, Armor: 10, Weak: true}

	dealt := aria.Strike(smolder)
	fmt.Printf("%s hits %s for %d! (%d HP left)\n", aria.Name, smolder.Name, dealt, smolder.HP)
}
```

```
Aria hits Smolder for 75! (125 HP left)
```

The game loop only ever calls `aria.Strike(smolder)`. It doesn't know about crits, armour, weakness or the minimum-1-damage rule. That's four rules the caller doesn't need to understand, and four rules we can rebalance later without touching the game loop.

Notice the lower-case helper methods `rawDamage` and `mitigate`. In a real `combat` package, they'd be invisible to other packages. Encapsulation and abstraction working together.

## Levels of abstraction

Good code is layered. Each layer talks in terms of the layer just below it:

1. **Game loop**: "the party fights the dragon".
2. **Combat**: "each hero strikes; the dragon breathes fire".
3. **Strike**: "compute raw damage, apply mitigation, subtract HP".
4. **Maths**: `max(1, dmg - armor)`.

A function should stay at **one** level. A `Fight` function full of armour maths is mixing layer 2 and layer 4, which makes it hard to read.

## Too much abstraction

Abstraction has a cost: every layer is something a reader must jump through. The goal is to hide *irrelevant* detail, not *all* detail. A signature like `func Do(input map[string]any) map[string]any` hides so much that the types no longer tell you anything. That's not simpler, it's just vaguer.

Go's main tool for abstraction is the **interface**, which lets you describe *what* something can do without saying *what it is*. That's the next lesson.
