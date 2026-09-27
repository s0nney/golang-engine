---
title: Adventurer's Bag
difficulty: medium
after: encapsulation
hints:
  - 'You don''t need to model individual slots. Keep a `map[string]int` of counts: an item with `n` units fills `ceil(n / stackSize)` slots, which is `(n + stackSize - 1) / stackSize` in integer maths.'
  - 'Make `Add` and `Remove` all-or-nothing: validate, then work out the **new** count and the slots it would need, and return an error before changing anything if it doesn''t fit. Only then write to the map.'
  - 'When an item''s count drops to 0, `delete` it from the map, so `Items` never lists it. `Items` should build a fresh slice (`slices.Sorted(maps.Keys(b.counts))` does it in one line), so callers can''t reach your map.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    var (
    	ErrInvalid   = errors.New("invalid item or quantity")
    	ErrBagFull   = errors.New("bag is full")
    	ErrNotEnough = errors.New("not enough of that item")
    )

    type Bag struct {
    }

    func NewBag(slots, stackSize int) (*Bag, error) {
    	return &Bag{}, nil
    }

    func (b *Bag) Add(item string, qty int) error {
    	return nil
    }

    func (b *Bag) Remove(item string, qty int) error {
    	return nil
    }

    func (b *Bag) Count(item string) int {
    	return 0
    }

    func (b *Bag) FreeSlots() int {
    	return 0
    }

    func (b *Bag) Items() []string {
    	return nil
    }

    func main() {
    	bag, err := NewBag(3, 10)
    	fmt.Println(err)                                                         // want <nil>
    	fmt.Println(bag.Add("arrow", 25), bag.FreeSlots())                       // want <nil> 0
    	fmt.Println(bag.Add("potion", 1))                                        // want bag is full
    	fmt.Println(bag.Remove("arrow", 6), bag.Count("arrow"), bag.FreeSlots()) // want <nil> 19 1
    	fmt.Println(bag.Add("potion", 2), bag.Items())                           // want <nil> [arrow potion]
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"maps"
    	"slices"
    )

    var (
    	ErrInvalid   = errors.New("invalid item or quantity")
    	ErrBagFull   = errors.New("bag is full")
    	ErrNotEnough = errors.New("not enough of that item")
    )

    // Bag is an inventory with a fixed number of slots. Each slot holds up to
    // stackSize units of a single item.
    type Bag struct {
    	slots     int
    	stackSize int
    	counts    map[string]int // item -> units; never holds 0
    	used      int            // slots in use
    }

    // NewBag returns an empty bag. Both slots and stackSize must be at least 1.
    func NewBag(slots, stackSize int) (*Bag, error) {
    	if slots < 1 || stackSize < 1 {
    		return nil, fmt.Errorf("NewBag(%d, %d): %w", slots, stackSize, ErrInvalid)
    	}
    	return &Bag{slots: slots, stackSize: stackSize, counts: map[string]int{}}, nil
    }

    // slotsFor returns how many slots n units of one item fill.
    func (b *Bag) slotsFor(n int) int {
    	return (n + b.stackSize - 1) / b.stackSize
    }

    // set changes item's count to n, keeping used in step.
    func (b *Bag) set(item string, n int) {
    	b.used += b.slotsFor(n) - b.slotsFor(b.counts[item])
    	if n == 0 {
    		delete(b.counts, item)
    	} else {
    		b.counts[item] = n
    	}
    }

    // Add puts qty units of item in the bag, or returns ErrBagFull and changes
    // nothing if they don't all fit.
    func (b *Bag) Add(item string, qty int) error {
    	if item == "" || qty <= 0 {
    		return ErrInvalid
    	}
    	old := b.counts[item]
    	n := old + qty
    	if b.used-b.slotsFor(old)+b.slotsFor(n) > b.slots {
    		return ErrBagFull
    	}
    	b.set(item, n)
    	return nil
    }

    // Remove takes qty units of item out of the bag, or returns ErrNotEnough and
    // changes nothing if the bag holds fewer.
    func (b *Bag) Remove(item string, qty int) error {
    	if item == "" || qty <= 0 {
    		return ErrInvalid
    	}
    	if b.counts[item] < qty {
    		return ErrNotEnough
    	}
    	b.set(item, b.counts[item]-qty)
    	return nil
    }

    // Count returns how many units of item the bag holds.
    func (b *Bag) Count(item string) int {
    	return b.counts[item]
    }

    // FreeSlots returns the number of empty slots.
    func (b *Bag) FreeSlots() int {
    	return b.slots - b.used
    }

    // Items returns the names of the items in the bag, sorted.
    func (b *Bag) Items() []string {
    	return slices.Sorted(maps.Keys(b.counts))
    }

    func main() {
    	bag, err := NewBag(3, 10)
    	fmt.Println(err)
    	fmt.Println(bag.Add("arrow", 25), bag.FreeSlots())
    	fmt.Println(bag.Add("potion", 1))
    	fmt.Println(bag.Remove("arrow", 6), bag.Count("arrow"), bag.FreeSlots())
    	fmt.Println(bag.Add("potion", 2), bag.Items())
    }
  tests: |
    package main

    import (
    	"errors"
    	"fmt"
    	"slices"
    	"strconv"
    	"strings"
    	"testing"
    )

    func newBag(t *testing.T, slots, stack int) *Bag {
    	t.Helper()
    	b, err := NewBag(slots, stack)
    	if err != nil || b == nil {
    		t.Fatalf("NewBag(%d, %d) = %v, %v, want a bag and no error", slots, stack, b, err)
    	}
    	return b
    }

    func errName(err error) string {
    	switch {
    	case err == nil:
    		return "ok"
    	case errors.Is(err, ErrBagFull):
    		return "full"
    	case errors.Is(err, ErrNotEnough):
    		return "short"
    	case errors.Is(err, ErrInvalid):
    		return "invalid"
    	}
    	return "unexpected error " + strconv.Quote(err.Error())
    }

    // run applies ops like "+arrow:25" (Add) and "-arrow:6" (Remove) and records
    // each result plus the bag's state afterwards.
    func run(b *Bag, script string) string {
    	var out []string
    	for _, op := range strings.Fields(script) {
    		item, qtyText, _ := strings.Cut(op[1:], ":")
    		qty, _ := strconv.Atoi(qtyText)
    		var err error
    		if op[0] == '+' {
    			err = b.Add(item, qty)
    		} else {
    			err = b.Remove(item, qty)
    		}
    		counts := []string{}
    		for _, it := range b.Items() {
    			counts = append(counts, fmt.Sprintf("%s=%d", it, b.Count(it)))
    		}
    		out = append(out, fmt.Sprintf("%s %v free=%d", errName(err), counts, b.FreeSlots()))
    	}
    	return strings.Join(out, "\n")
    }

    func TestBag(t *testing.T) {
    	tests := []struct {
    		slots, stack int
    		script       string
    		want         []string
    	}{
    		{3, 10, "+arrow:25 +potion:1 -arrow:6 +potion:2", []string{
    			"ok [arrow=25] free=0",
    			"full [arrow=25] free=0",
    			"ok [arrow=19] free=1",
    			"ok [arrow=19 potion=2] free=0",
    		}},
    		{2, 5, "+gem:5 +gem:5 +gem:1 +ore:1", []string{
    			"ok [gem=5] free=1",
    			"ok [gem=10] free=0",
    			"full [gem=10] free=0",
    			"full [gem=10] free=0",
    		}},
    		{2, 5, "+gem:3 +gem:2 -gem:1 +ore:5 +gem:1 +gem:1", []string{
    			"ok [gem=3] free=1",
    			"ok [gem=5] free=1",
    			"ok [gem=4] free=1",
    			"ok [gem=4 ore=5] free=0",
    			"ok [gem=5 ore=5] free=0",
    			"full [gem=5 ore=5] free=0",
    		}},
    		{4, 1, "+sword:1 +shield:1 -sword:1 -sword:1 +helm:3", []string{
    			"ok [sword=1] free=3",
    			"ok [shield=1 sword=1] free=2",
    			"ok [shield=1] free=3",
    			"short [shield=1] free=3",
    			"ok [helm=3 shield=1] free=0",
    		}},
    		{1, 99, "+potion:0 +potion:-4 +:3 -potion:0 +potion:99 -potion:100 -potion:99", []string{
    			"invalid [] free=1",
    			"invalid [] free=1",
    			"invalid [] free=1",
    			"invalid [] free=1",
    			"ok [potion=99] free=0",
    			"short [potion=99] free=0",
    			"ok [] free=1",
    		}},
    		{3, 10, "+rope:31 +rope:30 -ghost:1", []string{
    			"full [] free=3",
    			"ok [rope=30] free=0",
    			"short [rope=30] free=0",
    		}},
    	}
    	for _, tt := range tests {
    		b := newBag(t, tt.slots, tt.stack)
    		if got, want := run(b, tt.script), strings.Join(tt.want, "\n"); got != want {
    			t.Errorf("NewBag(%d, %d) then %q:\n got:\n%s\n want:\n%s", tt.slots, tt.stack, tt.script, got, want)
    		}
    	}
    }

    func TestNewBagRejectsBadSizes(t *testing.T) {
    	for _, tt := range [][2]int{{0, 10}, {3, 0}, {-1, 5}, {2, -2}} {
    		b, err := NewBag(tt[0], tt[1])
    		if !errors.Is(err, ErrInvalid) || b != nil {
    			t.Errorf("NewBag(%d, %d) = %v, %v, want nil and an error wrapping ErrInvalid", tt[0], tt[1], b, err)
    		}
    	}
    }

    func TestItemsIsACopy(t *testing.T) {
    	b := newBag(t, 5, 10)
    	b.Add("torch", 2)
    	b.Add("apple", 3)
    	items := b.Items()
    	if !slices.Equal(items, []string{"apple", "torch"}) {
    		t.Fatalf("Items() = %q, want [apple torch]", items)
    	}
    	items[0] = "dragon egg"
    	if got := b.Items(); !slices.Equal(got, []string{"apple", "torch"}) {
    		t.Errorf("changing the slice from Items() changed the bag: Items() is now %q", got)
    	}
    }

    func TestBagsAreIndependent(t *testing.T) {
    	a, b := newBag(t, 2, 5), newBag(t, 2, 5)
    	a.Add("coin", 10)
    	if b.Count("coin") != 0 || b.FreeSlots() != 2 {
    		t.Errorf("after 10 coins went into bag a, bag b has Count(\"coin\") = %d and FreeSlots() = %d, want 0 and 2", b.Count("coin"), b.FreeSlots())
    	}
    }
