---
title: What Is Encapsulation?
quiz:
  - question: What is the main goal of encapsulation?
    options:
      - text: Making code run faster by hiding it from the garbage collector
      - text: Preventing outside code from putting an object into an invalid state, and letting you change internals freely
        correct: true
      - text: Encrypting fields so hackers can't read them
      - text: Putting every function inside a struct
    explanation: |
      Encapsulation is about control. If only the type's own methods can touch its
      internals, those methods can guarantee the rules (HP never negative, gold never
      negative) and you can later change how the data is stored without breaking callers.
  - question: |
      What does this print?

      ```go
      type Hero struct{ HP, MaxHP int }

      func (h *Hero) Heal(n int) { h.HP = min(h.HP+n, h.MaxHP) }

      func main() {
      	h := &Hero{HP: 90, MaxHP: 100}
      	h.Heal(50)
      	h.HP += 50
      	fmt.Println(h.HP)
      }
      ```
    options:
      - text: '`100`'
      - text: '`150`'
        correct: true
      - text: '`190`'
      - text: '`140`'
    explanation: |
      `Heal` caps HP at 100. But `HP` is exported, so `main` can bypass the method
      and write `h.HP += 50`, giving 150. The rule is only as strong as the weakest
      code allowed to touch the field. That's the problem encapsulation solves.
---

The first pillar of OOP is **encapsulation**: bundling data with the code that manages it, and **hiding** the data so that only that code can change it.

## The problem

Here's a hero whose healing is capped at max HP:

```go
type Hero struct {
	Name  string
	HP    int
	MaxHP int
}

func (h *Hero) Heal(n int) {
	h.HP = min(h.HP+n, h.MaxHP)
}
```

`Heal` carefully enforces the rule. But nothing stops *other* code from ignoring it:

```go
package main

import "fmt"

type Hero struct {
	Name  string
	HP    int
	MaxHP int
}

func (h *Hero) Heal(n int) {
	h.HP = min(h.HP+n, h.MaxHP)
}

func main() {
	aria := &Hero{Name: "Aria", HP: 80, MaxHP: 100}
	aria.Heal(50)
	fmt.Println(aria.HP)

	// Somewhere far away, in the shop code...
	aria.HP += 500 // "super potion"
	fmt.Println(aria.HP)

	// ...and in the combat code
	aria.HP -= 9999
	fmt.Println(aria.HP)
}
```

```
100
600
-9399
```

Our hero now has 600 out of 100 HP, and then -9399 HP. Every system that touches `HP` has to remember the rules, and one of them forgot. In a big codebase, one of them *always* forgets.

## The idea

Encapsulation says: **the data should be private, and the only way to change it should be through methods that enforce the rules.**

It's like a bank. You don't walk into the vault and move gold around yourself. You go to the teller (a method) who checks you have enough before handing it over. The vault's internals are hidden, and the teller is the public interface.

That gives you two big wins:

1. **Invariants hold.** An *invariant* is a rule that should always be true, such as `0 <= HP <= MaxHP`. If only `Heal` and `TakeDamage` can change `HP`, you only need to get those two methods right.
2. **Freedom to change.** If nobody outside can see how `HP` is stored, you can later change it (say, to a float, or to track shields separately) without breaking any calling code.

## How other languages do it

- **Java and C#** have `private`, `protected` and `public` keywords on each field and method.
- **Python** has no real privacy. By convention, a leading underscore (`self._hp`) means "please don't touch", and a double underscore mangles the name. Nothing actually stops you.
- **Go** uses something much simpler: **the first letter of the name**. And the boundary isn't the type, it's the **package**.

We'll see exactly how that works in the next lesson. The key mindset shift to prepare for: in Go you don't ask "is this field private to this *struct*?", you ask "is this name visible outside this *package*?"
