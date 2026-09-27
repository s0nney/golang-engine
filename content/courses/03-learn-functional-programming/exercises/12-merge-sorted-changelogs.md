---
title: Merge Sorted Changelogs
difficulty: hard
after: iterators
hints:
  - 'A `for v := range a` loop can''t also step through `b` one value at a time. `iter.Pull(a)` turns `a` into a `next()` function you can call whenever you want the next value, plus a `stop()` function. Pull both inputs, and `defer stop()` for each so the sources are shut down however the loop ends.'
  - 'Keep one "current" value from each side (`va, okA := nextA()`). While both sides have a value, yield the smaller one (if `vb < va` yield `vb`, otherwise `va`) and pull the next value from that side only. When one side runs out, yield the rest of the other. Return as soon as `yield` returns false.'
  - '`MergeAll` needs no new iterator code: start with an empty sequence and fold every input in with `MergeSorted`, like `Reduce`. Merging an empty sequence with `s` gives `s`, so it works for zero, one or many inputs.'
exercise:
  starter: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"iter"
    	"slices"
    )

    func MergeSorted[T cmp.Ordered](a, b iter.Seq[T]) iter.Seq[T] {
    	return func(yield func(T) bool) {}
    }

    func MergeAll[T cmp.Ordered](seqs ...iter.Seq[T]) iter.Seq[T] {
    	return func(yield func(T) bool) {}
    }

    func main() {
    	edits := slices.Values([]int{100, 140, 190})
    	comments := slices.Values([]int{120, 130, 200})
    	fmt.Println(slices.Collect(MergeSorted(edits, comments))) // want: [100 120 130 140 190 200]

    	views := slices.Values([]int{105, 199})
    	fmt.Println(slices.Collect(MergeAll(edits, comments, views))) // want: [100 105 120 130 140 190 199 200]
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"iter"
    	"slices"
    )

    func MergeSorted[T cmp.Ordered](a, b iter.Seq[T]) iter.Seq[T] {
    	return func(yield func(T) bool) {
    		nextA, stopA := iter.Pull(a)
    		defer stopA()
    		nextB, stopB := iter.Pull(b)
    		defer stopB()
    		va, okA := nextA()
    		vb, okB := nextB()
    		for okA && okB {
    			if vb < va {
    				if !yield(vb) {
    					return
    				}
    				vb, okB = nextB()
    			} else {
    				if !yield(va) {
    					return
    				}
    				va, okA = nextA()
    			}
    		}
    		for okA {
    			if !yield(va) {
    				return
    			}
    			va, okA = nextA()
    		}
    		for okB {
    			if !yield(vb) {
    				return
    			}
    			vb, okB = nextB()
    		}
    	}
    }

    func MergeAll[T cmp.Ordered](seqs ...iter.Seq[T]) iter.Seq[T] {
    	merged := func(yield func(T) bool) {}
    	for _, s := range seqs {
    		merged = MergeSorted(merged, s)
    	}
    	return merged
    }

    func main() {
    	edits := slices.Values([]int{100, 140, 190})
    	comments := slices.Values([]int{120, 130, 200})
    	fmt.Println(slices.Collect(MergeSorted(edits, comments)))

    	views := slices.Values([]int{105, 199})
    	fmt.Println(slices.Collect(MergeAll(edits, comments, views)))
    }
  tests: |
    package main

    import (
    	"iter"
    	"slices"
    	"testing"
    )

    // source is a sorted stream for the tests. It yields start, start+step,
    // start+2*step, ... (endlessly, as far as MergeSorted can tell: it gives up
    // after a million values so an eager solution fails instead of hanging).
    // It counts how many values were pulled from it and whether it has finished.
    type source struct {
    	start, step int
    	pulled      int
    	finished    bool
    }

    func (s *source) seq() iter.Seq[int] {
    	return func(yield func(int) bool) {
    		defer func() { s.finished = true }()
    		for i := range 1_000_000 {
    			s.pulled++
    			if !yield(s.start + i*s.step) {
    				return
    			}
    		}
    	}
    }

    func first[T any](seq iter.Seq[T], n int) []T {
    	var out []T
    	for v := range seq {
    		if len(out) == n {
    			break
    		}
    		out = append(out, v)
    	}
    	return out
    }

    func TestMergeSorted(t *testing.T) {
    	tests := []struct {
    		a, b []int
    		want []int
    	}{
    		{[]int{1, 4, 9}, []int{2, 3, 10}, []int{1, 2, 3, 4, 9, 10}},
    		{[]int{1, 2, 3}, []int{7, 8}, []int{1, 2, 3, 7, 8}},
    		{[]int{7, 8}, []int{1, 2, 3}, []int{1, 2, 3, 7, 8}},
    		{[]int{1, 3, 3}, []int{3, 4}, []int{1, 3, 3, 3, 4}},
    		{nil, []int{5, 6}, []int{5, 6}},
    		{[]int{5, 6}, nil, []int{5, 6}},
    		{nil, nil, nil},
    		{[]int{-5, 0}, []int{-7, 100}, []int{-7, -5, 0, 100}},
    	}
    	for _, tt := range tests {
    		got := slices.Collect(MergeSorted(slices.Values(tt.a), slices.Values(tt.b)))
    		if !slices.Equal(got, tt.want) {
    			t.Errorf("MergeSorted(%v, %v) yielded %v, want %v", tt.a, tt.b, got, tt.want)
    		}
    	}
    }

    func TestMergeSortedIsLazy(t *testing.T) {
    	evens, odds := &source{0, 2, 0, false}, &source{1, 2, 0, false}
    	merged := MergeSorted(evens.seq(), odds.seq())
    	if evens.pulled+odds.pulled != 0 {
    		t.Fatalf("calling MergeSorted pulled %d values before anyone ranged over it, want 0", evens.pulled+odds.pulled)
    	}
    	got := first(merged, 10)
    	if !slices.Equal(got, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}) {
    		t.Errorf("first 10 values of MergeSorted(evens, odds) = %v, want [0 1 2 3 4 5 6 7 8 9]", got)
    	}
    	if total := evens.pulled + odds.pulled; total > 12 {
    		t.Errorf("taking 10 merged values pulled %d values from the sources (%d evens, %d odds), want at most 12: pull one value at a time", total, evens.pulled, odds.pulled)
    	}
    }

    func TestMergeSortedStopsSources(t *testing.T) {
    	a, b := &source{0, 3, 0, false}, &source{1, 5, 0, false}
    	first(MergeSorted(a.seq(), b.seq()), 4)
    	if !a.finished || !b.finished {
    		t.Errorf("after the consumer broke out of the loop, source a finished=%v and b finished=%v, want both true: call the stop functions from iter.Pull (defer them)", a.finished, b.finished)
    	}
    	c, d := &source{0, 1, 0, false}, &source{0, 1, 0, false}
    	shortC := func(yield func(int) bool) {
    		for v := range c.seq() {
    			if v == 3 || !yield(v) {
    				return
    			}
    		}
    	}
    	got := first(MergeSorted(shortC, d.seq()), 8)
    	if !slices.Equal(got, []int{0, 0, 1, 1, 2, 2, 3, 4}) {
    		t.Errorf("MergeSorted([0 1 2], 0 1 2 3 ...) first 8 = %v, want [0 0 1 1 2 2 3 4]", got)
    	}
    	if !c.finished || !d.finished {
    		t.Errorf("after one source ran out and the consumer broke, finished = %v and %v, want both true", c.finished, d.finished)
    	}
    }

    func TestMergeAll(t *testing.T) {
    	tests := []struct {
    		seqs [][]int
    		want []int
    	}{
    		{nil, nil},
    		{[][]int{{1, 5}}, []int{1, 5}},
    		{[][]int{{1, 5}, {2, 3}, {0, 9}}, []int{0, 1, 2, 3, 5, 9}},
    		{[][]int{{}, {4}, {}, {1, 4}}, []int{1, 4, 4}},
    	}
    	for _, tt := range tests {
    		var seqs []iter.Seq[int]
    		for _, s := range tt.seqs {
    			seqs = append(seqs, slices.Values(s))
    		}
    		got := slices.Collect(MergeAll(seqs...))
    		if !slices.Equal(got, tt.want) {
    			t.Errorf("MergeAll(%v) yielded %v, want %v", tt.seqs, got, tt.want)
    		}
    	}
    }

    func TestMergeAllHugeStreams(t *testing.T) {
    	srcs := []*source{{0, 3, 0, false}, {1, 3, 0, false}, {2, 3, 0, false}, {1_000, 1, 0, false}}
    	var seqs []iter.Seq[int]
    	for _, s := range srcs {
    		seqs = append(seqs, s.seq())
    	}
    	got := first(MergeAll(seqs...), 1_000)
    	for i, v := range got {
    		if v != i {
    			t.Fatalf("MergeAll of the streams 0,3,6... 1,4,7... 2,5,8... and 1000,1001...: value #%d = %d, want %d", i, v, i)
    		}
    	}
    	total := 0
    	for i, s := range srcs {
    		total += s.pulled
    		if !s.finished {
    			t.Errorf("source %d was not stopped after the consumer broke out of the loop", i)
    		}
    	}
    	if len(got) != 1_000 || total > 1_000+2*len(srcs) {
    		t.Errorf("taking %d values from MergeAll of 4 streams pulled %d values in total, want at most %d", len(got), total, 1_000+2*len(srcs))
    	}
    }
