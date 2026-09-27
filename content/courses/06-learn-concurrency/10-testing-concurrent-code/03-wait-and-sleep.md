---
title: synctest.Wait and synctest.Sleep
quiz:
  - question: What does `synctest.Wait()` do?
    options:
      - text: Sleeps for a short fixed time so goroutines can catch up
      - text: Blocks until every *other* goroutine in the bubble is durably blocked, meaning they've done everything they can
        correct: true
      - text: Waits for every goroutine in the bubble to exit
      - text: Advances the fake clock to the next timer
    explanation: |
      `Wait` returns once everything else in the bubble is stuck waiting on
      something, so any work they could do in response to your last action
      has been done. It doesn't advance time or require goroutines to exit.
  - question: 'A goroutine''s timer and the test''s `time.Sleep` both end at exactly t=30s. Why use `synctest.Sleep(30*time.Second)` instead of `time.Sleep`?'
    options:
      - text: '`time.Sleep` doesn''t work inside a bubble'
      - text: '`synctest.Sleep` is faster'
      - text: With `time.Sleep`, either goroutine might run first at t=30s; `synctest.Sleep` also waits until the others have finished reacting, so the test sees the settled state
        correct: true
      - text: '`synctest.Sleep` stops the fake clock'
    explanation: |
      When two goroutines wake at the same fake instant, their order is
      unspecified. `synctest.Sleep(d)` is `time.Sleep(d)` followed by
      `synctest.Wait()`, so the test only checks once the code under test
      has handled everything due at that moment.
