---
title: Retry with Backoff
difficulty: medium
after: context
hints:
  - '`time.Sleep` can''t be interrupted, so a cancelled context would still wait out the whole backoff. Wait with a `select` on a timer and `ctx.Done()` instead.'
  - 'Keep a `wait` variable that starts at `base` and doubles after each wait. Only wait **between** attempts: after the last failure, return right away.'
  - 'To return an error that matches both the cause and the last failure, use `fmt.Errorf("...: %w (last error: %w)", context.Cause(ctx), lastErr)`: `%w` may appear more than once, and `errors.Is` checks all of them. `errors.Join` works too.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"time"
    )

    // retry calls op until it succeeds, at most attempts times, waiting base,
    // 2*base, 4*base, ... between attempts. It gives up early if ctx is done.
    func retry(ctx context.Context, attempts int, base time.Duration, op func(context.Context) error) error {
    	return op(ctx)
    }

    func main() {
    	start := time.Now()
    	calls := 0
    	err := retry(context.Background(), 4, 10*time.Millisecond, func(ctx context.Context) error {
    		calls++
    		fmt.Printf("%4v: charge card, attempt %d\n", time.Since(start).Round(5*time.Millisecond), calls)
    		if calls < 3 {
    			return errors.New("payment gateway busy")
    		}
    		return nil
    	})
    	fmt.Println("result:", err)
    	// want attempts at about 0s, 10ms and 30ms, then "result: <nil>"
    }
  solution: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"time"
    )

    // retry calls op until it succeeds, at most attempts times, waiting base,
    // 2*base, 4*base, ... between attempts. It gives up early if ctx is done.
    func retry(ctx context.Context, attempts int, base time.Duration, op func(context.Context) error) error {
    	if ctx.Err() != nil {
    		return context.Cause(ctx)
    	}
    	var lastErr error
    	wait := base
    	for i := range attempts {
    		if lastErr = op(ctx); lastErr == nil {
    			return nil
    		}
    		if i == attempts-1 {
    			break // no point waiting after the last attempt
    		}
    		timer := time.NewTimer(wait)
    		select {
    		case <-timer.C:
    		case <-ctx.Done():
    			timer.Stop()
    			return fmt.Errorf("retry stopped: %w (last error: %w)", context.Cause(ctx), lastErr)
    		}
    		wait *= 2
    	}
    	return fmt.Errorf("gave up after %d attempts: %w", attempts, lastErr)
    }

    func main() {
    	start := time.Now()
    	calls := 0
    	err := retry(context.Background(), 4, 10*time.Millisecond, func(ctx context.Context) error {
    		calls++
    		fmt.Printf("%4v: charge card, attempt %d\n", time.Since(start).Round(5*time.Millisecond), calls)
    		if calls < 3 {
    			return errors.New("payment gateway busy")
    		}
    		return nil
    	})
    	fmt.Println("result:", err)
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

    // flaky is an op that fails with "failure #n" until call succeedOn
    // (0 = never), recording the fake time of every call.
    type flaky struct {
    	start     time.Time
    	succeedOn int
    	calls     []time.Duration
    	errs      []error
    }

    func newFlaky(succeedOn int) *flaky {
    	return &flaky{start: time.Now(), succeedOn: succeedOn}
    }

    func (f *flaky) op(ctx context.Context) error {
    	f.calls = append(f.calls, time.Since(f.start))
    	if len(f.calls) == f.succeedOn {
    		return nil
    	}
    	err := fmt.Errorf("failure #%d", len(f.calls))
    	f.errs = append(f.errs, err)
    	return err
    }

    func (f *flaky) last() error { return f.errs[len(f.errs)-1] }

    func secs(ds ...int) []time.Duration {
    	var out []time.Duration
    	for _, d := range ds {
    		out = append(out, time.Duration(d)*time.Second)
    	}
    	return out
    }

    func TestRetryBackoffSchedule(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		f := newFlaky(0)
    		err := retry(t.Context(), 4, time.Second, f.op)
    		took := time.Since(f.start)

    		if want := secs(0, 1, 3, 7); !slices.Equal(f.calls, want) {
    			t.Errorf("4 attempts with base 1s: op called at %v, want %v (wait 1s, 2s, 4s between attempts)", f.calls, want)
    		}
    		if took != 7*time.Second {
    			t.Errorf("retry returned after %v, want 7s: don't wait after the last attempt", took)
    		}
    		if len(f.errs) > 0 && !errors.Is(err, f.last()) {
    			t.Errorf("retry returned %v, want an error wrapping the last failure %q", err, f.last())
    		}
    		if len(f.errs) > 1 && errors.Is(err, f.errs[0]) {
    			t.Errorf("retry returned %v, which wraps the first failure; want the last one", err)
    		}
    	})
    }

    type ctxKey struct{}

    func TestRetryPassesContext(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		ctx := context.WithValue(t.Context(), ctxKey{}, "order-42")
    		var got []any
    		retry(ctx, 2, time.Second, func(c context.Context) error {
    			got = append(got, c.Value(ctxKey{}))
    			return errors.New("nope")
    		})
    		if !slices.Equal(got, []any{"order-42", "order-42"}) {
    			t.Errorf("op saw ctx values %v, want [order-42 order-42]: pass ctx (or a context derived from it) to op", got)
    		}
    	})
    }

    func TestRetryStopsOnSuccess(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		f := newFlaky(3)
    		err := retry(t.Context(), 10, 2*time.Second, f.op)
    		if err != nil {
    			t.Errorf("op succeeds on attempt 3: retry returned %v, want nil", err)
    		}
    		if want := secs(0, 2, 6); !slices.Equal(f.calls, want) {
    			t.Errorf("op succeeds on attempt 3 (base 2s): op called at %v, want %v", f.calls, want)
    		}
    		if took := time.Since(f.start); took != 6*time.Second {
    			t.Errorf("retry returned after %v, want 6s", took)
    		}
    	})
    }

    func TestRetryFirstTry(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		f := newFlaky(1)
    		if err := retry(t.Context(), 1, time.Second, f.op); err != nil || len(f.calls) != 1 {
    			t.Errorf("op succeeds at once: retry returned %v after %d calls, want nil after 1", err, len(f.calls))
    		}
    		g := newFlaky(0)
    		err := retry(t.Context(), 1, time.Second, g.op)
    		if len(g.calls) != 1 || !errors.Is(err, g.errs[0]) {
    			t.Errorf("1 attempt that fails: retry returned %v after %d calls, want failure #1 after 1 call", err, len(g.calls))
    		}
    		if took := time.Since(g.start); took != 0 {
    			t.Errorf("1 attempt that fails: retry took %v, want 0s", took)
    		}
    	})
    }

    func TestRetryCancelledDuringBackoff(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		errShutdown := errors.New("dispatcher shutting down")
    		ctx, cancel := context.WithCancelCause(t.Context())
    		time.AfterFunc(2*time.Second, func() { cancel(errShutdown) })

    		f := newFlaky(0)
    		err := retry(ctx, 5, time.Second, f.op)
    		took := time.Since(f.start)

    		if want := secs(0, 1); !slices.Equal(f.calls, want) {
    			t.Errorf("cancelled at 2s: op called at %v, want %v", f.calls, want)
    		}
    		if took != 2*time.Second {
    			t.Errorf("cancelled at 2s, during the 2s wait: retry returned after %v, want exactly 2s (don't time.Sleep; select on ctx.Done())", took)
    		}
    		if !errors.Is(err, errShutdown) {
    			t.Errorf("cancelled with cause %q: retry returned %v, want it to wrap context.Cause(ctx)", errShutdown, err)
    		}
    		if len(f.errs) > 0 && !errors.Is(err, f.last()) {
    			t.Errorf("cancelled after failure #2: retry returned %v, want it to wrap the last failure too", err)
    		}
    	})
    }

    func TestRetryDeadline(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
    		defer cancel()
    		f := newFlaky(0)
    		err := retry(ctx, 10, time.Second, f.op)
    		if took := time.Since(f.start); took != 5*time.Second {
    			t.Errorf("5s timeout: retry returned after %v, want 5s", took)
    		}
    		if !errors.Is(err, context.DeadlineExceeded) {
    			t.Errorf("5s timeout: retry returned %v, want it to wrap context.DeadlineExceeded", err)
    		}
    		if len(f.calls) != 3 {
    			t.Errorf("5s timeout, base 1s: op called %d times (at %v), want 3 (at 0s, 1s, 3s)", len(f.calls), f.calls)
    		}
    	})
    }

    func TestRetryAlreadyCancelled(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		errClosed := errors.New("restaurant closed")
    		ctx, cancel := context.WithCancelCause(t.Context())
    		cancel(errClosed)
    		f := newFlaky(1)
    		err := retry(ctx, 3, time.Second, f.op)
    		if len(f.calls) != 0 {
    			t.Errorf("ctx already cancelled: op called %d times, want 0", len(f.calls))
    		}
    		if !errors.Is(err, errClosed) {
    			t.Errorf("ctx already cancelled with cause %q: retry returned %v, want the cause", errClosed, err)
    		}
    	})
    }
