---
title: Named Types and Type Literals
quiz:
  - question: |
      Which of these is a **type literal** (an unnamed type)?
    options:
      - text: '`int`'
      - text: '`UserID`, declared as `type UserID int`'
      - text: '`map[string][]int`'
        correct: true
      - text: '`error`'
    explanation: |
      `map[string][]int` is spelled out from type constructors (`map`, `[]`), so it's a
      type literal. `int` and `error` are predeclared names, and `UserID` is a defined
      type: all three are named types.
  - question: |
      Does this compile?

      ```go
      type UserID int
      type OrderID int

      var u UserID = 7
      var o OrderID = 7
      fmt.Println(u == o)
      ```
    options:
      - text: Yes, and it prints `true`
      - text: Yes, and it prints `false`
      - text: 'No: `UserID` and `OrderID` are different types, so `==` has mismatched types'
        correct: true
      - text: No, because you can't compare integers with `==`
    explanation: |
      Every type definition creates a brand-new named type. Two named types are identical
      only if they come from the *same* declaration, so `UserID` and `OrderID` are
      different even though both are "an `int` underneath". That's the whole point:
      the compiler stops you mixing up user and order IDs.
---

Welcome to the deep end. You've used generics, interfaces and custom types for nine courses. This course is about *how they really work*: the rules the compiler follows, the corners where those rules surprise people, and how to design APIs that use them well.

Along the way you'll build **Stash**, a small generic collections and data-processing library: sets, ordered maps, an LRU cache, a typed event bus, and a struct validator driven by reflection. By the capstone you'll assemble them into a typed, validated, cached repository.

We start with the ground everything else stands on: what a *type* is in Go.

## Two families of types

Every Go type is one of two things.

A **named type** has a name:

- the **predeclared** types: `int`, `string`, `bool`, `float64`, `error`, `any`...
- **defined types** you declare with `type`: `type UserID int`
- **type parameters** inside generic code: the `T` in `func F[T any]`

A **type literal** (the spec also says *unnamed type*) is built on the spot from type constructors:

```go
[]string                  // slice
map[string]int            // map
*Entry                    // pointer
func(string) bool         // function
chan<- Event              // channel
struct{ Key, Val string } // struct
[4]byte                   // array
```

A type literal can mention named types (`*Entry`, `[]UserID`) and still be unnamed itself. The thing on the outside decides.

## Type identity

The first place this matters is **type identity**: when are two types *the same type*?

- A **named type** is identical only to itself. Every `type X ...` declaration creates a new, different type, even if it's spelled exactly like another one.
- Two **type literals** are identical if they have the same structure: same kind of constructor with identical element types (and, for structs, the same field names, types, tags and order).

```go
package main

import "fmt"

type UserID int
type OrderID int

func main() {
	var u UserID = 7
	var o OrderID = 7
	// fmt.Println(u == o) // compile error: mismatched types UserID and OrderID
	fmt.Println(int(u) == int(o))

	var p struct{ X, Y int }
	var q struct{ X, Y int }
	q.X = 3
	p = q // fine: two identical type literals are the same type
	fmt.Println(p)
}
```

```
true
{3 0}
```

`UserID` and `OrderID` are different types, so comparing them is a compile error until you convert both to `int`. The two anonymous structs, on the other hand, were written separately but are one and the same type, because type literals are compared by structure.

## Why Stash cares

Named types are Go's cheapest safety tool. Stash will use them for things like:

```go
type Key string    // a cache or map key
type Tags []string // a named slice with methods
type Bytes int64   // sizes, with a String method
```

A function that takes a `Key` can't be handed any old `string` variable by accident, and `Bytes` can carry methods that a plain `int64` can't.

But named types also create friction: a `[]Key` is not a `[]string`, and a `Tags` is not always interchangeable with `[]string`. Knowing exactly where that friction comes from (underlying types, assignability and conversion) is what the rest of this chapter is about. And it's the foundation for constraints like `~string` later in the course.

## Further reading

- [The Go spec: Types](https://go.dev/ref/spec#Types) and [Type identity](https://go.dev/ref/spec#Type_identity)
