---
title: Types and Values
quiz:
  - question: |
      With `type Bytes int64`, what do `reflect.TypeFor[Bytes]()` and its `.Kind()` print?
    options:
      - text: '`int64 int64`'
      - text: '`main.Bytes int64`'
        correct: true
      - text: '`main.Bytes main.Bytes`'
      - text: '`main.Bytes named`'
    explanation: |
      The `Type` is the exact type, `main.Bytes`. The `Kind` is the category of its
      underlying type, one of a fixed list (`Int64`, `Struct`, `Slice`, `Pointer`...), so
      every type defined from `int64` has kind `int64`.
  - question: Why does `reflect.TypeOf(error(nil))` return `nil`, while `reflect.TypeFor[error]()` returns the `error` interface type?
    options:
      - text: '`TypeOf` is buggy for interfaces'
      - text: '`TypeOf` takes an `any`, and a nil `error` converts to a nil `any` with no dynamic type left to report; `TypeFor` gets the type from its type argument, so there''s no value involved'
        correct: true
      - text: '`TypeFor` only works for interface types'
      - text: They return the same thing
    explanation: |
      `TypeOf` always reports a *dynamic* type, and a nil interface has none. That's the
      classic reason people wrote `reflect.TypeOf((*error)(nil)).Elem()`. Since Go 1.22,
      `reflect.TypeFor[error]()` says the same thing directly.
---

Everything so far has been checked at compile time. **Reflection** is the other end of the spectrum: inspecting and manipulating values whose types you only discover at runtime. It's how `encoding/json` fills your structs, how `fmt` prints anything, and how Stash's validator will read struct tags at the end of this chapter.

## Two core types

The `reflect` package revolves around two types:

- **`reflect.Type`** describes a Go type: its name, kind, fields, methods, element type, and so on.
- **`reflect.Value`** holds a value of some type, and lets you read (and sometimes set) it.

You get into reflection through an interface value, the two words from chapter 6. `reflect.TypeOf(x)` returns the type word; `reflect.ValueOf(x)` returns both, wrapped up.

```go
package main

import (
	"fmt"
	"reflect"
)

type Bytes int64

type Entry struct {
	Key  string
	Size Bytes
}

func main() {
	e := Entry{Key: "logo.png", Size: 2048}

	t := reflect.TypeOf(e) // from a value (via any)
	fmt.Println(t, t.Name(), t.Kind())

	fmt.Println(reflect.TypeFor[Bytes](), reflect.TypeFor[Bytes]().Kind())
	fmt.Println(reflect.TypeFor[*Entry](), reflect.TypeFor[*Entry]().Kind(), reflect.TypeFor[*Entry]().Elem())
	fmt.Println(reflect.TypeFor[error]().Kind(), reflect.TypeOf(error(nil)))

	v := reflect.ValueOf(e)
	size := v.Field(1)
	fmt.Println(size.Kind(), size.Int(), size.Type())

	key, ok := reflect.TypeAssert[string](v.Field(0))
	fmt.Println(key, ok)
	fmt.Println(v.Interface().(Entry).Size)
}
```

```
main.Entry Entry struct
main.Bytes int64
*main.Entry ptr main.Entry
interface <nil>
int64 2048 main.Bytes
logo.png true
2048
```

## Type vs Kind

This is the distinction everything else hangs on:

- The **Type** is the exact type: `main.Bytes`, `*main.Entry`, `[]string`, `main.Set[string]`.
- The **Kind** is the *category* of its underlying type, from a fixed list: `Bool`, `Int`...`Int64`, `Uint`..., `Float64`, `String`, `Slice`, `Array`, `Map`, `Struct`, `Pointer`, `Interface`, `Func`, `Chan`, and a few more.

Named types don't have their own kinds: `Bytes` has kind `Int64`, just like `int64` and `time.Duration`. Reflection code mostly **switches on Kind** (to decide how to walk a value) and **compares Types** (to check for one exact type).

## Getting a Type

- **`reflect.TypeOf(x any)`** gives the *dynamic* type of `x`. Because it takes an `any`, a nil interface has nothing to report, which is why `reflect.TypeOf(error(nil))` is `nil`.
- **`reflect.TypeFor[T]()`** (Go 1.22) gives the type `T`, with no value needed. It works for interface types too: `reflect.TypeFor[error]()`. Prefer it whenever you know the type statically.
- **`t.Elem()`** steps from a pointer, slice, array, map or channel type to its element type.

## Reading a Value

A `reflect.Value` has one getter per family of kinds: `Int()` for all signed integer kinds (returned as `int64`), `Uint()`, `Float()`, `String()`, `Bool()`, `Len()`, `Index(i)`, `Field(i)`, `MapIndex(k)`, `Elem()`... Calling the wrong one **panics**: `reflect: call of reflect.Value.Int on float64 Value`. The Kind tells you which getter is safe.

To get back to ordinary Go values:

- **`v.Interface()`** returns an `any`, which you then type-assert.
- **`reflect.TypeAssert[T](v)`** (Go 1.25) does both in one step and returns `(T, bool)`, avoiding the detour through `any` (and, for many types, an allocation).

## Reflection and generics

They look like rivals, but they meet in the middle. Generic code can call `reflect.TypeFor[T]()` to learn about its own type parameter, which is how a generic `Validate[T]` could cache per-type information. And reflection sees instantiated generic types as ordinary types: `reflect.TypeFor[Set[string]]()` prints `main.Set[string]`. What reflection can't do is create *new* instantiations at runtime; `Set[X]` for a type `X` discovered at runtime doesn't exist unless the compiler already built it.

## Further reading

- [The Laws of Reflection](https://go.dev/blog/laws-of-reflection) on the Go blog (lesson 6 summarises them)
