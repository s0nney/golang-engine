---
title: How Go Compiles Generics
quiz:
  - question: |
      In Go's implementation, which of these instantiations share one compiled copy of `func Sum[T Number](xs []T) T`?
    options:
      - text: '`Sum[int]` and `Sum[float64]`'
      - text: '`Sum[int]` and `Sum[Gold]`, where `type Gold int`'
        correct: true
      - text: None; every type argument gets its own copy
      - text: All of them; Go compiles one copy that boxes everything into interfaces
    explanation: |
      Instantiations are grouped by *GC shape*, and types with the same underlying type
      have the same shape. `int` and `Gold` share code (with different dictionaries);
      `float64` needs different machine instructions, so it gets its own copy.
  - question: Why can calling a method on a value of type parameter type be slower than calling it on a concrete type?
    options:
      - text: Generic code always boxes its arguments into interfaces
      - text: The shared shape code doesn't know which concrete method to call, so it looks the method up through the instantiation's dictionary and makes an indirect call, which also blocks inlining
        correct: true
      - text: Methods on type parameters are interpreted at runtime
      - text: Go re-compiles the function on every call
    explanation: |
      Code for a shape is shared by several types, so a call like `x.Value()` can't be
      resolved at compile time. It goes through a function pointer found in the dictionary,
      much like an interface call. Operators on basic shapes like `int` don't have this
      problem, because the shape itself determines the instruction.
---

Chapter 6 showed what an interface value costs. To reason about generic code's performance, you need the matching picture of what the compiler does with a type parameter. This lesson describes Go's implementation at a high level. The language spec doesn't require any of it, and details change between releases, but the broad strokes have held since Go 1.18.

## Two classic strategies

Languages with generics tend to pick one of two extremes:

- **Monomorphization** (C++ templates, Rust): compile a separate copy of the generic function for **every** type argument. Each copy is as fast as hand-written code, but binaries and compile times grow with every instantiation.
- **Boxing / dictionary passing** (Java's erased generics, broadly): compile **one** copy that works on boxed values and receives type information at runtime. Small binaries, but every value is behind a pointer and every operation is indirect.

## Go's middle road: GC shapes and dictionaries

Go uses a hybrid usually called **GC shape stenciling with dictionaries**:

1. **Stenciling by shape.** The compiler generates one copy of a generic function per **GC shape** of its type arguments. Roughly, a type's shape is its underlying type, as far as the machine and garbage collector are concerned. `int` and `type Gold int` have the same shape. `int` and `float64` don't (they need different instructions). And **all pointer types share a single shape**: to the machine, `*Entry` and `*Config` are just pointers.
2. **Dictionaries.** Each instantiation (each actual type argument list) also gets a **dictionary**, passed to the shared code as a hidden argument. It holds what the shape can't tell the code: the exact runtime types (needed to convert a `T` to `any`, or to `make` a `[]T`), and the method pointers for calling methods on `T` values.

Conceptually, for `func Sum[T Number](xs []T) T`:

```go
// Source:
Sum([]int{1, 2})
Sum([]Gold{3, 4})
Sum([]float64{0.5})

// Roughly what gets compiled:
// Sum[go.shape.int](dict_int, xs)       - shared by int and Gold
// Sum[go.shape.int](dict_Gold, xs)
// Sum[go.shape.float64](dict_float64, xs)
```

Those shape names are real. Build a program that calls a generic `First[T any]` with a `[]Gold` and a `[]*Entry`, and `go tool nm` on the binary lists exactly two copies: `main.First[go.shape.int]` and `main.First[go.shape.*uint8]` (every pointer type shares the `*uint8` shape). You'll meet the same names in CPU profiles. Panic stack traces abbreviate them to `main.First[...]`.

## What that means for speed

The model predicts where generic code is fast and where it isn't:

- **Operators on basic types are fast.** In `Sum`, `total += x` compiles to an integer add in the `int` shape and a float add in the `float64` shape. There's nothing to look up, and generic `Sum` performs like a hand-written `SumInts`.
- **Method calls on type parameters are indirect.** In `func SumValues[T Valuer](xs []T)`, the call `x.Value()` could be `Size.Value` or some other type's method sharing the shape, so the code fetches the method from the dictionary and calls through a pointer, like an interface call. It usually can't be inlined. For pointer type arguments, where every pointer type shares one shape, this is always the case.
- **Converting a `T` to an interface** needs the runtime type from the dictionary, and may allocate, just like any other conversion to an interface.
- **No boxing of values.** A `[]T` of `int`s is a real `[]int` in memory, not a slice of interfaces. That's often the biggest win over `[]any`-based code, and it's why Stash's collections hold values directly.

The compiler can sometimes do better: when a generic function is small enough to inline into a caller with a known type argument, the indirection can disappear. But don't count on it; measure.

## The honest summary

Generics in Go are primarily a tool for **type safety and code reuse**, not a performance feature. They're usually as fast as the equivalent interface code, often faster (no boxing), and for arithmetic on basic types, as fast as concrete code. When the difference matters, the next lesson shows how to find out instead of guessing.

## Further reading

- [Generics implementation: GC shape stenciling](https://github.com/golang/proposal/blob/master/design/generics-implementation-gcshape.md), the design document
