---
title: The Rules of Context
quiz:
  - question: Which of these follows the context conventions?
    options:
      - text: '`func (d *Dispatcher) Assign(order Order, ctx context.Context) error`'
      - text: '`type Dispatcher struct { ctx context.Context }` set once in `NewDispatcher`'
      - text: '`func (d *Dispatcher) Assign(ctx context.Context, order Order) error`'
        correct: true
      - text: '`d.Assign(nil, order)` when you have no context handy'
    explanation: |
      `ctx` goes first and is passed to each call that needs it. Storing it
      in a struct ties every call to one lifetime, and passing `nil` panics
      as soon as anything calls a method on it. Use `context.TODO()` as a
      placeholder instead.
  - question: What's a good use of `context.WithValue`?
    options:
      - text: Passing the database connection to every function
      - text: Passing an optional `retries` setting so you don't have to change function signatures
      - text: Carrying a request's trace ID so logs from every layer can include it
        correct: true
      - text: Returning results from a goroutine to its caller
    explanation: |
      Context values are for request-scoped data that crosses API
      boundaries, like trace IDs or the authenticated user. Dependencies and
      options should be explicit parameters or struct fields, where the
      compiler can check them.
exercise:
  starter: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"time"
    )

    var errDelivered = errors.New("order delivered")

    // pollCourier calls ping every interval and collects the locations it
    // returns, until ctx is done. It then returns what it collected along
    // with the reason ctx was cancelled (its cause).
    func pollCourier(ctx context.Context, every time.Duration, ping func() string) ([]string, error) {
    	var locs []string
    	for {
    		time.Sleep(every)
    		if err := ctx.Err(); err != nil {
    			return locs, err
    		}
    		locs = append(locs, ping())
    	}
    }

    func main() {
    	ctx, cancel := context.WithCancelCause(context.Background())
    	go func() {
    		time.Sleep(75 * time.Millisecond)
    		cancel(errDelivered)
    	}()

    	n := 0
    	ping := func() string {
    		n++
    		return fmt.Sprintf("block %d", n)
    	}
    	start := time.Now()
    	locs, err := pollCourier(ctx, 20*time.Millisecond, ping)
    	fmt.Println(locs)
    	fmt.Println("stopped because:", err)
    	fmt.Println("stopped after", time.Since(start).Round(5*time.Millisecond))
    }
  solution: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"time"
    )

    var errDelivered = errors.New("order delivered")

    // pollCourier calls ping every interval and collects the locations it
    // returns, until ctx is done. It then returns what it collected along
    // with the reason ctx was cancelled (its cause).
    func pollCourier(ctx context.Context, every time.Duration, ping func() string) ([]string, error) {
    	ticker := time.NewTicker(every)
    	defer ticker.Stop()

    	var locs []string
    	for {
    		select {
    		case <-ctx.Done():
    			return locs, context.Cause(ctx)
    		case <-ticker.C:
    			locs = append(locs, ping())
    		}
    	}
    }

    func main() {
    	ctx, cancel := context.WithCancelCause(context.Background())
    	go func() {
    		time.Sleep(75 * time.Millisecond)
    		cancel(errDelivered)
    	}()

    	n := 0
    	ping := func() string {
    		n++
    		return fmt.Sprintf("block %d", n)
    	}
    	start := time.Now()
    	locs, err := pollCourier(ctx, 20*time.Millisecond, ping)
    	fmt.Println(locs)
    	fmt.Println("stopped because:", err)
    	fmt.Println("stopped after", time.Since(start).Round(5*time.Millisecond))
    }
  tests: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"slices"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    func counter() func() string {
    	n := 0
    	return func() string {
    		n++
    		return fmt.Sprint("loc", n)
    	}
    }

    func TestStopsAtDeadline(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		ctx, cancel := context.WithTimeout(t.Context(), 35*time.Second)
    		defer cancel()

    		start := time.Now()
    		locs, err := pollCourier(ctx, 10*time.Second, counter())
    		took := time.Since(start)

    		if want := []string{"loc1", "loc2", "loc3"}; !slices.Equal(locs, want) {
    			t.Errorf("polling every 10s with a 35s timeout collected %v, want %v", locs, want)
    		}
    		if !errors.Is(err, context.DeadlineExceeded) {
    			t.Errorf("pollCourier returned error %v, want context.DeadlineExceeded", err)
    		}
    		if took != 35*time.Second {
    			t.Errorf("pollCourier returned after %v, want 35s (the moment the context expired, not at the next poll)", took)
    		}
    	})
    }

    func TestReturnsCause(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		ctx, cancel := context.WithCancelCause(t.Context())
    		go func() {
    			time.Sleep(25 * time.Second)
    			cancel(errDelivered)
    		}()

    		start := time.Now()
    		locs, err := pollCourier(ctx, 10*time.Second, counter())
    		took := time.Since(start)

    		if len(locs) != 2 {
    			t.Errorf("polling every 10s, cancelled at 25s: collected %d locations, want 2", len(locs))
    		}
    		if err != errDelivered {
    			t.Errorf("pollCourier returned error %v, want errDelivered (use context.Cause, not ctx.Err)", err)
    		}
    		if took != 25*time.Second {
    			t.Errorf("pollCourier returned after %v, want 25s (the moment the context was cancelled)", took)
    		}
    	})
    }

    func TestAlreadyCancelled(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		ctx, cancel := context.WithCancelCause(t.Context())
    		cancel(errDelivered)

    		start := time.Now()
    		locs, err := pollCourier(ctx, time.Minute, counter())
    		if len(locs) != 0 {
    			t.Errorf("pollCourier with an already-cancelled context collected %v, want nothing", locs)
    		}
    		if err != errDelivered {
    			t.Errorf("pollCourier returned error %v, want errDelivered", err)
    		}
    		if took := time.Since(start); took != 0 {
    			t.Errorf("pollCourier with an already-cancelled context took %v, want it to return immediately", took)
    		}
    	})
    }
