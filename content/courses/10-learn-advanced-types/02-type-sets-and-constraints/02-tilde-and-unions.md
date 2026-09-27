---
title: The Tilde and Unions
quiz:
  - question: |
      Given `type Bytes int64`, which constraint is **invalid**?
    options:
      - text: '`interface{ ~int64 }`'
      - text: '`interface{ Bytes }`'
      - text: '`interface{ ~Bytes }`'
        correct: true
      - text: '`interface{ Bytes | int64 }`'
    explanation: |
      In `~T`, `T` must be its own underlying type. `Bytes`'s underlying type is `int64`,
      so the compiler says `invalid use of ~ (underlying type of Bytes is int64)`. Write
      `~int64` to mean "`int64` and every type defined from it".
  - question: |
      Why doesn't this compile?

      ```go
      type Printable interface {
      	fmt.Stringer | ~string
      }
      ```
    options:
      - text: '`~string` isn''t allowed in unions'
      - text: A union term can't be an interface that has methods
        correct: true
      - text: Unions need at least three terms
      - text: '`fmt.Stringer` must be written `~fmt.Stringer`'
    explanation: |
      Union terms can be types, `~` terms, or interfaces *without* methods. Mixing
      method-based interfaces into unions would make type sets expensive to reason about,
      so the compiler rejects it: `cannot use fmt.Stringer in union (fmt.Stringer
      contains methods)`.
exercise:
  starter: |
    package main

    import "fmt"

    // Bytes is a size in bytes.
    type Bytes int64

    // Millis is a duration in milliseconds.
    type Millis float64

    // Measure should allow int, int64 and float64, AND any type defined
    // from them, like Bytes and Millis.
    type Measure interface {
    	int | int64 | float64 // TODO: allow named types too
    }

    // Total returns the sum of all values in s.
    func Total[T Measure](s []T) T {
    	var sum T
    	// ?
    	return sum
    }

    // Mean returns the average of s as a float64, or 0 for an empty slice.
    func Mean[T Measure](s []T) float64 {
    	// ?
    	return 0
    }

    func main() {
    	fmt.Println(Total([]int{1, 2, 3}), Mean([]float64{1, 2}))
    	// Uncomment once Measure allows named types:
    	// fmt.Println(Total([]Bytes{512, 1024}), Mean([]Millis{10, 20, 45}))
    }
  solution: |
    package main

    import "fmt"

    // Bytes is a size in bytes.
    type Bytes int64

    // Millis is a duration in milliseconds.
    type Millis float64

    // Measure allows int, int64 and float64, and any type defined from them.
    type Measure interface {
    	~int | ~int64 | ~float64
    }

    // Total returns the sum of all values in s.
    func Total[T Measure](s []T) T {
    	var sum T
    	for _, v := range s {
    		sum += v
    	}
    	return sum
    }

    // Mean returns the average of s as a float64, or 0 for an empty slice.
    func Mean[T Measure](s []T) float64 {
    	if len(s) == 0 {
    		return 0
    	}
    	return float64(Total(s)) / float64(len(s))
    }

    func main() {
    	fmt.Println(Total([]int{1, 2, 3}), Mean([]float64{1, 2}))
    	fmt.Println(Total([]Bytes{512, 1024}), Mean([]Millis{10, 20, 45}))
    }
  tests: |
    package main

    import "testing"

    func TestTotalBuiltin(t *testing.T) {
    	if got := Total([]int{1, 2, 3}); got != 6 {
    		t.Errorf("Total([]int{1, 2, 3}) = %d, want 6", got)
    	}
    	if got := Total([]float64{0.5, 0.25}); got != 0.75 {
    		t.Errorf("Total([]float64{0.5, 0.25}) = %v, want 0.75", got)
    	}
    	if got := Total([]int64(nil)); got != 0 {
    		t.Errorf("Total(nil) = %d, want 0", got)
    	}
    }

    func TestMeanBuiltin(t *testing.T) {
    	if got := Mean([]int{1, 2}); got != 1.5 {
    		t.Errorf("Mean([]int{1, 2}) = %v, want 1.5 (did you divide as integers?)", got)
    	}
    	if got := Mean([]float64{}); got != 0 {
    		t.Errorf("Mean([]float64{}) = %v, want 0", got)
    	}
    }

    func check[T Measure](t *testing.T) {}

    func TestNamedTypes(t *testing.T) {
    	check[Bytes](t)
    	check[Millis](t)
    	if got := Total([]Bytes{512, 1024}); got != 1536 {
    		t.Errorf("Total([]Bytes{512, 1024}) = %d, want 1536", got)
    	}
    	if got := Mean([]Millis{10, 20, 45}); got != 25 {
    		t.Errorf("Mean([]Millis{10, 20, 45}) = %v, want 25", got)
    	}
    }
---

Two small pieces of syntax do most of the work in constraints: the **tilde** and the **union**. Both have sharp edges worth knowing.

## `~T`: every type built on T

A plain type term like `int64` matches exactly one type. `~int64` matches **every type whose underlying type is `int64`**: `int64` itself, `type Bytes int64`, `type Offset int64`, and so on.

```go
type Plain interface{ int64 }  // only int64
type Loose interface{ ~int64 } // int64, Bytes, Offset, time.Duration...
```

Yes, `time.Duration` too: it's `type Duration int64`. Unless you *specifically* mean one exact type, you almost always want the tilde, otherwise your generic function rejects your users' named types.

The rule for what may follow the tilde: **`T` must be its own underlying type**. That means a predeclared type or a type literal, never a defined type, and never an interface:

```go
type Bytes int64

~int64           // fine
~[]byte          // fine: a type literal
~map[string]int  // fine
~Bytes           // error: invalid use of ~ (underlying type of Bytes is int64)
~fmt.Stringer    // error: interfaces can't follow ~
```

`~Bytes` would mean the same set as `~int64` anyway, so the compiler asks you to say it directly.

## Unions

A union `A | B | C` is the set of all types in *any* of its terms. Terms can be exact types, tilde terms, or interfaces **without methods** (which lets you build constraints out of other constraints):

```go
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Integer interface{ Signed | Unsigned }
```

The restrictions:

- **No method interfaces in a union.** `fmt.Stringer | ~string` fails with `cannot use fmt.Stringer in union (fmt.Stringer contains methods)`.
- **No `comparable` in a union**, and no interfaces that embed it.
- **Tilde terms must not overlap**, so `~int | int` is rejected (`overlapping terms int and ~int`): the second term adds nothing.

## What a union buys you

A union constraint lets the body use any operation that's valid for **all** types in the set. With `~int | ~int64 | ~float64`, you get `+`, `-`, `*`, `/`, `<`, and conversion to `float64`:

```go
func Average[T ~int | ~int64 | ~float64](xs []T) float64 {
	var sum T
	for _, x := range xs {
		sum += x // + works for every type in the set
	}
	return float64(sum) / float64(len(xs)) // so does conversion
}
```

Add `~string` to that union and `+` still works (strings concatenate), but `/` and `float64(sum)` don't, so the function stops compiling. The union is a promise in both directions: callers may pass any type in the set, and the body may only do what all of them support. Lesson 5 goes through those rules in detail.

## Your turn

Stash measures things: sizes as `type Bytes int64`, latencies as `type Millis float64`. Its `Measure` constraint only lists exact types, so neither is allowed.

1. Fix `Measure` so it accepts `int`, `int64`, `float64` **and any type defined from them**.
2. Complete `Total`, which sums a slice (0 for an empty one).
3. Complete `Mean`, which returns the average as a `float64` (0 for an empty slice). Careful with integer division!
