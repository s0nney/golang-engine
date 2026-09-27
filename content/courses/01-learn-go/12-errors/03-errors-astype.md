---
title: Inspecting Errors with errors.AsType
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import (
      	"errors"
      	"fmt"
      )

      type rateLimitError struct {
      	retryAfter int
      }

      func (e *rateLimitError) Error() string { return "rate limited" }

      func main() {
      	err := fmt.Errorf("send: %w", &rateLimitError{retryAfter: 30})
      	if rl, ok := errors.AsType[*rateLimitError](err); ok {
      		fmt.Println("retry in", rl.retryAfter)
      	} else {
      		fmt.Println("give up")
      	}
      }
      ```
    options:
      - text: '`give up`, because the error was wrapped'
      - text: '`retry in 30`'
        correct: true
      - text: '`retry in 0`'
      - text: It doesn't compile
    explanation: |
      `errors.AsType` unwraps the chain looking for an error of type
      `*rateLimitError`. It finds one inside the wrapper, returns it with
      `ok == true`, and its `retryAfter` field is 30.
  - question: When should you use `errors.AsType` instead of `errors.Is`?
    options:
      - text: When you want to check for one specific error *value*, such as `ErrOptedOut`
      - text: When you need an error of a particular *type* so you can read its fields
        correct: true
      - text: They do the same thing; `AsType` is just the newer name
    explanation: |
      `errors.Is` asks "is this particular error value in the chain?".
      `errors.AsType` asks "is there an error of this *type* in the chain?",
      and hands it back, fields and all.
---

`errors.Is` answers "is this a particular error?". But sometimes you need more than yes or no. If a carrier says "rate limited, retry in 30 seconds", you want that **30**. The information lives in the fields of a custom error type, so you need to pull the error out of the chain *as that type*.

## `errors.AsType`

Go 1.26 added `errors.AsType`, a generic function that searches the error chain for an error of a given type:

```go
func AsType[E error](err error) (E, bool)
```

You put the type you're looking for in square brackets. It returns the matching error and `true`, or the zero value and `false` if nothing in the chain has that type:

```go
package main

import (
	"errors"
	"fmt"
)

type carrierError struct {
	carrier    string
	retryAfter int // seconds
}

func (e *carrierError) Error() string {
	return fmt.Sprintf("%s is busy", e.carrier)
}

func send(to string) error {
	err := &carrierError{carrier: "Alpha Mobile", retryAfter: 30}
	return fmt.Errorf("send to %s: %w", to, err)
}

func main() {
	err := send("+1-555-0100")
	fmt.Println(err)

	if ce, ok := errors.AsType[*carrierError](err); ok {
		fmt.Printf("%s asked us to retry in %ds\n", ce.carrier, ce.retryAfter)
	}
}
```

```text
send to +1-555-0100: Alpha Mobile is busy
Alpha Mobile asked us to retry in 30s
```

The `if ..., ok := ...; ok` shape should look familiar: it's the comma-ok idiom again, just like map lookups. Inside the `if`, `ce` is a `*carrierError`, so you can read its fields directly.

The square brackets are **type arguments**, part of Go's generics, which you'll study in the last main chapter. For now, read `errors.AsType[*carrierError](err)` as "find me a `*carrierError` in `err`".

## Pointer or value?

Look carefully at the type you ask for. The `Error` method above has a pointer receiver, so it's `*carrierError` (the pointer) that satisfies `error`, and that's what `send` returned. You must ask for `*carrierError`, with the star. Asking for plain `carrierError` doesn't even compile, because the struct type itself has no `Error` method, so it isn't an `error`.

## The older `errors.As`

Before Go 1.26, you'd write the same thing with `errors.As`, which needs a variable to fill in:

```go
var ce *carrierError
if errors.As(err, &ce) {
	fmt.Println(ce.retryAfter)
}
```

It works, but it's clunkier: you declare the variable separately, it lives outside the `if`, and it's easy to pass the wrong kind of pointer. `errors.AsType` is shorter, type-safe and keeps the variable scoped to the `if`. You'll still see `errors.As` in older code, and `go fix` can modernise it for you.

## Choosing the right tool

| You want to know... | Use |
|--------------------|-----|
| Did anything go wrong? | `err != nil` |
| Is this a specific error value? | `errors.Is(err, ErrOptedOut)` |
| Is there an error of this type, and what's in it? | `errors.AsType[*carrierError](err)` |

Most of the time, `err != nil` is all you need. Reach for `Is` and `AsType` when your code has to react differently to different failures.

## Further reading

- [Package errors: AsType](https://pkg.go.dev/errors#AsType)
