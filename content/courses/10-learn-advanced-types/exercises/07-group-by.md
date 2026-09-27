---
title: Group By
difficulty: medium
after: type-inference
hints:
  - 'The `S ~[]E` constraint is what lets the groups keep the caller''s slice type: build each group as an `S` (for example `var g S` then `g = append(g, e)`), never as a plain `[]E`.'
  - 'You need the groups in **first-seen** order, and a map''s order is random. Keep a `map[K]int` from key to position plus a slice of groups (`[]S`) and a slice of keys (`[]K`) in that order. One pass over `s` fills them in O(n).'
  - 'Do the grouping *inside* the returned `func(yield func(K, S) bool)`, so nothing happens until someone ranges over it and every range starts fresh. Stop as soon as `yield` returns `false`.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"iter"
    )

    // GroupBy groups the elements of s by key(e). It yields each key once,
    // with its elements in their original order, in the order the keys
    // first appear in s. Each group has the same type as s.
    func GroupBy[S ~[]E, E any, K comparable](s S, key func(E) K) iter.Seq2[K, S] {
    	return func(yield func(K, S) bool) {}
    }

    type Job struct {
    	Queue string
    	ID    int
    }

    // Batch is a named slice type; GroupBy should hand back Batches.
    type Batch []Job

    func main() {
    	b := Batch{{"email", 1}, {"thumbs", 2}, {"email", 3}, {"billing", 4}, {"thumbs", 5}}
    	groups := 0
    	for queue, jobs := range GroupBy(b, func(j Job) string { return j.Queue }) {
    		fmt.Printf("%s: %v (%T)\n", queue, jobs, jobs)
    		groups++
    	}
    	fmt.Println(groups, "groups")
    	// want:
    	// email: [{email 1} {email 3}] (main.Batch)
    	// thumbs: [{thumbs 2} {thumbs 5}] (main.Batch)
    	// billing: [{billing 4}] (main.Batch)
    	// 3 groups
    }
  solution: |
    package main

    import (
    	"fmt"
    	"iter"
    )

    // GroupBy groups the elements of s by key(e). It yields each key once,
    // with its elements in their original order, in the order the keys
    // first appear in s. Each group has the same type as s.
    func GroupBy[S ~[]E, E any, K comparable](s S, key func(E) K) iter.Seq2[K, S] {
    	return func(yield func(K, S) bool) {
    		index := make(map[K]int)
    		var keys []K
    		var groups []S
    		for _, e := range s {
    			k := key(e)
    			i, ok := index[k]
    			if !ok {
    				i = len(groups)
    				index[k] = i
    				keys = append(keys, k)
    				groups = append(groups, nil)
    			}
    			groups[i] = append(groups[i], e)
    		}
    		for i, k := range keys {
    			if !yield(k, groups[i]) {
    				return
    			}
    		}
    	}
    }

    type Job struct {
    	Queue string
    	ID    int
    }

    // Batch is a named slice type; GroupBy should hand back Batches.
    type Batch []Job

    func main() {
    	b := Batch{{"email", 1}, {"thumbs", 2}, {"email", 3}, {"billing", 4}, {"thumbs", 5}}
    	for queue, jobs := range GroupBy(b, func(j Job) string { return j.Queue }) {
    		fmt.Printf("%s: %v (%T)\n", queue, jobs, jobs)
    	}
    }
  tests: |
    package main

    import (
    	"fmt"
    	"slices"
    	"testing"
    	"time"
    )

    type Region string

    type Order struct {
    	Region Region
    	Cents  int
    }

    type Orders []Order

    type shard struct {
    	Region Region
    	Big    bool
    }

    type pair[K comparable, S any] struct {
    	k K
    	g S
    }

    func collect[K comparable, S any](seq func(func(K, S) bool)) []pair[K, S] {
    	var out []pair[K, S]
    	for k, g := range seq {
    		out = append(out, pair[K, S]{k, g})
    	}
    	return out
    }

    func TestGroupByJobs(t *testing.T) {
    	b := Batch{{"email", 1}, {"thumbs", 2}, {"email", 3}, {"billing", 4}, {"thumbs", 5}}
    	orig := slices.Clone(b)
    	got := collect(GroupBy(b, func(j Job) string { return j.Queue }))
    	want := []pair[string, Batch]{
    		{"email", Batch{{"email", 1}, {"email", 3}}},
    		{"thumbs", Batch{{"thumbs", 2}, {"thumbs", 5}}},
    		{"billing", Batch{{"billing", 4}}},
    	}
    	if len(got) != len(want) {
    		t.Fatalf("GroupBy(%v, by queue) yielded %d groups %v, want %d: %v", b, len(got), got, len(want), want)
    	}
    	for i := range want {
    		if got[i].k != want[i].k || !slices.Equal(got[i].g, want[i].g) {
    			t.Errorf("GroupBy(%v, by queue) group %d = %q %v, want %q %v", b, i, got[i].k, got[i].g, want[i].k, want[i].g)
    		}
    	}
    	if !slices.Equal(b, orig) {
    		t.Errorf("GroupBy changed its input to %v", b)
    	}
    }

    func TestGroupByKeepsNamedSliceType(t *testing.T) {
    	os := Orders{{"eu", 500}, {"us", 120}, {"eu", 80}}
    	for region, group := range GroupBy(os, func(o Order) Region { return o.Region }) {
    		// This line only compiles if the group's type is Orders, not []Order.
    		var typed Orders = group
    		if got := fmt.Sprintf("%T", typed); got != "main.Orders" {
    			t.Errorf("group %q has type %s, want main.Orders", region, got)
    		}
    	}
    }

    func TestGroupByStructKeys(t *testing.T) {
    	os := Orders{{"eu", 500}, {"us", 120}, {"eu", 80}, {"us", 9000}, {"eu", 1000}}
    	got := collect(GroupBy(os, func(o Order) shard { return shard{o.Region, o.Cents >= 500} }))
    	want := []pair[shard, Orders]{
    		{shard{"eu", true}, Orders{{"eu", 500}, {"eu", 1000}}},
    		{shard{"us", false}, Orders{{"us", 120}}},
    		{shard{"eu", false}, Orders{{"eu", 80}}},
    		{shard{"us", true}, Orders{{"us", 9000}}},
    	}
    	if len(got) != len(want) {
    		t.Fatalf("GroupBy(%v, by shard) yielded %d groups %v, want %d", os, len(got), got, len(want))
    	}
    	for i := range want {
    		if got[i].k != want[i].k || !slices.Equal(got[i].g, want[i].g) {
    			t.Errorf("GroupBy(%v, by shard) group %d = %+v %v, want %+v %v", os, i, got[i].k, got[i].g, want[i].k, want[i].g)
    		}
    	}
    }

    func TestGroupByPlainSliceAndEmpty(t *testing.T) {
    	words := []string{"ant", "bee", "cat", "ape", "cow", "bat"}
    	got := collect(GroupBy(words, func(w string) byte { return w[0] }))
    	if len(got) != 3 || got[0].k != 'a' || got[1].k != 'b' || got[2].k != 'c' ||
    		!slices.Equal(got[0].g, []string{"ant", "ape"}) || !slices.Equal(got[2].g, []string{"cat", "cow"}) {
    		t.Errorf("GroupBy(%q, first letter) = %v, want [{a [ant ape]} {b [bee bat]} {c [cat cow]}]", words, got)
    	}
    	if got := collect(GroupBy(Batch(nil), func(j Job) int { return j.ID })); len(got) != 0 {
    		t.Errorf("GroupBy(nil) yielded %v, want nothing", got)
    	}
    	one := collect(GroupBy([]int{7, 7, 7}, func(int) bool { return true }))
    	if len(one) != 1 || !one[0].k || !slices.Equal(one[0].g, []int{7, 7, 7}) {
    		t.Errorf("GroupBy([7 7 7], always true) = %v, want [{true [7 7 7]}]", one)
    	}
    }

    func TestGroupByStopsEarly(t *testing.T) {
    	nums := []int{1, 2, 3, 4, 5, 6}
    	var seen []int
    	for k := range GroupBy(nums, func(n int) int { return n % 3 }) {
    		seen = append(seen, k)
    		if len(seen) == 2 {
    			break
    		}
    	}
    	if !slices.Equal(seen, []int{1, 2}) {
    		t.Errorf("breaking after 2 groups of GroupBy([1..6], n%%3) saw keys %v, want [1 2]", seen)
    	}
    }

    func TestGroupByIsReusableAndLazy(t *testing.T) {
    	calls := 0
    	seq := GroupBy([]int{1, 2, 3}, func(n int) int { calls++; return n % 2 })
    	if calls != 0 {
    		t.Errorf("GroupBy called key %d times before anyone ranged over it, want 0 (group inside the iterator)", calls)
    	}
    	a, b := collect(seq), collect(seq)
    	if len(a) != 2 || len(b) != 2 {
    		t.Errorf("ranging over the same GroupBy twice gave %v then %v, want 2 groups both times", a, b)
    	}
    }

    func TestGroupByGroupsAreIndependent(t *testing.T) {
    	in := []int{1, 2, 3, 4}
    	got := collect(GroupBy(in, func(n int) bool { return n%2 == 0 }))
    	if len(got) != 2 {
    		t.Fatalf("GroupBy([1 2 3 4], even) yielded %d groups, want 2", len(got))
    	}
    	odd := append(got[0].g, 99)
    	odd[0] = -1
    	if !slices.Equal(got[1].g, []int{2, 4}) || !slices.Equal(in, []int{1, 2, 3, 4}) {
    		t.Errorf("appending to / changing the odd group changed the even group to %v or the input to %v", got[1].g, in)
    	}
    }

    func TestGroupByLarge(t *testing.T) {
    	n := 200_000
    	in := make([]int, n)
    	for i := range in {
    		in[i] = i
    	}
    	start := time.Now()
    	groups := 0
    	for _, g := range GroupBy(in, func(v int) int { return v % 50_000 }) {
    		groups++
    		if len(g) != 4 {
    			t.Fatalf("group of size %d, want 4", len(g))
    		}
    	}
    	if groups != 50_000 {
    		t.Errorf("GroupBy(200,000 ints, v%%50,000) yielded %d groups, want 50,000", groups)
    	}
    	if d := time.Since(start); d > time.Second {
    		t.Errorf("GroupBy(200,000 ints into 50,000 groups) took %v: find each key's group with a map, not a search", d)
    	}
    }
