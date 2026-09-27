---
title: 'Build It: First Error Cancels the Rest'
quiz:
  - question: Three tasks run in a group. Task A fails after 1s. Tasks B and C would take 10s but watch the group's `ctx`. Roughly when does `Wait` return, and with what?
    options:
      - text: After 10s, with A's error
      - text: After 1s, with `nil`, since B and C succeeded
      - text: After about 1s, with A's error, because A's failure cancels `ctx` and B and C return early
        correct: true
      - text: After 1s, with `context canceled`
    explanation: |
      The first error cancels the shared context, B and C see `ctx.Done()`
      and give up, and `Wait` returns as soon as all three have returned.
      It reports A's error, the root cause, not the `context canceled`
      errors that B and C returned as a consequence.
  - question: Why does the group record only the **first** error?
    options:
      - text: Because Go functions can only return one error
      - text: Once the first task fails, the others usually fail *because* they were cancelled, so their errors are just noise
        correct: true
      - text: To save memory
      - text: Because `sync.WaitGroup` only stores one error
    explanation: |
      After the first failure everything else is cancelled, so later errors
      are nearly always "context canceled" or similar. The first error is
      the one that explains what went wrong.
exercise:
  starter: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"sync"
    	"time"
    )

    // Group runs tasks in goroutines. The first task to return an error
    // cancels the group's context, and Wait returns that first error.
    type Group struct {
    	cancel context.CancelCauseFunc
    	wg     sync.WaitGroup
    	once   sync.Once
    	err    error
    }

    // WithContext returns a new Group and a context derived from ctx that is
    // cancelled when a task fails, or when Wait returns.
    func WithContext(ctx context.Context) (*Group, context.Context) {
    	ctx, cancel := context.WithCancelCause(ctx)
    	return &Group{cancel: cancel}, ctx
    }

    // Go runs f in a new goroutine. If f returns an error and it's the
    // first one, record it and cancel the group's context with it as the cause.
    func (g *Group) Go(f func() error) {
    	g.wg.Go(func() {
    		f() // TODO: don't ignore the error
    	})
    }

    // Wait waits for every task to finish, cancels the context, and returns
    // the first error (or nil).
    func (g *Group) Wait() error {
    	g.wg.Wait()
    	return nil
    }

    var errCardDeclined = errors.New("card declined")

    // step pretends to do some work that takes d, giving up if ctx is cancelled.
    func step(ctx context.Context, name string, d time.Duration, err error) error {
    	select {
    	case <-time.After(d):
    		fmt.Println(name, "done")
    		return err
    	case <-ctx.Done():
    		fmt.Println(name, "stopped:", context.Cause(ctx))
    		return context.Cause(ctx)
    	}
    }

    func main() {
    	start := time.Now()
    	g, ctx := WithContext(context.Background())
    	g.Go(func() error { return step(ctx, "reserve courier", 80*time.Millisecond, nil) })
    	g.Go(func() error { return step(ctx, "charge card", 20*time.Millisecond, errCardDeclined) })
    	g.Go(func() error { return step(ctx, "notify restaurant", 60*time.Millisecond, nil) })

    	err := g.Wait()
    	fmt.Println("order failed:", err)
    	fmt.Println("took about", time.Since(start).Round(20*time.Millisecond))
    }
  solution: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"sync"
    	"time"
    )

    // Group runs tasks in goroutines. The first task to return an error
    // cancels the group's context, and Wait returns that first error.
    type Group struct {
    	cancel context.CancelCauseFunc
    	wg     sync.WaitGroup
    	once   sync.Once
    	err    error
    }

    // WithContext returns a new Group and a context derived from ctx that is
    // cancelled when a task fails, or when Wait returns.
    func WithContext(ctx context.Context) (*Group, context.Context) {
    	ctx, cancel := context.WithCancelCause(ctx)
    	return &Group{cancel: cancel}, ctx
    }

    // Go runs f in a new goroutine. If f returns an error and it's the
    // first one, record it and cancel the group's context with it as the cause.
    func (g *Group) Go(f func() error) {
    	g.wg.Go(func() {
    		if err := f(); err != nil {
    			g.once.Do(func() {
    				g.err = err
    				g.cancel(err)
    			})
    		}
    	})
    }

    // Wait waits for every task to finish, cancels the context, and returns
    // the first error (or nil).
    func (g *Group) Wait() error {
    	g.wg.Wait()
    	g.cancel(g.err)
    	return g.err
    }

    var errCardDeclined = errors.New("card declined")

    // step pretends to do some work that takes d, giving up if ctx is cancelled.
    func step(ctx context.Context, name string, d time.Duration, err error) error {
    	select {
    	case <-time.After(d):
    		fmt.Println(name, "done")
    		return err
    	case <-ctx.Done():
    		fmt.Println(name, "stopped:", context.Cause(ctx))
    		return context.Cause(ctx)
    	}
    }

    func main() {
    	start := time.Now()
    	g, ctx := WithContext(context.Background())
    	g.Go(func() error { return step(ctx, "reserve courier", 80*time.Millisecond, nil) })
    	g.Go(func() error { return step(ctx, "charge card", 20*time.Millisecond, errCardDeclined) })
    	g.Go(func() error { return step(ctx, "notify restaurant", 60*time.Millisecond, nil) })

    	err := g.Wait()
    	fmt.Println("order failed:", err)
    	fmt.Println("took about", time.Since(start).Round(20*time.Millisecond))
    }
  tests: |
    package main

    import (
    	"context"
    	"errors"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // work finishes after d with err, or returns early if ctx is cancelled.
    func work(ctx context.Context, d time.Duration, err error) error {
    	select {
    	case <-time.After(d):
    		return err
    	case <-ctx.Done():
    		return context.Cause(ctx)
    	}
    }

    func TestFirstErrorCancelsTheRest(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		errA := errors.New("A failed")
    		g, ctx := WithContext(t.Context())
    		start := time.Now()

    		g.Go(func() error { return work(ctx, time.Second, errA) })
    		g.Go(func() error { return work(ctx, time.Hour, nil) })
    		g.Go(func() error { return work(ctx, time.Hour, nil) })
    		err := g.Wait()

    		if err != errA {
    			t.Errorf("Wait() = %v, want the failing task's error %q", err, errA)
    		}
    		if took := time.Since(start); took != time.Second {
    			t.Errorf("Wait returned after %v, want 1s: the failure should cancel ctx so the hour-long tasks stop right away", took)
    		}
    		if cause := context.Cause(ctx); cause != errA {
    			t.Errorf("context.Cause(ctx) = %v, want %q (cancel the context with the first error as its cause)", cause, errA)
    		}
    	})
    }

    func TestOnlyFirstErrorIsKept(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		errA, errB := errors.New("A failed"), errors.New("B failed")
    		g, ctx := WithContext(t.Context())

    		g.Go(func() error { return work(ctx, time.Second, errA) })
    		g.Go(func() error { time.Sleep(2 * time.Second); return errB }) // ignores ctx
    		if err := g.Wait(); err != errA {
    			t.Errorf("Wait() = %v, want the first error %q, not a later one", err, errA)
    		}
    		if cause := context.Cause(ctx); cause != errA {
    			t.Errorf("context.Cause(ctx) = %v, want the first error %q", cause, errA)
    		}
    	})
    }

    func TestAllSucceed(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		g, ctx := WithContext(t.Context())
    		results := make([]int, 3)
    		for i := range results {
    			g.Go(func() error {
    				time.Sleep(time.Duration(i+1) * time.Second)
    				if ctx.Err() != nil {
    					t.Errorf("task %d: ctx was cancelled while every task was still succeeding", i)
    				}
    				results[i] = i * 10
    				return nil
    			})
    		}
    		if err := g.Wait(); err != nil {
    			t.Errorf("Wait() = %v, want nil when every task succeeds", err)
    		}
    		if results[0] != 0 || results[1] != 10 || results[2] != 20 {
    			t.Errorf("results = %v, want [0 10 20] (Wait must wait for every task)", results)
    		}
    		if ctx.Err() == nil {
    			t.Errorf("ctx is still live after Wait returned, want it cancelled so its resources are released")
    		}
    	})
    }

    func TestWaitWithNoTasks(t *testing.T) {
    	g, _ := WithContext(t.Context())
    	if err := g.Wait(); err != nil {
    		t.Errorf("Wait() with no tasks = %v, want nil", err)
    	}
    }
