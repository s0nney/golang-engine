---
title: Iteration and the maps Package
quiz:
  - question: |
      A map holds `{"alice": 1, "bob": 2, "carol": 3}`. In what order does
      `for name := range m` visit the names?
    options:
      - text: Alphabetical order
      - text: The order the keys were inserted
      - text: An unspecified order that can differ from one run to the next
        correct: true
      - text: Reverse alphabetical order
    explanation: |
      Map iteration order is not defined, and Go deliberately randomises it
      so that nobody accidentally relies on it. If you need a stable order,
      sort the keys first.
  - question: How do you get the keys of `m map[string]int` as a sorted slice?
    options:
      - text: '`slices.Sort(m)`'
      - text: '`slices.Sorted(maps.Keys(m))`'
        correct: true
      - text: '`maps.Keys(m)`, because keys always come out sorted'
      - text: '`sort(m.keys())`'
    explanation: |
      `maps.Keys` returns an *iterator* over the keys, in no particular order.
      `slices.Sorted` collects the values from an iterator into a new slice
      and sorts it.
---

You can loop over a map with `for ... range`, just like a slice. But there's a twist that trips up almost everyone.

## Ranging over a map

`range` over a map gives you each **key** and **value**:

```go
package main

import "fmt"

func main() {
	credits := map[string]int{"alice": 50, "bob": 0, "carol": 25}

	total := 0
	for name, c := range credits {
		fmt.Println(name, c)
		total += c
	}
	fmt.Println("total:", total)
}
```

Run it a few times, and you might see:

```text
bob 0
carol 25
alice 50
total: 75
```

and then:

```text
alice 50
bob 0
carol 25
total: 75
```

## Map order is random

**Map iteration order is not specified**, and Go deliberately randomises it. The same program can visit keys in a different order every time it runs.

Why? Maps are built for fast lookups, not ordering. Early on, people wrote code that accidentally depended on whatever order the map happened to produce, and it broke when Go's implementation changed. So Go now shuffles the order on purpose to flush out those bugs early.

The total is always 75, because the order didn't matter for adding up. But if you're printing a report for a customer, you want a stable order.

## Sorting the keys

To visit a map in a predictable order, get its keys, sort them, then loop over the sorted keys. The `maps` and `slices` packages make that a one-liner:

```go
package main

import (
	"fmt"
	"maps"
	"slices"
)

func main() {
	credits := map[string]int{"carol": 25, "alice": 50, "bob": 0}

	for _, name := range slices.Sorted(maps.Keys(credits)) {
		fmt.Println(name, credits[name])
	}
}
```

```text
alice 50
bob 0
carol 25
```

`maps.Keys` doesn't return a slice. It returns an **iterator**: something you can `range` over that produces values one at a time. (Its type is `iter.Seq[string]`.) `slices.Sorted` takes an iterator, collects everything into a new slice and sorts it.

If you want the keys in a slice without sorting them, use `slices.Collect`:

```go
names := slices.Collect(maps.Keys(credits)) // unsorted []string
```

You can also `range` directly over `maps.Keys(credits)` or `maps.Values(credits)`, and it works just like ranging over the map.

## More from the `maps` package

```go
package main

import (
	"fmt"
	"maps"
)

func main() {
	credits := map[string]int{"alice": 50, "bob": 0}

	backup := maps.Clone(credits) // independent copy
	backup["alice"] = 999

	fmt.Println(credits["alice"])
	fmt.Println(maps.Equal(credits, backup))

	extra := map[string]int{"dave": 10}
	maps.Copy(credits, extra) // add all of extra's pairs to credits
	fmt.Println(credits)
}
```

```text
50
false
map[alice:50 bob:0 dave:10]
```

| Function | What it does |
|----------|--------------|
| `maps.Keys(m)` | Iterator over the keys |
| `maps.Values(m)` | Iterator over the values |
| `maps.Clone(m)` | Independent shallow copy |
| `maps.Equal(a, b)` | Same keys with equal values? |
| `maps.Copy(dst, src)` | Copies every pair from `src` into `dst` |

## Changing a map while ranging

Deleting keys while you loop over a map is safe in Go. Keys you delete before they're reached won't be visited:

```go
for name, c := range credits {
	if c == 0 {
		delete(credits, name) // safe
	}
}
```

Adding keys during a loop is also allowed, but a new key may or may not be visited. Avoid relying on either outcome.
