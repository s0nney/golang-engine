---
title: 'Practice: Labels and Tallies'
exercise:
  starter: |
    package main

    import "fmt"

    // labeler returns a function that hands out numbered labels:
    // prefix-1, prefix-2, prefix-3, ... Each labeler counts on its own.
    func labeler(prefix string) func() string {
    	return func() string {
    		// ?
    		return prefix + "-1"
    	}
    }

    // wordTally returns two closures that share one running total:
    // add counts the words in a document and adds them to the total,
    // and total reports the total so far.
    // Hint: strings.Fields splits a document into words (add the import).
    func wordTally() (add func(doc string), total func() int) {
    	add = func(doc string) {
    		// ?
    	}
    	total = func() int {
    		// ?
    		return 0
    	}
    	return add, total
    }

    func main() {
    	fig := labeler("fig")
    	table := labeler("table")
    	fmt.Println(fig(), fig(), table(), fig()) // fig-1 fig-2 table-1 fig-3

    	add, total := wordTally()
    	add("Doc2Doc converts files")
    	add("and counts words")
    	fmt.Println("words so far:", total()) // 6
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // labeler returns a function that hands out numbered labels:
    // prefix-1, prefix-2, prefix-3, ... Each labeler counts on its own.
    func labeler(prefix string) func() string {
    	n := 0
    	return func() string {
    		n++
    		return fmt.Sprintf("%s-%d", prefix, n)
    	}
    }

    // wordTally returns two closures that share one running total:
    // add counts the words in a document and adds them to the total,
    // and total reports the total so far.
    func wordTally() (add func(doc string), total func() int) {
    	count := 0
    	add = func(doc string) {
    		count += len(strings.Fields(doc))
    	}
    	total = func() int {
    		return count
    	}
    	return add, total
    }

    func main() {
    	fig := labeler("fig")
    	table := labeler("table")
    	fmt.Println(fig(), fig(), table(), fig()) // fig-1 fig-2 table-1 fig-3

    	add, total := wordTally()
    	add("Doc2Doc converts files")
    	add("and counts words")
    	fmt.Println("words so far:", total()) // 6
    }
  tests: |
    package main

    import "testing"

    func TestLabeler(t *testing.T) {
    	fig := labeler("fig")
    	for i, want := range []string{"fig-1", "fig-2", "fig-3"} {
    		if got := fig(); got != want {
    			t.Errorf("call %d of labeler(\"fig\") = %q, want %q", i+1, got, want)
    		}
    	}
    }

    func TestLabelersAreIndependent(t *testing.T) {
    	a := labeler("note")
    	b := labeler("note")
    	a()
    	a()
    	if got := b(); got != "note-1" {
    		t.Errorf("a second labeler started at %q, want %q (each labeler needs its own counter)", got, "note-1")
    	}
    	if got := a(); got != "note-3" {
    		t.Errorf("first labeler's third call = %q, want %q", got, "note-3")
    	}
    }

    func TestWordTally(t *testing.T) {
    	add, total := wordTally()
    	if got := total(); got != 0 {
    		t.Errorf("total() before any add = %d, want 0", got)
    	}
    	add("one two three")
    	if got := total(); got != 3 {
    		t.Errorf("total() after adding 3 words = %d, want 3", got)
    	}
    	add("  four\tfive  ")
    	add("")
    	if got := total(); got != 5 {
    		t.Errorf("total() after adding 5 words in all = %d, want 5", got)
    	}

    	add2, total2 := wordTally()
    	add2("fresh")
    	if got := total2(); got != 1 {
    		t.Errorf("a new wordTally's total = %d, want 1 (tallies must not share state)", got)
    	}
    	if got := total(); got != 5 {
    		t.Errorf("the first tally changed to %d after using a second one, want 5", got)
    	}
    }
---

Two small Doc2Doc features, both perfect jobs for a stateful closure.

## Your task

1. `labeler(prefix)` returns a function that hands out numbered labels: `fig-1`,
   `fig-2`, `fig-3`, and so on. Every labeler counts **on its own**, so a `table`
   labeler starts at `table-1` no matter how many figures there are.
2. `wordTally()` returns **two** closures that share one running total: `add(doc)` adds
   the number of words in `doc`, and `total()` reports the sum so far. Every call to
   `wordTally` starts a fresh, independent tally.

```go
fig := labeler("fig")
fig() // "fig-1"
fig() // "fig-2"
```

## Tips

- Declare the counter variable inside `labeler`, *before* the `return`, so each call
  to `labeler` creates a new one and the returned closure captures it.
- In `wordTally`, both closures must capture the **same** variable. Declare it once,
  above both.
- `strings.Fields` splits text into words on any whitespace. You'll need to add the
  `"strings"` import yourself, and `fmt.Sprintf` builds the label.