---

Stash's job runner receives a mixed `Batch` of jobs and hands each queue its own
batch. That's a **group by**, and Stash wants one that works for any slice.

Write `GroupBy(s, key)`. It returns an `iter.Seq2[K, S]` that yields each
distinct key once, together with the elements of `s` that have that key:

- Keys come out in the order they **first appear** in `s`.
- Within a group, elements keep their original order.
- Each group has the **same type as `s`**: grouping a `Batch` gives `Batch`
  groups, not `[]Job`.

## Example

```go
b := Batch{{"email", 1}, {"thumbs", 2}, {"email", 3}, {"billing", 4}, {"thumbs", 5}}
for queue, jobs := range GroupBy(b, func(j Job) string { return j.Queue }) {
	fmt.Println(queue, jobs)
}
// email [{email 1} {email 3}]
// thumbs [{thumbs 2} {thumbs 5}]
// billing [{billing 4}]
```

Callers never spell out `S`, `E` or `K`: they're inferred from `s` and `key`.

## Constraints

- Don't modify `s`, and groups must not share memory with `s` or with each
  other: appending to one group can't change another.
- Keys can be any comparable type: strings, bytes, named types, structs.
- The iterator is **lazy** (no call to `key` until someone ranges over it),
  can be ranged over more than once, and stops when the loop breaks.
- Up to 200,000 elements and 50,000 groups: aim for O(n).
