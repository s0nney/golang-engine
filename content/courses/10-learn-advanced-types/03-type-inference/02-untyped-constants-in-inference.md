---
title: Untyped Constants in Inference
quiz:
  - question: |
      What does this print?

      ```go
      func Max[T cmp.Ordered](a, b T) T { return max(a, b) }

      fmt.Printf("%T %T %T\n", Max(1, 2), Max(1, 2.5), Max('a', 1))
      ```
    options:
      - text: '`int int int`'
      - text: '`int float64 int32`'
        correct: true
      - text: '`int float64 int`'
      - text: It doesn't compile, because `1` and `2.5` have different types
    explanation: |
      With only untyped constants, Go picks the default type of the "largest" kind
      present: integer < rune < floating-point < complex. `1, 2.5` gives `float64`;
      `'a', 1` gives rune, which is `int32`.
  - question: |
      With `type Gold int` and `var g Gold = 5`, what's `T` in `Max(g, 1)`?
    options:
      - text: '`int`, the default type of `1`'
      - text: '`Gold`'
        correct: true
      - text: Inference fails because `Gold` and `1` differ
      - text: '`any`'
    explanation: |
      Typed arguments are unified first, which fixes `T = Gold`. The untyped constant `1`
      then only has to be representable as a `Gold`, and it is.
---

Untyped constants make inference interesting. `Max(1, 2.5)` has no typed arguments at all, so where does `T` come from?

## Typed arguments go first

Inference runs in two passes:

1. **Typed arguments** are unified with their parameters, as in the last lesson.
2. Any type parameter that's *still* unknown, but has **untyped constant** arguments, gets a type from those constants.

So typed arguments always win. If `g` is a `Gold`, then in `Max(g, 1)` the first pass fixes `T = Gold`, and the constant `1` just needs to fit in a `Gold`:

```go
type Gold int

var g Gold = 5
Max(g, 1)   // T = Gold
Max(g, 1.5) // error: cannot use 1.5 (untyped float constant) as Gold value in argument to Max (truncated)
```

## Only untyped constants: the largest kind wins

When a type parameter's only arguments are untyped constants, Go picks the **default type of the largest kind** among them, in this order:

integer < rune < floating-point < complex

```go
package main

import (
	"cmp"
	"fmt"
)

func Max[T cmp.Ordered](a, b T) T { return max(a, b) }

func main() {
	fmt.Printf("%T %v\n", Max(1, 2), Max(1, 2))
	fmt.Printf("%T %v\n", Max(1, 2.5), Max(1, 2.5))
	fmt.Printf("%T %v\n", Max('a', 1), Max('a', 1))
}
```

```
int 2
float64 2.5
int32 97
```

That rule arrived in **Go 1.21**. Before it, Go used the first constant's default type and then complained if the rest didn't fit, so `Max(1, 2.5)` was a compile error. You may still see that in older answers online.

Mixing untyped constants of *different* basic kinds that can't share a type still fails. `Max(1, "x")` has an integer and a string constant, and there's no type that's both, and the compiler says so: `in call to Max, mismatched types untyped int and untyped string (cannot infer T)`.

## Why this is sensible

The rule mirrors how Go treats constant *expressions*. `1 + 2.5` is an untyped float constant `3.5`: the "bigger" kind wins so no information is lost. Inference picks the same way, so `Max(1, 2.5)` returns `2.5` instead of failing or silently truncating.

## Watch out for the default type

The default type can still surprise you when the function has other type parameters:

```go
func Repeat[T any](v T, n int) []T {
	out := make([]T, n)
	for i := range out {
		out[i] = v
	}
	return out
}

sizes := Repeat(0, 3) // []int, not []Bytes or []int64
```

If you meant `[]Bytes`, say so. Either give the constant a type, `Repeat(Bytes(0), 3)`, or instantiate explicitly, `Repeat[Bytes](0, 3)`. Both are clear, and the second leans on the constant being converted to the now-known `T`. Explicit instantiation is the subject of lesson 4.

## Further reading

- [Go 1.21 release notes: type inference](https://go.dev/doc/go1.21#language)
