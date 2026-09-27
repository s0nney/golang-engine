---
title: Rate-Limited Dispatcher
difficulty: hard
after: concurrency-patterns
hints:
  - 'Structure it as one loop that, for each delivery, waits in turn for: the next order, the rate limit (a timer until `lastStart + interval`), and a free slot (a buffered channel of size `maxInFlight` used as a semaphore). Every one of those waits is a `select` that also watches `ctx.Done()`. Waiting for the order first means you notice straight away when `orders` is closed.'
  - 'Derive `ctx, cancel := context.WithCancelCause(ctx)` and hand that ctx to every delivery. The first delivery to fail calls `cancel(err)`, which stops the loop and tells the others to give up. Guard "first" with a mutex (or `sync.Once`), and ignore errors that happen after ctx is already done: those are just deliveries reacting to the cancellation.'
  - 'When the loop ends, `wg.Wait()` for the deliveries still running, then decide what to return: the first delivery error if there was one, else `context.Cause(ctx)` if the loop stopped because ctx was done, else nil. `select` picks randomly when several cases are ready, so check `ctx.Err()` once more right before starting a delivery.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"fmt"
    	"time"
    )

    // dispatchAll delivers every order from orders, in order, starting each
    // delivery in its own goroutine. Starts are at least interval apart, and
    // at most maxInFlight deliveries run at once. The first delivery error
    // cancels the rest and is returned; if ctx is done, dispatchAll stops and
    // returns context.Cause(ctx). It returns only once every delivery it
    // started has finished.
    func dispatchAll(ctx context.Context, orders <-chan int, interval time.Duration, maxInFlight int, deliver func(context.Context, int) error) error {
    	for id := range orders {
    		if err := deliver(ctx, id); err != nil {
    			return err
    		}
    	}
    	return nil
    }

    func main() {
    	orders := make(chan int, 5)
    	for id := range 5 {
    		orders <- 100 + id
    	}
    	close(orders)

    	start := time.Now()
    	since := func() time.Duration { return time.Since(start).Round(10 * time.Millisecond) }
    	err := dispatchAll(context.Background(), orders, 10*time.Millisecond, 2, func(ctx context.Context, id int) error {
    		fmt.Printf("%5v: start order %d\n", since(), id)
    		time.Sleep(30 * time.Millisecond)
    		return nil
    	})
    	fmt.Printf("%5v: done, err = %v\n", since(), err)
    	// want starts at 0s, 10ms, 30ms, 40ms, 60ms, then done at 90ms
    }
  solution: |
    package main

    import (
    	"context"
    	"fmt"
    	"sync"
    	"time"
    )

    // dispatchAll delivers every order from orders, in order, starting each
    // delivery in its own goroutine. Starts are at least interval apart, and
    // at most maxInFlight deliveries run at once. The first delivery error
    // cancels the rest and is returned; if ctx is done, dispatchAll stops and
    // returns context.Cause(ctx). It returns only once every delivery it
    // started has finished.
    func dispatchAll(ctx context.Context, orders <-chan int, interval time.Duration, maxInFlight int, deliver func(context.Context, int) error) error {
    	ctx, cancel := context.WithCancelCause(ctx)
    	defer cancel(nil)

    	var (
    		wg       sync.WaitGroup
    		mu       sync.Mutex
    		firstErr error
    		slots    = make(chan struct{}, maxInFlight)
    		next     time.Time // earliest time the next delivery may start
    	)
    	stoppedByCtx := func() bool {
    		for {
    			// 1. The next order.
    			var id int
    			select {
    			case v, ok := <-orders:
    				if !ok {
    					return false
    				}
    				id = v
    			case <-ctx.Done():
    				return true
    			}
    			// 2. The rate limit.
    			if wait := time.Until(next); wait > 0 {
    				timer := time.NewTimer(wait)
    				select {
    				case <-timer.C:
    				case <-ctx.Done():
    					timer.Stop()
    					return true
    				}
    			}
    			// 3. A free slot.
    			select {
    			case slots <- struct{}{}:
    			case <-ctx.Done():
    				return true
    			}
    			if ctx.Err() != nil { // select picks randomly if both were ready
    				<-slots
    				return true
    			}

    			next = time.Now().Add(interval)
    			wg.Go(func() {
    				defer func() { <-slots }()
    				if err := deliver(ctx, id); err != nil {
    					mu.Lock()
    					defer mu.Unlock()
    					if firstErr == nil && ctx.Err() == nil {
    						firstErr = err
    						cancel(err)
    					}
    				}
    			})
    		}
    	}()

    	wg.Wait()
    	if firstErr != nil {
    		return firstErr
    	}
    	if stoppedByCtx {
    		return context.Cause(ctx)
    	}
    	return nil
    }

    func main() {
    	orders := make(chan int, 5)
    	for id := range 5 {
    		orders <- 100 + id
    	}
    	close(orders)

    	start := time.Now()
    	since := func() time.Duration { return time.Since(start).Round(10 * time.Millisecond) }
    	err := dispatchAll(context.Background(), orders, 10*time.Millisecond, 2, func(ctx context.Context, id int) error {
    		fmt.Printf("%5v: start order %d\n", since(), id)
    		time.Sleep(30 * time.Millisecond)
    		return nil
    	})
    	fmt.Printf("%5v: done, err = %v\n", since(), err)
    }
  tests: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"slices"
    	"sync"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // courierAPI is a fake deliver func. Each delivery takes took[id] (or
    // def), then returns fail[id]. If its ctx is cancelled first, it spends
    // windDown cleaning up and returns ctx.Err().
    type courierAPI struct {
    	start    time.Time
    	def      time.Duration
    	took     map[int]time.Duration
    	fail     map[int]error
    	windDown time.Duration

    	mu        sync.Mutex
    	starts    []string // "1s #7"
    	running   int
    	peak      int
    	cancelled []string // "1.5s #3: <cause>"
    }

    func newAPI(def time.Duration) *courierAPI {
    	return &courierAPI{start: time.Now(), def: def, took: map[int]time.Duration{}, fail: map[int]error{}}
    }

    func (a *courierAPI) deliver(ctx context.Context, id int) error {
    	a.mu.Lock()
    	a.starts = append(a.starts, fmt.Sprint(time.Since(a.start), " #", id))
    	a.running++
    	a.peak = max(a.peak, a.running)
    	d, ok := a.took[id]
    	if !ok {
    		d = a.def
    	}
    	a.mu.Unlock()
    	defer func() {
    		a.mu.Lock()
    		defer a.mu.Unlock()
    		a.running--
    	}()

    	select {
    	case <-time.After(d):
    		return a.fail[id]
    	case <-ctx.Done():
    		a.mu.Lock()
    		a.cancelled = append(a.cancelled, fmt.Sprint(time.Since(a.start), " #", id, ": ", context.Cause(ctx)))
    		a.mu.Unlock()
    		time.Sleep(a.windDown)
    		return ctx.Err() // context.Canceled, like most real APIs
    	}
    }

    func (a *courierAPI) snapshot() (starts []string, running, peak int, cancelled []string) {
    	a.mu.Lock()
    	defer a.mu.Unlock()
    	return slices.Clone(a.starts), a.running, a.peak, slices.Clone(a.cancelled)
    }

    func queue(ids ...int) chan int {
    	ch := make(chan int, len(ids))
    	for _, id := range ids {
    		ch <- id
    	}
    	close(ch)
    	return ch
    }

    // run calls dispatchAll and fails the test if it takes over a day.
    func run(t *testing.T, ctx context.Context, orders <-chan int, interval time.Duration, maxInFlight int, a *courierAPI) (error, time.Duration) {
    	t.Helper()
    	done := make(chan error, 1)
    	go func() { done <- dispatchAll(ctx, orders, interval, maxInFlight, a.deliver) }()
    	select {
    	case err := <-done:
    		return err, time.Since(a.start)
    	case <-time.After(24 * time.Hour):
    		t.Fatalf("dispatchAll hadn't returned after a day of fake time")
    		return nil, 0
    	}
    }

    func TestRateAndInFlightLimits(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		a := newAPI(3 * time.Second)
    		err, took := run(t, t.Context(), queue(1, 2, 3, 4, 5), time.Second, 2, a)
    		starts, running, peak, _ := a.snapshot()

    		want := []string{"0s #1", "1s #2", "3s #3", "4s #4", "6s #5"}
    		if !slices.Equal(starts, want) {
    			t.Errorf("5 orders, interval 1s, max 2 in flight, 3s deliveries:\nstarts %v\nwant   %v\n(each start waits for a free slot AND for 1s after the previous start)", starts, want)
    		}
    		if peak > 2 {
    			t.Errorf("%d deliveries ran at once, want at most 2", peak)
    		}
    		if err != nil {
    			t.Errorf("dispatchAll = %v, want nil", err)
    		}
    		if running != 0 || took != 9*time.Second {
    			t.Errorf("dispatchAll returned at %v with %d deliveries still running, want 9s with none running", took, running)
    		}
    	})
    }

    func TestRateLimitOnly(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		a := newAPI(100 * time.Millisecond)
    		err, took := run(t, t.Context(), queue(1, 2, 3, 4, 5, 6), 500*time.Millisecond, 10, a)
    		starts, _, _, _ := a.snapshot()
    		want := []string{"0s #1", "500ms #2", "1s #3", "1.5s #4", "2s #5", "2.5s #6"}
    		if !slices.Equal(starts, want) {
    			t.Errorf("6 orders, interval 500ms, max 10 in flight:\nstarts %v\nwant   %v", starts, want)
    		}
    		if err != nil || took != 2600*time.Millisecond {
    			t.Errorf("dispatchAll = %v at %v, want nil at 2.6s", err, took)
    		}
    	})
    }

    func TestLateOrdersStartOnArrival(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		orders := make(chan int)
    		go func() {
    			defer close(orders)
    			orders <- 1
    			time.Sleep(5 * time.Second)
    			orders <- 2
    			orders <- 3
    		}()
    		a := newAPI(time.Second)
    		err, took := run(t, t.Context(), orders, 2*time.Second, 5, a)
    		starts, _, _, _ := a.snapshot()
    		want := []string{"0s #1", "5s #2", "7s #3"}
    		if !slices.Equal(starts, want) {
    			t.Errorf("orders arriving at 0s, 5s, 5s with interval 2s:\nstarts %v\nwant   %v", starts, want)
    		}
    		if err != nil || took != 8*time.Second {
    			t.Errorf("dispatchAll = %v at %v, want nil at 8s", err, took)
    		}
    	})
    }

    func TestNoOrders(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		a := newAPI(time.Second)
    		if err, took := run(t, t.Context(), queue(), time.Second, 3, a); err != nil || took != 0 {
    			t.Errorf("no orders: dispatchAll = %v after %v, want nil after 0s", err, took)
    		}
    	})
    }

    func TestFirstErrorCancelsTheRest(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		errNoCourier := errors.New("no courier nearby")
    		a := newAPI(10 * time.Second)
    		a.took[2] = 500 * time.Millisecond
    		a.fail[2] = errNoCourier
    		a.windDown = 200 * time.Millisecond

    		err, took := run(t, t.Context(), queue(1, 2, 3, 4, 5, 6, 7, 8), time.Second, 3, a)
    		starts, running, _, cancelled := a.snapshot()

    		if !errors.Is(err, errNoCourier) {
    			t.Errorf("order #2 failed with %q: dispatchAll = %v, want that error", errNoCourier, err)
    		}
    		if want := []string{"0s #1", "1s #2"}; !slices.Equal(starts, want) {
    			t.Errorf("order #2 failed at 1.5s: starts %v, want %v (start nothing after the first failure)", starts, want)
    		}
    		if want := []string{"1.5s #1: no courier nearby"}; !slices.Equal(cancelled, want) {
    			t.Errorf("order #2 failed at 1.5s: cancelled deliveries %v, want %v (cancel their ctx with the error as its cause)", cancelled, want)
    		}
    		if running != 0 || took != 1700*time.Millisecond {
    			t.Errorf("dispatchAll returned at %v with %d deliveries running, want 1.7s (after #1's 200ms wind-down) with none running", took, running)
    		}
    	})
    }

    func TestParentCancelled(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		errShutdown := errors.New("dispatcher shutting down")
    		ctx, cancel := context.WithCancelCause(t.Context())
    		time.AfterFunc(2500*time.Millisecond, func() { cancel(errShutdown) })
    		orders := make(chan int, 10) // never closed
    		for id := range 10 {
    			orders <- id + 1
    		}
    		a := newAPI(time.Hour)
    		a.windDown = time.Second

    		err, took := run(t, ctx, orders, time.Second, 5, a)
    		starts, running, _, cancelled := a.snapshot()

    		if !errors.Is(err, errShutdown) {
    			t.Errorf("ctx cancelled with cause %q: dispatchAll = %v, want the cause (a delivery returning context.Canceled because it was cancelled isn't a delivery failure)", errShutdown, err)
    		}
    		if want := []string{"0s #1", "1s #2", "2s #3"}; !slices.Equal(starts, want) {
    			t.Errorf("ctx cancelled at 2.5s: starts %v, want %v", starts, want)
    		}
    		if len(cancelled) != 3 {
    			t.Errorf("ctx cancelled at 2.5s: deliveries that saw the cancellation: %v, want all 3", cancelled)
    		}
    		if running != 0 || took != 3500*time.Millisecond {
    			t.Errorf("dispatchAll returned at %v with %d deliveries running, want 3.5s (after the 1s wind-down) with none running", took, running)
    		}
    	})
    }

    func TestAlreadyCancelled(t *testing.T) {
    	errClosed := errors.New("city closed for the storm")
    	for range 100 { // select picks randomly among ready cases, so try many times
    		synctest.Test(t, func(t *testing.T) {
    			ctx, cancel := context.WithCancelCause(t.Context())
    			cancel(errClosed)
    			a := newAPI(time.Second)
    			err, _ := run(t, ctx, queue(1, 2, 3, 4, 5), time.Millisecond, 5, a)
    			if starts, _, _, _ := a.snapshot(); len(starts) != 0 {
    				t.Fatalf("ctx already cancelled: deliveries started %v, want none (check ctx.Err() right before starting one)", starts)
    			}
    			if !errors.Is(err, errClosed) {
    				t.Fatalf("ctx already cancelled with cause %q: dispatchAll = %v, want the cause", errClosed, err)
    			}
    		})
    	}
    }

    func TestCancelledWhileWaitingForOrders(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
    		defer cancel()
    		a := newAPI(time.Second)
    		err, took := run(t, ctx, make(chan int), time.Second, 1, a)
    		if !errors.Is(err, context.DeadlineExceeded) || took != time.Minute {
    			t.Errorf("no orders arrive, 1m timeout: dispatchAll = %v at %v, want context.DeadlineExceeded at 1m", err, took)
    		}
    	})
    }
