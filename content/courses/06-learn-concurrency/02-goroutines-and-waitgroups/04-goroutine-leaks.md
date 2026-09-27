---
title: Goroutine Leaks
quiz:
  - question: |
      `lookup` is called once per incoming request. What goes wrong over time?

      ```go
      func lookup(id string) (Order, error) {
      	ch := make(chan Order)
      	go func() { ch <- fetchOrder(id) }()
      	select {
      	case o := <-ch:
      		return o, nil
      	case <-time.After(time.Second):
      		return Order{}, errors.New("timeout")
      	}
      }
      ```
    options:
      - text: Nothing, the goroutine is garbage-collected when `lookup` returns
      - text: Every timed-out request leaves a goroutine blocked forever on `ch <-`, so memory grows until the service falls over
        correct: true
      - text: '`time.After` leaks a timer on every call'
      - text: It deadlocks on the first timeout
    explanation: |
      After a timeout nobody will ever receive from `ch`, so the goroutine's
      send blocks forever. Blocked goroutines are never collected. The fix is
      one character: `make(chan Order, 1)`, so the send always succeeds
      and the goroutine can exit even if nobody reads the result.
  - question: What's the best rule of thumb for avoiding goroutine leaks?
    options:
      - text: Call `runtime.GC()` regularly
      - text: Never use unbuffered channels
      - text: Before starting a goroutine, know exactly how and when it will stop
        correct: true
      - text: Keep the number of goroutines below GOMAXPROCS
    explanation: |
      Every goroutine needs a guaranteed exit path: its channel will be
      closed, its send will always find room or a receiver, or it watches
      a done channel or context. If you can't say when it stops, it might
      never stop.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"runtime"
    	"time"
    )

    // fastestETA asks every courier for an ETA concurrently and returns the
    // first answer that comes back. It returns 0 if there are no couriers.
    // It must not leave any goroutines behind.
    func fastestETA(couriers []string, eta func(courier string) time.Duration) time.Duration {
    	if len(couriers) == 0 {
    		return 0
    	}
    	results := make(chan time.Duration)
    	for _, c := range couriers {
    		go func() {
    			results <- eta(c)
    		}()
    	}
    	return <-results
    }

    func main() {
    	eta := func(c string) time.Duration {
    		d := time.Duration(len(c)) * 10 * time.Millisecond
    		time.Sleep(d) // pretend to ask the courier's phone
    		return d
    	}

    	fmt.Println("fastest ETA:", fastestETA([]string{"ana", "benedict", "cy"}, eta))

    	time.Sleep(100 * time.Millisecond) // give the slow couriers time to answer
    	fmt.Println("goroutines left behind:", runtime.NumGoroutine()-1)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"runtime"
    	"time"
    )

    func fastestETA(couriers []string, eta func(courier string) time.Duration) time.Duration {
    	if len(couriers) == 0 {
    		return 0
    	}
    	// Room for every answer, so no sender ever blocks.
    	results := make(chan time.Duration, len(couriers))
    	for _, c := range couriers {
    		go func() {
    			results <- eta(c)
    		}()
    	}
    	return <-results
    }

    func main() {
    	eta := func(c string) time.Duration {
    		d := time.Duration(len(c)) * 10 * time.Millisecond
    		time.Sleep(d)
    		return d
    	}

    	fmt.Println("fastest ETA:", fastestETA([]string{"ana", "benedict", "cy"}, eta))

    	time.Sleep(100 * time.Millisecond)
    	fmt.Println("goroutines left behind:", runtime.NumGoroutine()-1)
    }
  tests: |
    package main

    import (
    	"runtime"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    func TestFastestETANoLeak(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		before := runtime.NumGoroutine()
    		eta := func(c string) time.Duration {
    			d := time.Duration(len(c)) * time.Second
    			time.Sleep(d)
    			return d
    		}
    		fastestETA([]string{"a", "bb", "ccc", "dddd", "eeeee"}, eta)

    		time.Sleep(time.Hour) // plenty of time for every courier to answer
    		synctest.Wait()
    		if n := runtime.NumGoroutine() - before; n > 0 {
    			t.Errorf("%d goroutine(s) are still blocked after fastestETA returned: they leaked", n)
    		}
    	})
    }

    func TestFastestETA(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		etas := map[string]time.Duration{
    			"ana": 7 * time.Minute, "ben": 3 * time.Minute, "cy": 12 * time.Minute, "dee": 5 * time.Minute,
    		}
    		eta := func(c string) time.Duration {
    			time.Sleep(etas[c] / 60) // fake time: answers arrive after a few seconds
    			return etas[c]
    		}
    		start := time.Now()
    		got := fastestETA([]string{"ana", "ben", "cy", "dee"}, eta)
    		if got != 3*time.Minute {
    			t.Errorf("fastestETA = %v, want 3m0s", got)
    		}
    		if took := time.Since(start); took != 3*time.Second {
    			t.Errorf("fastestETA returned after %v, want 3s (as soon as the first answer arrives)", took)
    		}
    		time.Sleep(time.Hour) // let the slower couriers finish before the test ends
    	})
    }

    func TestFastestETANoCouriers(t *testing.T) {
    	if got := fastestETA(nil, func(string) time.Duration { return time.Minute }); got != 0 {
    		t.Errorf("fastestETA(nil) = %v, want 0", got)
    	}
    }

