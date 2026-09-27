---
title: Fan-Out and Fan-In
quiz:
  - question: Three planner goroutines all `range` over the same `orders` channel. How is each order handled?
    options:
      - text: Every planner receives every order, so each order is planned three times
      - text: Each order is received by exactly one planner, whichever is free
        correct: true
      - text: The orders are split into three equal thirds up front
      - text: Only the first planner receives anything
    explanation: |
      A value sent on a channel is received exactly once. Several receivers
      on one channel compete for values, which spreads the work across
      them automatically. Faster workers simply take more.
  - question: 'In `fanIn`, why is `close(out)` done in a separate goroutine after `wg.Wait()`, instead of by each forwarding goroutine when its input ends?'
    options:
      - text: Closing from a goroutine is faster
      - text: The first forwarder to finish would close `out` while the others are still sending, and they'd panic
        correct: true
      - text: '`close` can only be called from the goroutine that created the channel'
      - text: It doesn't matter where it's closed
    explanation: |
      `out` has several senders, so none of them can close it. The closer
      goroutine waits for all of them and then closes it exactly once, the
      same multiple-senders rule from the closing channels lesson.
---

A pipeline is only as fast as its slowest stage. In Dispatchly, planning a delivery route calls a maps API and takes 50ms per order, while every other stage takes microseconds. One route-planning goroutine is the bottleneck. The fix is to run **several copies of the slow stage** and combine their output. That's fan-out and fan-in.

- **Fan-out**: several goroutines receive from the **same** channel. Each value goes to exactly one of them, so the work spreads out automatically.
- **Fan-in**: several channels are merged into **one**, so the next stage sees a single stream.

```text
                ┌─► planner 1 ─┐
orders ─────────┼─► planner 2 ─┼──► fanIn ──► routes
                └─► planner 3 ─┘
```

## In code

```go
package main

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"sync"
	"time"
)

type Route struct {
	Order   string
	Minutes int
}

// planRoute is slow: it calls a maps API.
func planRoute(order string) Route {
	time.Sleep(50 * time.Millisecond)
	return Route{Order: order, Minutes: 10 + 3*len(order)}
}

// planner is one fan-out worker: it plans routes for orders from in.
func planner(ctx context.Context, in <-chan string) <-chan Route {
	out := make(chan Route)
	go func() {
		defer close(out)
		for o := range in {
			select {
			case out <- planRoute(o):
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// fanIn merges any number of channels into one.
func fanIn[T any](ctx context.Context, chans ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup
	for _, ch := range chans {
		wg.Go(func() {
			for v := range ch {
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	orders := make(chan string)
	go func() {
		defer close(orders)
		for _, o := range []string{"a", "bb", "ccc", "dddd", "eeeee", "ffffff"} {
			orders <- o
		}
	}()

	// Fan out: three planners share one input channel.
	p1 := planner(ctx, orders)
	p2 := planner(ctx, orders)
	p3 := planner(ctx, orders)

	start := time.Now()
	var routes []Route
	for r := range fanIn(ctx, p1, p2, p3) { // fan in
		routes = append(routes, r)
	}
	slices.SortFunc(routes, func(a, b Route) int { return cmp.Compare(a.Minutes, b.Minutes) })

	fmt.Println(routes)
	fmt.Println("took about", time.Since(start).Round(50*time.Millisecond))
}
```

```text
[{a 13} {bb 16} {ccc 19} {dddd 22} {eeeee 25} {ffffff 28}]
took about 100ms
```

Six orders at 50ms each would take 300ms with one planner. Three planners finish in about 100ms.

## Things to notice

**Fan-out needs nothing special.** Just start several stages that read the same channel. When the source closes `orders`, *all* planners' `range` loops end, because a close is seen by every receiver.

**`fanIn` is the multiple-senders pattern.** Each forwarding goroutine sends on `out`, so none of them may close it. A separate closer waits on the `WaitGroup` and closes `out` once. It's generic, so it works for any element type. (It's also a more general version of the `merge` you wrote with nil channels, which handled exactly two inputs in a single goroutine.)

**Order is lost.** With several workers, results come out in whatever order they finish. If order matters, carry an index with each item and put results back in place (as the worker pool in the next lesson does), or sort at the end, as here.

**How many workers?** For I/O-bound stages like this one, as many as the remote service will tolerate. That might be dozens. For CPU-bound stages, more than `runtime.GOMAXPROCS(0)` won't help. Measure it either way.

## Further reading

- [Go blog: Pipelines and cancellation](https://go.dev/blog/pipelines), the "Fan-out, fan-in" section
