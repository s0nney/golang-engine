---
title: 'Practice: A Lazy Pipeline'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"iter"
    	"slices"
    	"strings"
    )

    // Filter yields only the values of seq for which keep returns true.
    func Filter[T any](seq iter.Seq[T], keep func(T) bool) iter.Seq[T] {
    	return func(yield func(T) bool) {
    		// ?
    	}
    }

    // Map yields f(v) for every value v of seq.
    func Map[T, U any](seq iter.Seq[T], f func(T) U) iter.Seq[U] {
    	return func(yield func(U) bool) {
    		// ?
    	}
    }

    // Take yields at most the first n values of seq, then stops without
    // asking seq for any more.
    func Take[T any](seq iter.Seq[T], n int) iter.Seq[T] {
    	return func(yield func(T) bool) {
    		// ?
    	}
    }

    func main() {
    	doc := "# Intro\ntext\n# Installing Doc2Doc\n# Converting files\nmore text\n# Formats\n"

    	headings := Filter(strings.Lines(doc), func(l string) bool {
    		return strings.HasPrefix(l, "# ")
    	})
    	titles := Map(headings, func(l string) string {
    		return strings.TrimSpace(strings.TrimPrefix(l, "# "))
    	})
    	fmt.Printf("%q\n", slices.Collect(Take(titles, 3)))
    	// ["Intro" "Installing Doc2Doc" "Converting files"]
    }
  solution: |
    package main

    import (
    	"fmt"
    	"iter"
    	"slices"
    	"strings"
    )

    // Filter yields only the values of seq for which keep returns true.
    func Filter[T any](seq iter.Seq[T], keep func(T) bool) iter.Seq[T] {
    	return func(yield func(T) bool) {
    		for v := range seq {
    			if keep(v) && !yield(v) {
    				return
    			}
    		}
    	}
    }

    // Map yields f(v) for every value v of seq.
    func Map[T, U any](seq iter.Seq[T], f func(T) U) iter.Seq[U] {
    	return func(yield func(U) bool) {
    		for v := range seq {
    			if !yield(f(v)) {
    				return
    			}
    		}
    	}
    }

    // Take yields at most the first n values of seq, then stops without
    // asking seq for any more.
    func Take[T any](seq iter.Seq[T], n int) iter.Seq[T] {
    	return func(yield func(T) bool) {
    		if n <= 0 {
    			return
    		}
    		count := 0
    		for v := range seq {
    			if !yield(v) {
    				return
    			}
    			count++
    			if count == n {
    				return
    			}
    		}
    	}
    }

    func main() {
    	doc := "# Intro\ntext\n# Installing Doc2Doc\n# Converting files\nmore text\n# Formats\n"

    	headings := Filter(strings.Lines(doc), func(l string) bool {
    		return strings.HasPrefix(l, "# ")
    	})
    	titles := Map(headings, func(l string) string {
    		return strings.TrimSpace(strings.TrimPrefix(l, "# "))
    	})
    	fmt.Printf("%q\n", slices.Collect(Take(titles, 3)))
    	// ["Intro" "Installing Doc2Doc" "Converting files"]
    }
  tests: |
    package main

    import (
    	"iter"
    	"slices"
    	"testing"
    )

    // naturals yields 0, 1, 2, ... forever and counts how many it produced.
    func naturals(produced *int) iter.Seq[int] {
    	return func(yield func(int) bool) {
    		for i := 0; ; i++ {
    			*produced++
    			if !yield(i) {
    				return
    			}
    		}
    	}
    }

    func TestFilterAndMap(t *testing.T) {
    	words := slices.Values([]string{"go", "doc", "a", "html"})
    	long := slices.Collect(Filter(words, func(s string) bool { return len(s) > 2 }))
    	if want := []string{"doc", "html"}; !slices.Equal(long, want) {
    		t.Errorf("Filter(len > 2) = %q, want %q", long, want)
    	}
    	lens := slices.Collect(Map(words, func(s string) int { return len(s) }))
    	if want := []int{2, 3, 1, 4}; !slices.Equal(lens, want) {
    		t.Errorf("Map(len) = %v, want %v", lens, want)
    	}
    }

    func TestTake(t *testing.T) {
    	nums := slices.Values([]int{1, 2, 3, 4, 5})
    	if got := slices.Collect(Take(nums, 2)); !slices.Equal(got, []int{1, 2}) {
    		t.Errorf("Take([1 2 3 4 5], 2) = %v, want [1 2]", got)
    	}
    	if got := slices.Collect(Take(nums, 10)); !slices.Equal(got, []int{1, 2, 3, 4, 5}) {
    		t.Errorf("Take([1 2 3 4 5], 10) = %v, want all five", got)
    	}
    	if got := slices.Collect(Take(nums, 0)); len(got) != 0 {
    		t.Errorf("Take(seq, 0) = %v, want nothing", got)
    	}
    }

    func TestPipelineIsLazy(t *testing.T) {
    	produced := 0
    	evens := Filter(naturals(&produced), func(n int) bool { return n%2 == 0 })
    	squares := Map(evens, func(n int) int { return n * n })
    	if produced != 0 {
    		t.Fatalf("building the pipeline pulled %d values; it should do no work until ranged over", produced)
    	}
    	got := slices.Collect(Take(squares, 3))
    	if want := []int{0, 4, 16}; !slices.Equal(got, want) {
    		t.Errorf("first 3 even squares = %v, want %v", got, want)
    	}
    	if produced != 5 {
    		t.Errorf("the source produced %d values, want 5 (0..4): Take must stop pulling once it has enough", produced)
    	}
    }

    func TestBreakStopsEveryStage(t *testing.T) {
    	produced := 0
    	var got []int
    	for v := range Map(Filter(naturals(&produced), func(n int) bool { return n > 2 }), func(n int) int { return n * 10 }) {
    		got = append(got, v)
    		if len(got) == 2 {
    			break
    		}
    	}
    	if want := []int{30, 40}; !slices.Equal(got, want) {
    		t.Errorf("got %v, want %v", got, want)
    	}
    	if produced != 5 {
    		t.Errorf("source produced %d values after break, want 5: Filter and Map must return when yield returns false", produced)
    	}
    }
---

Doc2Doc wants the first few headings of documents that may be huge. Build the lazy
adapters that make that cheap.

## Your task

Implement three iterator adapters over `iter.Seq`:

- `Filter(seq, keep)` yields only the values where `keep` returns true.
- `Map(seq, f)` yields `f(v)` for every value.
- `Take(seq, n)` yields at most the first `n` values, then stops **without pulling any
  more** from `seq`.

All three must be lazy and must respect early exit. When `yield` returns `false`, return
right away. The tests feed your adapters an **infinite** sequence and count how many
values were produced, so any extra work gets caught:

```go
Take(Map(Filter(naturals, isEven), square), 3) // 0, 4, 16, pulling only 0..4
```

## Tips

- Each adapter returns a `func(yield func(T) bool)`. Inside it, `for v := range seq` and
  call `yield` for each value you want to pass on.
- The pattern `if !yield(v) { return }` is your best friend.
- In `Take`, stop *as soon as* you've yielded `n` values. Don't loop around again to
  discover you're done, or you'll pull one value too many.
- If you see the panic "range function continued iteration after function for loop body
  returned false", some adapter kept going after `yield` returned `false`.
