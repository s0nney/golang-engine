---
title: The Laws of Reflection, and When Not to Use It
quiz:
  - question: |
      Which task is the best fit for reflection?
    options:
      - text: Summing a `[]int` and a `[]float64` with one function
      - text: Calling `Area()` on a mix of shapes
      - text: Encoding *any* user-defined struct to a text format, driven by its field tags
        correct: true
      - text: A type-safe `Set` for any comparable element type
    explanation: |
      Summing numbers of different types is a job for generics, calling `Area` on shapes
      is an interface's job, and a `Set` is a generic type. Walking the fields of arbitrary
      structs you've never seen, guided by tags, is exactly what reflection is for.
  - question: According to the third law of reflection, what must be true to modify a value through a `reflect.Value`?
    options:
      - text: The value must be exported
      - text: The `reflect.Value` must be settable, meaning it refers to the original, addressable storage (typically reached through a pointer and `Elem`)
        correct: true
      - text: The type must implement `reflect.Setter`
      - text: You must call `reflect.Unlock` first
    explanation: |
      Settability is about whether the change could be seen: a `Value` made from a copy
      isn't settable, one reached through a pointer is. (Exported-ness also matters for
      struct fields, but it's settability that the law is about.)
---

The Go blog's classic post [The Laws of Reflection](https://go.dev/blog/laws-of-reflection) boils `reflect` down to three rules. With chapter 6's picture of interface values, they're short.

## The three laws

1. **Reflection goes from an interface value to a reflection object.** `reflect.ValueOf(x)` and `reflect.TypeOf(x)` take an `any` and unpack its two words.
2. **Reflection goes from a reflection object back to an interface value.** `v.Interface()` packs a type and value back into an `any`, which you assert to a concrete type. (Or `reflect.TypeAssert[T](v)` in one step.)
3. **To modify a reflection object, the value must be settable.** It has to refer to the original storage, which in practice means you started from a pointer and called `Elem()`.

```go
package main

import (
	"fmt"
	"reflect"
)

type Bytes int64

func main() {
	var size Bytes = 2048

	// Law 1: interface value -> reflection object.
	v := reflect.ValueOf(size)
	fmt.Println(v.Type(), v.Kind(), v.Int())

	// Law 2: reflection object -> interface value.
	back := v.Interface().(Bytes)
	fmt.Println(back == size)

	// Law 3: to modify a reflection object, it must be settable.
	fmt.Println(v.CanSet())
	p := reflect.ValueOf(&size).Elem()
	p.SetInt(4096)
	fmt.Println(size)
}
```

```
main.Bytes int64 2048
true
false
4096
```

Everything in this chapter is these three laws plus the Kind switch.

## The costs

Reflection is powerful, and it isn't free:

- **No compile-time checking.** Wrong kinds, wrong argument counts and unexported fields become **runtime panics**. Every reflective function needs tests for the shapes it claims to support.
- **Speed.** A reflective field read or method call does many times the work of a direct one: kind checks, flag checks, often an allocation for boxing results. For encoding a config file once, irrelevant. For a hot loop, measurable.
- **Readability.** `v.Field(i).SetInt(n)` says far less than `cfg.Workers = n`. Readers can't jump to definitions, and refactoring tools can't see through strings like `MethodByName("Put")`.
- **Binary size**, as the last lesson showed, when methods are looked up dynamically.

## Choosing the right tool

Before you reach for `reflect`, walk down this list and stop at the first thing that works:

1. **Concrete code.** One type? Just write it.
2. **Interfaces.** Different types with different behaviour behind a shared method set.
3. **Generics.** The same algorithm or container for many types.
4. **Code generation.** Per-type code that would be tedious to write by hand (`go generate` with a generator, as `stringer` does for `String` methods). It's typed and fast, at the cost of a build step.
5. **Reflection.** Truly arbitrary types you can't know in advance, usually driven by struct tags: encoders and decoders, validators, ORMs, dependency injection, test helpers like `reflect.DeepEqual`.

## Using it well when you do

- **Hide it.** Expose a typed or tag-driven API (`json.Marshal(v)`, `Validate(v)`) and keep `reflect.Value`s inside.
- **Cache per type.** Work out a struct's fields and parsed tags once per `reflect.Type` (a `sync.Map` keyed by type is common), not on every call. `encoding/json` does this.
- **Fail with good errors, not panics**, for anything a caller could get wrong, like passing a non-pointer.
- **Combine with generics** at the edge: `Columns[T]()` and `Validate[T](v T)` make the type explicit and avoid nil-interface surprises.

Next, you'll put all of this together in Stash's validator.

## Further reading

- [The Laws of Reflection](https://go.dev/blog/laws-of-reflection)
