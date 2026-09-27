---
title: Non-Blocking Operations with Default
quiz:
  - question: |
      What does this print?

      ```go
      alerts := make(chan string, 1)
      for _, a := range []string{"late", "cold", "lost"} {
      	select {
      	case alerts <- a:
      		fmt.Println("sent", a)
      	default:
      		fmt.Println("skipped", a)
      	}
      }
      ```
    options:
      - text: '`sent late`, `sent cold`, `sent lost`'
      - text: '`sent late`, `skipped cold`, `skipped lost`'
        correct: true
      - text: '`sent late`, then it deadlocks'
      - text: '`skipped late`, `skipped cold`, `skipped lost`'
    explanation: |
      The buffer has room for one value, so the first send succeeds. Nobody
      receives, so the buffer stays full, and the next two sends can't
      proceed. With a `default` case the `select` doesn't block: it runs
      `default` instead.
  - question: |
      What's wrong with this loop?

      ```go
      for {
      	select {
      	case o := <-orders:
      		dispatch(o)
      	default:
      	}
      }
      ```
    options:
      - text: Nothing, it's the idiomatic way to wait for orders
      - text: It busy-waits, spinning a CPU core at 100% whenever no order is waiting
        correct: true
      - text: It misses orders that arrive while `default` runs
      - text: It deadlocks when `orders` is empty
    explanation: |
      With an empty `default`, the `select` returns instantly whenever
      `orders` is empty, and the loop immediately tries again, millions of
      times a second. Just drop the `default`: a blocking receive waits
      without using any CPU.
---

A `select` with a `default` case **never blocks**. If none of the other cases can proceed right now, `default` runs instead. That turns any send or receive into a "try" operation.

```go
select {
case v := <-ch:
	// got a value
default:
	// nothing was ready right now
}
```

## Try-send: drop instead of blocking

Dispatchly emits a metrics event for everything that happens. Metrics are nice to have, but dispatching orders is the real job. If the metrics pipeline falls behind, it's better to drop an event than to slow dispatching down:

```go
package main

import "fmt"

func main() {
	metrics := make(chan string, 2)
	dropped := 0

	for _, event := range []string{"order.created", "order.paid", "courier.assigned", "order.delivered"} {
		select {
		case metrics <- event:
		default:
			dropped++ // buffer full: drop it rather than slow down dispatching
		}
	}
	fmt.Println("queued:", len(metrics), "dropped:", dropped)

	for range 3 {
		select {
		case e := <-metrics:
			fmt.Println("flushed", e)
		default:
			fmt.Println("nothing to flush")
		}
	}
}
```

```text
queued: 2 dropped: 2
flushed order.created
flushed order.paid
nothing to flush
```

The first loop is a **try-send**: it sends if there's room and counts a drop otherwise. The second is a **try-receive**: it takes a value if one is waiting and moves on if not. Always count or log what you drop. Silent data loss is much harder to debug than a counter that says `dropped: 4012`.

## Checking for a signal without waiting

`default` is also how a goroutine that's busy with other work can *peek* at a channel. Say a courier's route planner does long calculations in steps and should stop early if the order is cancelled:

```go
for _, stop := range stops {
	select {
	case <-cancelled:
		return errCancelled
	default:
	}
	planLeg(stop) // the real work
}
```

The `select` checks `cancelled` once per step and falls straight through if nothing's there. You'll use exactly this shape with contexts in the next chapter.

## Don't busy-wait

An empty `default` inside a `for` loop, with nothing else in the loop, is almost always a bug:

```go
for {
	select {
	case o := <-orders:
		dispatch(o)
	default:
		// nothing yet... try again immediately
	}
}
```

This **busy-waits**. When `orders` is empty it loops as fast as the CPU allows, burning a whole core to do nothing. Adding a `time.Sleep` in `default` only makes it slower to react. The fix is to delete `default`: a plain blocking `select` (or receive) waits without using any CPU, and wakes up the moment a value arrives.

Use `default` when your goroutine has **something else useful to do**, or when blocking is genuinely worse than skipping. If the answer to "what happens in default?" is "nothing, then try again", you don't want `default`.

## Further reading

- [Go by Example: Non-Blocking Channel Operations](https://gobyexample.com/non-blocking-channel-operations)
