---
title: Don't Repeat Yourself
quiz:
  - question: |
      The game has this code in three different files:

      ```go
      if hp-dmg < 0 {
      	hp = 0
      } else {
      	hp -= dmg
      }
      ```

      The designers now want a "last stand" rule: heroes survive at 1 HP once per fight. What's the main risk?
    options:
      - text: The compiler will reject the duplicated code
      - text: You might update two of the copies and forget the third, so the rule only works sometimes
        correct: true
      - text: Duplicated code runs slower
      - text: There's no risk, because duplicated code is always fine
    explanation: |
      Duplicated logic means duplicated maintenance. When a rule changes you have
      to find and fix every copy, and missing one creates an inconsistent bug.
      Putting the rule in one function fixes that.
  - question: Which proverb warns against taking DRY too far?
    options:
      - text: '"Don''t panic."'
      - text: '"A little copying is better than a little dependency."'
        correct: true
      - text: '"Errors are values."'
      - text: '"Make the zero value useful."'
    explanation: |
      Sometimes two bits of code look alike by coincidence. Forcing them to share
      code (or pulling in a whole package for one tiny function) couples things that
      should change independently.
exercise:
  starter: |
    package main

    import "fmt"

    // applyDamage returns the hero's HP after a hit of dmg.
    // Armor reduces every hit (never below 0 damage) and HP never drops below 0.
    func applyDamage(hp, dmg, armor int) int {
    	// ?
    	return hp
    }

    func main() {
    	hp := 100
    	hp = applyDamage(hp, 30, 5) // sword
    	fmt.Println("after sword:", hp)
    	hp = applyDamage(hp, 3, 5) // pebble
    	fmt.Println("after pebble:", hp)
    	hp = applyDamage(hp, 200, 5) // dragon fire
    	fmt.Println("after dragon fire:", hp)
    }
  solution: |
    package main

    import "fmt"

    func applyDamage(hp, dmg, armor int) int {
    	dmg = max(0, dmg-armor)
    	return max(0, hp-dmg)
    }

    func main() {
    	hp := 100
    	hp = applyDamage(hp, 30, 5)
    	fmt.Println("after sword:", hp)
    	hp = applyDamage(hp, 3, 5)
    	fmt.Println("after pebble:", hp)
    	hp = applyDamage(hp, 200, 5)
    	fmt.Println("after dragon fire:", hp)
    }
  tests: |
    package main

    import "testing"

    func TestApplyDamage(t *testing.T) {
    	for _, tt := range []struct {
    		hp, dmg, armor, want int
    	}{
    		{100, 30, 5, 75},
    		{75, 3, 5, 75},
    		{75, 200, 5, 0},
    		{50, 10, 0, 40},
    		{0, 10, 0, 0},
    	} {
    		if got := applyDamage(tt.hp, tt.dmg, tt.armor); got != tt.want {
    			t.Errorf("applyDamage(%d, %d, %d) = %d, want %d", tt.hp, tt.dmg, tt.armor, got, tt.want)
    		}
    	}
    }
---

**DRY** stands for **Don't Repeat Yourself**. The idea: every piece of knowledge in your program should live in exactly one place.

## The problem with copies

Suppose three places in our RPG apply damage: melee attacks, arrows and dragon fire. A rushed developer writes the same logic three times:

```go
// melee.go
heroHP -= dmg
if heroHP < 0 {
	heroHP = 0
}

// archery.go
heroHP -= arrowDmg
if heroHP < 0 {
	heroHP = 0
}

// dragon.go
heroHP -= fireDmg
if heroHP < 0 {
	heroHP = 0
}
```

Now the designers add armour, which reduces all incoming damage by 5. You have to find every copy and change it. Miss one and dragon fire ignores armour. Nobody notices until a player files a very angry bug report.

## One source of truth

Pull the rule into a function:

```go
package main

import "fmt"

const armor = 5

func applyDamage(hp, dmg int) int {
	dmg = max(0, dmg-armor)
	return max(0, hp-dmg)
}

func main() {
	hp := 100
	hp = applyDamage(hp, 30) // sword
	hp = applyDamage(hp, 12) // arrow
	hp = applyDamage(hp, 90) // dragon fire
	fmt.Println(hp)
}
```

```
0
```

The sword does 25, the arrow 7, which leaves 68, and the dragon's 85 finishes the job. The armour rule now lives in **one** place. Change `applyDamage` and every attack in the game follows.

This is the seed of OOP. Soon we'll attach `applyDamage` to a `Hero` type as a method, so the data (`hp`) and the rule that changes it live together.

## Don't overdo it

DRY is about *knowledge*, not about text that happens to look the same. Consider:

```go
const maxPartySize = 4
const maxInventorySlots = 4
```

Both are `4`, but they mean completely different things. If you "DRY them up" into one constant, then growing the party to 5 would secretly grow every inventory too. That's worse, not better.

Go has a proverb for this:

> A little copying is better than a little dependency.

If two pieces of code look alike but could change for different reasons, it's fine to let them stay separate. Duplication is cheaper than the wrong abstraction.

## Rule of thumb

- The same **rule** in two places? Extract it now.
- Similar-looking **code** that means different things? Leave it.
- Not sure yet? Wait until you see the pattern a third time, then extract it.

## Assignment

The armour rule has been copied around the codebase again. Put it in **one** place: complete `applyDamage(hp, dmg, armor int) int` so that armour reduces every hit (a hit can't do less than 0 damage) and HP never drops below 0. Then every attack in `main` goes through it.
