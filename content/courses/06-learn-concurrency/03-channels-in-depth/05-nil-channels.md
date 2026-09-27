---
title: Nil Channels
quiz:
  - question: |
      What does this print?

      ```go
      var feed chan string
      live := make(chan string, 1)
      live <- "ping"

      select {
      case m := <-feed:
      	fmt.Println("feed:", m)
      case m := <-live:
      	fmt.Println("live:", m)
      }
      ```
    options:
      - text: '`feed: ` (an empty string, because `feed` is nil)'
      - text: '`live: ping`'
        correct: true
      - text: It panics, because `feed` was never made
      - text: Either line, chosen at random
    explanation: |
      `feed` is a nil channel, so its case can never be ready. `select` only
      picks among ready cases, and the only one is `live`. Receiving from a
      nil channel doesn't panic, it just blocks forever, which inside a
      `select` means "this case is switched off".
  - question: 'Inside a `for`/`select` loop, why set `a = nil` when `v, ok := <-a` returns `ok == false`?'
    options:
      - text: To let the garbage collector free the channel sooner
      - text: Because a closed channel must be set to nil before it can be reused
      - text: 'A closed channel is *always* ready, so without it the loop would spin on zero values; a nil channel is *never* ready, so its case is disabled'
        correct: true
      - text: To close the channel a second time safely
    explanation: |
      Receiving from a closed channel returns immediately, every time. Leave
      it in the `select` and that case keeps winning, burning CPU on
      zero values. Setting the variable to nil takes the case out of the
      running while the other cases carry on.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"time"
    )

    // merge forwards every value from a and b onto the returned channel,
    // in whatever order they arrive. The returned channel is closed once
    // both a and b are closed.
    func merge(a, b <-chan int) <-chan int {
    	out := make(chan int)
    	go func() {
    		defer close(out)
    		for {
    			select {
    			case v, ok := <-a:
    				if !ok {
    					return // BUG: b may still have orders coming
    				}
    				out <- v
    			case v, ok := <-b:
    				if !ok {
    					return // BUG: a may still have orders coming
    				}
    				out <- v
    			}
    		}
    	}()
    	return out
    }

    // feed sends each order ID on a new channel, pausing between them,
    // then closes it.
    func feed(pause time.Duration, ids ...int) <-chan int {
    	ch := make(chan int)
    	go func() {
    		defer close(ch)
    		for _, id := range ids {
    			time.Sleep(pause)
    			ch <- id
    		}
    	}()
    	return ch
    }

    func main() {
    	app := feed(10*time.Millisecond, 101, 102)
    	phone := feed(15*time.Millisecond, 201, 202, 203)

    	for id := range merge(app, phone) {
    		fmt.Println("order", id)
    	}
    	fmt.Println("all orders in")
    }
  solution: |
    package main

    import (
    	"fmt"
    	"time"
    )

    // merge forwards every value from a and b onto the returned channel,
    // in whatever order they arrive. The returned channel is closed once
    // both a and b are closed.
    func merge(a, b <-chan int) <-chan int {
    	out := make(chan int)
    	go func() {
    		defer close(out)
    		for a != nil || b != nil {
    			select {
    			case v, ok := <-a:
    				if !ok {
    					a = nil
    					continue
    				}
    				out <- v
    			case v, ok := <-b:
    				if !ok {
    					b = nil
    					continue
    				}
    				out <- v
    			}
    		}
    	}()
    	return out
    }

    // feed sends each order ID on a new channel, pausing between them,
    // then closes it.
    func feed(pause time.Duration, ids ...int) <-chan int {
    	ch := make(chan int)
    	go func() {
    		defer close(ch)
    		for _, id := range ids {
    			time.Sleep(pause)
    			ch <- id
    		}
    	}()
    	return ch
    }

    func main() {
    	app := feed(10*time.Millisecond, 101, 102)
    	phone := feed(15*time.Millisecond, 201, 202, 203)

    	for id := range merge(app, phone) {
    		fmt.Println("order", id)
    	}
    	fmt.Println("all orders in")
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // collect reads from out until it is closed, or until an hour of (fake)
    // time passes without it closing.
    func collect(out <-chan int) (vals []int, closed bool) {
    	for {
    		select {
    		case v, ok := <-out:
    			if !ok {
    				return vals, true
    			}
    			vals = append(vals, v)
    		case <-time.After(time.Hour):
    			return vals, false
    		}
    	}
    }

    // timedFeed sends each value after a delay. Its channel is buffered so
    // the test's own goroutine never gets stuck, whatever merge does.
    func timedFeed(delays []time.Duration, vals []int) <-chan int {
    	ch := make(chan int, len(vals))
    	go func() {
    		defer close(ch)
    		for i, v := range vals {
    			time.Sleep(delays[i])
    			ch <- v
    		}
    	}()
    	return ch
    }

    func TestMergeKeepsReadingAfterOneInputCloses(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		s := time.Second
    		a := timedFeed([]time.Duration{1 * s, 1 * s}, []int{1, 2}) // closes at 2s
    		b := timedFeed([]time.Duration{3 * s, 1 * s, 1 * s}, []int{10, 20, 30})

    		got, closed := collect(merge(a, b))
    		slices.Sort(got)
    		want := []int{1, 2, 10, 20, 30}
    		if !slices.Equal(got, want) {
    			t.Errorf("merge delivered %v, want %v (did it stop as soon as the first input closed?)", got, want)
    		}
    		if !closed {
    			t.Errorf("the merged channel was never closed after both inputs were closed")
    		}
    		time.Sleep(time.Hour) // let the test's feeders finish
    	})
    }

    func TestMergeOtherOrder(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		s := time.Second
    		a := timedFeed([]time.Duration{2 * s, 5 * s}, []int{7, 8})
    		b := timedFeed([]time.Duration{1 * s}, []int{9}) // closes at 1s
    		got, closed := collect(merge(a, b))
    		slices.Sort(got)
    		if want := []int{7, 8, 9}; !slices.Equal(got, want) {
    			t.Errorf("merge delivered %v, want %v", got, want)
    		}
    		if !closed {
    			t.Errorf("the merged channel was never closed after both inputs were closed")
    		}
    		time.Sleep(time.Hour) // let the test's feeders finish
    	})
    }

    func TestMergeBothEmpty(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		a, b := make(chan int), make(chan int)
    		close(a)
    		close(b)
    		got, closed := collect(merge(a, b))
    		if len(got) != 0 {
    			t.Errorf("merge of two closed, empty channels delivered %v, want nothing", got)
    		}
    		if !closed {
    			t.Errorf("merge of two closed channels never closed its output")
    		}
    	})
    }
