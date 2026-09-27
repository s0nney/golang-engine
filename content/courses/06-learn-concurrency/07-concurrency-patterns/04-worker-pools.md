---
title: Worker Pools
quiz:
  - question: |
      What's the bug in this worker pool?

      ```go
      jobs := make(chan Order)
      var wg sync.WaitGroup
      for range 4 {
      	wg.Go(func() {
      		for o := range jobs {
      			charge(o)
      		}
      	})
      }
      for _, o := range orders {
      	jobs <- o
      }
      wg.Wait()
      ```
    options:
      - text: The workers should be started after the jobs are sent
      - text: '`jobs` is never closed, so the workers'' `range` loops never end and `wg.Wait()` blocks forever'
        correct: true
      - text: Four workers can't share one channel
      - text: '`wg.Go` can''t be used in a loop'
    explanation: |
      Workers stop when `range jobs` ends, which only happens when `jobs`
      is closed. Add `close(jobs)` after the sending loop and before
      `wg.Wait()`.
  - question: What limits how many orders are processed at once in a worker pool with N workers?
    options:
      - text: The capacity of the jobs channel
      - text: GOMAXPROCS
      - text: The number of worker goroutines, N, since each handles one job at a time
        correct: true
      - text: Nothing, all orders are processed at once
    explanation: |
      Each worker runs one job at a time, so at most N run concurrently.
      Buffering the jobs channel lets the sender get ahead, but it doesn't
      change how many jobs are *processed* at once.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"sync"
    	"time"
    )

    type Order struct {
    	ID    int
    	Cents int
    }

    type Receipt struct {
    	OrderID int
    	Charged int
    }

    // processAll runs process on every order using at most `workers` goroutines
    // at once, and returns the receipts in the same order as orders.
    func processAll(orders []Order, workers int, process func(Order) Receipt) []Receipt {
    	receipts := make([]Receipt, len(orders))
    	var wg sync.WaitGroup
    	for i, o := range orders {
    		wg.Go(func() {
    			receipts[i] = process(o)
    		})
    	}
    	wg.Wait()
    	return receipts
    }

    // gauge tracks how many payments are running now, and the peak.
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
    	var g gauge
    	charge := func(o Order) Receipt {
    		g.add(1)
    		time.Sleep(10 * time.Millisecond) // call the payment provider
    		g.add(-1)
    		return Receipt{OrderID: o.ID, Charged: o.Cents}
    	}

    	var orders []Order
    	for i := range 12 {
    		orders = append(orders, Order{ID: 100 + i, Cents: 500 + 25*i})
    	}

    	receipts := processAll(orders, 3, charge)
    	fmt.Println("first receipt:", receipts[0], "last:", receipts[len(receipts)-1])
    	fmt.Println("max payments in flight:", g.peak, "(the payment provider allows 3)")
    }
  solution: |
    package main

    import (
    	"fmt"
    	"sync"
    	"time"
    )

    type Order struct {
    	ID    int
    	Cents int
    }

    type Receipt struct {
    	OrderID int
    	Charged int
    }

    // processAll runs process on every order using at most `workers` goroutines
    // at once, and returns the receipts in the same order as orders.
    func processAll(orders []Order, workers int, process func(Order) Receipt) []Receipt {
    	receipts := make([]Receipt, len(orders))
    	jobs := make(chan int) // indexes into orders

    	var wg sync.WaitGroup
    	for range workers {
    		wg.Go(func() {
    			for i := range jobs {
    				receipts[i] = process(orders[i])
    			}
    		})
    	}

    	for i := range orders {
    		jobs <- i
    	}
    	close(jobs)
    	wg.Wait()
    	return receipts
    }

    // gauge tracks how many payments are running now, and the peak.
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
    	var g gauge
    	charge := func(o Order) Receipt {
    		g.add(1)
    		time.Sleep(10 * time.Millisecond) // call the payment provider
    		g.add(-1)
    		return Receipt{OrderID: o.ID, Charged: o.Cents}
    	}

    	var orders []Order
    	for i := range 12 {
    		orders = append(orders, Order{ID: 100 + i, Cents: 500 + 25*i})
    	}

    	receipts := processAll(orders, 3, charge)
    	fmt.Println("first receipt:", receipts[0], "last:", receipts[len(receipts)-1])
    	fmt.Println("max payments in flight:", g.peak, "(the payment provider allows 3)")
    }
  tests: |
    package main

    import (
    	"fmt"
    	"sync/atomic"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // tracker wraps a fake payment call that takes one (fake) second and
    // records the highest number of calls that were ever running at once.
    type tracker struct {
    	inFlight, max atomic.Int64
    }

    func (tr *tracker) process(o Order) Receipt {
    	n := tr.inFlight.Add(1)
    	for m := tr.max.Load(); n > m && !tr.max.CompareAndSwap(m, n); m = tr.max.Load() {
    	}
    	time.Sleep(time.Second)
    	tr.inFlight.Add(-1)
    	return Receipt{OrderID: o.ID, Charged: o.Cents * 2}
    }

    func makeOrders(n int) []Order {
    	orders := make([]Order, n)
    	for i := range orders {
    		orders[i] = Order{ID: i + 1, Cents: 100 * (i + 1)}
    	}
    	return orders
    }

    func TestBoundedConcurrency(t *testing.T) {
    	for _, tc := range []struct {
    		orders, workers int
    		wantMax         int64
    		wantTime        time.Duration
    	}{
    		{orders: 10, workers: 3, wantMax: 3, wantTime: 4 * time.Second},
    		{orders: 8, workers: 4, wantMax: 4, wantTime: 2 * time.Second},
    		{orders: 5, workers: 1, wantMax: 1, wantTime: 5 * time.Second},
    		{orders: 3, workers: 10, wantMax: 3, wantTime: 1 * time.Second},
    	} {
    		t.Run(fmt.Sprintf("%d orders %d workers", tc.orders, tc.workers), func(t *testing.T) {
    			synctest.Test(t, func(t *testing.T) {
    				var tr tracker
    				start := time.Now()
    				processAll(makeOrders(tc.orders), tc.workers, tr.process)
    				took := time.Since(start)

    				if got := tr.max.Load(); got != tc.wantMax {
    					t.Errorf("processAll(%d orders, %d workers): at most %d payments ran at once, want %d", tc.orders, tc.workers, got, tc.wantMax)
    				}
    				if took != tc.wantTime {
    					t.Errorf("processAll(%d orders, %d workers) with 1s per payment took %v, want %v", tc.orders, tc.workers, took, tc.wantTime)
    				}
    			})
    		})
    	}
    }

    func TestReceiptsInOrder(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		var tr tracker
    		orders := makeOrders(7)
    		got := processAll(orders, 3, tr.process)
    		if len(got) != len(orders) {
    			t.Fatalf("processAll returned %d receipts for %d orders", len(got), len(orders))
    		}
    		for i, o := range orders {
    			if want := (Receipt{OrderID: o.ID, Charged: o.Cents * 2}); got[i] != want {
    				t.Errorf("receipts[%d] = %v, want %v (receipts must be in the same order as orders)", i, got[i], want)
    			}
    		}
    	})
    }

    func TestNoOrders(t *testing.T) {
    	var tr tracker
    	if got := processAll(nil, 3, tr.process); len(got) != 0 {
    		t.Errorf("processAll(nil) = %v, want no receipts", got)
    	}
    }
