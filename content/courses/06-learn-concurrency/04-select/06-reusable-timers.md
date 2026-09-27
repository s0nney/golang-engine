---
title: Reusable Timers and Callbacks
quiz:
  - question: |
      What does this print?

      ```go
      t := time.NewTimer(time.Hour)
      fmt.Println(t.Stop(), t.Stop())
      ```
    options:
      - text: '`true true`'
      - text: '`true false`'
        correct: true
      - text: '`false false`'
      - text: It blocks for an hour
    explanation: |
      `Stop` reports whether *this call* stopped an active timer. The first
      call stops the pending hour-long timer and returns `true`. By the
      second call the timer is already stopped, so there's nothing to stop
      and it returns `false`.
  - question: '`timer := time.AfterFunc(d, refresh)`. Later, `timer.Stop()` returns `false` because `refresh` has already started. What does `Stop` do about the running `refresh`?'
    options:
      - text: It waits for `refresh` to return
      - text: It cancels `refresh` in the middle of what it's doing
      - text: Nothing. `refresh` keeps running in its own goroutine, and you have to coordinate with it yourself if you need to know when it's done
        correct: true
      - text: It panics, because the callback has already started
    explanation: |
      `Stop` can only prevent a callback that hasn't started yet. Once
      `f` is running in its own goroutine, `Stop` returns `false` straight
      away. If you need to wait for it, have the callback close a channel
      (or call `wg.Done`) when it finishes.
  - question: In a module with `go 1.27` in its `go.mod`, you call `timer.Reset(time.Minute)` on a timer that may have fired a moment ago. Can the next receive from `timer.C` deliver that old expiry?
    options:
      - text: Yes, so you must call `Stop` and drain `timer.C` before every `Reset`
      - text: No. Since Go 1.23, a receive after `Reset` never sees a value from the previous setting
        correct: true
      - text: Only if the channel is buffered
      - text: '`Reset` panics on a timer that has already fired'
    explanation: |
      Go 1.23 made timer channels unbuffered (synchronous), so `Reset` and
      `Stop` can guarantee no stale value is delivered afterwards. The old
      "stop and drain" dance you'll see in older code was needed before
      that, and isn't needed in new code.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"time"
    )

    var ErrIdle = errors.New("courier went quiet")

    // watchUpdates counts location updates from in. It returns the count and
    // a nil error once in is closed, or the count and ErrIdle as soon as no
    // update has arrived for `idle`.
    func watchUpdates(in <-chan string, idle time.Duration) (int, error) {
    	timer := time.NewTimer(idle)
    	defer timer.Stop()
    	count := 0
    	for {
    		select {
    		case _, ok := <-in:
    			if !ok {
    				return count, nil
    			}
    			count++
    		case <-timer.C:
    			return count, ErrIdle
    		}
    	}
    }

    func main() {
    	in := make(chan string)
    	go func() {
    		defer close(in)
    		for _, u := range []string{"at pickup", "leaving the restaurant", "2 blocks away", "at dropoff"} {
    			time.Sleep(30 * time.Millisecond)
    			in <- "ana " + u
    		}
    	}()

    	count, err := watchUpdates(in, 50*time.Millisecond)
    	fmt.Println("updates:", count, "error:", err)
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"time"
    )

    var ErrIdle = errors.New("courier went quiet")

    // watchUpdates counts location updates from in. It returns the count and
    // a nil error once in is closed, or the count and ErrIdle as soon as no
    // update has arrived for `idle`.
    func watchUpdates(in <-chan string, idle time.Duration) (int, error) {
    	timer := time.NewTimer(idle)
    	defer timer.Stop()
    	count := 0
    	for {
    		select {
    		case _, ok := <-in:
    			if !ok {
    				return count, nil
    			}
    			count++
    			timer.Reset(idle) // a fresh idle period starts now
    		case <-timer.C:
    			return count, ErrIdle
    		}
    	}
    }

    func main() {
    	in := make(chan string)
    	go func() {
    		defer close(in)
    		for _, u := range []string{"at pickup", "leaving the restaurant", "2 blocks away", "at dropoff"} {
    			time.Sleep(30 * time.Millisecond)
    			in <- "ana " + u
    		}
    	}()

    	count, err := watchUpdates(in, 50*time.Millisecond)
    	fmt.Println("updates:", count, "error:", err)
    }
  tests: |
    package main

    import (
    	"errors"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    func TestClosedInput(t *testing.T) {
    	for _, n := range []int{0, 1, 7} {
    		in := make(chan string, n)
    		for range n {
    			in <- "update"
    		}
    		close(in)
    		got, err := watchUpdates(in, time.Hour)
    		if got != n || err != nil {
    			t.Errorf("closed input with %d queued updates: got %d, %v; want %d, <nil>", n, got, err, n)
    		}
    	}
    }

    func TestIdleInput(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		start := time.Now()
    		got, err := watchUpdates(make(chan string), time.Minute)
    		if got != 0 || !errors.Is(err, ErrIdle) {
    			t.Errorf("no updates at all: got %d, %v; want 0, ErrIdle", got, err)
    		}
    		if took := time.Since(start); took != time.Minute {
    			t.Errorf("no updates at all: gave up after %v, want 1m0s", took)
    		}
    	})
    }

    func TestUpdatesExtendDeadline(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		in := make(chan string, 3) // buffered, so the sender never gets stuck
    		go func() {
    			for range 3 {
    				time.Sleep(40 * time.Second)
    				in <- "update"
    			}
    		}()
    		start := time.Now()
    		got, err := watchUpdates(in, time.Minute)
    		took := time.Since(start)
    		if got != 3 || !errors.Is(err, ErrIdle) {
    			t.Errorf("updates every 40s with a 1m idle limit: got %d, %v; want 3, ErrIdle", got, err)
    		}
    		if took != 3*time.Minute {
    			t.Errorf("gave up after %v, want 3m0s: 1m after the last update at 2m0s (reset the timer after each update)", took)
    		}
    		time.Sleep(time.Hour) // let the sender finish
    	})
    }
