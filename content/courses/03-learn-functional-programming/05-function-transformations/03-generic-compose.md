---
title: A Generic Compose
quiz:
  - question: |
      Given `func Compose[A, B, C any](f func(A) B, g func(B) C) func(A) C`, what is the type of `Compose(strings.Fields, countItems)` if `countItems` is a `func([]string) int`?
    options:
      - text: '`func([]string) int`'
      - text: '`func(string) []string`'
      - text: '`func(string) int`'
        correct: true
      - text: It doesn't compile
    explanation: |
      `strings.Fields` is `func(string) []string`, so `A` is `string` and `B` is
      `[]string`. `countItems` makes `C` an `int`. The result is `func(A) C`, which
      is `func(string) int`.
  - question: Why can't Go write a single variadic `Compose(fs ...)` that chains functions of *different* types?
    options:
      - text: Go doesn't allow variadic generic functions
      - text: Variadic parameters must all have the same type, and each step would need its own type parameters
        correct: true
      - text: Go only allows two type parameters per function
    explanation: |
      `...T` means "any number of values of one type T". A chain `string → []string →
      int → bool` needs a different type at each link, which Go's type system can't
      express in one parameter list. You compose two at a time instead.
---

The `compose` from the last lesson only worked on `Transform`, which takes a string
and returns a string. Real pipelines change types along the way: Doc2Doc takes a
document (`string`), splits it into words (`[]string`), counts them (`int`) and then
decides whether it's too long (`bool`).

Generics let you write a `Compose` that connects *any* two compatible functions.

## Three type parameters

```go
func Compose[A, B, C any](f func(A) B, g func(B) C) func(A) C {
	return func(a A) C {
		return g(f(a))
	}
}
```

Read the signature like a chain of dominoes: `f` goes from `A` to `B`, `g` goes from
`B` to `C`, so together they go from `A` to `C`. The shared `B` is what makes the
two functions fit together, and the compiler checks it.

## Doc2Doc: from text to verdict

```go
package main

import (
	"fmt"
	"strings"
)

func Compose[A, B, C any](f func(A) B, g func(B) C) func(A) C {
	return func(a A) C {
		return g(f(a))
	}
}

func count(words []string) int { return len(words) }

func tooLong(n int) bool { return n > 5 }

func main() {
	wordCount := Compose(strings.Fields, count)          // string -> int
	isTooLong := Compose(wordCount, tooLong)             // string -> bool
	shout := Compose(strings.TrimSpace, strings.ToUpper) // string -> string

	doc := "  Functional programming in Go is fun and practical  "
	fmt.Println(wordCount(doc))
	fmt.Println(isTooLong(doc))
	fmt.Println(shout(doc))
}
```

```text
8
true
FUNCTIONAL PROGRAMMING IN GO IS FUN AND PRACTICAL
```

Type inference does all the work. You never write `Compose[string, []string, int]`,
because Go figures out `A`, `B` and `C` from the arguments. And if you try to compose
functions that don't fit, like `Compose(count, strings.ToUpper)`, you get a compile
error rather than a runtime crash, since an `int` isn't a `string`.

## Going longer

What about three or more functions of different types? Go has no way to write one
variadic `Compose` for that, because every item in a `...` parameter must have the
same type. You have two options.

**Nest the calls**, as `isTooLong` does above: `Compose(Compose(f, g), h)`.

**Write fixed-size helpers**, if you need them often:

```go
func Compose3[A, B, C, D any](f func(A) B, g func(B) C, h func(C) D) func(A) D {
	return func(a A) D { return h(g(f(a))) }
}
```

When every step has the *same* type, the variadic `pipe` from the last lesson works
fine as a generic function:

```go
func Pipe[T any](steps ...func(T) T) func(T) T {
	return func(v T) T {
		for _, step := range steps {
			v = step(v)
		}
		return v
	}
}
```

## A note on style

Generic composition is neat, but heavy use of it makes Go code look alien to other
Go programmers. Long chains of `Compose(Compose(...))` are hard to debug, because a
stack trace shows anonymous `func1` frames instead of names you recognise. Use
composition where it clarifies, such as building configurable pipelines, and write
straight-line code where it doesn't.
