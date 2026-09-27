---
title: Promoted Methods
quiz:
  - question: |
      What does this print?

      ```go
      type Health struct{ HP int }

      func (h *Health) TakeDamage(n int) { h.HP -= n }

      type Golem struct{ Health }

      func (g *Golem) TakeDamage(n int) {
      	g.Health.TakeDamage(n / 2) // stone skin
      }

      func main() {
      	g := &Golem{Health{HP: 100}}
      	g.TakeDamage(40)
      	g.Health.TakeDamage(40)
      	fmt.Println(g.HP)
      }
      ```
    options:
      - text: '`20`'
      - text: '`40`'
        correct: true
      - text: '`60`'
      - text: '`80`'
    explanation: |
      `g.TakeDamage(40)` calls the Golem's own method, which halves it: 100 - 20 = 80.
      `g.Health.TakeDamage(40)` calls the embedded method directly, skipping the
      stone skin: 80 - 40 = 40.
  - question: |
      `Hero` embeds both `Walker` and `Swimmer`, and both have a `Move()` method. `Hero` has no `Move` of its own. What happens?
    options:
      - text: '`Walker.Move` wins because it''s listed first'
      - text: Compile error as soon as the struct is declared
      - text: Compile error only if you call `h.Move()`, because the selector is ambiguous
        correct: true
      - text: Both methods run, one after the other
    explanation: |
      Go doesn't pick a winner between embedded types at the same depth. The
      struct compiles fine, but `h.Move()` is ambiguous and won't compile. You can
      still call `h.Walker.Move()`, or define `Hero.Move` to decide.
exercise:
  starter: |
    package main

    import "fmt"

    type Health struct {
    	HP, MaxHP int
    }

    func (h *Health) TakeDamage(n int) { h.HP = max(0, h.HP-n) }
    func (h *Health) Heal(n int)       { h.HP = min(h.MaxHP, h.HP+n) }
    func (h *Health) Current() int     { return h.HP }

    // Knight should embed Health so it gets TakeDamage, Heal and Current for free.
    type Knight struct {
    	Name string
    }

    // NewKnight returns a knight at full health.
    func NewKnight(name string, hp int) *Knight {
    	return &Knight{Name: name}
    }

    // Golem embeds Health, but its stone skin halves all damage (rounding down).
    type Golem struct {
    	Health
    }

    func NewGolem(hp int) *Golem {
    	return &Golem{Health: Health{HP: hp, MaxHP: hp}}
    }

    func main() {
    	g := NewGolem(100)
    	g.TakeDamage(40)
    	fmt.Println("golem HP after a 40 damage hit:", g.Current()) // want 80

    	var k any = NewKnight("Sir Aria", 120)
    	_, ok := k.(interface{ TakeDamage(int) })
    	fmt.Println("knight can take damage:", ok) // want true
    }
  solution: |
    package main

    import "fmt"

    type Health struct {
    	HP, MaxHP int
    }

    func (h *Health) TakeDamage(n int) { h.HP = max(0, h.HP-n) }
    func (h *Health) Heal(n int)       { h.HP = min(h.MaxHP, h.HP+n) }
    func (h *Health) Current() int     { return h.HP }

    type Knight struct {
    	Health
    	Name string
    }

    func NewKnight(name string, hp int) *Knight {
    	return &Knight{Health: Health{HP: hp, MaxHP: hp}, Name: name}
    }

    type Golem struct {
    	Health
    }

    func NewGolem(hp int) *Golem {
    	return &Golem{Health: Health{HP: hp, MaxHP: hp}}
    }

    func (g *Golem) TakeDamage(n int) {
    	g.Health.TakeDamage(n / 2)
    }

    func main() {
    	g := NewGolem(100)
    	g.TakeDamage(40)
    	fmt.Println("golem HP after a 40 damage hit:", g.Current())

    	var k any = NewKnight("Sir Aria", 120)
    	_, ok := k.(interface{ TakeDamage(int) })
    	fmt.Println("knight can take damage:", ok)
    }
  tests: |
    package main

    import "testing"

    type fighter interface {
    	TakeDamage(int)
    	Heal(int)
    	Current() int
    }

    func TestKnightEmbedsHealth(t *testing.T) {
    	var v any = NewKnight("Sir Aria", 120)
    	k, ok := v.(fighter)
    	if !ok {
    		t.Fatal("*Knight has no TakeDamage/Heal/Current methods: embed Health in Knight")
    	}
    	if got := k.Current(); got != 120 {
    		t.Fatalf("NewKnight(\"Sir Aria\", 120).Current() = %d, want 120 (start at full health)", got)
    	}
    	k.TakeDamage(50)
    	if got := k.Current(); got != 70 {
    		t.Errorf("after TakeDamage(50): Current() = %d, want 70", got)
    	}
    	k.Heal(500)
    	if got := k.Current(); got != 120 {
    		t.Errorf("after Heal(500): Current() = %d, want 120 (capped at max HP)", got)
    	}
    	if name := NewKnight("Sir Aria", 120).Name; name != "Sir Aria" {
    		t.Errorf("NewKnight name = %q, want %q", name, "Sir Aria")
    	}
    }

    func TestGolemHalvesDamage(t *testing.T) {
    	g := NewGolem(100)
    	g.TakeDamage(40)
    	if got := g.Current(); got != 80 {
    		t.Errorf("golem with 100 HP after TakeDamage(40): Current() = %d, want 80", got)
    	}
    	g.TakeDamage(15)
    	if got := g.Current(); got != 73 {
    		t.Errorf("golem with 80 HP after TakeDamage(15): Current() = %d, want 73 (15/2 rounds down to 7)", got)
    	}
    	var f fighter = g
    	f.TakeDamage(10)
    	if got := g.Current(); got != 68 {
    		t.Errorf("golem used through an interface after TakeDamage(10): Current() = %d, want 68", got)
    	}
    }
