---
title: Warband
difficulty: hard
after: inheritance-vs-composition
hints:
  - '`Knight` gets `Alive`, `Heal` **and** `TakeDamage` promoted from `Vitals`, so it already satisfies `Unit`, but the promoted `TakeDamage` ignores the armor. Declare `(*Knight).TakeDamage` yourself: the outer method shadows the promoted one, and inside it you can still call `k.Vitals.TakeDamage(k.Absorb(n))`.'
  - '`Warband` should store its members as a `[]Unit` (copied from the variadic argument, so the caller can''t change it later). Because `*Warband` has the `Unit` methods too, a warband can be a member of another warband and all the recursion happens through the interface for free.'
  - 'For `Heal`, loop over the members and ask each one `h, ok := m.(Healable)`. Knights, nested warbands and any test type with a `Heal` method pass; wraiths don''t. Dead units stay dead: make `Vitals.Heal` return 0 when `HP` is 0.'
exercise:
  starter: |
    package main

    import "fmt"

    // Unit is anything that can fight: a single creature or a whole warband.
    type Unit interface {
    	Name() string
    	Alive() bool
    	// TakeDamage applies an attack of strength n and returns the HP lost.
    	TakeDamage(n int) int
    }

    // Healable is implemented by units that can be healed.
    type Healable interface {
    	// Heal restores up to n HP and returns how much was restored.
    	Heal(n int) int
    }

    // Vitals is a reusable health component.
    type Vitals struct {
    	HP, MaxHP int
    }

    func (v *Vitals) Alive() bool { return false }

    func (v *Vitals) TakeDamage(n int) int { return 0 }

    func (v *Vitals) Heal(n int) int { return 0 }

    // Armor is a reusable damage-reduction component.
    type Armor struct {
    	Defense int
    }

    // Absorb returns the part of an attack of strength n that gets through.
    func (a Armor) Absorb(n int) int { return n }

    // Knight is built from Vitals and Armor.
    type Knight struct {
    	Title string
    	Vitals
    	Armor
    }

    func (k *Knight) Name() string { return k.Title }

    // Wraith has no body to heal and shrugs off weak hits.
    type Wraith struct {
    	Title   string
    	Essence int
    }

    func (w *Wraith) Name() string         { return w.Title }
    func (w *Wraith) Alive() bool          { return false }
    func (w *Wraith) TakeDamage(n int) int { return 0 }

    // Warband is a Unit made of other Units, possibly other warbands.
    type Warband struct {
    }

    func NewWarband(name string, members ...Unit) *Warband {
    	return &Warband{}
    }

    func (w *Warband) Name() string         { return "" }
    func (w *Warband) Living() int          { return 0 }
    func (w *Warband) Alive() bool          { return false }
    func (w *Warband) TakeDamage(n int) int { return 0 }
    func (w *Warband) Heal(n int) int       { return 0 }

    func main() {
    	ayla := &Knight{Title: "Ser Ayla", HP: 30, MaxHP: 30, Defense: 3}
    	ghost := &Wraith{Title: "the Pale One", Essence: 12}
    	band := NewWarband("Ashen Company", ayla, ghost)

    	fmt.Println(band.TakeDamage(10), ayla.HP) // want 7 23
    	fmt.Println(band.Heal(5), ayla.HP)        // want 5 28
    	band.TakeDamage(40)
    	fmt.Println(band.Living(), band.TakeDamage(4), band.TakeDamage(20), band.Alive()) // want 1 0 12 false
    }
  solution: |
    package main

    import "fmt"

    // Unit is anything that can fight: a single creature or a whole warband.
    type Unit interface {
    	Name() string
    	Alive() bool
    	// TakeDamage applies an attack of strength n and returns the HP lost.
    	TakeDamage(n int) int
    }

    // Healable is implemented by units that can be healed.
    type Healable interface {
    	// Heal restores up to n HP and returns how much was restored.
    	Heal(n int) int
    }

    // Vitals is a reusable health component.
    type Vitals struct {
    	HP, MaxHP int
    }

    func (v *Vitals) Alive() bool { return v.HP > 0 }

    func (v *Vitals) TakeDamage(n int) int {
    	lost := max(0, min(n, v.HP))
    	v.HP -= lost
    	return lost
    }

    func (v *Vitals) Heal(n int) int {
    	if !v.Alive() || n <= 0 {
    		return 0
    	}
    	healed := min(n, v.MaxHP-v.HP)
    	v.HP += healed
    	return healed
    }

    // Armor is a reusable damage-reduction component.
    type Armor struct {
    	Defense int
    }

    // Absorb returns the part of an attack of strength n that gets through.
    func (a Armor) Absorb(n int) int { return max(0, n-a.Defense) }

    // Knight is built from Vitals and Armor.
    type Knight struct {
    	Title string
    	Vitals
    	Armor
    }

    func (k *Knight) Name() string { return k.Title }

    // TakeDamage lets the armor absorb its share before the vitals take the rest.
    func (k *Knight) TakeDamage(n int) int {
    	return k.Vitals.TakeDamage(k.Absorb(n))
    }

    // Wraith has no body to heal and shrugs off weak hits.
    type Wraith struct {
    	Title   string
    	Essence int
    }

    func (w *Wraith) Name() string { return w.Title }
    func (w *Wraith) Alive() bool  { return w.Essence > 0 }

    func (w *Wraith) TakeDamage(n int) int {
    	if n < 5 {
    		return 0
    	}
    	lost := min(n, w.Essence)
    	w.Essence -= lost
    	return lost
    }

    // Warband is a Unit made of other Units, possibly other warbands.
    type Warband struct {
    	name    string
    	members []Unit
    }

    // Compile-time checks that the design fits together.
    var (
    	_ Unit     = (*Knight)(nil)
    	_ Healable = (*Knight)(nil)
    	_ Unit     = (*Wraith)(nil)
    	_ Unit     = (*Warband)(nil)
    	_ Healable = (*Warband)(nil)
    )

    func NewWarband(name string, members ...Unit) *Warband {
    	return &Warband{name: name, members: append([]Unit(nil), members...)}
    }

    func (w *Warband) Name() string { return w.name }

    // Living returns how many direct members are alive.
    func (w *Warband) Living() int {
    	n := 0
    	for _, m := range w.members {
    		if m.Alive() {
    			n++
    		}
    	}
    	return n
    }

    func (w *Warband) Alive() bool { return w.Living() > 0 }

    // TakeDamage sends the whole attack at the first living member.
    func (w *Warband) TakeDamage(n int) int {
    	for _, m := range w.members {
    		if m.Alive() {
    			return m.TakeDamage(n)
    		}
    	}
    	return 0
    }

    // Heal offers n HP to every member that can be healed and returns the total restored.
    func (w *Warband) Heal(n int) int {
    	total := 0
    	for _, m := range w.members {
    		if h, ok := m.(Healable); ok {
    			total += h.Heal(n)
    		}
    	}
    	return total
    }

    func main() {
    	ayla := &Knight{Title: "Ser Ayla", HP: 30, MaxHP: 30, Defense: 3}
    	ghost := &Wraith{Title: "the Pale One", Essence: 12}
    	band := NewWarband("Ashen Company", ayla, ghost)

    	fmt.Println(band.TakeDamage(10), ayla.HP) // 7 23
    	fmt.Println(band.Heal(5), ayla.HP)        // 5 28
    	band.TakeDamage(40)
    	fmt.Println(band.Living(), band.TakeDamage(4), band.TakeDamage(20), band.Alive()) // 1 0 12 false
    }
  tests: |
    package main

    import "testing"

    func knight(title string, hp, def int) *Knight {
    	// Go 1.27: promoted fields of Vitals and Armor set directly.
    	return &Knight{Title: title, HP: hp, MaxHP: hp, Defense: def}
    }

    func TestVitals(t *testing.T) {
    	v := &Vitals{HP: 20, MaxHP: 25}
    	steps := []struct {
    		op         string
    		n, ret, hp int
    	}{
    		{"TakeDamage", 8, 8, 12},
    		{"TakeDamage", -3, 0, 12},
    		{"Heal", 100, 13, 25},
    		{"Heal", -5, 0, 25},
    		{"TakeDamage", 30, 25, 0},
    		{"Heal", 10, 0, 0}, // the dead stay dead
    		{"TakeDamage", 5, 0, 0},
    	}
    	for i, s := range steps {
    		var got int
    		if s.op == "Heal" {
    			got = v.Heal(s.n)
    		} else {
    			got = v.TakeDamage(s.n)
    		}
    		if got != s.ret || v.HP != s.hp {
    			t.Fatalf("step %d, Vitals.%s(%d): returned %d with HP %d, want %d with HP %d", i+1, s.op, s.n, got, v.HP, s.ret, s.hp)
    		}
    	}
    	if v.Alive() {
    		t.Errorf("Vitals with 0 HP: Alive() = true, want false")
    	}
    	if !(&Vitals{HP: 1, MaxHP: 1}).Alive() {
    		t.Errorf("Vitals with 1 HP: Alive() = false, want true")
    	}
    }

    func TestKnightArmor(t *testing.T) {
    	tests := []struct {
    		def, hit, lost, hp int
    	}{
    		{3, 10, 7, 23},
    		{3, 3, 0, 30},
    		{3, 1, 0, 30},
    		{0, 10, 10, 20},
    		{5, 100, 30, 0},
    		{3, -8, 0, 30},
    	}
    	for _, tt := range tests {
    		k := knight("Ser Ayla", 30, tt.def)
    		var u Unit = k // through the interface, as a Warband would call it
    		if got := u.TakeDamage(tt.hit); got != tt.lost || k.HP != tt.hp {
    			t.Errorf("knight with 30 HP and Defense %d hit for %d: lost %d, HP now %d; want lost %d, HP %d",
    				tt.def, tt.hit, got, k.HP, tt.lost, tt.hp)
    		}
    	}
    	k := knight("Ser Bram", 30, 4)
    	k.Vitals.TakeDamage(10)
    	if k.HP != 20 {
    		t.Errorf("k.Vitals.TakeDamage(10) should bypass the armor: HP = %d, want 20", k.HP)
    	}
    	if k.Name() != "Ser Bram" || !k.Alive() {
    		t.Errorf("knight Name() = %q, Alive() = %v, want \"Ser Bram\", true", k.Name(), k.Alive())
    	}
    	var h Healable = k
    	if got := h.Heal(15); got != 10 || k.HP != 30 {
    		t.Errorf("knight at 20/30 HP healed 15: restored %d, HP %d; want 10, 30", got, k.HP)
    	}
    }

    func TestWraith(t *testing.T) {
    	w := &Wraith{Title: "the Pale One", Essence: 12}
    	for _, s := range []struct{ hit, lost, left int }{{4, 0, 12}, {5, 5, 7}, {0, 0, 7}, {20, 7, 0}, {9, 0, 0}} {
    		if got := w.TakeDamage(s.hit); got != s.lost || w.Essence != s.left {
    			t.Fatalf("wraith hit for %d: lost %d, Essence %d; want lost %d, Essence %d", s.hit, got, w.Essence, s.lost, s.left)
    		}
    	}
    	if w.Alive() || w.Name() != "the Pale One" {
    		t.Errorf("wraith with 0 Essence: Name() = %q, Alive() = %v, want \"the Pale One\", false", w.Name(), w.Alive())
    	}
    	if _, ok := any(w).(Healable); ok {
    		t.Errorf("*Wraith must not be Healable: wraiths can't be healed")
    	}
    }

    func TestWarband(t *testing.T) {
    	ayla := knight("Ser Ayla", 30, 3)
    	ghost := &Wraith{Title: "the Pale One", Essence: 12}
    	bram := knight("Ser Bram", 10, 0)
    	band := NewWarband("Ashen Company", ayla, ghost, bram)

    	if band.Name() != "Ashen Company" || band.Living() != 3 || !band.Alive() {
    		t.Fatalf("new warband: Name() = %q, Living() = %d, Alive() = %v, want \"Ashen Company\", 3, true", band.Name(), band.Living(), band.Alive())
    	}
    	if got := band.TakeDamage(10); got != 7 || ayla.HP != 23 || ghost.Essence != 12 || bram.HP != 10 {
    		t.Errorf("warband hit for 10: lost %d, HPs %d/%d/%d, want 7 lost, only the first member hit (23/12/10)", got, ayla.HP, ghost.Essence, bram.HP)
    	}
    	if got := band.Heal(5); got != 5 || ayla.HP != 28 {
    		t.Errorf("warband healed 5 with Ayla at 23/30 and Bram full: restored %d, Ayla %d HP, want 5, 28", got, ayla.HP)
    	}
    	band.TakeDamage(50) // Ayla falls; no spill-over to the others
    	if ayla.HP != 0 || ghost.Essence != 12 || band.Living() != 2 {
    		t.Errorf("after a hit for 50: HPs %d/%d/%d, Living() = %d, want 0/12/10 and 2", ayla.HP, ghost.Essence, bram.HP, band.Living())
    	}
    	if got := band.TakeDamage(4); got != 0 || ghost.Essence != 12 || bram.HP != 10 {
    		t.Errorf("hit for 4 with the wraith in front: lost %d, want 0 (and Bram untouched)", got)
    	}
    	bram.TakeDamage(3)
    	if got := band.Heal(10); got != 3 || ayla.HP != 0 {
    		t.Errorf("Heal(10) with Ayla dead, a wraith, and Bram at 7/10: restored %d (Ayla %d HP), want 3 (Ayla stays at 0)", got, ayla.HP)
    	}
    	band.TakeDamage(12)
    	if got := band.TakeDamage(15); got != 10 || band.Alive() {
    		t.Errorf("last knight hit for 15: lost %d, Alive() = %v, want 10, false", got, band.Alive())
    	}
    	if got := band.TakeDamage(15); got != 0 {
    		t.Errorf("hitting a fallen warband lost %d HP, want 0", got)
    	}
    }

    func TestEmptyWarband(t *testing.T) {
    	band := NewWarband("Nobody")
    	if band.Alive() || band.Living() != 0 || band.TakeDamage(10) != 0 || band.Heal(10) != 0 {
    		t.Errorf("empty warband: Alive() = %v, Living() = %d, want false and 0, and no damage or healing", band.Alive(), band.Living())
    	}
    }

    func TestNestedWarbands(t *testing.T) {
    	a, b, c := knight("A", 10, 0), knight("B", 10, 0), knight("C", 10, 0)
    	vanguard := NewWarband("Vanguard", a, b)
    	host := NewWarband("Host", vanguard, c)

    	var u Unit = host
    	u.TakeDamage(10)
    	u.TakeDamage(4)
    	if a.HP != 0 || b.HP != 6 || c.HP != 10 {
    		t.Fatalf("host = [vanguard = [A, B], C], hits 10 then 4: HPs %d/%d/%d, want 0/6/10 (the vanguard takes hits first)", a.HP, b.HP, c.HP)
    	}
    	if host.Living() != 2 || vanguard.Living() != 1 {
    		t.Errorf("Living(): host %d, vanguard %d, want 2 (vanguard and C) and 1", host.Living(), vanguard.Living())
    	}
    	c.TakeDamage(5)
    	if got := host.Heal(3); got != 6 || b.HP != 9 || c.HP != 8 {
    		t.Errorf("host.Heal(3): restored %d, B %d, C %d; want 6 in total (B 9, C 8): nested warbands heal their members too", got, b.HP, c.HP)
    	}
    	u.TakeDamage(20)
    	if !host.Alive() || vanguard.Alive() || c.HP != 8 {
    		t.Errorf("after the vanguard falls: host.Alive() = %v, vanguard.Alive() = %v, C HP %d, want true, false, 8", host.Alive(), vanguard.Alive(), c.HP)
    	}
    	u.TakeDamage(3)
    	if c.HP != 5 {
    		t.Errorf("once the vanguard is gone, C should be hit: C has %d HP, want 5", c.HP)
    	}
    }

    // scarecrow is the test's own Unit. It records every hit it takes.
    type scarecrow struct {
    	hits  []int
    	straw int
    }

    func (s *scarecrow) Name() string { return "scarecrow" }
    func (s *scarecrow) Alive() bool  { return s.straw > 0 }
    func (s *scarecrow) TakeDamage(n int) int {
    	s.hits = append(s.hits, n)
    	lost := min(max(n, 0), s.straw)
    	s.straw -= lost
    	return lost
    }

    // mendable is a scarecrow that can be restuffed.
    type mendable struct{ scarecrow }

    func (m *mendable) Heal(n int) int { m.straw += n; return n }

    func TestCustomUnits(t *testing.T) {
    	s := &scarecrow{straw: 5}
    	m := &mendable{scarecrow{straw: 5}}
    	band := NewWarband("Field", s, m)
    	band.TakeDamage(3)
    	band.TakeDamage(9)
    	band.TakeDamage(2)
    	if len(s.hits) != 2 || s.hits[0] != 3 || s.hits[1] != 9 || len(m.hits) != 1 || m.hits[0] != 2 {
    		t.Errorf("hits 3, 9, 2 on [scarecrow(5), scarecrow(5)]: first saw %v, second saw %v, want [3 9] and [2]", s.hits, m.hits)
    	}
    	if got := band.Heal(4); got != 4 || m.straw != 7 || s.straw != 0 {
    		t.Errorf("Heal(4): restored %d, straw %d and %d; want 4, only the Healable one restuffed (0 and 7)", got, s.straw, m.straw)
    	}
    }

    func TestNewWarbandCopiesMembers(t *testing.T) {
    	a, b := knight("A", 10, 0), knight("B", 10, 0)
    	members := []Unit{a, b}
    	band := NewWarband("Pair", members...)
    	members[0] = &Wraith{Title: "impostor", Essence: 99}
    	band.TakeDamage(4)
    	if a.HP != 6 {
    		t.Errorf("changing the caller's slice after NewWarband changed the warband: A has %d HP, want 6", a.HP)
    	}
    }
