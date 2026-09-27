---
title: 'any vs Generics: A Typed Event Bus'
quiz:
  - question: |
      With Stash's `Bus`, what does the second `Publish` return?

      ```go
      var bus Bus
      bus.Subscribe(func(e Stored) { fmt.Println("stored", e.Key) })

      var ev any = Stored{Key: "b"}
      n := bus.Publish(ev)
      ```
    options:
      - text: '`1`, because the dynamic type of `ev` is `Stored`'
      - text: '`0`, because `E` is inferred from the *static* type `any`, and nobody subscribed to `any`'
        correct: true
      - text: It panics in the type assertion
      - text: It doesn't compile
    explanation: |
      Inference sees static types only (chapter 3). `Publish(ev)` is `Publish[any]`, which
      looks up the handlers registered for `any`. The typed API is only as good as the
      static types flowing into it.
  - question: Why does the `Bus` store handlers as `[]any` at all, instead of something fully typed?
    options:
      - text: Generics are slower than `any`
      - text: One bus holds handlers for many different event types, and Go has no type for "a map whose value type depends on the key", so the storage has to be heterogeneous
        correct: true
      - text: '`func(E)` values can''t be stored in slices'
      - text: So users can subscribe with any function signature
    explanation: |
      A `map[K]V` has one value type. Handlers of type `func(Stored)` and `func(Evicted)`
      can only share a map through an interface. Generics make the *edges* typed
      (`Subscribe` and `Publish`), and a guarded assertion bridges the middle.
