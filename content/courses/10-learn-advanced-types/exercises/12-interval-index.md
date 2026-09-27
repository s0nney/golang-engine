---
title: Interval Index
difficulty: hard
after: generics-performance-and-limits
hints:
  - 'Checking every interval on every query is O(n) per query: 100,000 queries over 100,000 intervals is 10 billion checks. Sorting by `Lo` helps you stop early (nothing after the first `Lo > hi` can match), but not skip the start: one long interval that begins early can still overlap `lo`, so you can''t just binary-search for `lo`.'
  - 'Sort a copy by `Lo` with `slices.SortStableFunc` (stable keeps equal `Lo`s in input order). Treat the sorted slice as an implicit balanced search tree: the root of `items[l:r]` is `mid := (l + r) / 2`, its children are `items[l:mid]` and `items[mid+1:r]`. Precompute `maxHi[mid]`, the largest `Hi` anywhere in `items[l:r]`, with one recursive pass.'
  - 'Query the tree in order: for the node `[l, r)`, give up if it''s empty or `maxHi[mid] < lo` (everything inside ends too early). Otherwise search the left half, then stop if `items[mid].Lo > hi` (this node and everything to its right starts too late), yield `items[mid]` if `items[mid].Hi >= lo`, then search the right half. Have the helper return `false` once `yield` does, so a `break` stops the whole search.'
