---
title: Paginate
difficulty: medium
after: iterators
hints:
  - 'Return `func(yield func([]T) bool) { ... }`. Inside, range over `seq` and append each value to a `page` slice. When the page reaches `size`, yield it.'
  - 'After yielding a page, start the next one with a **new** slice (`page = make([]T, 0, size)` or `page = nil`). Reusing the old backing array with `page = page[:0]` would overwrite pages the caller kept.'
  - 'If `yield` returns `false`, `return` straight away: that stops the `range` over `seq` too. After the loop, yield the last, partial page if it isn''t empty.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"iter"
    	"slices"
    )

    func Paginate[T any](seq iter.Seq[T], size int) iter.Seq[[]T] {
    	return func(yield func([]T) bool) {
    	}
    }

    func main() {
    	lines := slices.Values([]string{"l1", "l2", "l3", "l4", "l5", "l6", "l7"})
    	pages := 0
    	for page := range Paginate(lines, 3) {
    		fmt.Println(page)
    		pages++
    	}
    	fmt.Println(pages, "pages")
    	// want:
    	// [l1 l2 l3]
    	// [l4 l5 l6]
    	// [l7]
    	// 3 pages
    }
  solution: |
    package main

    import (
    	"fmt"
    	"iter"
    	"slices"
    )

    func Paginate[T any](seq iter.Seq[T], size int) iter.Seq[[]T] {
    	return func(yield func([]T) bool) {
    		if size < 1 {
    			return
    		}
    		page := make([]T, 0, size)
    		for v := range seq {
    			page = append(page, v)
    			if len(page) == size {
    				if !yield(page) {
    					return
    				}
    				page = make([]T, 0, size)
    			}
    		}
    		if len(page) > 0 {
    			yield(page)
    		}
    	}
    }

    func main() {
    	lines := slices.Values([]string{"l1", "l2", "l3", "l4", "l5", "l6", "l7"})
    	pages := 0
    	for page := range Paginate(lines, 3) {
    		fmt.Println(page)
    		pages++
    	}
    	fmt.Println(pages, "pages")
    }
  tests: |
    package main

    import (
    	"fmt"
    	"iter"
    	"slices"
    	"testing"
    )

    func upTo(n int) []int {
    	s := make([]int, n)
    	for i := range s {
    		s[i] = i + 1
    	}
    	return s
    }

    func TestPaginate(t *testing.T) {
    	tests := []struct {
    		input []int
    		size  int
    		want  [][]int
    	}{
    		{upTo(7), 3, [][]int{{1, 2, 3}, {4, 5, 6}, {7}}},
    		{upTo(6), 3, [][]int{{1, 2, 3}, {4, 5, 6}}},
    		{upTo(3), 5, [][]int{{1, 2, 3}}},
    		{upTo(4), 1, [][]int{{1}, {2}, {3}, {4}}},
    		{nil, 3, nil},
    		{upTo(4), 0, nil},
    		{upTo(4), -2, nil},
    	}
    	for _, tt := range tests {
    		got := slices.Collect(Paginate(slices.Values(tt.input), tt.size))
    		if fmt.Sprint(got) != fmt.Sprint(tt.want) {
    			t.Errorf("Paginate(%v, %d) yielded %v, want %v", tt.input, tt.size, got, tt.want)
    		}
    	}
    }

    func TestPaginatePagesAreIndependent(t *testing.T) {
    	pages := slices.Collect(Paginate(slices.Values(upTo(6)), 2))
    	want := "[[1 2] [3 4] [5 6]]"
    	if fmt.Sprint(pages) != want {
    		t.Fatalf("collecting Paginate(1..6, 2) gave %v, want %s: did you reuse one slice for every page?", pages, want)
    	}
    	pages[0][0] = 99
    	if pages[1][0] != 3 || pages[2][0] != 5 {
    		t.Errorf("changing page 0 changed another page (%v): each page needs its own backing array", pages)
    	}
    }

    // counted yields 1, 2, 3, ... and counts how many values it produced.
    // It stands in for an endless source, but gives up after a million values
    // so an eager solution fails instead of hanging.
    func counted(produced *int) iter.Seq[int] {
    	return func(yield func(int) bool) {
    		for i := 1; i <= 1_000_000; i++ {
    			*produced++
    			if !yield(i) {
    				return
    			}
    		}
    	}
    }

    func TestPaginateStopsEarly(t *testing.T) {
    	produced := 0
    	var got [][]int
    	for page := range Paginate(counted(&produced), 4) {
    		got = append(got, page)
    		if len(got) == 2 {
    			break
    		}
    	}
    	if fmt.Sprint(got) != "[[1 2 3 4] [5 6 7 8]]" {
    		t.Errorf("first two pages of Paginate(1, 2, 3, ..., 4) = %v, want [[1 2 3 4] [5 6 7 8]]", got)
    	}
    	if produced != 8 {
    		t.Errorf("taking 2 pages of 4 pulled %d values from the source, want 8: stop ranging over seq as soon as yield returns false", produced)
    	}
    }

    func TestPaginateIsLazy(t *testing.T) {
    	produced := 0
    	pages := Paginate(counted(&produced), 10)
    	if produced != 0 {
    		t.Errorf("calling Paginate pulled %d values before anyone ranged over it, want 0", produced)
    	}
    	for page := range pages {
    		if len(page) != 10 || page[0] != 1 {
    			t.Errorf("first page of an endless source = %v, want [1 ... 10]", page)
    		}
    		break
    	}
    }

    func TestPaginateReusable(t *testing.T) {
    	pages := Paginate(slices.Values(upTo(5)), 2)
    	a, b := slices.Collect(pages), slices.Collect(pages)
    	if fmt.Sprint(a) != fmt.Sprint(b) || len(a) != 3 {
    		t.Errorf("ranging over the same Paginate result twice gave %v and %v, want [[1 2] [3 4] [5]] both times", a, b)
    	}
    }
---

Doc2Doc's PDF exporter lays lines out in **pages** of a fixed size. Documents
can be enormous (or streamed from a slow source), so it works on an
`iter.Seq` of lines and yields one page at a time.

Write the generic `Paginate(seq, size)`. It returns an `iter.Seq[[]T]` that
yields the values of `seq` grouped into pages:

- every page holds exactly `size` values, except the last one, which holds
  whatever is left (and is never empty);
- if `size < 1` or `seq` is empty, it yields nothing;
- each page is a **separate slice**: a caller that keeps a page and edits it
  must not change any other page;
- it's **lazy**: it pulls values from `seq` only as the consumer asks for pages,
  and as soon as the consumer stops (e.g. with `break`), it stops pulling.

## Example

```go
lines := slices.Values([]string{"l1", "l2", "l3", "l4", "l5", "l6", "l7"})
for page := range Paginate(lines, 3) {
	fmt.Println(page)
}
// [l1 l2 l3]
// [l4 l5 l6]
// [l7]
```

Taking only the first 2 pages of 4 from an endless source pulls exactly 8
values from it.

## Constraints

- `seq` may be infinite, so never collect it all first.
- Calling `Paginate` itself must not pull anything; the work happens when
  someone ranges over the result. Ranging over it twice gives the same pages
  (as long as `seq` can be ranged over twice).
