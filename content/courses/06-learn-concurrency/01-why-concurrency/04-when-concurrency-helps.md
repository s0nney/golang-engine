---
title: When Concurrency Helps
quiz:
  - question: Which of these Dispatchly jobs will get the **biggest** speed-up from running its tasks in goroutines on a 4-core machine?
    options:
      - text: Resizing 10,000 menu photos, which is pure CPU work
      - text: Calling 200 restaurant APIs that each take about 300ms to reply
        correct: true
      - text: Adding up 1,000 order totals in a slice
      - text: Parsing one 2 KB JSON config file
    explanation: |
      The API calls are I/O-bound: nearly all the time is spent waiting, and
      waits overlap no matter how many cores you have, so 200 calls can take
      about as long as one. Photo resizing is CPU-bound and can speed up at most
      about 4x. The last two jobs are so tiny that goroutine overhead would
      outweigh any gain.
  - question: 'Amdahl''s law says: if 50% of a job must run sequentially, what is the best possible speed-up with unlimited cores?'
    options:
      - text: 2x
        correct: true
      - text: 4x
      - text: 50x
      - text: Unlimited
    explanation: |
      Even if the parallel half took zero time, the sequential half still
      takes 50% of the original time. The job can never run more than twice
      as fast.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"time"
    )

    type Menu struct {
    	Restaurant string
    	Items      int
    }

    // fetchMenus fetches the menu of every restaurant in ids concurrently.
    // The result must be in the same order as ids.
    func fetchMenus(ids []string, fetch func(id string) Menu) []Menu {
    	menus := make([]Menu, len(ids))
    	for i, id := range ids {
    		menus[i] = fetch(id)
    	}
    	return menus
    }

    func main() {
    	fetch := func(id string) Menu {
    		time.Sleep(50 * time.Millisecond) // a slow restaurant API
    		return Menu{Restaurant: id, Items: len(id)}
    	}

    	start := time.Now()
    	menus := fetchMenus([]string{"pho-king", "wok-this-way", "thai-tanic"}, fetch)
    	fmt.Println(menus)
    	fmt.Println("took", time.Since(start).Round(50*time.Millisecond))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"sync"
    	"time"
    )

    type Menu struct {
    	Restaurant string
    	Items      int
    }

    func fetchMenus(ids []string, fetch func(id string) Menu) []Menu {
    	menus := make([]Menu, len(ids))
    	var wg sync.WaitGroup
    	for i, id := range ids {
    		wg.Go(func() {
    			menus[i] = fetch(id)
    		})
    	}
    	wg.Wait()
    	return menus
    }

    func main() {
    	fetch := func(id string) Menu {
    		time.Sleep(50 * time.Millisecond)
    		return Menu{Restaurant: id, Items: len(id)}
    	}

    	start := time.Now()
    	menus := fetchMenus([]string{"pho-king", "wok-this-way", "thai-tanic"}, fetch)
    	fmt.Println(menus)
    	fmt.Println("took", time.Since(start).Round(50*time.Millisecond))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"slices"
    	"sync/atomic"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    func TestFetchMenusOrder(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		ids := []string{"a", "bb", "ccc", "dddd", "eeeee"}
    		var calls atomic.Int64
    		fetch := func(id string) Menu {
    			calls.Add(1)
    			// Later restaurants answer faster, so finishing order differs from ids order.
    			time.Sleep(time.Duration(10-len(id)) * time.Second)
    			return Menu{Restaurant: id, Items: len(id)}
    		}
    		got := fetchMenus(ids, fetch)
    		var want []Menu
    		for _, id := range ids {
    			want = append(want, Menu{Restaurant: id, Items: len(id)})
    		}
    		if !slices.Equal(got, want) {
    			t.Errorf("fetchMenus(%q) = %v, want %v (same order as ids)", ids, got, want)
    		}
    		if n := calls.Load(); n != int64(len(ids)) {
    			t.Errorf("fetch was called %d times, want %d (once per restaurant)", n, len(ids))
    		}
    	})
    }

    func TestFetchMenusConcurrent(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		// Inside synctest, time is fake: each fetch "takes" 1s instantly.
    		var ids []string
    		for i := range 20 {
    			ids = append(ids, fmt.Sprint("restaurant-", i))
    		}
    		fetch := func(id string) Menu {
    			time.Sleep(time.Second)
    			return Menu{Restaurant: id}
    		}
    		start := time.Now()
    		fetchMenus(ids, fetch)
    		if took := time.Since(start); took != time.Second {
    			t.Errorf("fetching 20 menus that each take 1s took %v, want 1s: are the fetches running concurrently?", took)
    		}
    	})
    }

    func TestFetchMenusEmpty(t *testing.T) {
    	got := fetchMenus(nil, func(string) Menu { return Menu{} })
    	if len(got) != 0 {
    		t.Errorf("fetchMenus(nil) = %v, want an empty slice", got)
    	}
    }
