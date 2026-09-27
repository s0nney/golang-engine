---
title: Generic Functions as Values
quiz:
  - question: |
      Which line does **not** compile?

      ```go
      func Double[T ~int | ~float64](v T) T { return v * 2 }
      ```
    options:
      - text: '`f := Double[int]`'
      - text: '`var g func(float64) float64 = Double`'
      - text: '`h := Double`'
        correct: true
      - text: '`out := Apply([]int{1, 2}, Double)` (with `func Apply[T, U any](xs []T, f func(T) U) []U`)'
    explanation: |
      A generic function isn't a value until it's instantiated. `h := Double` gives the
      compiler nothing to infer `T` from. Assigning to a typed variable, or passing it to a
      parameter whose type is known (Go 1.21+), supplies that information.
  - question: |
      What does this print?

      ```go
      xs := []int{3, 1, 2}
      slices.SortFunc(xs, cmp.Compare)
      fmt.Println(xs)
      ```
    options:
      - text: '`[1 2 3]`'
        correct: true
      - text: '`[3 2 1]`'
      - text: It doesn't compile, because `cmp.Compare` isn't instantiated
      - text: '`[3 1 2]`'
    explanation: |
      `SortFunc`'s second parameter has type `func(a, b E) int`, and `E` is already known
      to be `int` from `xs`. Since Go 1.21 that's enough to infer `cmp.Compare[int]`
      automatically.
---

In [Learn Functional Programming](/courses/learn-functional-programming/first-class-functions/functions-as-values) you passed functions around as values. Generic functions can be values too, but only once every type parameter is pinned down.

## Instantiate, then use

A generic function on its own is a *template*, not a value. You can't store it in a variable, because there's no single function type to give that variable:

```go
func Double[T ~int | ~float64](v T) T { return v * 2 }

h := Double // cannot use generic function Double without instantiation
```

Instantiate it and you get an ordinary function value:

```go
f := Double[int] // f is a func(int) int
```

## Inference from the target type (Go 1.21)

Since Go 1.21, the compiler can also infer type arguments from **where the function value is going**:

```go
var g func(float64) float64 = Double // T = float64, from g's type
```

The same works when passing a generic function as an argument to a parameter of known function type, including a parameter of *another generic function*:

```go
package main

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
)

func Double[T ~int | ~float64](v T) T { return v * 2 }

func Apply[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, 0, len(xs))
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

func main() {
	xs := []int{3, 1, 2}
	slices.SortFunc(xs, cmp.Compare) // cmp.Compare[int], inferred
	fmt.Println(xs)

	fmt.Println(Apply(xs, Double))       // Double[int]; U = int
	fmt.Println(Apply(xs, strconv.Itoa)) // not generic: U = string
}
```

```
[1 2 3]
[2 4 6]
[1 2 3]
```

In `Apply(xs, Double)`, the compiler first learns `T = int` from `xs`. The second parameter's type then becomes `func(int) U`, and unifying that with `Double`'s generic signature `func(T') T'` gives `T' = int` and `U = int`. Both generic functions get instantiated in one go.

Before Go 1.21 you had to write `slices.SortFunc(xs, cmp.Compare[int])` and `Apply(xs, Double[int])`. Both forms still work, and the explicit one is sometimes clearer.

## Generic methods need it too

Go 1.27's generic methods follow the same rule. A method value like `inv.First` for a generic method `First[T Item]()` must be instantiated, `inv.First[*Sword]`, before it can be stored or passed. The [Learn OOP lesson on generic methods](/courses/learn-oop/generics-and-oop/generic-methods) covers that case.

## A table of parsers

Instantiated functions are ordinary values, so they can go into maps and slices. That's how you build registries around generic code:

```go
package main

import (
	"fmt"
	"strconv"
)

func ParseInto[T any](parse func(string) (T, error)) func(string) (any, error) {
	return func(s string) (any, error) {
		return parse(s) // T converts to any on return
	}
}

func main() {
	parsers := map[string]func(string) (any, error){
		"int":  ParseInto(strconv.Atoi),
		"bool": ParseInto(strconv.ParseBool),
	}
	for _, kind := range []string{"int", "bool"} {
		v, err := parsers[kind](map[string]string{"int": "42", "bool": "true"}[kind])
		fmt.Printf("%s: %v (%T) %v\n", kind, v, v, err)
	}
}
```

```
int: 42 (int) <nil>
bool: true (bool) <nil>
```

`ParseInto(strconv.Atoi)` infers `T = int` from `Atoi`'s signature and returns a non-generic `func(string) (any, error)`. The generic code is instantiated once per entry, and the map holds plain values. That boundary, where precise generic types meet a dynamic `any`, is something you'll see again with Stash's event bus.

## Recap of the chapter

- Inference unifies **typed arguments** with parameter types, structurally.
- **Untyped constants** only fill in what's left, using the largest kind's default type.
- **Constraints** with a single type term (`~[]E`) can reveal more type parameters.
- It never looks at **results**, and never at **generic types**. Put uninferrable type parameters first so callers can use partial instantiation.
- A generic function becomes a value once instantiated, explicitly or from its target type.
