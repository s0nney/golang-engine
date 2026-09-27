---
title: Wordiest Paragraph
difficulty: easy
after: first-class-functions
hints:
  - '`score` is just a value of type `func(T) int`. Call it like any other function: `score(items[i])`.'
  - 'Start with `best := items[0]` and `bestScore := score(best)`, then walk the rest. Replace the best only when a score is **strictly** bigger, so the first of several ties wins.'
  - 'Handle the empty slice first: declare `var zero T` and return `zero, false`.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // MaxBy returns the element of items with the highest score(item), and true.
    // If several elements share the highest score, it returns the FIRST of them.
    // If items is empty, it returns T's zero value and false.
    func MaxBy[T any](items []T, score func(T) int) (T, bool) {
    	// 1. Handle an empty items slice.
    	// 2. Call score on each item and remember the best one so far.
    	var zero T
    	return zero, false
    }

    func main() {
    	paragraphs := []string{
    		"Doc2Doc converts documents.",
    		"It reads Markdown, writes HTML and never loses a heading.",
    		"Short one.",
    	}
    	words := func(p string) int { return len(strings.Fields(p)) }
    	fmt.Println(MaxBy(paragraphs, words)) // want: It reads Markdown, writes HTML and never loses a heading. true
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    func MaxBy[T any](items []T, score func(T) int) (T, bool) {
    	if len(items) == 0 {
    		var zero T
    		return zero, false
    	}
    	best, bestScore := items[0], score(items[0])
    	for _, item := range items[1:] {
    		if s := score(item); s > bestScore {
    			best, bestScore = item, s
    		}
    	}
    	return best, true
    }

    func main() {
    	paragraphs := []string{
    		"Doc2Doc converts documents.",
    		"It reads Markdown, writes HTML and never loses a heading.",
    		"Short one.",
    	}
    	words := func(p string) int { return len(strings.Fields(p)) }
    	fmt.Println(MaxBy(paragraphs, words))
    }
  tests: |
    package main

    import (
    	"strings"
    	"testing"
    )

    func words(p string) int { return len(strings.Fields(p)) }

    func TestMaxByStrings(t *testing.T) {
    	tests := []struct {
    		items  []string
    		score  func(string) int
    		name   string
    		want   string
    		wantOK bool
    	}{
    		{[]string{"a b", "a b c d", "a"}, words, "words", "a b c d", true},
    		{[]string{"one two", "three four", "five"}, words, "words", "one two", true},
    		{[]string{"tiny", "a much longer line", "mid line"}, func(s string) int { return len(s) }, "len", "a much longer line", true},
    		{[]string{"tiny", "a much longer line", "mid line"}, func(s string) int { return -len(s) }, "-len", "tiny", true},
    		{[]string{"solo"}, words, "words", "solo", true},
    		{[]string{"", ""}, words, "words", "", true},
    		{nil, words, "words", "", false},
    		{[]string{}, words, "words", "", false},
    	}
    	for _, tt := range tests {
    		got, ok := MaxBy(tt.items, tt.score)
    		if got != tt.want || ok != tt.wantOK {
    			t.Errorf("MaxBy(%q, %s) = %q, %v, want %q, %v", tt.items, tt.name, got, ok, tt.want, tt.wantOK)
    		}
    	}
    }

    func TestMaxByNegativeScores(t *testing.T) {
    	scores := []int{-9, -3, -7, -3}
    	got, ok := MaxBy(scores, func(n int) int { return n })
    	if got != -3 || !ok {
    		t.Errorf("MaxBy(%v, identity) = %d, %v, want -3, true (all scores are negative, so don't start the best score at 0)", scores, got, ok)
    	}
    }

    func TestMaxByStructs(t *testing.T) {
    	type doc struct {
    		Name  string
    		Pages int
    	}
    	docs := []doc{{"guide", 12}, {"spec", 40}, {"notes", 3}, {"manual", 40}}
    	got, ok := MaxBy(docs, func(d doc) int { return d.Pages })
    	if got.Name != "spec" || !ok {
    		t.Errorf("MaxBy(docs, pages) = %v, %v, want {spec 40}, true (the first of two 40-page docs)", got, ok)
    	}
    }

    func TestMaxByCallsScoreOncePerItem(t *testing.T) {
    	calls := 0
    	counted := func(s string) int {
    		calls++
    		return len(s)
    	}
    	MaxBy([]string{"a", "bb", "ccc", "dd"}, counted)
    	if calls != 4 {
    		t.Errorf("MaxBy with 4 items called score %d times, want 4 (once per item)", calls)
    	}
    }
---

Doc2Doc's summary view highlights the **wordiest paragraph** of a document. Or
the longest line. Or the biggest file. Rather than write the same loop three
times, write it once and pass the "how big is it?" rule in as a function.

Complete the generic `MaxBy(items, score)`. It calls `score` on each item and
returns the item with the **highest** score, and `true`.

- If several items share the highest score, return the **first** of them.
- If `items` is empty, return the zero value of `T` and `false`.
- Call `score` exactly **once** per item.

## Examples

```go
words := func(p string) int { return len(strings.Fields(p)) }

MaxBy([]string{"a b", "a b c d", "a"}, words)          // "a b c d", true
MaxBy([]string{"one two", "three four", "five"}, words) // "one two", true (tie: first wins)
MaxBy([]int{-9, -3, -7}, func(n int) int { return n })  // -3, true
MaxBy([]string{}, words)                                // "", false
```

## Constraints

- Scores can be negative, so don't assume the best score starts at 0.
- `items` must not be modified.
