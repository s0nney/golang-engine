---
title: Ping Every Courier
difficulty: easy
after: goroutines-and-waitgroups
hints:
  - 'Start one goroutine per courier, and use a `sync.WaitGroup` so `pingAll` doesn''t return until every one of them has finished.'
  - '`wg.Go(func() { ... })` (Go 1.25+) does the `Add(1)` and the `defer Done()` for you. Since Go 1.22 each loop iteration has its own `c`, so the closure can use it directly.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"sync"
    	"time"
    )

    // pingAll calls ping once for every courier in couriers. The calls run
    // concurrently, and pingAll returns only once every call has finished.
    func pingAll(couriers []string, ping func(courier string)) {
    	// 1. Create a sync.WaitGroup.
    	// 2. Start one goroutine per courier that calls ping(courier).
    	// 3. Wait for all of them before returning.
    	_ = sync.WaitGroup{}
    }

    func main() {
    	var mu sync.Mutex
    	var pinged []string
    	ping := func(courier string) {
    		time.Sleep(100 * time.Millisecond) // a slow push notification
    		mu.Lock()
    		defer mu.Unlock()
    		pinged = append(pinged, courier)
    	}

    	start := time.Now()
    	pingAll([]string{"ana", "ben", "cy", "dee"}, ping)
    	fmt.Printf("pinged %d couriers in %v\n", len(pinged), time.Since(start).Round(50*time.Millisecond))
    	// want: pinged 4 couriers in 100ms
    }
  solution: |
    package main

    import (
    	"fmt"
    	"sync"
    	"time"
    )

    // pingAll calls ping once for every courier in couriers. The calls run
    // concurrently, and pingAll returns only once every call has finished.
    func pingAll(couriers []string, ping func(courier string)) {
    	var wg sync.WaitGroup
    	for _, c := range couriers {
    		wg.Go(func() { ping(c) })
    	}
    	wg.Wait()
    }

    func main() {
    	var mu sync.Mutex
    	var pinged []string
    	ping := func(courier string) {
    		time.Sleep(100 * time.Millisecond) // a slow push notification
    		mu.Lock()
    		defer mu.Unlock()
    		pinged = append(pinged, courier)
    	}

    	start := time.Now()
    	pingAll([]string{"ana", "ben", "cy", "dee"}, ping)
    	fmt.Printf("pinged %d couriers in %v\n", len(pinged), time.Since(start).Round(50*time.Millisecond))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"slices"
    	"sync"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // recorder is a ping func that takes a second (of fake time) per call.
    type recorder struct {
    	mu     sync.Mutex
    	pinged []string
    }

    func (r *recorder) ping(courier string) {
    	time.Sleep(time.Second)
    	r.mu.Lock()
    	defer r.mu.Unlock()
    	r.pinged = append(r.pinged, courier)
    }

    func (r *recorder) sorted() []string {
    	r.mu.Lock()
    	defer r.mu.Unlock()
    	return slices.Sorted(slices.Values(r.pinged))
    }

    func TestPingAll(t *testing.T) {
    	for _, couriers := range [][]string{
    		{"ana", "ben", "cy", "dee"},
    		{"solo"},
    		{"ana", "ana", "ben"},
    	} {
    		synctest.Test(t, func(t *testing.T) {
    			r := &recorder{}
    			start := time.Now()
    			pingAll(couriers, r.ping)
    			took := time.Since(start)

    			if got, want := r.sorted(), slices.Sorted(slices.Values(couriers)); !slices.Equal(got, want) {
    				t.Errorf("pingAll(%q): when it returned, pinged %q, want %q (wait for every goroutine before returning)", couriers, got, want)
    			}
    			if took != time.Second {
    				t.Errorf("pingAll(%q) with pings taking 1s each took %v, want 1s: the pings must run concurrently", couriers, took)
    			}
    			synctest.Sleep(time.Minute) // let any stragglers finish
    			if got := len(r.sorted()); got != len(couriers) {
    				t.Errorf("pingAll(%q) called ping %d times, want %d", couriers, got, len(couriers))
    			}
    		})
    	}
    }

    func TestPingAllEmpty(t *testing.T) {
    	calls := 0
    	pingAll(nil, func(string) { calls++ })
    	if calls != 0 {
    		t.Errorf("pingAll(nil) called ping %d times, want 0", calls)
    	}
    }

    func TestPingAllMany(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		couriers := make([]string, 500)
    		for i := range couriers {
    			couriers[i] = fmt.Sprint("courier-", i)
    		}
    		r := &recorder{}
    		start := time.Now()
    		pingAll(couriers, r.ping)
    		if got := len(r.sorted()); got != 500 {
    			t.Errorf("pingAll(500 couriers) pinged %d, want 500", got)
    		}
    		if took := time.Since(start); took != time.Second {
    			t.Errorf("pingAll(500 couriers) took %v, want 1s", took)
    		}
    	})
    }
---

Lunch rush at Dispatchly: before assigning a new batch of orders, the app
**pings every courier** to check they're still online. Each ping is a slow
push notification, so sending them one at a time takes far too long.

Complete `pingAll(couriers, ping)`:

- Call `ping(courier)` once for **every** courier in `couriers`, each in its
  own goroutine, so the pings run at the same time.
- Return only once **every** ping has finished.

## Example

```go
pingAll([]string{"ana", "ben", "cy", "dee"}, ping) // each ping takes 100ms
// returns after about 100ms, not 400ms, with all four couriers pinged
```

## Constraints

- `couriers` may be empty, and may list the same courier twice (ping them
  twice).
- The tests run in a `testing/synctest` bubble, so time is fake and exact: with
  pings that take 1s each, `pingAll` must take exactly 1s, whether there are 4
  couriers or 500.
