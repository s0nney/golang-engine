---
title: GOMAXPROCS
quiz:
  - question: What does `runtime.GOMAXPROCS(0)` do?
    options:
      - text: Sets the limit to zero, so no goroutines can run
      - text: Resets the limit to the number of CPUs
      - text: Returns the current setting without changing it
        correct: true
      - text: Panics, because the argument must be at least 1
    explanation: |
      `GOMAXPROCS(n)` sets the limit and returns the *previous* value, but
      if `n < 1` it changes nothing. So passing `0` is the idiomatic way to
      ask "what's the current setting?".
  - question: |
      Dispatchly's API server runs with `GOMAXPROCS=4`. At some moment, 4 goroutines are
      crunching numbers and 3,000 goroutines are waiting for database replies. What happens?
    options:
      - text: Only 4 goroutines can exist at a time, so the other 2,996 are never created
      - text: Everything is fine. Waiting goroutines don't use any of the 4 slots
        correct: true
      - text: The program deadlocks, because every slot is taken
      - text: The runtime raises GOMAXPROCS to 3,004 automatically
    explanation: |
      GOMAXPROCS limits how many threads *execute Go code* at the same time.
      A goroutine blocked on a channel, a timer or network I/O isn't executing
      anything, so it doesn't take a slot. You can have millions of goroutines
      with GOMAXPROCS=4.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"runtime"
    	"sync/atomic"
    )

    // sumOrders splits totals into at most `parts` chunks, sums each chunk
    // in its own goroutine with sumChunk, and returns the grand total.
    func sumOrders(totals []int, parts int, sumChunk func([]int) int) int {
    	// TODO: split the work into chunks and sum them concurrently.
    	return sumChunk(totals)
    }

    func main() {
    	totals := make([]int, 1_000_000)
    	for i := range totals {
    		totals[i] = 500 + i%1000 // cents
    	}

    	var calls atomic.Int64
    	sumChunk := func(chunk []int) int {
    		calls.Add(1)
    		sum := 0
    		for _, t := range chunk {
    			sum += t
    		}
    		return sum
    	}

    	parts := runtime.GOMAXPROCS(0)
    	fmt.Println("total cents:", sumOrders(totals, parts, sumChunk))
    	fmt.Println("chunks:", calls.Load(), "(at most", parts, "wanted)")
    }
  solution: |
    package main

    import (
    	"fmt"
    	"runtime"
    	"sync"
    	"sync/atomic"
    )

    // sumOrders splits totals into at most `parts` chunks, sums each chunk
    // in its own goroutine with sumChunk, and returns the grand total.
    func sumOrders(totals []int, parts int, sumChunk func([]int) int) int {
    	size := (len(totals) + parts - 1) / parts
    	sums := make([]int, parts)
    	var wg sync.WaitGroup
    	for i := range parts {
    		start := i * size
    		if start >= len(totals) {
    			break
    		}
    		end := min(start+size, len(totals))
    		wg.Go(func() {
    			sums[i] = sumChunk(totals[start:end])
    		})
    	}
    	wg.Wait()

    	total := 0
    	for _, s := range sums {
    		total += s
    	}
    	return total
    }

    func main() {
    	totals := make([]int, 1_000_000)
    	for i := range totals {
    		totals[i] = 500 + i%1000 // cents
    	}

    	var calls atomic.Int64
    	sumChunk := func(chunk []int) int {
    		calls.Add(1)
    		sum := 0
    		for _, t := range chunk {
    			sum += t
    		}
    		return sum
    	}

    	parts := runtime.GOMAXPROCS(0)
    	fmt.Println("total cents:", sumOrders(totals, parts, sumChunk))
    	fmt.Println("chunks:", calls.Load(), "(at most", parts, "wanted)")
    }
  tests: |
    package main

    import (
    	"fmt"
    	"sync"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    func orders(n int) []int {
    	totals := make([]int, n)
    	for i := range totals {
    		totals[i] = 1 << (i % 20) // distinct-ish values, so a missed or doubled chunk changes the sum
    	}
    	return totals
    }

    func TestSumOrders(t *testing.T) {
    	for _, tc := range []struct{ n, parts int }{
    		{10, 4}, {9, 4}, {8, 3}, {100, 1}, {2, 4}, {1, 8}, {0, 3}, {1000, 12},
    	} {
    		t.Run(fmt.Sprintf("%d orders %d parts", tc.n, tc.parts), func(t *testing.T) {
    			synctest.Test(t, func(t *testing.T) {
    				totals := orders(tc.n)
    				want := 0
    				for _, v := range totals {
    					want += v
    				}
    				maxSize := (tc.n + tc.parts - 1) / tc.parts
    				var mu sync.Mutex
    				var sizes []int
    				sumChunk := func(chunk []int) int {
    					mu.Lock()
    					sizes = append(sizes, len(chunk))
    					mu.Unlock()
    					time.Sleep(time.Second) // the expensive part
    					sum := 0
    					for _, v := range chunk {
    						sum += v
    					}
    					return sum
    				}
    				start := time.Now()
    				got := sumOrders(totals, tc.parts, sumChunk)
    				took := time.Since(start)
    				if got != want {
    					t.Errorf("sumOrders(%d orders, %d parts) = %d, want %d", tc.n, tc.parts, got, want)
    				}
    				if len(sizes) > tc.parts {
    					t.Errorf("sumOrders(%d orders, %d parts) made %d chunks, want at most %d", tc.n, tc.parts, len(sizes), tc.parts)
    				}
    				for _, size := range sizes {
    					if size == 0 {
    						t.Errorf("sumOrders(%d orders, %d parts) passed an empty chunk to sumChunk", tc.n, tc.parts)
    					}
    					if size > maxSize {
    						t.Errorf("sumOrders(%d orders, %d parts) made a chunk of %d orders, want at most %d: split the work up", tc.n, tc.parts, size, maxSize)
    					}
    				}
    				if tc.n > 0 && took != time.Second {
    					t.Errorf("sumOrders(%d orders, %d parts) took %v with 1s per chunk, want 1s: are the chunks summed concurrently?", tc.n, tc.parts, took)
    				}
    			})
    		})
    	}
    }
---

In the last lesson you saw that concurrent code *may* run in parallel. How much parallelism does a Go program actually get? That's controlled by one number: **GOMAXPROCS**.

## What it means

GOMAXPROCS is the maximum number of operating-system threads that can be **executing Go code at the same instant**. If it's 8, at most 8 goroutines are running at any moment, one per core. All the others are either waiting for their turn or blocked on something (a channel, a timer, the network).

Note what it does *not* limit:

- the number of goroutines. You can have a million.
- threads blocked in system calls. Those don't count against the limit.

## Reading the current value

```go
package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Println("logical CPUs:", runtime.NumCPU())
	fmt.Println("GOMAXPROCS:  ", runtime.GOMAXPROCS(0))
}
```

On an 8-core laptop this prints something like:

```text
logical CPUs: 8
GOMAXPROCS:   8
```

`runtime.GOMAXPROCS(n)` sets the value and returns the *old* one, but when `n < 1` it only reports the current value. That's why you'll see `GOMAXPROCS(0)` used as a getter.

## Where the default comes from

Modern Go picks a sensible default for you. It's the smallest of:

- the number of logical CPUs on the machine,
- the CPUs your process is allowed to use (its *affinity mask*),
- on Linux, the container's CPU limit (the cgroup quota), rounded up to a whole number. (A CPU limit on its own never pushes the default below 2.)

The container rule arrived in Go 1.25 (it applies when your `go.mod` says `go 1.25` or later), and it matters for Dispatchly. Our services run in Kubernetes pods with a 2-CPU limit on 64-core machines. Older Go versions would have set GOMAXPROCS to 64, then had the kernel throttle the process for using more CPU than allowed. Now the default is 2. The runtime also re-checks these values periodically and **updates the default** if they change.

## Overriding it

You can override the default in two ways:

```text
$ GOMAXPROCS=2 ./dispatchly   # environment variable
```

```go
runtime.GOMAXPROCS(2) // from code
```

Setting a custom value turns off the automatic updates. `runtime.SetDefaultGOMAXPROCS()` restores the default behaviour.

In practice you almost never need to touch it. The default is good, and the right fix for "my program is slow" is almost never "change GOMAXPROCS".

## Seeing it in action

Here's CPU-heavy work (counting primes the slow way), split across four goroutines. We run it twice, first limited to one thread and then to four:

```go
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func countPrimes(from, to int) int {
	n := 0
	for x := max(from, 2); x < to; x++ {
		prime := true
		for d := 2; d*d <= x; d++ {
			if x%d == 0 {
				prime = false
				break
			}
		}
		if prime {
			n++
		}
	}
	return n
}

func run(procs int) {
	runtime.GOMAXPROCS(procs)
	start := time.Now()
	var wg sync.WaitGroup
	for i := range 4 {
		wg.Go(func() { countPrimes(i*2_000_000, (i+1)*2_000_000) })
	}
	wg.Wait()
	fmt.Printf("GOMAXPROCS=%d: %v\n", procs, time.Since(start).Round(time.Millisecond))
}

func main() {
	run(1)
	run(4)
}
```

On a machine with at least four cores you'll see something like:

```text
GOMAXPROCS=1: 1.919s
GOMAXPROCS=4: 609ms
```

With one thread, the four goroutines take turns on a single core: concurrent, but not parallel. With four threads they really do run at the same time. For CPU-bound work like this, parallelism is the only thing that helps, and it's capped by your core count. You'll dig into that in the lesson after next, and put it to work right now.

## Your turn

Dispatchly's end-of-day report adds up every order total, and there are millions of them. That's CPU-bound work, so the way to speed it up is to split it into a few big **chunks**, one per core, and sum the chunks in parallel.

Complete `sumOrders(totals, parts, sumChunk)`:

- Split `totals` into **at most `parts`** contiguous chunks, each with at most `ceil(len(totals) / parts)` elements. In Go that's `(len(totals) + parts - 1) / parts`. Never pass an empty chunk.
- Call `sumChunk` on every chunk, each in its own goroutine (use `wg.Go`), and return the grand total.

`sumChunk` is passed in so the tests can check the chunking and time it. Hint: give each goroutine its own slot in a `sums` slice, then add the slots up after `wg.Wait()`. You can assume `parts >= 1`.

## Further reading

- [`runtime.GOMAXPROCS` documentation](https://pkg.go.dev/runtime#GOMAXPROCS)
- [Container-aware GOMAXPROCS](https://go.dev/blog/container-aware-gomaxprocs) on the Go blog
