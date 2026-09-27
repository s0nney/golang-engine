---
title: Wyvern Lore
difficulty: easy
after: inheritance-vs-composition
hints:
  - 'Because `Wyvern` embeds `Creature`, `w.Name`, `w.HP` and `w.Hurt(...)` already work on a wyvern. You only write the methods that are **new** or **different**.'
  - 'To reuse the embedded method inside your own `Describe`, call it through the field name: `w.Creature.Describe()`. Calling `w.Describe()` there would call itself forever.'
  - '`Bite` just calls `target.Hurt(w.Venom)`. A `*Wyvern` counts as a `Hurter` because `Hurt` is promoted from the embedded `Creature`.'
exercise:
  starter: |
    package main

    import "fmt"

    type Creature struct {
    	Name string
    	HP   int
    }

    // Describe returns e.g. "Snik (30 HP)".
    func (c Creature) Describe() string {
    	return "?"
    }

    // Hurt lowers HP by n, but never below 0.
    func (c *Creature) Hurt(n int) {
    	c.HP = max(0, c.HP-n)
    }

    // Hurter is anything that can take damage.
    type Hurter interface {
    	Hurt(n int)
    }

    // Wyvern is a Creature with a venomous sting.
    type Wyvern struct {
    	Creature
    	Venom int
    }

    // TODO: give Wyvern its own Describe that returns e.g. "Vex (40 HP), venom 7".
    // It must reuse Creature's Describe instead of repeating its format.

    // Bite hurts target by the wyvern's venom.
    func (w *Wyvern) Bite(target Hurter) {
    	// ?
    }

    func main() {
    	vex := &Wyvern{Name: "Vex", HP: 40, Venom: 7} // Go 1.27: promoted fields in a literal
    	snik := &Creature{Name: "Snik", HP: 30}
    	vex.Hurt(5)                  // promoted from Creature
    	fmt.Println(vex.Describe())  // want Vex (35 HP), venom 7
    	fmt.Println(snik.Describe()) // want Snik (30 HP)
    }
  solution: |
    package main

    import "fmt"

    type Creature struct {
    	Name string
    	HP   int
    }

    // Describe returns e.g. "Snik (30 HP)".
    func (c Creature) Describe() string {
    	return fmt.Sprintf("%s (%d HP)", c.Name, c.HP)
    }

    // Hurt lowers HP by n, but never below 0.
    func (c *Creature) Hurt(n int) {
    	c.HP = max(0, c.HP-n)
    }

    // Hurter is anything that can take damage.
    type Hurter interface {
    	Hurt(n int)
    }

    // Wyvern is a Creature with a venomous sting.
    type Wyvern struct {
    	Creature
    	Venom int
    }

    // Describe returns e.g. "Vex (40 HP), venom 7".
    func (w Wyvern) Describe() string {
    	return fmt.Sprintf("%s, venom %d", w.Creature.Describe(), w.Venom)
    }

    // Bite hurts target by the wyvern's venom.
    func (w *Wyvern) Bite(target Hurter) {
    	target.Hurt(w.Venom)
    }

    func main() {
    	vex := &Wyvern{Name: "Vex", HP: 40, Venom: 7}
    	snik := &Creature{Name: "Snik", HP: 30}
    	vex.Hurt(5)
    	vex.Bite(snik)
    	fmt.Println(vex.Describe())
    	fmt.Println(snik.Describe())
    }
  tests: |
    package main

    import "testing"

    func TestCreatureDescribe(t *testing.T) {
    	for _, tt := range []struct {
    		c    Creature
    		want string
    	}{
    		{Creature{Name: "Snik", HP: 30}, "Snik (30 HP)"},
    		{Creature{Name: "Old Bones", HP: 0}, "Old Bones (0 HP)"},
    	} {
    		if got := tt.c.Describe(); got != tt.want {
    			t.Errorf("%#v.Describe() = %q, want %q", tt.c, got, tt.want)
    		}
    	}
    }

    type describer interface{ Describe() string }

    func TestWyvernDescribe(t *testing.T) {
    	var d describer = Wyvern{Name: "Vex", HP: 40, Venom: 7}
    	if got, want := d.Describe(), "Vex (40 HP), venom 7"; got != want {
    		t.Errorf("Wyvern{Name: \"Vex\", HP: 40, Venom: 7}.Describe() = %q, want %q", got, want)
    	}
    	w := Wyvern{Name: "Zix", HP: 3}
    	if got, want := w.Describe(), "Zix (3 HP), venom 0"; got != want {
    		t.Errorf("Wyvern{Name: \"Zix\", HP: 3}.Describe() = %q, want %q", got, want)
    	}
    	if got, want := w.Creature.Describe(), "Zix (3 HP)"; got != want {
    		t.Errorf("w.Creature.Describe() = %q, want %q: the embedded Creature keeps its own Describe", got, want)
    	}
    }

    func TestBite(t *testing.T) {
    	vex := &Wyvern{Name: "Vex", HP: 40, Venom: 7}
    	snik := &Creature{Name: "Snik", HP: 30}
    	vex.Bite(snik)
    	vex.Bite(snik)
    	if snik.HP != 16 {
    		t.Errorf("after two bites with venom 7, Snik (30 HP) has %d HP, want 16", snik.HP)
    	}

    	rival := &Wyvern{Name: "Kaz", HP: 10, Venom: 2}
    	vex.Bite(rival) // a *Wyvern is a Hurter too, via the promoted Hurt
    	rival.Bite(vex)
    	if rival.HP != 3 || vex.HP != 38 {
    		t.Errorf("after Vex and Kaz bite each other, HP = %d and %d, want 38 and 3", vex.HP, rival.HP)
    	}
    	vex.Bite(rival)
    	if rival.HP != 0 {
    		t.Errorf("a bite for more than the target's HP left it at %d HP, want 0", rival.HP)
    	}
    	if vex.Describe() != "Vex (38 HP), venom 7" {
    		t.Errorf("vex.Describe() = %q, want %q", vex.Describe(), "Vex (38 HP), venom 7")
    	}
    }

    type dummy struct{ hits []int }

    func (d *dummy) Hurt(n int) { d.hits = append(d.hits, n) }

    func TestBiteUsesHurter(t *testing.T) {
    	d := &dummy{}
    	(&Wyvern{Venom: 4}).Bite(d)
    	if len(d.hits) != 1 || d.hits[0] != 4 {
    		t.Errorf("biting a training dummy recorded hits %v, want [4]: Bite should call target.Hurt(w.Venom) once", d.hits)
    	}
    }
