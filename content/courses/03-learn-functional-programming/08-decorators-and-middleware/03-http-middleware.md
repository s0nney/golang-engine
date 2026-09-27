---
title: Middleware
quiz:
  - question: 'With `type Converter func(string) string`, what is the signature of a Doc2Doc middleware?'
    options:
      - text: '`func(string) string`'
      - text: '`func(Converter)`'
      - text: '`func(Converter) Converter`'
        correct: true
      - text: '`func(Converter) string`'
    explanation: |
      Middleware takes the next converter and returns a new converter that wraps it.
      Because the input and output types match, middleware can be stacked as many
      times as you like.
  - question: 'In this lesson, what happens when `maxSize(20)` receives a 32-byte document?'
    options:
      - text: It truncates the document to 20 bytes and passes it on
      - text: It returns a placeholder without calling `next`, so the real converter never runs
        correct: true
      - text: It calls `next` and then throws away the result
      - text: It panics
    explanation: |
      Middleware decides whether to call `next` at all. Returning early
      short-circuits the chain, so `markdownToHTML` never runs, which is why its
      "runs" line is missing from the second conversion's output.
  - question: 'Why is `maxSize` written as `func(limit int) Middleware` instead of being a `Middleware` itself?'
    options:
      - text: Because Go doesn't allow functions with three levels of nesting
      - text: Because it needs a setting (`limit`), so it's a function that *builds* a middleware, which is currying from the last chapter
        correct: true
      - text: Because middleware must never take arguments
    explanation: |
      Configurable middleware takes its settings first and returns a
      `Middleware`: `maxSize(20)(convert)`. Simple middleware like `logCalls`
      already has the right shape.
---

Wrapping one function is handy. Once you have several wrappers that all take and
return the same function type, you have **middleware**: layers you can put around
any converter, in any combination. It's the most common form of the decorator idea
in real Go code.

## One shape for every converter

Every Doc2Doc converter has the same shape: a document goes in, a converted document
comes out. Give that shape a name, and give "something that wraps a converter" a name
too:

```go
// Converter turns a document in one format into another.
type Converter func(string) string

// Middleware wraps a Converter and returns an enhanced one.
type Middleware func(Converter) Converter
```

`Converter` is the same shape as the `Transform` type from earlier in the chapter. A
`Middleware` is any function that takes the *next* converter and returns a new one
that calls it (or doesn't).

## Two middlewares for Doc2Doc

Every converter needs logging, and none of them should waste time on enormous
documents. Neither concern belongs inside `markdownToHTML`, so write them as
middleware:

```go
package main

import (
	"fmt"
	"strings"
)

// Converter turns a document in one format into another.
type Converter func(string) string

// Middleware wraps a Converter and returns an enhanced one.
type Middleware func(Converter) Converter

func logCalls(next Converter) Converter {
	return func(doc string) string {
		fmt.Printf("-> converting %d bytes\n", len(doc))
		out := next(doc)
		fmt.Printf("<- produced %d bytes\n", len(out))
		return out
	}
}

func maxSize(limit int) Middleware {
	return func(next Converter) Converter {
		return func(doc string) string {
			if len(doc) > limit {
				return "<!-- document too large -->"
			}
			return next(doc) // only reached for small documents
		}
	}
}

func markdownToHTML(doc string) string {
	fmt.Println("   markdownToHTML runs")
	title := strings.TrimPrefix(doc, "# ")
	return "<h1>" + title + "</h1>"
}

func main() {
	var convert Converter = markdownToHTML
	convert = maxSize(20)(convert)
	convert = logCalls(convert)

	for _, doc := range []string{"# Hello", "# A heading that is far too long"} {
		fmt.Println(convert(doc))
	}
}
```

```text
-> converting 7 bytes
   markdownToHTML runs
<- produced 14 bytes
<h1>Hello</h1>
-> converting 32 bytes
<- produced 27 bytes
<!-- document too large -->
```

Notice what happened on the second document. `logCalls` ran (it's the outer layer),
but `maxSize` stopped the document and `markdownToHTML` never ran. `logCalls` still
logged the placeholder coming back out, because from the outside a short-circuit
looks like any other result.

## Two kinds of middleware

- **Simple middleware** like `logCalls` has the signature
  `func(Converter) Converter` directly, so it *is* a `Middleware`.
- **Configurable middleware** like `maxSize` needs settings, so it's a function
  that *returns* a middleware. That's currying from the last chapter:
  `maxSize(20)(convert)`.

## Before, after, or instead

Everything a middleware can do falls into three spots:

- **Before** calling `next`: inspect or change the input (trim whitespace, normalise
  line endings, reject bad input).
- **After** `next` returns: inspect or change the output (log its size, add a footer,
  cache it).
- **Instead of** calling `next`: short-circuit with a different answer, like
  `maxSize` does.

`markdownToHTML` stays a plain, pure function that knows nothing about logging or
limits. Each concern lives in one small wrapper you can test on its own.

## Coming back in HTTP

You'll meet this exact pattern again in [Learn HTTP Servers](/courses/learn-http-servers),
where Go's web handlers are wrapped with middleware of the shape
`func(http.Handler) http.Handler`. Different type, same idea.

Wrapping by hand, one assignment per layer, gets tedious. The next lesson adds a
`Chain` helper.
