---
title: Range Over Functions
quiz:
  - question: What is `iter.Seq[string]`?
    options:
      - text: A slice of strings with extra methods
      - text: A channel of strings
      - text: '`func(yield func(string) bool)`, a function that produces strings by calling `yield`'
        correct: true
      - text: An interface with `Next` and `Value` methods
    explanation: |
      `iter.Seq[V]` is just a named function type. The iterator calls `yield`
      once per value. A `for ... range` loop over it supplies the `yield` function
      for you, and the loop body runs inside it.
  - question: |
      What does this print?

      ```go
      func upTo(n int) iter.Seq[int] {
          return func(yield func(int) bool) {
              for i := 1; i <= n; i++ {
                  if !yield(i * 10) {
                      return
                  }
              }
          }
      }

      func main() {
          for v := range upTo(3) {
              fmt.Print(v, " ")
          }
      }
      ```
    options:
      - text: '`1 2 3 `'
      - text: '`0 10 20 `'
      - text: It doesn't compile, because you can't range over a function
      - text: '`10 20 30 `'
        correct: true
    explanation: |
      Since Go 1.23, `range` works over functions of this shape. Each `yield(i * 10)`
      runs the loop body once with `v` set to that value.
---

Since Go 1.23 you can `range` over a **function**. That sounds odd at first, but
it's the most functional feature Go has added in years, and it's how the standard
library now exposes sequences.

## The problem

Say Doc2Doc wants to process the words of a huge document. `strings.Fields(doc)`
builds a slice of *all* the words up front, which may be millions of strings, even if
you only need the first ten. You want a way to hand out words *one at a time* while
still using a normal `for` loop.

## An iterator is a function

The `iter` package defines:

```go
type Seq[V any] func(yield func(V) bool)
```

An iterator is a function that takes a `yield` callback and calls it once for each
value. That's all. Here's one that produces the words of a document:

```go
package main

import (
	"fmt"
	"iter"
	"unicode"
)

func Words(doc string) iter.Seq[string] {
	return func(yield func(string) bool) {
		start := -1
		for i, r := range doc {
			switch {
			case !unicode.IsSpace(r) && start < 0:
				start = i
			case unicode.IsSpace(r) && start >= 0:
				if !yield(doc[start:i]) {
					return
				}
				start = -1
			}
		}
		if start >= 0 {
			yield(doc[start:])
		}
	}
}

func main() {
	for w := range Words("Doc2Doc  loves\tlazy words") {
		fmt.Printf("[%s] ", w)
	}
	fmt.Println()
}
```

```text
[Doc2Doc] [loves] [lazy] [words] 
```

## What `range` does

When you write

```go
for w := range Words(doc) {
	fmt.Println(w)
}
```

the compiler turns the loop body into the `yield` function and passes it to the
iterator, roughly like this:

```go
Words(doc)(func(w string) bool {
	fmt.Println(w)
	return true // keep going
})
```

Every time the iterator calls `yield(word)`, the loop body runs once. When the body
finishes normally, `yield` returns `true`. If the body executes `break` or `return`,
`yield` returns `false`, and a well-behaved iterator must stop. More on that soon.

## Why this is functional

An iterator is a **higher-order function** (it takes `yield`) returned from another
function (`Words`), and it usually **closes over** its input (`doc`). Everything
you've learned so far comes together here.

It's also **lazy**: `Words` doesn't do any work until someone ranges over it, and it
does only as much as the loop asks for. No slice of all the words is ever built.

## Ranging over integers, too

Go 1.22 also added `for i := range 10`, which counts from 0 to 9. You've been using it
in this course. Together with range-over-func, `range` now works on slices, arrays,
strings, maps, channels, integers and iterator functions.

## Further reading

- [Go by Example: Range over Iterators](https://gobyexample.com/range-over-iterators)
