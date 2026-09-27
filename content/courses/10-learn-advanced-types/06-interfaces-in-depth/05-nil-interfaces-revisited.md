---
title: Nil Interfaces, Revisited
quiz:
  - question: |
      What does this print?

      ```go
      func IsNil[T any](v T) bool { return any(v) == nil }

      var p *Entry
      fmt.Println(IsNil[error](nil), IsNil(p))
      ```
    options:
      - text: '`true true`'
      - text: '`true false`'
        correct: true
      - text: '`false false`'
      - text: It doesn't compile, because you can't compare a `T` to `nil`
    explanation: |
      For `T = error`, `v` is a nil interface, and converting it to `any` keeps it nil. For
      `T = *Entry`, `any(v)` is an interface whose type word says `*Entry` and whose value
      is a nil pointer, which is *not* a nil interface.
  - question: |
      Why doesn't `func IsNil[T any](v T) bool { return v == nil }` compile?
    options:
      - text: Type parameters can never be compared
      - text: '`nil` isn''t a valid value for every type in `any`''s type set (an `int` can''t be `nil`)'
        correct: true
      - text: It needs `T comparable`
      - text: Generic functions can't return `bool`
    explanation: |
      Comparing with `nil` is an operation, and operations on a `T` must be valid for every
      type in its type set. Even `comparable` wouldn't help, since `int` is comparable but
      has no `nil`. Converting to `any` compiles, but, as the lesson shows, answers a
      different question.
---

[Learn OOP](/courses/learn-oop/polymorphism/the-nil-interface-gotcha) covered the classic bug: an interface holding a nil pointer is not a nil interface. With the two-word picture from lesson 1, it's easy to see why: a nil interface has **both** words empty, while an interface holding a nil `*Entry` has a type word saying `*Entry` and an empty data word. Generics add a couple of new ways to trip over it.

## The generic IsNil that isn't

```go
package main

import (
	"errors"
	"fmt"
)

type Entry struct{ Key string }

type ValidationError struct{ Field string }

func (e *ValidationError) Error() string { return e.Field + " is invalid" }

// IsNil looks reasonable, but only detects nil *interfaces*.
func IsNil[T any](v T) bool { return any(v) == nil }

func check(name string) error {
	var verr *ValidationError
	if name == "" {
		verr = &ValidationError{Field: "name"}
	}
	return verr // BUG: a nil *ValidationError is a non-nil error
}

func main() {
	fmt.Println(IsNil[error](nil), IsNil[any](nil))
	fmt.Println(IsNil[*Entry](nil), IsNil[[]int](nil), IsNil[map[string]int](nil))

	err := check("stash")
	fmt.Println(err == nil, err)

	if v, ok := errors.AsType[*ValidationError](check("")); ok {
		fmt.Println("invalid field:", v.Field)
	}
}
```

```
true true
false false false
false <nil>
invalid field: name
```

`IsNil` only reports `true` when `T` is itself an interface type holding nothing. For pointers, slices and maps it's always `false`, even when they're nil, because `any(v)` wraps them with a type word. And you can't write `v == nil` for a `T any` at all: `nil` isn't a value of every type.

If you genuinely need "is this nil?" in generic code:

- **Constrain the shape.** A function that takes `p *T` can compare `p == nil` directly. Same for `s []T` and `m map[K]V`.
- **Use reflection** when the kind is unknown: `reflect.ValueOf(&v).Elem()` gives you a `reflect.Value` whose `IsNil` method works for pointers, maps, slices, channels, functions and interfaces. Chapter 7 covers it.
- **Or redesign** so you don't need to ask. A `(T, bool)` result (chapter 4) usually removes the question entirely.

## Returning concrete error types

The other half of the output is the classic bug in its natural habitat. `check` declares `var verr *ValidationError` and returns it as an `error`. When there's nothing wrong, it returns a **typed nil**, and the caller's `err == nil` is `false`. Worse, `fmt` prints `<nil>`, which makes the bug look impossible in logs.

The rule: **functions that return `error` should return a literal `nil` on success**, never a nil variable of a concrete type:

```go
func check(name string) error {
	if name == "" {
		return &ValidationError{Field: "name"}
	}
	return nil
}
```

## Getting the concrete type back

When callers need the details, `errors.AsType` (Go 1.26) is the generic way to ask. It walks the wrap chain and returns a typed value:

```go
if v, ok := errors.AsType[*ValidationError](err); ok {
	fmt.Println("invalid field:", v.Field)
}
```

Note the type argument is the **pointer** type, because that's the type whose method set has `Error`. `errors.AsType[ValidationError]` would fail to compile: `ValidationError does not satisfy error (method Error has pointer receiver)`, which is the addressability rule from lesson 2 surfacing in a constraint.

## Zero values of interface type parameters

One more consequence for generic containers: if `T` is an interface type, `var zero T` is a nil interface. A `Stack[error]` that returns `(zero, false)` from an empty `Pop` returns a nil `error`, which is exactly what callers expect, but only if they check the `bool`. If `T` is `*Entry`, the zero is a nil pointer. Either way, the `bool` is what makes `(T, bool)` APIs safe for every `T`.
