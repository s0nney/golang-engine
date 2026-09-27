---
title: Dependency Injection
quiz:
  - question: What is dependency injection?
    options:
      - text: A framework that Go requires for large programs
      - text: Passing the things a type depends on (like a logger or random number source) in from outside, instead of creating them inside
        correct: true
      - text: Using `init()` functions to set global variables
      - text: Embedding every dependency as an anonymous field
    explanation: |
      Dependency injection just means "hand me my dependencies". In Go, that's
      usually a constructor parameter of a small interface type. No framework needed.
  - question: |
      `Attack` calls `rand.IntN(20)` directly to roll a d20. Why is that hard to test?
    options:
      - text: '`rand.IntN` is too slow for tests'
      - text: The result is random, so a test can't reliably check what a critical hit or a miss does
        correct: true
      - text: The `math/rand/v2` package can't be imported in tests
      - text: It isn't hard to test at all
    explanation: |
      When a function reaches out to a hidden global dependency, the test can't
      control it. Inject a `Roller` interface instead, and tests can pass dice that
      always roll 20 or always roll 1.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"math/rand/v2"
    )

    type Roller interface {
    	Roll(sides int) int
    }

    // RandomDice is the real implementation.
    type RandomDice struct{}

    func (RandomDice) Roll(sides int) int { return rand.IntN(sides) + 1 }

    type Dragon struct {
    	name string
    	dice Roller
    }

    // NewDragon should store the dice it's given.
    func NewDragon(name string, dice Roller) *Dragon {
    	return &Dragon{name: name}
    }

    // FireBreath rolls three 6-sided dice and returns their sum. If all
    // three dice show the same number it's an inferno: double the sum.
    //
    // It should roll with d.dice, not reach for math/rand itself.
    func (d *Dragon) FireBreath() int {
    	a, b, c := rand.IntN(6)+1, rand.IntN(6)+1, rand.IntN(6)+1
    	sum := a + b + c
    	if a == b && b == c {
    		return sum * 2
    	}
    	return sum
    }

    func main() {
    	smaug := NewDragon("Smaug", RandomDice{})
    	dmg := smaug.FireBreath()
    	fmt.Println("damage in range?", dmg >= 3 && dmg <= 36)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"math/rand/v2"
    )

    type Roller interface {
    	Roll(sides int) int
    }

    // RandomDice is the real implementation.
    type RandomDice struct{}

    func (RandomDice) Roll(sides int) int { return rand.IntN(sides) + 1 }

    type Dragon struct {
    	name string
    	dice Roller
    }

    func NewDragon(name string, dice Roller) *Dragon {
    	return &Dragon{name: name, dice: dice}
    }

    // FireBreath rolls three 6-sided dice and returns their sum. If all
    // three dice show the same number it's an inferno: double the sum.
    func (d *Dragon) FireBreath() int {
    	a, b, c := d.dice.Roll(6), d.dice.Roll(6), d.dice.Roll(6)
    	sum := a + b + c
    	if a == b && b == c {
    		return sum * 2
    	}
    	return sum
    }

    func main() {
    	smaug := NewDragon("Smaug", RandomDice{})
    	dmg := smaug.FireBreath()
    	fmt.Println("damage in range?", dmg >= 3 && dmg <= 36)
    }
  tests: |
    package main

    import "testing"

    // scripted is a fake Roller that returns pre-set rolls and records
    // how it was used.
    type scripted struct {
    	rolls []int
    	sides []int
    }

    func (s *scripted) Roll(sides int) int {
    	s.sides = append(s.sides, sides)
    	if len(s.rolls) == 0 {
    		return 1
    	}
    	r := s.rolls[0]
    	s.rolls = s.rolls[1:]
    	return r
    }

    func TestFireBreath(t *testing.T) {
    	for _, tt := range []struct {
    		rolls []int
    		want  int
    	}{
    		{[]int{1, 2, 3}, 6},
    		{[]int{6, 5, 6}, 17},
    		{[]int{4, 4, 4}, 24},
    		{[]int{1, 1, 1}, 6},
    		{[]int{2, 2, 5}, 9},
    	} {
    		dice := &scripted{rolls: append([]int(nil), tt.rolls...)}
    		d := NewDragon("Smaug", dice)
    		if got := d.FireBreath(); got != tt.want {
    			t.Errorf("FireBreath() with rolls %v = %d, want %d (is the Dragon using the injected dice?)", tt.rolls, got, tt.want)
    		}
    		if len(dice.sides) != 3 {
    			t.Errorf("FireBreath() rolled the injected dice %d times, want 3", len(dice.sides))
    		}
    		for _, s := range dice.sides {
    			if s != 6 {
    				t.Errorf("FireBreath() called Roll(%d), want Roll(6)", s)
    			}
    		}
    	}
    }
