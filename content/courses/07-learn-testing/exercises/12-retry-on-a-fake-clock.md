---
title: Retry on a Fake Clock
difficulty: hard
after: test-doubles
hints:
  - '`FakeClock` only needs a `now time.Time` and a `Sleeps []time.Duration`. `Sleep(d)` appends `d` and moves `now` forward; `Advance(d)` only moves `now`. Nothing ever really waits, so a test of two seconds of retries runs in microseconds.'
  - 'Write one scenario helper and reuse it: given a script of attempts (how long each takes, and what error it returns), make a fresh clock, build an `op` that plays the next step of the script (calling `clock.Advance(step.took)` first), call `retry(clock, op)`, and return the number of calls, `clock.Sleeps` and the error. Compare them with what the policy says, using `slices.Equal` for the sleeps.'
  - 'Cover every rule and both sides of every limit: first-try success; success after failures; failing every time (count the attempts and the capped delays, and check the error with `errors.Is` for **both** `ErrGaveUp` and the op''s error); a permanent error; slow attempts that run out the budget; and an attempt that ends exactly on the budget, where the next sleep is still allowed.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"time"
    )

    // FakeClock is a Clock for tests. Time only moves when Sleep or Advance
    // is called.
    type FakeClock struct {
    	Sleeps []time.Duration // every Sleep call, in order
    }

    // NewFakeClock returns a FakeClock whose Now is start.
    func NewFakeClock(start time.Time) *FakeClock {
    	return &FakeClock{}
    }

    // Now returns the fake current time.
    func (c *FakeClock) Now() time.Time { return time.Time{} }

    // Sleep records d in Sleeps and moves the clock forward by d.
    func (c *FakeClock) Sleep(d time.Duration) {}

    // Advance moves the clock forward by d without recording a sleep, for
    // pretending that some work took time.
    func (c *FakeClock) Advance(d time.Duration) {}

    // checkRetry tests retry against the retry policy using FakeClocks. It
    // returns nil if retry follows the policy, or an error describing the
    // first difference.
    func checkRetry(retry RetryFunc) error {
    	return nil
    }

    // ---- Ledgerly code (already done) ----

    type Clock interface {
    	Now() time.Time
    	Sleep(d time.Duration)
    }

    type RetryFunc func(clock Clock, op func() error) error

    var (
    	ErrPermanent = errors.New("permanent failure")
    	ErrGaveUp    = errors.New("gave up")
    )

    const (
    	maxAttempts = 5
    	firstDelay  = 100 * time.Millisecond
    	maxDelay    = 500 * time.Millisecond
    	budget      = 2 * time.Second
    )

    // Retry calls op until it succeeds, following Ledgerly's retry policy
    // for syncing with the bank (see the rules in the problem statement).
    func Retry(clock Clock, op func() error) error {
    	start := clock.Now()
    	delay := firstDelay
    	for attempt := 1; ; attempt++ {
    		err := op()
    		if err == nil || errors.Is(err, ErrPermanent) {
    			return err
    		}
    		if attempt == maxAttempts || clock.Now().Sub(start)+delay > budget {
    			return fmt.Errorf("%w after %d attempts: %w", ErrGaveUp, attempt, err)
    		}
    		clock.Sleep(delay)
    		delay = min(delay*2, maxDelay)
    	}
    }

    // impatientRetry never waits between attempts.
    func impatientRetry(clock Clock, op func() error) error {
    	var err error
    	for range maxAttempts {
    		if err = op(); err == nil {
    			return nil
    		}
    	}
    	return fmt.Errorf("%w: %w", ErrGaveUp, err)
    }

    func main() {
    	c := NewFakeClock(time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC))
    	c.Sleep(100 * time.Millisecond)
    	c.Advance(time.Second)
    	fmt.Println("clock:", c.Now().Format("15:04:05.000"), "sleeps:", c.Sleeps) // want 09:00:01.100 [100ms]

    	calls := 0
    	c = NewFakeClock(time.Time{})
    	err := Retry(c, func() error {
    		calls++
    		return errors.New("bank timeout")
    	})
    	fmt.Println("Retry of an op that always fails:", calls, "calls, sleeps", c.Sleeps, "->", err)

    	fmt.Println("checkRetry(Retry):         ", checkRetry(Retry))
    	fmt.Println("checkRetry(impatientRetry):", checkRetry(impatientRetry))
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"slices"
    	"time"
    )

    // FakeClock is a Clock for tests. Time only moves when Sleep or Advance
    // is called.
    type FakeClock struct {
    	now    time.Time
    	Sleeps []time.Duration // every Sleep call, in order
    }

    // NewFakeClock returns a FakeClock whose Now is start.
    func NewFakeClock(start time.Time) *FakeClock {
    	return &FakeClock{now: start}
    }

    // Now returns the fake current time.
    func (c *FakeClock) Now() time.Time { return c.now }

    // Sleep records d in Sleeps and moves the clock forward by d.
    func (c *FakeClock) Sleep(d time.Duration) {
    	c.Sleeps = append(c.Sleeps, d)
    	c.now = c.now.Add(d)
    }

    // Advance moves the clock forward by d without recording a sleep, for
    // pretending that some work took time.
    func (c *FakeClock) Advance(d time.Duration) { c.now = c.now.Add(d) }

    // step is one scripted attempt: how long it takes and what it returns.
    type step struct {
    	took time.Duration
    	err  error
    }

    // scenario runs retry against a scripted op. Once the script runs out,
    // its last step repeats.
    func scenario(retry RetryFunc, script ...step) (calls int, sleeps []time.Duration, err error) {
    	clock := NewFakeClock(time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC))
    	err = retry(clock, func() error {
    		s := script[min(calls, len(script)-1)]
    		calls++
    		clock.Advance(s.took)
    		return s.err
    	})
    	return calls, clock.Sleeps, err
    }

    func ms(n ...int) []time.Duration {
    	var d []time.Duration
    	for _, v := range n {
    		d = append(d, time.Duration(v)*time.Millisecond)
    	}
    	return d
    }

    // checkRetry tests retry against the retry policy using FakeClocks. It
    // returns nil if retry follows the policy, or an error describing the
    // first difference.
    func checkRetry(retry RetryFunc) error {
    	flaky := errors.New("bank timeout")
    	denied := fmt.Errorf("bank says no: %w", ErrPermanent)
    	ok := step{}
    	fail := step{err: flaky}

    	tests := []struct {
    		name       string
    		script     []step
    		calls      int
    		sleeps     []time.Duration
    		wantErrs   []error // each must match with errors.Is; nil means success
    		notWantErr error   // must not match, if set
    	}{
    		{"succeeds first time", []step{ok}, 1, nil, nil, nil},
    		{"succeeds on the third attempt", []step{fail, fail, ok}, 3, ms(100, 200), nil, nil},
    		{"always fails", []step{fail}, 5, ms(100, 200, 400, 500), []error{ErrGaveUp, flaky}, nil},
    		{"permanent error", []step{{err: denied}}, 1, nil, []error{ErrPermanent}, ErrGaveUp},
    		{"permanent error after a timeout", []step{fail, {err: denied}}, 2, ms(100), []error{ErrPermanent}, ErrGaveUp},
    		{"slow attempts run out the budget", []step{{400 * time.Millisecond, flaky}}, 4, ms(100, 200, 400), []error{ErrGaveUp, flaky}, nil},
    		{"sleep that ends exactly on the budget", []step{{1900 * time.Millisecond, flaky}, fail}, 2, ms(100), []error{ErrGaveUp, flaky}, nil},
    	}
    	for _, tt := range tests {
    		calls, sleeps, err := scenario(retry, tt.script...)
    		if calls != tt.calls || !slices.Equal(sleeps, tt.sleeps) {
    			return fmt.Errorf("%s: %d calls with sleeps %v, want %d calls with sleeps %v", tt.name, calls, sleeps, tt.calls, tt.sleeps)
    		}
    		if tt.wantErrs == nil && err != nil {
    			return fmt.Errorf("%s: returned %v, want nil", tt.name, err)
    		}
    		for _, want := range tt.wantErrs {
    			if !errors.Is(err, want) {
    				return fmt.Errorf("%s: returned %v, want an error matching %q", tt.name, err, want)
    			}
    		}
    		if tt.notWantErr != nil && errors.Is(err, tt.notWantErr) {
    			return fmt.Errorf("%s: returned %v, which shouldn't match %q", tt.name, err, tt.notWantErr)
    		}
    	}
    	return nil
    }

    // ---- Ledgerly code (already done) ----

    type Clock interface {
    	Now() time.Time
    	Sleep(d time.Duration)
    }

    type RetryFunc func(clock Clock, op func() error) error

    var (
    	ErrPermanent = errors.New("permanent failure")
    	ErrGaveUp    = errors.New("gave up")
    )

    const (
    	maxAttempts = 5
    	firstDelay  = 100 * time.Millisecond
    	maxDelay    = 500 * time.Millisecond
    	budget      = 2 * time.Second
    )

    // Retry calls op until it succeeds, following Ledgerly's retry policy
    // for syncing with the bank (see the rules in the problem statement).
    func Retry(clock Clock, op func() error) error {
    	start := clock.Now()
    	delay := firstDelay
    	for attempt := 1; ; attempt++ {
    		err := op()
    		if err == nil || errors.Is(err, ErrPermanent) {
    			return err
    		}
    		if attempt == maxAttempts || clock.Now().Sub(start)+delay > budget {
    			return fmt.Errorf("%w after %d attempts: %w", ErrGaveUp, attempt, err)
    		}
    		clock.Sleep(delay)
    		delay = min(delay*2, maxDelay)
    	}
    }

    // impatientRetry never waits between attempts.
    func impatientRetry(clock Clock, op func() error) error {
    	var err error
    	for range maxAttempts {
    		if err = op(); err == nil {
    			return nil
    		}
    	}
    	return fmt.Errorf("%w: %w", ErrGaveUp, err)
    }

    func main() {
    	c := NewFakeClock(time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC))
    	c.Sleep(100 * time.Millisecond)
    	c.Advance(time.Second)
    	fmt.Println("clock:", c.Now().Format("15:04:05.000"), "sleeps:", c.Sleeps) // want 09:00:01.100 [100ms]

    	calls := 0
    	c = NewFakeClock(time.Time{})
    	err := Retry(c, func() error {
    		calls++
    		return errors.New("bank timeout")
    	})
    	fmt.Println("Retry of an op that always fails:", calls, "calls, sleeps", c.Sleeps, "->", err)

    	fmt.Println("checkRetry(Retry):         ", checkRetry(Retry))
    	fmt.Println("checkRetry(impatientRetry):", checkRetry(impatientRetry))
    }
  tests: |
    package main

    import (
    	"errors"
    	"fmt"
    	"slices"
    	"testing"
    	"time"
    )

    func TestFakeClock(t *testing.T) {
    	start := time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)
    	c := NewFakeClock(start)
    	if !c.Now().Equal(start) {
    		t.Fatalf("NewFakeClock(%v).Now() = %v, want the start time", start, c.Now())
    	}
    	var clock Clock = c
    	clock.Sleep(100 * time.Millisecond)
    	c.Advance(2 * time.Second)
    	clock.Sleep(300 * time.Millisecond)
    	if want := start.Add(2400 * time.Millisecond); !c.Now().Equal(want) {
    		t.Errorf("after Sleep(100ms), Advance(2s), Sleep(300ms): Now() = %v, want %v", c.Now(), want)
    	}
    	if want := []time.Duration{100 * time.Millisecond, 300 * time.Millisecond}; !slices.Equal(c.Sleeps, want) {
    		t.Errorf("after Sleep(100ms), Advance(2s), Sleep(300ms): Sleeps = %v, want %v (Advance isn't a sleep)", c.Sleeps, want)
    	}
    	other := NewFakeClock(start)
    	if len(other.Sleeps) != 0 || !other.Now().Equal(start) {
    		t.Error("two FakeClocks share state")
    	}
    }

    // policy is a configurable retry; the zero value plus defaults() is correct.
    type policy struct {
    	attempts       int
    	maxDelay       time.Duration
    	retryPermanent bool
    	sleepAfterLast bool
    	ignoreBudget   bool
    	strictBudget   bool
    	wrap           string // "both", "gaveup", "op"
    	keepOldErr     bool
    }

    func defaults() policy {
    	return policy{attempts: 5, maxDelay: 500 * time.Millisecond, wrap: "both"}
    }

    func (p policy) retry(clock Clock, op func() error) error {
    	start := clock.Now()
    	delay := 100 * time.Millisecond
    	var last error
    	for attempt := 1; ; attempt++ {
    		err := op()
    		if err == nil {
    			if p.keepOldErr {
    				return last
    			}
    			return nil
    		}
    		last = err
    		if errors.Is(err, ErrPermanent) && !p.retryPermanent {
    			return err
    		}
    		elapsed := clock.Now().Sub(start) + delay
    		over := elapsed > 2*time.Second
    		if p.strictBudget {
    			over = elapsed >= 2*time.Second
    		}
    		if p.ignoreBudget {
    			over = false
    		}
    		if attempt >= p.attempts || over {
    			if p.sleepAfterLast {
    				clock.Sleep(delay)
    			}
    			switch p.wrap {
    			case "gaveup":
    				return fmt.Errorf("%w after %d attempts: %v", ErrGaveUp, attempt, err)
    			case "op":
    				return fmt.Errorf("after %d attempts: %w", attempt, err)
    			}
    			return fmt.Errorf("%w after %d attempts: %w", ErrGaveUp, attempt, err)
    		}
    		clock.Sleep(delay)
    		delay = min(delay*2, p.maxDelay)
    	}
    }

    func runCheck(t *testing.T, retry RetryFunc) (err error, panicked string) {
    	t.Helper()
    	done := make(chan struct{})
    	go func() {
    		defer close(done)
    		defer func() {
    			if p := recover(); p != nil {
    				panicked = fmt.Sprint(p)
    			}
    		}()
    		err = checkRetry(retry)
    	}()
    	select {
    	case <-done:
    	case <-time.After(2 * time.Second):
    		t.Fatal("checkRetry took over 2 seconds: use FakeClocks, never real time")
    	}
    	return err, panicked
    }

    func TestCheckAcceptsCorrectRetries(t *testing.T) {
    	for name, retry := range map[string]RetryFunc{"Retry": Retry, "another correct retry": defaults().retry} {
    		if err, p := runCheck(t, retry); p != "" {
    			t.Errorf("checkRetry(%s) panicked: %s", name, p)
    		} else if err != nil {
    			t.Errorf("checkRetry(%s) = %v, want nil: it follows the policy", name, err)
    		}
    	}
    }

    func TestCheckCatchesBugs(t *testing.T) {
    	with := func(f func(*policy)) RetryFunc {
    		p := defaults()
    		f(&p)
    		return p.retry
    	}
    	bugs := []struct {
    		bug   string
    		retry RetryFunc
    	}{
    		{"makes 4 attempts instead of 5", with(func(p *policy) { p.attempts = 4 })},
    		{"makes 6 attempts instead of 5", with(func(p *policy) { p.attempts = 6 })},
    		{"doesn't cap the delay at 500ms", with(func(p *policy) { p.maxDelay = time.Hour })},
    		{"retries permanent errors", with(func(p *policy) { p.retryPermanent = true })},
    		{"sleeps once more after the last attempt", with(func(p *policy) { p.sleepAfterLast = true })},
    		{"ignores the 2s budget", with(func(p *policy) { p.ignoreBudget = true })},
    		{"gives up when a sleep would end exactly on the budget", with(func(p *policy) { p.strictBudget = true })},
    		{"doesn't wrap the op's error when giving up", with(func(p *policy) { p.wrap = "gaveup" })},
    		{"doesn't wrap ErrGaveUp when giving up", with(func(p *policy) { p.wrap = "op" })},
    		{"returns the previous error after a later success", with(func(p *policy) { p.keepOldErr = true })},
    		{"never waits between attempts", impatientRetry},
    	}
    	for _, b := range bugs {
    		if err, p := runCheck(t, b.retry); p != "" {
    			t.Errorf("checkRetry panicked on a retry that %s: %s", b.bug, p)
    		} else if err == nil {
    			t.Errorf("checkRetry returned nil for a retry that %s", b.bug)
    		}
    	}
    }
