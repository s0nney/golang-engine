---
title: Logging and Timing Wrappers
quiz:
  - question: |
      Why does `Timed` in this lesson use `defer` to report the duration?
    options:
      - text: '`defer` makes the function run faster'
      - text: So the duration is reported after the wrapped function returns, even if it panics
        correct: true
      - text: '`time.Since` only works inside deferred functions'
    explanation: |
      A deferred call runs when the surrounding function exits, however it exits.
      The timer covers the whole call and still reports if the wrapped function
      panics.
  - question: 'What''s the benefit of making `Logged` generic, as in `Logged[In, Out any]`?'
    options:
      - text: Generic functions run faster
      - text: One wrapper works for functions of any input and output types, such as `func(string) int` and `func([]byte) string`
        correct: true
      - text: It lets the wrapper change the wrapped function's signature
    explanation: |
      Without generics you'd need a separate logging wrapper for every function
      type. Type parameters let one wrapper cover them all, as long as they take
      one input and return one output.
---

The two most common decorators in any language are "log this call" and "time this
call". Let's build both, using generics so they work for any function with one input
and one output.

## A logging wrapper

```go
package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func Logged[In, Out any](logger *slog.Logger, name string, f func(In) Out) func(In) Out {
	return func(in In) Out {
		logger.Info("call", "func", name, "in", in)
		out := f(in)
		logger.Info("done", "func", name, "out", out)
		return out
	}
}

func wordCount(doc string) int {
	return len(strings.Fields(doc))
}

func main() {
	// A text logger without timestamps, so the output is predictable.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}))

	count := Logged(logger, "wordCount", wordCount)
	shout := Logged(logger, "upper", strings.ToUpper)

	fmt.Println(count("functional go is fun"))
	fmt.Println(shout("doc2doc"))
}
```

```text
level=INFO msg=call func=wordCount in="functional go is fun"
level=INFO msg=done func=wordCount out=4
4
level=INFO msg=call func=upper in=doc2doc
level=INFO msg=done func=upper out=DOC2DOC
DOC2DOC
```

`wordCount` and `strings.ToUpper` know nothing about logging. The logger is passed in
rather than grabbed from a global, which keeps the wrapper easy to test.

## A timing wrapper

```go
func Timed[In, Out any](name string, report func(string, time.Duration), f func(In) Out) func(In) Out {
	return func(in In) Out {
		start := time.Now()
		defer func() { report(name, time.Since(start)) }()
		return f(in)
	}
}
```

Used like this:

```go
render := Timed("render", func(name string, d time.Duration) {
	fmt.Printf("%s took %v\n", name, d)
}, renderMarkdown)

html := render(doc) // prints something like: render took 1.304ms
```

Three details worth copying:

- **`defer` does the reporting.** It runs after `f` returns, even if `f` panics, so
  every call is measured.
- **`report` is a function parameter.** In production it might record a metric; in a
  test it can append to a slice so you can check it was called.
- **The duration isn't printed by the wrapper itself.** The wrapper measures and the
  caller decides what to do with the measurement. Small pieces, loosely joined.

## Stacking them

Because both wrappers return the same type they take, you can stack them:

```go
convert := Logged(logger, "convert", Timed("convert", report, convertDoc))
```

## The limits

Go's generics can't abstract over *how many* parameters a function has. `Logged`
works for `func(In) Out`, but a `func(string, int) (string, error)` needs its own
wrapper (or you bundle the arguments into a struct). Wrappers in Go are usually
written for one well-known signature, which is exactly what the next lesson's
`Converter` type gives you.
