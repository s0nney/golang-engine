---
title: The Go Scheduler
quiz:
  - question: In the G/M/P model, what is a **P**?
    options:
      - text: A goroutine waiting to be scheduled
      - text: An operating-system thread
      - text: A processor slot with its own run queue; a thread must hold one to run Go code
        correct: true
      - text: A physical CPU core
    explanation: |
      G is a goroutine, M is a machine (an OS thread), and P is a logical
      processor. There are exactly GOMAXPROCS Ps, and an M needs a P to
      execute goroutines. That's how GOMAXPROCS limits parallelism.
  - question: |
      A goroutine calls a blocking system call (say, reading a file) and its
      thread gets stuck in the kernel. What does the scheduler do?
    options:
      - text: Every other goroutine on that P waits until the syscall returns
      - text: It hands the P to another thread so the remaining goroutines keep running
        correct: true
      - text: It kills the goroutine and restarts it later
      - text: It raises GOMAXPROCS by one for the duration of the call
    explanation: |
      The blocked thread (M) keeps the goroutine, but it gives up its P. Another
      M picks up the P and carries on with the run queue. When the syscall
      returns, the goroutine goes back into a run queue.
  - question: |
      What does this print?

      ```go
      func main() {
      	for range 1000 {
      		go func() { time.Sleep(time.Hour) }()
      	}
      	time.Sleep(10 * time.Millisecond)
      	fmt.Println(runtime.NumGoroutine())
      }
      ```
    options:
      - text: '`1`'
      - text: '`1000`'
      - text: '`1001`'
        correct: true
      - text: It panics, because you can't have more goroutines than GOMAXPROCS
    explanation: |
      `runtime.NumGoroutine` counts every goroutine that exists, including
      `main` itself. The 1000 sleepers are all alive (just blocked), so the
      total is 1001.
---

You write `go f()` and somehow `f` runs, alongside thousands of others, on a handful of CPU cores. The piece of the Go runtime that makes this happen is the **scheduler**. You don't need to know its internals to write good concurrent code, but a mental model explains a lot of behaviour you'll see later.

## Goroutines are not threads

Operating systems schedule **threads**. Threads are fairly expensive: each one reserves a large stack (often 1 to 8 MB), and switching between them means a trip into the kernel.

Goroutines are managed by the Go runtime instead. A new goroutine starts with a tiny stack (a few KB) that grows as needed, and switching between goroutines happens in user space, which is much cheaper. So Go multiplexes many goroutines onto a few threads. This is called **M:N scheduling**: M goroutines on N threads.

## G, M and P

The scheduler juggles three kinds of object:

- **G**: a goroutine. Its stack, its instruction pointer, and what it's waiting for.
- **M**: a *machine*, which means an OS thread. Ms actually execute code.
- **P**: a *processor*, a slot for running Go code. There are exactly **GOMAXPROCS** Ps. Each P has a local **run queue** of goroutines that are ready to run.

An M must hold a P to run goroutines. That's how GOMAXPROCS limits parallelism: 4 Ps means at most 4 threads running Go code at once.

```text
          P0                 P1
      ┌────────┐         ┌────────┐
 M0 ──┤ G1 run │   M1 ───┤ G5 run │
      └────────┘         └────────┘
      run queue:         run queue:
      G2 G3 G4           G6

 global run queue: G7 G8
 blocked (no P needed): G9 on a channel, G10 in time.Sleep, G11 waiting on the network
```

## What happens when a goroutine blocks

This is where the design pays off:

- **Channel operations, mutexes, `time.Sleep`**: the G is parked (set aside) and the M immediately runs the next G from the P's queue. When something wakes the G up, it's put back in a run queue. Blocked goroutines cost almost nothing.
- **Network I/O**: Go uses a *network poller* (epoll on Linux, kqueue on macOS). A goroutine waiting on a socket is parked just like one waiting on a channel. This is why a Go server can hold 100,000 open connections with a handful of threads.
- **Blocking system calls** (such as some file operations): the M really does get stuck in the kernel. The scheduler detaches the P from it and hands the P to another M (creating a new thread if needed), so the other goroutines keep running.

## Keeping everyone busy

Two more tricks keep all the Ps busy and fair:

- **Work stealing.** When a P's run queue is empty, its M steals half the goroutines from another P's queue (or takes some from the global queue). Work spreads out automatically.
- **Preemption.** A goroutine can't hog a P forever. If one runs for more than about 10ms, the runtime interrupts it and lets someone else have a turn, even if it's stuck in a tight loop with no function calls.

You can see preemption at work. This program limits Go to a single P and starts a goroutine that spins forever. `main` still gets to run:

```go
package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	runtime.GOMAXPROCS(1)

	go func() {
		for {
			// busy-looping courier tracker that never yields
		}
	}()

	time.Sleep(20 * time.Millisecond)
	fmt.Println("main still gets a turn")
}
```

```text
main still gets a turn
```

Before Go 1.14 this program would hang forever, because the scheduler could only switch goroutines at function calls. Today the runtime preempts the spinning goroutine asynchronously.

## Counting goroutines

`runtime.NumGoroutine()` reports how many goroutines exist right now, including `main`. It's a handy debugging tool, and you'll use it later to catch goroutine leaks:

```go
package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	for range 10_000 {
		go func() { time.Sleep(time.Hour) }() // 10,000 idle couriers
	}
	time.Sleep(50 * time.Millisecond)
	fmt.Println("goroutines:", runtime.NumGoroutine())
}
```

```text
goroutines: 10001
```

Ten thousand sleeping goroutines use a few tens of megabytes at most and no CPU at all. Try that with OS threads!

## What this means for you

- Goroutines are cheap, so it's fine to start one per request, per order or per connection.
- Blocking is cheap too. Write straightforward blocking code and let the scheduler overlap the waits.
- But "cheap" isn't "free". A goroutine that blocks forever is never cleaned up. That's a **leak**, and you'll learn to spot leaks in the next chapter.

## Further reading

- [`runtime` package documentation](https://pkg.go.dev/runtime)