---

The bestiary describes every creature the same way, and wyverns are creatures
with a venomous sting. `Wyvern` already **embeds** `Creature`, so it gets `Name`,
`HP` and `Hurt` for free. Finish the job:

1. Write `Creature.Describe`, returning `"<Name> (<HP> HP)"`.
2. Give `Wyvern` its **own** `Describe` (value receiver) returning
   `"<Name> (<HP> HP), venom <Venom>"`. Reuse the creature's description
   instead of repeating its format.
3. Write `(*Wyvern).Bite(target Hurter)`: it hurts `target` by the wyvern's
   `Venom`. Any `Hurter` can be bitten, including another wyvern.

Since Go 1.27 you can set promoted fields directly in a struct literal, so
`Wyvern{Name: "Vex", HP: 40, Venom: 7}` works. The tests build wyverns that way.

## Example

```go
vex := &Wyvern{Name: "Vex", HP: 40, Venom: 7}
snik := &Creature{Name: "Snik", HP: 30}

vex.Hurt(5)             // promoted from Creature
vex.Bite(snik)
vex.Describe()          // "Vex (35 HP), venom 7"
snik.Describe()         // "Snik (23 HP)"
vex.Creature.Describe() // "Vex (35 HP)"
```

## Constraints

- Keep `Creature` embedded in `Wyvern` (no named field), and don't change
  `Creature.Hurt`.
