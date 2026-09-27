---
title: Wrap-Up and What's Next
quiz:
  - question: You need to process values of many unrelated struct types, guided by their field tags, in a library that has never seen those types. What's the right tool?
    options:
      - text: A type parameter with an `any` constraint
      - text: Reflection, hidden behind a small typed or `any`-accepting API
        correct: true
      - text: A type switch listing every struct type
      - text: '`unsafe.Pointer`'
    explanation: |
      Generics can't see fields or tags, and a type switch can't list types it doesn't
      know about. Reading arbitrary struct layouts at runtime is exactly reflection's job,
      as in `encoding/json` and Stash's `Validate`.
  - question: 'Which of these is **not** something Go 1.27 generics can do?'
    options:
      - text: Declare a method with its own type parameters
      - text: Use a generic type alias
      - text: Declare an interface method with type parameters
        correct: true
      - text: Infer type arguments from an assigned function type
    explanation: |
      Generic methods (1.27), generic aliases (1.24) and inference from function
      types (1.21) all work. Interface methods still can't have type parameters, and a
      generic method can't satisfy an interface method.
---

You started this course knowing how to *use* generics and interfaces. You finish it knowing how they *work*, and with a small library to show for it.

## What you covered

1. **The type system**: named types and type literals, underlying types, identity, assignability and conversion, untyped constants, and aliases, including generic aliases.
2. **Type sets and constraints**: interfaces as sets of types, `~T` and unions, `comparable` after Go 1.20, `cmp.Ordered` and its NaN caveats, and exactly which operations a `T` allows.
3. **Type inference**: unification, untyped constants, constraint type inference with `S ~[]E`, and explicit and partial instantiation.
4. **Designing generic APIs**: zero values and comma-ok, the `PT interface{ *T; M() }` pattern, receiver gotchas, and choosing between generic methods and functions.
5. **Generic data structures**: `Set`, `OrderedMap`, an `LRU` cache, a heap over `container/heap`, and iterator APIs that plug into `slices` and `maps`.
6. **Interfaces in depth**: the two words inside an interface value, addressability, optional interfaces, type switches in generic code, nil interfaces, and the typed event bus.
7. **Reflection**: `Type` and `Value`, kinds, fields and tags, settability, calling methods, the laws, and a tag-driven validator.
8. **Performance and limits**: GC shapes and dictionaries, measuring with `b.Loop` and `AllocsPerRun`, the things generics can't express, and a glimpse of `unsafe`.
9. **The capstone**: all of it assembled into a validated, cached, observable `Repo[K, V]`.

## The one-paragraph version

Reach for **concrete code** first. Use **interfaces** when behaviour varies and **generics** when the same code works for many types; combine them freely, with `any` hidden inside typed edges when you need heterogeneous storage. Keep **reflection** for truly unknown types, behind a small API. And when performance matters, **measure**: GC shapes and dictionaries make generic arithmetic as fast as concrete code, and method calls on type parameters about as costly as interface calls.

## Ideas for growing Stash

- Make `Cached` safe for concurrent use with a `sync.Mutex`, and test it under the race detector.
- Cache `Validate`'s parsed rules per `reflect.Type` in a `sync.Map`, and benchmark the difference.
- Add a `Set`-based secondary index: `repo.Index(func(e Entry) string { return e.Owner })`.
- Write a `SQLStore[K, V]` that satisfies `Store[K, V]` using `database/sql` and the `Columns[T]` helper from chapter 7.
- Expose `Where` results as a paginated API with `slices.Chunk`.

## Next up: cryptography

Types keep data *correct*. The next course is about keeping it *secret* and *trustworthy*. You'll build Keybox, a small end-to-end-encrypted secrets vault, and learn the primitives behind it properly: randomness, hashes, MACs, key derivation, authenticated encryption, key exchange, signatures and TLS, all with Go's standard library.

Continue with [Learn Cryptography](/courses/learn-cryptography).
