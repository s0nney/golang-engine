---
title: Stacking Satchel
difficulty: hard
after: generics-and-oop
hints:
  - '`Satchel[S Stackable[S]]` knows only two things about its items: `Key()` and `Merge(S) S`. That''s all `Put` needs: if a stack with `item.Key()` exists, replace it with `existing.Merge(item)`; otherwise claim a free slot.'
  - 'Model the slots explicitly so "lowest free slot" is easy: a `[]S` of length `slots`, a parallel `[]bool` saying which are filled, and a `map[string]int` from key to slot index for fast lookups. `Take` clears the slot (set it to `var zero S`) and deletes the key.'
  - '`Transfer` must be all-or-nothing: `Get` the stack, try `dst.Put(item)`, and only `Take` it from the source if that succeeded. Check `dst == s` first: putting a stack into its own satchel would merge it with itself.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"iter"
    )

    // Stackable is an item that can merge with another item of its own type.
    // Items with the same Key share one slot.
    type Stackable[S Stackable[S]] interface {
    	Key() string
    	Merge(other S) S
    }

    var (
    	ErrFull    = errors.New("satchel is full")
    	ErrMissing = errors.New("no such stack")
    )

    type Potion struct {
    	Kind  string
    	Doses int
    }

    func (p Potion) Key() string               { return "" }
    func (p Potion) Merge(other Potion) Potion { return p }

    type Arrows struct {
    	Tip   string
    	Count int
    }

    func (a Arrows) Key() string               { return "" }
    func (a Arrows) Merge(other Arrows) Arrows { return a }

    type Satchel[S Stackable[S]] struct {
    }

    func NewSatchel[S Stackable[S]](slots int) *Satchel[S] {
    	return &Satchel[S]{}
    }

    func (s *Satchel[S]) Put(item S) error {
    	return nil
    }

    func (s *Satchel[S]) Get(key string) (S, bool) {
    	var zero S
    	return zero, false
    }

    func (s *Satchel[S]) Take(key string) (S, bool) {
    	var zero S
    	return zero, false
    }

    func (s *Satchel[S]) Len() int {
    	return 0
    }

    func (s *Satchel[S]) All() iter.Seq[S] {
    	return func(yield func(S) bool) {}
    }

    func (s *Satchel[S]) Transfer(key string, dst *Satchel[S]) error {
    	return nil
    }

    func main() {
    	belt := NewSatchel[Potion](2)
    	belt.Put(Potion{"healing", 2})
    	belt.Put(Potion{"fire", 1})
    	belt.Put(Potion{"healing", 3})            // merges: no new slot needed
    	fmt.Println(belt.Put(Potion{"frost", 1})) // want: satchel is full
    	for p := range belt.All() {
    		fmt.Println(p.Kind, p.Doses) // want healing 5, then fire 1
    	}

    	quiver := NewSatchel[Arrows](3)
    	quiver.Put(Arrows{"steel", 20})
    	quiver.Put(Arrows{"steel", 12})
    	fmt.Println(quiver.Get("steel arrows")) // want {steel 32} true
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"iter"
    )

    // Stackable is an item that can merge with another item of its own type.
    // Items with the same Key share one slot.
    type Stackable[S Stackable[S]] interface {
    	Key() string
    	Merge(other S) S
    }

    var (
    	ErrFull    = errors.New("satchel is full")
    	ErrMissing = errors.New("no such stack")
    )

    type Potion struct {
    	Kind  string
    	Doses int
    }

    func (p Potion) Key() string               { return p.Kind }
    func (p Potion) Merge(other Potion) Potion { return Potion{p.Kind, p.Doses + other.Doses} }

    type Arrows struct {
    	Tip   string
    	Count int
    }

    func (a Arrows) Key() string               { return a.Tip + " arrows" }
    func (a Arrows) Merge(other Arrows) Arrows { return Arrows{a.Tip, a.Count + other.Count} }

    // Satchel holds stacks of one Stackable type in a fixed number of slots.
    type Satchel[S Stackable[S]] struct {
    	slots  []S
    	filled []bool
    	index  map[string]int // key -> slot
    }

    // NewSatchel returns an empty satchel with the given number of slots
    // (a negative number counts as 0).
    func NewSatchel[S Stackable[S]](slots int) *Satchel[S] {
    	slots = max(0, slots)
    	return &Satchel[S]{
    		slots:  make([]S, slots),
    		filled: make([]bool, slots),
    		index:  map[string]int{},
    	}
    }

    // Put merges item into the stack with the same key, or starts a new stack in
    // the lowest free slot. It returns ErrFull if a new stack is needed but no slot
    // is free.
    func (s *Satchel[S]) Put(item S) error {
    	if i, ok := s.index[item.Key()]; ok {
    		s.slots[i] = s.slots[i].Merge(item)
    		return nil
    	}
    	for i, used := range s.filled {
    		if !used {
    			s.slots[i], s.filled[i] = item, true
    			s.index[item.Key()] = i
    			return nil
    		}
    	}
    	return ErrFull
    }

    // Get returns the stack with the given key.
    func (s *Satchel[S]) Get(key string) (S, bool) {
    	if i, ok := s.index[key]; ok {
    		return s.slots[i], true
    	}
    	var zero S
    	return zero, false
    }

    // Take removes and returns the stack with the given key, freeing its slot.
    func (s *Satchel[S]) Take(key string) (S, bool) {
    	i, ok := s.index[key]
    	if !ok {
    		var zero S
    		return zero, false
    	}
    	item := s.slots[i]
    	var zero S
    	s.slots[i], s.filled[i] = zero, false
    	delete(s.index, key)
    	return item, true
    }

    // Len returns the number of stacks in the satchel.
    func (s *Satchel[S]) Len() int {
    	return len(s.index)
    }

    // All yields the stacks in slot order.
    func (s *Satchel[S]) All() iter.Seq[S] {
    	return func(yield func(S) bool) {
    		for i, item := range s.slots {
    			if s.filled[i] && !yield(item) {
    				return
    			}
    		}
    	}
    }

    // Transfer moves the whole stack with the given key into dst. If dst can't
    // take it, both satchels are left unchanged and the error is returned.
    func (s *Satchel[S]) Transfer(key string, dst *Satchel[S]) error {
    	item, ok := s.Get(key)
    	if !ok {
    		return fmt.Errorf("transfer %q: %w", key, ErrMissing)
    	}
    	if dst == s {
    		return nil // already there; Put would merge the stack with itself
    	}
    	if err := dst.Put(item); err != nil {
    		return fmt.Errorf("transfer %q: %w", key, err)
    	}
    	s.Take(key)
    	return nil
    }

    func main() {
    	belt := NewSatchel[Potion](2)
    	belt.Put(Potion{"healing", 2})
    	belt.Put(Potion{"fire", 1})
    	belt.Put(Potion{"healing", 3})            // merges: no new slot needed
    	fmt.Println(belt.Put(Potion{"frost", 1})) // satchel is full
    	for p := range belt.All() {
    		fmt.Println(p.Kind, p.Doses) // healing 5, fire 1
    	}

    	quiver := NewSatchel[Arrows](3)
    	quiver.Put(Arrows{"steel", 20})
    	quiver.Put(Arrows{"steel", 12})
    	fmt.Println(quiver.Get("steel arrows")) // {steel 32} true
    }
  tests: |
    package main

    import (
    	"errors"
    	"fmt"
    	"strings"
    	"testing"
    )

    func contents[S Stackable[S]](s *Satchel[S]) string {
    	var parts []string
    	for item := range s.All() {
    		parts = append(parts, fmt.Sprint(item))
    	}
    	return "[" + strings.Join(parts, " ") + "]"
    }

    func TestItems(t *testing.T) {
    	if got := (Potion{"healing", 2}).Merge(Potion{"healing", 3}); got != (Potion{"healing", 5}) {
    		t.Errorf("Potion{healing 2}.Merge(Potion{healing 3}) = %v, want {healing 5}", got)
    	}
    	if got := (Arrows{"steel", 20}).Merge(Arrows{"steel", 12}); got != (Arrows{"steel", 32}) {
    		t.Errorf("Arrows{steel 20}.Merge(Arrows{steel 12}) = %v, want {steel 32}", got)
    	}
    	if k := (Potion{"fire", 1}).Key(); k != "fire" {
    		t.Errorf("Potion{fire 1}.Key() = %q, want \"fire\"", k)
    	}
    	if k := (Arrows{"steel", 1}).Key(); k != "steel arrows" {
    		t.Errorf("Arrows{steel 1}.Key() = %q, want \"steel arrows\"", k)
    	}
    }

    func TestPutAndGet(t *testing.T) {
    	belt := NewSatchel[Potion](2)
    	steps := []struct {
    		put     Potion
    		wantErr error
    		want    string
    	}{
    		{Potion{"healing", 2}, nil, "[{healing 2}]"},
    		{Potion{"fire", 1}, nil, "[{healing 2} {fire 1}]"},
    		{Potion{"healing", 3}, nil, "[{healing 5} {fire 1}]"},
    		{Potion{"frost", 1}, ErrFull, "[{healing 5} {fire 1}]"},
    		{Potion{"fire", 4}, nil, "[{healing 5} {fire 5}]"},
    	}
    	for _, st := range steps {
    		err := belt.Put(st.put)
    		if !errors.Is(err, st.wantErr) || (st.wantErr == nil && err != nil) {
    			t.Errorf("Put(%v) = %v, want %v", st.put, err, st.wantErr)
    		}
    		if got := contents(belt); got != st.want {
    			t.Errorf("after Put(%v): satchel holds %s, want %s", st.put, got, st.want)
    		}
    	}
    	if belt.Len() != 2 {
    		t.Errorf("Len() = %d, want 2", belt.Len())
    	}
    	if p, ok := belt.Get("healing"); !ok || p != (Potion{"healing", 5}) {
    		t.Errorf("Get(\"healing\") = %v, %v, want {healing 5}, true", p, ok)
    	}
    	if p, ok := belt.Get("frost"); ok || p != (Potion{}) {
    		t.Errorf("Get(\"frost\") = %v, %v, want {}, false", p, ok)
    	}
    }

    func TestTakeFreesLowestSlot(t *testing.T) {
    	quiver := NewSatchel[Arrows](3)
    	for _, a := range []Arrows{{"steel", 10}, {"fire", 5}, {"bone", 8}} {
    		quiver.Put(a)
    	}
    	if a, ok := quiver.Take("fire arrows"); !ok || a != (Arrows{"fire", 5}) {
    		t.Errorf("Take(\"fire arrows\") = %v, %v, want {fire 5}, true", a, ok)
    	}
    	if a, ok := quiver.Take("fire arrows"); ok {
    		t.Errorf("second Take(\"fire arrows\") = %v, true, want false", a)
    	}
    	if quiver.Len() != 2 || contents(quiver) != "[{steel 10} {bone 8}]" {
    		t.Errorf("after Take: Len() = %d, holds %s, want 2 and [{steel 10} {bone 8}]", quiver.Len(), contents(quiver))
    	}
    	quiver.Put(Arrows{"silver", 3}) // goes into the freed middle slot
    	if got, want := contents(quiver), "[{steel 10} {silver 3} {bone 8}]"; got != want {
    		t.Errorf("new stack after a Take: holds %s, want %s (new stacks fill the lowest free slot)", got, want)
    	}
    	quiver.Take("steel arrows")
    	quiver.Put(Arrows{"fire", 1}) // a fresh stack: the old fire arrows are gone
    	if got, want := contents(quiver), "[{fire 1} {silver 3} {bone 8}]"; got != want {
    		t.Errorf("holds %s, want %s", got, want)
    	}
    }

    func TestTinySatchels(t *testing.T) {
    	for _, n := range []int{0, -4} {
    		s := NewSatchel[Potion](n)
    		if err := s.Put(Potion{"healing", 1}); !errors.Is(err, ErrFull) || s.Len() != 0 {
    			t.Errorf("NewSatchel(%d).Put(...) = %v with Len() %d, want ErrFull and 0", n, err, s.Len())
    		}
    	}
    }

    // Rune is the test's own Stackable: runes of the same glyph keep the highest power.
    type Rune struct {
    	Glyph string
    	Power int
    }

    func (r Rune) Key() string { return r.Glyph }
    func (r Rune) Merge(o Rune) Rune {
    	return Rune{r.Glyph, max(r.Power, o.Power)}
    }

    // Ledger is a Stackable pointer type: merging appends entries in place.
    type Ledger struct {
    	Owner   string
    	Entries []int
    }

    func (l *Ledger) Key() string { return l.Owner }
    func (l *Ledger) Merge(o *Ledger) *Ledger {
    	l.Entries = append(l.Entries, o.Entries...)
    	return l
    }

    func TestOwnStackables(t *testing.T) {
    	stones := NewSatchel[Rune](2)
    	stones.Put(Rune{"ᚠ", 3})
    	stones.Put(Rune{"ᚠ", 9})
    	stones.Put(Rune{"ᚠ", 4})
    	if r, _ := stones.Get("ᚠ"); r.Power != 9 {
    		t.Errorf("runes merged to power %d, want 9: Put must use the item's own Merge", r.Power)
    	}

    	books := NewSatchel[*Ledger](1)
    	books.Put(&Ledger{"guild", []int{5}})
    	books.Put(&Ledger{"guild", []int{-2, 7}})
    	if l, ok := books.Get("guild"); !ok || fmt.Sprint(l.Entries) != "[5 -2 7]" {
    		t.Errorf("ledger satchel: Get(\"guild\") = %v, %v, want entries [5 -2 7]", l, ok)
    	}
    	if l, ok := books.Get("thieves"); ok || l != nil {
    		t.Errorf("Get of a missing *Ledger = %v, %v, want nil, false", l, ok)
    	}
    }

    func TestAllStopsEarly(t *testing.T) {
    	s := NewSatchel[Potion](3)
    	s.Put(Potion{"a", 1})
    	s.Put(Potion{"b", 1})
    	s.Put(Potion{"c", 1})
    	n := 0
    	for range s.All() {
    		n++
    		break
    	}
    	if n != 1 {
    		t.Errorf("breaking out of All after one stack ran the loop body %d times", n)
    	}
    }

    func TestTransfer(t *testing.T) {
    	pack := NewSatchel[Potion](3)
    	belt := NewSatchel[Potion](1)
    	pack.Put(Potion{"healing", 4})
    	pack.Put(Potion{"fire", 2})
    	pack.Put(Potion{"frost", 1})

    	if err := pack.Transfer("fire", belt); err != nil {
    		t.Fatalf("Transfer(\"fire\") into an empty belt = %v, want nil", err)
    	}
    	if contents(pack) != "[{healing 4} {frost 1}]" || contents(belt) != "[{fire 2}]" {
    		t.Errorf("after Transfer(\"fire\"): pack %s, belt %s, want [{healing 4} {frost 1}] and [{fire 2}]", contents(pack), contents(belt))
    	}

    	err := pack.Transfer("healing", belt)
    	if !errors.Is(err, ErrFull) {
    		t.Errorf("Transfer(\"healing\") into a full belt = %v, want an error wrapping ErrFull", err)
    	}
    	if contents(pack) != "[{healing 4} {frost 1}]" || contents(belt) != "[{fire 2}]" {
    		t.Errorf("a failed Transfer changed things: pack %s, belt %s; both must be unchanged", contents(pack), contents(belt))
    	}

    	pack.Put(Potion{"fire", 3})
    	if err := pack.Transfer("fire", belt); err != nil || contents(belt) != "[{fire 5}]" || pack.Len() != 2 {
    		t.Errorf("Transfer(\"fire\") onto an existing fire stack in a full belt: err %v, belt %s, pack Len %d; want nil, [{fire 5}], 2", err, contents(belt), pack.Len())
    	}

    	if err := pack.Transfer("gold", belt); !errors.Is(err, ErrMissing) {
    		t.Errorf("Transfer of a missing stack = %v, want an error wrapping ErrMissing", err)
    	}
    	if err := pack.Transfer("healing", pack); err != nil || contents(pack) != "[{healing 4} {frost 1}]" {
    		t.Errorf("Transfer into the same satchel: err %v, holds %s, want nil and no change", err, contents(pack))
    	}
    }
