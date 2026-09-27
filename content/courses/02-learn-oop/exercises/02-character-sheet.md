---
title: Character Sheet
difficulty: easy
after: types-and-methods
hints:
  - 'All four methods only **read** `s`, so value receivers (`func (s Stats) ...`) are right. `Add` builds and returns a brand-new `Stats` instead of changing `s`.'
  - 'For `Best`, start with `"STR"` as the answer and only switch when a later stat is **strictly** greater. That gives the tie order STR, DEX, INT for free.'
  - '`String` makes `Stats` a `fmt.Stringer`: once it exists, `fmt.Println(s)` calls it for you. `fmt.Sprintf("STR %d / DEX %d / INT %d", ...)` builds the text.'
exercise:
  starter: |
    package main

    import "fmt"

    // Stats are a hero's three core attributes.
    type Stats struct {
    	STR, DEX, INT int
    }

    // Total returns STR + DEX + INT.
    func (s Stats) Total() int {
    	return 0
    }

    // Best returns the name of the highest stat: "STR", "DEX" or "INT".
    // Ties go to the stat listed first (STR, then DEX, then INT).
    func (s Stats) Best() string {
    	return ""
    }

    // Add returns s with each stat increased by the matching stat in bonus.
    // It must not change s.
    func (s Stats) Add(bonus Stats) Stats {
    	return s
    }

    // String formats s like "STR 12 / DEX 9 / INT 15".
    // Adding this method makes Stats print nicely with fmt.Println.
    func (s Stats) String() string {
    	return "?"
    }

    func main() {
    	base := Stats{STR: 12, DEX: 9, INT: 15}
    	ring := Stats{DEX: 7}
    	geared := base.Add(ring)
    	fmt.Println(base, "total", base.Total(), "best", base.Best())
    	fmt.Println(geared, "total", geared.Total(), "best", geared.Best())
    	// want:
    	// STR 12 / DEX 9 / INT 15 total 36 best INT
    	// STR 12 / DEX 16 / INT 15 total 43 best DEX
    }
  solution: |
    package main

    import "fmt"

    // Stats are a hero's three core attributes.
    type Stats struct {
    	STR, DEX, INT int
    }

    // Total returns STR + DEX + INT.
    func (s Stats) Total() int {
    	return s.STR + s.DEX + s.INT
    }

    // Best returns the name of the highest stat: "STR", "DEX" or "INT".
    // Ties go to the stat listed first (STR, then DEX, then INT).
    func (s Stats) Best() string {
    	best, value := "STR", s.STR
    	if s.DEX > value {
    		best, value = "DEX", s.DEX
    	}
    	if s.INT > value {
    		best = "INT"
    	}
    	return best
    }

    // Add returns s with each stat increased by the matching stat in bonus.
    // It must not change s.
    func (s Stats) Add(bonus Stats) Stats {
    	return Stats{
    		STR: s.STR + bonus.STR,
    		DEX: s.DEX + bonus.DEX,
    		INT: s.INT + bonus.INT,
    	}
    }

    // String formats s like "STR 12 / DEX 9 / INT 15".
    func (s Stats) String() string {
    	return fmt.Sprintf("STR %d / DEX %d / INT %d", s.STR, s.DEX, s.INT)
    }

    func main() {
    	base := Stats{STR: 12, DEX: 9, INT: 15}
    	ring := Stats{DEX: 7}
    	geared := base.Add(ring)
    	fmt.Println(base, "total", base.Total(), "best", base.Best())
    	fmt.Println(geared, "total", geared.Total(), "best", geared.Best())
    }
  tests: |
    package main

    import (
    	"fmt"
    	"testing"
    )

    // lit shows a Stats as a literal, without calling its String method.
    func lit(s Stats) string {
    	return fmt.Sprintf("Stats{STR: %d, DEX: %d, INT: %d}", s.STR, s.DEX, s.INT)
    }

    func TestTotalAndBest(t *testing.T) {
    	tests := []struct {
    		s     Stats
    		total int
    		best  string
    	}{
    		{Stats{12, 9, 15}, 36, "INT"},
    		{Stats{12, 16, 15}, 43, "DEX"},
    		{Stats{18, 3, 4}, 25, "STR"},
    		{Stats{}, 0, "STR"},
    		{Stats{7, 7, 7}, 21, "STR"},
    		{Stats{2, 9, 9}, 20, "DEX"},
    		{Stats{9, 2, 9}, 20, "STR"},
    		{Stats{-3, -1, -2}, -6, "DEX"},
    	}
    	for _, tt := range tests {
    		if got := tt.s.Total(); got != tt.total {
    			t.Errorf("%s.Total() = %d, want %d", lit(tt.s), got, tt.total)
    		}
    		if got := tt.s.Best(); got != tt.best {
    			t.Errorf("%s.Best() = %q, want %q", lit(tt.s), got, tt.best)
    		}
    	}
    }

    func TestAdd(t *testing.T) {
    	base := Stats{12, 9, 15}
    	got := base.Add(Stats{1, 7, -5})
    	if want := (Stats{13, 16, 10}); got != want {
    		t.Errorf("%s.Add(%s) = %s, want %s", lit(base), lit(Stats{1, 7, -5}), lit(got), lit(want))
    	}
    	if base != (Stats{12, 9, 15}) {
    		t.Errorf("Add changed the original: it is now %s, want %s", lit(base), lit(Stats{12, 9, 15}))
    	}
    	if got := base.Add(Stats{}); got != base {
    		t.Errorf("%s.Add(Stats{}) = %s, want it unchanged", lit(base), lit(got))
    	}
    }

    func TestString(t *testing.T) {
    	tests := []struct {
    		s    Stats
    		want string
    	}{
    		{Stats{12, 9, 15}, "STR 12 / DEX 9 / INT 15"},
    		{Stats{}, "STR 0 / DEX 0 / INT 0"},
    		{Stats{100, -2, 3}, "STR 100 / DEX -2 / INT 3"},
    	}
    	for _, tt := range tests {
    		if got := tt.s.String(); got != tt.want {
    			t.Errorf("%s.String() = %q, want %q", lit(tt.s), got, tt.want)
    		} else if got := fmt.Sprint(tt.s); got != tt.want {
    			t.Errorf("fmt.Sprint(stats) = %q, want %q: is String defined on Stats (not *Stats)?", got, tt.want)
    		}
    	}
    }
---

Every hero in the guild ledger has a character sheet with three core stats.
Give the `Stats` type some behaviour:

- `Total()` returns `STR + DEX + INT`.
- `Best()` returns the name of the highest stat: `"STR"`, `"DEX"` or `"INT"`.
  On a tie, the stat that comes first in that list wins.
- `Add(bonus)` returns a **new** `Stats` with each stat raised by the matching
  stat in `bonus` (say, from a magic ring). The original must not change.
- `String()` formats the sheet as `"STR 12 / DEX 9 / INT 15"`. With this method,
  `Stats` satisfies `fmt.Stringer`, so `fmt.Println(s)` prints it that way.

## Example

```go
base := Stats{STR: 12, DEX: 9, INT: 15}
geared := base.Add(Stats{DEX: 7})

fmt.Println(base)   // STR 12 / DEX 9 / INT 15
fmt.Println(geared) // STR 12 / DEX 16 / INT 15
base.Total()        // 36
base.Best()         // "INT"
geared.Best()       // "DEX"
Stats{7, 7, 7}.Best() // "STR" (tie)
```

## Constraints

- Stats can be negative (curses exist).
- None of these methods changes the receiver, so use value receivers.
