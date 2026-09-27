---
title: WaitGroup.Go vs Add and Done
quiz:
  - question: |
      What's the bug here?

      ```go
      var wg sync.WaitGroup
      for _, o := range orders {
      	go func() {
      		wg.Add(1)
      		defer wg.Done()
      		dispatch(o)
      	}()
      }
      wg.Wait()
      ```
    options:
      - text: '`defer wg.Done()` should be `wg.Done()` at the end of the function'
      - text: '`wg.Add(1)` runs inside the goroutine, so `Wait` may return before some goroutines have even called `Add`'
        correct: true
      - text: '`o` is shared between all goroutines'
      - text: There is no bug
    explanation: |
      `Add` must happen *before* `Wait` can observe the counter. If `main`
      reaches `Wait` while the counter is still 0, `Wait` returns immediately
      and orders go undispatched. Call `Add` before the `go` statement, or
      better, use `wg.Go`, which gets this right for you. `go vet` also reports
      "WaitGroup.Add called from inside new goroutine".
  - question: |
      This program prints `dispatching A1` and then crashes with "all goroutines are asleep - deadlock!". Why?

      ```go
      func dispatch(o string, wg sync.WaitGroup) {
      	defer wg.Done()
      	fmt.Println("dispatching", o)
      }

      func main() {
      	var wg sync.WaitGroup
      	wg.Add(1)
      	go dispatch("A1", wg)
      	wg.Wait()
      }
      ```
    options:
      - text: '`fmt.Println` can''t be called from a goroutine'
      - text: '`wg` is passed by value, so `Done` decrements a copy and the original counter never reaches zero'
        correct: true
      - text: '`Add(1)` should be `Add(0)`'
      - text: The goroutine never starts
    explanation: |
      A `WaitGroup` must never be copied after first use. `dispatch` gets
      its own copy, so its `Done` has no effect on the `wg` that `main` is
      waiting on. Pass `*sync.WaitGroup`, or avoid passing it at all by
      using `wg.Go` with a closure. `go vet`'s copylocks check catches this.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"time"
    )

    var errNoCourier = errors.New("no courier nearby")

    // dispatchAll dispatches every order concurrently and waits for all of
    // them. It returns nil if every dispatch succeeded. Otherwise it returns
    // the failures joined with errors.Join, in the same order as orders, each
    // wrapped as "order #<index> (<id>): <error>".
    func dispatchAll(orders []string, dispatch func(order string) error) error {
    	var errs []error
    	for i, o := range orders {
    		if err := dispatch(o); err != nil {
    			errs = append(errs, fmt.Errorf("order #%d (%s): %w", i, o, err))
    		}
    	}
    	return errors.Join(errs...)
    }

    func main() {
    	dispatch := func(order string) error {
    		time.Sleep(20 * time.Millisecond) // find a courier
    		if order == "C3" || order == "E5" {
    			return errNoCourier
    		}
    		return nil
    	}

    	start := time.Now()
    	err := dispatchAll([]string{"A1", "B2", "C3", "D4", "E5"}, dispatch)
    	fmt.Println(err)
    	fmt.Println("took", time.Since(start).Round(20*time.Millisecond))
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"sync"
    	"time"
    )

    var errNoCourier = errors.New("no courier nearby")

    // dispatchAll dispatches every order concurrently and waits for all of
    // them. It returns nil if every dispatch succeeded. Otherwise it returns
    // the failures joined with errors.Join, in the same order as orders, each
    // wrapped as "order #<index> (<id>): <error>".
    func dispatchAll(orders []string, dispatch func(order string) error) error {
    	errs := make([]error, len(orders))
    	var wg sync.WaitGroup
    	for i, o := range orders {
    		wg.Go(func() {
    			if err := dispatch(o); err != nil {
    				errs[i] = fmt.Errorf("order #%d (%s): %w", i, o, err)
    			}
    		})
    	}
    	wg.Wait()
    	return errors.Join(errs...)
    }

    func main() {
    	dispatch := func(order string) error {
    		time.Sleep(20 * time.Millisecond) // find a courier
    		if order == "C3" || order == "E5" {
    			return errNoCourier
    		}
    		return nil
    	}

    	start := time.Now()
    	err := dispatchAll([]string{"A1", "B2", "C3", "D4", "E5"}, dispatch)
    	fmt.Println(err)
    	fmt.Println("took", time.Since(start).Round(20*time.Millisecond))
    }
  tests: |
    package main

    import (
    	"errors"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    var errClosed = errors.New("restaurant closed")

    func TestDispatchAllErrorsInOrder(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		orders := []string{"A1", "B2", "C3", "D4", "E5", "F6"}
    		dispatch := func(o string) error {
    			// Later orders finish first, so completion order differs from input order.
    			time.Sleep(time.Duration(8-int(o[0]-'A')) * time.Second)
    			switch o {
    			case "B2":
    				return errClosed
    			case "E5", "F6":
    				return errNoCourier
    			}
    			return nil
    		}

    		start := time.Now()
    		err := dispatchAll(orders, dispatch)
    		took := time.Since(start)

    		want := "order #1 (B2): restaurant closed\norder #4 (E5): no courier nearby\norder #5 (F6): no courier nearby"
    		if err == nil || err.Error() != want {
    			t.Errorf("dispatchAll error =\n%v\nwant\n%s", err, want)
    		}
    		if !errors.Is(err, errNoCourier) || !errors.Is(err, errClosed) {
    			t.Errorf("errors.Is can't find the original errors in %v: wrap them with %%w", err)
    		}
    		if took != 8*time.Second {
    			t.Errorf("dispatching 6 orders took %v, want 8s (the slowest one): are they running concurrently?", took)
    		}
    	})
    }

    func TestDispatchAllSucceeds(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		calls := make(chan string, 10)
    		err := dispatchAll([]string{"A1", "B2", "C3"}, func(o string) error {
    			calls <- o
    			time.Sleep(time.Second)
    			return nil
    		})
    		if err != nil {
    			t.Errorf("dispatchAll with no failures = %v, want nil", err)
    		}
    		if len(calls) != 3 {
    			t.Errorf("dispatch was called %d times, want 3 (once per order)", len(calls))
    		}
    	})
    }

    func TestDispatchAllNoOrders(t *testing.T) {
    	if err := dispatchAll(nil, func(string) error { return errClosed }); err != nil {
    		t.Errorf("dispatchAll(nil) = %v, want nil", err)
    	}
    }
