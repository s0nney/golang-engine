---
title: AfterFunc and WithoutCancel
quiz:
  - question: 'You call `stop := context.AfterFunc(ctx, f)`, then later `stop()` returns `true`. What does that tell you?'
    options:
      - text: '`f` has finished running'
      - text: '`f` is running right now'
      - text: '`f` will never run, because `stop` unregistered it before `ctx` was cancelled'
        correct: true
      - text: '`ctx` has been cancelled'
    explanation: |
      `stop` returns `true` only if it prevented `f` from running. If it
      returns `false`, either `ctx` was already cancelled and `f` has been
      started in its own goroutine, or `f` had already been stopped.
  - question: 'An HTTP handler starts a goroutine to write an audit record *after* the response is sent. Which context should that goroutine use?'
    options:
      - text: The request's `ctx` as-is
      - text: '`context.WithoutCancel(ctx)`, ideally wrapped in its own `WithTimeout`'
        correct: true
      - text: '`nil`'
      - text: A context stored in a global variable
    explanation: |
      The request context is cancelled as soon as the handler returns, which
      would abort the audit write. `WithoutCancel` keeps the request's
      values (like a trace ID) but drops its cancellation. Because it has no
      deadline either, give it one of its own so a stuck write can't hang
      forever.
---

Two smaller tools round out the `context` package. One runs cleanup code when a context is cancelled; the other deliberately escapes cancellation.

## context.AfterFunc

`context.AfterFunc(ctx, f)` runs `f` **in its own goroutine** once `ctx` is cancelled. If `ctx` is already cancelled, `f` starts straight away. It returns a `stop` function that unregisters `f`:

```go
package main

import (
	"context"
	"fmt"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	released := make(chan struct{})
	context.AfterFunc(ctx, func() {
		fmt.Println("releasing courier ana back to the pool")
		close(released)
	})

	fmt.Println("courier ana reserved for order A1")
	cancel() // the customer cancels
	<-released

	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	stop := context.AfterFunc(ctx2, func() {
		fmt.Println("never printed")
	})
	fmt.Println("delivered, stop() =", stop())
	fmt.Println("stop() again =", stop())
}
```

```text
courier ana reserved for order A1
releasing courier ana back to the pool
delivered, stop() = true
stop() again = false
```

`f` runs in a separate goroutine, so `main` waits on `released` to be sure it has run. `stop()` returns `true` only if it actually prevented `f` from running. `stop` doesn't wait for a running `f` to finish, so coordinate yourself if you need that.

Why not just start a goroutine that does `<-ctx.Done(); cleanup()`? That goroutine sits around for the whole life of the context, and if the context is never cancelled, it leaks. `AfterFunc` costs no goroutine until the context is actually cancelled, and `stop` cleans it up.

### Making blocking calls cancellable

The killer use for `AfterFunc` is bridging contexts to things that *don't* take one. A network connection's `Read` blocks with no `ctx` parameter, but setting a deadline in the past wakes it up:

```go
stop := context.AfterFunc(ctx, func() {
	conn.SetReadDeadline(time.Now()) // unblocks any Read in progress
})
defer stop()

n, err := conn.Read(buf) // now effectively respects ctx
```

The same trick works for `sync.Cond.Wait`: call `Broadcast` from the `AfterFunc`.

## context.WithoutCancel

Sometimes work must **outlive** the request that started it. After Dispatchly confirms an order, it writes an audit record and sends a receipt email. If those use the request's context, they're cancelled the moment the handler returns, possibly halfway through.

You could use `context.Background()`, but then you'd lose the request's values, like its trace ID. `context.WithoutCancel(parent)` returns a child that keeps the parent's **values** but **none of its cancellation**:

```go
func confirmOrder(ctx context.Context, o Order) error {
	if err := save(ctx, o); err != nil {
		return err
	}

	auditCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	go func() {
		defer cancel()
		writeAudit(auditCtx, o) // keeps running after confirmOrder returns
	}()
	return nil
}
```

The returned context never reports an error, has no deadline, and its `Done()` returns `nil`, a channel that blocks forever. That's why the example wraps it in `WithTimeout`: detaching from the request shouldn't mean "allowed to hang forever".

Use `WithoutCancel` sparingly. Most work *should* stop when its request is cancelled. It's for genuine "finish this no matter what" tasks such as audit logs, cleanup and metrics.

## The family so far

| Function | Cancelled when |
| --- | --- |
| `WithCancel(parent)` | you call `cancel()` or the parent is cancelled |
| `WithCancelCause(parent)` | same, and you pass a cause |
| `WithTimeout` / `WithDeadline` | the time runs out, `cancel()`, or the parent |
| `WithTimeoutCause` / `WithDeadlineCause` | same, with a cause for the timeout |
| `WithoutCancel(parent)` | never |
| `AfterFunc(ctx, f)` | (not a context: runs `f` when `ctx` is cancelled) |
