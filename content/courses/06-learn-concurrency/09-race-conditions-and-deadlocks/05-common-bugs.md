---
title: Common Concurrency Bugs and Their Fixes
quiz:
  - question: |
      What's the bug?

      ```go
      func (c *Cache) Get(key string) Menu {
      	c.mu.RLock()
      	m, ok := c.menus[key]
      	c.mu.RUnlock()
      	if !ok {
      		m = fetchMenu(key) // slow
      		c.mu.Lock()
      		c.menus[key] = m
      		c.mu.Unlock()
      	}
      	return m
      }
      ```
    options:
      - text: A data race on `c.menus`
      - text: A deadlock between `RLock` and `Lock`
      - text: Nothing serious; concurrent misses may fetch the same menu twice, which wastes work but stores a valid menu either way
        correct: true
      - text: '`fetchMenu` must be called while holding the lock'
    explanation: |
      Every map access is locked, so there's no data race, and the locks
      aren't nested, so no deadlock. Two goroutines can both miss and both
      fetch. That's a race condition, but a benign one here, since both
      store an equally good menu. Whether a race condition is a bug depends
      on what the operation promises.
  - question: 'What''s wrong with calling `wg.Add(1)` *inside* the goroutine, as in `go func() { wg.Add(1); defer wg.Done(); work() }()`?'
    options:
      - text: Nothing, it's equivalent to calling it before `go`
      - text: '`wg.Wait()` might run before the goroutine gets to `Add`, see a zero counter, and return too early'
        correct: true
      - text: It panics with "negative WaitGroup counter"
      - text: It leaks the goroutine
    explanation: |
      `Add` must happen before `Wait` could possibly observe the counter.
      Inside the goroutine there's no such guarantee. `wg.Go` avoids the
      problem entirely by doing the `Add` for you, before starting the
      goroutine.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"sync"
    	"time"
    )

    // Dispatcher assigns orders to couriers. A courier can only carry one
    // order at a time.
    type Dispatcher struct {
    	mu       sync.Mutex
    	assigned map[string]string // courier -> order; guarded by mu

    	// notify pushes the new job to the courier's phone. It's slow.
    	notify func(courier, order string)
    }

    func NewDispatcher(notify func(courier, order string)) *Dispatcher {
    	return &Dispatcher{assigned: make(map[string]string), notify: notify}
    }

    // Assign gives order to courier if the courier is free, notifies them,
    // and reports whether it did.
    func (d *Dispatcher) Assign(courier, order string) bool {
    	d.mu.Lock()
    	_, busy := d.assigned[courier]
    	d.mu.Unlock()
    	if busy {
    		return false
    	}

    	d.notify(courier, order)

    	d.mu.Lock()
    	d.assigned[courier] = order
    	d.mu.Unlock()
    	return true
    }

    func main() {
    	d := NewDispatcher(func(courier, order string) {
    		time.Sleep(10 * time.Millisecond) // talk to the phone
    		fmt.Printf("%s, please pick up %s\n", courier, order)
    	})

    	var wg sync.WaitGroup
    	for _, order := range []string{"A1", "B2", "C3"} {
    		wg.Go(func() { d.Assign("ana", order) })
    	}
    	wg.Wait()
    	fmt.Println("ana is carrying:", d.assigned["ana"])
    }
  solution: |
    package main

    import (
    	"fmt"
    	"sync"
    	"time"
    )

    // Dispatcher assigns orders to couriers. A courier can only carry one
    // order at a time.
    type Dispatcher struct {
    	mu       sync.Mutex
    	assigned map[string]string // courier -> order; guarded by mu

    	// notify pushes the new job to the courier's phone. It's slow.
    	notify func(courier, order string)
    }

    func NewDispatcher(notify func(courier, order string)) *Dispatcher {
    	return &Dispatcher{assigned: make(map[string]string), notify: notify}
    }

    // Assign gives order to courier if the courier is free, notifies them,
    // and reports whether it did.
    func (d *Dispatcher) Assign(courier, order string) bool {
    	// Check and claim in one critical section, so no other Assign can
    	// slip in between.
    	d.mu.Lock()
    	if _, busy := d.assigned[courier]; busy {
    		d.mu.Unlock()
    		return false
    	}
    	d.assigned[courier] = order
    	d.mu.Unlock()

    	// The slow call happens without holding the lock.
    	d.notify(courier, order)
    	return true
    }

    func main() {
    	d := NewDispatcher(func(courier, order string) {
    		time.Sleep(10 * time.Millisecond) // talk to the phone
    		fmt.Printf("%s, please pick up %s\n", courier, order)
    	})

    	var wg sync.WaitGroup
    	for _, order := range []string{"A1", "B2", "C3"} {
    		wg.Go(func() { d.Assign("ana", order) })
    	}
    	wg.Wait()
    	fmt.Println("ana is carrying:", d.assigned["ana"])
    }
  tests: |
    package main

    import (
    	"slices"
    	"sync"
    	"testing"
    	"time"
    )

    // gatedNotify records notifications and blocks each one until release is closed.
    type gatedNotify struct {
    	started chan string
    	release chan struct{}
    }

    func newGate() *gatedNotify {
    	return &gatedNotify{started: make(chan string, 10), release: make(chan struct{})}
    }

    func (g *gatedNotify) notify(courier, order string) {
    	g.started <- courier + ":" + order
    	<-g.release
    }

    // startedWithin collects notifications that start within d.
    func (g *gatedNotify) startedWithin(d time.Duration, max int) []string {
    	var got []string
    	timeout := time.After(d)
    	for len(got) < max {
    		select {
    		case s := <-g.started:
    			got = append(got, s)
    		case <-timeout:
    			return got
    		}
    	}
    	return got
    }

    func TestSameCourierOnlyOnce(t *testing.T) {
    	g := newGate()
    	d := NewDispatcher(g.notify)

    	results := make([]bool, 2)
    	var wg sync.WaitGroup
    	for i, order := range []string{"A1", "B2"} {
    		wg.Go(func() { results[i] = d.Assign("ana", order) })
    	}

    	notified := g.startedWithin(100*time.Millisecond, 2)
    	close(g.release)
    	wg.Wait()

    	if len(notified) != 1 {
    		t.Errorf("two concurrent Assign calls for the same courier sent %d notifications %v, want 1: ana can only carry one order", len(notified), notified)
    	}
    	if results[0] == results[1] {
    		t.Errorf("Assign returned %v and %v, want exactly one true", results[0], results[1])
    	}
    	winner := "A1"
    	if results[1] {
    		winner = "B2"
    	}
    	if got := d.assigned["ana"]; got != winner {
    		t.Errorf("ana is assigned %q, want %q (the order whose Assign returned true)", got, winner)
    	}
    }

    func TestDifferentCouriersInParallel(t *testing.T) {
    	g := newGate()
    	d := NewDispatcher(g.notify)

    	var wg sync.WaitGroup
    	wg.Go(func() { d.Assign("ana", "A1") })
    	wg.Go(func() { d.Assign("ben", "B2") })

    	notified := g.startedWithin(time.Second, 2)
    	close(g.release)
    	wg.Wait()

    	slices.Sort(notified)
    	if want := []string{"ana:A1", "ben:B2"}; !slices.Equal(notified, want) {
    		t.Errorf("notifications that started while the first was still in progress: %v, want %v (don't hold the lock during the slow notify call)", notified, want)
    	}
    }

    func TestSequential(t *testing.T) {
    	var calls []string
    	d := NewDispatcher(func(c, o string) { calls = append(calls, c+":"+o) })

    	if !d.Assign("cy", "C3") {
    		t.Errorf("Assign(cy, C3) on a free courier = false, want true")
    	}
    	if d.Assign("cy", "D4") {
    		t.Errorf("Assign(cy, D4) on a busy courier = true, want false")
    	}
    	if !slices.Equal(calls, []string{"cy:C3"}) {
    		t.Errorf("notify calls = %v, want [cy:C3]", calls)
    	}
    }