exercise:
  starter: |
    package main

    import "fmt"

    // Stored is published when an entry is added to Stash.
    type Stored struct{ Key string }

    // Evicted is published when an entry is evicted.
    type Evicted struct {
    	Key    string
    	Reason string
    }

    // Bus delivers events to handlers subscribed to their type.
    // The zero value is ready to use.
    type Bus struct {
    	handlers map[any][]any // typeKey[E]() -> []func(E)
    }

    // typeKey returns a comparable value that's unique to the type E:
    // a nil *E stored in an interface.
    func typeKey[E any]() any { return (*E)(nil) }

    // Subscribe registers fn to be called for every published E.
    func (b *Bus) Subscribe[E any](fn func(E)) {
    	// ?
    }

    // Publish calls every handler subscribed to E, in subscription order,
    // and returns how many there were.
    func (b *Bus) Publish[E any](e E) int {
    	// ?
    	return 0
    }

    func main() {
    	var bus Bus
    	bus.Subscribe(func(e Stored) { fmt.Println("stored", e.Key) })
    	bus.Subscribe(func(e Evicted) { fmt.Println("evicted", e.Key, "because", e.Reason) })
    	bus.Subscribe(func(e Stored) { fmt.Println("audit: stored", e.Key) })

    	fmt.Println(bus.Publish(Stored{Key: "logo.png"}))
    	fmt.Println(bus.Publish(Evicted{Key: "old.css", Reason: "capacity"}))
    	fmt.Println(bus.Publish("nobody listens"))
    }
  solution: |
    package main

    import "fmt"

    // Stored is published when an entry is added to Stash.
    type Stored struct{ Key string }

    // Evicted is published when an entry is evicted.
    type Evicted struct {
    	Key    string
    	Reason string
    }

    // Bus delivers events to handlers subscribed to their type.
    // The zero value is ready to use.
    type Bus struct {
    	handlers map[any][]any // typeKey[E]() -> []func(E)
    }

    // typeKey returns a comparable value that's unique to the type E:
    // a nil *E stored in an interface.
    func typeKey[E any]() any { return (*E)(nil) }

    // Subscribe registers fn to be called for every published E.
    func (b *Bus) Subscribe[E any](fn func(E)) {
    	if b.handlers == nil {
    		b.handlers = make(map[any][]any)
    	}
    	k := typeKey[E]()
    	b.handlers[k] = append(b.handlers[k], fn)
    }

    // Publish calls every handler subscribed to E, in subscription order,
    // and returns how many there were.
    func (b *Bus) Publish[E any](e E) int {
    	hs := b.handlers[typeKey[E]()]
    	for _, h := range hs {
    		h.(func(E))(e)
    	}
    	return len(hs)
    }

    func main() {
    	var bus Bus
    	bus.Subscribe(func(e Stored) { fmt.Println("stored", e.Key) })
    	bus.Subscribe(func(e Evicted) { fmt.Println("evicted", e.Key, "because", e.Reason) })
    	bus.Subscribe(func(e Stored) { fmt.Println("audit: stored", e.Key) })

    	fmt.Println(bus.Publish(Stored{Key: "logo.png"}))
    	fmt.Println(bus.Publish(Evicted{Key: "old.css", Reason: "capacity"}))
    	fmt.Println(bus.Publish("nobody listens"))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func TestPublishReachesOnlyMatchingType(t *testing.T) {
    	var bus Bus
    	var got []string
    	bus.Subscribe(func(e Stored) { got = append(got, "A:"+e.Key) })
    	bus.Subscribe(func(e Evicted) { got = append(got, "evicted:"+e.Key+":"+e.Reason) })
    	bus.Subscribe(func(e Stored) { got = append(got, "B:"+e.Key) })

    	if n := bus.Publish(Stored{Key: "x"}); n != 2 {
    		t.Errorf("Publish(Stored) returned %d, want 2 handlers", n)
    	}
    	if n := bus.Publish(Evicted{Key: "y", Reason: "ttl"}); n != 1 {
    		t.Errorf("Publish(Evicted) returned %d, want 1 handler", n)
    	}
    	want := []string{"A:x", "B:x", "evicted:y:ttl"}
    	if !slices.Equal(got, want) {
    		t.Errorf("handlers saw %q, want %q (in subscription order)", got, want)
    	}
    }

    func TestPublishWithNoSubscribers(t *testing.T) {
    	var bus Bus
    	if n := bus.Publish(42); n != 0 {
    		t.Errorf("Publish on an empty Bus returned %d, want 0", n)
    	}
    	bus.Subscribe(func(n int) {})
    	if n := bus.Publish("text"); n != 0 {
    		t.Errorf("Publish(string) with only an int subscriber returned %d, want 0", n)
    	}
    	if n := bus.Publish(int64(1)); n != 0 {
    		t.Errorf("Publish(int64) with only an int subscriber returned %d, want 0", n)
    	}
    	if n := bus.Publish(7); n != 1 {
    		t.Errorf("Publish(7) with one int subscriber returned %d, want 1", n)
    	}
    }

    func TestSeparateBuses(t *testing.T) {
    	var a, b Bus
    	calls := 0
    	a.Subscribe(func(Stored) { calls++ })
    	if n := b.Publish(Stored{}); n != 0 || calls != 0 {
    		t.Errorf("publishing on bus b reached a handler on bus a")
    	}
    }
---

You've now seen both of Go's tools for "code that works with many types" up close. This last lesson of the chapter puts them side by side, then combines them in Stash's **typed event bus**.

## The trade-offs

| | Interfaces (`any`, `io.Writer`...) | Generics (`[T any]`) |
|---|---|---|
| Checked | method calls at compile time; `any` content only at runtime | everything at compile time |
| One container, many types at once | **yes**: a `[]Shape` holds circles and squares | no: a `List[Circle]` holds only circles |
| Result types | you get the interface back, and assert | you get the exact type back |
| Behaviour per type | dynamic dispatch through methods | the same code for every `T` |
| Cost | two-word values, possible boxing, indirect calls | usually none of those; details in chapter 8 |

A short version: **interfaces abstract over behaviour; generics abstract over types.** If different values need to do *different things*, that's an interface. If you're doing *the same thing* to values of some type you don't know yet, that's a type parameter.

## When you need both

Some problems need heterogeneous storage *and* typed APIs. The event bus is one: subscribers want `func(Stored)`, not `func(any)` plus an assertion, and publishers want to hand over a `Stored`. But one bus holds handlers for *many* event types, and a Go map has one value type.

The answer is a common Go pattern: **typed at the edges, `any` in the middle.**

```go
type Bus struct {
	handlers map[any][]any // typeKey[E]() -> []func(E)
}

func (b *Bus) Subscribe[E any](fn func(E))
func (b *Bus) Publish[E any](e E) int
```

Users see only `Subscribe` and `Publish`, both generic methods (Go 1.27), both fully typed, with `E` inferred from the argument: `bus.Subscribe(func(e Stored) {...})`, `bus.Publish(Stored{Key: "a"})`. Inside, handlers are stored as `any` and asserted back to `func(E)` on the way out. That assertion can't fail: everything filed under `E`'s key was put there by `Subscribe[E]`, as a `func(E)`.

The standard library does the same thing: `atomic.Pointer[T]` is a typed wrapper around an untyped `unsafe.Pointer`, and typed helpers around `sync.Map` or `context.Context` values follow the same shape.

## A key for each type

The map needs a key that's different for each event type. Chapter 7 will give you `reflect.TypeFor[E]()`, but there's a neat trick that needs no reflection:

```go
func typeKey[E any]() any { return (*E)(nil) }
```

A nil `*E` converted to `any` is an interface whose type word says `*E` and whose data word is empty. Two such values are equal exactly when their types are equal (lesson 1's comparison rules), so `typeKey[Stored]() == typeKey[Stored]()` but `typeKey[Stored]() != typeKey[Evicted]()`. It's comparable, cheap, and correct even when `E` is itself an interface type.

## The one catch: static types

Because `E` is inferred, it's the **static** type of the argument. If an event arrives as an `any` (say, decoded from a queue), `bus.Publish(ev)` publishes an `E = any` and reaches only `any` subscribers, whatever is inside. Callers who have an `any` must convert it first, or you can offer a separate method that type-switches. Document it; the quiz for this lesson is exactly this bug.

## Your turn

Implement Stash's `Bus`:

- `Subscribe[E](fn)` appends `fn` to the handlers stored under `typeKey[E]()`, creating the map on first use so the zero `Bus` works.
- `Publish[E](e)` calls every handler for `E`, in subscription order, by asserting each one to `func(E)`, and returns how many it called. With no subscribers, it returns 0.
