---
title: Timeouts with time.After
quiz:
  - question: |
      Bids arrive every 2 seconds, forever. How long does this loop run?

      ```go
      for {
      	select {
      	case b := <-bids:
      		record(b)
      	case <-time.After(5 * time.Second):
      		return
      	}
      }
      ```
    options:
      - text: About 5 seconds
      - text: About 7 seconds
      - text: Forever
        correct: true
      - text: It returns immediately
    explanation: |
      `time.After` is called again on every iteration, creating a fresh
      5-second timer each time. A bid arrives after 2 seconds, which
      restarts the countdown, so the timeout never fires. That's an *idle*
      timeout. For a total time limit, create the timer once, before the loop.
  - question: 'Since Go 1.23, what happens to the timer behind a `time.After` call whose `select` picked a different case?'
    options:
      - text: It leaks until it fires, so you should use `time.NewTimer` and `Stop` instead
      - text: It's garbage-collected once nothing refers to its channel, even if it hasn't fired
        correct: true
      - text: It fires into the next `select` statement by mistake
      - text: It panics when it fires with no receiver
    explanation: |
      Before Go 1.23 an unfired `time.After` timer stayed in memory until it
      fired, and older code avoided it in hot loops. Since Go 1.23 unreferenced
      timers are collected, and the `time` docs say there's no reason to
      prefer `NewTimer` when `After` will do.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"time"
    )

    type Bid struct {
    	Courier string
    	Cents   int
    }

    // collectBids gathers bids until the window (measured from when
    // collectBids is called) runs out, or until bids is closed.
    func collectBids(bids <-chan Bid, window time.Duration) []Bid {
    	var got []Bid
    	for {
    		select {
    		case b, ok := <-bids:
    			if !ok {
    				return got
    			}
    			got = append(got, b)
    		case <-time.After(window):
    			return got
    		}
    	}
    }

    func main() {
    	bids := make(chan Bid, 10)
    	go func() {
    		for _, c := range []string{"ana", "ben", "cy", "dee", "eli"} {
    			time.Sleep(30 * time.Millisecond) // couriers trickle in
    			bids <- Bid{Courier: c, Cents: 400 + len(c)*50}
    		}
    	}()

    	start := time.Now()
    	got := collectBids(bids, 100*time.Millisecond)
    	fmt.Println("bids:", got)
    	fmt.Println("window took", time.Since(start).Round(50*time.Millisecond))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"time"
    )

    type Bid struct {
    	Courier string
    	Cents   int
    }

    // collectBids gathers bids until the window (measured from when
    // collectBids is called) runs out, or until bids is closed.
    func collectBids(bids <-chan Bid, window time.Duration) []Bid {
    	var got []Bid
    	deadline := time.After(window) // one timer for the whole window
    	for {
    		select {
    		case b, ok := <-bids:
    			if !ok {
    				return got
    			}
    			got = append(got, b)
    		case <-deadline:
    			return got
    		}
    	}
    }

    func main() {
    	bids := make(chan Bid, 10)
    	go func() {
    		for _, c := range []string{"ana", "ben", "cy", "dee", "eli"} {
    			time.Sleep(30 * time.Millisecond) // couriers trickle in
    			bids <- Bid{Courier: c, Cents: 400 + len(c)*50}
    		}
    	}()

    	start := time.Now()
    	got := collectBids(bids, 100*time.Millisecond)
    	fmt.Println("bids:", got)
    	fmt.Println("window took", time.Since(start).Round(50*time.Millisecond))
    }
  tests: |
    package main

    import (
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // sendBids sends one bid every interval, then closes the channel if
    // closeAfter is true. The channel is buffered so this goroutine never
    // gets stuck, whatever collectBids does.
    func sendBids(n int, interval time.Duration, closeAfter bool) <-chan Bid {
    	ch := make(chan Bid, n)
    	go func() {
    		for i := range n {
    			time.Sleep(interval)
    			ch <- Bid{Courier: string(rune('a' + i)), Cents: 500 + i}
    		}
    		if closeAfter {
    			close(ch)
    		}
    	}()
    	return ch
    }

    func TestWindowIsNotResetByBids(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		bids := sendBids(10, 25*time.Second, true) // bids at 25s, 50s, 75s, ...
    		start := time.Now()
    		got := collectBids(bids, time.Minute)
    		took := time.Since(start)

    		if len(got) != 2 {
    			t.Errorf("collectBids(window=1m) with a bid every 25s collected %d bids, want 2 (the 1m window must start when collectBids is called, not restart after every bid)", len(got))
    		}
    		if took != time.Minute {
    			t.Errorf("collectBids(window=1m) returned after %v, want exactly 1m0s", took)
    		}
    		time.Sleep(time.Hour) // let the sender finish
    	})
    }

    func TestReturnsWhenChannelCloses(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		bids := sendBids(2, 10*time.Second, true) // closes at 20s
    		start := time.Now()
    		got := collectBids(bids, time.Minute)
    		took := time.Since(start)

    		if len(got) != 2 || got[0].Courier != "a" || got[1].Courier != "b" {
    			t.Errorf("collectBids = %v, want the 2 bids [{a 500} {b 501}] in order", got)
    		}
    		if took != 20*time.Second {
    			t.Errorf("collectBids returned after %v, want 20s (as soon as the bids channel was closed)", took)
    		}
    	})
    }

    func TestNoBids(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		bids := make(chan Bid) // nobody ever bids
    		start := time.Now()
    		got := collectBids(bids, 45*time.Second)
    		if len(got) != 0 {
    			t.Errorf("collectBids with no bidders = %v, want no bids", got)
    		}
    		if took := time.Since(start); took != 45*time.Second {
    			t.Errorf("collectBids(window=45s) with no bidders returned after %v, want 45s", took)
    		}
    	})
    }
