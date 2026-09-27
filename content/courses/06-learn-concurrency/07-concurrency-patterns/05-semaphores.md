---
title: Bounded Parallelism with a Semaphore
quiz:
  - question: |
      What's wrong with this semaphore?

      ```go
      sem := make(chan struct{}, 3)
      for _, r := range restaurants {
      	wg.Go(func() {
      		sem <- struct{}{}
      		defer func() { <-sem }()
      		syncMenu(r)
      	})
      }
      ```
    options:
      - text: It allows more than 3 syncs at once
      - text: Nothing is wrong with the limit, but it starts a goroutine for every restaurant up front, and most of them just sit waiting for a slot
        correct: true
      - text: It deadlocks, because the release happens before the acquire
      - text: The channel should be unbuffered
    explanation: |
      Acquiring inside the goroutine still caps concurrent `syncMenu` calls
      at 3, so it's correct. But with 100,000 restaurants you'd create
      100,000 goroutines immediately. Acquiring in the loop *before* `go`
      means at most 3 (plus one waiting loop) exist at any time.
  - question: A semaphore channel is `make(chan struct{}, 5)`. What does its capacity represent?
    options:
      - text: The number of goroutines that may wait for a slot
      - text: The number of tasks allowed to run at the same time
        correct: true
      - text: The total number of tasks
      - text: How many times each task may retry
    explanation: |
      Each running task holds one value in the buffer. When all 5 slots are
      full, the next send (acquire) blocks until a running task receives
      (releases) one.
---

A worker pool bounds concurrency with a fixed set of long-lived goroutines. There's another way that's often simpler: keep "one goroutine per task", but make each task acquire a slot from a **semaphore** before it runs. In Go, a semaphore is just a buffered channel.

## A buffered channel as a semaphore

The channel's **capacity is the limit**. Sending takes a slot, receiving gives it back:

```go
sem := make(chan struct{}, 3)

sem <- struct{}{} // acquire: blocks if 3 slots are taken
// ... do the work ...
<-sem             // release
```

The values carry no information, hence `struct{}`. What matters is how many are in the buffer: once it holds 3, the next send blocks until someone receives.

Dispatchly re-syncs restaurant menus every night, and the menu service allows three concurrent calls:

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

// gauge tracks how many syncs are running now, and the peak.
type gauge struct {
	mu        sync.Mutex
	now, peak int
}

func (g *gauge) add(delta int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.now += delta
	g.peak = max(g.peak, g.now)
}

func main() {
	restaurants := []string{"pho-king", "wok-this-way", "thai-tanic", "grill-seeker",
		"curry-up", "bun-intended", "lord-of-the-fries", "brewed-awakening"}

	var g gauge
	syncMenu := func(r string) {
		g.add(1)
		defer g.add(-1)
		time.Sleep(20 * time.Millisecond) // call r's API
	}

	sem := make(chan struct{}, 3) // at most 3 syncs at once
	var wg sync.WaitGroup
	for _, r := range restaurants {
		sem <- struct{}{} // acquire: blocks while 3 syncs are running
		wg.Go(func() {
			defer func() { <-sem }() // release
			syncMenu(r)
		})
	}
	wg.Wait()
	fmt.Println("synced", len(restaurants), "menus, peak concurrency:", g.peak)
}
```

```text
synced 8 menus, peak concurrency: 3
```

## Acquire before `go`

Notice where the acquire happens: in the loop, **before** starting the goroutine. The loop itself blocks when all slots are busy, so there are never more than 3 sync goroutines alive at once. If you acquire *inside* the goroutine instead, the limit still holds for the work, but you start a goroutine for every restaurant immediately, and thousands of them sit around waiting for a slot. It's correct, just wasteful.

Always release with `defer`, so a task that returns early (or hits an error) still frees its slot. A leaked slot permanently shrinks your limit, and after enough of them everything stops.

## Making the acquire cancellable

A bare `sem <- struct{}{}` can block for a long time. When there's a context, acquire in a `select`:

```go
for _, r := range restaurants {
	select {
	case sem <- struct{}{}:
	case <-ctx.Done():
		wg.Wait() // let the running syncs finish
		return ctx.Err()
	}
	wg.Go(func() {
		defer func() { <-sem }()
		syncMenu(ctx, r)
	})
}
```

## Semaphore or worker pool?

They give the same guarantee: at most N at once. Pick the one that reads better:

| Semaphore | Worker pool |
| --- | --- |
| One short-lived goroutine per task | N long-lived goroutines |
| Easy to add to an existing "goroutine per task" loop | Workers can hold per-worker state (a connection, a buffer) |
| Tasks can be different functions | Jobs are data sent over a channel |

The `golang.org/x/sync/semaphore` package offers a **weighted** semaphore, where a big task can take several slots at once. For equal-sized tasks, the buffered channel is all you need.
