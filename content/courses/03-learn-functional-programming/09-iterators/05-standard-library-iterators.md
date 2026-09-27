---
title: Iterators in the Standard Library
quiz:
  - question: |
      What does this print?

      ```go
      m := map[string]int{"md": 3, "txt": 1, "html": 2}
      fmt.Println(slices.Sorted(maps.Keys(m)))
      ```
    options:
      - text: '`[html md txt]`'
        correct: true
      - text: '`[md txt html]`'
      - text: '`[1 2 3]`'
      - text: The order is random
    explanation: |
      `maps.Keys` yields the keys in random order, but `slices.Sorted` collects them
      into a slice and sorts it, giving `[html md txt]` every time.
  - question: |
      How many times does this loop body run?

      ```go
      for line := range strings.Lines("a\nb\n\nc") {
          _ = line
      }
      ```
    options:
      - text: 4
        correct: true
      - text: 3
      - text: 5
    explanation: |
      `strings.Lines` yields `"a\n"`, `"b\n"`, `"\n"` and `"c"`. The empty line
      counts, and the final line has no newline but is still yielded. Four lines.
  - question: What's the difference between `strings.Split(s, ",")` and `strings.SplitSeq(s, ",")`?
    options:
      - text: '`SplitSeq` splits on every character in the separator'
      - text: '`SplitSeq` removes empty parts'
      - text: '`SplitSeq` returns a lazy `iter.Seq[string]` and doesn''t build a slice'
        correct: true
    explanation: |
      Both produce the same parts. `Split` allocates a `[]string` holding all of
      them, while `SplitSeq` hands them out one at a time for a `for range` loop.
---

Since Go 1.23, iterators are everywhere in the standard library. Knowing them saves
you from writing your own, and saves allocations too.

## slices

| Function | What it yields |
|---|---|
| `slices.All(s)` | index and value pairs, like `range s` (`Seq2`) |
| `slices.Values(s)` | just the values (`Seq`) |
| `slices.Backward(s)` | index and value pairs, from the end |
| `slices.Chunk(s, n)` | sub-slices of up to `n` elements |

And functions that **consume** iterators:

| Function | What it does |
|---|---|
| `slices.Collect(seq)` | builds a new slice from a `Seq` |
| `slices.AppendSeq(s, seq)` | appends a `Seq` to an existing slice |
| `slices.Sorted(seq)` | collects and sorts |
| `slices.SortedFunc(seq, cmp)` | collects and sorts with a comparison function |

## maps

`maps.Keys(m)`, `maps.Values(m)` and `maps.All(m)` iterate over a map (in random
order, as always). `maps.Collect(seq2)` builds a map from a `Seq2`, and
`maps.Insert(m, seq2)` adds pairs to an existing map.

The classic "sorted keys" idiom is now a one-liner: `slices.Sorted(maps.Keys(m))`.

## strings (and bytes)

| Function | Yields |
|---|---|
| `strings.Lines(s)` | each line, **including** its trailing `\n` |
| `strings.SplitSeq(s, sep)` | the parts between separators |
| `strings.FieldsSeq(s)` | the words, split on whitespace |
| `strings.FieldsFuncSeq(s, f)` | the parts split wherever `f` returns true |

The `bytes` package has matching versions for `[]byte`.

## Doc2Doc: a word-frequency report

Let's put it all together. Count word frequencies across a document and print the top
three, without ever building a slice of all the words:

```go
package main

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"
)

func main() {
	doc := `Go makes tools.
Doc2Doc is a Go tool.
Tools convert docs, and Go converts fast.`

	counts := map[string]int{}
	for line := range strings.Lines(doc) {
		for word := range strings.FieldsSeq(line) {
			word = strings.ToLower(strings.Trim(word, ".,"))
			counts[word]++
		}
	}

	words := slices.SortedFunc(maps.Keys(counts), func(a, b string) int {
		return cmp.Or(
			cmp.Compare(counts[b], counts[a]), // most frequent first
			cmp.Compare(a, b),                 // then alphabetical
		)
	})

	for i, w := range slices.All(words[:3]) {
		fmt.Printf("%d. %s (%d)\n", i+1, w, counts[w])
	}

	for chunk := range slices.Chunk(slices.Sorted(maps.Keys(counts)), 4) {
		fmt.Println(chunk)
	}
}
```

```text
1. go (3)
2. tools (2)
3. a (1)
[a and convert converts]
[doc2doc docs fast go]
[is makes tool tools]
```

A few things to notice:

- `strings.Lines` and `strings.FieldsSeq` never build a `[]string`.
- `slices.SortedFunc(maps.Keys(counts), ...)` goes straight from map keys to a
  sorted slice.
- `cmp.Or` returns its first non-zero argument, a neat way to sort by several keys.
- `slices.Chunk` breaks the word list into rows of four for display.

## Where the course has taken you

That word counter uses first-class functions (the comparison), closures (it captures
`counts`), higher-order functions (`SortedFunc`), and lazy iterators, all in plain,
idiomatic Go. That's functional programming the Go way: borrow the ideas that make
code clearer, and leave the rest.

## Further reading

- [The Go Blog: Range Over Function Types](https://go.dev/blog/range-functions)
