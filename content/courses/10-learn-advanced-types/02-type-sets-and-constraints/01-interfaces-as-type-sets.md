---
title: Interfaces as Type Sets
quiz:
  - question: |
      Which of these can be used as the type of an ordinary variable, like `var x I`?
    options:
      - text: '`interface{ ~int | ~float64 }`'
      - text: '`interface{ String() string; ~int64 }`'
      - text: '`comparable`'
      - text: '`interface{ String() string }`'
        correct: true
    explanation: |
      Only *basic* interfaces, the ones made of nothing but methods, can be variable
      types. As soon as an interface lists type terms or embeds `comparable`, it's a
      *general* interface and may only appear as a constraint.
  - question: |
      What is the type set of this interface?

      ```go
      type Weird interface {
      	int
      	string
      }
      ```
    options:
      - text: '`int` and `string`'
      - text: It's empty, because no type is both an `int` and a `string`
        correct: true
      - text: All types
      - text: It doesn't compile
    explanation: |
      Separate lines in an interface *intersect*. A union needs `|` on one line:
      `int | string`. The empty interface compiles, but no type satisfies it, so any
      generic function constrained by it can never be called.
---

You've written constraints like `cmp.Ordered` and `interface{ ~int | ~float64 }` in [Learn Go](/courses/learn-go/generics-basics/constraints) and [Learn OOP](/courses/learn-oop/generics-and-oop/constraints). This chapter looks at the model underneath them, because it explains every rule (and every error message) you'll meet.

## Two ways to read an interface

Before generics, an interface was a **method set**: `fmt.Stringer` is "anything with `String() string`".

Since Go 1.18, the spec reads every interface as a **type set**: the set of all types that satisfy it.

- `fmt.Stringer`'s type set is every type with a `String() string` method. It's infinite.
- `any` (`interface{}`) has *every* type in its type set.
- `interface{ int | string }` has a type set of exactly two types.

A type satisfies a constraint (or implements an interface) when it's **in the type set**. That one idea covers both methods and unions.

## Building type sets

An interface's body is a list of *elements*, one per line. Each line describes a set, and the interface's type set is the **intersection** of all of them:

```go
type SizeLike interface {
	~int64          // all types whose underlying type is int64...
	String() string // ...AND that have a String method
}
```

Within a single line, `|` means **union**:

```go
type Number interface {
	~int | ~int64 | ~float64 // any of these
}
```

Embedding another interface adds its type set as another line to intersect with:

```go
type OrderedStringer interface {
	cmp.Ordered  // any ordered type...
	fmt.Stringer // ...that is also a Stringer
}
```

So `SizeLike` is satisfied by a `type Bytes int64` with a `String` method, and by nothing else in this program:

```go
package main

import (
	"fmt"
	"slices"
	"strconv"
)

type Bytes int64

func (b Bytes) String() string { return strconv.FormatInt(int64(b), 10) + "B" }

type SizeLike interface {
	~int64
	String() string
}

// Largest can compare (thanks to ~int64) and print (thanks to String).
func Largest[T SizeLike](xs ...T) string {
	return slices.Max(xs).String()
}

func main() {
	fmt.Println(Largest[Bytes](512, 4096, 1024))
}
```

```
4096B
```

Because every type in `SizeLike`'s type set is an `int64` underneath, `slices.Max` can compare them. Because every one has `String`, we can call it.

## Basic vs general interfaces

That power comes with a restriction:

- A **basic interface** lists only methods. It can be used anywhere: as a variable type, a field type, a parameter, *and* as a constraint.
- A **general interface** contains type terms (`int`, `~int`, unions) or embeds `comparable`. It can **only** be used as a constraint.

```go
var n Number // compile error: cannot use type Number outside a type constraint: interface contains type constraints
```

Why? An interface *value* has to be able to hold any member of the type set at runtime and dispatch calls through its methods. A union like `~int | ~float64` has no methods to dispatch through, and nothing to do with a value once you have it. Go may lift the restriction one day, but not yet.

## Empty type sets

Lines intersect, so it's easy to write a constraint nothing satisfies:

```go
type Weird interface {
	int
	string
}
```

It compiles, but the type set is empty. You meant `int | string`. If you ever get an error like `int does not satisfy Weird`, check for a missing `|`.

## Further reading

- [The Go spec: Interface types](https://go.dev/ref/spec#Interface_types)
