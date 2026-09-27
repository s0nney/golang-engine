---
title: Writing Iterators
quiz:
  - question: Which loop fits an `iter.Seq2[int, string]`?
    options:
      - text: '`for line := range seq.Values() { ... }`'
      - text: '`for seq.Next() { ... }`'
      - text: '`for i := 0; i < len(seq); i++ { ... }`'
      - text: '`for n, line := range seq { ... }`'
        correct: true
    explanation: |
      `iter.Seq2[K, V]` yields pairs, so you range over it with two variables, just
      like ranging over a slice or a map. You can also write `for n := range seq`
      to take only the first value of each pair.
  - question: |
      By convention, what should a method named `All` on a collection type return?
    options:
      - text: An iterator over every element, such as `iter.Seq[T]` or `iter.Seq2[K, V]`
        correct: true
      - text: A slice copy of every element
      - text: The number of elements
    explanation: |
      The standard library uses `All` for "iterate over everything", as in
      `slices.All` and `maps.All`, and `Backward`, `Keys` and `Values` for other
      orders and views. Following the convention makes your types feel familiar.
---

Now let's write iterators for Doc2Doc's own types, and meet `iter.Seq`'s two-valued
sibling.

## iter.Seq2: pairs

```go
type Seq2[K, V any] func(yield func(K, V) bool)
```

`Seq2` yields two values at a time, like the index and value of a slice or the key
and value of a map. Doc2Doc's error messages need line numbers, so here's an iterator
that yields `(lineNumber, line)` pairs:

```go
package main

import (
	"fmt"
	"iter"
	"strings"
)

func NumberedLines(doc string) iter.Seq2[int, string] {
	return func(yield func(int, string) bool) {
		n := 1
		for line := range strings.Lines(doc) {
			if !yield(n, strings.TrimSuffix(line, "\n")) {
				return
			}
			n++
		}
	}
}

func main() {
	doc := "# Title\nsome text   \nmore text\n"
	for n, line := range NumberedLines(doc) {
		if strings.HasSuffix(line, " ") {
			fmt.Printf("line %d: trailing whitespace\n", n)
		}
	}
}
```

```text
line 2: trailing whitespace
```

`NumberedLines` is itself built on top of another iterator, `strings.Lines`.
Iterators compose nicely.

## Iterator methods on your types

Collection types conventionally offer iterator methods named after what they yield:
`All`, `Backward`, `Keys`, `Values`. Here's a Doc2Doc section tree with an `All`
method that walks every section, depth first:

```go
package main

import (
	"fmt"
	"iter"
)

type Section struct {
	Title string
	Subs  []*Section
}

// All yields every section in the tree with its depth.
func (s *Section) All() iter.Seq2[int, *Section] {
	return func(yield func(int, *Section) bool) {
		s.walk(0, yield)
	}
}

func (s *Section) walk(depth int, yield func(int, *Section) bool) bool {
	if !yield(depth, s) {
		return false
	}
	for _, sub := range s.Subs {
		if !sub.walk(depth+1, yield) {
			return false
		}
	}
	return true
}

func main() {
	doc := &Section{Title: "Guide", Subs: []*Section{
		{Title: "Install", Subs: []*Section{{Title: "Linux"}, {Title: "macOS"}}},
		{Title: "Usage"},
	}}
	for depth, s := range doc.All() {
		fmt.Printf("%*s%s\n", depth*2, "", s.Title)
	}
}
```

```text
Guide
  Install
    Linux
    macOS
  Usage
```

This is recursion from chapter 4 hiding behind a simple `for` loop. The caller has no
idea there's a tree walk going on. Notice that `walk` returns a `bool`: if the loop
body breaks, `yield` returns `false`, and that `false` must travel all the way back
up the recursion so every level stops.

## Iterators vs returning a slice

Why not just return a `[]*Section`?

- **Laziness.** An iterator does work only as the loop asks for it. Break after the
  first match, and the rest of the tree is never visited.
- **No allocation.** No slice is built to hold every element.
- **Encapsulation.** Callers can't modify your internal slice through the iterator.

A slice is still better when the caller needs random access, `len`, or to loop over
the data several times. And if they want a slice anyway, `slices.Collect(seq)` makes
one.

## Further reading

- [The Go Blog: Range Over Function Types](https://go.dev/blog/range-functions)