---

The `context` package comes with a set of conventions. They're written in its documentation and followed by the whole Go ecosystem, so code that breaks them sticks out.

## 1. ctx is the first parameter, named ctx

```go
func (d *Dispatcher) Assign(ctx context.Context, order Order) (Courier, error)
```

Always first, always called `ctx`. Anyone reading a signature can immediately see that the function may block and can be cancelled. If a function takes a `ctx`, it should respect it.

## 2. Don't store a Context in a struct

```go
// Don't
type Dispatcher struct {
	ctx context.Context
}
```

A context belongs to one *operation*: this request, this batch. A `Dispatcher` lives much longer and serves many requests, each with its own deadline and cancellation. Stored in a struct, a context either gets shared by requests it has nothing to do with, or it quietly goes stale. Pass it explicitly to each method that needs it.

(The standard library bends this rule in a few places for backwards compatibility, like `http.Request`, which holds its request's context. Your code shouldn't need to.)

## 3. Never pass a nil Context

Even if a function "doesn't really use it", pass `context.TODO()` if you don't have a real context yet. Calling `ctx.Done()` on a nil interface panics.

## 4. Derive, don't replace

Inside a function, build new contexts from the one you were given: `WithTimeout(ctx, ...)`, not `WithTimeout(context.Background(), ...)`. Starting a fresh root cuts your work off from the caller's cancellation and deadline. The one legitimate exception is `WithoutCancel`, when you really mean to outlive the caller.

## 5. Values are for request-scoped data only

`context.WithValue` attaches a key/value pair, and `ctx.Value(key)` looks it up by walking up the tree. Use it for data that belongs to the *request* and has to cross API boundaries, like a trace ID or the authenticated user, and use an **unexported key type** so no other package can collide with yours:

```go
package main

import (
	"context"
	"fmt"
)

type traceKey struct{} // unexported: no other package can use this key

func withTrace(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceKey{}, id)
}

func traceID(ctx context.Context) string {
	id, ok := ctx.Value(traceKey{}).(string)
	if !ok {
		return "no-trace"
	}
	return id
}

func assignCourier(ctx context.Context, order string) {
	fmt.Printf("[%s] assigning a courier to %s\n", traceID(ctx), order)
}

func main() {
	ctx := withTrace(context.Background(), "req-7f3a")
	assignCourier(ctx, "A1")
	assignCourier(context.Background(), "B2")
}
```

```text
[req-7f3a] assigning a courier to A1
[no-trace] assigning a courier to B2
```

Don't use values for things a function *needs*: database handles, config, loggers, optional parameters. Those are invisible in signatures and unchecked by the compiler. If it's required, make it a parameter.

## 6. Return promptly, and say why

When the context is done, stop what you're doing and return, ideally with `ctx.Err()` or `context.Cause(ctx)` (possibly wrapped). A function that notices cancellation but carries on for another ten seconds isn't really respecting it. Wherever your code *waits*, the wait should be a `select` with a `case <-ctx.Done():`.

A context is safe to use from many goroutines at once, so it's fine to hand the same `ctx` to every goroutine working on a request.

## Your turn

While a courier is on a delivery, Dispatchly polls their phone for its location. `pollCourier` should call `ping` once every `every`, collecting the results, until `ctx` is done. Then it returns what it collected along with **why** it stopped, the context's *cause*.

The starter uses `time.Sleep` and checks `ctx.Err()` after each nap. Press **Run**: it stops late, and the reason is just `context canceled`. Rewrite it so it:

- returns **the moment** `ctx` is done, not at the next poll (and immediately if `ctx` is already done),
- returns `context.Cause(ctx)` as its error,
- polls on a steady rhythm with a `time.Ticker`.
