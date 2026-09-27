---
title: Event Hub
difficulty: hard
after: interfaces-in-depth
hints:
  - 'A `Hub` can''t store `func(Stored)` and `func(Evicted)` in one slice, but it can store what they have in common. In `Subscribe[E]`, wrap `fn` in a non-generic closure: `func(v any) bool { e, ok := v.(E); if ok { fn(e) }; return ok }`. The type assertion `v.(E)` does all the matching: for a concrete `E` it checks the exact dynamic type, for an interface `E` it checks that the dynamic type implements it.'
  - 'Keep subscribers as pointers to a small struct (`deliver func(any) bool`, `active bool`) in subscription order. `cancel` sets `active = false` and removes the subscriber from the hub''s slice. In `Publish`, loop over a **copy** of the slice (`slices.Clone`) so subscribers added mid-publish aren''t seen, and check `active` right before each call so one cancelled mid-publish is skipped.'
  - 'For `Latest[E]`, remember every published event in order and scan backwards for the first `ev.(E)` that succeeds. Record the event *before* delivering it, so a handler (or a nested `Publish`) calling `Latest` already sees it.'
exercise:
  starter: |
    package main

    import "fmt"

    type Hub struct {
    	// your fields here
    }

    func NewHub() *Hub {
    	return &Hub{}
    }

    func (h *Hub) Subscribe[E any](fn func(E)) (cancel func()) {
    	return func() {}
    }

    func (h *Hub) Publish(event any) int {
    	return 0
    }

    func (h *Hub) Latest[E any]() (E, bool) {
    	var zero E
    	return zero, false
    }

    type Stored struct{ Key string }

    type Evicted struct {
    	Key    string
    	Reason string
    }

    func (e Evicted) String() string { return "evicted " + e.Key + " (" + e.Reason + ")" }

    func main() {
    	h := NewHub()
    	h.Subscribe(func(e Stored) { fmt.Println("stored:", e.Key) })
    	stop := h.Subscribe(func(s fmt.Stringer) { fmt.Println("log:", s) })
    	fmt.Println(h.Publish(Stored{"user:1"}))          // want stored: user:1, then 1
    	fmt.Println(h.Publish(Evicted{"user:1", "full"})) // want log: evicted user:1 (full), then 1
    	stop()
    	fmt.Println(h.Publish(Evicted{"user:2", "ttl"})) // want 0
    	fmt.Println(h.Latest[Stored]())                  // want {user:1} true
    }
  solution: |
    package main

    import (
    	"fmt"
    	"slices"
    )

    type subscriber struct {
    	deliver func(any) bool
    	active  bool
    }

    // Hub delivers events to subscribers chosen by the events' dynamic types.
    type Hub struct {
    	subs    []*subscriber
    	history []any
    }

    // NewHub returns an empty hub.
    func NewHub() *Hub {
    	return &Hub{}
    }

    // Subscribe registers fn for every event whose dynamic type is E or, when E
    // is an interface type, implements E. cancel unsubscribes; it's idempotent.
    func (h *Hub) Subscribe[E any](fn func(E)) (cancel func()) {
    	s := &subscriber{active: true}
    	s.deliver = func(v any) bool {
    		e, ok := v.(E)
    		if ok {
    			fn(e)
    		}
    		return ok
    	}
    	h.subs = append(h.subs, s)
    	return func() {
    		if !s.active {
    			return
    		}
    		s.active = false
    		h.subs = slices.DeleteFunc(h.subs, func(x *subscriber) bool { return x == s })
    	}
    }

    // Publish delivers event to every matching subscriber, in subscription
    // order, and returns how many received it.
    func (h *Hub) Publish(event any) int {
    	if event == nil {
    		return 0
    	}
    	h.history = append(h.history, event)
    	n := 0
    	for _, s := range slices.Clone(h.subs) {
    		if s.active && s.deliver(event) {
    			n++
    		}
    	}
    	return n
    }

    // Latest returns the most recently published event that matches E.
    func (h *Hub) Latest[E any]() (E, bool) {
    	for i := len(h.history) - 1; i >= 0; i-- {
    		if e, ok := h.history[i].(E); ok {
    			return e, true
    		}
    	}
    	var zero E
    	return zero, false
    }

    type Stored struct{ Key string }

    type Evicted struct {
    	Key    string
    	Reason string
    }

    func (e Evicted) String() string { return "evicted " + e.Key + " (" + e.Reason + ")" }

    func main() {
    	h := NewHub()
    	h.Subscribe(func(e Stored) { fmt.Println("stored:", e.Key) })
    	stop := h.Subscribe(func(s fmt.Stringer) { fmt.Println("log:", s) })
    	fmt.Println(h.Publish(Stored{"user:1"}))
    	fmt.Println(h.Publish(Evicted{"user:1", "full"}))
    	stop()
    	fmt.Println(h.Publish(Evicted{"user:2", "ttl"}))
    	fmt.Println(h.Latest[Stored]())
    }
  tests: |
    package main

    import (
    	"errors"
    	"fmt"
    	"slices"
    	"testing"
    )

    type Key string

    type Flushed struct{ Count int }

    // *Flushed and Flushed are different event types.

    type QuotaError struct{ Bucket string }

    func (e *QuotaError) Error() string { return "quota exceeded: " + e.Bucket }

    type recorder struct{ log []string }

    func (r *recorder) add(format string, args ...any) {
    	r.log = append(r.log, fmt.Sprintf(format, args...))
    }

    func (r *recorder) check(t *testing.T, name string, want ...string) {
    	t.Helper()
    	if !slices.Equal(r.log, want) {
    		t.Errorf("%s: handlers ran as %q, want %q", name, r.log, want)
    	}
    	r.log = nil
    }

    func TestConcreteTypes(t *testing.T) {
    	h := NewHub()
    	var r recorder
    	h.Subscribe(func(e Stored) { r.add("stored %s", e.Key) })
    	h.Subscribe(func(e Evicted) { r.add("evicted %s", e.Key) })
    	h.Subscribe(func(e Stored) { r.add("stored again %s", e.Key) })
    	h.Subscribe(func(k Key) { r.add("key %s", k) })
    	h.Subscribe(func(s string) { r.add("string %s", s) })
    	h.Subscribe(func(f *Flushed) { r.add("*flushed %d", f.Count) })

    	if n := h.Publish(Stored{"a"}); n != 2 {
    		t.Errorf("Publish(Stored) = %d, want 2", n)
    	}
    	r.check(t, "Publish(Stored)", "stored a", "stored again a")
    	if n := h.Publish(Key("k1")); n != 1 {
    		t.Errorf("Publish(Key) = %d, want 1", n)
    	}
    	r.check(t, "Publish(Key)", "key k1")
    	if n := h.Publish("plain"); n != 1 {
    		t.Errorf("Publish(string) = %d, want 1", n)
    	}
    	r.check(t, "Publish(string): Key and string are different types", "string plain")
    	if n := h.Publish(Flushed{3}); n != 0 {
    		t.Errorf("Publish(Flushed{3}) = %d, want 0: only a *Flushed subscriber exists", n)
    	}
    	r.check(t, "Publish(Flushed value)")
    	if n := h.Publish(&Flushed{4}); n != 1 {
    		t.Errorf("Publish(&Flushed{4}) = %d, want 1", n)
    	}
    	r.check(t, "Publish(*Flushed)", "*flushed 4")
    	if n := h.Publish(42); n != 0 {
    		t.Errorf("Publish(42) with no int subscribers = %d, want 0", n)
    	}
    	if n := h.Publish(nil); n != 0 {
    		t.Errorf("Publish(nil) = %d, want 0", n)
    	}
    	r.check(t, "unmatched events")
    }

    func TestInterfaceSubscribers(t *testing.T) {
    	h := NewHub()
    	var r recorder
    	h.Subscribe(func(e any) { r.add("any %T", e) })
    	h.Subscribe(func(s fmt.Stringer) { r.add("stringer %s", s) })
    	h.Subscribe(func(err error) { r.add("error %v", err) })
    	h.Subscribe(func(e Evicted) { r.add("evicted %s", e.Key) })

    	if n := h.Publish(Evicted{"k", "ttl"}); n != 3 {
    		t.Errorf("Publish(Evicted) = %d, want 3 (any, Stringer, Evicted)", n)
    	}
    	r.check(t, "Publish(Evicted)", "any main.Evicted", "stringer evicted k (ttl)", "evicted k")
    	if n := h.Publish(&QuotaError{"b1"}); n != 2 {
    		t.Errorf("Publish(*QuotaError) = %d, want 2 (any, error)", n)
    	}
    	r.check(t, "Publish(*QuotaError)", "any *main.QuotaError", "error quota exceeded: b1")
    	if n := h.Publish(QuotaError{"b2"}); n != 1 {
    		t.Errorf("Publish(QuotaError value) = %d, want 1: Error has a pointer receiver, so QuotaError isn't an error", n)
    	}
    	r.check(t, "Publish(QuotaError value)", "any main.QuotaError")
    	if n := h.Publish(fmt.Errorf("wrapped: %w", errors.New("x"))); n != 2 {
    		t.Errorf("Publish(fmt.Errorf(...)) = %d, want 2 (any, error)", n)
    	}
    	r.log = nil
    }

    func TestCancel(t *testing.T) {
    	h := NewHub()
    	var r recorder
    	c1 := h.Subscribe(func(e Stored) { r.add("one") })
    	h.Subscribe(func(e Stored) { r.add("two") })
    	c1()
    	c1() // cancelling twice is harmless
    	if n := h.Publish(Stored{"a"}); n != 1 {
    		t.Errorf("Publish after cancelling one of two subscribers = %d, want 1", n)
    	}
    	r.check(t, "after cancel", "two")
    	c3 := h.Subscribe(func(e Stored) { r.add("three") })
    	h.Publish(Stored{"b"})
    	r.check(t, "new subscriber after a cancel", "two", "three")
    	c3()
    	h.Publish(Stored{"c"})
    	r.check(t, "cancel the newest", "two")
    }

    func TestChangesDuringPublish(t *testing.T) {
    	h := NewHub()
    	var r recorder
    	var cancelLater func()
    	h.Subscribe(func(e Stored) {
    		r.add("first %s", e.Key)
    		if e.Key == "a" {
    			cancelLater()
    			h.Subscribe(func(e Stored) { r.add("added %s", e.Key) })
    		}
    	})
    	cancelLater = h.Subscribe(func(e Stored) { r.add("cancelled %s", e.Key) })
    	h.Subscribe(func(e Stored) { r.add("last %s", e.Key) })

    	if n := h.Publish(Stored{"a"}); n != 2 {
    		t.Errorf("Publish(a) = %d, want 2", n)
    	}
    	r.check(t, "cancel + subscribe inside a handler", "first a", "last a")
    	h.Publish(Stored{"b"})
    	r.check(t, "next publish", "first b", "last b", "added b")
    }

    func TestNestedPublishAndLatest(t *testing.T) {
    	h := NewHub()
    	var r recorder
    	h.Subscribe(func(e Evicted) {
    		latest, ok := h.Latest[Evicted]()
    		r.add("evicted %s (latest %s %v)", e.Key, latest.Key, ok)
    		h.Publish(Flushed{1})
    	})
    	h.Subscribe(func(f Flushed) { r.add("flushed %d", f.Count) })
    	h.Subscribe(func(e Evicted) { r.add("evicted again %s", e.Key) })
    	h.Publish(Evicted{"k", "full"})
    	r.check(t, "nested publish runs immediately", "evicted k (latest k true)", "flushed 1", "evicted again k")

    	if f, ok := h.Latest[Flushed](); !ok || f.Count != 1 {
    		t.Errorf("Latest[Flushed]() = %v, %v, want {1}, true", f, ok)
    	}
    	if s, ok := h.Latest[fmt.Stringer](); !ok || s.String() != "evicted k (full)" {
    		t.Errorf("Latest[fmt.Stringer]() = %v, %v, want the Evicted event, true", s, ok)
    	}
    	if e, ok := h.Latest[any](); !ok || e != (Flushed{1}) {
    		t.Errorf("Latest[any]() = %v, %v, want {1} (the most recently published event), true", e, ok)
    	}
    }

    func TestLatest(t *testing.T) {
    	h := NewHub()
    	if s, ok := h.Latest[Stored](); ok || s != (Stored{}) {
    		t.Errorf("Latest[Stored]() on a new hub = %v, %v, want {}, false", s, ok)
    	}
    	if e, ok := h.Latest[error](); ok || e != nil {
    		t.Errorf("Latest[error]() on a new hub = %v, %v, want nil, false", e, ok)
    	}
    	h.Publish(Stored{"a"}) // no subscribers: still remembered
    	h.Publish(Key("k"))
    	h.Publish(Stored{"b"})
    	h.Publish(&QuotaError{"q"})
    	h.Publish(Key("z"))
    	if s, ok := h.Latest[Stored](); !ok || s.Key != "b" {
    		t.Errorf("Latest[Stored]() = %v, %v, want {b}, true", s, ok)
    	}
    	if k, ok := h.Latest[Key](); !ok || k != "z" {
    		t.Errorf("Latest[Key]() = %q, %v, want \"z\", true", k, ok)
    	}
    	if _, ok := h.Latest[string](); ok {
    		t.Errorf("Latest[string]() found an event, but only Key events were published")
    	}
    	if e, ok := h.Latest[error](); !ok || e.Error() != "quota exceeded: q" {
    		t.Errorf("Latest[error]() = %v, %v, want the *QuotaError, true", e, ok)
    	}
    }

    func TestHubsAreIndependent(t *testing.T) {
    	a, b := NewHub(), NewHub()
    	got := 0
    	a.Subscribe(func(Stored) { got++ })
    	if n := b.Publish(Stored{"x"}); n != 0 || got != 0 {
    		t.Errorf("publishing on hub b reached a subscriber of hub a")
    	}
    	if _, ok := a.Latest[Stored](); ok {
    		t.Errorf("hub a's Latest saw an event published on hub b")
    	}
    }
