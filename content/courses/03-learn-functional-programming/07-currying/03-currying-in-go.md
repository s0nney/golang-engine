---
title: When Currying Helps in Go
quiz:
  - question: '`slices.ContainsFunc` needs a `func(string) bool`, but your check `longerThan(n int, s string) bool` takes two arguments. What''s the idiomatic fix?'
    options:
      - text: Change `slices.ContainsFunc` to accept two arguments
      - text: Use a global variable for `n`
      - text: It can't be done
      - text: 'Write `func longerThan(n int) func(string) bool` so `slices.ContainsFunc(lines, longerThan(80))` fits'
        correct: true
    explanation: |
      Returning a function with `n` already fixed adapts your check to the
      one-argument shape that `ContainsFunc` wants. This is currying's most common
      real-world use in Go.
  - question: Which of these is **most** likely to be flagged in a Go code review?
    options:
      - text: '`render(cfg)(tmpl)(page)(w)(r)` with five levels of nested function types'
        correct: true
      - text: '`func handleConvert(store Store) http.HandlerFunc`'
      - text: '`slices.DeleteFunc(lines, hasPrefix("//"))`'
    explanation: |
      One level of "configure, then return a function" is idiomatic Go. Deeply
      curried chains are hard to read, hard to debug and unfamiliar to most Go
      programmers.
---

So far, currying might seem like a party trick. In Go it has one killer use:
**adapting a function to fit a shape that some other code demands.**

## Fitting the shape

Tons of Go APIs take a function with a fixed signature: `slices.ContainsFunc` wants
`func(E) bool`, `strings.FieldsFunc` wants `func(rune) bool`, and `http.Handle` wants
an `http.Handler`. Your logic often needs extra information that doesn't fit into
that signature. So you take the extra information first and return a function of the
right shape:

```go
package main

import (
	"fmt"
	"slices"
	"strings"
)

func longerThan(limit int) func(string) bool {
	return func(line string) bool {
		return len(line) > limit
	}
}

func hasPrefix(prefix string) func(string) bool {
	return func(line string) bool {
		return strings.HasPrefix(line, prefix)
	}
}

func main() {
	lines := []string{
		"# Style guide",
		"// TODO: remove this",
		"Keep lines short.",
		"This line is far too long for the Doc2Doc style checker to accept.",
	}

	fmt.Println(slices.ContainsFunc(lines, longerThan(60)))
	fmt.Println(slices.IndexFunc(lines, longerThan(60)))

	clean := slices.DeleteFunc(slices.Clone(lines), hasPrefix("//"))
	fmt.Println(len(clean), clean[1])
}
```

```text
true
3
3 Keep lines short.
```

`longerThan(60)` reads nicely at the call site, almost like English. (Notice the
`slices.Clone` before `DeleteFunc`: `DeleteFunc` modifies the slice you give it, and
we don't want to wreck `lines`.)

## Dependency injection for handlers

The biggest real-world use: HTTP handlers. A handler must have the signature
`func(http.ResponseWriter, *http.Request)`, with no room for a database or config.
So you take those first:

```go
func handleConvert(store DocStore, maxSize int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxSize)
		// ...use store to load and convert the document...
	}
}

mux.Handle("POST /convert", handleConvert(store, 10<<20))
```

That's one level of currying: dependencies first, request later. It's the standard
way to structure Go web servers, and it makes handlers trivial to test with a fake
store.

## Where to stop

One level of "configure, then return a function" is idiomatic Go. Beyond that,
readability drops fast:

```go
// Please don't.
render(cfg)(tmpl)(page)(w)(r)
```

Each extra layer adds a nested `func` type and an anonymous frame in stack traces.
If you have many settings, reach for a struct (`Renderer{Cfg: cfg, Tmpl: tmpl}`) or
a single factory that takes all of them at once.

## Summary

- **Currying**: one argument at a time. Rare in Go beyond one level.
- **Partial application**: pre-fill some arguments. Very common, usually as a closure
  or a method value.
- **The Go sweet spot**: `func thing(config) func(args) result`, which lets your logic
  plug into APIs that demand a fixed signature.
