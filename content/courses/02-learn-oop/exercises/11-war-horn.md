---
title: War Horn
difficulty: hard
after: polymorphism
hints:
  - 'Functions can''t be compared with `==`, and the same listener may be subscribed twice, so you can''t find a subscription by its listener. Store a `[]subscription` where each one has a unique `id` (a counter in the `Horn`), and let the returned `unsubscribe` closure remove the entry with **its** id, which also makes a second call harmless.'
  - '`ListenerFunc` works like `http.HandlerFunc`: a named function type with a method `func (f ListenerFunc) Notify(e Event) { f(e) }`. That one line lets any closure be passed where a `Listener` is expected.'
  - 'Listeners can call back into the horn while `Blow` is looping. Loop over a **copy** of the subscriptions (`slices.Clone`), taken when `Blow` starts, so changes wait for the next blow and nested blows are safe. `SubscribeOnce` can then be built on `Subscribe`: wrap `l` in a `ListenerFunc` that cancels its own subscription (and remembers it fired) before calling `l.Notify`.'
exercise:
  starter: |
    package main

    import "fmt"

    // Event is something that happened in the realm.
    type Event struct {
    	Kind   string // e.g. "dragon-sighted"
    	Source string
    	Amount int
    }

    // Listener is anything that wants to hear about events.
    type Listener interface {
    	Notify(e Event)
    }

    // ListenerFunc lets an ordinary function be used as a Listener.
    type ListenerFunc func(e Event)

    // Notify calls f(e).
    func (f ListenerFunc) Notify(e Event) {}

    type Horn struct {
    }

    func (h *Horn) Subscribe(kind string, l Listener) (unsubscribe func()) {
    	return func() {}
    }

    func (h *Horn) SubscribeOnce(kind string, l Listener) (unsubscribe func()) {
    	return func() {}
    }

    func (h *Horn) Listeners(kind string) int {
    	return 0
    }

    func (h *Horn) Blow(e Event) int {
    	return 0
    }

    // Scribe is a Listener that writes down everything it hears.
    type Scribe struct {
    	Lines []string
    }

    func (s *Scribe) Notify(e Event) {
    	s.Lines = append(s.Lines, fmt.Sprintf("%s from %s (%d)", e.Kind, e.Source, e.Amount))
    }

    func main() {
    	var horn Horn
    	scribe := &Scribe{}
    	horn.Subscribe("*", scribe)
    	stop := horn.Subscribe("dragon-sighted", ListenerFunc(func(e Event) {
    		fmt.Println("To arms! A dragon over", e.Source)
    	}))

    	fmt.Println(horn.Blow(Event{"dragon-sighted", "Eastwatch", 1})) // want "To arms! ..." then 2
    	stop()
    	fmt.Println(horn.Blow(Event{"dragon-sighted", "Westwatch", 2})) // want 1
    	fmt.Println(scribe.Lines)                                       // want both events
    }
  solution: |
    package main

    import (
    	"fmt"
    	"slices"
    )

    // Event is something that happened in the realm.
    type Event struct {
    	Kind   string // e.g. "dragon-sighted"
    	Source string
    	Amount int
    }

    // Listener is anything that wants to hear about events.
    type Listener interface {
    	Notify(e Event)
    }

    // ListenerFunc lets an ordinary function be used as a Listener.
    type ListenerFunc func(e Event)

    // Notify calls f(e).
    func (f ListenerFunc) Notify(e Event) { f(e) }

    type subscription struct {
    	id   int
    	kind string
    	l    Listener
    }

    // Horn delivers events to subscribed listeners. The zero Horn is ready to use.
    type Horn struct {
    	subs   []subscription
    	nextID int
    }

    // Subscribe registers l for events of the given kind ("*" means every kind)
    // and returns a function that cancels this subscription.
    func (h *Horn) Subscribe(kind string, l Listener) (unsubscribe func()) {
    	h.nextID++
    	id := h.nextID
    	h.subs = append(h.subs, subscription{id: id, kind: kind, l: l})
    	return func() {
    		h.subs = slices.DeleteFunc(h.subs, func(s subscription) bool { return s.id == id })
    	}
    }

    // SubscribeOnce is like Subscribe, but l hears at most one event.
    func (h *Horn) SubscribeOnce(kind string, l Listener) (unsubscribe func()) {
    	fired := false
    	unsubscribe = h.Subscribe(kind, ListenerFunc(func(e Event) {
    		if fired {
    			return
    		}
    		fired = true
    		unsubscribe()
    		l.Notify(e)
    	}))
    	return unsubscribe
    }

    // Listeners returns how many subscriptions would hear an event of kind.
    func (h *Horn) Listeners(kind string) int {
    	n := 0
    	for _, s := range h.subs {
    		if s.kind == kind || s.kind == "*" {
    			n++
    		}
    	}
    	return n
    }

    // Blow delivers e to every matching subscription, in the order they were
    // made, and returns how many were notified. Subscriptions added or removed
    // while it runs take effect from the next Blow.
    func (h *Horn) Blow(e Event) int {
    	n := 0
    	for _, s := range slices.Clone(h.subs) {
    		if s.kind == e.Kind || s.kind == "*" {
    			s.l.Notify(e)
    			n++
    		}
    	}
    	return n
    }

    // Scribe is a Listener that writes down everything it hears.
    type Scribe struct {
    	Lines []string
    }

    func (s *Scribe) Notify(e Event) {
    	s.Lines = append(s.Lines, fmt.Sprintf("%s from %s (%d)", e.Kind, e.Source, e.Amount))
    }

    func main() {
    	var horn Horn
    	scribe := &Scribe{}
    	horn.Subscribe("*", scribe)
    	stop := horn.Subscribe("dragon-sighted", ListenerFunc(func(e Event) {
    		fmt.Println("To arms! A dragon over", e.Source)
    	}))

    	fmt.Println(horn.Blow(Event{"dragon-sighted", "Eastwatch", 1})) // To arms! ... then 2
    	stop()
    	fmt.Println(horn.Blow(Event{"dragon-sighted", "Westwatch", 2})) // 1
    	fmt.Println(scribe.Lines)
    }
  tests: |
    package main

    import (
    	"slices"
    	"strings"
    	"testing"
    )

    // watcher is the test's own Listener: it appends "name:kind" to a shared log.
    type watcher struct {
    	name string
    	log  *[]string
    }

    func (w *watcher) Notify(e Event) { *w.log = append(*w.log, w.name+":"+e.Kind) }

    func ev(kind string) Event { return Event{Kind: kind, Source: "test"} }

    func expect(t *testing.T, what string, log []string, want ...string) {
    	t.Helper()
    	if !slices.Equal(log, want) {
    		t.Errorf("%s: listeners heard [%s], want [%s]", what, strings.Join(log, " "), strings.Join(want, " "))
    	}
    }

    func TestListenerFunc(t *testing.T) {
    	var got Event
    	var l Listener = ListenerFunc(func(e Event) { got = e })
    	l.Notify(Event{"raid", "Northgate", 7})
    	if got != (Event{"raid", "Northgate", 7}) {
    		t.Errorf("ListenerFunc(f).Notify(e) passed %+v to f, want {raid Northgate 7}", got)
    	}
    }

    func TestBlowMatchesKindsInOrder(t *testing.T) {
    	var log []string
    	var h Horn // the zero Horn must work
    	h.Subscribe("raid", &watcher{"a", &log})
    	h.Subscribe("*", &watcher{"all", &log})
    	h.Subscribe("dragon", &watcher{"b", &log})
    	h.Subscribe("raid", &watcher{"c", &log})

    	if n := h.Blow(ev("raid")); n != 3 {
    		t.Errorf("Blow(raid) = %d, want 3", n)
    	}
    	expect(t, "Blow(raid)", log, "a:raid", "all:raid", "c:raid")

    	log = nil
    	if n := h.Blow(ev("feast")); n != 1 {
    		t.Errorf("Blow(feast) = %d, want 1 (only the \"*\" listener)", n)
    	}
    	expect(t, "Blow(feast)", log, "all:feast")

    	if got := [3]int{h.Listeners("raid"), h.Listeners("dragon"), h.Listeners("feast")}; got != [3]int{3, 2, 1} {
    		t.Errorf("Listeners(raid, dragon, feast) = %v, want [3 2 1]", got)
    	}

    	var empty Horn
    	if n := empty.Blow(ev("raid")); n != 0 || empty.Listeners("raid") != 0 {
    		t.Errorf("a Horn with no listeners: Blow = %d, Listeners = %d, want 0 and 0", n, empty.Listeners("raid"))
    	}
    }

    func TestUnsubscribe(t *testing.T) {
    	var log []string
    	var h Horn
    	w := &watcher{"w", &log}
    	first := h.Subscribe("raid", w)
    	h.Subscribe("raid", w) // the same listener twice: two subscriptions
    	other := h.Subscribe("raid", &watcher{"o", &log})

    	h.Blow(ev("raid"))
    	expect(t, "same listener subscribed twice", log, "w:raid", "w:raid", "o:raid")

    	log = nil
    	first()
    	h.Blow(ev("raid"))
    	expect(t, "after cancelling one of w's two subscriptions", log, "w:raid", "o:raid")

    	log = nil
    	first() // again: must not remove anything else
    	other()
    	if n := h.Blow(ev("raid")); n != 1 {
    		t.Errorf("Blow after unsubscribing = %d, want 1", n)
    	}
    	expect(t, "after calling the first unsubscribe twice and cancelling o", log, "w:raid")
    }

    func TestUnsubscribeFuncs(t *testing.T) {
    	// Function values can't be compared, so the Horn must tell
    	// subscriptions apart some other way.
    	calls := 0
    	f := ListenerFunc(func(Event) { calls++ })
    	var h Horn
    	cancelA := h.Subscribe("raid", f)
    	h.Subscribe("raid", f)
    	cancelA()
    	h.Blow(ev("raid"))
    	if calls != 1 {
    		t.Errorf("same ListenerFunc subscribed twice, one cancelled: it ran %d times, want 1", calls)
    	}
    }

    func TestChangesDuringBlow(t *testing.T) {
    	var log []string
    	var h Horn
    	var cancelB func()
    	h.Subscribe("raid", ListenerFunc(func(e Event) {
    		log = append(log, "a:"+e.Kind)
    		cancelB()                                   // B was already due to hear this blow
    		h.Subscribe("raid", &watcher{"late", &log}) // not until the next blow
    	}))
    	cancelB = h.Subscribe("raid", &watcher{"b", &log})

    	if n := h.Blow(ev("raid")); n != 2 {
    		t.Errorf("first Blow = %d, want 2", n)
    	}
    	expect(t, "first blow (changes take effect next time)", log, "a:raid", "b:raid")

    	log = nil
    	h.Blow(ev("raid"))
    	expect(t, "second blow", log, "a:raid", "late:raid")
    }

    func TestNestedBlow(t *testing.T) {
    	var log []string
    	var h Horn
    	h.Subscribe("dragon-slain", ListenerFunc(func(e Event) {
    		log = append(log, "herald:"+e.Kind)
    		if n := h.Blow(Event{Kind: "loot", Amount: e.Amount * 10}); n != 2 {
    			t.Errorf("nested Blow(loot) = %d, want 2", n)
    		}
    	}))
    	h.Subscribe("*", &watcher{"scribe", &log})
    	h.Subscribe("loot", &watcher{"quartermaster", &log})

    	if n := h.Blow(ev("dragon-slain")); n != 2 {
    		t.Errorf("Blow(dragon-slain) = %d, want 2: count only this blow's listeners, not the nested one's", n)
    	}
    	expect(t, "a listener blowing the horn itself", log,
    		"herald:dragon-slain", "scribe:loot", "quartermaster:loot", "scribe:dragon-slain")
    }

    func TestSubscribeOnce(t *testing.T) {
    	var log []string
    	var h Horn
    	h.SubscribeOnce("raid", &watcher{"once", &log})
    	h.Subscribe("raid", &watcher{"always", &log})
    	h.Blow(ev("feast"))
    	h.Blow(ev("raid"))
    	h.Blow(ev("raid"))
    	expect(t, "SubscribeOnce, then feast, raid, raid", log, "once:raid", "always:raid", "always:raid")
    	if n := h.Listeners("raid"); n != 1 {
    		t.Errorf("after a once-listener fired, Listeners(raid) = %d, want 1", n)
    	}

    	log = nil
    	cancel := h.SubscribeOnce("*", &watcher{"never", &log})
    	cancel()
    	h.Blow(ev("raid"))
    	expect(t, "SubscribeOnce cancelled before any event", log, "always:raid")

    	// A once-listener that blows the horn itself still hears only one event.
    	log = nil
    	var h2 Horn
    	h2.SubscribeOnce("*", ListenerFunc(func(e Event) {
    		log = append(log, "echo:"+e.Kind)
    		h2.Blow(ev("echo"))
    	}))
    	h2.Blow(ev("shout"))
    	h2.Blow(ev("shout"))
    	expect(t, "a once-listener that blows the horn", log, "echo:shout")
    }

    func TestHornsAreIndependent(t *testing.T) {
    	var log []string
    	var a, b Horn
    	a.Subscribe("*", &watcher{"a", &log})
    	if n := b.Blow(ev("raid")); n != 0 || len(log) != 0 {
    		t.Errorf("blowing horn b notified %d listeners of horn a", n)
    	}
    }
