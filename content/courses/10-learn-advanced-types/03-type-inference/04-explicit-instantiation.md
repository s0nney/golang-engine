---
title: When Inference Fails
quiz:
  - question: |
      Given `func Convert[To, From ~int | ~float64](v From) To`, which call compiles?
    options:
      - text: '`Convert(3)`'
      - text: '`Convert[float64](3)`'
        correct: true
      - text: '`Convert[, float64](3)`'
      - text: '`var f float64 = Convert(3)`'
    explanation: |
      `To` only appears in the result, so it can't be inferred. You can supply a
      *prefix* of the type arguments, and the rest are inferred: `To = float64` explicitly,
      `From = int` from the untyped constant. There's no syntax for skipping a leading one,
      which is why uninferrable type parameters should come first.
  - question: Why doesn't `var s Set` compile, given `type Set[T comparable] struct{ ... }`?
    options:
      - text: '`Set` needs a constructor'
      - text: A generic type must always be instantiated with type arguments, like `Set[string]`; types are never inferred
        correct: true
      - text: Structs can't have type parameters
      - text: It needs to be `var s *Set`
    explanation: |
      Inference only happens for *function calls*. Everywhere a generic type is named
      (variables, fields, composite literals, conversions) you must spell out the
      type arguments.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strconv"
    )

    // Parse converts s into a T. T is only in the result, so callers must
    // instantiate it explicitly: Parse[int]("42").
    func Parse[T int | float64 | bool | string](s string) (T, error) {
    	var out T
    	var err error
    	switch p := any(&out).(type) {
    	case *string:
    		*p = s
    	case *int:
    		*p, err = strconv.Atoi(s)
    		// TODO: *float64 (strconv.ParseFloat with 64 bits) and *bool (strconv.ParseBool)
    	}
    	return out, err
    }

    // ParseAll parses every string with Parse. It stops at the first error
    // and returns nil and that error.
    func ParseAll[T int | float64 | bool | string](ss []string) ([]T, error) {
    	// ?
    	return nil, nil
    }

    func main() {
    	n, err := Parse[int]("42")
    	fmt.Println(n+1, err)
    	ok, err := Parse[bool]("true")
    	fmt.Println(ok, err)
    	sizes, err := ParseAll[float64]([]string{"1.5", "2", "0.25"})
    	fmt.Println(sizes, err)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strconv"
    )

    // Parse converts s into a T. T is only in the result, so callers must
    // instantiate it explicitly: Parse[int]("42").
    func Parse[T int | float64 | bool | string](s string) (T, error) {
    	var out T
    	var err error
    	switch p := any(&out).(type) {
    	case *string:
    		*p = s
    	case *int:
    		*p, err = strconv.Atoi(s)
    	case *float64:
    		*p, err = strconv.ParseFloat(s, 64)
    	case *bool:
    		*p, err = strconv.ParseBool(s)
    	}
    	return out, err
    }

    // ParseAll parses every string with Parse. It stops at the first error
    // and returns nil and that error.
    func ParseAll[T int | float64 | bool | string](ss []string) ([]T, error) {
    	out := make([]T, 0, len(ss))
    	for _, s := range ss {
    		v, err := Parse[T](s)
    		if err != nil {
    			return nil, err
    		}
    		out = append(out, v)
    	}
    	return out, nil
    }

    func main() {
    	n, err := Parse[int]("42")
    	fmt.Println(n+1, err)
    	ok, err := Parse[bool]("true")
    	fmt.Println(ok, err)
    	sizes, err := ParseAll[float64]([]string{"1.5", "2", "0.25"})
    	fmt.Println(sizes, err)
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func TestParse(t *testing.T) {
    	if n, err := Parse[int]("42"); err != nil || n != 42 {
    		t.Errorf(`Parse[int]("42") = (%d, %v), want (42, nil)`, n, err)
    	}
    	if f, err := Parse[float64]("2.5"); err != nil || f != 2.5 {
    		t.Errorf(`Parse[float64]("2.5") = (%v, %v), want (2.5, nil)`, f, err)
    	}
    	if b, err := Parse[bool]("true"); err != nil || !b {
    		t.Errorf(`Parse[bool]("true") = (%v, %v), want (true, nil)`, b, err)
    	}
    	if s, err := Parse[string]("hi"); err != nil || s != "hi" {
    		t.Errorf(`Parse[string]("hi") = (%q, %v), want ("hi", nil)`, s, err)
    	}
    }

    func TestParseErrors(t *testing.T) {
    	if _, err := Parse[int]("forty-two"); err == nil {
    		t.Errorf(`Parse[int]("forty-two") returned no error`)
    	}
    	if _, err := Parse[bool]("maybe"); err == nil {
    		t.Errorf(`Parse[bool]("maybe") returned no error`)
    	}
    	if _, err := Parse[float64]("1.2.3"); err == nil {
    		t.Errorf(`Parse[float64]("1.2.3") returned no error`)
    	}
    }

    func TestParseAll(t *testing.T) {
    	got, err := ParseAll[int]([]string{"3", "1", "2"})
    	if err != nil || !slices.Equal(got, []int{3, 1, 2}) {
    		t.Errorf(`ParseAll[int]([3 1 2]) = (%v, %v), want ([3 1 2], nil)`, got, err)
    	}
    	bad, err := ParseAll[int]([]string{"1", "two", "3"})
    	if err == nil || bad != nil {
    		t.Errorf(`ParseAll[int]([1 two 3]) = (%v, %v), want (nil, an error)`, bad, err)
    	}
    	empty, err := ParseAll[bool](nil)
    	if err != nil || len(empty) != 0 {
    		t.Errorf(`ParseAll[bool](nil) = (%v, %v), want ([], nil)`, empty, err)
    	}
    }
---

Inference is convenient, but it has limits. When it can't work out a type argument, you supply it yourself. That's **explicit instantiation**, and a little care in how you order type parameters makes it painless for your callers.

## Where inference gives up

**1. A type parameter that appears only in the result.**

```go
func Zero[T any]() T
func New[T any]() *T

Zero()      // cannot infer T
Zero[int]() // fine
```

**2. Arguments that carry no type information.** An untyped `nil` fits any pointer, slice, map, channel or function type, so it can't pick one:

```go
func First[T any](xs []T) T

First(nil)      // in call to First, cannot infer T
First[Key](nil) // fine (and panics at runtime, but that's another story)
```

**3. Generic types.** Inference only applies to *function calls*. Whenever you name a generic type (in a variable, field, composite literal or conversion) you write the type arguments:

```go
var s Set[string]   // not: var s Set
idx := Index[Key]{} // not: Index{}
```

That's also why constructors are handy: `NewSet("a", "b")` can infer `T = string` from its arguments, while `Set{...}` never could.

## Partial instantiation

You can supply a **prefix** of the type arguments and let inference do the rest:

```go
package main

import "fmt"

func Convert[To, From ~int | ~int64 | ~float64](v From) To {
	return To(v)
}

type Millis float64

func main() {
	fmt.Printf("%T\n", Convert[float64](3))       // To explicit, From = int inferred
	fmt.Printf("%T\n", Convert[Millis](int64(5))) // From = int64 inferred
}
```

```
float64
main.Millis
```

There's no syntax for skipping a *leading* type argument, so this only works because `To` comes first. The design rule: **put type parameters that can't be inferred first**, and ones that can at the end. If `Convert` were declared `[From, To ...]`, every call would need both.

## Explicit instantiation inside generic code

Inside a generic function, you often need to instantiate another generic function with your own type parameter: `Parse[T](s)`. There's nothing special about it: `T` is just a type.

The exercise uses one more trick you'll see properly in chapter 6. You can't type-switch on a `T` value directly, but you *can* switch on `any(&out)`, a pointer to your result, and write through it:

```go
var out T
switch p := any(&out).(type) {
case *int:
	*p, err = strconv.Atoi(s) // writes into out
}
```

Notice the constraint in the exercise is `int | float64 | bool | string` **without** tildes. That's deliberate: a `type Gold int` would make `&out` a `*Gold`, which wouldn't match `case *int`, and `Parse` would silently return a zero. Constraints should promise only what the body can deliver.

## Your turn

Stash reads config values from strings. Finish:

- `Parse[T](s)` by adding the `*float64` (`strconv.ParseFloat(s, 64)`) and `*bool` (`strconv.ParseBool`) cases, following the `*int` one.
- `ParseAll[T](ss)`, which parses every string with `Parse[T]`, stopping at the first error and returning `nil` and that error.
