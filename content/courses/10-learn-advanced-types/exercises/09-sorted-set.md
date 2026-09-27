---
title: Sorted Set
difficulty: medium
after: generic-data-structures
hints:
  - 'Store the values in one sorted slice with no duplicates. `NewSortedSet` can clone its arguments, then use `slices.Sort` and `slices.Compact`.'
  - '`slices.BinarySearch(s.items, v)` returns the position where `v` is (or would be inserted) and whether it was found. That gives you `Has` in O(log n), and `Add` becomes `slices.Insert` at that position when it wasn''t found.'
  - 'For `Range(lo, hi)`, binary-search for `lo` to find the first index, then walk forward yielding values while they''re `< hi`. Return as soon as `yield` returns `false`.'
exercise:
  starter: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"iter"
    )

    type SortedSet[T cmp.Ordered] struct {
    	// your fields here
    }

    func NewSortedSet[T cmp.Ordered](vs ...T) *SortedSet[T] {
    	return &SortedSet[T]{}
    }

    func (s *SortedSet[T]) Add(v T) bool {
    	return false
    }

    func (s *SortedSet[T]) Has(v T) bool {
    	return false
    }

    func (s *SortedSet[T]) Len() int {
    	return 0
    }

    func (s *SortedSet[T]) All() iter.Seq[T] {
    	return func(yield func(T) bool) {}
    }

    func (s *SortedSet[T]) Range(lo, hi T) iter.Seq[T] {
    	return func(yield func(T) bool) {}
    }

    func main() {
    	s := NewSortedSet(30, 10, 50, 20, 10)
    	fmt.Println(s.Add(40), s.Add(20)) // want true false
    	fmt.Println(s.Len(), s.Has(40))   // want 5 true
    	for v := range s.All() {
    		fmt.Print(v, " ") // want 10 20 30 40 50
    	}
    	fmt.Println()
    	for v := range s.Range(15, 40) {
    		fmt.Print(v, " ") // want 20 30
    	}
    	fmt.Println()
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"iter"
    	"slices"
    )

    // SortedSet keeps distinct values in ascending order.
    type SortedSet[T cmp.Ordered] struct {
    	items []T // sorted, no duplicates
    }

    // NewSortedSet returns a set holding vs.
    func NewSortedSet[T cmp.Ordered](vs ...T) *SortedSet[T] {
    	items := slices.Clone(vs)
    	slices.Sort(items)
    	return &SortedSet[T]{items: slices.Compact(items)}
    }

    // Add inserts v and reports whether it was new.
    func (s *SortedSet[T]) Add(v T) bool {
    	i, found := slices.BinarySearch(s.items, v)
    	if found {
    		return false
    	}
    	s.items = slices.Insert(s.items, i, v)
    	return true
    }

    // Has reports whether v is in the set.
    func (s *SortedSet[T]) Has(v T) bool {
    	_, found := slices.BinarySearch(s.items, v)
    	return found
    }

    // Len returns the number of values in the set.
    func (s *SortedSet[T]) Len() int { return len(s.items) }

    // All yields every value in ascending order.
    func (s *SortedSet[T]) All() iter.Seq[T] {
    	return func(yield func(T) bool) {
    		for _, v := range s.items {
    			if !yield(v) {
    				return
    			}
    		}
    	}
    }

    // Range yields the values v with lo <= v < hi in ascending order.
    func (s *SortedSet[T]) Range(lo, hi T) iter.Seq[T] {
    	return func(yield func(T) bool) {
    		i, _ := slices.BinarySearch(s.items, lo)
    		for ; i < len(s.items) && s.items[i] < hi; i++ {
    			if !yield(s.items[i]) {
    				return
    			}
    		}
    	}
    }

    func main() {
    	s := NewSortedSet(30, 10, 50, 20, 10)
    	fmt.Println(s.Add(40), s.Add(20))
    	fmt.Println(s.Len(), s.Has(40))
    	for v := range s.All() {
    		fmt.Print(v, " ")
    	}
    	fmt.Println()
    	for v := range s.Range(15, 40) {
    		fmt.Print(v, " ")
    	}
    	fmt.Println()
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    	"time"
    )

    type Version string

    type Score float64

    type Priority int8

    func TestNewAndAll(t *testing.T) {
    	in := []int{30, 10, 50, 20, 10, 30}
    	s := NewSortedSet(in...)
    	if got := slices.Collect(s.All()); !slices.Equal(got, []int{10, 20, 30, 50}) {
    		t.Errorf("NewSortedSet(%v).All() = %v, want [10 20 30 50]", in, got)
    	}
    	if s.Len() != 4 {
    		t.Errorf("NewSortedSet(%v).Len() = %d, want 4", in, s.Len())
    	}
    	if !slices.Equal(in, []int{30, 10, 50, 20, 10, 30}) {
    		t.Errorf("NewSortedSet changed its arguments to %v", in)
    	}
    	empty := NewSortedSet[string]()
    	if empty.Len() != 0 || len(slices.Collect(empty.All())) != 0 || empty.Has("") {
    		t.Errorf("an empty set should have Len 0, yield nothing and not Has(\"\")")
    	}
    }

    func TestAddAndHas(t *testing.T) {
    	s := NewSortedSet[Priority]()
    	for _, step := range []struct {
    		v    Priority
    		want bool
    	}{{5, true}, {-3, true}, {5, false}, {0, true}, {127, true}, {-128, true}, {0, false}} {
    		if got := s.Add(step.v); got != step.want {
    			t.Errorf("Add(%d) = %v, want %v", step.v, got, step.want)
    		}
    	}
    	if got := slices.Collect(s.All()); !slices.Equal(got, []Priority{-128, -3, 0, 5, 127}) {
    		t.Errorf("after the Adds, All() = %v, want [-128 -3 0 5 127]", got)
    	}
    	for _, v := range []Priority{-128, -3, 0, 5, 127} {
    		if !s.Has(v) {
    			t.Errorf("Has(%d) = false, want true", v)
    		}
    	}
    	for _, v := range []Priority{-127, 1, 4, 6, 126} {
    		if s.Has(v) {
    			t.Errorf("Has(%d) = true, want false", v)
    		}
    	}
    }

    func TestNamedStringType(t *testing.T) {
    	s := NewSortedSet[Version]("v1.10", "v1.2", "v1.9", "v2.0")
    	s.Add("v1.0")
    	want := []Version{"v1.0", "v1.10", "v1.2", "v1.9", "v2.0"} // byte order, not semver!
    	if got := slices.Collect(s.All()); !slices.Equal(got, want) {
    		t.Errorf("All() = %q, want %q", got, want)
    	}
    	if got := slices.Collect(s.Range("v1.1", "v1.9")); !slices.Equal(got, []Version{"v1.10", "v1.2"}) {
    		t.Errorf("Range(v1.1, v1.9) = %q, want [v1.10 v1.2]", got)
    	}
    }

    func TestRange(t *testing.T) {
    	s := NewSortedSet[Score](1.5, 2, 2.5, 3, 10)
    	tests := []struct {
    		lo, hi Score
    		want   []Score
    	}{
    		{2, 3, []Score{2, 2.5}},
    		{1.9, 3.1, []Score{2, 2.5, 3}},
    		{0, 100, []Score{1.5, 2, 2.5, 3, 10}},
    		{3, 3, nil},
    		{5, 1, nil},
    		{10, 11, []Score{10}},
    		{11, 20, nil},
    		{-5, 1.5, nil},
    		{-5, 1.6, []Score{1.5}},
    	}
    	for _, tt := range tests {
    		got := slices.Collect(s.Range(tt.lo, tt.hi))
    		if !slices.Equal(got, tt.want) {
    			t.Errorf("Range(%v, %v) = %v, want %v", tt.lo, tt.hi, got, tt.want)
    		}
    	}
    }

    func TestIteratorsStopEarly(t *testing.T) {
    	s := NewSortedSet(1, 2, 3, 4, 5, 6)
    	var got []int
    	for v := range s.Range(2, 6) {
    		got = append(got, v)
    		if v == 3 {
    			break
    		}
    	}
    	if !slices.Equal(got, []int{2, 3}) {
    		t.Errorf("breaking out of Range(2, 6) at 3 saw %v, want [2 3]", got)
    	}
    	got = nil
    	for v := range s.All() {
    		got = append(got, v)
    		if len(got) == 2 {
    			break
    		}
    	}
    	if !slices.Equal(got, []int{1, 2}) {
    		t.Errorf("breaking out of All() after 2 values saw %v, want [1 2]", got)
    	}
    }

    func TestLarge(t *testing.T) {
    	n := 200_000
    	vs := make([]int, n)
    	for i := range vs {
    		vs[i] = (i * 7919) % n * 2 // every even number below 2n, shuffled
    	}
    	start := time.Now()
    	s := NewSortedSet(vs...)
    	hits := 0
    	for i := range n {
    		if i%1000 == 0 && time.Since(start) > time.Second {
    			t.Fatalf("still running Has calls (%d of %d) after a second: use binary search, not a scan", i, n)
    		}
    		if s.Has(i) {
    			hits++
    		}
    	}
    	sum := 0
    	for i := range n {
    		for v := range s.Range(2*i, 2*i+5) {
    			sum += v - 2*i
    		}
    	}
    	d := time.Since(start)
    	if s.Len() != n || hits != n/2 || sum != 6*n-10 {
    		t.Errorf("large set: Len() = %d, %d Has hits, range sum %d, want %d, %d, %d", s.Len(), hits, sum, n, n/2, 6*n-10)
    	}
    	if d > time.Second {
    		t.Errorf("building a %d-value set plus %d Has and %d Range calls took %v: use binary search, not a scan", n, n, n, d)
    	}
    }
