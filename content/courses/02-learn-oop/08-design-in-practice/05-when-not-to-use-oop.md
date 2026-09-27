---
title: When Not to Use OOP
quiz:
  - question: |
      You need to compute experience points needed for a level: `100 * level * level`. Which is the most idiomatic Go?
    options:
      - text: An `XPCalculator` interface, an `xpCalculatorImpl` struct, and a `NewXPCalculator` constructor
      - text: '`func XPForLevel(level int) int { return 100 * level * level }`'
        correct: true
      - text: A method on a global singleton `var Calc = &Calculator{}`
      - text: A generic `Calculator[T]` type with a self-referential constraint
    explanation: |
      It's a pure calculation with no state and one implementation. A plain
      function is the clearest option. Wrapping it in objects adds ceremony and
      no value.
  - question: When is introducing an interface most clearly justified?
    options:
      - text: Always, for every struct, just in case
      - text: When there are (or you need to substitute, e.g. in tests) at least two implementations that code must treat uniformly
        correct: true
      - text: Only when the interface has at least five methods
      - text: Never; Go discourages interfaces
    explanation: |
      Interfaces pay for themselves when they let code work with more than one
      concrete type. With only one implementation and no need to fake it, the
      interface is just an extra layer to read through.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"io"
    	"os"
    )

    // Roller rolls a die with the given number of sides (chapter 8).
    type Roller interface {
    	Roll(sides int) int
    }

    // FixedDice returns pre-set rolls in order, then 1s.
    type FixedDice struct {
    	rolls []int
    }

    func (f *FixedDice) Roll(sides int) int {
    	if len(f.rolls) == 0 {
    		return 1
    	}
    	r := f.rolls[0]
    	f.rolls = f.rolls[1:]
    	return r
    }

    // Stats is the shared component embedded by every creature (chapter 5).
    type Stats struct {
    	HP, Power int
    }

    func (s *Stats) TakeDamage(n int) { s.HP = max(0, s.HP-n) }
    func (s *Stats) Alive() bool      { return s.HP > 0 }

    // Combatant is anything that can fight (chapters 4 and 6).
    type Combatant interface {
    	Name() string
    	Alive() bool
    	TakeDamage(n int)
    	Attack(dice Roller) int
    }

    type Hero struct {
    	Stats
    	name string
    }

    func (h *Hero) Name() string { return h.name }

    // Attack rolls one 20-sided die: 20 is a critical hit (double Power),
    // 1 is a fumble (0), anything else deals Power.
    func (h *Hero) Attack(dice Roller) int {
    	// ?
    	return 0
    }

    type Dragon struct {
    	Stats
    }

    func (d *Dragon) Name() string { return "the dragon" }

    // Attack is fire breath: roll three 6-sided dice and return the sum,
    // doubled if all three match.
    func (d *Dragon) Attack(dice Roller) int {
    	// ?
    	return 0
    }

    // Compile-time checks (chapter 8).
    var (
    	_ Combatant = (*Hero)(nil)
    	_ Combatant = (*Dragon)(nil)
    )

    // StalemateError is returned when nobody wins in time (chapter 6).
    type StalemateError struct {
    	Rounds int
    }

    // Error should read like "no winner after 10 rounds".
    func (e *StalemateError) Error() string {
    	// ?
    	return ""
    }

    // Battle runs up to maxRounds rounds. In each round a attacks b, then
    // b attacks a. Every attack writes one line to log:
    //
    //	<attacker> hits <defender> for <damage>
    //
    // As soon as a defender is no longer alive, write "<attacker> wins!"
    // and return the attacker. If both survive every round, return nil
    // and a *StalemateError.
    func Battle(a, b Combatant, dice Roller, maxRounds int, log io.Writer) (Combatant, error) {
    	// ?
    	return nil, nil
    }

    func main() {
    	aria := &Hero{name: "Aria", HP: 40, Power: 15}
    	dragon := &Dragon{HP: 45}
    	dice := &FixedDice{rolls: []int{
    		20,      // Aria crits: 30
    		2, 3, 4, // dragon breath: 9
    		1,       // Aria fumbles: 0
    		5, 5, 5, // inferno: 30
    		12, // Aria hits: 15, and the dragon falls
    	}}
    	winner, err := Battle(aria, dragon, dice, 10, os.Stdout)
    	if err != nil || winner == nil {
    		fmt.Println("no winner:", err)
    		return
    	}
    	fmt.Println("Winner:", winner.Name(), "with", aria.HP, "HP left")
    }
  solution: |
    package main

    import (
    	"fmt"
    	"io"
    	"os"
    )

    // Roller rolls a die with the given number of sides (chapter 8).
    type Roller interface {
    	Roll(sides int) int
    }

    // FixedDice returns pre-set rolls in order, then 1s.
    type FixedDice struct {
    	rolls []int
    }

    func (f *FixedDice) Roll(sides int) int {
    	if len(f.rolls) == 0 {
    		return 1
    	}
    	r := f.rolls[0]
    	f.rolls = f.rolls[1:]
    	return r
    }

    // Stats is the shared component embedded by every creature (chapter 5).
    type Stats struct {
    	HP, Power int
    }

    func (s *Stats) TakeDamage(n int) { s.HP = max(0, s.HP-n) }
    func (s *Stats) Alive() bool      { return s.HP > 0 }

    // Combatant is anything that can fight (chapters 4 and 6).
    type Combatant interface {
    	Name() string
    	Alive() bool
    	TakeDamage(n int)
    	Attack(dice Roller) int
    }

    type Hero struct {
    	Stats
    	name string
    }

    func (h *Hero) Name() string { return h.name }

    func (h *Hero) Attack(dice Roller) int {
    	switch dice.Roll(20) {
    	case 20:
    		return h.Power * 2
    	case 1:
    		return 0
    	default:
    		return h.Power
    	}
    }

    type Dragon struct {
    	Stats
    }

    func (d *Dragon) Name() string { return "the dragon" }

    func (d *Dragon) Attack(dice Roller) int {
    	a, b, c := dice.Roll(6), dice.Roll(6), dice.Roll(6)
    	sum := a + b + c
    	if a == b && b == c {
    		return sum * 2
    	}
    	return sum
    }

    // Compile-time checks (chapter 8).
    var (
    	_ Combatant = (*Hero)(nil)
    	_ Combatant = (*Dragon)(nil)
    )

    // StalemateError is returned when nobody wins in time (chapter 6).
    type StalemateError struct {
    	Rounds int
    }

    func (e *StalemateError) Error() string {
    	return fmt.Sprintf("no winner after %d rounds", e.Rounds)
    }

    func Battle(a, b Combatant, dice Roller, maxRounds int, log io.Writer) (Combatant, error) {
    	for range maxRounds {
    		for _, turn := range [][2]Combatant{{a, b}, {b, a}} {
    			attacker, defender := turn[0], turn[1]
    			dmg := attacker.Attack(dice)
    			defender.TakeDamage(dmg)
    			fmt.Fprintf(log, "%s hits %s for %d\n", attacker.Name(), defender.Name(), dmg)
    			if !defender.Alive() {
    				fmt.Fprintf(log, "%s wins!\n", attacker.Name())
    				return attacker, nil
    			}
    		}
    	}
    	return nil, &StalemateError{Rounds: maxRounds}
    }

    func main() {
    	aria := &Hero{name: "Aria", HP: 40, Power: 15}
    	dragon := &Dragon{HP: 45}
    	dice := &FixedDice{rolls: []int{
    		20,      // Aria crits: 30
    		2, 3, 4, // dragon breath: 9
    		1,       // Aria fumbles: 0
    		5, 5, 5, // inferno: 30
    		12, // Aria hits: 15, and the dragon falls
    	}}
    	winner, err := Battle(aria, dragon, dice, 10, os.Stdout)
    	if err != nil || winner == nil {
    		fmt.Println("no winner:", err)
    		return
    	}
    	fmt.Println("Winner:", winner.Name(), "with", aria.HP, "HP left")
    }
  tests: |
    package main

    import (
    	"errors"
    	"strings"
    	"testing"
    )

    type scripted struct{ rolls, sides []int }

    func (s *scripted) Roll(sides int) int {
    	s.sides = append(s.sides, sides)
    	if len(s.rolls) == 0 {
    		return 1
    	}
    	r := s.rolls[0]
    	s.rolls = s.rolls[1:]
    	return r
    }

    // dummy is a training dummy: a Combatant the solution has never seen.
    type dummy struct{ hp int }

    func (d *dummy) Name() string           { return "the dummy" }
    func (d *dummy) Alive() bool            { return d.hp > 0 }
    func (d *dummy) TakeDamage(n int)       { d.hp -= n }
    func (d *dummy) Attack(dice Roller) int { return 0 }

    func TestHeroAttack(t *testing.T) {
    	for _, tt := range []struct{ roll, want int }{{20, 30}, {1, 0}, {2, 15}, {19, 15}} {
    		h := &Hero{name: "Aria", HP: 10, Power: 15}
    		dice := &scripted{rolls: []int{tt.roll}}
    		if got := h.Attack(dice); got != tt.want {
    			t.Errorf("Hero{Power: 15}.Attack with a roll of %d = %d, want %d", tt.roll, got, tt.want)
    		}
    		if len(dice.sides) != 1 || dice.sides[0] != 20 {
    			t.Errorf("Hero.Attack rolled %v, want exactly one Roll(20)", dice.sides)
    		}
    	}
    }

    func TestDragonAttack(t *testing.T) {
    	for _, tt := range []struct {
    		rolls []int
    		want  int
    	}{{[]int{1, 2, 3}, 6}, {[]int{6, 6, 6}, 36}, {[]int{3, 3, 2}, 8}} {
    		d := &Dragon{HP: 10}
    		dice := &scripted{rolls: append([]int(nil), tt.rolls...)}
    		if got := d.Attack(dice); got != tt.want {
    			t.Errorf("Dragon.Attack with rolls %v = %d, want %d", tt.rolls, got, tt.want)
    		}
    		if len(dice.sides) != 3 {
    			t.Errorf("Dragon.Attack rolled %d dice, want 3", len(dice.sides))
    		}
    		for _, s := range dice.sides {
    			if s != 6 {
    				t.Errorf("Dragon.Attack called Roll(%d), want Roll(6)", s)
    			}
    		}
    	}
    }

    func TestBattleHeroWins(t *testing.T) {
    	aria := &Hero{name: "Aria", HP: 40, Power: 15}
    	dragon := &Dragon{HP: 45}
    	dice := &scripted{rolls: []int{20, 2, 3, 4, 1, 5, 5, 5, 12}}
    	var log strings.Builder
    	winner, err := Battle(aria, dragon, dice, 10, &log)
    	if err != nil {
    		t.Fatalf("Battle returned error %v, want nil", err)
    	}
    	if winner != Combatant(aria) {
    		t.Errorf("Battle winner = %v, want Aria", winner)
    	}
    	want := `Aria hits the dragon for 30
    the dragon hits Aria for 9
    Aria hits the dragon for 0
    the dragon hits Aria for 30
    Aria hits the dragon for 15
    Aria wins!
    `
    	if log.String() != want {
    		t.Errorf("battle log =\n%s\nwant\n%s", log.String(), want)
    	}
    	if aria.HP != 1 || dragon.HP != 0 {
    		t.Errorf("after battle: Aria HP = %d (want 1), dragon HP = %d (want 0)", aria.HP, dragon.HP)
    	}
    }

    func TestBattleStalemateAndOtherCombatants(t *testing.T) {
    	d := &dummy{hp: 1000}
    	dragon := &Dragon{HP: 5}
    	hero := &Hero{name: "Borin", HP: 5, Power: 3}
    	var log strings.Builder
    	winner, err := Battle(d, hero, &scripted{}, 3, &log)
    	if winner != nil {
    		t.Errorf("Battle(dummy, hero) winner = %v, want nil", winner)
    	}
    	se, ok := errors.AsType[*StalemateError](err)
    	if !ok {
    		t.Fatalf("Battle(dummy, hero) error = %v, want a *StalemateError", err)
    	}
    	if se.Rounds != 3 || err.Error() != "no winner after 3 rounds" {
    		t.Errorf("StalemateError = {Rounds: %d} %q, want {Rounds: 3} \"no winner after 3 rounds\"", se.Rounds, err.Error())
    	}
    	if n := strings.Count(log.String(), "\n"); n != 6 {
    		t.Errorf("a 3-round stalemate should log 6 attacks, got %d lines:\n%s", n, log.String())
    	}

    	log.Reset()
    	winner, err = Battle(&dummy{hp: 1000}, dragon, &scripted{rolls: []int{6, 6, 6}}, 5, &log)
    	if err == nil && winner != nil {
    		t.Errorf("Battle(dummy, dragon): dummy deals 0 damage and the dragon can't reach 1000 HP in one breath, got winner %s", winner.Name())
    	}
    	log.Reset()
    	d2 := &dummy{hp: 10}
    	winner, err = Battle(d2, &Hero{name: "Cara", HP: 50, Power: 10}, &scripted{rolls: []int{7}}, 5, &log)
    	if err != nil || winner == nil || winner.Name() != "Cara" {
    		t.Fatalf("Battle(dummy with 10 HP, Cara with Power 10) = (%v, %v), want Cara to win", winner, err)
    	}
    	want := "the dummy hits Cara for 0\nCara hits the dummy for 10\nCara wins!\n"
    	if log.String() != want {
    		t.Errorf("battle log =\n%s\nwant\n%s", log.String(), want)
    	}
    }