exercise:
  starter: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"iter"
    )

    // Interval covers the keys from Lo to Hi, inclusive, and carries a Value.
    type Interval[K cmp.Ordered, V any] struct {
    	Lo, Hi K
    	Value  V
    }

    type IntervalIndex[K cmp.Ordered, V any] struct {
    	// your fields here
    }

    func NewIntervalIndex[K cmp.Ordered, V any](items []Interval[K, V]) *IntervalIndex[K, V] {
    	return &IntervalIndex[K, V]{}
    }

    func (x *IntervalIndex[K, V]) Len() int {
    	return 0
    }

    func (x *IntervalIndex[K, V]) Overlapping(lo, hi K) iter.Seq[Interval[K, V]] {
    	return func(yield func(Interval[K, V]) bool) {}
    }

    func main() {
    	idx := NewIntervalIndex([]Interval[int, string]{
    		{10, 20, "backup"},
    		{0, 100, "maintenance"},
    		{15, 16, "deploy"},
    		{30, 40, "reindex"},
    	})
    	fmt.Println("intervals:", idx.Len()) // want 4
    	for iv := range idx.Overlapping(18, 32) {
    		fmt.Println(iv.Lo, iv.Hi, iv.Value)
    	}
    	// want:
    	// 0 100 maintenance
    	// 10 20 backup
    	// 30 40 reindex
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"iter"
    	"slices"
    )

    // Interval covers the keys from Lo to Hi, inclusive, and carries a Value.
    type Interval[K cmp.Ordered, V any] struct {
    	Lo, Hi K
    	Value  V
    }

    // IntervalIndex answers "which intervals overlap [lo, hi]?" in
    // O(log n + k)-ish time. It's an implicit balanced search tree over the
    // intervals sorted by Lo, with each subtree's largest Hi precomputed.
    type IntervalIndex[K cmp.Ordered, V any] struct {
    	items []Interval[K, V] // sorted by Lo (stable)
    	maxHi []K              // maxHi[mid] = largest Hi in the subtree rooted at mid
    }

    // NewIntervalIndex builds an index over items, ignoring intervals with Lo > Hi.
    func NewIntervalIndex[K cmp.Ordered, V any](items []Interval[K, V]) *IntervalIndex[K, V] {
    	valid := make([]Interval[K, V], 0, len(items))
    	for _, iv := range items {
    		if iv.Lo <= iv.Hi {
    			valid = append(valid, iv)
    		}
    	}
    	slices.SortStableFunc(valid, func(a, b Interval[K, V]) int { return cmp.Compare(a.Lo, b.Lo) })
    	x := &IntervalIndex[K, V]{items: valid, maxHi: make([]K, len(valid))}
    	if len(valid) > 0 {
    		x.build(0, len(valid))
    	}
    	return x
    }

    func (x *IntervalIndex[K, V]) build(l, r int) K {
    	mid := (l + r) / 2
    	m := x.items[mid].Hi
    	if l < mid {
    		m = max(m, x.build(l, mid))
    	}
    	if mid+1 < r {
    		m = max(m, x.build(mid+1, r))
    	}
    	x.maxHi[mid] = m
    	return m
    }

    // Len returns the number of intervals in the index.
    func (x *IntervalIndex[K, V]) Len() int { return len(x.items) }

    // Overlapping yields the intervals that share at least one key with
    // [lo, hi], ordered by Lo and then by their position in the input.
    func (x *IntervalIndex[K, V]) Overlapping(lo, hi K) iter.Seq[Interval[K, V]] {
    	return func(yield func(Interval[K, V]) bool) {
    		if lo <= hi {
    			x.search(0, len(x.items), lo, hi, yield)
    		}
    	}
    }

    func (x *IntervalIndex[K, V]) search(l, r int, lo, hi K, yield func(Interval[K, V]) bool) bool {
    	if l >= r {
    		return true
    	}
    	mid := (l + r) / 2
    	if x.maxHi[mid] < lo {
    		return true
    	}
    	if !x.search(l, mid, lo, hi, yield) {
    		return false
    	}
    	iv := x.items[mid]
    	if iv.Lo > hi {
    		return true
    	}
    	if iv.Hi >= lo && !yield(iv) {
    		return false
    	}
    	return x.search(mid+1, r, lo, hi, yield)
    }

    func main() {
    	idx := NewIntervalIndex([]Interval[int, string]{
    		{10, 20, "backup"},
    		{0, 100, "maintenance"},
    		{15, 16, "deploy"},
    		{30, 40, "reindex"},
    	})
    	fmt.Println("intervals:", idx.Len())
    	for iv := range idx.Overlapping(18, 32) {
    		fmt.Println(iv.Lo, iv.Hi, iv.Value)
    	}
    }
  tests: |
    package main

    import (
    	"cmp"
    	"iter"
    	"slices"
    	"testing"
    	"time"
    )

    type Timestamp int64

    type Version string

    type Lease struct {
    	Holder string
    }

    func values[K cmp.Ordered, V any](seq iter.Seq[Interval[K, V]]) []V {
    	var out []V
    	for iv := range seq {
    		out = append(out, iv.Value)
    	}
    	return out
    }

    func TestExample(t *testing.T) {
    	in := []Interval[int, string]{
    		{10, 20, "backup"},
    		{0, 100, "maintenance"},
    		{15, 16, "deploy"},
    		{30, 40, "reindex"},
    	}
    	orig := slices.Clone(in)
    	idx := NewIntervalIndex(in)
    	if idx.Len() != 4 {
    		t.Errorf("Len() = %d, want 4", idx.Len())
    	}
    	tests := []struct {
    		lo, hi int
    		want   []string
    	}{
    		{18, 32, []string{"maintenance", "backup", "reindex"}},
    		{15, 15, []string{"maintenance", "backup", "deploy"}},
    		{20, 20, []string{"maintenance", "backup"}},
    		{21, 29, []string{"maintenance"}},
    		{-5, -1, nil},
    		{100, 200, []string{"maintenance"}},
    		{101, 200, nil},
    		{-10, 1000, []string{"maintenance", "backup", "deploy", "reindex"}},
    		{40, 30, nil},
    	}
    	for _, tt := range tests {
    		if got := values(idx.Overlapping(tt.lo, tt.hi)); !slices.Equal(got, tt.want) {
    			t.Errorf("Overlapping(%d, %d) = %q, want %q", tt.lo, tt.hi, got, tt.want)
    		}
    	}
    	if !slices.Equal(in, orig) {
    		t.Errorf("NewIntervalIndex changed its input to %v", in)
    	}
    	in[0].Value = "changed"
    	if got := values(idx.Overlapping(18, 18)); !slices.Equal(got, []string{"maintenance", "backup"}) {
    		t.Errorf("after changing the caller's slice, Overlapping(18, 18) = %q: the index must keep its own copy", got)
    	}
    }

    func TestTiesKeepInputOrderAndInvalidDropped(t *testing.T) {
    	in := []Interval[Timestamp, *Lease]{
    		{5, 9, &Lease{"c"}},
    		{5, 6, &Lease{"a"}},
    		{9, 2, &Lease{"backwards"}},
    		{5, 5, &Lease{"b"}},
    		{1, 4, &Lease{"early"}},
    	}
    	idx := NewIntervalIndex(in)
    	if idx.Len() != 4 {
    		t.Errorf("Len() = %d, want 4 (the Lo > Hi interval is ignored)", idx.Len())
    	}
    	var got []string
    	for iv := range idx.Overlapping(3, 5) {
    		got = append(got, iv.Value.Holder)
    	}
    	if want := []string{"early", "c", "a", "b"}; !slices.Equal(got, want) {
    		t.Errorf("Overlapping(3, 5) holders = %q, want %q (sorted by Lo, ties in input order)", got, want)
    	}
    	for iv := range idx.Overlapping(0, 100) {
    		if iv.Value.Holder == "backwards" {
    			t.Errorf("Overlapping yielded the Lo > Hi interval %v; it should be ignored", iv)
    		}
    	}
    }

    func TestOtherKeyTypes(t *testing.T) {
    	vs := NewIntervalIndex([]Interval[Version, int]{
    		{"v1.0", "v1.4", 1},
    		{"v1.2", "v2.0", 2},
    		{"v2.1", "v3.0", 3},
    	})
    	if got := values(vs.Overlapping("v1.5", "v2.05")); !slices.Equal(got, []int{2}) {
    		t.Errorf("Version index Overlapping(v1.5, v2.05) = %v, want [2]", got)
    	}
    	fs := NewIntervalIndex([]Interval[float64, bool]{{0.5, 1.5, true}, {1.5, 2.5, false}})
    	if got := values(fs.Overlapping(1.5, 1.5)); !slices.Equal(got, []bool{true, false}) {
    		t.Errorf("float index Overlapping(1.5, 1.5) = %v, want [true false] (ends are inclusive)", got)
    	}
    	empty := NewIntervalIndex[int, string](nil)
    	if empty.Len() != 0 || len(values(empty.Overlapping(0, 10))) != 0 {
    		t.Errorf("an empty index should have Len 0 and yield nothing")
    	}
    }

    func TestStopsEarly(t *testing.T) {
    	var in []Interval[int, int]
    	for i := range 100 {
    		in = append(in, Interval[int, int]{i, i + 50, i})
    	}
    	idx := NewIntervalIndex(in)
    	var got []int
    	for iv := range idx.Overlapping(60, 70) {
    		got = append(got, iv.Value)
    		if len(got) == 3 {
    			break
    		}
    	}
    	if !slices.Equal(got, []int{10, 11, 12}) {
    		t.Errorf("breaking after 3 results of Overlapping(60, 70) saw %v, want [10 11 12]", got)
    	}
    	if again := values(idx.Overlapping(60, 70)); len(again) != 61 {
    		t.Errorf("ranging over Overlapping(60, 70) a second time gave %d results, want 61", len(again))
    	}
    }

    func TestAgainstBruteForce(t *testing.T) {
    	var in []Interval[int, int]
    	seed := uint32(7)
    	next := func(n int) int { seed = seed*1664525 + 1013904223; return int(seed>>8) % n }
    	for i := range 500 {
    		lo := next(1000)
    		in = append(in, Interval[int, int]{lo, lo + next(60), i})
    	}
    	idx := NewIntervalIndex(in)
    	sorted := slices.Clone(in)
    	slices.SortStableFunc(sorted, func(a, b Interval[int, int]) int { return a.Lo - b.Lo })
    	for range 300 {
    		lo := next(1100) - 50
    		hi := lo + next(40)
    		var want []int
    		for _, iv := range sorted {
    			if iv.Lo <= hi && lo <= iv.Hi {
    				want = append(want, iv.Value)
    			}
    		}
    		if got := values(idx.Overlapping(lo, hi)); !slices.Equal(got, want) {
    			t.Fatalf("Overlapping(%d, %d) = %v, want %v", lo, hi, got, want)
    		}
    	}
    }

    func TestLarge(t *testing.T) {
    	n := 100_000
    	in := make([]Interval[Timestamp, int], 0, n)
    	for i := range n - 3 {
    		lo := Timestamp((i * 7919) % n * 10)
    		in = append(in, Interval[Timestamp, int]{lo, lo + Timestamp(i%25), i})
    	}
    	// A few very long intervals, so "only look near lo" shortcuts don't work.
    	in = append(in,
    		Interval[Timestamp, int]{0, Timestamp(n * 10), -1},
    		Interval[Timestamp, int]{5, 6, -2},
    		Interval[Timestamp, int]{Timestamp(n * 5), Timestamp(n * 10), -3})
    	start := time.Now()
    	idx := NewIntervalIndex(in)
    	total := 0
    	for q := range n {
    		lo := Timestamp(q * 10)
    		for range idx.Overlapping(lo+3, lo+4) {
    			total++
    		}
    	}
    	d := time.Since(start)
    	if want := 293_991; total != want {
    		t.Errorf("large test: %d results in total, want %d", total, want)
    	}
    	if d > time.Second {
    		t.Errorf("building an index of %d intervals and running %d queries took %v: skip whole groups of intervals instead of checking every one", n, n, d)
    	}
    }