---

Embedding promotes **methods** as well as fields. This is where it really starts to feel like code reuse.

## Methods come along for free

```go
package main

import "fmt"

type Health struct {
	HP, MaxHP int
}

func (h *Health) TakeDamage(n int) { h.HP = max(0, h.HP-n) }
func (h *Health) Heal(n int)       { h.HP = min(h.MaxHP, h.HP+n) }
func (h *Health) Alive() bool      { return h.HP > 0 }

type Hero struct {
	Health
	Name string
}

type Goblin struct {
	Health
	Loot string
}

func main() {
	aria := &Hero{Health: Health{HP: 100, MaxHP: 100}, Name: "Aria"}
	snik := &Goblin{Health: Health{HP: 25, MaxHP: 25}, Loot: "rusty key"}

	aria.TakeDamage(40)
	aria.Heal(15)
	snik.TakeDamage(30)

	fmt.Println(aria.Name, aria.HP, aria.Alive())
	fmt.Println(snik.Loot, snik.HP, snik.Alive())
}
```

```
Aria 75 true
rusty key 0 false
```

`Hero` and `Goblin` each got `TakeDamage`, `Heal` and `Alive` without writing a line. `aria.TakeDamage(40)` is shorthand for `aria.Health.TakeDamage(40)`. The receiver is the embedded `Health`, not the `Hero`. Hold on to that detail; it matters a lot in the last lesson of this chapter.

## "Overriding" by shadowing

If the outer type declares a method with the same name, it **shadows** the promoted one. Callers get the outer version, and the outer version can still call the inner one explicitly, which feels a lot like `super()`:

```go
type Golem struct {
	Health
}

// Golems have stone skin: they take half damage.
func (g *Golem) TakeDamage(n int) {
	g.Health.TakeDamage(n / 2) // explicitly call the embedded version
}
```

Now `golem.TakeDamage(40)` only removes 20 HP. There's no `super` keyword; you name the embedded field. It's explicit, which makes it easy to see exactly which code runs.

## Method sets with embedding

Promotion follows the method set rules from chapter 2:

- Embed a **value** `Health`, whose methods have pointer receivers. Then `*Hero` gets those methods, but a plain `Hero` value stored in an interface doesn't.
- Embed a **pointer** `*Health`. Then both `Hero` and `*Hero` get them.

In practice: if the embedded type has pointer-receiver methods, use the outer type through a pointer too (`&Hero{...}`, constructors returning `*Hero`), and everything lines up.

## Ambiguity is a compile error

What if you embed two types that both have a method with the same name?

```go
type Walker struct{}

func (Walker) Move() string { return "walks" }

type Swimmer struct{}

func (Swimmer) Move() string { return "swims" }

type Amphibian struct {
	Walker
	Swimmer
}
```

Python would pick one using its method resolution order, and you'd need to know the rules to predict which. Go refuses to guess. The struct compiles, but `a.Move()` is a compile error: *ambiguous selector a.Move*. You resolve it by being explicit:

```go
func (a Amphibian) Move() string {
	return a.Walker.Move() + " and " + a.Swimmer.Move()
}
```

Clear, not clever. Anyone reading `Amphibian.Move` knows exactly what happens.

## Nested embedding

Embedding can nest: if `Health` itself embedded a `Shield`, then `Hero` would get `Shield`'s methods too. The shallowest match wins, and ties at the same depth are ambiguous. Keep nesting shallow. Two levels of embedding is usually the most a reader can hold in their head.

## Assignment

- Embed `Health` in `Knight`, and make `NewKnight` start the knight at full health (`HP` and `MaxHP` both `hp`). The knight should then get `TakeDamage`, `Heal` and `Current` for free.
- Give `Golem` its own `TakeDamage` method that halves the damage (integer division rounds down) and then calls the embedded `Health.TakeDamage`.

## Further reading

- [Effective Go: Embedding](https://go.dev/doc/effective_go#embedding)
