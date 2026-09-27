---
title: 'Build It: Dispatchly Batch Engine'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"fmt"
    )

    func dispatch(ctx context.Context, ids []int, workers int, deliver func(context.Context, int) error) error {
    	return nil
    }

    func main() {
    	err := dispatch(context.Background(), []int{1, 2, 3}, 2, func(ctx context.Context, id int) error {
    		fmt.Println("delivered", id)
    		return nil
    	})
    	fmt.Println("batch:", err)
    }
  solution: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"sync"
    )

    func dispatch(ctx context.Context, ids []int, workers int, deliver func(context.Context, int) error) error {
    	if workers < 1 { return errors.New("workers must be positive") }
    	ctx, cancel := context.WithCancelCause(ctx)
    	defer cancel(nil)
    	jobs := make(chan int)
    	var wg sync.WaitGroup
    	for range min(workers, len(ids)) {
    		wg.Go(func() {
    			for id := range jobs {
    				if ctx.Err() != nil { return }
    				if err := deliver(ctx, id); err != nil {
    					cancel(err)
    					return
    				}
    			}
    		})
    	}
    send:
    	for _, id := range ids {
    		select {
    		case jobs <- id:
    		case <-ctx.Done(): break send
    		}
    	}
    	close(jobs)
    	wg.Wait()
    	return context.Cause(ctx)
    }

    func main() {
    	err := dispatch(context.Background(), []int{1, 2, 3}, 2, func(ctx context.Context, id int) error {
    		fmt.Println("delivered", id)
    		return nil
    	})
    	fmt.Println("batch:", err)
    }
  tests: |
    package main

    import (
    	"context"
    	"errors"
    	"sync"
    	"sync/atomic"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    func TestBatch(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		var mu sync.Mutex
    		seen := make(map[int]int)
    		var active, peak atomic.Int32
    		err := dispatch(t.Context(), []int{1,2,3,4,5,6}, 2, func(ctx context.Context, id int) error {
    			n := active.Add(1)
    			defer active.Add(-1)
    			for p := peak.Load(); n > p && !peak.CompareAndSwap(p, n); p = peak.Load() {}
    			time.Sleep(time.Second)
    			mu.Lock()
    			seen[id]++
    			mu.Unlock()
    			return nil
    		})
    		if err != nil || peak.Load() != 2 || active.Load() != 0 {
    			t.Fatalf("err=%v peak=%d active=%d; want nil, 2, 0", err, peak.Load(), active.Load())
    		}
    		for _, id := range []int{1,2,3,4,5,6} {
    			if seen[id] != 1 { t.Errorf("id %d delivered %d times", id, seen[id]) }
    		}
    	})
    }

    func TestCancellationAndFailure(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		want := errors.New("courier unavailable")
    		var active atomic.Int32
    		err := dispatch(t.Context(), []int{1,2,3}, 2, func(ctx context.Context, id int) error {
    			active.Add(1)
    			defer active.Add(-1)
    			if id == 1 { time.Sleep(time.Second); return want }
    			<-ctx.Done()
    			return ctx.Err()
    		})
    		if !errors.Is(err, want) || active.Load() != 0 { t.Fatalf("err=%v active=%d", err, active.Load()) }
    		ctx, cancel := context.WithCancel(t.Context())
    		cancel()
    		err = dispatch(ctx, []int{1}, 1, func(context.Context, int) error { t.Error("called after cancellation"); return nil })
    		if !errors.Is(err, context.Canceled) { t.Errorf("got %v, want cancellation", err) }
    	})
    }

    func TestEmptyAndInvalid(t *testing.T) {
    	fn := func(context.Context, int) error { t.Error("unexpected delivery"); return nil }
    	if err := dispatch(t.Context(), nil, 2, fn); err != nil { t.Fatal(err) }
    	if err := dispatch(t.Context(), []int{1}, 0, fn); err == nil { t.Fatal("zero workers must fail") }
    }
---

Dispatchly needs one entry point for a batch of deliveries. Combine the worker pool,
cancellation causes, channel ownership, and deterministic testing from this course.

Complete `dispatch` with this contract:

- Reject a worker count below one. An empty batch with a valid, live context succeeds.
- Run up to `workers` deliveries concurrently. On a successful batch, call `deliver`
  exactly once per input entry. Order doesn't matter.
- Derive a cancellable context. A delivery failure cancels its siblings; return that
  failure as the cause. Parent cancellation also stops the batch.
- Stop scheduling once cancellation is observed and check cancellation before calling
  `deliver`. Already running callbacks must cooperate by observing their context.
- Wait for all started workers before returning, including on failure.

The callback is supplied by the caller and is safe to call concurrently. It must
return when cancelled. As with every cooperative cancellation API, you cannot force
a callback that ignores its context to stop.

Start by sketching ownership: one producer sends and closes the jobs channel;
workers consume it; the caller joins the workers. A useful building block is:

```go
ctx, cancel := context.WithCancelCause(ctx)
defer cancel(nil)
```

The first non-nil cancellation cause wins, so a sibling's `context.Canceled` must
not hide the original delivery error. A cancellation can race with a job handoff;
checking the context again before the callback keeps already-cancelled jobs from
starting. Cancellation does not undo deliveries that already succeeded.

**Run** prints each delivered ID and the batch result. **Submit** checks bounded
parallelism, exactly-once delivery on success, cancellation, and worker cleanup using
fake time. On your machine, also run the tests with `go test -race`.

You now have a batch engine with an explicit lifetime and failure contract. Next,
[Learn Testing and Tooling in Go](/courses/learn-testing) expands these checks into
table tests, test doubles, fuzzing, benchmarks, and a repeatable development workflow.
