---
title: Loot Appraiser
difficulty: easy
after: polymorphism
hints:
  - 'A type switch can have an **interface** as a case: `case Valuer:` matches any value whose type has a `Value() int` method. Cases are tried top to bottom, so put it first.'
  - 'Inside `case []any:` the variable already has type `[]any`. Loop over it and call `appraise` on each element: a pouch inside a pouch just works.'
  - 'A named type like `type Coins int` is **not** `int`, so it won''t match `case int`. Anything that matches no case, including `nil`, falls through to `default` and is worth 0.'
exercise:
  starter: |
    package main

    import "fmt"

    // Valuer is loot that knows its own worth.
    type Valuer interface {
    	Value() int
    }

    type Gem struct {
    	Carats int
    }

    func (g Gem) Value() int { return g.Carats * 50 }

    // appraise returns the value, in gold, of one piece of loot:
    //   - anything that implements Valuer is worth Value()
    //   - an int is a pile of gold coins, worth that many gold (negative piles are worth 0)
    //   - a string is a scroll, worth 10 gold
    //   - a []any is a pouch, worth the sum of everything inside it
    //   - anything else (including nil) is worth 0
    func appraise(loot any) int {
    	return 0
    }

    func main() {
    	pouch := []any{Gem{Carats: 2}, 35, "Scroll of Fire", nil}
    	fmt.Println(appraise(Gem{Carats: 3})) // want 150
    	fmt.Println(appraise(pouch))          // want 145
    	fmt.Println(appraise(3.5))            // want 0
    }
  solution: |
    package main

    import "fmt"

    // Valuer is loot that knows its own worth.
    type Valuer interface {
    	Value() int
    }

    type Gem struct {
    	Carats int
    }

    func (g Gem) Value() int { return g.Carats * 50 }

    const scrollValue = 10

    // appraise returns the value, in gold, of one piece of loot.
    func appraise(loot any) int {
    	switch v := loot.(type) {
    	case Valuer:
    		return v.Value()
    	case int:
    		return max(0, v)
    	case string:
    		return scrollValue
    	case []any:
    		total := 0
    		for _, item := range v {
    			total += appraise(item)
    		}
    		return total
    	default:
    		return 0
    	}
    }

    func main() {
    	pouch := []any{Gem{Carats: 2}, 35, "Scroll of Fire", nil}
    	fmt.Println(appraise(Gem{Carats: 3}))
    	fmt.Println(appraise(pouch))
    	fmt.Println(appraise(3.5))
    }
  tests: |
    package main

    import "testing"

    // Crown is the test's own Valuer.
    type Crown struct{ Jewels int }

    func (c *Crown) Value() int { return 1000 + 100*c.Jewels }

    // Coins is a named type whose underlying type is int. It's not an int.
    type Coins int

    // Tally is an int-based type that is also a Valuer.
    type Tally int

    func (t Tally) Value() int { return int(t) * 2 }

    func TestAppraise(t *testing.T) {
    	tests := []struct {
    		name string
    		loot any
    		want int
    	}{
    		{"gem", Gem{Carats: 3}, 150},
    		{"tiny gem", Gem{}, 0},
    		{"coins", 35, 35},
    		{"no coins", 0, 0},
    		{"a debt note", -20, 0},
    		{"scroll", "Scroll of Fire", 10},
    		{"blank scroll", "", 10},
    		{"nil", nil, 0},
    		{"float", 3.5, 0},
    		{"bool", true, 0},
    		{"named int type", Coins(99), 0},
    		{"*Crown (the test's own Valuer)", &Crown{Jewels: 2}, 1200},
    		{"Crown value: no Value method on the value type", Crown{Jewels: 2}, 0},
    		{"Valuer beats int", Tally(21), 42},
    		{"empty pouch", []any{}, 0},
    		{"pouch", []any{Gem{Carats: 2}, 35, "Scroll of Fire", nil}, 145},
    		{"pouch in a pouch", []any{5, []any{5, []any{"scroll", Gem{Carats: 1}}}, -3}, 70},
    		{"[]int is not a pouch", []int{1, 2, 3}, 0},
    		{"hoard", []any{&Crown{}, Tally(5), Coins(7), 1000, []any{"a", "b"}}, 2030},
    	}
    	for _, tt := range tests {
    		if got := appraise(tt.loot); got != tt.want {
    			t.Errorf("%s: appraise(%v) = %d, want %d", tt.name, tt.loot, got, tt.want)
    		}
    	}
    }
---

After every dungeon crawl the party dumps its loot on the tavern table, and the
appraiser has to price a heap of **completely different things**. Write
`appraise(loot any) int`, returning the value of one piece of loot in gold:

| Loot | Worth |
|---|---|
| anything that implements `Valuer` | its `Value()` |
| an `int` (a pile of coins) | that many gold, but a negative pile is worth 0 |
| a `string` (a scroll) | 10 gold |
| a `[]any` (a pouch) | the sum of everything inside it, pouches included |
| anything else, including `nil` | 0 |

The rows are checked **top to bottom**: a type that has a `Value` method is
priced by it even if it's also based on `int`. Only the exact types `int`,
`string` and `[]any` count: `type Coins int` is not an `int`, and a `[]int` is
not a pouch.

## Examples

```go
appraise(Gem{Carats: 3})                                   // 150 (Gem is a Valuer)
appraise([]any{Gem{Carats: 2}, 35, "Scroll of Fire", nil}) // 100 + 35 + 10 + 0 = 145
appraise([]any{5, []any{5, "scroll"}})                     // 20
appraise(3.5)                                              // 0
```

## Constraints

- The tests also pass in loot types of their own that implement `Valuer`.
