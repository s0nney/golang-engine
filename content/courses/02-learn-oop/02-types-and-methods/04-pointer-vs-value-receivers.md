---
title: Pointer vs Value Receivers
quiz:
  - question: |
      What does this print?

      ```go
      type Hero struct{ HP int }

      func (h Hero) Heal(n int) { h.HP += n }

      func main() {
      	aria := Hero{HP: 50}
      	aria.Heal(20)
      	fmt.Println(aria.HP)
      }
      ```
    options:
      - text: '`70`'
      - text: '`50`'
        correct: true
      - text: '`20`'
      - text: It doesn't compile, because `Heal` needs a pointer
    explanation: |
      `Heal` has a value receiver, so it heals a *copy* of `aria`. The copy is
      thrown away when the method returns. To change the original, use a pointer
      receiver: `func (h *Hero) Heal(n int)`.
  - question: |
      `Hero` has `func (h *Hero) TakeDamage(n int)`. Which value can NOT be stored in a variable of interface type `interface{ TakeDamage(int) }`?
    options:
      - text: '`&Hero{}`'
      - text: '`Hero{}`'
        correct: true
      - text: '`new(Hero)`'
      - text: A `*Hero` returned by `NewHero`
    explanation: |
      A pointer-receiver method belongs to the method set of `*Hero` only, not
      `Hero`. So only `*Hero` values satisfy the interface. The other three options
      are all `*Hero`.
  - question: Which is the usual advice for choosing receivers?
    options:
      - text: Always use value receivers, because pointers are slow
      - text: Use pointer receivers when methods modify the receiver or the struct is large, and keep all methods on a type consistent
        correct: true
      - text: Use value receivers for exported methods and pointer receivers for unexported ones
      - text: Mix them freely; it makes no difference
    explanation: |
      If any method needs a pointer receiver (to mutate, or to avoid copying a big
      struct), give all the type's methods pointer receivers so the type behaves
      consistently.
exercise:
  starter: |
    package main

    import "fmt"

    type Hero struct {
    	Name  string
    	HP    int
    	MaxHP int
    	XP    int
    	Level int
    }

    // TakeDamage lowers HP by n, never below 0.
    func (h Hero) TakeDamage(n int) {
    	h.HP = max(0, h.HP-n)
    }

    // Heal raises HP by n, never above MaxHP.
    func (h Hero) Heal(n int) {
    	h.HP = min(h.MaxHP, h.HP+n)
    }

    // GainXP adds n experience. Every 100 XP is traded in for one level.
    func (h Hero) GainXP(n int) {
    	h.XP += n
    	for h.XP >= 100 {
    		h.XP -= 100
    		h.Level++
    	}
    }

    // IsAlive reports whether the hero is still standing.
    func (h Hero) IsAlive() bool {
    	return h.HP > 0
    }

    func main() {
    	aria := Hero{Name: "Aria", HP: 100, MaxHP: 100, Level: 1}
    	aria.TakeDamage(70)
    	aria.Heal(20)
    	aria.GainXP(250)
    	fmt.Printf("%+v alive=%v\n", aria, aria.IsAlive())
    	// want {Name:Aria HP:50 MaxHP:100 XP:50 Level:3} alive=true
    }
  solution: |
    package main

    import "fmt"

    type Hero struct {
    	Name  string
    	HP    int
    	MaxHP int
    	XP    int
    	Level int
    }

    // TakeDamage lowers HP by n, never below 0.
    func (h *Hero) TakeDamage(n int) {
    	h.HP = max(0, h.HP-n)
    }

    // Heal raises HP by n, never above MaxHP.
    func (h *Hero) Heal(n int) {
    	h.HP = min(h.MaxHP, h.HP+n)
    }

    // GainXP adds n experience. Every 100 XP is traded in for one level.
    func (h *Hero) GainXP(n int) {
    	h.XP += n
    	for h.XP >= 100 {
    		h.XP -= 100
    		h.Level++
    	}
    }

    // IsAlive reports whether the hero is still standing.
    func (h *Hero) IsAlive() bool {
    	return h.HP > 0
    }

    func main() {
    	aria := Hero{Name: "Aria", HP: 100, MaxHP: 100, Level: 1}
    	aria.TakeDamage(70)
    	aria.Heal(20)
    	aria.GainXP(250)
    	fmt.Printf("%+v alive=%v\n", aria, aria.IsAlive())
    }
  tests: |
    package main

    import "testing"

    func TestTakeDamageAndHeal(t *testing.T) {
    	h := Hero{Name: "Aria", HP: 100, MaxHP: 100, Level: 1}
    	h.TakeDamage(70)
    	if h.HP != 30 {
    		t.Fatalf("after TakeDamage(70) from 100 HP: HP = %d, want 30 (did the change stick?)", h.HP)
    	}
    	h.Heal(500)
    	if h.HP != 100 {
    		t.Errorf("after Heal(500): HP = %d, want 100 (capped at MaxHP)", h.HP)
    	}
    	h.TakeDamage(250)
    	if h.HP != 0 || h.IsAlive() {
    		t.Errorf("after TakeDamage(250): HP = %d, IsAlive() = %v, want 0 and false", h.HP, h.IsAlive())
    	}
    }

    func TestGainXP(t *testing.T) {
    	h := Hero{Name: "Borin", HP: 50, MaxHP: 50, Level: 1}
    	h.GainXP(250)
    	if h.Level != 3 || h.XP != 50 {
    		t.Errorf("after GainXP(250) at level 1: Level = %d, XP = %d, want Level 3, XP 50", h.Level, h.XP)
    	}
    }

    type damageable interface{ TakeDamage(int) }

    func TestThroughInterface(t *testing.T) {
    	h := &Hero{Name: "Cyra", HP: 40, MaxHP: 40}
    	var d damageable = h
    	d.TakeDamage(15)
    	if h.HP != 25 {
    		t.Errorf("TakeDamage(15) called through an interface: HP = %d, want 25", h.HP)
    	}
    }

    func TestConsistentReceivers(t *testing.T) {
    	// If every method has a pointer receiver, a plain Hero value has no methods
    	// in its method set, so it can't satisfy even a one-method interface.
    	var v any = Hero{}
    	if _, ok := v.(interface{ IsAlive() bool }); ok {
    		t.Error("IsAlive still has a value receiver: give every Hero method a pointer receiver, to be consistent")
    	}
    }