---

`time.After` is perfect for a one-off timeout. But some deadlines *move*. Dispatchly flags a courier as "gone quiet" when their phone hasn't sent a location update for a minute, and every update pushes that deadline back. For that you want a single **reusable timer** that you reset, not a new one for every event.

## NewTimer, Reset and Stop

`time.NewTimer(d)` returns a `*time.Timer`. Its channel `C` receives the time once, after `d`. Two methods change its plans:

- **`Reset(d)`** reschedules it to fire `d` from now, whether or not it has already fired.
- **`Stop()`** cancels a pending expiry. It returns `true` if it stopped an active timer, `false` if the timer had already fired or been stopped.

Neither method closes `C`, so never `range` over a timer's channel, and don't leave a goroutine whose *only* way out is receiving from `C` of a timer someone might stop.

Here Dispatchly rings a courier who isn't answering, doubling the wait between attempts, with one timer for the whole conversation:

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	answered := make(chan string)
	go func() {
		time.Sleep(100 * time.Millisecond) // ana finally picks up
		answered <- "ana"
	}()

	delay := 10 * time.Millisecond
	timer := time.NewTimer(delay)
	defer timer.Stop()
	for {
		select {
		case who := <-answered:
			fmt.Println(who, "answered")
			return
		case <-timer.C:
			delay *= 2
			fmt.Println("no answer, ringing again in", delay)
			timer.Reset(delay)
		}
	}
}
```

```text
no answer, ringing again in 20ms
no answer, ringing again in 40ms
no answer, ringing again in 80ms
ana answered
```

You could write this with a fresh `time.After(delay)` on every iteration, and since Go 1.23 that wouldn't leak anything. A single timer is still tidier: it states "there is exactly one pending deadline", and it avoids allocating a new timer for every event in a busy loop.

## No more stop-and-drain

In older Go code you'll see this ritual before every `Reset`:

```go
if !t.Stop() {
	<-t.C // drain a value that may already be sitting in the channel
}
t.Reset(d)
```

Before Go 1.23, a timer that had fired left its value sitting in a buffered channel, and a later receive could pick up that *stale* expiry. Since Go 1.23 (for modules whose `go.mod` says `go 1.23` or later) timer channels are synchronous: after `Reset` or `Stop` returns, no value from the old setting will ever be delivered. Just call `Reset`. Don't copy the drain ritual into new code: on a modern timer, that `<-t.C` can block forever.

## Callbacks with time.AfterFunc

`time.AfterFunc(d, f)` doesn't give you a channel at all. After `d` it calls `f` **in its own goroutine**. The returned timer's `C` is `nil`, but `Stop` and `Reset` work.

The catch is what `Stop` can and can't do. If it returns `true`, `f` will never run. If it returns `false`, `f` has already started, and `Stop` **doesn't wait for it**. If you need to know that the callback has finished, it has to tell you:

```go
done := make(chan struct{})
timer := time.AfterFunc(time.Second, func() {
	defer close(done)
	refreshStatusBoard()
})

// ... later, during shutdown:
if !timer.Stop() {
	<-done // the callback already started: wait for it to finish
}
```

This only works for a one-shot callback. Calling `Reset` on an `AfterFunc` timer after it has fired schedules `f` to run *again*, possibly while the previous run is still going. Anything the callback touches must then be safe for concurrent use.

(There's also `context.AfterFunc`, which runs a function when a *context* is cancelled rather than after a delay. You'll meet it in the next chapter.)

## Your turn

`watchUpdates` counts a courier's location updates. It should return `ErrIdle` as soon as **no update has arrived for `idle`**, or a nil error once the `in` channel is closed. You can assume `idle > 0`.

The starter creates its timer once and never touches it again, so the courier is declared "gone quiet" `idle` after the watch *started*, no matter how many updates arrive. Press **Run** to see ana get cut off after one update. Fix it so every update starts a fresh idle period.

## Further reading

- [`time.Timer` documentation](https://pkg.go.dev/time#Timer)