---

You've now got a full toolbox: methods, encapsulation, interfaces, embedding, generics, dependency injection, functional options. The last lesson is about the most underrated skill of all: **knowing when to leave the tools in the box**.

## Plain functions are fine

Go is not an "everything must be an object" language. Many things are best as ordinary functions:

```go
package main

import "fmt"

func xpForLevel(level int) int {
	return 100 * level * level
}

func levelForXP(xp int) int {
	level := 1
	for xpForLevel(level+1) <= xp {
		level++
	}
	return level
}

func main() {
	fmt.Println(xpForLevel(5))
	fmt.Println(levelForXP(2600))
}
```

```
2500
5
```

No `XPCalculator` interface. No `XPService` struct. No constructor. It's a calculation, so it's a function. The standard library is full of these: `strings.ToUpper`, `slices.Sort`, `math.Max`.

## Warning signs of over-engineering

If you catch yourself doing any of these, stop and ask whether the abstraction is earning its keep:

- **An interface with exactly one implementation**, and no test that needs to fake it. Go's advice: don't define the interface until a second implementation, or a consumer that needs one, shows up. Implicit satisfaction means you can add it later without touching the type.
- **Getters and setters for every field** with no rules to protect. Export the field.
- **A `Manager`, `Helper`, `Service` or `Impl` type** whose methods don't use any of its fields. Those are functions wearing a costume.
- **Deep embedding chains** that recreate a class hierarchy (`Entity` → `Creature` → `Monster` → `Dragon`). Readers have to walk four types to find a method.
- **Factories for factories.** If a constructor needs a constructor, something has gone wrong.