---

Concurrency isn't a magic "go faster" button. Sometimes it makes a program dramatically faster, sometimes a little faster, and sometimes *slower*. The deciding question is: **what is the program waiting for?**

## I/O-bound work: concurrency shines

A task is **I/O-bound** when it spends most of its time waiting for input/output: network calls, database queries, disk reads. During that wait the CPU sits idle.

Dispatchly's "menu refresh" job calls 200 restaurant APIs, each taking about 300ms. Sequentially, that's a full minute. Concurrently, all 200 waits overlap, and the job takes roughly as long as the *slowest* call. It doesn't matter if you have 1 core or 64, because waiting doesn't need a core.

## CPU-bound work: limited by your cores

A task is **CPU-bound** when it spends its time computing: resizing images, hashing passwords, running route optimisation. Here goroutines only help if they run in **parallel**, so the speed-up is capped by GOMAXPROCS (normally your core count). Eight cores give you at most 8x, and usually less, because:

- the goroutines compete for memory bandwidth and caches,
- the work may not split evenly (one goroutine gets the hard part),
- some of the job can't be split at all.

That last point has a name. **Amdahl's law** says that if a fraction of a job must run sequentially, that fraction limits the total speed-up. If 10% of a job is sequential, then even with a thousand cores the best you can do is 10x.

## Tiny work: concurrency hurts

Starting a goroutine, and the synchronisation needed to collect its result, costs something on the order of a microsecond. If the work itself takes nanoseconds, the overhead dominates:

```go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	totals := make([]int64, 100_000)
	for i := range totals {
		totals[i] = int64(i % 50)
	}

	// Sequential: a tight loop.
	var sum int64
	for _, t := range totals {
		sum += t
	}

	// One goroutine per order: correct, but much slower.
	var total atomic.Int64
	var wg sync.WaitGroup
	for _, t := range totals {
		wg.Go(func() { total.Add(t) })
	}
	wg.Wait()

	fmt.Println(sum == total.Load())
}
```

```text
true
```

Both give the same answer, but the second version starts 100,000 goroutines to do 100,000 additions, and it runs far slower than the plain loop. (You'll meet `atomic.Int64` properly in chapter 6.) If you do want parallelism for a big CPU job, split it into a few large **chunks**, roughly one per core, not one goroutine per item.

## Rules of thumb

- **Waiting on many independent things?** Use goroutines, one per task is fine.
- **Heavy computation on big data?** Split into about GOMAXPROCS chunks. Measure!
- **Tiny, fast work?** Keep it sequential.
- **Tasks depend on each other's results?** They can't overlap much. Rethink the design first.

And always **measure** before and after. Go's benchmarks (`b.Loop()`) make that easy.

## Your turn

Dispatchly's app shows menus from several restaurants at once. `fetchMenus` fetches them one after another, so five restaurants that each take a second to reply keep the customer waiting five seconds.

Change `fetchMenus` so it calls `fetch` for every restaurant **concurrently**, waits for all of them, and returns the menus **in the same order as `ids`**. Use a `sync.WaitGroup` and its `Go` method.

Hint: each goroutine can write into its own index of a pre-sized slice. Writing to *different* elements of a slice from different goroutines is safe. It's only writes to the *same* memory that race.

The tests use `testing/synctest`, which gives them a fake clock so they can check your timing exactly without really waiting. You'll learn how that works in the final chapter.
