---
title: Untyped Constants and Default Types
quiz:
  - question: |
      What does this print?

      ```go
      x := 0.1
      fmt.Println(x+0.2 == 0.3, 0.1+0.2 == 0.3)
      ```
    options:
      - text: '`true true`'
      - text: '`false false`'
      - text: '`false true`'
        correct: true
      - text: '`true false`'
    explanation: |
      `x` is a `float64` variable, so `x+0.2` is computed in binary floating point and
      picks up rounding error. `0.1+0.2 == 0.3` involves only untyped constants, which the
      compiler evaluates exactly, so it's `true`.
  - question: |
      Why does `2 * time.Second` compile but `n * time.Second` (with `n := 2`) doesn't?
    options:
      - text: '`time.Second` can only be multiplied by literals, as a special case'
      - text: '`2` is an untyped constant that becomes a `time.Duration`; `n` is already an `int`, and `int` and `time.Duration` are mismatched types'
        correct: true
      - text: '`n` might be negative'
      - text: Multiplication of variables needs `math.Mul`
    explanation: |
      An untyped constant takes on whatever type the context needs, as long as the value
      fits. `n := 2` gave `n` the default type `int`, and Go never mixes two different
      types in arithmetic. Write `time.Duration(n) * time.Second`.
  - question: What's the type of `v` in `v := 'A'`?
    options:
      - text: '`byte`'
      - text: '`string`'
      - text: '`rune` (that is, `int32`)'
        correct: true
      - text: '`int`'
    explanation: |
      `'A'` is an untyped *rune* constant, and the default type of that kind is `rune`,
      an alias for `int32`. When there's no other type to go by, as with `:=`, the
      default type is used.
---

`const limit = 1 << 20` has no type. Not "an `int` you didn't write down": genuinely **no type**. That's what lets one constant work as an `int8`, a `Bytes` and a `float64` in the same program.

## Typed and untyped constants

```go
const typed int = 10 // a typed constant: it's an int, full stop
const untyped = 10   // an untyped integer constant
```

Untyped constants still have a **kind**: boolean, rune, integer, floating-point, complex or string. The kind decides what they can turn into. When an untyped constant is used somewhere that needs a type (assigned to a variable, passed to a function, combined with a typed operand) it's converted **implicitly** to that type, as long as the value fits.

```go
package main

import "fmt"

type Bytes int64

func main() {
	const limit = 1 << 20

	var small int8 = limit >> 14 // 64 fits in an int8
	var b Bytes = limit          // becomes a Bytes, no conversion needed
	f := limit / 3.0             // 3.0 is a float constant, so this is float division
	i := limit / 3               // both integer constants: integer division

	fmt.Println(small, b, f, i)
}
```

```
64 1048576 349525.3333333333 349525
```

That's rule 5 from the previous lesson: an untyped constant is assignable to any type that can represent its value. `var tiny int8 = 200` fails with `cannot use 200 (untyped int constant) as int8 value in variable declaration (overflows)`.

## Default types

When nothing in the context gives a type (a short variable declaration, an `any` parameter like `fmt.Println`'s) the constant gets its kind's **default type**:

| Kind | Example | Default type |
|---|---|---|
| boolean | `true` | `bool` |
| rune | `'x'` | `rune` (`int32`) |
| integer | `42` | `int` |
| floating-point | `2.0` | `float64` |
| complex | `1i` | `complex128` |
| string | `"s"` | `string` |

That's why `fmt.Printf("%T", 'x')` prints `int32`. And it's the source of a classic surprise:

```go
d := 2 * time.Second               // 2 becomes a time.Duration: fine
n := 2                             // n is an int (the default type)
d = n * time.Second                // compile error: mismatched types int and time.Duration
d = time.Duration(n) * time.Second // fine
```

## Exact arithmetic

Constant expressions are evaluated **exactly**, at compile time, with at least 256 bits of precision. Only the final value has to fit its type:

```go
package main

import "fmt"

func main() {
	const huge = 1 << 100   // far too big for any integer type...
	fmt.Println(huge >> 98) // ...but this result fits in an int

	x := 0.1
	fmt.Println(x+0.2 == 0.3, 0.1+0.2 == 0.3)
}
```

```
4
false true
```

The first comparison uses a `float64` variable and suffers binary rounding. The second is all constants, computed exactly.

## Typed constants and iota

Give a constant a type when it *is* a value of that type, like units or enums. `iota` works nicely with a named type, and every following line repeats the typed expression:

```go
type Bytes int64

const (
	KB Bytes = 1 << (10 * (iota + 1)) // 1 << 10
	MB                                // 1 << 20
	GB                                // 1 << 30
)
```

Now `KB`, `MB` and `GB` are `Bytes`. They can't be added to an `int` by accident, and they'll use any `String` method you give `Bytes`.

## Why this matters for generics

Keep the default types in mind. When you call a generic function with only untyped constants, like `Max(1, 2.5)`, inference has to pick *some* type for `T`, and the default types decide it. Chapter 3 covers exactly how.

## Further reading

- [The Go blog: Constants](https://go.dev/blog/constants)
