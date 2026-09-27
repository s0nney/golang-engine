---
title: Loot Table
difficulty: medium
after: generics-and-oop
hints:
  - 'Store a slice of a small generic struct, `type entry[T any] struct { item T; weight int }`, plus the running total. The zero value (nil slice, total 0) is then an empty table for free.'
  - 'For `Pick`, walk the entries and subtract each weight from `roll` until `roll < weight`. A negative roll, or one still left over after the last entry, is out of range: return `var zero T` and `false`.'
  - '`Convert` is a Go 1.27 **generic method**: `func (t *LootTable[T]) Convert[U any](f func(T) U) *LootTable[U]`. Build a fresh `LootTable[U]` with its own slice, so adding to it can''t touch the original. For `All`, return a `func(yield func(T, int) bool)` and stop as soon as `yield` returns false.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"iter"
    )

    var ErrBadWeight = errors.New("weight must be at least 1")

    type LootTable[T any] struct {
    }

    func (t *LootTable[T]) Add(item T, weight int) error {
    	return nil
    }

    func (t *LootTable[T]) Total() int {
    	return 0
    }

    func (t *LootTable[T]) Pick(roll int) (T, bool) {
    	var zero T
    	return zero, false
    }

    func (t *LootTable[T]) All() iter.Seq2[T, int] {
    	return func(yield func(T, int) bool) {}
    }

    func (t *LootTable[T]) Convert[U any](f func(T) U) *LootTable[U] {
    	return &LootTable[U]{}
    }

    func main() {
    	var table LootTable[string]
    	table.Add("copper coin", 60)
    	table.Add("healing potion", 30)
    	table.Add("dragon scale", 10)
    	fmt.Println(table.Total())               // want 100
    	fmt.Println(table.Pick(75))              // want healing potion true
    	fmt.Println(table.Add("cursed ring", 0)) // want an error

    	lengths := table.Convert(func(s string) int { return len(s) })
    	for n, w := range lengths.All() {
    		fmt.Println(n, w) // want 11 60, then 14 30, then 12 10
    	}
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"iter"
    )

    var ErrBadWeight = errors.New("weight must be at least 1")

    type entry[T any] struct {
    	item   T
    	weight int
    }

    // LootTable picks items with probability proportional to their weight.
    // The zero LootTable is an empty table, ready to use.
    type LootTable[T any] struct {
    	entries []entry[T]
    	total   int
    }

    // Add appends item with the given weight.
    func (t *LootTable[T]) Add(item T, weight int) error {
    	if weight < 1 {
    		return fmt.Errorf("Add with weight %d: %w", weight, ErrBadWeight)
    	}
    	t.entries = append(t.entries, entry[T]{item, weight})
    	t.total += weight
    	return nil
    }

    // Total returns the sum of all weights.
    func (t *LootTable[T]) Total() int {
    	return t.total
    }

    // Pick returns the item whose slice of [0, Total) contains roll.
    func (t *LootTable[T]) Pick(roll int) (T, bool) {
    	if roll >= 0 {
    		for _, e := range t.entries {
    			if roll < e.weight {
    				return e.item, true
    			}
    			roll -= e.weight
    		}
    	}
    	var zero T
    	return zero, false
    }

    // All yields every item with its weight, in the order they were added.
    func (t *LootTable[T]) All() iter.Seq2[T, int] {
    	return func(yield func(T, int) bool) {
    		for _, e := range t.entries {
    			if !yield(e.item, e.weight) {
    				return
    			}
    		}
    	}
    }

    // Convert returns a new table with f applied to every item and the same weights.
    func (t *LootTable[T]) Convert[U any](f func(T) U) *LootTable[U] {
    	out := &LootTable[U]{total: t.total}
    	for _, e := range t.entries {
    		out.entries = append(out.entries, entry[U]{f(e.item), e.weight})
    	}
    	return out
    }

    func main() {
    	var table LootTable[string]
    	table.Add("copper coin", 60)
    	table.Add("healing potion", 30)
    	table.Add("dragon scale", 10)
    	fmt.Println(table.Total())  // 100
    	fmt.Println(table.Pick(0))  // copper coin true
    	fmt.Println(table.Pick(75)) // healing potion true
    	fmt.Println(table.Pick(99)) // dragon scale true
    	fmt.Println(table.Add("cursed ring", 0))

    	lengths := table.Convert(func(s string) int { return len(s) })
    	for n, w := range lengths.All() {
    		fmt.Println(n, w)
    	}
    }
  tests: |
    package main

    import (
    	"errors"
    	"fmt"
    	"strings"
    	"testing"
    )

    type Drop struct {
    	Name string
    	Qty  int
    }

    // dump lists a table's contents as "item×weight" using All.
    func dump[T any](t *LootTable[T]) string {
    	var parts []string
    	for item, w := range t.All() {
    		parts = append(parts, fmt.Sprintf("%v×%d", item, w))
    	}
    	return "[" + strings.Join(parts, " ") + "]"
    }

    func chest() *LootTable[string] {
    	t := &LootTable[string]{}
    	t.Add("coin", 60)
    	t.Add("potion", 30)
    	t.Add("scale", 10)
    	return t
    }

    func TestZeroTable(t *testing.T) {
    	var table LootTable[Drop]
    	got, ok := table.Pick(0)
    	if table.Total() != 0 || ok || got != (Drop{}) || dump(&table) != "[]" {
    		t.Errorf("zero LootTable: Total() = %d, Pick(0) = %v, %v, All = %s; want 0, {} false, []", table.Total(), got, ok, dump(&table))
    	}
    }

    func TestPick(t *testing.T) {
    	table := chest()
    	if table.Total() != 100 {
    		t.Errorf("Total() = %d, want 100", table.Total())
    	}
    	tests := []struct {
    		roll int
    		want string
    		ok   bool
    	}{
    		{0, "coin", true}, {59, "coin", true}, {60, "potion", true}, {89, "potion", true},
    		{90, "scale", true}, {99, "scale", true}, {100, "", false}, {-1, "", false}, {5000, "", false},
    	}
    	for _, tt := range tests {
    		got, ok := table.Pick(tt.roll)
    		if got != tt.want || ok != tt.ok {
    			t.Errorf("table [coin×60 potion×30 scale×10]: Pick(%d) = %q, %v, want %q, %v", tt.roll, got, ok, tt.want, tt.ok)
    		}
    	}
    }

    func TestEveryRoll(t *testing.T) {
    	var table LootTable[Drop]
    	table.Add(Drop{"arrow", 5}, 7)
    	table.Add(Drop{"gem", 1}, 1)
    	table.Add(Drop{"arrow", 5}, 2) // the same item may appear twice
    	table.Add(Drop{"map", 1}, 4)
    	counts := map[Drop]int{}
    	for roll := range table.Total() {
    		d, ok := table.Pick(roll)
    		if !ok {
    			t.Fatalf("Pick(%d) returned false with Total() = %d", roll, table.Total())
    		}
    		counts[d]++
    	}
    	want := map[Drop]int{{"arrow", 5}: 9, {"gem", 1}: 1, {"map", 1}: 4}
    	if fmt.Sprint(counts) != fmt.Sprint(want) {
    		t.Errorf("picking every roll from 0 to Total()-1 gave %v, want %v", counts, want)
    	}
    }

    func TestAddRejectsBadWeights(t *testing.T) {
    	table := chest()
    	for _, w := range []int{0, -3} {
    		if err := table.Add("cursed ring", w); !errors.Is(err, ErrBadWeight) {
    			t.Errorf("Add(\"cursed ring\", %d) = %v, want an error wrapping ErrBadWeight", w, err)
    		}
    	}
    	if err := table.Add("feather", 1); err != nil {
    		t.Errorf("Add(\"feather\", 1) = %v, want nil", err)
    	}
    	if got, want := dump(table), "[coin×60 potion×30 scale×10 feather×1]"; got != want || table.Total() != 101 {
    		t.Errorf("after rejected adds and one good one: All = %s, Total() = %d, want %s and 101", got, table.Total(), want)
    	}
    }

    func TestAllStopsEarly(t *testing.T) {
    	var seen []string
    	for item := range chest().All() {
    		seen = append(seen, item)
    		if len(seen) == 2 {
    			break // All must stop when yield returns false
    		}
    	}
    	if len(seen) != 2 || seen[0] != "coin" || seen[1] != "potion" {
    		t.Errorf("breaking out of All after two items saw %q, want [coin potion]", seen)
    	}
    }

    func TestConvert(t *testing.T) {
    	table := chest()
    	var calls []string
    	drops := table.Convert(func(name string) Drop {
    		calls = append(calls, name)
    		return Drop{strings.ToUpper(name), len(name)}
    	})
    	if got, want := dump(drops), "[{COIN 4}×60 {POTION 6}×30 {SCALE 5}×10]"; got != want {
    		t.Errorf("Convert: All = %s, want %s", got, want)
    	}
    	if drops.Total() != 100 {
    		t.Errorf("Convert: Total() = %d, want 100", drops.Total())
    	}
    	if d, ok := drops.Pick(95); !ok || d.Name != "SCALE" {
    		t.Errorf("converted table: Pick(95) = %v, %v, want {SCALE 5}, true", d, ok)
    	}
    	if strings.Join(calls, ",") != "coin,potion,scale" {
    		t.Errorf("Convert called f with %q, want each item once, in order", calls)
    	}

    	drops.Add(Drop{"EGG", 1}, 900)
    	if table.Total() != 100 || dump(table) != "[coin×60 potion×30 scale×10]" {
    		t.Errorf("adding to the converted table changed the original: %s (Total %d)", dump(table), table.Total())
    	}

    	var empty LootTable[int]
    	words := empty.Convert(func(n int) string { return fmt.Sprint(n) })
    	if words.Total() != 0 || dump(words) != "[]" {
    		t.Errorf("converting an empty table gave %s with Total() = %d, want an empty table", dump(words), words.Total())
    	}
    	words.Add("usable", 2)
    	if w, _ := words.Pick(1); w != "usable" {
    		t.Errorf("the table Convert returned can't be added to: Pick(1) = %q, want \"usable\"", w)
    	}
    }
