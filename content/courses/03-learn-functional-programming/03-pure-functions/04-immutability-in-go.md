---
title: Immutability in Go
quiz:
  - question: |
      What does this print?

      ```go
      base := make([]string, 1, 4)
      base[0] = "# Title"
      a := append(base, "draft")
      b := append(base, "final")
      fmt.Println(a[1], b[1])
      ```
    options:
      - text: '`draft final`'
      - text: '`final final`'
        correct: true
      - text: '`draft draft`'
      - text: It panics with an index out of range error
    explanation: |
      `base` has spare capacity, so both `append` calls write into the *same*
      backing array at index 1. The second append overwrites the first, and `a`
      and `b` both see `"final"`.
  - question: |
      After `copyDocs := slices.Clone(docs)`, where `docs` is a `[]Doc` and `Doc` has a `Tags []string` field, which statement is true?
    options:
      - text: Changing `copyDocs[0].Tags[0]` also changes `docs[0].Tags[0]`
        correct: true
      - text: '`copyDocs` and `docs` share nothing, so it''s a deep copy'
      - text: '`slices.Clone` doesn''t work on slices of structs'
    explanation: |
      `slices.Clone` is a *shallow* copy. Each `Doc` struct is copied, but a slice
      field inside it is just a header pointing at the same backing array. For a
      deep copy you must clone the inner slices too.
---

Functional programming loves **immutable** data: once created, a value never
changes. Want a different value? Make a new one.

Go has no `immutable` keyword. `const` only works for numbers, strings and booleans.
So immutability in Go is mostly a *discipline*, and you need to know where the
traps are.

## What's already safe

- **Strings are immutable.** `strings.ToUpper(s)` always returns a new string, and
  nothing can change the bytes of an existing string.
- **Numbers, booleans, arrays and structs are copied** when you assign them or pass
  them to a function, as long as they don't contain slices, maps or pointers.

## The traps: slices and maps

A slice is a small header (pointer, length and capacity) that points at a backing
array. Copying the slice copies the header, **not** the array:

```go
package main

import (
	"fmt"
	"maps"
	"slices"
)

func main() {
	lines := []string{"intro", "body"}
	alias := lines
	alias[0] = "CHANGED"
	fmt.Println(lines[0]) // the original changed too

	safe := slices.Clone(lines)
	safe[0] = "safe edit"
	fmt.Println(lines[0], "/", safe[0])

	meta := map[string]string{"author": "Sam"}
	metaCopy := maps.Clone(meta)
	metaCopy["author"] = "Alex"
	fmt.Println(meta["author"], metaCopy["author"])
}
```

```text
CHANGED
CHANGED / safe edit
Sam Alex
```

Maps are the same: a map value is a reference to shared data, so two variables can
point at one map. `slices.Clone` and `maps.Clone` make a fresh copy you can change
freely.

## The append trap

This one catches experienced Go programmers. If a slice has spare capacity, `append`
writes into the *shared* backing array instead of allocating a new one:

```go
package main

import (
	"fmt"
	"slices"
)

func main() {
	header := make([]string, 0, 10)
	header = append(header, "# Report")

	draft := append(header, "draft body")
	final := append(header, "final body")
	fmt.Println(draft[1], "|", final[1])

	draft = append(slices.Clone(header), "draft body")
	final = append(slices.Clone(header), "final body")
	fmt.Println(draft[1], "|", final[1])
}
```

```text
final body | final body
draft body | final body
```

The first `draft` got silently overwritten. Cloning before appending gives each
version its own array. `slices.Concat(header, extra)` also always returns a new
slice.

## Shallow vs deep

`slices.Clone` and `maps.Clone` are **shallow**. If the elements themselves contain
slices, maps or pointers, those inner parts are still shared. A Doc2Doc `Doc` with a
`Tags []string` field needs a deep copy:

```go
func (d Doc) Clone() Doc {
	d.Tags = slices.Clone(d.Tags) // d is already a copy; now so are its tags
	return d
}
```

## Making data hard to mutate

When a type must stay immutable, hide its fields and hand out copies:

```go
type Document struct {
	lines []string // unexported: other packages can't touch it
}

func (d Document) Lines() []string {
	return slices.Clone(d.lines)
}

func (d Document) WithLine(l string) Document {
	return Document{lines: append(slices.Clone(d.lines), l)}
}
```

Copying has a cost, so don't clone everything reflexively. Clone at **boundaries**:
when you store a slice someone handed you, or when you return internal state.