---

A `sync.WaitGroup` waits for a group of goroutines to finish. It's a counter: when it hits zero, everyone blocked in `Wait` is released. Go gives you two ways to drive that counter.

## The classic way: `Add` and `Done`

```go
var wg sync.WaitGroup
for _, o := range orders {
	wg.Add(1) // one more task to wait for
	go func() {
		defer wg.Done() // this task is finished
		dispatch(o)
	}()
}
wg.Wait() // block until the counter is 0
```

Three rules keep this correct:

1. **`Add` before `go`.** Call `Add` in the goroutine that will call `Wait`, before starting the new goroutine. If the new goroutine calls `Add` itself, `Wait` might run first, see a zero counter, and return early.
2. **`defer wg.Done()`.** Deferring guarantees `Done` runs on every return path, including panics that get recovered. A missed `Done` means `Wait` blocks forever.
3. **Never copy a WaitGroup.** Pass `*sync.WaitGroup` if a function needs it. A copy has its own counter, so the original never reaches zero.

If you know the count up front, you can `Add` it in one go: `wg.Add(len(orders))`.

## The modern way: `wg.Go`

Go 1.25 added `WaitGroup.Go`, which bundles all three rules into one call:

```go
var wg sync.WaitGroup
for _, o := range orders {
	wg.Go(func() {
		dispatch(o)
	})
}
wg.Wait()
```

`wg.Go(f)` increments the counter, starts `f` in a new goroutine, and decrements the counter when `f` returns. You can't put `Add` in the wrong place or forget `Done`, and since the closure captures `wg` directly, you never pass it around by value.

**Prefer `wg.Go` in new code.** In fact, `go fix` includes a modernizer that rewrites the `Add(1)` / `go func` / `defer Done()` pattern into `wg.Go` for you. You'll still meet `Add` and `Done` in older code, and they're useful when the goroutine is started by code you don't control.

One caveat from the documentation: the function passed to `Go` must not panic. If it might, recover inside it, as you saw in the last lesson.

## Getting results out

A WaitGroup only tells you *that* the goroutines finished, not *what* they produced. A simple, race-free approach is to give each goroutine its own slot in a pre-sized slice:

```go
package main

import (
	"errors"
	"fmt"
	"sync"
)

func dispatch(order string) error {
	if order == "" {
		return errors.New("empty order ID")
	}
	return nil
}

func main() {
	orders := []string{"A1", "", "C3", ""}

	errs := make([]error, len(orders))
	var wg sync.WaitGroup
	for i, o := range orders {
		wg.Go(func() {
			if err := dispatch(o); err != nil {
				errs[i] = fmt.Errorf("order #%d: %w", i, err)
			}
		})
	}
	wg.Wait()

	if err := errors.Join(errs...); err != nil {
		fmt.Println("some dispatches failed:")
		fmt.Println(err)
	}
}
```

```text
some dispatches failed:
order #1: empty order ID
order #3: empty order ID
```

Each goroutine writes only `errs[i]`, so no two goroutines touch the same memory. `errors.Join` skips `nil` errors and returns `nil` if they're all `nil`. After `wg.Wait()` returns, it's safe for `main` to read everything the goroutines wrote: `Wait` *synchronizes* with every `Done`, so their writes are visible.

## Reusing a WaitGroup

A few more rules from the documentation, in plain words:

- **Several goroutines may call `Wait`** on the same WaitGroup. They're all released when the counter hits zero.
- **A running task may start more tasks** (call `wg.Go` or `wg.Add`) while the counter is still above zero. The count can't have reached zero yet, because the task doing the adding hasn't finished.
- **Reusing it for a new batch** is fine, but only after every `Wait` for the previous batch has returned. When in doubt, declare a fresh one: the zero value is ready to use.

## Your turn

When the lunch rush starts, Dispatchly dispatches a whole batch of orders at once. `dispatchAll` does them one after another, so a batch of ten orders that each take a second keeps the last customer waiting ten seconds.

Make `dispatchAll` dispatch every order **concurrently** with `wg.Go`, wait for all of them, and return the failures joined with `errors.Join`, **in the same order as `orders`**, each wrapped as `order #<index> (<id>): <error>` (keep the `%w`). If every order succeeds it must return `nil`. The pattern from "Getting results out" above is exactly what you need.

## Further reading

- [`sync.WaitGroup` documentation](https://pkg.go.dev/sync#WaitGroup)
- [Go by Example: WaitGroups](https://gobyexample.com/waitgroups)