---

Doc2Doc's history panel shows one timeline of everything that happened to a
document: edits, comments, views. Each kind of event comes from its own
**sorted** stream of timestamps, and some of those streams are practically
endless (a busy document's view log). The panel only ever shows the first page
or so, so merging must be **lazy**.

Write two generic functions over sorted `iter.Seq`s:

- `MergeSorted(a, b)` yields every value of `a` and `b` in ascending order.
  Both inputs are sorted ascending; duplicates are kept.
- `MergeAll(seqs...)` does the same for any number of sorted sequences
  (zero sequences yields nothing).

Both must be lazy and polite to their sources:

- calling them pulls nothing; values are pulled only as the consumer asks;
- a merge holds at most **one** value from each input that it hasn't yielded
  yet, so taking the first `n` values of a two-way merge pulls at most `n + 2`
  values from the inputs;
- when the consumer stops early (e.g. `break`) or an input runs out, every
  input sequence is shut down properly: its `yield` returns `false` and its
  function returns, so its cleanup (`defer`) runs.

## Example

```go
edits := slices.Values([]int{100, 140, 190})
comments := slices.Values([]int{120, 130, 200})
slices.Collect(MergeSorted(edits, comments)) // [100 120 130 140 190 200]

views := slices.Values([]int{105, 199})
slices.Collect(MergeAll(edits, comments, views)) // [100 105 120 130 140 190 199 200]
```

Merging the endless streams `0, 2, 4, …` and `1, 3, 5, …` and breaking after 10
values yields `0 1 2 … 9` and pulls no more than 12 values from the two streams.

## Constraints

- Inputs may be endless. The tests use streams of up to a million values each
  and only look at the first thousand merged values, checking how many values
  were pulled from every stream.
- Don't collect an input into a slice.
