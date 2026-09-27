---
title: Kitchen Deadline
difficulty: easy
after: context
hints:
  - '`context.WithTimeout` gives you the deadline, but when it expires `context.Cause(ctx)` is just `context.DeadlineExceeded`. There''s a variant that lets you choose the cause.'
  - '`context.WithTimeoutCause(parent, d, cause)` returns a context that expires after `d` with `ctx.Err() == context.DeadlineExceeded` and `context.Cause(ctx) == cause`. It returns a plain `context.CancelFunc`, so you can return its results directly.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"time"
    )

    // ErrKitchenTooSlow is the cause of a prep deadline running out.
    var ErrKitchenTooSlow = errors.New("kitchen took too long")

    // prepDeadline returns a context derived from parent that expires prep
    // from now. When it expires, ctx.Err() is context.DeadlineExceeded and
    // context.Cause(ctx) is ErrKitchenTooSlow.
    func prepDeadline(parent context.Context, prep time.Duration) (context.Context, context.CancelFunc) {
    	// TODO: this has no deadline at all. Look for a context.With... function
    	// that takes both a timeout and a cause.
    	return context.WithCancel(parent)
    }

    func main() {
    	ctx, cancel := prepDeadline(context.Background(), 50*time.Millisecond)
    	defer cancel()

    	select {
    	case <-time.After(200 * time.Millisecond): // the kitchen is slow today
    		fmt.Println("order ready")
    	case <-ctx.Done():
    		fmt.Println("gave up:", ctx.Err(), "/", context.Cause(ctx))
    	}
    	// want: gave up: context deadline exceeded / kitchen took too long
    }
  solution: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"time"
    )

    // ErrKitchenTooSlow is the cause of a prep deadline running out.
    var ErrKitchenTooSlow = errors.New("kitchen took too long")

    // prepDeadline returns a context derived from parent that expires prep
    // from now. When it expires, ctx.Err() is context.DeadlineExceeded and
    // context.Cause(ctx) is ErrKitchenTooSlow.
    func prepDeadline(parent context.Context, prep time.Duration) (context.Context, context.CancelFunc) {
    	return context.WithTimeoutCause(parent, prep, ErrKitchenTooSlow)
    }

    func main() {
    	ctx, cancel := prepDeadline(context.Background(), 50*time.Millisecond)
    	defer cancel()

    	select {
    	case <-time.After(200 * time.Millisecond): // the kitchen is slow today
    		fmt.Println("order ready")
    	case <-ctx.Done():
    		fmt.Println("gave up:", ctx.Err(), "/", context.Cause(ctx))
    	}
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

    func TestPrepDeadlineExpires(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		start := time.Now()
    		ctx, cancel := prepDeadline(t.Context(), 20*time.Minute)
    		defer cancel()

    		if dl, ok := ctx.Deadline(); !ok {
    			t.Errorf("prepDeadline(ctx, 20m): ctx.Deadline() reports no deadline, want one 20m from now")
    		} else if !dl.Equal(start.Add(20 * time.Minute)) {
    			t.Errorf("prepDeadline(ctx, 20m): ctx.Deadline() is %v from now, want 20m", dl.Sub(start))
    		}

    		synctest.Sleep(20*time.Minute - time.Nanosecond)
    		if err := ctx.Err(); err != nil {
    			t.Fatalf("1ns before the 20m prep time: ctx.Err() = %v, want nil", err)
    		}

    		synctest.Sleep(time.Nanosecond)
    		if err := ctx.Err(); !errors.Is(err, context.DeadlineExceeded) {
    			t.Errorf("after 20m: ctx.Err() = %v, want context.DeadlineExceeded", err)
    		}
    		if cause := context.Cause(ctx); !errors.Is(cause, ErrKitchenTooSlow) {
    			t.Errorf("after 20m: context.Cause(ctx) = %v, want ErrKitchenTooSlow", cause)
    		}
    	})
    }

    func TestPrepDeadlineCancelledEarly(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		ctx, cancel := prepDeadline(t.Context(), time.Hour)
    		synctest.Sleep(time.Minute)
    		cancel() // the order was ready early
    		if err := ctx.Err(); !errors.Is(err, context.Canceled) {
    			t.Errorf("after cancel(): ctx.Err() = %v, want context.Canceled", err)
    		}
    		if cause := context.Cause(ctx); errors.Is(cause, ErrKitchenTooSlow) {
    			t.Errorf("after cancel() before the deadline: context.Cause(ctx) = %v, want context.Canceled, not ErrKitchenTooSlow", cause)
    		}
    		synctest.Sleep(2 * time.Hour)
    		if cause := context.Cause(ctx); !errors.Is(cause, context.Canceled) {
    			t.Errorf("an hour after cancel(): context.Cause(ctx) = %v, want it to stay context.Canceled", cause)
    		}
    	})
    }

    func TestPrepDeadlineFollowsParent(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		errCustomerLeft := errors.New("customer cancelled the order")
    		parent, cancelParent := context.WithCancelCause(t.Context())
    		ctx, cancel := prepDeadline(parent, time.Hour)
    		defer cancel()

    		synctest.Sleep(time.Minute)
    		cancelParent(errCustomerLeft)
    		if err := ctx.Err(); !errors.Is(err, context.Canceled) {
    			t.Errorf("after the parent was cancelled: ctx.Err() = %v, want context.Canceled", err)
    		}
    		if cause := context.Cause(ctx); !errors.Is(cause, errCustomerLeft) {
    			t.Errorf("after the parent was cancelled: context.Cause(ctx) = %v, want the parent's cause %q (derive ctx from parent)", cause, errCustomerLeft)
    		}
    	})
    }

    func TestPrepDeadlineShorterParent(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		errShiftOver := errors.New("the shift is over")
    		parent, cancelParent := context.WithTimeoutCause(t.Context(), time.Minute, errShiftOver)
    		defer cancelParent()
    		ctx, cancel := prepDeadline(parent, time.Hour)
    		defer cancel()

    		synctest.Sleep(time.Minute)
    		if cause := context.Cause(ctx); !errors.Is(cause, errShiftOver) {
    			t.Errorf("parent expired after 1m: context.Cause(ctx) = %v, want the parent's cause %q", cause, errShiftOver)
    		}
    	})
    }
---

Every Dispatchly order has a **prep time**. If the kitchen hasn't finished by
then, the app stops waiting and offers the customer a refund. When that
happens, the logs should say *why* the order was abandoned, not just
"context deadline exceeded".

Complete `prepDeadline(parent, prep)`. It returns a context derived from
`parent`, and its cancel function, such that:

- The context expires `prep` from now. `ctx.Deadline()` reports that time.
- When it expires, `ctx.Err()` is `context.DeadlineExceeded` and
  `context.Cause(ctx)` is `ErrKitchenTooSlow`.
- Calling the cancel function before then cancels it as usual (`Err()` and
  `Cause` are `context.Canceled`).
- If `parent` is cancelled or expires first, `ctx` is too, and its cause is
  the parent's cause.

## Example

```go
ctx, cancel := prepDeadline(context.Background(), 50*time.Millisecond)
defer cancel()
<-ctx.Done()
fmt.Println(ctx.Err())           // context deadline exceeded
fmt.Println(context.Cause(ctx))  // kitchen took too long
```

## Constraints

- One line of code does it, if you pick the right function from the
  `context` package.
- The tests use `synctest` and check the exact nanosecond the context expires.
