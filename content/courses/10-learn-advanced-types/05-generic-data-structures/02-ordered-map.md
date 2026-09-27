---
title: 'Stash: OrderedMap[K, V]'
quiz:
  - question: |
      What does this print?

      ```go
      var m OrderedMap[string, int]
      m.Set("a", 1)
      m.Set("b", 2)
      m.Set("a", 3)
      m.Delete("b")
      m.Set("b", 4)
      for k, v := range m.All() {
      	fmt.Print(k, v, " ")
      }
      ```
    options:
      - text: '`a1 b2 a3 b4`'
      - text: '`a3 b4`'
        correct: true
      - text: '`b4 a3`'
      - text: '`a1 b4`'
    explanation: |
      Updating `"a"` keeps its original position and replaces its value. Deleting `"b"`
      removes it from the order, so setting it again puts it at the end, after `"a"`.
  - question: Why does `All` return `iter.Seq2[K, V]` rather than a `[]K` of keys?
    options:
      - text: Slices can't hold type parameters
      - text: It lets callers `range` over pairs directly, stop early, and pass the sequence to functions like `maps.Collect`, without copying
        correct: true
      - text: '`iter.Seq2` is faster than a slice for every use'
      - text: Maps can only be read through iterators
    explanation: |
      An `iter.Seq2` is a function, so nothing is copied up front, callers can `break` out,
      and the standard library has consumers ready: `maps.Collect`, `maps.Insert` and more.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"iter"
    )

    // OrderedMap is a map that remembers the order keys were first inserted.
    // The zero value is an empty map, ready to use.
    type OrderedMap[K comparable, V any] struct {
    	keys []K
    	vals map[K]V
    }

    // Set stores v under k. A new key goes to the end; an existing key keeps its place.
    func (m *OrderedMap[K, V]) Set(k K, v V) {
    	if m.vals == nil {
    		m.vals = make(map[K]V)
    	}
    	if _, ok := m.vals[k]; !ok {
    		m.keys = append(m.keys, k)
    	}
    	m.vals[k] = v
    }

    // Get returns the value stored under k, if any.
    func (m *OrderedMap[K, V]) Get(k K) (V, bool) {
    	v, ok := m.vals[k]
    	return v, ok
    }

    // Len reports the number of keys.
    func (m *OrderedMap[K, V]) Len() int { return len(m.keys) }

    // All yields key-value pairs in insertion order.
    func (m *OrderedMap[K, V]) All() iter.Seq2[K, V] {
    	return func(yield func(K, V) bool) {
    		for _, k := range m.keys {
    			if !yield(k, m.vals[k]) {
    				return
    			}
    		}
    	}
    }

    // Delete removes k and reports whether it was present.
    func (m *OrderedMap[K, V]) Delete(k K) bool {
    	// ?
    	return false
    }

    // Keys yields the keys in insertion order.
    func (m *OrderedMap[K, V]) Keys() iter.Seq[K] {
    	// ?
    	return func(yield func(K) bool) {}
    }

    // Backward yields key-value pairs from the newest key to the oldest.
    func (m *OrderedMap[K, V]) Backward() iter.Seq2[K, V] {
    	// ?
    	return func(yield func(K, V) bool) {}
    }

    func main() {
    	var cfg OrderedMap[string, int]
    	cfg.Set("port", 8080)
    	cfg.Set("workers", 4)
    	cfg.Set("timeout", 30)
    	fmt.Println(cfg.Delete("workers"), cfg.Delete("nope"))
    	for k := range cfg.Keys() {
    		fmt.Print(k, " ")
    	}
    	fmt.Println()
    	for k, v := range cfg.Backward() {
    		fmt.Print(k, "=", v, " ")
    	}
    	fmt.Println()
    }
  solution: |
    package main

    import (
    	"fmt"
    	"iter"
    	"slices"
    )

    // OrderedMap is a map that remembers the order keys were first inserted.
    // The zero value is an empty map, ready to use.
    type OrderedMap[K comparable, V any] struct {
    	keys []K
    	vals map[K]V
    }

    // Set stores v under k. A new key goes to the end; an existing key keeps its place.
    func (m *OrderedMap[K, V]) Set(k K, v V) {
    	if m.vals == nil {
    		m.vals = make(map[K]V)
    	}
    	if _, ok := m.vals[k]; !ok {
    		m.keys = append(m.keys, k)
    	}
    	m.vals[k] = v
    }

    // Get returns the value stored under k, if any.
    func (m *OrderedMap[K, V]) Get(k K) (V, bool) {
    	v, ok := m.vals[k]
    	return v, ok
    }

    // Len reports the number of keys.
    func (m *OrderedMap[K, V]) Len() int { return len(m.keys) }

    // All yields key-value pairs in insertion order.
    func (m *OrderedMap[K, V]) All() iter.Seq2[K, V] {
    	return func(yield func(K, V) bool) {
    		for _, k := range m.keys {
    			if !yield(k, m.vals[k]) {
    				return
    			}
    		}
    	}
    }

    // Delete removes k and reports whether it was present.
    func (m *OrderedMap[K, V]) Delete(k K) bool {
    	if _, ok := m.vals[k]; !ok {
    		return false
    	}
    	delete(m.vals, k)
    	i := slices.Index(m.keys, k)
    	m.keys = slices.Delete(m.keys, i, i+1)
    	return true
    }

    // Keys yields the keys in insertion order.
    func (m *OrderedMap[K, V]) Keys() iter.Seq[K] {
    	return slices.Values(m.keys)
    }

    // Backward yields key-value pairs from the newest key to the oldest.
    func (m *OrderedMap[K, V]) Backward() iter.Seq2[K, V] {
    	return func(yield func(K, V) bool) {
    		for i := len(m.keys) - 1; i >= 0; i-- {
    			k := m.keys[i]
    			if !yield(k, m.vals[k]) {
    				return
    			}
    		}
    	}
    }

    func main() {
    	var cfg OrderedMap[string, int]
    	cfg.Set("port", 8080)
    	cfg.Set("workers", 4)
    	cfg.Set("timeout", 30)
    	fmt.Println(cfg.Delete("workers"), cfg.Delete("nope"))
    	for k := range cfg.Keys() {
    		fmt.Print(k, " ")
    	}
    	fmt.Println()
    	for k, v := range cfg.Backward() {
    		fmt.Print(k, "=", v, " ")
    	}
    	fmt.Println()
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func build() *OrderedMap[string, int] {
    	m := &OrderedMap[string, int]{}
    	for i, k := range []string{"a", "b", "c", "d"} {
    		m.Set(k, i+1)
    	}
    	return m
    }

    func TestDelete(t *testing.T) {
    	m := build()
    	if !m.Delete("b") {
    		t.Errorf("Delete(%q) = false, want true", "b")
    	}
    	if m.Delete("b") {
    		t.Errorf("second Delete(%q) = true, want false", "b")
    	}
    	if m.Delete("zzz") {
    		t.Errorf("Delete of a missing key = true, want false")
    	}
    	if _, ok := m.Get("b"); ok {
    		t.Errorf("Get(%q) after Delete still finds it", "b")
    	}
    	if m.Len() != 3 {
    		t.Errorf("Len() after one Delete = %d, want 3", m.Len())
    	}
    	var keys []string
    	for k := range m.All() {
    		keys = append(keys, k)
    	}
    	if !slices.Equal(keys, []string{"a", "c", "d"}) {
    		t.Errorf("All() keys after Delete(b) = %v, want [a c d]", keys)
    	}
    	m.Set("b", 99)
    	keys = keys[:0]
    	for k := range m.All() {
    		keys = append(keys, k)
    	}
    	if !slices.Equal(keys, []string{"a", "c", "d", "b"}) {
    		t.Errorf("re-adding a deleted key should put it at the end; got %v, want [a c d b]", keys)
    	}
    	var empty OrderedMap[int, int]
    	if empty.Delete(1) {
    		t.Errorf("Delete on a zero-value OrderedMap = true, want false")
    	}
    }

    func TestKeys(t *testing.T) {
    	m := build()
    	if got := slices.Collect(m.Keys()); !slices.Equal(got, []string{"a", "b", "c", "d"}) {
    		t.Errorf("Keys() = %v, want [a b c d]", got)
    	}
    	var first []string
    	for k := range m.Keys() {
    		first = append(first, k)
    		if len(first) == 2 {
    			break
    		}
    	}
    	if !slices.Equal(first, []string{"a", "b"}) {
    		t.Errorf("breaking out of Keys() after 2 got %v, want [a b]", first)
    	}
    }

    func TestBackward(t *testing.T) {
    	m := build()
    	var keys []string
    	var vals []int
    	for k, v := range m.Backward() {
    		keys = append(keys, k)
    		vals = append(vals, v)
    	}
    	if !slices.Equal(keys, []string{"d", "c", "b", "a"}) || !slices.Equal(vals, []int{4, 3, 2, 1}) {
    		t.Errorf("Backward() = %v %v, want [d c b a] [4 3 2 1]", keys, vals)
    	}
    	n := 0
    	for range m.Backward() {
    		n++
    		if n == 1 {
    			break
    		}
    	}
    	if n != 1 {
    		t.Errorf("Backward() kept going after break")
    	}
    }
---

Go maps iterate in random order on purpose. Most of the time that's fine, but config files, HTTP headers, JSON objects you want to round-trip, and reports all want **insertion order**. Stash's `OrderedMap[K, V]` provides it.

## Two structures, one type

The simplest correct design keeps two things in sync:

- a `map[K]V` for fast lookups, and
- a `[]K` recording the order keys were first added.

```go
package main

import (
	"fmt"
	"iter"
	"maps"
	"slices"
)

// OrderedMap is a map that remembers the order keys were first inserted.
type OrderedMap[K comparable, V any] struct {
	keys []K
	vals map[K]V
}

// Set stores v under k. A new key goes to the end; an existing key keeps its place.
func (m *OrderedMap[K, V]) Set(k K, v V) {
	if m.vals == nil {
		m.vals = make(map[K]V)
	}
	if _, ok := m.vals[k]; !ok {
		m.keys = append(m.keys, k)
	}
	m.vals[k] = v
}

func (m *OrderedMap[K, V]) Get(k K) (V, bool) {
	v, ok := m.vals[k]
	return v, ok
}

func (m *OrderedMap[K, V]) Len() int { return len(m.keys) }

// All yields key-value pairs in insertion order.
func (m *OrderedMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, k := range m.keys {
			if !yield(k, m.vals[k]) {
				return
			}
		}
	}
}

func main() {
	var cfg OrderedMap[string, int]
	cfg.Set("port", 8080)
	cfg.Set("workers", 4)
	cfg.Set("timeout", 30)
	cfg.Set("port", 9090) // update: keeps its place

	for k, v := range cfg.All() {
		fmt.Println(k, v)
	}
	plain := maps.Collect(cfg.All()) // into an ordinary map
	fmt.Println(len(plain), slices.Sorted(maps.Keys(plain)))
}
```

```
port 9090
workers 4
timeout 30
3 [port timeout workers]
```

Notice how naturally it plugs into the standard library: `All()` returns an `iter.Seq2[K, V]`, the same type `maps.All` returns, so `maps.Collect` turns it into a plain map in one call.

## The semantics to decide

An ordered map has a few questions a plain map doesn't:

1. **Does updating a key move it?** Stash says no: `Set` on an existing key only replaces the value. (A "last-modified order" map would move it; that's closer to the LRU cache in the next lesson.)
2. **What does deleting do?** The key leaves the order too, so setting it again appends it at the end.
3. **What does iteration see if you modify the map mid-loop?** Keep it simple and document it: don't. Many Go iterators, including `maps.All` for a plain map, give only loose guarantees in that case.

## The cost of Delete

With a slice of keys, `Delete` has to find the key (`slices.Index`) and close the gap (`slices.Delete`), which is **O(n)**. For config-sized maps that's nothing. If you need fast deletes on big maps, store each key's entry in a doubly linked list and keep `map[K]*node` instead; that's exactly the structure the LRU cache uses next, so you'll see how it works.

`slices.Delete` also zeroes the vacated slot at the end (since Go 1.22), so the old key isn't kept alive by the slice's spare capacity. That's the "clear what you pop" advice from chapter 4, built in.

## Your turn

Finish Stash's `OrderedMap`:

- `Delete(k)` removes the key from both the map and the order, and reports whether it was present. Re-adding a deleted key must put it at the end.
- `Keys()` yields the keys in insertion order. (`slices.Values` already turns a slice into an `iter.Seq`.)
- `Backward()` yields pairs from newest to oldest, and must stop when the loop `break`s.
