---
title: Deadlocks
quiz:
  - question: |
      What happens when this runs?

      ```go
      func main() {
      	var wg sync.WaitGroup
      	results := make(chan int)
      	for i := range 3 {
      		wg.Go(func() { results <- i })
      	}
      	wg.Wait()
      	close(results)
      	for r := range results {
      		fmt.Println(r)
      	}
      }
      ```
    options:
      - text: It prints 0, 1 and 2 in some order
      - text: 'It crashes with `fatal error: all goroutines are asleep - deadlock!`'
        correct: true
      - text: It prints nothing and exits
      - text: It panics with "send on closed channel"
    explanation: |
      The senders block on the unbuffered channel until someone receives,
      but `main` won't receive until `wg.Wait()` returns, and that waits
      for the senders. Everyone is waiting on everyone. Move the
      `Wait`+`close` into its own goroutine, as in the closer pattern.
  - question: A web server has a goroutine leak where 1,000 goroutines are stuck forever on a channel send. Why doesn't Go report a deadlock?
    options:
      - text: Deadlock detection only works in tests
      - text: The runtime only reports a deadlock when *every* goroutine is blocked, and the server's other goroutines (like the HTTP listener) are still able to run
        correct: true
      - text: Channel sends can't deadlock, only mutexes can
      - text: Deadlock detection is disabled once GOMAXPROCS is above 1
    explanation: |
      The built-in check is global: it fires only when no goroutine at all
      can make progress. A partial deadlock, where some goroutines are
      stuck while others keep running, is silent. That's why leak
      detection and good tests matter.
---

A **deadlock** is a set of goroutines each waiting for something only another goroutine in the set can provide. Nobody can move, so nobody ever will.

## "All goroutines are asleep"

The Go runtime detects one special case: when **every** goroutine in the program is blocked, it crashes with a helpful message. You've seen the simplest version already, a send with no receiver in `main`. Here's a slightly sneakier one:

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	results := make(chan int)
	for i := range 3 {
		wg.Go(func() { results <- i * 10 })
	}
	wg.Wait() // waits for senders that are waiting for a receiver
	close(results)
	for r := range results {
		fmt.Println(r)
	}
}
```

```text
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [sync.WaitGroup.Wait]:
...
```

The senders wait for `main` to receive. `main` waits for the senders to finish. The trace shows what each goroutine is stuck on: `[sync.WaitGroup.Wait]`, `[chan send]`, `[chan receive]`, `[sync.Mutex.Lock]`. Read those labels first. They usually point straight at the cycle.

## Lock ordering

The classic deadlock with mutexes: two goroutines take the same two locks in opposite orders.

```go
func transfer(from, to *Account, cents int) {
	from.mu.Lock()
	defer from.mu.Unlock()
	time.Sleep(time.Millisecond) // widen the window, as a slow audit log would

	to.mu.Lock()
	defer to.mu.Unlock()

	from.balance -= cents
	to.balance += cents
}

// meanwhile, at the same time:
go transfer(restaurant, courier, 450) // locks restaurant, then courier
go transfer(courier, restaurant, 100) // locks courier, then restaurant
```

Each goroutine grabs its first lock and then waits forever for the other's. The fix is a **global lock order**: whenever you need several locks, always take them in the same order, for example by account ID:

```go
first, second := from, to
if second.id < first.id {
	first, second = second, first
}
first.mu.Lock()
defer first.mu.Unlock()
second.mu.Lock()
defer second.mu.Unlock()
```

Better still, avoid holding two locks at once when you can.

## Other common shapes

- **Locking twice.** Go's mutexes aren't reentrant. A method that holds `mu` and calls another method that also locks `mu` deadlocks with itself.
- **A forgotten close.** A `range` over a channel nobody closes, or a `WaitGroup` whose task never returns.
- **Holding a lock while sending on a channel.** If the receiver needs the same lock before it can receive, neither can go on.
- **A goroutine waiting on itself.** For example, sending on an unbuffered channel and then, in the same goroutine, receiving from it.

## Partial deadlocks are silent

The runtime's check only fires when *everything* is stuck. In a real service there's always something still running, like the HTTP listener, a ticker or a signal handler, so a deadlock among a few goroutines produces no crash, just requests that never finish and goroutines that pile up. To the runtime, that's indistinguishable from a program waiting patiently for work.

So don't rely on the crash. Your defences are:

- **Timeouts and contexts on every wait** (chapters 4 and 5), so a stuck operation eventually fails loudly instead of hanging forever.
- **Goroutine dumps** (`SIGQUIT`, or the `goroutine` profile) to see what everything is waiting on.
- **The `goroutineleak` profile**, which finds goroutines blocked on something unreachable (final chapter).
- **Tests that detect hangs**, which `testing/synctest` makes easy, also in the final chapter.
