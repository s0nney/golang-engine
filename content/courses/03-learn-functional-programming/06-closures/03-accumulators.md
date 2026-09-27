---
title: Accumulators
quiz:
  - question: |
      What does this print?

      ```go
      func accumulator() func(string) int {
          total := 0
          return func(doc string) int {
              total += len(strings.Fields(doc))
              return total
          }
      }

      func main() {
          add := accumulator()
          add("one two")
          add("three")
          fmt.Println(add(""))
      }
      ```
    options:
      - text: '`0`'
      - text: '`3`'
        correct: true
      - text: '`1`'
      - text: '`2`'
    explanation: |
      `total` persists between calls: 2 after the first call, 3 after the second.
      The empty string adds zero words, so the last call returns 3.
  - question: 'In `wordStats` from this lesson, what do `add` and `report` have in common?'
    options:
      - text: They capture the same `counts` map, so words added through `add` show up in `report`
        correct: true
      - text: Nothing, because each closure gets its own copy of `counts`
      - text: They must be called in the same goroutine
    explanation: |
      Closures created in the same call share the variables of that call. Both
      functions refer to the one `counts` variable created by that `wordStats` call.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // vocabulary returns an accumulator. Each call adds the words of doc
    // (split with strings.Fields, compared case-insensitively) to the set of
    // words seen so far, and returns how many DISTINCT words it has seen in
    // total. Separate accumulators must not share state.
    func vocabulary() func(doc string) int {
    	// ?
    	return func(doc string) int {
    		return len(strings.Fields(doc))
    	}
    }

    func main() {
    	add := vocabulary()
    	fmt.Println(add("the quick fox")) // 3
    	fmt.Println(add("The lazy dog"))  // 5: "the" is not new
    	fmt.Println(add("THE FOX"))       // 5

    	other := vocabulary()
    	fmt.Println(other("brand new counter")) // 3
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    func vocabulary() func(doc string) int {
    	seen := map[string]bool{}
    	return func(doc string) int {
    		for _, w := range strings.Fields(doc) {
    			seen[strings.ToLower(w)] = true
    		}
    		return len(seen)
    	}
    }

    func main() {
    	add := vocabulary()
    	fmt.Println(add("the quick fox"))
    	fmt.Println(add("The lazy dog"))
    	fmt.Println(add("THE FOX"))

    	other := vocabulary()
    	fmt.Println(other("brand new counter"))
    }
  tests: |
    package main

    import "testing"

    func TestVocabulary(t *testing.T) {
    	add := vocabulary()
    	for _, tt := range []struct {
    		doc  string
    		want int
    	}{
    		{"the quick fox", 3},
    		{"The lazy dog", 5},
    		{"THE FOX", 5},
    		{"", 5},
    		{"jumps  over\nthe\tdog", 7},
    	} {
    		if got := add(tt.doc); got != tt.want {
    			t.Errorf("after adding %q: distinct words = %d, want %d", tt.doc, got, tt.want)
    		}
    	}
    }

    func TestVocabularyIndependent(t *testing.T) {
    	a, b := vocabulary(), vocabulary()
    	a("one two three")
    	if got := b("one"); got != 1 {
    		t.Errorf("a second vocabulary() saw %d words after adding %q, want 1 (accumulators must not share state)", got, "one")
    	}
    	if got := a("four"); got != 4 {
    		t.Errorf("first vocabulary() = %d after 4 distinct words, want 4", got)
    	}
    }
---

A counter only goes up by one. An **accumulator** is its bigger sibling: a closure
that takes input on every call and folds it into a running result, like a `Reduce`
that you feed one item at a time.

## A running word count

Doc2Doc processes documents one by one as it reads them from disk. It can't wait
until it has all of them to compute totals, so it keeps a running tally:

```go
package main

import (
	"fmt"
	"strings"
)

func wordCounter() func(doc string) int {
	total := 0
	return func(doc string) int {
		total += len(strings.Fields(doc))
		return total
	}
}

func main() {
	add := wordCounter()
	for _, doc := range []string{"Hello Doc2Doc", "Convert all the things", "Done"} {
		fmt.Println("running total:", add(doc))
	}
}
```

```text
running total: 2
running total: 6
running total: 7
```

## Several closures sharing state

A function can return more than one closure. If they're created in the same call,
they **share** the captured variables. That gives you a mini-object with several
methods:

```go
package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

func wordStats() (add func(string), report func() string) {
	counts := map[string]int{}

	add = func(doc string) {
		for _, w := range strings.Fields(strings.ToLower(doc)) {
			counts[w]++
		}
	}
	report = func() string {
		var b strings.Builder
		for _, w := range slices.Sorted(maps.Keys(counts)) {
			fmt.Fprintf(&b, "%s=%d ", w, counts[w])
		}
		return strings.TrimSpace(b.String())
	}
	return add, report
}

func main() {
	add, report := wordStats()
	add("Go is fun")
	add("go is fast")
	fmt.Println(report())
}
```

```text
fast=1 fun=1 go=2 is=2
```

`add` writes to `counts`, `report` reads from it, and no other code can touch it.
The map's keys are sorted before printing, because Go deliberately randomises map
iteration order.

## Don't leak the internals

Look at `report` again. It returns a *string*, not the map. If it returned `counts`
directly, the caller could change the map and corrupt the accumulator's state behind
its back. If you must hand out a map or slice, hand out a copy (`maps.Clone`,
`slices.Clone`).

## Accumulators vs Reduce

Both fold many values into one. The difference is control:

- **`Reduce`** is pure. It gets the whole slice up front and returns a final answer.
- **An accumulator** is stateful. It gets values one at a time, whenever they arrive,
  and can report a partial answer at any point.

Prefer `Reduce` (or a loop) when you have all the data. Reach for an accumulator when
data trickles in over time, such as lines from a stream or events from a server.

## Your turn

Doc2Doc's style checker wants to know how rich a book's vocabulary is, chapter by chapter. Complete `vocabulary()`. It returns an accumulator: each call adds the words of one document to the set of words seen so far and returns how many **distinct** words that is in total.

- Split words with `strings.Fields` and compare them case-insensitively (`The` and `the` are the same word).
- Keep the set in a `map[string]bool` captured by the closure, created inside `vocabulary` so that two accumulators never share state.