---

Stash schedules maintenance jobs as time **intervals**: a backup from 10 to 20,
a reindex from 30 to 40. Before starting new work it asks "which jobs overlap
this window?", tens of thousands of times a minute.

Build a generic `IntervalIndex[K, V]` over `Interval[K, V]` values, where
`[Lo, Hi]` is a closed range of keys (both ends included) and `Value` is any
payload:

- `NewIntervalIndex(items)` builds the index. Intervals with `Lo > Hi` are
  ignored. The index keeps its own copy: later changes to `items` don't affect
  it, and `items` itself isn't modified.
- `Len()` returns the number of intervals in the index.
- `Overlapping(lo, hi)` yields every interval that shares at least one key with
  `[lo, hi]` (that is, `iv.Lo <= hi && lo <= iv.Hi`), ordered by `Lo`, with
  equal `Lo`s in their input order. If `lo > hi` it yields nothing. It stops
  as soon as the loop breaks, and can be ranged over again.

## Example

```go
idx := NewIntervalIndex([]Interval[int, string]{
	{10, 20, "backup"},
	{0, 100, "maintenance"},
	{15, 16, "deploy"},
	{30, 40, "reindex"},
})
for iv := range idx.Overlapping(18, 32) {
	fmt.Println(iv.Lo, iv.Hi, iv.Value)
}
// 0 100 maintenance
// 10 20 backup
// 30 40 reindex
```

## Constraints

- `K` is any `cmp.Ordered` type: the hidden tests use `type Timestamp int64`,
  `type Version string` and `float64`, with struct-pointer and other payloads.
- The index is built once and then only queried; there's no insert or delete.
- **Performance**: the big test builds an index of 100,000 intervals (including
  a few that span nearly everything) and runs 100,000 queries, each with a
  handful of results, under a one-second limit. Checking every interval per
  query means 10 billion checks. Aim for O(n log n) to build and roughly
  O(log n + k) per query with `k` results.
