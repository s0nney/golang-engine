---
title: 'Build It: Dispatchly Batch Engine'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"fmt"
    )

    // dispatch delivers every order ID in ids using at most workers goroutines.
    // The first delivery error cancels the rest of the batch and is returned.
    func dispatch(ctx context.Context, ids []int, workers int, deliver func(context.Context, int) error) error {
    	// TODO: validate workers, derive a cancellable context, start the
    	// workers, feed them the IDs, wait for them, and return the cause.
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

    // dispatch delivers every order ID in ids using at most workers goroutines.
    // The first delivery error cancels the rest of the batch and is returned.
    func dispatch(ctx context.Context, ids []int, workers int, deliver func(context.Context, int) error) error {
    	if workers < 1 {
    		return errors.New("workers must be positive")
    	}
    	ctx, cancel := context.WithCancelCause(ctx)
    	defer cancel(nil)

    	jobs := make(chan int)
    	var wg sync.WaitGroup
    	for range min(workers, len(ids)) {
    		wg.Go(func() {
    			for id := range jobs {
    				if ctx.Err() != nil {
    					return
    				}
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
    		case <-ctx.Done():
    			break send
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
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // tracker records how many deliveries run at once and how often each ID ran.
    type tracker struct {
    	mu     sync.Mutex
    	active int
    	peak   int
    	seen   map[int]int
    }

    func (tr *tracker) start(id int) {
    	tr.mu.Lock()
    	defer tr.mu.Unlock()
    	tr.active++
    	tr.peak = max(tr.peak, tr.active)
    	tr.seen[id]++
    }

    func (tr *tracker) stop() {
    	tr.mu.Lock()
    	defer tr.mu.Unlock()
    	tr.active--
    }

    func TestBatchRunsEachOrderOnce(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		tr := &tracker{seen: map[int]int{}}
    		ids := []int{1, 2, 3, 4, 5, 6}
    		err := dispatch(t.Context(), ids, 2, func(ctx context.Context, id int) error {
    			tr.start(id)
    			defer tr.stop()
    			time.Sleep(time.Second)
    			return nil
    		})
    		if err != nil {
    			t.Fatalf("dispatch returned %v, want nil", err)
    		}
    		if tr.peak != 2 {
    			t.Errorf("peak concurrent deliveries = %d, want 2 (workers = 2)", tr.peak)
    		}
    		if tr.active != 0 {
    			t.Errorf("%d deliveries still running after dispatch returned; wait for your workers", tr.active)
    		}
    		for _, id := range ids {
    			if tr.seen[id] != 1 {
    				t.Errorf("order %d delivered %d times, want exactly 1", id, tr.seen[id])
    			}
    		}
    	})
    }

    func TestFailureCancelsSiblings(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		want := errors.New("courier unavailable")
    		tr := &tracker{seen: map[int]int{}}
    		err := dispatch(t.Context(), []int{1, 2, 3}, 2, func(ctx context.Context, id int) error {
    			tr.start(id)
    			defer tr.stop()
    			if id == 1 {
    				time.Sleep(time.Second)
    				return want
    			}
    			<-ctx.Done() // a well-behaved delivery waits until it's cancelled
    			return ctx.Err()
    		})
    		if !errors.Is(err, want) {
    			t.Errorf("dispatch returned %v, want the delivery error %q as the cause", err, want)
    		}
    		if tr.active != 0 {
    			t.Errorf("%d deliveries still running after dispatch returned; wait for your workers", tr.active)
    		}
    		if tr.seen[3] != 0 {
    			t.Errorf("order 3 was delivered after the batch failed; stop scheduling once cancelled")
    		}
    	})
    }

    func TestParentCancellation(t *testing.T) {
    	ctx, cancel := context.WithCancel(t.Context())
    	cancel()
    	err := dispatch(ctx, []int{1, 2}, 1, func(context.Context, int) error {
    		t.Error("deliver called although the parent context was already cancelled")
    		return nil
    	})
    	if !errors.Is(err, context.Canceled) {
    		t.Errorf("dispatch with a cancelled parent returned %v, want context.Canceled", err)
    	}
    }

    func TestEmptyAndInvalid(t *testing.T) {
    	fn := func(context.Context, int) error {
    		t.Error("deliver called for an empty batch")
    		return nil
    	}
    	if err := dispatch(t.Context(), nil, 2, fn); err != nil {
    		t.Errorf("dispatch(empty batch) = %v, want nil", err)
    	}
    	if err := dispatch(t.Context(), []int{1}, 0, fn); err == nil {
    		t.Error("dispatch with 0 workers returned nil, want an error")
    	}
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