---

When Ledgerly syncs with the bank, the bank sometimes times out. The fix is
a retry loop with **exponential backoff**, and retry loops are notoriously
hard to test: the real one waits for seconds, and its behaviour depends on
how long each attempt took. The answer is to inject the clock, and in tests
use a fake one where sleeping is instant and time only moves when you say so.

This problem has two parts.

## Part 1: FakeClock

Implement `FakeClock`, which satisfies `Clock`:

- `NewFakeClock(start)` returns a clock whose `Now()` is `start`.
- `Sleep(d)` appends `d` to `Sleeps` and moves the clock forward by `d`. It
  returns immediately.
- `Advance(d)` moves the clock forward by `d` **without** recording a sleep.
  Tests use it to pretend an attempt took a while.

## Part 2: checkRetry

Write `checkRetry(retry)`, which tests any `RetryFunc` against the policy
below using FakeClocks, and returns `nil` if it complies or an error
describing the first difference. `retry(clock, op)` calls `op` until it
succeeds or the policy says stop:

1. If `op` returns `nil`, return `nil` straight away.
2. If `op` returns an error wrapping `ErrPermanent`, return that error
   straight away, without retrying.
3. Otherwise, sleep and try again. The delays are **100ms, 200ms, 400ms**,
   doubling each time but **capped at 500ms**.