---

Stash's cache, store and metrics all talk through an **event hub**. With Go
1.27's generic methods, a single non-generic `Hub` can offer a typed API:
subscribers write `h.Subscribe(func(e Evicted) {...})` and never type-assert.

Implement `Hub`:

- `Subscribe[E](fn)` registers `fn` and returns a `cancel` function. `fn`
  receives every published event whose **dynamic type** is `E`, or, when `E`
  is an interface type (`fmt.Stringer`, `error`, `any`...), whose dynamic type
  implements `E`. `cancel` stops future deliveries; calling it again does
  nothing.
- `Publish(event)` delivers `event` synchronously to every matching subscriber
  **in subscription order** and returns how many received it. `Publish(nil)`
  delivers nothing and returns 0.
- `Latest[E]()` returns the most recently published event that matches `E`
  (by the same rule), and `true`, whether or not anybody was subscribed. If
  there's none, it returns the zero `E` and `false`.

Handlers may use the hub while an event is being delivered:

- A subscriber **cancelled** mid-publish doesn't receive the rest of that
  event, if it hadn't already.
- A subscriber **added** mid-publish only receives later events.
- A nested `Publish` is delivered immediately, before the outer one continues.
  The outer event already counts for `Latest` while it's being delivered.

## Example

```go
h := NewHub()
h.Subscribe(func(e Stored) { fmt.Println("stored:", e.Key) })
stop := h.Subscribe(func(s fmt.Stringer) { fmt.Println("log:", s) })

h.Publish(Stored{"user:1"})          // stored: user:1       -> 1
h.Publish(Evicted{"user:1", "full"}) // log: evicted user:1 (full) -> 1
stop()
h.Publish(Evicted{"user:2", "ttl"})  // (nobody)             -> 0
h.Latest[Stored]()                   // {user:1}, true
```

## Constraints

- Matching is exactly what a type assertion does: `Key` and `string` are
  different types, as are `Flushed` and `*Flushed`. A type whose `Error` method
  has a pointer receiver only counts as an `error` when you publish a pointer.
- Hubs don't share subscribers or history.
- There's no concurrency here: everything runs on one goroutine.
