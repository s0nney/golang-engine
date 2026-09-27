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