---

Placing an order in Dispatchly means three things happening at once: reserve a courier, charge the card, notify the restaurant. They run concurrently to save time, but they succeed or fail **together**. If the card is declined, there's no point waiting for the restaurant's slow API. Everything should stop, and the caller should hear "card declined", not three variations of "context canceled".

This is so common that the Go team maintains a package for it, `golang.org/x/sync/errgroup`. It lives outside the standard library, so in this lesson you'll build the core of it yourself from pieces you already know.

## The API

```go
g, ctx := WithContext(parentCtx)

g.Go(func() error { return reserveCourier(ctx, order) })
g.Go(func() error { return chargeCard(ctx, order) })
g.Go(func() error { return notifyRestaurant(ctx, order) })

if err := g.Wait(); err != nil {
	return fmt.Errorf("placing order %s: %w", order.ID, err)
}
```

- **`WithContext`** returns a group and a context derived from the parent.
- **`Go`** runs a task in a new goroutine. Tasks return an `error`.
- **`Wait`** blocks until every task has returned, then returns the **first** non-nil error, or `nil`.
- The first failure **cancels `ctx`**, so the other tasks, which should all be watching `ctx`, stop early.

## The ingredients

Everything you need is in the standard library:

| Need | Tool |
| --- | --- |
| Run tasks and wait for them all | `sync.WaitGroup` and its `Go` method |
| Cancel the others, with the reason | `context.WithCancelCause` |
| Keep only the first error, even if several fail at once | `sync.Once` |

