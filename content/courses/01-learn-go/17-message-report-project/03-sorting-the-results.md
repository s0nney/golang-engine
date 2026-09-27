---
title: Sorting the Results
quiz:
  - question: Why can't the report just `range` over the word-count map and print it?
    options:
      - text: Maps can't be used in a `for ... range` loop
      - text: Map iteration order is not specified, so the report would come out in a different order on different runs
        correct: true
      - text: '`range` over a map only gives you the keys'
    explanation: |
      Go deliberately randomizes map iteration order. To print "most common
      word first", copy the entries into a slice and sort the slice.
  - question: |
      With this comparison function, which entry comes first?

      ```go
      func(a, b WordCount) int {
      	if c := cmp.Compare(b.Count, a.Count); c != 0 {
      		return c
      	}
      	return cmp.Compare(a.Word, b.Word)
      }
      ```

      Entries: `{"yes", 2}`, `{"lunch", 3}`, `{"at", 2}`
    options:
      - text: '`{"yes", 2}`'
      - text: '`{"at", 2}`'
      - text: '`{"lunch", 3}`'
        correct: true
    explanation: |
      Comparing `b.Count` with `a.Count` (note the swapped order) sorts by
      count, **highest first**, so `lunch` with 3 wins. Only ties fall through
      to the second comparison, which puts `at` before `yes` alphabetically.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    )

    type WordCount struct {
    	Word  string
    	Count int
    }

    // topWords returns at most n entries from counts, sorted by Count
    // (highest first) and then by Word (alphabetical) to break ties.
    func topWords(counts map[string]int, n int) []WordCount {
    	// ?
    	return nil
    }

    func main() {
    	counts := map[string]int{"yes": 2, "lunch": 3, "at": 2, "noon": 1}
    	fmt.Println(topWords(counts, 3))
    	fmt.Println(topWords(counts, 10))
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"slices"
    )

    type WordCount struct {
    	Word  string
    	Count int
    }

    func topWords(counts map[string]int, n int) []WordCount {
    	var entries []WordCount
    	for word, count := range counts {
    		entries = append(entries, WordCount{word, count})
    	}
    	slices.SortFunc(entries, func(a, b WordCount) int {
    		if c := cmp.Compare(b.Count, a.Count); c != 0 {
    			return c
    		}
    		return cmp.Compare(a.Word, b.Word)
    	})
    	return entries[:min(n, len(entries))]
    }

    func main() {
    	counts := map[string]int{"yes": 2, "lunch": 3, "at": 2, "noon": 1}
    	fmt.Println(topWords(counts, 3))
    	fmt.Println(topWords(counts, 10))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func TestTopWords(t *testing.T) {
    	counts := map[string]int{"yes": 2, "lunch": 3, "at": 2, "noon": 1, "hi": 5}
    	for _, tc := range []struct {
    		n    int
    		want []WordCount
    	}{
    		{3, []WordCount{{"hi", 5}, {"lunch", 3}, {"at", 2}}},
    		{1, []WordCount{{"hi", 5}}},
    		{10, []WordCount{{"hi", 5}, {"lunch", 3}, {"at", 2}, {"yes", 2}, {"noon", 1}}},
    	} {
    		// Run several times: map order changes between runs, the result must not.
    		for range 20 {
    			got := topWords(counts, tc.n)
    			if !slices.Equal(got, tc.want) {
    				t.Fatalf("topWords(%v, %d) = %v, want %v", counts, tc.n, got, tc.want)
    			}
    		}
    	}
    }

    func TestTopWordsTies(t *testing.T) {
    	counts := map[string]int{"c": 1, "a": 1, "d": 1, "b": 1}
    	want := []WordCount{{"a", 1}, {"b", 1}}
    	for range 20 {
    		if got := topWords(counts, 2); !slices.Equal(got, want) {
    			t.Fatalf("topWords(%v, 2) = %v, want %v (ties should be alphabetical)", counts, got, want)
    		}
    	}
    }

    func TestTopWordsEmpty(t *testing.T) {
    	if got := topWords(map[string]int{}, 3); len(got) != 0 {
    		t.Errorf("topWords(empty map, 3) = %v, want no entries", got)
    	}
    }
---

We have a `map[string]int` of word counts. A report should list the **most common
words first**. There's a catch: maps have no order. Go even shuffles the iteration
order on purpose so nobody accidentally depends on it. Range over the same map twice
and you may see two different orders.

The fix is a two-step dance you'll use constantly in Go:

1. Copy the map entries into a **slice**.
2. **Sort** the slice.

## A struct for each entry

A slice needs one element type, and each entry has two parts, so we make a small
struct:

```go
type WordCount struct {
	Word  string
	Count int
}
```

## slices.SortFunc

The `slices` package can sort any slice when you tell it how to compare two
elements. The comparison function returns a negative number if `a` should come
first, a positive number if `b` should, and `0` if they're equal. `cmp.Compare`
builds exactly that number for you:

```go
package main

import (
	"cmp"
	"fmt"
	"slices"
)

type WordCount struct {
	Word  string
	Count int
}

func main() {
	counts := map[string]int{"yes": 2, "lunch": 3, "at": 2, "noon": 1}

	var entries []WordCount
	for word, count := range counts {
		entries = append(entries, WordCount{word, count})
	}

	slices.SortFunc(entries, func(a, b WordCount) int {
		if c := cmp.Compare(b.Count, a.Count); c != 0 {
			return c // bigger counts first
		}
		return cmp.Compare(a.Word, b.Word) // ties: alphabetical
	})

	for _, e := range entries {
		fmt.Printf("%-6s %d\n", e.Word, e.Count)
	}
}
```

```text
lunch  3
at     2
yes    2
noon   1
```

Two details make this work:

- **Swapping `a` and `b`** in the first `cmp.Compare` reverses the order, so the
  biggest count comes first.
- The **tie-breaker** matters. Without it, `at` and `yes` (both 2) could come out
  in either order, and the report would change from run to run. Sorting by a second
  key makes the output *deterministic*: same input, same output, every time. That
  also makes it testable.

`%-6s` pads the word to 6 characters, left-aligned, so the counts line up in a
column.

## Keeping the top N

Most reports only show the top few words. Once the slice is sorted, that's a slice
expression. Use the `min` built-in so you don't slice past the end when there are
fewer words than you asked for:

```go
top := entries[:min(3, len(entries))]
```

## Your turn

Complete `topWords`. Copy the map into a `[]WordCount`, sort it with
`slices.SortFunc` (count descending, then word ascending), and return at most `n`
entries. An empty map should give back an empty result, not a panic.

**Run** should print:

```text
[{lunch 3} {at 2} {yes 2}]
[{lunch 3} {at 2} {yes 2} {noon 1}]
```