## A before and after

The over-engineered version:

```go
type DamageCalculator interface {
	Calculate(attack, defense int) int
}

type damageCalculatorImpl struct{}

func NewDamageCalculator() DamageCalculator {
	return &damageCalculatorImpl{}
}

func (c *damageCalculatorImpl) Calculate(attack, defense int) int {
	return max(1, attack-defense)
}

// usage
calc := NewDamageCalculator()
dmg := calc.Calculate(30, 12)
```

The Go version:

```go
func damage(attack, defense int) int {
	return max(1, attack-defense)
}

// usage
dmg := damage(30, 12)
```

Same behaviour, a quarter of the code, nothing to navigate through. If one day there really are several damage formulas chosen at runtime, *then* introduce a small interface (or just a `func(int, int) int` field), in the package that needs it.

## So when is OOP the right call?

Use the tools from this course when they solve a real problem:

| Tool | Worth it when... |
|------|------------------|
| Methods on a type | behaviour clearly belongs to that data (`hero.TakeDamage`) |
| Unexported fields + constructor | the type has invariants to protect |
| Interfaces | code must work with **several** concrete types, or you need a fake in tests |
| Embedding | you're genuinely reusing a component's methods |
| Generics | the same algorithm or container is needed for several types |
| Functional options | many optional settings, most callers use few |