exercise:
  starter: |
    package main

    import (
    	"context"
    	"fmt"
    	"time"
    )

    // debounce forwards values from in to the returned channel, but only once
    // things have gone quiet: after each value it waits for `quiet`, and if
    // another value arrives in the meantime, it starts waiting again with the
    // newer value. Only the latest value of a burst is sent.
    //
    // When in is closed, any pending value is sent straight away and the
    // returned channel is closed. When ctx is cancelled, any pending value is
    // dropped and the returned channel is closed.
    func debounce(ctx context.Context, in <-chan string, quiet time.Duration) <-chan string {
    	out := make(chan string)
    	go func() {
    		defer close(out)
    		for v := range in {
    			out <- v
    		}
    	}()
    	return out
    }

    func main() {
    	pings := make(chan string)
    	go func() {
    		defer close(pings)
    		for _, loc := range []string{"5th & Main", "6th & Main", "7th & Main"} {
    			pings <- loc // a burst of GPS pings
    			time.Sleep(5 * time.Millisecond)
    		}
    		time.Sleep(60 * time.Millisecond)
    		pings <- "at the restaurant"
    		time.Sleep(60 * time.Millisecond)
    	}()

    	start := time.Now()
    	for loc := range debounce(context.Background(), pings, 30*time.Millisecond) {
    		fmt.Printf("%6v: map shows %s\n", time.Since(start).Round(10*time.Millisecond), loc)
    	}
    }
  solution: |
    package main

    import (
    	"context"
    	"fmt"
    	"time"
    )

    // debounce forwards values from in to the returned channel, but only once
    // things have gone quiet: after each value it waits for `quiet`, and if
    // another value arrives in the meantime, it starts waiting again with the
    // newer value. Only the latest value of a burst is sent.
    //
    // When in is closed, any pending value is sent straight away and the
    // returned channel is closed. When ctx is cancelled, any pending value is
    // dropped and the returned channel is closed.
    func debounce(ctx context.Context, in <-chan string, quiet time.Duration) <-chan string {
    	out := make(chan string)
    	go func() {
    		defer close(out)

    		var latest string
    		var timer <-chan time.Time // nil while nothing is pending
    		for {
    			select {
    			case v, ok := <-in:
    				if !ok {
    					if timer != nil { // flush the pending value
    						select {
    						case out <- latest:
    						case <-ctx.Done():
    						}
    					}
    					return
    				}
    				latest = v
    				timer = time.After(quiet) // (re)start the quiet period
    			case <-timer:
    				select {
    				case out <- latest:
    				case <-ctx.Done():
    					return
    				}
    				timer = nil
    			case <-ctx.Done():
    				return
    			}
    		}
    	}()
    	return out
    }

    func main() {
    	pings := make(chan string)
    	go func() {
    		defer close(pings)
    		for _, loc := range []string{"5th & Main", "6th & Main", "7th & Main"} {
    			pings <- loc // a burst of GPS pings
    			time.Sleep(5 * time.Millisecond)
    		}
    		time.Sleep(60 * time.Millisecond)
    		pings <- "at the restaurant"
    		time.Sleep(60 * time.Millisecond)
    	}()

    	start := time.Now()
    	for loc := range debounce(context.Background(), pings, 30*time.Millisecond) {
    		fmt.Printf("%6v: map shows %s\n", time.Since(start).Round(10*time.Millisecond), loc)
    	}
    }
  tests: |
    package main

    import (
    	"context"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // harness wires up debounce inside a synctest bubble.
    type harness struct {
    	t      *testing.T
    	in     chan string
    	out    <-chan string
    	cancel context.CancelFunc
    	closed bool
    }

    func start(t *testing.T, quiet time.Duration) *harness {
    	ctx, cancel := context.WithCancel(t.Context())
    	h := &harness{t: t, in: make(chan string), cancel: cancel}
    	h.out = debounce(ctx, h.in, quiet)
    	return h
    }

    // stop shuts everything down so no goroutine is left behind, whatever
    // debounce did.
    func (h *harness) stop() {
    	h.cancel()
    	go func() {
    		for range h.out {
    		}
    	}()
    	if !h.closed {
    		close(h.in)
    	}
    	synctest.Wait()
    }

    func (h *harness) send(v string) {
    	h.t.Helper()
    	select {
    	case h.in <- v:
    	case <-time.After(time.Hour):
    		h.t.Fatalf("debounce didn't accept %q within an hour: it must keep reading its input while waiting (is it stuck sending on its output?)", v)
    	}
    }

    // expect checks what is waiting on the output right now, without waiting.
    func (h *harness) expect(want string) {
    	h.t.Helper()
    	select {
    	case got, ok := <-h.out:
    		switch {
    		case !ok && want != "":
    			h.t.Errorf("at %v: output closed, want %q", elapsed(), want)
    		case ok && want == "":
    			h.t.Errorf("at %v: debounce sent %q, want nothing yet", elapsed(), got)
    		case ok && got != want:
    			h.t.Errorf("at %v: debounce sent %q, want %q", elapsed(), got, want)
    		}
    	default:
    		if want != "" {
    			h.t.Errorf("at %v: debounce sent nothing, want %q", elapsed(), want)
    		}
    	}
    }

    var epoch = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

    func elapsed() time.Duration { return time.Since(epoch) }

    func TestDebounceSendsLatestAfterQuiet(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		h := start(t, 5*time.Second)
    		defer h.stop()

    		h.send("a")
    		synctest.Sleep(time.Second)
    		h.send("b")
    		synctest.Sleep(time.Second)
    		h.send("c") // t=2s: the quiet period restarts from here

    		synctest.Sleep(4 * time.Second) // t=6s
    		h.expect("")
    		synctest.Sleep(time.Second) // t=7s: 5s after "c"
    		h.expect("c")
    		synctest.Sleep(time.Minute)
    		h.expect("") // nothing else was sent

    		h.send("d") // a new, separate burst
    		synctest.Sleep(5 * time.Second)
    		h.expect("d")
    	})
    }

    func TestDebounceFlushesOnClose(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		h := start(t, 5*time.Second)
    		defer h.stop()

    		h.send("x")
    		synctest.Sleep(time.Second)
    		close(h.in)
    		h.closed = true
    		synctest.Wait()
    		h.expect("x") // sent immediately, without waiting out the quiet period

    		select {
    		case _, ok := <-h.out:
    			if ok {
    				t.Errorf("debounce sent a second value after its input closed")
    			}
    		case <-time.After(time.Minute):
    			t.Errorf("debounce's output was not closed after its input closed")
    		}
    	})
    }

    func TestDebounceStopsOnCancel(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		h := start(t, 5*time.Second)
    		defer h.stop()

    		h.send("y")
    		synctest.Sleep(time.Second)
    		h.cancel()

    		select {
    		case v, ok := <-h.out:
    			if ok {
    				t.Errorf("after cancel, debounce sent %q, want the pending value dropped and the output closed", v)
    			}
    		case <-time.After(time.Minute):
    			t.Errorf("debounce's output was not closed after ctx was cancelled")
    		}
    	})
    }