---

When a monster dies, the game rolls on its **loot table**: each entry has a
weight, and heavier entries drop more often. The same logic serves gold, items,
whole enemy spawns... so write it once as a generic type, `LootTable[T]`.

- The zero `LootTable[T]` is an empty table, ready to use.
- `Add(item, weight)` appends an entry. A `weight` below 1 returns an error
  wrapping `ErrBadWeight` and leaves the table unchanged. The same item may be
  added more than once.
- `Total()` returns the sum of all weights.
- `Pick(roll)` turns a roll in `[0, Total())` into an item: the first entry owns
  rolls `0` to `weight-1`, the next entry the following `weight` rolls, and so
  on. For a roll outside that range it returns the zero `T` and `false`. (The
  caller does the random rolling, which keeps `Pick` testable.)
- `All()` returns an `iter.Seq2[T, int]` over the items and their weights, in
  the order they were added.
- `Convert(f)` is a **generic method** (new in Go 1.27): it returns a brand-new
  `*LootTable[U]` with `f` applied to every item and the same weights. The two
  tables must be independent afterwards.

## Example

```go
var table LootTable[string]
table.Add("copper coin", 60)    // rolls 0-59
table.Add("healing potion", 30) // rolls 60-89
table.Add("dragon scale", 10)   // rolls 90-99

table.Total()               // 100
table.Pick(75)              // "healing potion", true
table.Pick(100)             // "", false
table.Add("cursed ring", 0) // error wrapping ErrBadWeight

lengths := table.Convert(func(s string) int { return len(s) }) // a *LootTable[int]
for n, w := range lengths.All() {
	fmt.Println(n, w) // 11 60, then 14 30, then 12 10
}
```

## Constraints

- Tables hold a few dozen entries, so a linear scan in `Pick` is fine.
