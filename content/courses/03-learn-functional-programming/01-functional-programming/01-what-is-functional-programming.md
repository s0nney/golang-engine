---
title: What Is Functional Programming?
quiz:
  - question: Which statement best describes functional programming?
    options:
      - text: A style where every function must be written in one line
      - text: A language feature that only Haskell and Lisp have
      - text: A style that builds programs by composing functions that take inputs and return new values, avoiding shared mutable state
        correct: true
      - text: Any program that uses the `func` keyword
    explanation: |
      Functional programming (FP) is a *style*. You build programs out of small
      functions that turn inputs into outputs, and you avoid changing shared state.
      You can write in that style in almost any language, including Go.
  - question: |
      What does this program print?

      ```go
      func shout(s string) string { return strings.ToUpper(s) + "!" }
      func trim(s string) string  { return strings.TrimSpace(s) }

      func main() {
          fmt.Println(shout(trim("  hi  ")))
      }
      ```
    options:
      - text: '`  HI  !`'
      - text: '`HI!`'
        correct: true
      - text: '`hi!`'
      - text: '`HI  !`'
    explanation: |
      Nested calls run from the inside out. `trim` runs first and returns `"hi"`,
      then `shout` upper-cases it and adds `!`, giving `HI!`.
---

Welcome to functional programming! In this course you'll help build **Doc2Doc**, a
command-line tool that converts, formats and analyses text documents. Doc2Doc is a
perfect fit for a functional style, because almost everything it does is
"take some text in, hand some different text back".

## So what is it?

**Functional programming** (FP) is a way of writing programs by combining functions.
Each function takes some input and returns an output, and the program is built by
plugging the output of one function into the input of the next, like pipes.

A few ideas show up again and again:

- **Functions are values.** You can store them in variables, pass them to other
  functions and return them from functions.
- **Pure functions.** Given the same input, a function always returns the same
  output and doesn't change anything else in the program.
- **Immutability.** Instead of changing data in place, you create new data.
- **Composition.** Big behaviour comes from gluing small functions together.

## A tiny taste

Here's how Doc2Doc might clean up a document title:

```go
package main

import (
	"fmt"
	"strings"
)

func trim(s string) string    { return strings.TrimSpace(s) }
func toUpper(s string) string { return strings.ToUpper(s) }
func exclaim(s string) string { return s + "!" }

func main() {
	title := "  welcome to doc2doc  "
	fmt.Println(exclaim(toUpper(trim(title))))
	fmt.Printf("%q\n", title)
}
```

```text
WELCOME TO DOC2DOC!
"  welcome to doc2doc  "
```

Every function is tiny, easy to test and does exactly one job. None of them change
`title`: they each return a *new* string. When the program prints `title` at the
end, it's untouched. That's the functional mindset in a nutshell.

## It's a style, not a language

Some languages, such as Haskell, are built around FP. Go isn't. Go is a pragmatic
language that happily mixes styles: plain loops, structs with methods, interfaces
*and* functions as values. You don't have to pick one. In this course you'll learn
the functional tools Go gives you, and you'll also learn to spot when a plain `for`
loop is the clearer choice. Being honest about that trade-off is part of writing
good Go.

## Why bother?

Functional code tends to be:

- **Easier to test.** A pure function needs no setup: call it and check the result.
- **Easier to reason about.** You can understand a function by reading it alone,
  without wondering what else changed behind your back.
- **Safer with concurrency.** Data that nobody mutates can be shared between
  goroutines without locks.

By the end of the course, you'll be writing higher-order functions, closures,
middleware, lazy iterators and even sum types, all in idiomatic Go.