---

Our combat system rolls dice. The quick way is to call the random number generator straight from the code:

```go
func (h *Hero) Attack() int {
	roll := rand.IntN(20) + 1 // a d20
	if roll == 20 {
		return h.power * 2 // critical hit!
	}
	if roll == 1 {
		return 0 // fumble
	}
	return h.power
}
```

It works. But how do you test that a critical hit doubles damage? You'd have to call `Attack` over and over and hope for a 20. And what if you want a "replay" mode that reproduces a fight exactly?

The problem: `Attack` **reaches out** to a hidden global dependency. **Dependency injection** flips it around: the hero is **handed** the thing it depends on.

## Injecting an interface

First, describe what we need, as a small interface declared where it's used:

```go
type Roller interface {
	Roll(sides int) int
}
```

Then give the hero a `Roller` through its constructor, and use it:

```go
package main

import (
	"fmt"
	"math/rand/v2"
)

type Roller interface {
	Roll(sides int) int
}

// RandomDice is the real implementation.
type RandomDice struct{}

func (RandomDice) Roll(sides int) int { return rand.IntN(sides) + 1 }

// FixedDice returns pre-set rolls in order: perfect for tests and replays.
type FixedDice struct {
	rolls []int
}

func (f *FixedDice) Roll(sides int) int {
	r := f.rolls[0]
	f.rolls = f.rolls[1:]
	return r
}

type Hero struct {
	name  string
	power int
	dice  Roller
}

func NewHero(name string, power int, dice Roller) *Hero {
	return &Hero{name: name, power: power, dice: dice}
}

func (h *Hero) Attack() int {
	switch h.dice.Roll(20) {
	case 20:
		return h.power * 2
	case 1:
		return 0
	default:
		return h.power
	}
}

func main() {
	// A scripted fight: normal hit, critical, fumble.
	aria := NewHero("Aria", 15, &FixedDice{rolls: []int{11, 20, 1}})
	fmt.Println(aria.Attack(), aria.Attack(), aria.Attack())

	// A real fight uses real dice.
	borin := NewHero("Borin", 12, RandomDice{})
	dmg := borin.Attack()
	fmt.Println(dmg == 0 || dmg == 12 || dmg == 24)
}
```

```
15 30 0
true
```

The scripted hero's output is completely predictable. The real hero still rolls random numbers. `Hero` has no idea which it has, and doesn't need to.

## Testing becomes easy

With injection, a test for critical hits is short and deterministic:

```go
func TestCriticalHitDoublesDamage(t *testing.T) {
	h := NewHero("Aria", 15, &FixedDice{rolls: []int{20}})
	if got := h.Attack(); got != 30 {
		t.Errorf("Attack() = %d, want 30", got)
	}
}
```

The same pattern works for anything a type talks to: the battle log (`io.Writer`, as in the last lesson), the clock (`func() time.Time`), the save system, or a network connection.

## No framework required

In some languages, "dependency injection" means a large framework with annotations and containers. In Go it's just:

1. Declare a small interface for what you need.
2. Accept it as a constructor parameter (or struct field).
3. Pass the real thing in `main`, and a fake in tests.

`main` becomes the place where your program is **wired together**: it creates the real dice, the real log file and the real save store, then hands them to the types that need them.

## Functions can be dependencies too

If the dependency is a single operation, a function type is often simpler than an interface:

```go
type Hero struct {
	name  string
	power int
	roll  func(sides int) int
}

func NewHero(name string, power int, roll func(sides int) int) *Hero {
	return &Hero{name: name, power: power, roll: roll}
}

aria := NewHero("Aria", 15, func(int) int { return 20 }) // always crits
borin := NewHero("Borin", 12, func(n int) int { return rand.IntN(n) + 1 })
```

Both are fine. Use an interface when there are several related methods or the implementation carries state (like `FixedDice`); use a function when there's just one operation.

## Your turn

The dragon's `FireBreath` reaches straight for `math/rand`, so nobody can test that an inferno (three matching dice) really doubles the damage. Refactor it:

1. Make `NewDragon` store the `Roller` it's given.
2. Make `FireBreath` roll three 6-sided dice **through `d.dice`**, returning their sum, doubled if all three match.

The tests inject their own scripted dice (you won't see that type), which is exactly the point of injection.

## Further reading

- [Learn Go with Tests: Dependency Injection](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/dependency-injection) and [Mocking](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/mocking)
