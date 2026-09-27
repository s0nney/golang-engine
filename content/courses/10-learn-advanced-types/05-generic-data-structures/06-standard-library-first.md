---
title: The Standard Library First
quiz:
  - question: |
      You need the keys of a `map[string]int` in sorted order. Which is the idiomatic one-liner?
    options:
      - text: '`sort.Strings(maps.Keys(m))`'
      - text: '`slices.Sorted(maps.Keys(m))`'
        correct: true
      - text: '`slices.Sort(maps.Keys(m))`'
      - text: '`maps.SortedKeys(m)`'
    explanation: |
      `maps.Keys` returns an `iter.Seq[string]`, not a slice, so `sort.Strings` and
      `slices.Sort` can't take it. `slices.Sorted` collects any sequence into a new
      slice and sorts it. There's no `maps.SortedKeys`.
  - question: When is writing your own generic collection (like Stash's `OrderedMap`) justified?
    options:
      - text: Whenever you use a map in more than one function
      - text: When it maintains an invariant the built-ins don't, such as insertion order, a size limit, or recency, that callers would otherwise have to maintain by hand
        correct: true
      - text: Always; built-in maps and slices aren't type-safe
      - text: Never; the standard library covers everything
    explanation: |
      Built-in slices and maps plus the `slices` and `maps` packages cover most needs. A
      custom type earns its place by owning an invariant (order, capacity, eviction) so
      that every caller doesn't have to get it right separately.
---

Stash now has a set, an ordered map, an LRU cache and a heap. Before you build the next collection, it's worth taking stock of what the standard library's generic packages already give you, because the best generic code is often code you didn't have to write.

## A tour in one program

```go
package main

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"
)

type Entry struct {
	Key  string
	Size int
}

func main() {
	sizes := map[string]int{"logo.png": 2048, "app.js": 512, "index.html": 128}

	// Sorted keys: iterator in, sorted slice out.
	fmt.Println(slices.Sorted(maps.Keys(sizes)))

	// Largest entry, with a custom comparison.
	entries := make([]Entry, 0, len(sizes))
	for k, v := range sizes {
		entries = append(entries, Entry{k, v})
	}
	biggest := slices.MaxFunc(entries, func(a, b Entry) int { return cmp.Compare(a.Size, b.Size) })
	fmt.Println(biggest)

	// Batches of 2, in key order.
	for batch := range slices.Chunk(slices.Sorted(maps.Keys(sizes)), 2) {
		fmt.Println(strings.Join(batch, "+"))
	}

	// Merge one map into another, from an iterator.
	extra := map[string]int{"favicon.ico": 64}
	maps.Insert(sizes, maps.All(extra))
	fmt.Println(len(sizes))
}
```

```
[app.js index.html logo.png]
{logo.png 2048}
app.js+index.html
logo.png
4
```

Every function here is generic, and they compose through two shapes: **slices** and **iterators**. `maps.Keys` produces an `iter.Seq`, `slices.Sorted` consumes one, and `slices.Chunk` produces another.

## The highlights

**`slices`** (Go 1.21, with iterator functions added in 1.23):

- search and compare: `Contains`, `Index`, `IndexFunc`, `BinarySearch`, `Equal`, `Compare`, `Max`, `MaxFunc`, `Min`, `MinFunc`
- reorder: `Sort`, `SortFunc`, `SortStableFunc`, `Reverse`
- reshape: `Insert`, `Delete`, `DeleteFunc`, `Compact`, `Clone`, `Concat`, `Repeat`, `Grow`, `Clip`
- iterators: `All`, `Values`, `Backward`, `Collect`, `AppendSeq`, `Sorted`, `SortedFunc`, `Chunk`

**`maps`**: `Keys`, `Values`, `All`, `Collect`, `Insert`, `Clone`, `Copy`, `DeleteFunc`, `Equal`, `EqualFunc`.

**`cmp`**: `Compare`, `Less`, `Or`, and the `Ordered` constraint.

**`iter`**: the `Seq` and `Seq2` types, plus `Pull` and `Pull2`.

## A map is often enough

Before writing a collection type, try the plain version:

| You want | Plain Go |
|---|---|
| a set used in one function | `map[T]struct{}` or `map[T]bool` |
| deduplicate a sorted slice | `slices.Compact` |
| sorted keys, once | `slices.Sorted(maps.Keys(m))` |
| group by key | a `map[K][]V` and `append` |
| a stack | a slice with `append` and `s[:len(s)-1]` |

If the plain version is short and local, keep it. Stash's types earn their existence differently: each one **owns an invariant** that callers would otherwise have to maintain by hand, everywhere. `OrderedMap` keeps order and contents in sync; `LRU` keeps size under the limit and recency up to date; `Heap` keeps the heap property. That's the test to apply to your own collection ideas.

## Make yours interoperate

When you do write a collection, make it fit in with the standard library rather than replace it:

- Expose contents as `iter.Seq`/`iter.Seq2` (`All`, `Keys`, `Values`), so `slices.Collect`, `slices.Sorted`, `maps.Collect` and `maps.Insert` work on it for free.
- Accept iterators where you'd accept "some values", so callers can pass `slices.Values(s)`, `maps.Keys(m)` or another collection's `All()`.
- Use the standard comparison shape, `func(a, b T) int`, so `cmp.Compare` plugs in.

## One more: unique

For completeness, Go 1.23 added the `unique` package: `unique.Make(v)` returns a `unique.Handle[T]` for any comparable value, and two handles are equal exactly when their values are. Comparing handles is a pointer comparison, however big the value is, and equal values share one canonical copy in memory. It's handy for interning keys that repeat millions of times, like hostnames in logs. Stash doesn't need it yet, but now you know it exists.

With the collections done, the next chapter turns to the other half of Go's type system: interfaces, and what's actually inside an interface value.