---

Fake time solves "waiting is slow". The other half of flakiness is "did the goroutine get to it yet?" That's what `synctest.Wait`, and its Go 1.27 companion `synctest.Sleep`, are for.

## synctest.Wait

```go
func Wait()
```

`Wait` blocks until every *other* goroutine in the bubble is durably blocked. At that point they've done everything they can in response to whatever you just did, so it's safe to look at the results. It replaces every "sleep a little and hope" in your tests.

Here's a Dispatchly `Tracker` that marks a courier offline when their phone stops pinging:

```go
type Tracker struct {
	pings   chan struct{}
	timeout time.Duration
	online  atomic.Bool
}

func (tr *Tracker) Ping()        { tr.pings <- struct{}{} }
func (tr *Tracker) Online() bool { return tr.online.Load() }

// Run processes pings until ctx is cancelled.
func (tr *Tracker) Run(ctx context.Context) {
	var expired <-chan time.Time
	for {
		select {
		case <-tr.pings:
			tr.online.Store(true)
			expired = time.After(tr.timeout)
		case <-expired:
			tr.online.Store(false)
		case <-ctx.Done():
			return
		}
	}
}
```

After `tr.Ping()` returns, `Run` has *received* the ping, but may not have run `online.Store(true)` yet. `synctest.Wait()` waits until `Run` is back in its `select`, blocked, so the store has definitely happened:

```go
tr.Ping()
synctest.Wait() // let Run handle the ping
if !tr.Online() {
	t.Fatal("courier offline right after a ping")
}
```

## The tie problem

Now test the timeout. The obvious version:

```go
tr.Ping()
synctest.Wait()
time.Sleep(30 * time.Second) // no Wait: who runs first at t=30s?
if tr.Online() {
	t.Fatal("courier still online 30s after the last ping")
}
```

This test is **flaky**. Running it 200 times failed exactly 100 of them. At the fake instant t=30s, two goroutines wake up at once: the test, whose sleep ended, and `Run`, whose timer fired. Which runs first is up to the scheduler. Half the time the test checks `Online()` before `Run` has marked the courier offline.

The fix is to sleep and then wait for everything else to settle: `time.Sleep(d)` followed by `synctest.Wait()`. That pair is so common that **Go 1.27 added `synctest.Sleep`**, which does exactly that:

```go
func TestTrackerGoesOffline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		tr := NewTracker(30 * time.Second)
		go tr.Run(t.Context())

		tr.Ping()
		synctest.Wait() // let Run handle the ping
		if !tr.Online() {
			t.Fatal("courier offline right after a ping")
		}

		synctest.Sleep(29 * time.Second)
		if !tr.Online() {
			t.Fatal("courier went offline before the 30s timeout")
		}

		synctest.Sleep(time.Second) // exactly 30s after the ping
		if tr.Online() {
			t.Fatal("courier still online 30s after the last ping")
		}
	})
}
```

This passes every time. It also checks the boundary precisely: still online at 29s, offline at exactly 30s. That kind of test is almost impossible with real time.

A good rule: **in a synctest test, use `synctest.Sleep` wherever you'd use `time.Sleep`** to move time forward before checking something.

## Deterministic, not just fast

Look back at the exercises in this course. Many of their tests check things like "`collectBids` returned after exactly 1m0s" or "10 orders with 3 workers took exactly 4s". That precision is only possible because fake time is exact and `Wait` removes scheduling guesswork. Tests like these don't flake, don't need retries, and run in milliseconds.

## Your turn

When a courier is moving, their phone sends a burst of GPS pings. Redrawing the customer's map for each one is wasteful, so Dispatchly **debounces** them: after a ping, wait until the phone has been quiet for a while, then send only the latest location.

Implement `debounce(ctx, in, quiet)`:

- After each value from `in`, wait for `quiet`. If another value arrives first, start waiting again with the newer value. When the wait completes, send the latest value on the output.
- Keep reading `in` while waiting. Never block the sender just because a value is pending.
- When `in` is closed, send any pending value **immediately**, then close the output.
- When `ctx` is cancelled, drop any pending value and close the output.

The starter just forwards every value as it arrives. The tests use `synctest` and `synctest.Sleep` to check the exact moment each value comes out. Hint: the nil-channel trick is handy for a timer that's only sometimes active.
