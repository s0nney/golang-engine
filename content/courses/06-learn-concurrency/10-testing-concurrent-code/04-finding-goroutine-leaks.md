---
title: Finding Goroutine Leaks
quiz:
  - question: What does the `goroutineleak` profile report?
    options:
      - text: Every goroutine that has been running for more than a minute
      - text: Goroutines blocked on a channel, lock or similar that no running goroutine can reach, so nothing could ever wake them
        correct: true
      - text: The number of goroutines created per second
      - text: Goroutines that are using too much memory
    explanation: |
      Collecting the profile runs a special garbage-collection cycle. A
      blocked goroutine whose channel or lock is unreachable from any
      goroutine that could still run can never wake up, so it's reported as
      leaked. A goroutine blocked on something still reachable isn't
      reported, even if it will in fact never wake.
  - question: |
      A `synctest.Test` fails with `deadlock: main bubble goroutine has exited but blocked goroutines remain`. What does it mean?
    options:
      - text: The test function deadlocked before finishing
      - text: The test function returned, but goroutines it started are still blocked, a leak, and the panic lists their stacks
        correct: true
      - text: The fake clock ran out of time
      - text: A mutex was locked when the test ended
    explanation: |
      `synctest.Test` waits for every goroutine in the bubble to exit after
      your test function returns. If some are blocked and nothing in the
      bubble can wake them, they've leaked, and the panic prints where each
      one is stuck.
---

Goroutine leaks were the first real bug in this course, and they're one of the most common in production Go. Here are two tools that catch them automatically: one in tests, one in running programs.

## In tests: synctest catches leaks for free

`synctest.Test` doesn't just run your test. When your function returns, it waits for **every goroutine in the bubble** to exit. If some are still blocked and nothing can ever wake them, it fails loudly. Here's the leaky `lookup` from chapter 2, with an unbuffered result channel:

```go
func lookup(id string) (Order, error) {
	ch := make(chan Order)
	go func() { ch <- fetchOrder(id) }() // fetchOrder takes 2s
	select {
	case o := <-ch:
		return o, nil
	case <-time.After(time.Second):
		return Order{}, errors.New("timeout")
	}
}

func TestLookupTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		if _, err := lookup("A1"); err == nil {
			t.Fatal("want a timeout error")
		}
	})
}
```

The assertion passes, since there really is a timeout. But the test fails anyway:

```text
--- FAIL: TestLookupTimeout (0.00s)
panic: deadlock: main bubble goroutine has exited but blocked goroutines remain [recovered, repanicked]
...
goroutine 10 [sleep (durable), synctest bubble 1]:
time.Sleep(0x77359400)
dispatch.fetchOrder(...)
	/home/you/dispatch/leak_test.go:13
dispatch.lookup.func1()
	/home/you/dispatch/leak_test.go:19 +0x33
created by dispatch.lookup in goroutine 9
```

The panic lists each goroutine left behind, where it's stuck and where it was started.

The fix is `make(chan Order, 1)`, so the send can always complete. The test also has to let the fetch finish: once the test function returns, the bubble's clock stops, which is why the stack above is still in `Sleep`. Add `synctest.Sleep(2 * time.Second)` after the assertion. With the buffered channel the fetch goroutine now sends and exits, and the test passes. With the unbuffered channel it's stuck sending forever, and the bubble reports the leak, every run.

Without synctest, the standard trick is to compare `runtime.NumGoroutine()` before and after, allowing some time for goroutines to exit, as the pipeline exercise's tests did.

## In production: the goroutineleak profile

Tests only cover the paths you test. For running services, Go has a **`goroutineleak` profile**. It was added as an experiment in Go 1.26 and is available by default in Go 1.27.

Collecting it runs a special garbage-collection cycle. The garbage collector already knows which memory is reachable from running goroutines. If a goroutine is blocked on a channel, mutex or similar that **no goroutine that could still run can reach**, then nothing can ever wake it up, so it's reported as leaked:

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"runtime/pprof"
	"time"
)

type Order struct{ ID string }

func fetchOrder(id string) Order {
	time.Sleep(20 * time.Millisecond) // a slow database
	return Order{ID: id}
}

func lookup(id string) (Order, error) {
	ch := make(chan Order) // unbuffered: the bug
	go func() { ch <- fetchOrder(id) }()
	select {
	case o := <-ch:
		return o, nil
	case <-time.After(5 * time.Millisecond):
		return Order{}, errors.New("timeout")
	}
}

func main() {
	for _, id := range []string{"A1", "B2", "C3"} {
		_, err := lookup(id)
		fmt.Println(id, err)
	}
	time.Sleep(50 * time.Millisecond) // let the slow fetches finish

	pprof.Lookup("goroutineleak").WriteTo(os.Stdout, 1)
}
```

```text
A1 timeout
B2 timeout
C3 timeout
goroutineleak profile: total 3
3 @ 0x47f06a 0x41421c 0x413e17 0x4e0816 0x484ec1
#	0x4e0815	main.lookup.func1+0x55	/home/you/dispatch/main.go:20
```

Three goroutines, all leaked at the same line: the send in `lookup`'s goroutine. With `WriteTo(w, 2)` you get full stack traces in the same format as a panic, each marked like `[chan send (leaked)]`.

In a real service you'd rather not add code for this. If the service imports `net/http/pprof`, the profile is served at `/debug/pprof/goroutineleak`, next to the familiar `goroutine` profile.

### What it can and can't see

- **What it reports is definitely leaked.** Nothing can ever wake those goroutines.
- **It can miss leaks.** If the channel a goroutine waits on is still referenced from somewhere live (a global, a long-lived struct, another goroutine's variables), the goroutine *might* be woken as far as the garbage collector can tell, so it isn't reported, even if your program will never actually use that channel again. In experiments the detection is also sensitive to exactly how the code captures its variables. Treat a clean profile as good news, not as proof.
- **It costs a GC cycle** each time you collect it. That's fine to do now and then on a live service, but don't collect it in a tight loop.

## From tests to production

You now have the core tools for concurrent Go:

- **Structure**: goroutines, channels, `select`, `context`, and the `sync` package.
- **Patterns**: generators, pipelines, fan-out and fan-in, worker pools, semaphores and error groups.
- **Safety nets**: the race detector for data races, `synctest` for fast deterministic tests that also catch deadlocks and leaks, and the `goroutineleak` profile for what slips through to production.

The habits matter more than any single API: know how every goroutine will stop, watch `ctx.Done()` wherever you wait, keep locks short, and test with fake time instead of sleeps. Dispatchly's couriers are in good hands.

One job left: in the next lesson you'll assemble these pieces into Dispatchly's batch engine and test it with fake time.