---

Dispatchly hands deliveries to a partner courier company whose API has two
limits: you may **start** at most one delivery per `interval`, and at most
`maxInFlight` may be running at once. Break either and your account is
suspended.

Implement `dispatchAll(ctx, orders, interval, maxInFlight, deliver)`:

- Read order IDs from `orders` and call `deliver(ctx, id)` for each one, in
  order, each in its own goroutine.
- **Rate:** two starts are always at least `interval` apart. The first
  delivery starts straight away.
- **In flight:** at most `maxInFlight` deliveries run at once. A delivery
  that has to wait for a slot starts as soon as one frees up (and the rate
  allows), and the next start is measured from when it actually started.
- An order that arrives late starts as soon as it arrives, if the limits
  allow.
- **First error cancels:** when a delivery returns an error, start nothing
  more, cancel the context the other deliveries got (with that error as its
  cause), and return that error.
- If `ctx` is done, start nothing more and return `context.Cause(ctx)`.
  `orders` might never be closed, so don't wait on it forever.
- Return `nil` once `orders` is closed and every delivery succeeded. In every
  case, return only **after** every delivery you started has finished.

## Example

5 orders queued, `interval` 1s, `maxInFlight` 2, every delivery takes 3s:

```
t=0s  start #1
t=1s  start #2        (1s after #1)
t=3s  start #3        (#1 finished; the rate allowed it since 2s)
t=4s  start #4        (#2 finished, 1s after #3)
t=6s  start #5        (#3 finished; the rate allowed it since 5s)
t=9s  return nil      (#5 finished)
```

## Constraints

- `interval` > 0 and `maxInFlight` ≥ 1.
- Errors a delivery returns *because* it was cancelled are not "first"
  errors: return the original failure (or the context's cause).
- The tests run in a `synctest` bubble and check every start time exactly,
  the peak number of deliveries in flight, which deliveries saw the
  cancellation and with what cause, and that nothing is left running.
