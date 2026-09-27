---
title: Live Ordered Map
difficulty: hard
after: generic-data-structures
hints:
  - 'A slice of entries plus a `map[K]int` of positions gives you O(1) `Set` and `Get`, but removing from the middle of a slice is O(n). Instead, leave a **tombstone**: mark the entry dead, remove the key from the position map, and count the dead entries.'
  - 'Iterate by index, re-reading `len(m.entries)` on every step and skipping dead entries. New keys are appended, so the loop reaches them; deleted keys are dead by the time the loop gets there; updated values are read fresh. Delete-then-Set appends a new entry, which moves the key to the end.'
  - 'Tombstones pile up, so compact now and then: when more than half the entries are dead, rebuild the slice with only the live ones and fix every position in the map. Never compact while an iteration is running, because it would shift the indexes under it. Count active iterators (`m.iterating++` and `defer func() { m.iterating-- }()` so `break` is handled too), and compact from `Delete` or at the end of the last iteration.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"iter"
    )

    type OrderedMap[K comparable, V any] struct {
    	// your fields here
    }

    func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
    	return &OrderedMap[K, V]{}
    }

    func (m *OrderedMap[K, V]) Set(k K, v V) {
    }

    func (m *OrderedMap[K, V]) Get(k K) (V, bool) {
    	var zero V
    	return zero, false
    }

    func (m *OrderedMap[K, V]) Delete(k K) bool {
    	return false
    }

    func (m *OrderedMap[K, V]) Len() int {
    	return 0
    }

    func (m *OrderedMap[K, V]) All() iter.Seq2[K, V] {
    	return func(yield func(K, V) bool) {}
    }

    func main() {
    	m := NewOrderedMap[string, int]()
    	m.Set("a", 1)
    	m.Set("b", 2)
    	m.Set("c", 3)
    	m.Set("d", 4)
    	for k, v := range m.All() {
    		fmt.Print(k, "=", v, " ")
    		if k == "a" {
    			m.Delete("c")  // not visited yet: skipped
    			m.Set("e", 5)  // new key: visited at the end
    			m.Set("b", 20) // updated before we get there: 20 is seen
    		}
    	}
    	fmt.Println()
    	fmt.Println("len:", m.Len())
    	// want:
    	// a=1 b=20 d=4 e=5
    	// len: 4
    }
  solution: |
    package main

    import (
    	"fmt"
    	"iter"
    )

    type entry[K comparable, V any] struct {
    	key  K
    	val  V
    	dead bool
    }

    // OrderedMap is a map that remembers insertion order and can be changed
    // while it's being iterated.
    type OrderedMap[K comparable, V any] struct {
    	entries   []entry[K, V] // insertion order, with tombstones
    	pos       map[K]int     // key -> index in entries (live keys only)
    	dead      int           // number of tombstones in entries
    	iterating int           // active All() loops
    }

    // NewOrderedMap returns an empty map.
    func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
    	return &OrderedMap[K, V]{pos: make(map[K]int)}
    }

    // Set stores v under k. A new key goes to the end; an existing key keeps
    // its place.
    func (m *OrderedMap[K, V]) Set(k K, v V) {
    	if i, ok := m.pos[k]; ok {
    		m.entries[i].val = v
    		return
    	}
    	m.pos[k] = len(m.entries)
    	m.entries = append(m.entries, entry[K, V]{key: k, val: v})
    }

    // Get returns the value stored under k.
    func (m *OrderedMap[K, V]) Get(k K) (V, bool) {
    	if i, ok := m.pos[k]; ok {
    		return m.entries[i].val, true
    	}
    	var zero V
    	return zero, false
    }

    // Delete removes k and reports whether it was present.
    func (m *OrderedMap[K, V]) Delete(k K) bool {
    	i, ok := m.pos[k]
    	if !ok {
    		return false
    	}
    	delete(m.pos, k)
    	var zero entry[K, V]
    	m.entries[i] = zero // drop references so the GC can reclaim them
    	m.entries[i].dead = true
    	m.dead++
    	m.maybeCompact()
    	return true
    }

    // Len returns the number of keys.
    func (m *OrderedMap[K, V]) Len() int { return len(m.pos) }

    func (m *OrderedMap[K, V]) maybeCompact() {
    	if m.iterating > 0 || m.dead <= len(m.entries)/2 {
    		return
    	}
    	live := make([]entry[K, V], 0, len(m.pos))
    	for _, e := range m.entries {
    		if !e.dead {
    			m.pos[e.key] = len(live)
    			live = append(live, e)
    		}
    	}
    	m.entries, m.dead = live, 0
    }

    // All yields every key and value in insertion order. The map may be
    // changed during the loop: deleted keys that haven't been reached are
    // skipped, new keys are visited, and updated values are seen.
    func (m *OrderedMap[K, V]) All() iter.Seq2[K, V] {
    	return func(yield func(K, V) bool) {
    		m.iterating++
    		defer func() {
    			m.iterating--
    			m.maybeCompact()
    		}()
    		for i := 0; i < len(m.entries); i++ {
    			if e := m.entries[i]; !e.dead && !yield(e.key, e.val) {
    				return
    			}
    		}
    	}
    }

    func main() {
    	m := NewOrderedMap[string, int]()
    	m.Set("a", 1)
    	m.Set("b", 2)
    	m.Set("c", 3)
    	m.Set("d", 4)
    	for k, v := range m.All() {
    		fmt.Print(k, "=", v, " ")
    		if k == "a" {
    			m.Delete("c")
    			m.Set("e", 5)
    			m.Set("b", 20)
    		}
    	}
    	fmt.Println()
    	fmt.Println("len:", m.Len())
    }
  tests: |
    package main

    import (
    	"fmt"
    	"slices"
    	"testing"
    	"time"
    )

    type SKU string

    type cell struct{ Row, Col int }

    type kv[K comparable, V any] struct {
    	K K
    	V V
    }

    func items[K comparable, V any](m *OrderedMap[K, V]) []kv[K, V] {
    	var out []kv[K, V]
    	for k, v := range m.All() {
    		out = append(out, kv[K, V]{k, v})
    	}
    	return out
    }

    func keys[K comparable, V any](m *OrderedMap[K, V]) []K {
    	var out []K
    	for k := range m.All() {
    		out = append(out, k)
    	}
    	return out
    }

    func TestBasics(t *testing.T) {
    	m := NewOrderedMap[SKU, int]()
    	if _, ok := m.Get("x"); ok || m.Len() != 0 || len(keys(m)) != 0 {
    		t.Fatalf("a new map should be empty")
    	}
    	m.Set("mug", 3)
    	m.Set("hat", 1)
    	m.Set("pen", 9)
    	m.Set("mug", 4) // update keeps the position
    	if got := items(m); !slices.Equal(got, []kv[SKU, int]{{"mug", 4}, {"hat", 1}, {"pen", 9}}) {
    		t.Errorf("after Set mug, hat, pen, mug again: All() = %v, want [{mug 4} {hat 1} {pen 9}]", got)
    	}
    	if v, ok := m.Get("pen"); v != 9 || !ok {
    		t.Errorf("Get(pen) = %d, %v, want 9, true", v, ok)
    	}
    	if !m.Delete("hat") || m.Delete("hat") || m.Delete("nope") {
    		t.Errorf("Delete should report true for a present key and false otherwise")
    	}
    	if v, ok := m.Get("hat"); v != 0 || ok {
    		t.Errorf("Get(hat) after Delete = %d, %v, want 0, false", v, ok)
    	}
    	m.Set("hat", 2) // re-added: goes to the end
    	if got := keys(m); !slices.Equal(got, []SKU{"mug", "pen", "hat"}) || m.Len() != 3 {
    		t.Errorf("after deleting and re-adding hat: keys = %v (Len %d), want [mug pen hat] (Len 3)", got, m.Len())
    	}
    }

    func TestOtherTypes(t *testing.T) {
    	m := NewOrderedMap[cell, *string]()
    	x, o := "x", "o"
    	m.Set(cell{1, 1}, &x)
    	m.Set(cell{0, 2}, &o)
    	m.Set(cell{2, 0}, nil)
    	if v, ok := m.Get(cell{2, 0}); v != nil || !ok {
    		t.Errorf("Get of a key stored with a nil value = %v, %v, want nil, true", v, ok)
    	}
    	if got := keys(m); !slices.Equal(got, []cell{{1, 1}, {0, 2}, {2, 0}}) {
    		t.Errorf("keys = %v, want [{1 1} {0 2} {2 0}]", got)
    	}
    }

    func TestChangesDuringIteration(t *testing.T) {
    	m := NewOrderedMap[string, int]()
    	for i, k := range []string{"a", "b", "c", "d"} {
    		m.Set(k, i+1)
    	}
    	var seen []string
    	for k, v := range m.All() {
    		seen = append(seen, fmt.Sprintf("%s=%d", k, v))
    		if k == "a" {
    			m.Delete("c")
    			m.Set("e", 5)
    			m.Set("b", 20)
    		}
    	}
    	if want := []string{"a=1", "b=20", "d=4", "e=5"}; !slices.Equal(seen, want) {
    		t.Errorf("example loop saw %v, want %v", seen, want)
    	}

    	m = NewOrderedMap[string, int]()
    	for i := range 6 {
    		m.Set(fmt.Sprint(i), i)
    	}
    	seen = nil
    	for k := range m.All() {
    		seen = append(seen, k)
    		m.Delete(k) // delete the current key
    		if k == "1" {
    			m.Delete("0") // already visited
    			m.Delete("4") // not visited yet
    		}
    	}
    	if want := []string{"0", "1", "2", "3", "5"}; !slices.Equal(seen, want) || m.Len() != 0 {
    		t.Errorf("deleting while iterating saw %v with %d keys left, want %v and 0 left", seen, m.Len(), want)
    	}

    	m = NewOrderedMap[string, int]()
    	m.Set("x", 1)
    	m.Set("y", 2)
    	seen = nil
    	for k, v := range m.All() {
    		seen = append(seen, fmt.Sprintf("%s=%d", k, v))
    		if k == "x" && v == 1 {
    			m.Delete("x")
    			m.Set("x", 100) // re-added: moves to the end and is visited again
    		}
    	}
    	if want := []string{"x=1", "y=2", "x=100"}; !slices.Equal(seen, want) {
    		t.Errorf("delete + re-add during iteration saw %v, want %v", seen, want)
    	}
    }

    func TestLastKeyDeletedThenAdded(t *testing.T) {
    	m := NewOrderedMap[int, string]()
    	m.Set(1, "one")
    	m.Set(2, "two")
    	var seen []int
    	for k := range m.All() {
    		seen = append(seen, k)
    		if k == 2 {
    			m.Delete(2) // the current key is the last one
    			m.Set(3, "three")
    		}
    	}
    	if !slices.Equal(seen, []int{1, 2, 3}) {
    		t.Errorf("deleting the last key while on it and then adding a new one: saw %v, want [1 2 3]", seen)
    	}
    }

    func TestBreakAndNestedLoops(t *testing.T) {
    	m := NewOrderedMap[int, int]()
    	for i := range 10 {
    		m.Set(i, i*i)
    	}
    	n := 0
    	for range m.All() {
    		n++
    		if n == 3 {
    			break
    		}
    	}
    	var pairs []string
    	for a := range m.All() {
    		for b := range m.All() {
    			if b > a {
    				m.Delete(b) // the outer loop must skip these too
    			}
    		}
    		pairs = append(pairs, fmt.Sprint(a))
    	}
    	if !slices.Equal(pairs, []string{"0"}) || m.Len() != 1 {
    		t.Errorf("nested loops deleting keys: outer loop saw %v with %d keys left, want [0] and 1 left", pairs, m.Len())
    	}
    	for i := 20; i < 40; i++ {
    		m.Set(i, i)
    	}
    	for i := 20; i < 38; i++ {
    		m.Delete(i)
    	}
    	if got := keys(m); !slices.Equal(got, []int{0, 38, 39}) {
    		t.Errorf("after more Sets and Deletes, keys = %v, want [0 38 39]", got)
    	}
    }

    func TestLarge(t *testing.T) {
    	n := 200_000
    	start := time.Now()
    	m := NewOrderedMap[int, int]()
    	for i := range n {
    		m.Set(i, i)
    	}
    	for i := range n / 2 {
    		if i%1000 == 0 && time.Since(start) > time.Second {
    			t.Fatalf("still deleting (%d of %d) after a second: Delete must be O(1)", i, n/2)
    		}
    		m.Delete((i * 7919) % n) // delete half the keys, scattered
    	}
    	count, prev, ordered := 0, -1, true
    	for k := range m.All() {
    		count++
    		ordered = ordered && k > prev
    		prev = k
    		m.Delete(k)
    	}
    	if count != n/2 || !ordered || m.Len() != 0 {
    		t.Errorf("large test: iterated %d keys (in order: %v) with %d left, want %d, true, 0", count, ordered, m.Len(), n/2)
    	}

    	// A queue: add at the back, drop from the front, peek at the front.
    	q := NewOrderedMap[int, int]()
    	for i := range n {
    		if i%1000 == 0 && time.Since(start) > time.Second {
    			t.Fatalf("still in the queue workload (step %d of %d) after a second: Delete must be O(1), and dead entries must not pile up", i, n)
    		}
    		q.Set(i, i)
    		if i >= 10 {
    			q.Delete(i - 10)
    		}
    		for k := range q.All() {
    			if want := max(0, i-9); k != want {
    				t.Fatalf("queue: front is %d after step %d, want %d", k, i, want)
    			}
    			break
    		}
    	}
    	if d := time.Since(start); d > time.Second {
    		t.Errorf("200,000 Sets, Deletes and iterations took %v: Delete must be O(1), and dead entries must not pile up", d)
    	}
    }