---

When a dragon is sighted, the watchtower blows the **war horn** and everybody
who cares reacts: the scribe writes it down, the captain rings the bells, the
quartermaster counts the loot afterwards. Build the horn as an observer: it
knows nothing about its listeners except that they satisfy a one-method
interface.

```go
type Listener interface {
	Notify(e Event)
}
```

1. `ListenerFunc` is a function type. Give it a `Notify` method so that any
   `func(Event)` can be used as a `Listener` via `ListenerFunc(f)`.
2. `(*Horn).Subscribe(kind, l)` registers `l` for events of that `Kind`, or for
   **every** event if `kind` is `"*"`. It returns an `unsubscribe` function that
   cancels **this** subscription only. Calling it again does nothing. The same
   listener may be subscribed several times, and hears an event once per matching
   subscription.
3. `(*Horn).Blow(e)` notifies every matching subscription in the order they were
   made (exact-kind and `"*"` subscriptions interleaved) and returns how many it
   notified.
4. Listeners may subscribe, unsubscribe or even `Blow` the horn from inside
   `Notify`. Subscriptions added or removed during a `Blow` take effect from the
   **next** `Blow`, and a nested `Blow` delivers its own event right away. Each
   `Blow` counts only its own notifications.
5. `(*Horn).SubscribeOnce(kind, l)` is like `Subscribe`, but `l` hears at most
   **one** event and is then removed.
6. `(*Horn).Listeners(kind)` returns how many subscriptions would hear an event
   of that kind.

The zero `Horn` must be ready to use.

## Example

```go
var horn Horn
scribe := &Scribe{}
horn.Subscribe("*", scribe)
stop := horn.Subscribe("dragon-sighted", ListenerFunc(func(e Event) {
	fmt.Println("To arms! A dragon over", e.Source)
}))

horn.Blow(Event{"dragon-sighted", "Eastwatch", 1}) // prints "To arms! ...", returns 2
stop()
horn.Blow(Event{"dragon-sighted", "Westwatch", 2}) // returns 1: only the scribe
horn.Listeners("feast")                             // 1: the scribe hears everything
```

## Constraints

- The tests subscribe their own `Listener` types and closures, and some of them
  change the horn from inside `Notify`.
- No goroutines are involved: everything happens on one goroutine, so you don't
  need a mutex.