---

Dispatchly calls restaurant APIs, courier phones and payment providers. Any of them can hang. A program that waits forever on a slow partner is a program that's down. **Every wait on something outside your control needs a timeout.**

## time.After

`time.After(d)` returns a channel that receives the current time once, after `d`. Put it in a `select` alongside the operation you're waiting on, and whichever finishes first wins:

```go
package main

import (
	"errors"
	"fmt"
	"time"
)

var errTimeout = errors.New("restaurant did not answer in time")

func confirm(restaurant string, delay, timeout time.Duration) (string, error) {
	reply := make(chan string, 1) // buffered: the sender never gets stuck
	go func() {
		time.Sleep(delay) // pretend to call the restaurant's API
		reply <- restaurant + " confirmed"
	}()

	select {
	case r := <-reply:
		return r, nil
	case <-time.After(timeout):
		return "", errTimeout
	}
}

func main() {
	fmt.Println(confirm("pho-king", 10*time.Millisecond, 50*time.Millisecond))
	fmt.Println(confirm("thai-tanic", 200*time.Millisecond, 50*time.Millisecond))
}
```

```text
pho-king confirmed <nil>
 restaurant did not answer in time
```

(The second line starts with a space because `Println` prints the empty string, then a space, then the error.)

Look at `reply`: it's **buffered with capacity 1**. After a timeout nobody will ever receive from it, and with an unbuffered channel the goroutine would block forever on its send. That's the leak from chapter 2. The buffer gives the late reply somewhere to go, so the goroutine can finish.

Also note what a timeout does *not* do: the slow call keeps running in its goroutine. `select` only stops *waiting*. To actually cancel the work you need to tell it to stop, which is what contexts are for in the next chapter.

## Per-operation vs total timeouts

Where you call `time.After` decides what you're measuring.

**Inside the loop**, a new timer is created every iteration. That's an **idle timeout**: "give up if nothing arrives for 5 seconds".

```go
for {
	select {
	case b := <-bids:
		record(b)
	case <-time.After(5 * time.Second): // restarts after every bid
		return
	}
}
```

**Outside the loop**, one timer covers the whole thing. That's a **total timeout**: "give up 5 seconds from now, no matter what".

```go
deadline := time.After(5 * time.Second) // created once
for {
	select {
	case b := <-bids:
		record(b)
	case <-deadline:
		return
	}
}
```

Both are useful, but they're easy to mix up. If bids trickle in every couple of seconds, the first version never times out at all.

## Timers and garbage collection

Older Go code avoids `time.After` in loops and uses `time.NewTimer` with `defer t.Stop()`, because before Go 1.23 an unfired timer stayed in memory until it fired. Since Go 1.23 the garbage collector frees unreferenced timers even if they haven't fired, and the `time` documentation says there's no reason to prefer `NewTimer` when `After` will do. Use `time.NewTimer` when you actually need its methods, for example `Reset` to push a deadline back.

## Your turn

When a new order comes in, Dispatchly opens a **bidding window**: nearby couriers send bids, and after the window closes, the best one wins. `collectBids` gathers bids until the window runs out or the `bids` channel is closed, whichever comes first. The window is measured from when `collectBids` is called.

The starter has the classic bug from this lesson: every bid restarts the window, so a steady stream of bids keeps it open forever. Press **Run** and look at how long the window took. Then fix `collectBids`. It must still return straight away if `bids` is closed before the window ends.
