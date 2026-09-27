---
title: Designing the Repository
quiz:
  - question: Why is `Store[K, V]` an interface rather than a generic struct?
    options:
      - text: Interfaces are faster than generic types
      - text: 'Different backends (an in-memory `OrderedMap`, a cache in front of it, a database later) have different behaviour behind the same methods, which is what interfaces are for'
        correct: true
      - text: Generic types can't have methods
      - text: Interfaces can't have type parameters
    explanation: |
      The element types are generic (`K`, `V`), but the *behaviour* varies by backend. That
      combination, a generic interface, lets `Cached` wrap any `Store` and lets `Repo`
      accept any of them.
  - question: 'In Stash''s repository, where does reflection run?'
    options:
      - text: On every `Get`
      - text: Only in `Validate`, once per `Put`, to read the value's `validate` tags
        correct: true
      - text: In the LRU cache, to compare keys
      - text: Everywhere, since the repository is generic
    explanation: |
      Everything else is statically typed: generic collections, a generic interface, and
      typed event methods. Reflection is confined to the one job that needs it, reading
      struct tags of a type the library has never seen, and hidden behind `Validate`.
---

Over eight chapters you've built Stash's parts one at a time. The capstone assembles them into a single, useful type: a **repository**, `Repo[K, V]`, that stores values by key, validates them on the way in, caches reads, announces changes, and iterates in a predictable order.

## The pieces you have

| Piece | From | Role in the repository |
|---|---|---|
| `OrderedMap[K, V]` | chapter 5 | the in-memory backend, iterating in insertion order |
| `LRU[K, V]` | chapter 5 | the read cache |
| iterators (`iter.Seq2`) | chapter 5 | `All()` and filtered views |
| `Bus` with generic `Subscribe`/`Publish` | chapter 6 | change events |
| `Validate` with struct tags | chapter 7 | rejecting bad values |
| API design rules | chapters 3 and 4 | inference-friendly constructors, comma-ok results |

## The layers

```
Repo[K, V]            validates, publishes events, filters
  └── Store[K, V]     (interface)
        Cached[K, V]  LRU cache in front of another Store
          └── Store[K, V]
                *OrderedMap[K, V]   the actual data
```

The seam in the middle is a **generic interface**:

```go
// Store is anything that stores values by key.
type Store[K comparable, V any] interface {
	Get(k K) (V, bool)
	Set(k K, v V)
	Delete(k K) bool
	All() iter.Seq2[K, V]
}
```

Interfaces *can* have type parameters on the interface type itself (it's only *methods* that can't have their own). A `Store[string, Entry]` is an ordinary interface type after instantiation, so `*OrderedMap[string, Entry]` satisfies it with no extra code, and so will `*Cached[string, Entry]`. Swapping the backend for a database-backed store later wouldn't change `Repo` at all.

## Where each technique lives

A good design uses each tool for the job it's best at:

- **Generics** for the containers and the repository itself: `Repo[string, Entry].Get` returns an `Entry`, not an `any`.
- **An interface** for the backend seam, where behaviour varies.
- **`any` inside the `Bus`**, hidden behind typed generic methods.
- **Reflection only in `Validate`**, the one place that must read tags on types Stash has never seen.

## Events

The repository publishes two event types, themselves generic:

```go
type Stored[K comparable, V any] struct {
	Key   K
	Value V
}

type Deleted[K comparable] struct{ Key K }
```

A subscriber writes `repo.Events.Subscribe(func(e Stored[string, Entry]) { ... })`, and `E` is inferred as `Stored[string, Entry]`. Because the bus keys handlers by `typeKey[E]()`, a `Stored[string, Entry]` subscriber never sees a `Stored[int, Config]` event from another repository sharing the bus: different instantiations are different types.

## The plan

1. **Next lesson:** build `Cached[K, V]`, the caching `Store` decorator.
2. **Then:** assemble `Repo[K, V]` from all the pieces, the course's final exercise.
3. **Finally:** a look back, and where to go next.
