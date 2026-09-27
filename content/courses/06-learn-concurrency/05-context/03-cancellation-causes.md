---
title: Cancellation Causes
quiz:
  - question: |
      What does this print?

      ```go
      ctx, cancel := context.WithCancelCause(context.Background())
      cancel(errors.New("payment declined"))
      cancel(errors.New("courier unavailable"))
      fmt.Println(ctx.Err(), "/", context.Cause(ctx))
      ```
    options:
      - text: '`payment declined / payment declined`'
      - text: '`context canceled / courier unavailable`'
      - text: '`context canceled / payment declined`'
        correct: true
      - text: '`nil / payment declined`'
    explanation: |
      `Err` keeps its usual meaning, `context.Canceled`, so existing code
      that checks it still works. `Cause` returns the error passed to the
      **first** cancel call. Later calls do nothing, just like calling a
      plain `cancel` twice.
  - question: What does `context.Cause(ctx)` return for a context made with plain `context.WithCancel`, after `cancel()` is called?
    options:
      - text: '`nil`'
      - text: '`context.Canceled`, the same as `ctx.Err()`'
        correct: true
      - text: It panics, because the context has no cause
      - text: An error saying no cause was set
    explanation: |
      If no cause was set anywhere up the tree, `Cause` falls back to
      `ctx.Err()`. So it's always safe to call `context.Cause` instead of
      `ctx.Err()` when you want the most specific reason available.
---

`ctx.Err()` only ever says `context canceled` or `context deadline exceeded`. When an order's goroutines all stop and the logs say `context canceled` three hundred times, you'll want to know **why**. Was it the customer? A failed payment? A sibling goroutine that hit an error? Since Go 1.20 a context can carry that reason: its **cause**.

## WithCancelCause

`context.WithCancelCause` works like `WithCancel`, except its cancel function takes an error:

```go
package main

import (
	"context"
	"errors"
	"fmt"
)

var errOrderCancelled = errors.New("customer cancelled the order")

func main() {
	ctx, cancel := context.WithCancelCause(context.Background())

	cancel(errOrderCancelled)
	cancel(errors.New("too late, already cancelled")) // ignored

	fmt.Println("Err:  ", ctx.Err())
	fmt.Println("Cause:", context.Cause(ctx))
	fmt.Println(errors.Is(context.Cause(ctx), errOrderCancelled))

	child, stop := context.WithCancel(ctx) // children inherit the cause
	defer stop()
	fmt.Println("child:", context.Cause(child))
}
```

```text
Err:   context canceled
Cause: customer cancelled the order
true
child: customer cancelled the order
```

A few things to notice:

- **`ctx.Err()` doesn't change.** It's still `context.Canceled`, so all existing code that checks `errors.Is(err, context.Canceled)` keeps working.
- **`context.Cause(ctx)`** returns the cause. It's a function, not a method, because the `Context` interface couldn't gain a method without breaking every implementation.
- **The first cause wins.** Later cancels are ignored.
- **Descendants see the cause too**, so a goroutine deep in the tree can find out why its whole branch was cancelled.
- `cancel(nil)` sets the cause to `context.Canceled`.

If no cause was ever set, `context.Cause` returns the same as `ctx.Err()`. That means you can always use `Cause` where you'd have used `Err` and get the most specific reason available.

## Timeouts with causes

`WithTimeoutCause` and `WithDeadlineCause` attach a cause that's used **if the timer fires**:

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var errKitchenSlow = errors.New("kitchen took longer than 30 minutes")

func main() {
	ctx, cancel := context.WithTimeoutCause(context.Background(), 20*time.Millisecond, errKitchenSlow)
	defer cancel()

	<-ctx.Done()
	fmt.Println("Err:  ", ctx.Err())
	fmt.Println("Cause:", context.Cause(ctx))
}
```

```text
Err:   context deadline exceeded
Cause: kitchen took longer than 30 minutes
```

Their returned cancel function is a plain `CancelFunc` with no argument. If you cancel early by hand, the cause is just `context.Canceled`.

## Returning the cause

When a function stops because of its context, return the cause, so the caller sees the real reason:

```go
select {
case <-ctx.Done():
	return context.Cause(ctx)
case loc := <-pings:
	// ...
}
```

The most important use is coming up in chapter 7: when one of several goroutines fails, it cancels the shared context **with its error as the cause**, and every other goroutine stops. The caller then gets the error that started it all, not a pile of `context canceled`.

## Checking for cancellation

Code that wants to know "was I cancelled or did I fail?" should still check `ctx.Err()`, or use `errors.Is` against `context.Canceled` and `context.DeadlineExceeded`. Use `context.Cause` when you want to *report* or *react to* the specific reason. The two work together:

```go
if ctx.Err() != nil {
	log.Printf("order %s stopped: %v", id, context.Cause(ctx))
}
```
