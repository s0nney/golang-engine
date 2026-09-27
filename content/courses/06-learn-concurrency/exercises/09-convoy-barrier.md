---
title: Convoy Barrier
difficulty: medium
after: coordination-patterns
hints:
  - 'Give every round (generation) its own `release` channel. Under the mutex, an arrival bumps the count and either waits on the current round''s channel or, if it''s the `n`th, closes it and starts a fresh round (new channel, count back to 0).'
  - 'Grab the round''s channel while holding the lock, then unlock **before** waiting on it in a `select` with `ctx.Done()`. Never block while holding the mutex.'
  - 'On cancellation, lock again and check whether *your* round is still the current one. If it is, take your arrival back (`count--`) and return the cause. If it isn''t, the round was released while you were being cancelled, so you did make it: return nil.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"fmt"
    	"sync"
    	"time"
    )

    // Barrier releases couriers in groups of n.
    type Barrier struct {
    }

    // NewBarrier returns a barrier for groups of n couriers (n >= 1).
    func NewBarrier(n int) *Barrier {
    	return &Barrier{}
    }

    // Await blocks until n couriers (including this one) have arrived in the
    // current round, then returns nil. The barrier then starts a new round.
    // If ctx is done first, the courier leaves the round (it no longer counts
    // towards n) and Await returns context.Cause(ctx).
    func (b *Barrier) Await(ctx context.Context) error {
    	return nil
    }

    func main() {
    	b := NewBarrier(3)
    	start := time.Now()
    	var wg sync.WaitGroup
    	for i, name := range []string{"ana", "ben", "cy", "dee", "eve", "fox"} {
    		wg.Go(func() {
    			time.Sleep(time.Duration(i*10) * time.Millisecond) // couriers trickle in
    			b.Await(context.Background())
    			fmt.Printf("%v: %s leaves\n", time.Since(start).Round(10*time.Millisecond), name)
    		})
    	}
    	wg.Wait()
    	// want: ana, ben and cy leave at 20ms; dee, eve and fox at 50ms
    }
  solution: |
    package main

    import (
    	"context"
    	"fmt"
    	"sync"
    	"time"
    )

    // Barrier releases couriers in groups of n.
    type Barrier struct {
    	n       int
    	mu      sync.Mutex
    	count   int           // arrivals in the current round
    	release chan struct{} // closed when the current round is complete
    }

    // NewBarrier returns a barrier for groups of n couriers (n >= 1).
    func NewBarrier(n int) *Barrier {
    	return &Barrier{n: n, release: make(chan struct{})}
    }

    // Await blocks until n couriers (including this one) have arrived in the
    // current round, then returns nil. The barrier then starts a new round.
    // If ctx is done first, the courier leaves the round (it no longer counts
    // towards n) and Await returns context.Cause(ctx).
    func (b *Barrier) Await(ctx context.Context) error {
    	if ctx.Err() != nil {
    		return context.Cause(ctx)
    	}
    	b.mu.Lock()
    	b.count++
    	if b.count == b.n {
    		close(b.release) // release this round...
    		b.release = make(chan struct{})
    		b.count = 0 // ...and start the next one
    		b.mu.Unlock()
    		return nil
    	}
    	release := b.release
    	b.mu.Unlock()

    	select {
    	case <-release:
    		return nil
    	case <-ctx.Done():
    		b.mu.Lock()
    		defer b.mu.Unlock()
    		if release != b.release {
    			return nil // our round completed while we were being cancelled
    		}
    		b.count--
    		return context.Cause(ctx)
    	}
    }

    func main() {
    	b := NewBarrier(3)
    	start := time.Now()
    	var wg sync.WaitGroup
    	for i, name := range []string{"ana", "ben", "cy", "dee", "eve", "fox"} {
    		wg.Go(func() {
    			time.Sleep(time.Duration(i*10) * time.Millisecond) // couriers trickle in
    			b.Await(context.Background())
    			fmt.Printf("%v: %s leaves\n", time.Since(start).Round(10*time.Millisecond), name)
    		})
    	}
    	wg.Wait()
    }
  tests: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"sync"
    	"sync/atomic"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // arrive calls b.Await(ctx) in a goroutine and returns a channel that
    // receives its result.
    func arrive(b *Barrier, ctx context.Context) <-chan error {
    	ch := make(chan error, 1)
    	go func() { ch <- b.Await(ctx) }()
    	synctest.Wait()
    	return ch
    }

    // released reports which of the arrivals have returned.
    func released(chs ...<-chan error) []bool {
    	var out []bool
    	for _, ch := range chs {
    		out = append(out, len(ch) == 1)
    	}
    	return out
    }

    // cleanup releases anyone still waiting so the bubble can end.
    func cleanup(b *Barrier, n int) {
    	for range n {
    		go b.Await(context.Background())
    	}
    	synctest.Wait()
    }

    func TestBarrierReleasesTogether(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		b := NewBarrier(3)
    		ana := arrive(b, t.Context())
    		ben := arrive(b, t.Context())
    		if got := released(ana, ben); got[0] || got[1] {
    			defer cleanup(b, 3)
    			t.Fatalf("NewBarrier(3) with 2 arrivals: returned = %v, want [false false]", got)
    		}
    		cy := arrive(b, t.Context())
    		if got := released(ana, ben, cy); fmt.Sprint(got) != "[true true true]" {
    			defer cleanup(b, 3)
    			t.Fatalf("NewBarrier(3) with 3 arrivals: returned = %v, want [true true true]", got)
    		}
    		for i, ch := range []<-chan error{ana, ben, cy} {
    			if err := <-ch; err != nil {
    				t.Errorf("arrival %d: Await = %v, want nil", i+1, err)
    			}
    		}
    	})
    }

    func TestBarrierIsReusable(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		b := NewBarrier(2)
    		arrive(b, t.Context())
    		arrive(b, t.Context())

    		cy := arrive(b, t.Context())
    		if released(cy)[0] {
    			defer cleanup(b, 1)
    			t.Fatalf("NewBarrier(2): after one full round, a lone third arrival was let through; the second round needs 2 arrivals too")
    		}
    		dee := arrive(b, t.Context())
    		if got := released(cy, dee); !got[0] || !got[1] {
    			defer cleanup(b, 2)
    			t.Fatalf("NewBarrier(2): second round with 2 arrivals: returned = %v, want [true true]", got)
    		}
    	})
    }

    func TestBarrierCancelledArrivalLeaves(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		errShiftOver := errors.New("shift over")
    		b := NewBarrier(3)
    		ana := arrive(b, t.Context())
    		ctx, cancel := context.WithCancelCause(t.Context())
    		ben := arrive(b, ctx)

    		synctest.Sleep(5 * time.Second)
    		cancel(errShiftOver)
    		synctest.Wait()
    		if got := released(ben); !got[0] {
    			defer cleanup(b, 3)
    			t.Fatalf("Await still blocked after its ctx was cancelled")
    		}
    		if err := <-ben; !errors.Is(err, errShiftOver) {
    			t.Errorf("cancelled Await = %v, want the cause %q", err, errShiftOver)
    		}

    		cy := arrive(b, t.Context())
    		if got := released(ana, cy); got[0] || got[1] {
    			t.Fatalf("NewBarrier(3): ana waiting, ben cancelled, cy arrives: returned = %v, want [false false] (ben left, so only 2 of 3 are here)", got)
    		}
    		dee := arrive(b, t.Context())
    		if got := released(ana, cy, dee); fmt.Sprint(got) != "[true true true]" {
    			defer cleanup(b, 3)
    			t.Fatalf("ana, cy and dee arrived: returned = %v, want [true true true]", got)
    		}
    	})
    }

    func TestBarrierAlreadyCancelled(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		errClosed := errors.New("hub closed")
    		ctx, cancel := context.WithCancelCause(t.Context())
    		cancel(errClosed)
    		b := NewBarrier(2)
    		if err := <-arrive(b, ctx); !errors.Is(err, errClosed) {
    			t.Errorf("Await with a cancelled ctx = %v, want %q", err, errClosed)
    		}
    		ana := arrive(b, t.Context())
    		if released(ana)[0] {
    			t.Fatalf("NewBarrier(2): a cancelled caller plus ana released the round; the cancelled caller must not count")
    		}
    		arrive(b, t.Context())
    		if !released(ana)[0] {
    			t.Fatalf("NewBarrier(2): ana still waiting after a second courier arrived")
    		}
    	})
    }

    func TestBarrierManyRounds(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		const couriers, rounds = 4, 50
    		b := NewBarrier(couriers)
    		var arrived [couriers]atomic.Int32
    		var wg sync.WaitGroup
    		for c := range couriers {
    			wg.Go(func() {
    				for r := range rounds {
    					arrived[c].Add(1)
    					if err := b.Await(t.Context()); err != nil {
    						t.Errorf("Await = %v, want nil", err)
    						return
    					}
    					for other := range couriers {
    						if n := int(arrived[other].Load()); n < r+1 {
    							t.Errorf("courier %d passed round %d while courier %d had only arrived %d times", c, r+1, other, n)
    							return
    						}
    					}
    				}
    			})
    		}
    		done := make(chan struct{})
    		go func() {
    			wg.Wait()
    			close(done)
    		}()
    		select {
    		case <-done:
    		case <-time.After(time.Hour):
    			t.Fatalf("%d couriers doing %d rounds each got stuck: every round must release all %d and start afresh", couriers, rounds, couriers)
    		}
    	})
    }
---

Dispatchly's airport hub sends couriers out in **convoys**: nobody leaves
until `n` couriers are ready, and then they all go together. The next `n`
couriers form the next convoy, and so on, all day.

Implement a reusable `Barrier`:

- `NewBarrier(n)` creates a barrier for groups of `n`.
- `Await(ctx)` blocks until `n` couriers, counting this one, have arrived in
  the current round, and then returns `nil` for all of them at once. The
  barrier immediately starts a new, empty round.
- If `ctx` is done while waiting, the courier **leaves** the round (it no
  longer counts towards `n`) and `Await` returns `context.Cause(ctx)`. If
  `ctx` is already done, `Await` returns the cause without arriving at all.

## Example

```go
b := NewBarrier(3)
go b.Await(ctx) // ana waits
go b.Await(ctx) // ben waits
go b.Await(ctx) // cy arrives: ana, ben and cy all return nil
go b.Await(ctx) // dee waits: this is a new round
```

## Constraints

- Arrivals from different rounds must never mix: a courier arriving right
  after a round is released waits for `n` fresh arrivals.
- The tests use `synctest.Wait` to check exactly who is still blocked after
  each arrival, and `synctest` reports any goroutine left waiting at the end.