---

Most concurrency bugs in real Go code are variations on a small number of themes. You've met all of them in this course. Here they are in one place, as a checklist to run through when something is flaky.

## 1. Unsynchronized shared state

**Symptom:** lost updates, impossible values, `fatal error: concurrent map writes`.
**Detect:** `go test -race`.
**Fix:** guard with a mutex, use an atomic, or give the state a single owner goroutine.

## 2. Check-then-act across critical sections

**Symptom:** two goroutines both "win": a courier with two orders, a wallet overdrawn, a job run twice.
**Detect:** code review. Look for `Unlock` followed by `Lock` of the same mutex in one function. The race detector won't help.
**Fix:** do the check and the act inside **one** critical section, or use `CompareAndSwap`.

## 3. Holding a lock during slow work

**Symptom:** everything gets slow under load, because every goroutine queues behind one slow network call. Can escalate to a deadlock if the slow work needs the same lock or a channel whose receiver does.
**Fix:** hold locks only while touching shared memory. Copy what you need, unlock, then call out.

Bugs 2 and 3 pull in opposite directions: "lock *more* around the check and act" versus "lock *less* around slow work". The trick is to claim, record or reserve under the lock, then do the slow part outside it.

## 4. Goroutine leaks

**Symptom:** memory and `runtime.NumGoroutine()` creep up forever.
**Detect:** goroutine counts, goroutine dumps, the `goroutineleak` profile, `synctest`.
**Fix:** every goroutine needs a guaranteed exit: a buffered result channel, a closed input, or `ctx.Done()` in every `select`.

## 5. WaitGroup misuse

**Symptom:** `Wait` returns before work is done, or panics with a negative counter.
**Fix:** use `wg.Go`, which does `Add` before starting the goroutine and `Done` when it returns. If you use `Add`/`Done`, call `Add` *before* `go`, never inside the goroutine.

## 6. Closing channels wrongly

**Symptom:** `panic: send on closed channel`, `panic: close of closed channel`, or a `range` that never ends.
**Fix:** only the sender closes. With several senders, a separate closer waits on a WaitGroup, then closes. Receivers that want to stop early cancel a context instead.

## 7. Deadlocks

**Symptom:** `all goroutines are asleep`, or requests that simply never finish.
**Fix:** a global lock order, never locking a mutex twice in one goroutine, closer goroutines instead of waiting before receiving, and timeouts on anything that waits.

## 8. Relying on timing

**Symptom:** tests that pass locally and fail in CI. `time.Sleep(10 * time.Millisecond) // wait for the goroutine to start`.
**Fix:** never use sleeps for synchronization. Wait on the actual event: a channel, a WaitGroup, or in tests `synctest.Wait`.

## Your turn

This is bug 2 in the wild. Dispatchly's `Assign` checks whether a courier is free, sends a push notification to their phone (slow), then records the assignment. Every map access is locked, so there's no data race, but when three dispatchers assign orders to ana at the same moment, she gets all three. Press **Run** to see her phone buzz three times.

Fix `Assign` so that:

- a courier can never be given two orders, however many goroutines call `Assign` at once (exactly one call returns `true`, and `assigned` holds that call's order),
- the slow `notify` call is **not** made while holding the lock, so assigning *different* couriers still happens in parallel.