---

A goroutine that never finishes is a **goroutine leak**. Leaks are the most common concurrency bug in real Go services, because nothing warns you. The program works, the tests pass, and memory creeps up for days until the service is restarted or killed.

## Why leaked goroutines are never cleaned up

The garbage collector frees memory that nothing refers to. A goroutine, though, is a *live* thing: the runtime can't know that the channel it's waiting on will never be used again, so it can't safely remove it. A blocked goroutine keeps its stack (at least a few KB), and everything its stack points to, alive forever.

One leaked goroutine per request is nothing. Multiply that by ten thousand requests per minute and you have an outage.

## How goroutines leak

Nearly every leak is a goroutine **blocked forever** on one of these:

1. **Sending with no receiver.** The receiver gave up (timed out, returned early on an error) and nobody will ever read the value.
2. **Receiving with no sender.** The goroutine ranges over a channel that nobody ever closes, or waits for a reply that never comes.
3. **Waiting for a signal that never comes.** A `for`/`select` loop with no way to be told to stop.

Here's the classic, number 1. Dispatchly asks three couriers for an ETA and takes the first answer:

```go
package main

import (
	"fmt"
	"runtime"
	"time"
)

func firstAnswer(couriers []string) string {
	answers := make(chan string) // unbuffered
	for _, c := range couriers {
		go func() {
			time.Sleep(time.Duration(len(c)) * time.Millisecond) // ask the courier
			answers <- c
		}()
	}
	return <-answers // take one answer... and abandon the rest
}

func main() {
	for range 100 {
		firstAnswer([]string{"ana", "benedict", "cy"})
	}
	time.Sleep(50 * time.Millisecond)
	fmt.Println("goroutines:", runtime.NumGoroutine())
}
```

```text
goroutines: 201
```

Each call reads one answer. The other two goroutines block forever on `answers <- c`, because nobody will ever receive again. After 100 calls, 200 goroutines are stuck, plus `main`.

## How to spot leaks

- **Watch `runtime.NumGoroutine()`.** Export it as a metric. A count that climbs steadily under constant load is a leak.
- **Dump the stacks.** The `goroutine` profile from `runtime/pprof` (or the `/debug/pprof/goroutine?debug=1` page from `net/http/pprof`) groups goroutines by stack trace. Hundreds stuck on the same `chan send` line is your leak. Sending `SIGQUIT` (Ctrl+\\ in a terminal) to a Go program also prints every goroutine's stack.
- **The `goroutineleak` profile.** An experiment in Go 1.26 and available by default in Go 1.27, this profile asks the garbage collector to find goroutines blocked on channels or locks that nothing else can reach, which means they are leaked for certain. You'll use it in the final chapter.
- **Test for them.** `testing/synctest` fails a test that leaves goroutines blocked, and you can compare `runtime.NumGoroutine()` before and after. The tests for this lesson's exercise do exactly that.

## How to prevent them

The golden rule: **never start a goroutine without knowing how it will stop.** For each goroutine, you should be able to say "it exits when X happens", and X must be guaranteed. The usual tools:

- **Buffered channels** sized so every send has room, as in the exercise below.
- **Closing channels** so `range` loops end (chapter 3).
- **A done channel or a `context.Context`** that tells a goroutine to give up, checked in a `select` (chapters 4 and 5).
- **The caller waits** with a `WaitGroup`, so a stuck goroutine shows up as a stuck caller, which is much easier to notice.

## Your turn

`fastestETA` has exactly the leak from this lesson: it takes the first ETA and abandons every other goroutine, blocked forever on its send. Press **Run** to see the leftovers.

Fix it so that no goroutine is left behind, even though only the first answer is used. It should still return as soon as the first answer arrives. You don't need anything you haven't seen yet: think about what would let every send complete even when nobody receives.