---

Potions stack with potions of the same kind, arrows with arrows of the same
tip. Rather than write a satchel for each, write one generic `Satchel` that
works for **any** item type that knows how to merge with **its own type**:

```go
type Stackable[S Stackable[S]] interface {
	Key() string     // items with the same key share one slot
	Merge(other S) S // combine two items with the same key
}
```

This is a self-referential constraint (Go 1.26+): `Potion` satisfies
`Stackable[Potion]` because its `Merge` takes and returns a `Potion`.

1. Finish `Potion` (key: its `Kind`, merging adds `Doses`) and `Arrows` (key:
   `Tip + " arrows"`, merging adds `Count`).
2. `NewSatchel[S](slots)` returns an empty satchel with that many slots (a
   negative number counts as 0).
3. `Put(item)` merges `item` into the stack with the same key using the item
   type's own `Merge` (the existing stack is the receiver). If there's no such
   stack, it starts a new one in the **lowest-numbered free slot**, or returns
   `ErrFull` if every slot is taken. Merging never needs a free slot.
4. `Get(key)` returns the stack with that key and `true`, or the zero `S` and
   `false`. `Take(key)` does the same but also removes the stack, freeing its
   slot. `Len()` is the number of stacks.
5. `All()` returns an `iter.Seq[S]` over the stacks in slot order.
6. `Transfer(key, dst)` moves a whole stack into another satchel of the same
   type (merging if `dst` already has that key). If the stack doesn't exist,
   return an error wrapping `ErrMissing`. If `dst` can't take it, return an
   error wrapping `ErrFull` and change **neither** satchel. Transferring into
   the same satchel changes nothing and returns `nil`.

## Example

```go
belt := NewSatchel[Potion](2)
belt.Put(Potion{"healing", 2})
belt.Put(Potion{"fire", 1})
belt.Put(Potion{"healing", 3}) // merges into slot 0: {healing 5}
belt.Put(Potion{"frost", 1})   // ErrFull
belt.Take("healing")           // {healing 5}, true: slot 0 is free again
belt.Put(Potion{"frost", 1})   // goes into slot 0
for p := range belt.All() {
	fmt.Println(p) // {frost 1}, then {fire 1}
}
```

## Constraints

- The tests also use their own `Stackable` types, including a pointer type, so
  the satchel may rely on nothing but `Key` and `Merge`.