## The Go mindset

Go's creators designed the language to push back against complexity. The course started with "clear is better than clever", and it ends there too. The best Go code usually looks almost boring: concrete types, small interfaces discovered from real needs, plain functions for plain calculations, and exactly as much abstraction as the problem calls for. No more.

## Final project: the dragon fight

Time to put the whole course together. The editor holds a small battle engine built from pieces you've already met:

- **`Stats`** is a component embedded in both `Hero` and `Dragon` (chapter 5), so they share `TakeDamage` and `Alive`. Thanks to Go 1.27 you can write `&Hero{name: "Aria", HP: 40, Power: 15}`, setting the promoted fields directly.
- **`Combatant`** is the small interface `Battle` works through (chapters 4 and 6), with compile-time checks that both types satisfy it (this chapter).
- **`Roller`** is the injected dice (this chapter), so the fight can be scripted.
- **`StalemateError`** is a custom error type (chapter 6).
- The battle log is an injected **`io.Writer`**, so `main` prints to `os.Stdout` and the tests capture it in a `strings.Builder`.

Finish the four missing pieces:

1. `(*Hero).Attack`: roll one d20. A 20 deals double `Power`, a 1 deals 0, anything else deals `Power`.
2. `(*Dragon).Attack`: roll three d6 and return the sum, doubled if all three match.
3. `(*StalemateError).Error`: `no winner after N rounds`.
4. `Battle(a, b, dice, maxRounds, log)`: each round `a` attacks `b`, then `b` attacks `a`. Each attack deals damage and writes `<attacker> hits <defender> for <damage>` to `log`. When a defender falls, write `<attacker> wins!` and return the attacker with a nil error. If both survive `maxRounds` rounds, return `nil` and a `*StalemateError`.

`Battle` must only use the `Combatant` interface: the tests send in a training dummy you've never seen.

## What's next

You can now model a program as types with behaviour: encapsulated state, small interfaces, composition instead of inheritance, and generics where a container needs them. You also know when to skip all that and write a function.

That last point is where the next course picks up. In [Learn Functional Programming in Go](/courses/learn-functional-programming) functions take centre stage: passing them around as values, closures, pure functions, iterators and sum types, all while building a document converter. You'll see the same ideas from a different angle, like dependency injection done with a single `func` instead of an interface.

Now go slay some dragons. Clearly.

## Further reading

- [Go FAQ: Is Go an object-oriented language?](https://go.dev/doc/faq#Is_Go_an_object-oriented_language)
- [Go Proverbs](https://go-proverbs.github.io/)