---

"Start a goroutine for every order" works until Black Friday hits and 50,000 goroutines all open a connection to the payment provider, which allows 10. Most real systems need **bounded concurrency**: do the work in parallel, but never more than N things at once. The classic way to get it is a **worker pool**: a fixed number of goroutines pulling jobs from a shared channel.

## The shape

```go
func chargeAll(orders []Order, workers int) {
	jobs := make(chan Order)

	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for o := range jobs { // each worker takes jobs until none are left
				charge(o)
			}
		})
	}

	for _, o := range orders {
		jobs <- o
	}
	close(jobs) // no more jobs: workers finish their current one and exit
	wg.Wait()   // wait until they have
}
```

It's fan-out with a fixed number of workers. The number of goroutines, `workers`, is exactly the concurrency limit: there are never more than `workers` calls to `charge` in flight, whether there are 10 orders or 10 million.

The order of the last three steps matters:

1. **Send all jobs.** Each send blocks until a worker is free to take it, which gives you natural back-pressure.
2. **`close(jobs)`**, so every worker's `range` ends once the queue is empty.
3. **`wg.Wait()`** until all workers have returned.

Forget step 2 and step 3 waits forever.

## Collecting results

There are two common ways to get results back.

**A results channel** is best when you want to stream results as they finish. Workers send to `results`, and a closer goroutine closes it after `wg.Wait()`, exactly like `fanIn`:

```go
results := make(chan Receipt)
// ... workers send to results ...
go func() {
	wg.Wait()
	close(results)
}()
for r := range results {
	record(r)
}
```

**Writing by index** is simplest when you want a slice in input order. Send *indexes* as jobs, and have each worker write to its own slot. No two workers ever touch the same element, so there's no data race and no lock:

```go
receipts := make([]Receipt, len(orders))
jobs := make(chan int)
// worker: for i := range jobs { receipts[i] = charge(orders[i]) }
```

After `wg.Wait()`, every write is visible to the caller.

## Choosing the pool size

- **Rate-limited dependency** (payment provider, restaurant API): the provider's limit.
- **CPU-bound work**: `runtime.GOMAXPROCS(0)`. More workers only adds switching.
- **Other I/O-bound work**: start at a few times the number of cores and measure.

## Stopping early

To make a pool cancellable, add a `ctx`: the sending loop selects on `ctx.Done()` so it stops handing out jobs, and each worker checks `ctx.Err()` before starting a job. The error group in the next chapter builds on exactly that.

## Your turn

Dispatchly's payment provider allows at most a few concurrent charges per merchant account. `processAll` must charge every order using **at most `workers` concurrent calls** to `process`, and return the receipts **in the same order** as `orders`.

The starter starts one goroutine per order. Press **Run** and look at the peak. Rewrite `processAll` as a worker pool. The tests check the peak concurrency, the order of the receipts, and that the work really runs in parallel: with 10 orders, 3 workers and 1 second per charge, it should take 4 seconds, not 10. (The tests use `synctest`'s fake clock, so they run instantly.)