---

Stash needs an index that answers "which keys fall between these two?" A
**sorted set** keeps distinct values in ascending order, so membership checks
and range queries are both fast.

Implement `SortedSet[T cmp.Ordered]`:

- `NewSortedSet(vs...)` returns a set holding the distinct values in `vs`.
  It must not change the caller's slice.
- `Add(v)` inserts `v` and reports whether it was new.
- `Has(v)` reports whether `v` is in the set, in **O(log n)**.
- `Len()` returns the number of values.
- `All()` yields every value in ascending order.
- `Range(lo, hi)` yields the values `v` with `lo <= v < hi`, ascending, in
  **O(log n + k)** for `k` results. If `lo >= hi` it yields nothing.

Both iterators must stop as soon as the loop body breaks.

## Example

```go
s := NewSortedSet(30, 10, 50, 20, 10)
s.Add(40)                       // true
s.Add(20)                       // false (already there)
slices.Collect(s.All())         // [10 20 30 40 50]
slices.Collect(s.Range(15, 40)) // [20 30]
```

## Constraints

- The hidden tests use their own ordered types, such as `type Version string`,
  `type Score float64` and `type Priority int8`. Strings sort in byte order.
- No NaN values.
- The performance test builds a set of 200,000 values, then runs 200,000 `Has`
  and 200,000 small `Range` queries under a one-second limit. Scanning the whole
  set each time would take minutes.
