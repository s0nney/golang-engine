---
title: Context and Cancellation
quiz:
  - question: 'A parent context has two children made with `context.WithCancel`. What happens when you call the **first child''s** `cancel`?'
    options:
      - text: The parent and both children are cancelled
      - text: Only the first child (and any contexts derived from it) is cancelled
        correct: true
      - text: Both children are cancelled, but not the parent
      - text: Nothing, only the root context can be cancelled
    explanation: |
      Cancellation flows **down** the tree, never up or sideways. Cancelling
      a context cancels it and everything derived from it. The parent and
      its other children are unaffected.
  - question: What does `ctx.Err()` return while `ctx` has **not** been cancelled yet?
    options:
      - text: '`context.Canceled`'
      - text: '`nil`'
        correct: true
      - text: It blocks until the context is cancelled
      - text: '`context.DeadlineExceeded`'
    explanation: |
      `Err` never blocks. It returns `nil` until `Done` is closed, and after
      that a non-nil error explaining why: `context.Canceled` or
      `context.DeadlineExceeded`.
  - question: Why must you always call the `cancel` function returned by `context.WithCancel`, even if the work finished normally?
    options:
      - text: Otherwise the program won't compile
      - text: To release the resources the context holds, such as its link to the parent; `go vet` warns if you don't
        correct: true
      - text: To make `ctx.Err()` return `nil` again
      - text: It closes the parent context's `Done` channel
    explanation: |
      A cancellable child is registered with its parent until it's
      cancelled. Forgetting `cancel` keeps that registration (and anything
      it references) around until the parent is cancelled, which may be
      never. `defer cancel()` right after creating it is the idiom.
---

A done channel stops goroutines, but it doesn't scale. When an HTTP request comes in, Dispatchly might validate the order, call three restaurant APIs, query the database and ask a dozen couriers for ETAs, each in its own goroutines, some starting helpers of their own. When the customer closes the app, *all* of that should stop. `context.Context` is the standard way to carry that "stop" signal, plus deadlines, through a whole call tree.

## The interface

```go
type Context interface {
	Done() <-chan struct{}
	Err() error
	Deadline() (deadline time.Time, ok bool)
	Value(key any) any
}
```

- **`Done()`** returns a channel that's **closed** when the context is cancelled. It's the done channel from the last chapter, built in.
- **`Err()`** returns `nil` while the context is live, then `context.Canceled` or `context.DeadlineExceeded` once it's done.
- **`Deadline()`** reports when the context will be cancelled automatically, if ever (next lesson).
- **`Value(key)`** carries request-scoped values (last lesson of this chapter).

## Making contexts

Every context tree starts at a root, `context.Background()`. It's never cancelled and has no deadline. `context.TODO()` is the same thing with a different name. Use it as a placeholder when you haven't worked out which context to pass yet.

In tests, use `t.Context()`: it's cancelled automatically just before the test finishes.

To get a context you can cancel, derive one from a parent:

```go
ctx, cancel := context.WithCancel(parent)
defer cancel()
```

## Watching for cancellation

A function that might take a while accepts a `ctx` and watches `ctx.Done()` wherever it waits:

```go
package main

import (
	"context"
	"fmt"
	"time"
)

// searchCouriers looks for a free courier until it finds one or ctx is cancelled.
func searchCouriers(ctx context.Context) {
	for attempt := 1; ; attempt++ {
		select {
		case <-ctx.Done():
			fmt.Println("search stopped:", ctx.Err())
			return
		case <-time.After(20 * time.Millisecond):
			fmt.Println("attempt", attempt, "- no courier free yet")
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopped := make(chan struct{})
	go func() {
		searchCouriers(ctx)
		close(stopped)
	}()

	time.Sleep(50 * time.Millisecond)
	fmt.Println("customer cancelled the order")
	cancel()
	<-stopped
}
```

```text
attempt 1 - no courier free yet
attempt 2 - no courier free yet
customer cancelled the order
search stopped: context canceled
```

For a loop doing CPU work with no channel to wait on, check without blocking at the top of each iteration:

```go
for _, leg := range route {
	if err := ctx.Err(); err != nil {
		return err
	}
	plan(leg)
}
```

## The tree

Each `With...` function wraps a parent and returns a child. Cancelling a context cancels **all its descendants**, but never its parent or siblings:

```text
Background
└── request ctx          (cancelled when the customer disconnects)
    ├── restaurant ctx   (WithTimeout 2s)
    └── courier ctx      (WithCancel: stop once one courier accepts)
        ├── ask ana
        └── ask ben
```

When the customer disconnects, everything below the request stops. When a courier accepts, the courier branch stops, but the restaurant call carries on. You get this for free just by passing each function the right `ctx`.

## Always call cancel

`WithCancel` (and friends) return a `cancel` function, and **you must call it**, typically with `defer cancel()` right away. Until it's called (or the parent is cancelled) the child stays registered with its parent, holding memory. `go vet` reports code paths that lose a `cancel` function. Calling it more than once is fine: only the first call does anything.

## Cancellation is cooperative

Cancelling a context doesn't kill anything. There's no way in Go to stop a goroutine from the outside. `cancel()` closes a channel. It's up to each function to notice and return. Every standard library function that takes a `ctx`, like `http.NewRequestWithContext` or `database/sql` queries, does this for you. Your own code must watch `ctx.Done()` or check `ctx.Err()`, and return promptly, usually returning `ctx.Err()` (or a wrapped version of it) as its error.

## Further reading

- [Go blog: Go Concurrency Patterns: Context](https://go.dev/blog/context)
- [`context` package documentation](https://pkg.go.dev/context)