---

Our hero needs to take damage. Let's try the obvious thing:

```go
func (h Hero) TakeDamage(n int) {
	h.HP -= n
}
```

It compiles. It even runs. And the hero never gets hurt.

## Value receivers get a copy

Remember: assigning or passing a struct copies it. A **value receiver** `(h Hero)` is no different. The method receives its own private copy of the hero, dents *that*, and throws it away.

A **pointer receiver** `(h *Hero)` receives the address of the original, so changes stick:

```go
package main

import "fmt"

type Hero struct {
	Name string
	HP   int
}

func (h Hero) TakeDamageBroken(n int) {
	h.HP -= n // changes a copy
}

func (h *Hero) TakeDamage(n int) {
	h.HP -= n // changes the real hero
}

func main() {
	aria := Hero{Name: "Aria", HP: 100}

	aria.TakeDamageBroken(30)
	fmt.Println(aria.HP)

	aria.TakeDamage(30)
	fmt.Println(aria.HP)
}
```

```
100
70
```

Notice we called `aria.TakeDamage(30)` even though `aria` is a `Hero`, not a `*Hero`. Go automatically takes the address for you (`(&aria).TakeDamage(30)`) because `aria` is a variable. It works the other way too: you can call a value method on a pointer, and Go dereferences it.

Python has no equivalent choice, since `self` always refers to the shared object. Go makes you decide, which feels fussy at first but makes it obvious which methods can change things.

## When to use which

Use a **pointer receiver** when:

- the method **modifies** the receiver (`TakeDamage`, `LevelUp`),
- the struct is **large**, so copying it on every call is wasteful,
- the struct contains something that **must not be copied**, like a `sync.Mutex`.

Use a **value receiver** for small, immutable-feeling types: a `Point{X, Y}`, a `Gold` amount, a `Color`.

And the most important rule: **be consistent**. If one method on `Hero` needs a pointer receiver, give them all pointer receivers. Mixing the two is legal but confusing, and it causes the method set surprise below.

## Method sets

Every type has a **method set**: the methods you can call on a value of that type when it's stored in an interface.

| Type    | Method set contains                                  |
|---------|------------------------------------------------------|
| `Hero`  | methods with value receivers `(h Hero)`              |
| `*Hero` | methods with value receivers **and** pointer receivers |

Why the asymmetry? A pointer can always be dereferenced to get a value, but a value stored inside an interface has no stable address to point to. So Go refuses to pretend.

This matters as soon as interfaces show up (chapter 4 onwards):

```go
type Damageable interface {
	TakeDamage(n int)
}

var d Damageable = &Hero{Name: "Aria", HP: 100} // OK
var e Damageable = Hero{Name: "Aria", HP: 100}  // compile error
```

The second line fails with an error like *"Hero does not implement Damageable (method TakeDamage has pointer receiver)"*. When you see that, the fix is almost always to use `&Hero{...}`, or a constructor that returns `*Hero`.

## Another place the auto-address fails

Go can only take the address of something *addressable*, such as a variable, a slice element or a struct field. Map values are not addressable:

```go
heroes := map[string]Hero{"aria": {Name: "Aria", HP: 100}}
heroes["aria"].TakeDamage(10) // compile error: cannot call pointer method TakeDamage on Hero
```

Store pointers instead (`map[string]*Hero`) when you want to call mutating methods on map entries.

## Assignment

Aria's hero sheet is broken: she takes damage, heals and gains experience, yet `main` prints her exactly as she started. Every method on `Hero` was written with a value receiver, so each one changes a copy.

Fix the receivers so that `TakeDamage`, `Heal` and `GainXP` change the real hero. Then follow the consistency rule: once some methods need a pointer receiver, give **all** of `Hero`'s methods one, including `IsAlive`. The method bodies are already correct; only the receivers need to change.

## Further reading

- [A Tour of Go: Pointer receivers](https://go.dev/tour/methods/4) and [Choosing a value or pointer receiver](https://go.dev/tour/methods/8)
- [Effective Go: Pointers vs. Values](https://go.dev/doc/effective_go#pointers_vs_values)
