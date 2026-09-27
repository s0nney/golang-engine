---
title: Option and Result
quiz:
  - question: What does an `Option[T]` represent?
    options:
      - text: Either some value of type `T`, or nothing at all
        correct: true
      - text: A value of type `T` that is always present
      - text: A list of `T` values
      - text: A command-line flag
    explanation: |
      `Option` is a two-case sum type: `Some(value)` or `None`. It makes "might be
      missing" part of the type instead of relying on `nil` or a magic value.
  - question: |
      Using this lesson's `Result`, what does this print?

      ```go
      r := Try(strconv.Atoi("12x"))
      doubled := r.Map(func(n int) int { return n * 2 })
      fmt.Println(doubled.Or(-1))
      ```
    options:
      - text: '`24`'
      - text: '`0`'
      - text: It panics
      - text: '`-1`'
        correct: true
    explanation: |
      `strconv.Atoi("12x")` fails, so `r` holds an error. `Map` skips the function
      for an error result and passes the error along, and `Or(-1)` returns the
      fallback because there's no value.
  - question: What is the idiomatic Go equivalent of `Result[T]` in a function signature?
    options:
      - text: '`func f() *T`'
      - text: '`func f() any`'
      - text: '`func f() (T, error)`'
        correct: true
      - text: '`func f() (T, bool, error)`'
    explanation: |
      Multiple return values with an `error` last are Go's built-in way of saying
      "a value or a failure". A `Result` type is sometimes handy inside pipelines,
      but `(T, error)` is what other Go code expects.
---

Two sum types show up in almost every functional language:

- **Option** (also called *Maybe*): either `Some(value)` or `None`.
- **Result** (also called *Either*): either `Ok(value)` or `Err(error)`.

Go handles both ideas with multiple return values: `(T, bool)` for "maybe missing"
(think `v, ok := m[key]`) and `(T, error)` for "might fail". Those are idiomatic and
you should keep using them. But building generic versions is a great exercise, and
they're genuinely handy in pipelines where a value has to travel through several
functions.

## A generic Option

Since Go 1.27 a generic type can have **generic methods**, so `Map` can change the
element type:

```go
type Option[T any] struct {
	value T
	ok    bool
}

func Some[T any](v T) Option[T] { return Option[T]{value: v, ok: true} }
func None[T any]() Option[T]    { return Option[T]{} }

func (o Option[T]) Get() (T, bool) { return o.value, o.ok }

func (o Option[T]) Or(fallback T) T {
	if o.ok {
		return o.value
	}
	return fallback
}

func (o Option[T]) Map[U any](f func(T) U) Option[U] {
	if !o.ok {
		return None[U]()
	}
	return Some(f(o.value))
}
```

The fields are unexported, so nobody can build an `Option` that is "missing" but
secretly holds a value. `Get` returns the familiar `(T, bool)` pair, so an `Option`
converts back to idiomatic Go with no fuss.

(A struct with an `ok` flag isn't a *true* sum type. Its two states just share one
struct. But since the fields are hidden, callers can only see the two valid states.)

## A generic Result

```go
package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Result[T any] struct {
	value T
	err   error
}

func Ok[T any](v T) Result[T]             { return Result[T]{value: v} }
func Err[T any](err error) Result[T]      { return Result[T]{err: err} }
func Try[T any](v T, err error) Result[T] { return Result[T]{value: v, err: err} }

func (r Result[T]) Unwrap() (T, error) { return r.value, r.err }

func (r Result[T]) Or(fallback T) T {
	if r.err != nil {
		return fallback
	}
	return r.value
}

func (r Result[T]) Map[U any](f func(T) U) Result[U] {
	if r.err != nil {
		return Err[U](r.err)
	}
	return Ok(f(r.value))
}

func (r Result[T]) Then[U any](f func(T) (U, error)) Result[U] {
	if r.err != nil {
		return Err[U](r.err)
	}
	return Try(f(r.value))
}

// headingLevel parses "level=3" style settings from a Doc2Doc config line.
func headingLevel(line string) Result[int] {
	_, v, ok := strings.Cut(line, "=")
	if !ok {
		return Err[int](errors.New("missing '='"))
	}
	return Try(strconv.Atoi(v)).
		Then(func(n int) (int, error) {
			if n < 1 || n > 6 {
				return 0, fmt.Errorf("level %d out of range", n)
			}
			return n, nil
		})
}

func main() {
	for _, line := range []string{"level=2", "level=9", "level=two", "level"} {
		tag := headingLevel(line).Map(func(n int) string { return fmt.Sprintf("<h%d>", n) })
		if s, err := tag.Unwrap(); err != nil {
			fmt.Printf("%-10s error: %v\n", line, err)
		} else {
			fmt.Printf("%-10s %s\n", line, s)
		}
	}
}
```

```text
level=2    <h2>
level=9    error: level 9 out of range
level=two  error: strconv.Atoi: parsing "two": invalid syntax
level      error: missing '='
```

`Try` adapts any `(T, error)` function call into a `Result`, and `Unwrap` turns it
back. `Map` and `Then` run the next step only when there's no error, so there's a
single error check at the end instead of one after every step. That's the
"railway" style: values travel on the success track, and the first error switches
to the failure track and stays there.

## Should you use these?

Honestly, **mostly not in public APIs**. Every Go programmer knows `(T, error)`, and
tools, linters and the standard library all expect it. `if err != nil` is verbose, but
it's explicit and familiar. Reasonable places for `Option` and `Result` are inside a
package, for data pipelines or for values stored in collections (a
`[]Result[Doc]` from converting many files concurrently, for example). Convert back to
`(T, error)` at your package's boundary.
