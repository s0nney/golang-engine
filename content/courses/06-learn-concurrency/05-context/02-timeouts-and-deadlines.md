---
title: Timeouts and Deadlines
quiz:
  - question: |
      What does `d` end up as?

      ```go
      parent, c1 := context.WithTimeout(context.Background(), 2*time.Second)
      defer c1()
      child, c2 := context.WithTimeout(parent, 10*time.Second)
      defer c2()
      dl, _ := child.Deadline()
      d := time.Until(dl).Round(time.Second)
      ```
    options:
      - text: '`10s`'
      - text: '`12s`'
      - text: '`2s`'
        correct: true
      - text: '`8s`'
    explanation: |
      A child can never outlive its parent. The parent will be cancelled in
      2 seconds, which cancels the child too, so the child's effective
      deadline is the earlier of the two. A context's deadline can only be
      shortened as you go down the tree, never extended.
  - question: A context made with `context.WithTimeout` hits its timeout. What does `ctx.Err()` return?
    options:
      - text: '`context.Canceled`'
      - text: '`context.DeadlineExceeded`'
        correct: true
      - text: '`nil`, because nobody called `cancel`'
      - text: 'An `*errors.TimeoutError`'
    explanation: |
      A context cancelled because its deadline passed reports
      `context.DeadlineExceeded`. One cancelled by calling `cancel()`
      reports `context.Canceled`. Test for them with `errors.Is`, because
      they're often wrapped.
---

Cancelling by hand covers "the customer gave up". The other everyday case is "this is taking too long". Contexts have that built in.

## WithTimeout and WithDeadline

```go
ctx, cancel := context.WithTimeout(parent, 2*time.Second)
defer cancel()
```

The returned context is cancelled automatically after 2 seconds, or earlier if you call `cancel` or the parent is cancelled. `WithDeadline` does the same with an absolute time:

```go
ctx, cancel := context.WithDeadline(parent, time.Now().Add(2*time.Second))
```

`WithTimeout(parent, d)` is literally `WithDeadline(parent, time.Now().Add(d))`. Use whichever reads better: timeouts for "this step may take 2 seconds", deadlines for "the lunch rush promo ends at 14:00".

You still need `defer cancel()`. If the work finishes in 10ms, calling `cancel` releases the timer straight away instead of holding it for the full 2 seconds.

## Using it

Functions don't care *why* a context might be cancelled. They watch `ctx.Done()` as before:

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// quote asks a restaurant for a prep-time estimate. It takes `work` to answer.
func quote(ctx context.Context, restaurant string, work time.Duration) (time.Duration, error) {
	select {
	case <-time.After(work):
		return 15 * time.Minute, nil
	case <-ctx.Done():
		return 0, fmt.Errorf("quote from %s: %w", restaurant, ctx.Err())
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	d, err := quote(ctx, "pho-king", 10*time.Millisecond)
	fmt.Println(d, err)

	d, err = quote(ctx, "thai-tanic", 200*time.Millisecond)
	fmt.Println(d, err)
	fmt.Println(errors.Is(err, context.DeadlineExceeded))
}
```

```text
15m0s <nil>
0s quote from thai-tanic: context deadline exceeded
true
```

Notice that both calls share **one** 50ms budget. The first used 10ms, so the second had 40ms left. That's the big advantage over `time.After` in each function: the time limit belongs to the *request*, and every step underneath spends from the same budget.

## Canceled vs DeadlineExceeded

`ctx.Err()` tells you which kind of stop happened:

| Cause | `ctx.Err()` |
| --- | --- |
| Someone called `cancel()` | `context.Canceled` |
| The deadline passed | `context.DeadlineExceeded` |

Check with `errors.Is`, since functions usually wrap it, as `quote` does. The distinction matters for how you react: a timeout from a restaurant API might be worth a retry or an alert, while a customer who cancelled needs neither.

## Deadlines only get shorter

A child's deadline can never be later than its parent's. If the incoming request has 2 seconds left and you call `WithTimeout(ctx, 10*time.Second)`, the child still ends in 2 seconds, because the parent cancels it. So it's always safe to add a timeout for a single step. It can only tighten the budget.

A function can also look at the remaining budget with `ctx.Deadline()` and skip work that can't finish in time:

```go
if dl, ok := ctx.Deadline(); ok && time.Until(dl) < 500*time.Millisecond {
	return errNotEnoughTime // don't even start the slow route optimizer
}
```

## Where to put timeouts

Put them at the edges, where you know how long something *should* take: an incoming request as a whole, and each outgoing call to something you don't control. Don't sprinkle them through every internal function. Those just pass `ctx` along and let the caller's budget decide.

## Further reading

- [Go by Example: Context](https://gobyexample.com/context)
