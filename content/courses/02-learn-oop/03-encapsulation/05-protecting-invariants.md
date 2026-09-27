---
title: Protecting Invariants
quiz:
  - question: What is an invariant?
    options:
      - text: A variable that can never be reassigned
      - text: A condition about an object's state that must always be true between method calls
        correct: true
      - text: A function with no side effects
      - text: A type that can't be embedded
    explanation: |
      "Gold is never negative" and "the party never has more than 4 members" are
      invariants. Encapsulation lets a type guarantee them by only allowing changes
      through methods that check the rules.
  - question: |
      What does this print?

      ```go
      func (w *Wallet) Spend(n int) error {
      	if n > w.gold {
      		return ErrNotEnoughGold
      	}
      	w.gold -= n
      	return nil
      }

      func main() {
      	w := &Wallet{gold: 30}
      	err := w.Spend(50)
      	fmt.Println(w.gold, err != nil)
      }
      ```
    options:
      - text: '`-20 true`'
      - text: '`30 true`'
        correct: true
      - text: '`30 false`'
      - text: '`-20 false`'
    explanation: |
      `Spend` checks the rule *before* changing anything, so the failed purchase
      leaves the wallet untouched at 30 and returns an error. Validate first, then
      mutate: the object is never left half-updated.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    var ErrNotEnoughGold = errors.New("not enough gold")

    // Wallet holds a party's gold. The gold must never be negative.
    type Wallet struct {
    	gold int
    }

    // Gold returns the current amount of gold.
    func (w *Wallet) Gold() int {
    	return w.gold
    }

    // Earn adds n gold. Negative amounts are ignored.
    func (w *Wallet) Earn(n int) {
    	w.gold += n
    }

    // Spend removes n gold. It returns ErrNotEnoughGold (and changes nothing)
    // if n is negative or more than the wallet holds.
    func (w *Wallet) Spend(n int) error {
    	w.gold -= n
    	return nil
    }

    func main() {
    	var w Wallet
    	w.Earn(100)
    	w.Earn(-50)
    	fmt.Println("gold:", w.Gold())
    	fmt.Println("buy sword:", w.Spend(80), "gold:", w.Gold())
    	fmt.Println("buy shield:", w.Spend(60), "gold:", w.Gold())
    	fmt.Println("sneaky refund:", w.Spend(-1000), "gold:", w.Gold())
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    var ErrNotEnoughGold = errors.New("not enough gold")

    type Wallet struct {
    	gold int
    }

    func (w *Wallet) Gold() int {
    	return w.gold
    }

    func (w *Wallet) Earn(n int) {
    	w.gold += max(0, n)
    }

    func (w *Wallet) Spend(n int) error {
    	if n < 0 || n > w.gold {
    		return ErrNotEnoughGold
    	}
    	w.gold -= n
    	return nil
    }

    func main() {
    	var w Wallet
    	w.Earn(100)
    	w.Earn(-50)
    	fmt.Println("gold:", w.Gold())
    	fmt.Println("buy sword:", w.Spend(80), "gold:", w.Gold())
    	fmt.Println("buy shield:", w.Spend(60), "gold:", w.Gold())
    	fmt.Println("sneaky refund:", w.Spend(-1000), "gold:", w.Gold())
    }
  tests: |
    package main

    import (
    	"errors"
    	"testing"
    )

    func TestEarnIgnoresNegative(t *testing.T) {
    	var w Wallet
    	w.Earn(100)
    	w.Earn(-50)
    	if got := w.Gold(); got != 100 {
    		t.Errorf("after Earn(100), Earn(-50): Gold() = %d, want 100", got)
    	}
    }

    func TestSpend(t *testing.T) {
    	var w Wallet
    	w.Earn(100)
    	if err := w.Spend(80); err != nil {
    		t.Fatalf("Spend(80) with 100 gold returned %v, want nil", err)
    	}
    	if got := w.Gold(); got != 20 {
    		t.Fatalf("after Spend(80): Gold() = %d, want 20", got)
    	}
    	if err := w.Spend(60); !errors.Is(err, ErrNotEnoughGold) {
    		t.Errorf("Spend(60) with 20 gold returned %v, want ErrNotEnoughGold", err)
    	}
    	if got := w.Gold(); got != 20 {
    		t.Errorf("after failed Spend(60): Gold() = %d, want 20 (unchanged)", got)
    	}
    	if err := w.Spend(-1000); !errors.Is(err, ErrNotEnoughGold) {
    		t.Errorf("Spend(-1000) returned %v, want ErrNotEnoughGold", err)
    	}
    	if got := w.Gold(); got != 20 {
    		t.Errorf("after Spend(-1000): Gold() = %d, want 20 (unchanged)", got)
    	}
    	if err := w.Spend(20); err != nil {
    		t.Errorf("Spend(20) with exactly 20 gold returned %v, want nil", err)
    	}
    }
---

An **invariant** is a rule about an object's state that must *always* hold. For our RPG:

- A wallet's gold is never negative.
- A party has at most 4 members, and no hero appears twice.
- A hero's HP is between 0 and max HP.

Encapsulation exists to protect invariants. Let's build a `Party` type that makes breaking its rules impossible from outside its package.

## Validate, then mutate

The golden rule for methods that change state: **check everything first, change things last**. If a check fails, return an error and leave the object exactly as it was.

```go
package main

import (
	"errors"
	"fmt"
	"slices"
)

var (
	ErrPartyFull = errors.New("party is full")
	ErrAlreadyIn = errors.New("hero already in party")
	ErrNotEnough = errors.New("not enough gold")
)

const maxPartySize = 4

type Party struct {
	members []string
	gold    int
}

func (p *Party) Add(name string) error {
	if len(p.members) >= maxPartySize {
		return ErrPartyFull
	}
	if slices.Contains(p.members, name) {
		return fmt.Errorf("%w: %s", ErrAlreadyIn, name)
	}
	p.members = append(p.members, name)
	return nil
}

func (p *Party) Earn(n int) {
	p.gold += max(0, n)
}

func (p *Party) Spend(n int) error {
	if n < 0 || n > p.gold {
		return ErrNotEnough
	}
	p.gold -= n
	return nil
}

func (p *Party) Members() []string { return slices.Clone(p.members) }
func (p *Party) Gold() int         { return p.gold }

func main() {
	var p Party
	for _, name := range []string{"Aria", "Borin", "Aria", "Cyra", "Dax", "Elm"} {
		if err := p.Add(name); err != nil {
			fmt.Println("error:", err)
		}
	}
	fmt.Println(p.Members())

	p.Earn(100)
	if err := p.Spend(250); err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(p.Gold())

	roster := p.Members()
	roster[0] = "Imposter"
	fmt.Println(p.Members()[0])
}
```

```
error: hero already in party: Aria
error: party is full
[Aria Borin Cyra Dax]
error: not enough gold
100
Aria
```

Look at what each piece does:

- `Add` checks both rules before appending, so a rejected hero never sneaks in.
- `Spend` refuses negative amounts too, otherwise `Spend(-1000)` would be a free money glitch.
- `Members` returns a **clone**, so editing `roster` can't corrupt the real party.
- The zero value `var p Party` is a valid, empty party. The invariants hold from the very first moment.

Errors are values here, declared once as sentinel variables, so callers can check `errors.Is(err, ErrPartyFull)` and react (maybe by showing "Your party is full!" in the UI).

## Invariants make code simpler everywhere else

Once `Party` guarantees its rules, the rest of the game can stop worrying. The shop doesn't need to check for negative gold. The UI doesn't need to de-duplicate the roster. The save system can trust that a party has at most 4 members.

That's the real payoff of encapsulation: you pay the cost of careful checking **once**, inside the type, and every caller gets correctness for free.

## Remember the package rule

This only holds for code *outside* the package. Inside it, a careless function could still write `p.gold = -5`. So keep packages focused, and keep the code that touches a type's internals close to the type itself, ideally in the same file.

## Assignment

The `Wallet` below has an unexported `gold` field, but its methods don't protect it. Fix them so the invariant "gold is never negative" always holds:

- `Earn(n)` ignores negative amounts.
- `Spend(n)` returns `ErrNotEnoughGold` and changes **nothing** if `n` is negative or more than the wallet holds. Otherwise it removes the gold and returns `nil`.
