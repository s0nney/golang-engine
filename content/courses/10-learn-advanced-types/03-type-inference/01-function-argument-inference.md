---
title: Function Argument Inference
quiz:
  - question: |
      What happens here?

      ```go
      func Max[T cmp.Ordered](a, b T) T { return max(a, b) }

      var hits int = 3
      var misses int64 = 5
      fmt.Println(Max(hits, misses))
      ```
    options:
      - text: It prints `5`, with `T` inferred as `int64`
      - text: It prints `5`, with `T` inferred as `int`
      - text: 'It doesn''t compile: `int64` of `misses` doesn''t match the inferred type `int` for `T`'
        correct: true
      - text: It prints `5` as an `any`
    explanation: |
      Both parameters have type `T`, so both arguments must have the same type. Go never
      converts typed values implicitly, not even for inference. Convert one yourself:
      `Max(int64(hits), misses)`.
  - question: |
      With `type Tags []string` and `func First[T any](xs []T) T`, what is `T` in `First(Tags{"go"})`?
    options:
      - text: '`Tags`'
      - text: '`string`'
        correct: true
      - text: '`[]string`'
      - text: Inference fails, because `Tags` isn't `[]T`
    explanation: |
      Unification matches `Tags` against `[]T`. They aren't identical, but when one side
      is a named type, inference may look at its underlying type, `[]string`, which does
      match `[]T` with `T = string`.
---

Most of the time you call generic functions without writing a single type argument: `slices.Max(scores)`, not `slices.Max[[]int, int](scores)`. **Type inference** fills them in. This chapter shows how it works, so you can predict when it will succeed, what it will pick, and what to do when it fails.

## The idea: unification

For each call, the compiler lines up every **typed argument** with its **parameter type**, and looks for type arguments that make them identical. That matching process is called **unification**.

```go
func Apply[T, U any](xs []T, f func(T) U) []U
```

Call it as `Apply([]int{1, 2}, strconv.Itoa)`:

1. `[]int` against `[]T` gives `T = int`.
2. `func(int) string` (the type of `strconv.Itoa`) against `func(T) U` gives `T = int` again (consistent, good) and `U = string`.

Every type parameter now has a value, so the call is really `Apply[int, string](...)`.

Unification works **structurally**, through any depth of type literals: `map[K][]V` against `map[string][]Entry` gives `K = string`, `V = Entry`.

## Named types and underlying types

What if an argument's type is a named type like `Tags`, and the parameter is a type literal like `[]T`? They're not identical, but inference is allowed to use the **underlying type** of the named side when matching against a literal:

```go
package main

import "fmt"

type Tags []string

func First[T any](xs []T) T { return xs[0] }

func main() {
	tags := Tags{"go", "types"}
	fmt.Println(First(tags)) // Tags ~ []string, so T = string
}
```

```
go
```

The call is legal because a `Tags` is *assignable* to `[]string` (one side is unnamed). Notice what's lost, though: if `First` returned `[]T`, you'd get back a plain `[]string`, not a `Tags`. Lesson 3 shows how to keep the named type.

## Every use must agree

If a type parameter appears in several places, all the arguments must agree **exactly**. There's no implicit conversion during inference:

```go
var hits int = 3
var misses int64 = 5
Max(hits, misses) // in call to Max, type int64 of misses does not match inferred type int for T
```

Fix it at the call site with a conversion: `Max(int64(hits), misses)`. Writing `Max[int64](hits, misses)` doesn't help, because an `int` variable isn't assignable to an `int64` parameter either.

## Interfaces infer the static type

An argument's **static** type is what inference sees, never the dynamic type inside an interface:

```go
package main

import (
	"errors"
	"fmt"
)

func TypeParam[T any](v T) string {
	return fmt.Sprintf("%T", new(T)) // %T of a *T shows T itself
}

func main() {
	var err error = errors.New("boom")
	fmt.Println(TypeParam(err), TypeParam(42))
	fmt.Printf("%T\n", err)
}
```

```
*error *int
*errors.errorString
```

Inside `TypeParam(err)`, `T` is `error`, the variable's declared type, even though the value inside is an `*errors.errorString`. (Printing `%T` of `v` itself would show the dynamic type, since `fmt` receives it as an `any`.) Inference happens at compile time, and the compiler only knows static types.

## Only arguments, never results

Inference looks at **arguments**. It never looks at how you use the result:

```go
func Zero[T any]() T { var z T; return z }

var n int = Zero() // in call to Zero, cannot infer T
```

Even though the assignment makes it "obvious", Go won't infer `T` from the variable you assign to. Lesson 4 covers the fix: explicit instantiation.

## Further reading

- [Everything You Always Wanted to Know About Type Inference](https://go.dev/blog/type-inference), a deep dive on the Go blog
