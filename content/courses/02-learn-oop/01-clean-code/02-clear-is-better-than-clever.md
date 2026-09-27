---
title: Clear Is Better Than Clever
quiz:
  - question: |
      Two functions do the same thing. Which is more idiomatic Go?

      ```go
      // A
      func alive(hp int) bool { return !(hp <= 0 == true) }

      // B
      func isAlive(hp int) bool { return hp > 0 }
      ```
    options:
      - text: A, because packing logic into fewer characters is faster
      - text: B, because it says exactly what it means with no mental gymnastics
        correct: true
      - text: They're equally good, since both compile
    explanation: |
      Both compile and return the same result, but B reads like a sentence. Go
      culture values code that's obvious over code that's clever. As the Go
      proverb says, "clear is better than clever".
  - question: What does `gofmt` do for readability?
    options:
      - text: It renames your variables to be more descriptive
      - text: It rewrites your algorithms to be simpler
      - text: It formats every Go file the same way, so nobody argues about style
        correct: true
      - text: It deletes unused functions
    explanation: |
      `gofmt` handles indentation, spacing and alignment. Because everyone uses it,
      all Go code looks the same and you can focus on what code *does* rather than
      how it's laid out.
exercise:
  starter: |
    package main

    import "fmt"

    // xpForLevel returns the XP a hero needs to reach level l.
    // (It works out 100 * l * l. Can you tell?)
    func xpForLevel(l int) int {
    	return l*l<<6 + l*l<<5 + l*l<<2
    }

    // isLowHealth reports whether a living hero is at or below a quarter of max HP.
    func isLowHealth(h, m int) bool {
    	return h > 0 && h<<2 <= m
    }

    func main() {
    	fmt.Println(xpForLevel(1), xpForLevel(2), xpForLevel(10))                    // the designers want 150 600 15000
    	fmt.Println(isLowHealth(30, 100), isLowHealth(31, 100), isLowHealth(0, 100)) // they want true false false
    }
  solution: |
    package main

    import "fmt"

    // xpForLevel returns the XP a hero needs to reach level l.
    func xpForLevel(level int) int {
    	return 150 * level * level
    }

    // isLowHealth reports whether a living hero is at or below 30% of max HP.
    func isLowHealth(hp, maxHP int) bool {
    	isAlive := hp > 0
    	atOrBelow30Percent := hp*10 <= maxHP*3
    	return isAlive && atOrBelow30Percent
    }

    func main() {
    	fmt.Println(xpForLevel(1), xpForLevel(2), xpForLevel(10))
    	fmt.Println(isLowHealth(30, 100), isLowHealth(31, 100), isLowHealth(0, 100))
    }
  tests: |
    package main

    import "testing"

    func TestXP(t *testing.T) {
    	for _, tt := range []struct{ level, want int }{{0, 0}, {1, 150}, {2, 600}, {3, 1350}, {10, 15000}} {
    		if got := xpForLevel(tt.level); got != tt.want {
    			t.Errorf("xpForLevel(%d) = %d, want %d (150 * level * level)", tt.level, got, tt.want)
    		}
    	}
    }

    func TestLow(t *testing.T) {
    	for _, tt := range []struct {
    		hp, maxHP int
    		want      bool
    	}{
    		{30, 100, true},
    		{25, 100, true},
    		{31, 100, false},
    		{100, 100, false},
    		{0, 100, false},
    		{3, 10, true},
    		{4, 10, false},
    		{45, 150, true},
    		{46, 150, false},
    	} {
    		if got := isLowHealth(tt.hp, tt.maxHP); got != tt.want {
    			t.Errorf("isLowHealth(%d, %d) = %v, want %v (alive and at or below 30%% of max HP)", tt.hp, tt.maxHP, got, tt.want)
    		}
    	}
    }
---

Most of a programmer's time is spent **reading** code, not writing it. You read to fix bugs, to add features, and to remember what on earth you were thinking last Tuesday. So the single most valuable property of code is that it's **easy to read**.

## A Go proverb

Rob Pike, one of Go's creators, collected a list of short sayings known as the [Go Proverbs](https://go-proverbs.github.io/). One of them sums up Go's whole attitude:

> Clear is better than clever.

Clever code makes the author feel smart. Clear code makes the *reader* feel smart. Guess which one your teammates prefer.

## Clever vs clear

Here's a clever way to work out a hero's damage:

```go
func dmg(a, d, c int) int {
	return max(0, a-d) << (c & 1)
}
```

Did you spot that `<< (c & 1)` doubles the damage when `c` is odd? Probably not at a glance. Here's the same logic written clearly:

```go
func damage(attack, defense int, critical bool) int {
	dmg := max(0, attack-defense)
	if critical {
		dmg *= 2
	}
	return dmg
}
```

It's longer, and it's much better. Every line says what it means. A new teammate could change the critical multiplier to 3 without asking anybody.

Let's run it:

```go
package main

import "fmt"

func damage(attack, defense int, critical bool) int {
	dmg := max(0, attack-defense)
	if critical {
		dmg *= 2
	}
	return dmg
}

func main() {
	fmt.Println(damage(50, 20, false))
	fmt.Println(damage(50, 20, true))
	fmt.Println(damage(10, 20, true))
}
```

```
30
60
0
```

## Go helps you be boring

Go's design pushes you towards clear code:

- **`gofmt`** formats all Go code identically. No debates about tabs or brace placement.
- **Few features.** There's usually one obvious way to write a loop (`for`), so readers don't need to learn five.
- **Explicit errors.** `if err != nil` is verbose, but you can always see where things can fail.
- **No hidden magic.** No operator overloading, no implicit constructors, no exceptions jumping across your call stack.

The Go team's motto is that code is read far more than it's written, so it optimises for the reader.

## Why this matters for OOP

Object-oriented programming is, at heart, a set of techniques for keeping big programs readable: group related data, hide details, give things meaningful names. If an "OOP design" makes the code *harder* to follow (deep class hierarchies, five layers of indirection to deal damage), it has failed at its only job.

Keep that in mind for the rest of the course. Every technique we learn is only worth using when it makes the code clearer.

## Assignment

A "clever" teammate wrote two helpers, and now the designers want changes:

- `xpForLevel` secretly computes `100 × level × level` with bit shifts. The new curve is **150 × level × level**.
- `isLowHealth` warns when a living hero is at or below a quarter of max HP. The healer now wants the warning at **30%** or below. A fallen hero (0 HP) is never "low", it's gone.

Rewrite both so the rule is obvious at a glance: give the parameters real names and drop the bit tricks. Keep the function names, because the rest of the game calls them. Integer division rounds down, so compare `hp*10` with `maxHP*3` rather than dividing.

## Further reading

- [Go Proverbs](https://go-proverbs.github.io/)
