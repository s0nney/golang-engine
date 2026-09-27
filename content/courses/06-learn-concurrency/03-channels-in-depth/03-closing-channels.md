---
title: Closing Channels
quiz:
  - question: Three goroutines send results on the same channel. Who should close it?
    options:
      - text: Each sender, when it finishes
      - text: The receiver, once it has read enough values
      - text: A separate goroutine that waits for all three senders (for example with a `WaitGroup`) and then closes it
        correct: true
      - text: Nobody, closing channels is optional and only an optimization
    explanation: |
      If each sender closes, the second `close` panics, and a sender that's
      still running panics on its next send. The receiver can't close safely
      either, because senders may still be sending. A single closer that waits
      for every sender is the standard fix.
  - question: Which of these operations does **not** panic?
    options:
      - text: Closing a channel that's already closed
      - text: Sending on a closed channel
      - text: Receiving from a closed channel
        correct: true
      - text: Closing a `nil` channel
    explanation: |
      Receiving from a closed channel is perfectly normal: it returns
      buffered values first, then zero values with `ok == false`. The other
      three are all runtime panics.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"sync"
    )

    // streamTotals starts one goroutine per restaurant. Each sends that
    // restaurant's order totals on the returned channel. The channel must be
    // closed once every restaurant has sent all of its totals.
    func streamTotals(restaurants [][]int) <-chan int {
    	totals := make(chan int)
    	var wg sync.WaitGroup
    	for _, orders := range restaurants {
    		wg.Go(func() {
    			for _, cents := range orders {
    				totals <- cents
    			}
    		})
    	}
    	// TODO: close totals once every sender is done, without blocking here.
    	return totals
    }

    func main() {
    	restaurants := [][]int{
    		{1200, 850},      // pho-king
    		{990},            // wok-this-way
    		{1500, 700, 640}, // thai-tanic
    	}

    	sum := 0
    	for cents := range streamTotals(restaurants) {
    		sum += cents
    	}
    	fmt.Println("revenue:", sum)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"sync"
    )

    // streamTotals starts one goroutine per restaurant. Each sends that
    // restaurant's order totals on the returned channel. The channel must be
    // closed once every restaurant has sent all of its totals.
    func streamTotals(restaurants [][]int) <-chan int {
    	totals := make(chan int)
    	var wg sync.WaitGroup
    	for _, orders := range restaurants {
    		wg.Go(func() {
    			for _, cents := range orders {
    				totals <- cents
    			}
    		})
    	}
    	go func() {
    		wg.Wait()
    		close(totals)
    	}()
    	return totals
    }

    func main() {
    	restaurants := [][]int{
    		{1200, 850},      // pho-king
    		{990},            // wok-this-way
    		{1500, 700, 640}, // thai-tanic
    	}

    	sum := 0
    	for cents := range streamTotals(restaurants) {
    		sum += cents
    	}
    	fmt.Println("revenue:", sum)
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // drain reads totals until the channel is closed, or until an hour of
    // (fake) time passes without it closing.
    func drain(ch <-chan int) (got []int, closed bool) {
    	for {
    		select {
    		case v, ok := <-ch:
    			if !ok {
    				return got, true
    			}
    			got = append(got, v)
    		case <-time.After(time.Hour):
    			return got, false
    		}
    	}
    }

    func TestStreamTotals(t *testing.T) {
    	for _, restaurants := range [][][]int{
    		{{1200, 850}, {990}, {1500, 700, 640}},
    		{{100}},
    		{{5, 6, 7, 8}, {}, {9}},
    		{},
    	} {
    		synctest.Test(t, func(t *testing.T) {
    			var want []int
    			for _, r := range restaurants {
    				want = append(want, r...)
    			}
    			slices.Sort(want)

    			got, closed := drain(streamTotals(restaurants))
    			slices.Sort(got)
    			if !slices.Equal(got, want) {
    				t.Errorf("streamTotals(%v) sent %v, want %v (in any order)", restaurants, got, want)
    			}
    			if !closed {
    				t.Errorf("streamTotals(%v): the channel was never closed, so a range over it would wait forever", restaurants)
    			}
    		})
    	}
    }
---

You already know the basics from Learn Go: `close(ch)` says **no more values will be sent**, receivers can still drain what's buffered, after that `v, ok := <-ch` returns the zero value with `ok == false`, and sending on a closed channel panics. Closing is also a **broadcast**: every receiver, current and future, sees it.

This lesson is about the parts that bite in real code: the complete set of rules, and the question of *who* is allowed to close.

## The rules

Closing has sharp edges. These are all runtime panics, not compile errors:

| Operation | On an open channel | On a closed channel | On a `nil` channel |
| --- | --- | --- | --- |
| `close(ch)` | closes it | **panic** | **panic** |
| `ch <- v` | sends (may block) | **panic** | blocks forever |
| `<-ch` | receives (may block) | zero value, `ok == false` | blocks forever |

```text
panic: close of closed channel
panic: send on closed channel
```

There's no way to ask "is this channel closed?" without receiving from it, and even if there were, the answer could change before you acted on it. So you can't close "safely" by checking first. You have to *know* that nobody else will close it and nobody will send on it again.

## Who closes?

That leads to one rule: **the sender closes, never the receiver.** More precisely, the *owner* closes: the one goroutine that knows no more sends will happen.

- **One sender:** it closes the channel when it's done, usually with `defer close(out)`.
- **Many senders:** none of them can close it, because the others may still be sending. Add a separate **closer** goroutine that waits for all of them:

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	couriers := [][]string{
		{"ana: picked up A1", "ana: delivered A1"},
		{"ben: picked up B2"},
	}

	updates := make(chan string)
	var wg sync.WaitGroup
	for _, msgs := range couriers {
		wg.Go(func() {
			for _, m := range msgs {
				updates <- m
			}
		})
	}

	// The closer: once every sender is done, close the channel.
	go func() {
		wg.Wait()
		close(updates)
	}()

	n := 0
	for range updates {
		n++
	}
	fmt.Println("status updates received:", n)
}
```

```text
status updates received: 3
```

Why is `wg.Wait()` in its own goroutine? If `main` called it before reading, the senders would block on the unbuffered channel, `Wait` would never return, and everything would deadlock. The closer lets `main` start receiving right away. You'll use this shape again and again.

- **The receiver wants to stop early:** it must *not* close the channel to make the senders quit, because they'd panic on their next send. It needs a separate way to say "stop", like a done channel or a context, both coming up in the next two chapters.

## Your turn

Dispatchly's revenue report starts one goroutine per restaurant, and each one sends that restaurant's order totals on a shared channel. `main` adds them up with a `range` loop. Press **Run**: the program crashes with a deadlock, because nothing ever closes `totals`, so the `range` never ends.

Fix `streamTotals` so the channel is closed exactly once, after **every** sender has finished. `streamTotals` itself must return straight away, so the caller can start receiving.
