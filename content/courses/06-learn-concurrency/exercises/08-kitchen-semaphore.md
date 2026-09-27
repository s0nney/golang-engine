---
title: Kitchen Semaphore
difficulty: medium
after: concurrency-patterns
hints:
  - 'A buffered channel with capacity `n` is a ready-made semaphore: sending a token takes a slot (and blocks when all `n` are taken), receiving one frees a slot.'
  - '`Acquire` selects between sending a token and `<-ctx.Done()`. `TryAcquire` is the same send with a `default` case. `Release` receives a token, with a `default` case that panics.'
  - 'When a slot is free **and** ctx is already cancelled, both `select` cases are ready and Go picks one at random, so a cancelled caller would sometimes get a slot anyway. Check `ctx.Err()` first.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"fmt"
    	"sync"
    	"time"
    )

    // Semaphore limits how many holders there can be at once.
    type Semaphore struct {
    }

    // NewSemaphore returns a semaphore with n slots (n >= 1).
    func NewSemaphore(n int) *Semaphore {
    	return &Semaphore{}
    }

    // Acquire takes a slot, waiting until one is free. If ctx is done first
    // (or already), it takes no slot and returns context.Cause(ctx).
    func (s *Semaphore) Acquire(ctx context.Context) error {
    	return nil
    }

    // TryAcquire takes a slot if one is free right now and reports whether
    // it did. It never waits.
    func (s *Semaphore) TryAcquire() bool {
    	return true
    }

    // Release frees a slot. It panics if no slot is held.
    func (s *Semaphore) Release() {
    }

    func main() {
    	grills := NewSemaphore(2)
    	start := time.Now()
    	var wg sync.WaitGroup
    	for _, order := range []string{"burger", "steak", "kebab", "corn"} {
    		wg.Go(func() {
    			if err := grills.Acquire(context.Background()); err != nil {
    				fmt.Println(order, "error:", err)
    				return
    			}
    			defer grills.Release()
    			time.Sleep(50 * time.Millisecond)
    			fmt.Printf("%v: %s grilled\n", time.Since(start).Round(50*time.Millisecond), order)
    		})
    	}
    	wg.Wait()
    	// want: two orders grilled at 50ms, two at 100ms
    }
  solution: |
    package main

    import (
    	"context"
    	"fmt"
    	"sync"
    	"time"
    )

    // Semaphore limits how many holders there can be at once.
    type Semaphore struct {
    	slots chan struct{} // one token per slot in use
    }

    // NewSemaphore returns a semaphore with n slots (n >= 1).
    func NewSemaphore(n int) *Semaphore {
    	return &Semaphore{slots: make(chan struct{}, n)}
    }

    // Acquire takes a slot, waiting until one is free. If ctx is done first
    // (or already), it takes no slot and returns context.Cause(ctx).
    func (s *Semaphore) Acquire(ctx context.Context) error {
    	if ctx.Err() != nil {
    		return context.Cause(ctx)
    	}
    	select {
    	case s.slots <- struct{}{}:
    		return nil
    	case <-ctx.Done():
    		return context.Cause(ctx)
    	}
    }

    // TryAcquire takes a slot if one is free right now and reports whether
    // it did. It never waits.
    func (s *Semaphore) TryAcquire() bool {
    	select {
    	case s.slots <- struct{}{}:
    		return true
    	default:
    		return false
    	}
    }

    // Release frees a slot. It panics if no slot is held.
    func (s *Semaphore) Release() {
    	select {
    	case <-s.slots:
    	default:
    		panic("semaphore: Release without Acquire")
    	}
    }

    func main() {
    	grills := NewSemaphore(2)
    	start := time.Now()
    	var wg sync.WaitGroup
    	for _, order := range []string{"burger", "steak", "kebab", "corn"} {
    		wg.Go(func() {
    			if err := grills.Acquire(context.Background()); err != nil {
    				fmt.Println(order, "error:", err)
    				return
    			}
    			defer grills.Release()
    			time.Sleep(50 * time.Millisecond)
    			fmt.Printf("%v: %s grilled\n", time.Since(start).Round(50*time.Millisecond), order)
    		})
    	}
    	wg.Wait()
    }
  tests: |
    package main

    import (
    	"context"
    	"errors"
    	"sync"
    	"sync/atomic"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    func TestSemaphoreLimitsHolders(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		sem := NewSemaphore(3)
    		var now, peak atomic.Int32
    		start := time.Now()
    		var wg sync.WaitGroup
    		for range 10 {
    			wg.Go(func() {
    				if err := sem.Acquire(t.Context()); err != nil {
    					t.Errorf("Acquire with a live context = %v, want nil", err)
    					return
    				}
    				n := now.Add(1)
    				for {
    					p := peak.Load()
    					if n <= p || peak.CompareAndSwap(p, n) {
    						break
    					}
    				}
    				time.Sleep(time.Second)
    				now.Add(-1)
    				sem.Release()
    			})
    		}
    		wg.Wait()
    		if p := peak.Load(); p != 3 {
    			t.Errorf("10 goroutines sharing NewSemaphore(3): at most %d held a slot at once, want 3", p)
    		}
    		if took := time.Since(start); took != 4*time.Second {
    			t.Errorf("10 one-second jobs through 3 slots took %v, want 4s", took)
    		}
    	})
    }

    func TestAcquireWaitsForRelease(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		sem := NewSemaphore(1)
    		if err := sem.Acquire(t.Context()); err != nil {
    			t.Fatalf("first Acquire = %v, want nil", err)
    		}
    		var got atomic.Bool
    		go func() {
    			if sem.Acquire(t.Context()) == nil {
    				got.Store(true)
    			}
    		}()
    		synctest.Wait()
    		if got.Load() {
    			t.Fatalf("second Acquire on NewSemaphore(1) succeeded while the only slot was held")
    		}
    		sem.Release()
    		synctest.Wait()
    		if !got.Load() {
    			t.Fatalf("waiting Acquire still blocked after Release freed the slot")
    		}
    		sem.Release()
    	})
    }

    func TestAcquireCancelled(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		errClosing := errors.New("kitchen closing")
    		sem := NewSemaphore(1)
    		sem.Acquire(t.Context())

    		ctx, cancel := context.WithCancelCause(t.Context())
    		time.AfterFunc(5*time.Second, func() { cancel(errClosing) })
    		start := time.Now()
    		err := sem.Acquire(ctx)
    		if !errors.Is(err, errClosing) {
    			t.Errorf("Acquire cancelled with cause %q while waiting = %v, want the cause", errClosing, err)
    		}
    		if took := time.Since(start); took != 5*time.Second {
    			t.Errorf("Acquire cancelled after 5s returned after %v, want 5s", took)
    		}

    		// The cancelled Acquire must not have taken the slot.
    		sem.Release()
    		if !sem.TryAcquire() {
    			t.Errorf("after the holder released, TryAcquire = false: did the cancelled Acquire take a slot?")
    		}
    	})
    }

    func TestAcquireAlreadyCancelled(t *testing.T) {
    	errClosed := errors.New("kitchen closed")
    	ctx, cancel := context.WithCancelCause(context.Background())
    	cancel(errClosed)
    	sem := NewSemaphore(1)
    	for i := range 200 {
    		if err := sem.Acquire(ctx); !errors.Is(err, errClosed) {
    			t.Fatalf("call %d: Acquire with an already-cancelled ctx and a free slot = %v, want the cause %q (check ctx before select: select picks randomly among ready cases)", i+1, err, errClosed)
    		}
    	}
    	if !sem.TryAcquire() {
    		t.Errorf("TryAcquire = false after only cancelled Acquires: they must not take slots")
    	}
    }

    func TestTryAcquire(t *testing.T) {
    	sem := NewSemaphore(2)
    	got := []bool{sem.TryAcquire(), sem.TryAcquire(), sem.TryAcquire()}
    	if got[0] != true || got[1] != true || got[2] != false {
    		t.Errorf("three TryAcquire calls on NewSemaphore(2) = %v, want [true true false]", got)
    	}
    	sem.Release()
    	if !sem.TryAcquire() {
    		t.Errorf("TryAcquire after a Release = false, want true")
    	}
    }

    func TestReleaseWithoutAcquirePanics(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		sem := NewSemaphore(2)
    		done := make(chan any, 1)
    		go func() {
    			defer func() { done <- recover() }()
    			sem.Release()
    		}()
    		select {
    		case r := <-done:
    			if r == nil {
    				t.Errorf("Release on a semaphore with no slots held returned normally, want a panic")
    			}
    		case <-time.After(time.Hour):
    			t.Errorf("Release on a semaphore with no slots held blocked, want a panic")
    			sem.TryAcquire() // unblock it so the test can end
    		}
    	})
    }
---

Dispatchly's partner kitchens only have so many grills. Orders are cooked
concurrently, but never more at once than there are grills, and an order
that's cancelled while waiting for a grill must give up its place in the
queue at once.

Implement a `Semaphore` with `n` slots:

- `NewSemaphore(n)` creates it, with all `n` slots free.
- `Acquire(ctx)` takes a slot, waiting until one is free. If `ctx` is done
  before that, it returns `context.Cause(ctx)` without taking a slot. If `ctx`
  is **already** done, it must fail every time, even if a slot is free.
- `TryAcquire()` takes a slot only if one is free right now, and reports
  whether it did.
- `Release()` frees a slot, letting one waiting `Acquire` through. Calling it
  when no slot is held is a bug in the caller: panic.

## Example

```go
grills := NewSemaphore(2)
grills.TryAcquire()          // true
grills.Acquire(ctx)          // nil: second slot
grills.TryAcquire()          // false: both taken
grills.Release()
grills.TryAcquire()          // true
```

## Constraints

- Many goroutines use one semaphore at once. Don't start goroutines inside the
  methods.
- The tests use `synctest`: 10 one-second jobs through 3 slots must take
  exactly 4s with never more than 3 running, and a waiting `Acquire` must
  return at the exact moment its context is cancelled.