4. Give up after **5 attempts**, or when sleeping the next delay would take
   the total time since the first call **past 2s**. (Ending exactly at
   2s is fine.) Don't sleep before giving up.
5. When giving up, return an error that wraps **both** `ErrGaveUp` and the
   last error from `op`, so `errors.Is` matches either.

## Example

```text
op fails every time, instantly:
    5 calls, sleeps [100ms 200ms 400ms 500ms], error "gave up after 5 attempts: bank timeout"
op takes 400ms and fails every time:
    4 calls, sleeps [100ms 200ms 400ms]: after 4 attempts it's 2.3s in, so it stops
```

## How you're graded

Your `FakeClock` is tested on its own. Then your `checkRetry` must return
`nil` for `Retry` and for a second correct implementation, and an error for
**eleven** broken retries: wrong attempt counts, a missing cap, retried
permanent errors, a sleep after the last attempt, an ignored budget, an
off-by-one at the budget's edge, half-wrapped errors, and more. It must not
panic, and it must finish in well under a second, which rules out real
sleeping.

**Run** tries your clock, shows `Retry` on a failing op, and runs your
`checkRetry` on `Retry` and on `impatientRetry`, which never waits.

## Constraints

- Only interact with `retry` through its arguments: a clock you create, and
  an `op` you write.
- Use a fresh `FakeClock` for each scenario so sleeps don't pile up across
  scenarios.
