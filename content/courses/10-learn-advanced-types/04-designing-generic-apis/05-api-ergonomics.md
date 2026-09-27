---
title: Generic API Ergonomics
quiz:
  - question: |
      Which signature is the better design?

      ```go
      // A
      func WriteAll[W io.Writer](w W, lines []string) error

      // B
      func WriteAll(w io.Writer, lines []string) error
      ```
    options:
      - text: A, because generics are always faster
      - text: B, because the function only calls methods on `w`, and an interface parameter says that more simply
        correct: true
      - text: A, because B can't accept a `*os.File`
      - text: They're equivalent in every way, so it doesn't matter
    explanation: |
      If all you do with a value is call its methods, an ordinary interface parameter is
      simpler for callers and readers. Type parameters earn their keep when they connect
      types: the input's element type to the output's, or two arguments to each other.
  - question: |
      Stash has `func SortedKeys[K cmp.Ordered, V any](m map[K]V) []K`. A user wants keys ordered case-insensitively. What's the idiomatic addition?
    options:
      - text: A `SortedKeysFunc[K comparable, V any](m map[K]V, cmp func(a, b K) int) []K` variant
        correct: true
      - text: Change `K cmp.Ordered` to `K any` and type-switch on strings inside
      - text: Add a global flag that switches the comparison
      - text: Nothing; users should copy the function
    explanation: |
      The standard library pattern is a pair: one version using the natural order
      (`slices.Sort`, `slices.Index`) and a `Func` version that takes a comparison or
      predicate (`slices.SortFunc`, `slices.IndexFunc`). The `Func` version also works for
      types that aren't ordered at all.
---

Generic code has two audiences: the compiler, which just needs the types to line up, and the humans who call it. This lesson is a checklist for the second audience, drawn from how the standard library's `slices`, `maps` and `iter` packages are designed.

## 1. Use a type parameter only to connect types

Ask what the type parameter *relates*. In `func Max[T cmp.Ordered](a, b T) T`, `T` ties both arguments and the result together: that's real value. In `func WriteAll[W io.Writer](w W, lines []string) error`, `W` relates nothing; an `io.Writer` parameter says the same thing more plainly.

The Go blog post [When To Use Generics](https://go.dev/blog/when-generics) boils it down to this: reach for type parameters when you notice you're about to write the exact same code several times, differing only in the types. Otherwise, write ordinary code.

## 2. Preserve the caller's types

Return what you were given. `S ~[]E` and `M ~map[K]V` keep named types like `Tags` intact (chapter 3), and cost callers nothing:

```go
func Filter[S ~[]E, E any](s S, keep func(E) bool) S // not ([]E) []E
```

## 3. Order type parameters for partial instantiation

Put parameters that can't be inferred **first**, so callers can supply just those: `Convert[To, From]`, `ParseAll[T, PT]`. Keep the list short; three is a lot, and more usually means the function is doing too much.

## 4. Offer a Func variant

Constraints like `comparable` and `cmp.Ordered` are convenient but rigid. Pair them with a variant that takes a function, as the standard library does:

| Natural order / equality | Custom |
|---|---|
| `slices.Sort(s)` | `slices.SortFunc(s, cmp)` |
| `slices.Index(s, v)` | `slices.IndexFunc(s, pred)` |
| `slices.Equal(a, b)` | `slices.EqualFunc(a, b, eq)` |
| `slices.Max(s)` | `slices.MaxFunc(s, cmp)` |

The `Func` version has the weaker constraint (usually `any`), so it also covers structs, case-insensitive strings, and reverse order. Accept comparison functions in the standard shape `func(a, b T) int` so `cmp.Compare` and friends plug straight in.

## 5. Accept and return iterators where it fits

If a function only needs to *walk* its input once, accepting an `iter.Seq[T]` makes it work with slices (`slices.Values`), maps (`maps.Keys`), and every Stash collection. If a collection exposes its contents, an `All() iter.Seq[T]` method is more flexible than returning a copied slice. Chapter 5 builds this into Stash.

## 6. Don't fake generics with `any` and a type switch

```go
func Format[T any](v T) string {
	switch x := any(v).(type) {
	case int: // ...
	case string: // ...
	default:
		panic("unsupported type")
	}
}
```

This compiles for *every* `T` but works for only a few, so mistakes turn into runtime panics instead of compile errors. Either use a union constraint that lists exactly the supported types (as `Parse` did in chapter 3), or accept an interface. A type switch *is* the right tool for optional fast paths with a sensible fallback, which chapter 6 covers.

## 7. Name type parameters for readers

Conventional single letters carry meaning: `T` for a general type, `E` for an element, `S` for a slice, `K` and `V` for keys and values, `M` for a map. When a parameter has a domain meaning, use a word: `func Subscribe[Event any](...)` reads better than `[E any]` in a function that also has elements.

## 8. Document what the constraint means

`comparable` hides a runtime panic for interface types (chapter 2), and `cmp.Ordered` has NaN caveats. Say in the doc comment what happens in those cases, the way `slices.Sort`'s docs mention NaNs.

## Where Stash is heading

With these rules, Stash's next chapter builds real data structures: a `Set[T comparable]`, an insertion-`OrderedMap[K, V]`, an `LRU[K, V]` cache and a generic heap, all exposing iterators so they plug into the rest of the standard library.