---

The zero value of a channel type is `nil`. You get one whenever you declare a channel without `make`:

```go
var updates chan string // nil
```

A nil channel is not broken. It has very precise behaviour:

- **Sending** on it blocks forever.
- **Receiving** from it blocks forever.
- **Closing** it panics.

Blocking forever sounds useless, and outside a `select` it's always a bug (usually a forgotten `make`). Inside a `select`, though, it's a feature.

## A nil case is never chosen

`select` waits until one of its cases is ready and runs it. A case on a nil channel can never be ready, so `select` simply ignores it. Setting a channel variable to `nil` **switches that case off**:

```go
package main

import "fmt"

func main() {
	var paused chan string // nil: this courier's feed is switched off
	live := make(chan string, 1)
	live <- "ana is at the restaurant"

	for range 2 {
		select {
		case msg := <-paused:
			fmt.Println("paused feed:", msg) // never happens
		case msg := <-live:
			fmt.Println("live feed:", msg)
		default:
			fmt.Println("nothing to report")
		}
	}
}
```

```text
live feed: ana is at the restaurant
nothing to report
```

## Why you need it: closed channels are always ready

Here's the problem nil channels solve. A closed channel is the *opposite* of a nil one: a receive on it is always ready, returning the zero value at once. Suppose you wait on two order sources:

```go
for {
	select {
	case id := <-app:
		dispatch(id)
	case id := <-phone:
		dispatch(id)
	}
}
```

As soon as `app` is closed, its case is ready on every iteration, and the loop spins, dispatching order `0` millions of times a second. Returning when `app` closes isn't right either: `phone` may still have orders coming.

The fix is to check `ok` and set the closed channel to nil, so only the live source stays in the `select`. When both are nil, you're done:

```go
for app != nil || phone != nil {
	select {
	case id, ok := <-app:
		if !ok {
			app = nil // app is finished: disable this case
			continue
		}
		dispatch(id)
	case id, ok := <-phone:
		if !ok {
			phone = nil
			continue
		}
		dispatch(id)
	}
}
```

This changes only the *local variable*. The channel itself is untouched, so any other goroutine holding it is unaffected.

## Switching sends on and off

The same trick works for sends. A dispatcher holding a backlog of orders only wants its send case active when there's something to send:

```go
var sendTo chan<- Order // nil while the backlog is empty
var next Order
if len(backlog) > 0 {
	sendTo = couriers
	next = backlog[0]
}

select {
case o := <-incoming:
	backlog = append(backlog, o)
case sendTo <- next: // only possible when sendTo isn't nil
	backlog = backlog[1:]
}
```

One `select`, two behaviours, and no special-casing.

## Your turn

Dispatchly receives orders from the app and from phone calls, each on its own channel. `merge` should forward every order from both onto one channel, then close it once **both** inputs are closed.

The starter version returns as soon as *either* input closes, dropping orders from the other. Press **Run** to see orders go missing. Fix `merge` with the nil-channel technique from this lesson.