---

The war room needs a combat model where a single knight and a whole army can be
treated the same way. Everything that fights is a `Unit`:

```go
type Unit interface {
	Name() string
	Alive() bool
	TakeDamage(n int) int // returns the HP actually lost
}
```

Units that can be healed also implement `Healable` (`Heal(n int) int`, returning
the HP actually restored). Build the pieces by **composition**:

1. **`Vitals`** (a component): `Alive` when `HP > 0`. `TakeDamage(n)` loses
   `n` HP but never goes below 0. `Heal(n)` restores up to `n` HP, capped at
   `MaxHP`. Dead units (0 HP) can't be healed, and a negative `n` does nothing
   in both methods.
2. **`Armor`** (a component): `Absorb(n)` returns how much of an attack gets
   through: `n - Defense`, but never less than 0.
3. **`Knight`** embeds both. A knight is a `Unit` and `Healable`, and its armor
   absorbs part of every hit **before** its vitals take the rest. Calling
   `k.Vitals.TakeDamage` directly still skips the armor.
4. **`Wraith`** keeps its own `Essence` instead of `Vitals`. Hits weaker than 5
   do nothing; stronger hits lose Essence like HP. Wraiths are **not**
   `Healable`.
5. **`Warband`**, made with `NewWarband(name, members...)`, is itself a `Unit`
   and `Healable`:
   - `TakeDamage(n)` sends the whole attack to its **first living** member
     (no spill-over) and returns what that member lost, or 0 if none is alive.
   - `Heal(n)` offers `n` HP to **every** member that is `Healable` and returns
     the total restored.
   - `Alive()` is true while any member is alive; `Living()` counts the living
     direct members.
   - Warbands can contain warbands, and the tests use their own `Unit` types.

The tests build knights with Go 1.27 promoted-field literals, such as
`&Knight{Title: "Ser Ayla", HP: 30, MaxHP: 30, Defense: 3}`, so keep the
embedded fields as they are.

## Example

```go
ayla := &Knight{Title: "Ser Ayla", HP: 30, MaxHP: 30, Defense: 3}
ghost := &Wraith{Title: "the Pale One", Essence: 12}
band := NewWarband("Ashen Company", ayla, ghost)

band.TakeDamage(10) // 7: Ayla's armor stops 3, she drops to 23
band.Heal(5)        // 5: Ayla back to 28; the wraith isn't Healable
band.TakeDamage(40) // 28: Ayla falls, nothing spills over to the wraith
band.TakeDamage(4)  // 0: too weak to hurt the wraith
band.TakeDamage(20) // 12
band.Alive()        // false
```

## Constraints

- `NewWarband` must copy its members: changing the caller's slice afterwards
  doesn't change the warband.
- Don't give `Knight` a named `Vitals` or `Armor` field: the embedding is the
  point.