Here's the skeleton:

```go
type Group struct {
	cancel context.CancelCauseFunc
	wg     sync.WaitGroup
	once   sync.Once
	err    error
}

func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancelCause(ctx)
	return &Group{cancel: cancel}, ctx
}
```

Using a cause means that anything checking `context.Cause(ctx)`, such as a task deciding what to log, learns the real reason ("card declined") rather than just `context canceled`.

## Why sync.Once for the error?

Two tasks might fail at nearly the same moment. Without protection, both would write `g.err` at once, which is a data race, and the later one might overwrite the earlier. `once.Do` guarantees the recording function runs exactly once: the first failure wins, and every later one is ignored. A mutex would work too, but `Once` says exactly what you mean.

## Why Wait cancels the context too

Even when every task succeeds, `Wait` should cancel `ctx` before returning. As you saw in the context chapter, a cancellable context holds resources until it's cancelled. The group created it, so the group is responsible for cancelling it. It also means a task can't accidentally keep using `ctx` after the group is done.

## Your turn

`WithContext` and the `Group` struct are done. Finish `Go` and `Wait`:

- **`Go(f)`** runs `f` in a goroutine tracked by the WaitGroup. If `f` returns an error and it's the **first** error in the group, record it and cancel the context **with that error as the cause**.
- **`Wait()`** waits for every task, then cancels the context (even if everything succeeded) and returns the first error, or `nil`.

Press **Run** first. In the starter the card is declined, but the other steps run to completion and the order "fails" with `<nil>`. When you're done, the other steps should stop as soon as the card is declined.

## Further reading

- [`golang.org/x/sync/errgroup`](https://pkg.go.dev/golang.org/x/sync/errgroup), the real thing, which also has `SetLimit` to bound concurrency like a semaphore