---

Dispatchly's payment gateway sometimes says "busy, try again". Hammering it
makes things worse, so failed charges are retried with **exponential
backoff**: wait a bit, then twice as long, then twice as long again.

Implement `retry(ctx, attempts, base, op)`:

- Call `op(ctx)` up to `attempts` times. As soon as it returns `nil`, return
  `nil`.
- Between attempts wait `base`, then `2*base`, then `4*base`, and so on. Don't
  wait after the last attempt.
- If every attempt fails, return an error that wraps the **last** failure
  (`errors.Is(err, lastFailure)` is true).
- If `ctx` is done while waiting, stop **immediately** and return an error that
  wraps both `context.Cause(ctx)` and the last failure.
- If `ctx` is already done when `retry` is called, don't call `op` at all:
  return `context.Cause(ctx)`.

## Example

With `attempts = 4`, `base = 1s` and an `op` that always fails:

```
t=0s  attempt 1 fails   wait 1s
t=1s  attempt 2 fails   wait 2s
t=3s  attempt 3 fails   wait 4s
t=7s  attempt 4 fails   return "gave up after 4 attempts: failure #4"
```

If the context is cancelled at t=2s, `retry` returns at exactly t=2s with an
error wrapping the cancellation cause and failure #2.

## Constraints

- `attempts` ≥ 1 and `base` > 0. `op` runs in the calling goroutine.
- The tests use `synctest`, so every timing is checked to the nanosecond, and
  a wait that ignores cancellation shows up as returning too late.