---

Stash's `OrderedMap` from the lessons remembers insertion order, but it has two
problems in production: deleting a key is O(n), and changing the map inside a
`for range m.All()` loop gives surprising results. Build a version that fixes
both.

Implement `OrderedMap[K, V]`:

- `NewOrderedMap()` returns an empty map.
- `Set(k, v)` stores `v`. A **new** key goes to the end; an existing key keeps
  its place and gets the new value.
- `Get(k)` returns the value and `true`, or the zero value and `false`.
- `Delete(k)` removes `k` and reports whether it was there.
- `Len()` returns the number of keys.
- `All()` yields keys and values in order.

`Set`, `Get` and `Delete` must be **O(1)** (amortized).

## Changing the map during iteration

It's safe to call `Set` and `Delete` inside a `for range m.All()` loop,
and the loop sees the map as it is *now*:

- A key deleted before the loop reaches it is **not** visited.
- A new key added during the loop **is** visited (it's at the end).
- A value updated before the loop reaches it is seen with its new value.
- A key deleted and then set again moves to the end, so it's visited again
  even if it was visited before.

## Example

```go
m := NewOrderedMap[string, int]()
m.Set("a", 1); m.Set("b", 2); m.Set("c", 3); m.Set("d", 4)
for k, v := range m.All() {
	fmt.Print(k, "=", v, " ")
	if k == "a" {
		m.Delete("c")  // not visited yet: skipped
		m.Set("e", 5)  // new key: visited at the end
		m.Set("b", 20) // updated before we get there: 20 is seen
	}
}
// a=1 b=20 d=4 e=5
```

## Constraints

- The hidden tests use named, struct and pointer key and value types, nested
  loops over the same map, `break`, and deleting the key the loop is currently
  on, including the very last one.
- **Performance**: 200,000 `Set`s, scattered `Delete`s, and a queue workload
  that deletes from the front and peeks at the new front 200,000 times, all
  under one second. An O(n) delete, or deleted entries that pile up and are
  scanned on every peek, won't make it.