---

An adventurer's bag has a fixed number of **slots**. Each slot holds up to
`stackSize` units of **one** item: with a stack size of 10, 25 arrows fill 3
slots (10 + 10 + 5), and a potion needs a slot of its own.

Build `Bag` so it can never get into an impossible state (more items than slots,
negative counts, items with a count of 0) through its methods.

- `NewBag(slots, stackSize)` returns an empty bag. If either number is below 1,
  it returns `nil` and an error wrapping `ErrInvalid`.
- `Add(item, qty)` puts `qty` units of `item` in the bag. Units of the same item
  always stack together as tightly as possible. If they don't **all** fit,
  return `ErrBagFull` and add **none** of them.
- `Remove(item, qty)` takes `qty` units out. If the bag holds fewer, return
  `ErrNotEnough` and remove nothing. Removing frees up slots.
- `Add` and `Remove` return `ErrInvalid` for an empty item name or a `qty` of 0
  or less.
- `Count(item)` returns the units held (0 if none), `FreeSlots()` the empty
  slots, and `Items()` the names of the items in the bag, sorted, as a slice the
  caller may modify freely.

## Example

```go
bag, _ := NewBag(3, 10)
bag.Add("arrow", 25)   // nil: 3 slots used, FreeSlots() == 0
bag.Add("potion", 1)   // ErrBagFull
bag.Remove("arrow", 6) // nil: 19 arrows fit in 2 slots, FreeSlots() == 1
bag.Add("potion", 2)   // nil
bag.Items()            // [arrow potion]
bag.Remove("gem", 1)   // ErrNotEnough
```

## Constraints

- Up to a few thousand operations; any correct approach is fast enough.
- Returned errors are compared with `errors.Is`, so you may wrap them.
