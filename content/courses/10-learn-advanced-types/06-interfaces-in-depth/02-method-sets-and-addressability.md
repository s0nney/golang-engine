---
title: Method Sets and Addressability
quiz:
  - question: |
      With `func (c *Counter) Inc()`, which line compiles?
    options:
      - text: '`m["hits"].Inc()` where `m` is a `map[string]Counter`'
      - text: '`Counter{}.Inc()`'
      - text: '`s[0].Inc()` where `s` is a `[]Counter`'
        correct: true
      - text: '`getCounter().Inc()` where `getCounter` returns a `Counter`'
    explanation: |
      Calling a pointer method on a value needs its address, so the value must be
      *addressable*. Slice elements are (they live in a backing array). Map elements,
      composite literals and function results aren't, so those lines fail with
      `cannot call pointer method Inc on Counter`.
  - question: Why can't a map element be addressable?
    options:
      - text: Maps are read-only
      - text: The map may move its elements in memory when it grows, which would leave any pointer to an element dangling
        correct: true
      - text: Map values are always interfaces
      - text: It's an arbitrary restriction that may be lifted
    explanation: |
      Maps rearrange their storage as they grow. If `&m[k]` were allowed, that pointer
      could silently point at stale memory. So Go forbids taking the address, and with it
      calling pointer methods directly on map elements.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"slices"
    )

    // Counter counts events.
    type Counter struct {
    	n int
    }

    func (c *Counter) Inc()      { c.n++ }
    func (c *Counter) Add(d int) { c.n += d }
    func (c Counter) Value() int { return c.n }

    // Registry holds named counters. The zero value is ready to use.
    // BUG: counts never go up! Fix Registry so Inc and Add stick.
    type Registry struct {
    	counters map[string]Counter
    }

    // Inc increments the named counter, creating it if needed.
    func (r *Registry) Inc(name string) {
    	if r.counters == nil {
    		r.counters = make(map[string]Counter)
    	}
    	c := r.counters[name]
    	c.Inc()
    }

    // Add adds d to the named counter, creating it if needed.
    func (r *Registry) Add(name string, d int) {
    	if r.counters == nil {
    		r.counters = make(map[string]Counter)
    	}
    	c := r.counters[name]
    	c.Add(d)
    }

    // Value returns the named counter's value, or 0 if it doesn't exist.
    func (r *Registry) Value(name string) int {
    	return r.counters[name].Value()
    }

    // Names returns the counter names in sorted order.
    func (r *Registry) Names() []string {
    	var names []string
    	for name := range r.counters {
    		names = append(names, name)
    	}
    	slices.Sort(names)
    	return names
    }

    func main() {
    	var r Registry
    	r.Inc("hits")
    	r.Inc("hits")
    	r.Add("bytes", 512)
    	fmt.Println(r.Names(), r.Value("hits"), r.Value("bytes"), r.Value("nope"))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"slices"
    )

    // Counter counts events.
    type Counter struct {
    	n int
    }

    func (c *Counter) Inc()      { c.n++ }
    func (c *Counter) Add(d int) { c.n += d }
    func (c Counter) Value() int { return c.n }

    // Registry holds named counters. The zero value is ready to use.
    type Registry struct {
    	counters map[string]*Counter
    }

    // counter returns the named counter, creating it if needed.
    func (r *Registry) counter(name string) *Counter {
    	if r.counters == nil {
    		r.counters = make(map[string]*Counter)
    	}
    	c, ok := r.counters[name]
    	if !ok {
    		c = &Counter{}
    		r.counters[name] = c
    	}
    	return c
    }

    // Inc increments the named counter, creating it if needed.
    func (r *Registry) Inc(name string) { r.counter(name).Inc() }

    // Add adds d to the named counter, creating it if needed.
    func (r *Registry) Add(name string, d int) { r.counter(name).Add(d) }

    // Value returns the named counter's value, or 0 if it doesn't exist.
    func (r *Registry) Value(name string) int {
    	c, ok := r.counters[name]
    	if !ok {
    		return 0
    	}
    	return c.Value()
    }

    // Names returns the counter names in sorted order.
    func (r *Registry) Names() []string {
    	var names []string
    	for name := range r.counters {
    		names = append(names, name)
    	}
    	slices.Sort(names)
    	return names
    }

    func main() {
    	var r Registry
    	r.Inc("hits")
    	r.Inc("hits")
    	r.Add("bytes", 512)
    	fmt.Println(r.Names(), r.Value("hits"), r.Value("bytes"), r.Value("nope"))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func TestRegistryCounts(t *testing.T) {
    	var r Registry
    	for range 3 {
    		r.Inc("hits")
    	}
    	r.Add("bytes", 512)
    	r.Add("bytes", 256)
    	if got := r.Value("hits"); got != 3 {
    		t.Errorf(`after 3 Inc("hits"): Value("hits") = %d, want 3`, got)
    	}
    	if got := r.Value("bytes"); got != 768 {
    		t.Errorf(`after Add("bytes", 512) and Add("bytes", 256): Value("bytes") = %d, want 768`, got)
    	}
    	if got := r.Value("missing"); got != 0 {
    		t.Errorf(`Value("missing") = %d, want 0`, got)
    	}
    	if got := r.Names(); !slices.Equal(got, []string{"bytes", "hits"}) {
    		t.Errorf("Names() = %v, want [bytes hits]", got)
    	}
    }

    func TestZeroRegistryReads(t *testing.T) {
    	var r Registry
    	if got := r.Value("x"); got != 0 {
    		t.Errorf(`zero Registry: Value("x") = %d, want 0`, got)
    	}
    	if got := r.Names(); len(got) != 0 {
    		t.Errorf("zero Registry: Names() = %v, want empty", got)
    	}
    	if got := r.Value("x"); got != 0 || len(r.Names()) != 0 {
    		t.Errorf("Value must not create counters")
    	}
    }
---

[Learn OOP](/courses/learn-oop/types-and-methods/pointer-vs-value-receivers) introduced method sets: the method set of `T` has only value-receiver methods, while `*T` has both. That's what decides interface satisfaction. This lesson digs into the *other* rule underneath it: **addressability**.

## Auto-address, but only when possible

When you call a pointer method on a value, Go quietly takes its address for you:

```go
var c Counter
c.Inc() // really (&c).Inc()
```

That shortcut only works if the value is **addressable**: if it has a stable location in memory that `&` can point at. These are addressable:

- variables: `c`
- pointer dereferences: `*p`
- slice elements: `s[i]` (they live in the backing array)
- fields of addressable structs: `h.stats.hits`, and array elements of addressable arrays

These are **not**:

- map elements: `m[k]`
- function and method results: `getCounter()`
- composite literals: `Counter{}` (though `&Counter{}` is special-cased and allowed)
- values stored inside interfaces
- constants and most other expressions: `a + b`, `string(bs)`

Calling a pointer method on any of them fails with `cannot call pointer method Inc on Counter`.

## Why map elements aren't addressable

A map moves its elements around as it grows. If `&m[k]` were allowed, the pointer could end up pointing at memory the map no longer uses. So Go forbids it, and the "fix" people reach for introduces a subtler bug:

```go
c := m["hits"] // c is a COPY of the element
c.Inc()        // increments the copy; m["hits"] is unchanged
```

It compiles and does nothing. The two real fixes:

1. **Write it back**: `c := m["hits"]; c.Inc(); m["hits"] = c`. Fine for small values.
2. **Store pointers**: `map[string]*Counter`. Then `m["hits"].Inc()` compiles, because a pointer is already an address, and the counter lives at a fixed location outside the map.

## Why interface values aren't addressable

The same reasoning explains the method set rule for interfaces. An interface holds its *own copy* of a value (last lesson). If you could call a pointer method on that copy, you'd modify a hidden copy the caller can't see, which is almost certainly a bug. So Go says a `Counter` value in an interface can only use `Counter`'s value methods, and only a `*Counter` gets the pointer methods. That's why `var _ Incrementer = Counter{}` fails and `var _ Incrementer = &Counter{}` works.

## In generic code

Type parameters follow the same rules. If `T`'s constraint has a pointer method, only pointer types satisfy it, and a `[]T` of those is a slice of pointers (often nil ones: the trap from chapter 4). The `PT interface{ *T; M() }` pattern exists precisely to get addressable `T` values *and* pointer methods at the same time.

## Your turn

Stash's metrics `Registry` stores `Counter` values in a map, and its `Inc` and `Add` modify a copy, so the counts never change. Fix it so that:

- `Inc` and `Add` really update the named counter, creating it on first use;
- `Value` returns 0 for unknown names **without** creating them;
- the zero `Registry` is still ready to use.

Storing `*Counter` in the map is the cleanest fix. A small helper that fetches or creates a counter keeps `Inc` and `Add` to one line each.
