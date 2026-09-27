---
title: 'Stash: Set[T]'
quiz:
  - question: Why does Stash's `Set` use `map[T]struct{}` rather than `map[T]bool`?
    options:
      - text: '`map[T]bool` doesn''t compile for generic `T`'
      - text: '`struct{}` takes zero bytes, and there''s no confusing third state like a key mapped to `false`'
        correct: true
      - text: '`struct{}` values are faster to hash'
      - text: Maps with `bool` values can't be ranged over
    explanation: |
      An empty struct has size zero, so the map stores only keys. It also means
      membership is just "is the key present?", with no way to accidentally store
      `s.m[v] = false` and have it half-count as a member.
  - question: |
      Why is `Set`'s element type constrained by `comparable` and not `any`?
    options:
      - text: It's a style convention only
      - text: Because the values are map keys, and map keys must support `==`
        correct: true
      - text: So the set can be sorted
      - text: So `Set` can implement `fmt.Stringer`
    explanation: |
      The constraint follows from the implementation: `map[T]struct{}` requires `T` to be
      comparable. (Remember that since Go 1.20 that includes interface types, which can
      panic if their dynamic values aren't comparable.)
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"iter"
    	"slices"
    )

    // Set is an unordered collection of unique values.
    // The zero value is an empty set, ready to use.
    type Set[T comparable] struct {
    	m map[T]struct{}
    }

    // NewSet returns a set containing vs.
    func NewSet[T comparable](vs ...T) *Set[T] {
    	s := &Set[T]{}
    	s.Add(vs...)
    	return s
    }

    func (s *Set[T]) Add(vs ...T) {
    	if s.m == nil {
    		s.m = make(map[T]struct{}, len(vs))
    	}
    	for _, v := range vs {
    		s.m[v] = struct{}{}
    	}
    }

    func (s *Set[T]) Has(v T) bool {
    	_, ok := s.m[v]
    	return ok
    }

    func (s *Set[T]) Remove(v T) { delete(s.m, v) }

    func (s *Set[T]) Len() int { return len(s.m) }

    // All returns an iterator over the set's values, in no particular order.
    func (s *Set[T]) All() iter.Seq[T] {
    	return func(yield func(T) bool) {
    		for v := range s.m {
    			if !yield(v) {
    				return
    			}
    		}
    	}
    }

    // Union returns a new set with every value in a or b.
    func Union[T comparable](a, b *Set[T]) *Set[T] {
    	// ?
    	return &Set[T]{}
    }

    // Intersect returns a new set with the values in both a and b.
    func Intersect[T comparable](a, b *Set[T]) *Set[T] {
    	// ?
    	return &Set[T]{}
    }

    // Difference returns a new set with the values in a that aren't in b.
    func Difference[T comparable](a, b *Set[T]) *Set[T] {
    	// ?
    	return &Set[T]{}
    }

    // SubsetOf reports whether every value of s is also in other.
    func (s *Set[T]) SubsetOf(other *Set[T]) bool {
    	// ?
    	return false
    }

    func main() {
    	backend := NewSet("go", "sql", "http")
    	frontend := NewSet("http", "css", "js")
    	fmt.Println(slices.Sorted(Union(backend, frontend).All()))
    	fmt.Println(slices.Sorted(Intersect(backend, frontend).All()))
    	fmt.Println(slices.Sorted(Difference(backend, frontend).All()))
    	fmt.Println(NewSet("go").SubsetOf(backend), frontend.SubsetOf(backend))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"iter"
    	"slices"
    )

    // Set is an unordered collection of unique values.
    // The zero value is an empty set, ready to use.
    type Set[T comparable] struct {
    	m map[T]struct{}
    }

    // NewSet returns a set containing vs.
    func NewSet[T comparable](vs ...T) *Set[T] {
    	s := &Set[T]{}
    	s.Add(vs...)
    	return s
    }

    func (s *Set[T]) Add(vs ...T) {
    	if s.m == nil {
    		s.m = make(map[T]struct{}, len(vs))
    	}
    	for _, v := range vs {
    		s.m[v] = struct{}{}
    	}
    }

    func (s *Set[T]) Has(v T) bool {
    	_, ok := s.m[v]
    	return ok
    }

    func (s *Set[T]) Remove(v T) { delete(s.m, v) }

    func (s *Set[T]) Len() int { return len(s.m) }

    // All returns an iterator over the set's values, in no particular order.
    func (s *Set[T]) All() iter.Seq[T] {
    	return func(yield func(T) bool) {
    		for v := range s.m {
    			if !yield(v) {
    				return
    			}
    		}
    	}
    }

    // Union returns a new set with every value in a or b.
    func Union[T comparable](a, b *Set[T]) *Set[T] {
    	out := &Set[T]{}
    	for v := range a.All() {
    		out.Add(v)
    	}
    	for v := range b.All() {
    		out.Add(v)
    	}
    	return out
    }

    // Intersect returns a new set with the values in both a and b.
    func Intersect[T comparable](a, b *Set[T]) *Set[T] {
    	if a.Len() > b.Len() {
    		a, b = b, a // loop over the smaller set
    	}
    	out := &Set[T]{}
    	for v := range a.All() {
    		if b.Has(v) {
    			out.Add(v)
    		}
    	}
    	return out
    }

    // Difference returns a new set with the values in a that aren't in b.
    func Difference[T comparable](a, b *Set[T]) *Set[T] {
    	out := &Set[T]{}
    	for v := range a.All() {
    		if !b.Has(v) {
    			out.Add(v)
    		}
    	}
    	return out
    }

    // SubsetOf reports whether every value of s is also in other.
    func (s *Set[T]) SubsetOf(other *Set[T]) bool {
    	if s.Len() > other.Len() {
    		return false
    	}
    	for v := range s.All() {
    		if !other.Has(v) {
    			return false
    		}
    	}
    	return true
    }

    func main() {
    	backend := NewSet("go", "sql", "http")
    	frontend := NewSet("http", "css", "js")
    	fmt.Println(slices.Sorted(Union(backend, frontend).All()))
    	fmt.Println(slices.Sorted(Intersect(backend, frontend).All()))
    	fmt.Println(slices.Sorted(Difference(backend, frontend).All()))
    	fmt.Println(NewSet("go").SubsetOf(backend), frontend.SubsetOf(backend))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func sorted(s *Set[int]) []int { return slices.Sorted(s.All()) }

    func TestUnion(t *testing.T) {
    	a, b := NewSet(1, 2, 3), NewSet(3, 4)
    	if got := sorted(Union(a, b)); !slices.Equal(got, []int{1, 2, 3, 4}) {
    		t.Errorf("Union({1 2 3}, {3 4}) = %v, want [1 2 3 4]", got)
    	}
    	if got := sorted(a); !slices.Equal(got, []int{1, 2, 3}) {
    		t.Errorf("Union must not modify its inputs: a is now %v", got)
    	}
    	var empty Set[int]
    	if got := sorted(Union(&empty, &empty)); len(got) != 0 {
    		t.Errorf("Union of two empty sets = %v, want []", got)
    	}
    }

    func TestIntersect(t *testing.T) {
    	a, b := NewSet(1, 2, 3, 4), NewSet(3, 4, 5)
    	if got := sorted(Intersect(a, b)); !slices.Equal(got, []int{3, 4}) {
    		t.Errorf("Intersect({1 2 3 4}, {3 4 5}) = %v, want [3 4]", got)
    	}
    	if got := sorted(Intersect(b, a)); !slices.Equal(got, []int{3, 4}) {
    		t.Errorf("Intersect({3 4 5}, {1 2 3 4}) = %v, want [3 4]", got)
    	}
    	if got := sorted(Intersect(a, NewSet(9))); len(got) != 0 {
    		t.Errorf("Intersect with a disjoint set = %v, want []", got)
    	}
    }

    func TestDifference(t *testing.T) {
    	a, b := NewSet(1, 2, 3, 4), NewSet(3, 4, 5)
    	if got := sorted(Difference(a, b)); !slices.Equal(got, []int{1, 2}) {
    		t.Errorf("Difference({1 2 3 4}, {3 4 5}) = %v, want [1 2]", got)
    	}
    	if got := sorted(Difference(b, a)); !slices.Equal(got, []int{5}) {
    		t.Errorf("Difference({3 4 5}, {1 2 3 4}) = %v, want [5]", got)
    	}
    }

    func TestSubsetOf(t *testing.T) {
    	big := NewSet("go", "sql", "http")
    	if !NewSet("go", "http").SubsetOf(big) {
    		t.Errorf("{go http} should be a subset of {go sql http}")
    	}
    	if NewSet("go", "css").SubsetOf(big) {
    		t.Errorf("{go css} should not be a subset of {go sql http}")
    	}
    	var empty Set[string]
    	if !empty.SubsetOf(big) {
    		t.Errorf("the empty set is a subset of every set")
    	}
    	if !big.SubsetOf(big) {
    		t.Errorf("a set is a subset of itself")
    	}
    }
---

Time to build Stash for real. First up, the collection Go famously doesn't ship: a **set**.

## Why not just a map?

Plenty of Go code uses `map[string]bool` or `map[string]struct{}` as a set, and for a few lines inside one function that's fine. But the operations get repeated everywhere (`if _, ok := m[k]; ok`, loops for union and intersection), and a bare map says nothing about intent. A small generic type fixes both.

## The design

```go
package main

import (
	"fmt"
	"iter"
	"slices"
)

// Set is an unordered collection of unique values.
// The zero value is an empty set, ready to use.
type Set[T comparable] struct {
	m map[T]struct{}
}

// NewSet returns a set containing vs.
func NewSet[T comparable](vs ...T) *Set[T] {
	s := &Set[T]{}
	s.Add(vs...)
	return s
}

func (s *Set[T]) Add(vs ...T) {
	if s.m == nil {
		s.m = make(map[T]struct{}, len(vs))
	}
	for _, v := range vs {
		s.m[v] = struct{}{}
	}
}

func (s *Set[T]) Has(v T) bool {
	_, ok := s.m[v]
	return ok
}

func (s *Set[T]) Remove(v T) { delete(s.m, v) }

func (s *Set[T]) Len() int { return len(s.m) }

// All returns an iterator over the set's values, in no particular order.
func (s *Set[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range s.m {
			if !yield(v) {
				return
			}
		}
	}
}

func main() {
	tags := NewSet("go", "types", "go")
	tags.Add("generics")
	tags.Remove("types")
	fmt.Println(tags.Len(), tags.Has("go"), tags.Has("types"))
	fmt.Println(slices.Sorted(tags.All()))

	var ids Set[int] // zero value works too
	ids.Add(3, 1, 2)
	fmt.Println(slices.Sorted(ids.All()))
}
```

```
2 true false
[generics go]
[1 2 3]
```

Every decision here comes from earlier chapters:

- **A struct wrapping the map**, not `type Set[T comparable] map[T]struct{}`, so the representation stays private and could change later.
- **`map[T]struct{}`**: zero-byte values, and `comparable` is required because `T` is a map key.
- **Pointer receivers and lazy initialisation** (chapter 4), so `var ids Set[int]` works. `Has`, `Remove` and `Len` are safe on a nil map already: reading, deleting and `len` all work on one.
- **`NewSet(vs ...T)`** infers `T` from its arguments: `NewSet("go")` is a `*Set[string]`. With no arguments you'd write `NewSet[string]()`.
- **`All() iter.Seq[T]`** instead of returning a slice: callers can `range` over it, stop early, or hand it to `slices.Sorted`, `slices.Collect` or any other iterator consumer. Because map order is random, the doc comment says so, and the example sorts before printing.

## Operations between sets

Union, intersection and difference combine two sets symmetrically, which (per chapter 4) makes them read well as **functions**: `Union(a, b)`. "Is `s` a subset of `other`?" reads naturally as a method, `s.SubsetOf(other)`. Stash uses both.

Two small performance points:

- For **intersection**, loop over the *smaller* set and look things up in the bigger one.
- For **subset**, bail out early: a bigger set can't be a subset of a smaller one.

## Your turn

Implement for Stash's `Set`:

- `Union(a, b)`: a **new** set with the values in either.
- `Intersect(a, b)`: a new set with the values in both.
- `Difference(a, b)`: a new set with the values in `a` but not in `b`.
- `(s) SubsetOf(other)`: whether every value in `s` is in `other`. The empty set is a subset of everything.

None of them may modify their inputs. Use `All()` to loop, and remember that a zero-value `Set` is empty, not broken.
