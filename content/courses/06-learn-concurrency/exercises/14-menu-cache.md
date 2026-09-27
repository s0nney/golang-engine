---
title: Menu Cache
difficulty: hard
after: sync-primitives
hints:
  - 'Keep two maps under one mutex: the cached menus, and the loads in progress ("flights"). A flight holds a `done` channel that''s closed when the load finishes, the result, a count of waiting callers and the load''s cancel func. A miss either joins the existing flight or starts one in a new goroutine. Never hold the mutex while calling `load` or waiting.'
  - 'For least-recently-used eviction, `container/list` gives you an O(1) doubly linked list: `MoveToFront` on every hit, `PushFront` on insert, and remove the `Back()` when there are more than `capacity` menus. Store the `*list.Element` in the map.'
  - 'Run the load with `context.WithCancel(context.WithoutCancel(ctx))`: it keeps the first caller''s values but not its cancellation. A caller whose ctx is done decrements the waiter count and leaves; the one that brings it to zero cancels the load and removes the flight, so the next `Get` starts afresh. When a load finishes, only cache it if it succeeded and its flight is still the current one.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"fmt"
    	"sync"
    	"time"
    )

    type Menu struct {
    	Restaurant string
    	Version    int
    }

    // MenuCache caches up to capacity menus, evicting the least recently used.
    // Concurrent misses for the same restaurant share a single load.
    type MenuCache struct {
    	load func(ctx context.Context, restaurant string) (Menu, error)
    }

    func NewMenuCache(capacity int, load func(ctx context.Context, restaurant string) (Menu, error)) *MenuCache {
    	return &MenuCache{load: load}
    }

    // Get returns restaurant's menu, from the cache or by loading it.
    func (c *MenuCache) Get(ctx context.Context, restaurant string) (Menu, error) {
    	return c.load(ctx, restaurant)
    }

    // Len returns how many menus are cached.
    func (c *MenuCache) Len() int {
    	return 0
    }

    func main() {
    	var mu sync.Mutex
    	loads := 0
    	cache := NewMenuCache(2, func(ctx context.Context, r string) (Menu, error) {
    		mu.Lock()
    		loads++
    		mu.Unlock()
    		time.Sleep(20 * time.Millisecond) // a slow restaurant API
    		return Menu{r, 1}, nil
    	})

    	var wg sync.WaitGroup
    	for range 5 { // five customers open the same menu at once
    		wg.Go(func() { cache.Get(context.Background(), "Pho Real") })
    	}
    	wg.Wait()
    	m, err := cache.Get(context.Background(), "Pho Real")
    	fmt.Println(m, err, "loads:", loads, "cached:", cache.Len())
    	// want: {Pho Real 1} <nil> loads: 1 cached: 1
    }
  solution: |
    package main

    import (
    	"container/list"
    	"context"
    	"fmt"
    	"sync"
    	"time"
    )

    type Menu struct {
    	Restaurant string
    	Version    int
    }

    // flight is a load in progress.
    type flight struct {
    	done    chan struct{} // closed once menu and err are set
    	menu    Menu
    	err     error
    	waiters int
    	cancel  context.CancelFunc
    }

    type entry struct {
    	restaurant string
    	menu       Menu
    }

    // MenuCache caches up to capacity menus, evicting the least recently used.
    // Concurrent misses for the same restaurant share a single load.
    type MenuCache struct {
    	load     func(ctx context.Context, restaurant string) (Menu, error)
    	capacity int

    	mu      sync.Mutex
    	items   map[string]*list.Element // values are *entry
    	lru     *list.List               // most recently used at the front
    	flights map[string]*flight
    }

    func NewMenuCache(capacity int, load func(ctx context.Context, restaurant string) (Menu, error)) *MenuCache {
    	return &MenuCache{
    		load:     load,
    		capacity: capacity,
    		items:    map[string]*list.Element{},
    		lru:      list.New(),
    		flights:  map[string]*flight{},
    	}
    }

    // Get returns restaurant's menu, from the cache or by loading it.
    func (c *MenuCache) Get(ctx context.Context, restaurant string) (Menu, error) {
    	c.mu.Lock()
    	if el, ok := c.items[restaurant]; ok {
    		c.lru.MoveToFront(el)
    		c.mu.Unlock()
    		return el.Value.(*entry).menu, nil
    	}
    	if ctx.Err() != nil {
    		c.mu.Unlock()
    		return Menu{}, context.Cause(ctx)
    	}
    	f, ok := c.flights[restaurant]
    	if !ok {
    		loadCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
    		f = &flight{done: make(chan struct{}), cancel: cancel}
    		c.flights[restaurant] = f
    		go c.run(loadCtx, restaurant, f)
    	}
    	f.waiters++
    	c.mu.Unlock()

    	select {
    	case <-f.done:
    		return f.menu, f.err
    	case <-ctx.Done():
    		c.mu.Lock()
    		defer c.mu.Unlock()
    		f.waiters--
    		if f.waiters == 0 && c.flights[restaurant] == f {
    			delete(c.flights, restaurant) // nobody wants it any more
    			f.cancel()
    		}
    		return Menu{}, context.Cause(ctx)
    	}
    }

    func (c *MenuCache) run(ctx context.Context, restaurant string, f *flight) {
    	menu, err := c.load(ctx, restaurant)
    	c.mu.Lock()
    	defer c.mu.Unlock()
    	if c.flights[restaurant] == f {
    		delete(c.flights, restaurant)
    		if err == nil {
    			c.add(restaurant, menu)
    		}
    	}
    	f.menu, f.err = menu, err
    	close(f.done)
    	f.cancel()
    }

    // add caches a menu. c.mu must be held.
    func (c *MenuCache) add(restaurant string, menu Menu) {
    	if el, ok := c.items[restaurant]; ok {
    		el.Value.(*entry).menu = menu
    		c.lru.MoveToFront(el)
    		return
    	}
    	c.items[restaurant] = c.lru.PushFront(&entry{restaurant, menu})
    	if c.lru.Len() > c.capacity {
    		oldest := c.lru.Back()
    		c.lru.Remove(oldest)
    		delete(c.items, oldest.Value.(*entry).restaurant)
    	}
    }

    // Len returns how many menus are cached.
    func (c *MenuCache) Len() int {
    	c.mu.Lock()
    	defer c.mu.Unlock()
    	return c.lru.Len()
    }

    func main() {
    	var mu sync.Mutex
    	loads := 0
    	cache := NewMenuCache(2, func(ctx context.Context, r string) (Menu, error) {
    		mu.Lock()
    		loads++
    		mu.Unlock()
    		time.Sleep(20 * time.Millisecond) // a slow restaurant API
    		return Menu{r, 1}, nil
    	})

    	var wg sync.WaitGroup
    	for range 5 { // five customers open the same menu at once
    		wg.Go(func() { cache.Get(context.Background(), "Pho Real") })
    	}
    	wg.Wait()
    	m, err := cache.Get(context.Background(), "Pho Real")
    	fmt.Println(m, err, "loads:", loads, "cached:", cache.Len())
    }
  tests: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"maps"
    	"sync"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // api is a fake load func. Each load takes 1s (fake time); the nth load
    // of a restaurant returns Version n, or fails[restaurant] if it's set.
    type api struct {
    	mu        sync.Mutex
    	loads     map[string]int
    	fails     map[string]error
    	cancelled []string // "0.5s Pho Real"
    	start     time.Time
    }

    func newAPI() *api {
    	return &api{loads: map[string]int{}, fails: map[string]error{}, start: time.Now()}
    }

    func (a *api) load(ctx context.Context, r string) (Menu, error) {
    	a.mu.Lock()
    	a.loads[r]++
    	version := a.loads[r]
    	a.mu.Unlock()
    	select {
    	case <-time.After(time.Second):
    	case <-ctx.Done():
    		a.mu.Lock()
    		a.cancelled = append(a.cancelled, fmt.Sprint(time.Since(a.start), " ", r))
    		a.mu.Unlock()
    		return Menu{}, ctx.Err()
    	}
    	a.mu.Lock()
    	defer a.mu.Unlock()
    	if err := a.fails[r]; err != nil {
    		delete(a.fails, r) // fail once
    		return Menu{}, err
    	}
    	return Menu{r, version}, nil
    }

    func (a *api) count() map[string]int {
    	a.mu.Lock()
    	defer a.mu.Unlock()
    	return maps.Clone(a.loads)
    }

    type result struct {
    	menu Menu
    	err  error
    	at   time.Duration
    }

    // get calls c.Get in a goroutine; read the result from the channel.
    func get(c *MenuCache, ctx context.Context, r string) <-chan result {
    	start := time.Now()
    	ch := make(chan result, 1)
    	go func() {
    		m, err := c.Get(ctx, r)
    		ch <- result{m, err, time.Since(start)}
    	}()
    	return ch
    }

    func wait(t *testing.T, ch <-chan result) result {
    	t.Helper()
    	select {
    	case r := <-ch:
    		return r
    	case <-time.After(time.Hour):
    		t.Fatalf("Get hadn't returned after an hour")
    		return result{}
    	}
    }

    // serialLoads is set if loads can't run in parallel. The synctest tests
    // would then hang (fake time can't advance while goroutines wait for a
    // mutex), so they stop early instead.
    var serialLoads bool

    func requireParallelLoads(t *testing.T) {
    	t.Helper()
    	if serialLoads {
    		t.Fatal("not run: fix TestLoadsRunInParallel first")
    	}
    }

    // TestLoadsRunInParallel uses real time, with a wide margin: 8 loads of
    // 200ms each should take 200ms in parallel, and 1.6s one after another.
    func TestLoadsRunInParallel(t *testing.T) {
    	c := NewMenuCache(10, func(ctx context.Context, r string) (Menu, error) {
    		time.Sleep(200 * time.Millisecond)
    		return Menu{r, 1}, nil
    	})
    	start := time.Now()
    	var wg sync.WaitGroup
    	for i := range 8 {
    		wg.Go(func() { c.Get(context.Background(), fmt.Sprint("R", i)) })
    	}
    	wg.Wait()
    	if took := time.Since(start); took > time.Second {
    		serialLoads = true
    		t.Fatalf("8 different restaurants with 200ms loads took %v, want about 200ms: loads must run in parallel, so don't hold the mutex while calling load", took.Round(time.Millisecond))
    	}
    }

    func TestHitsAndMisses(t *testing.T) {
    	requireParallelLoads(t)
    	synctest.Test(t, func(t *testing.T) {
    		a := newAPI()
    		c := NewMenuCache(10, a.load)
    		for i, want := range []result{
    			{Menu{"Pho Real", 1}, nil, time.Second}, // miss
    			{Menu{"Pho Real", 1}, nil, 0},           // hit
    			{Menu{"Pho Real", 1}, nil, 0},           // hit
    		} {
    			if got := wait(t, get(c, t.Context(), "Pho Real")); got != want {
    				t.Errorf("Get #%d of Pho Real = %v, %v after %v; want %v, %v after %v", i+1, got.menu, got.err, got.at, want.menu, want.err, want.at)
    			}
    		}
    		if got := a.count(); got["Pho Real"] != 1 {
    			t.Errorf("load called %d times for Pho Real, want 1", got["Pho Real"])
    		}
    		if n := c.Len(); n != 1 {
    			t.Errorf("Len() = %d, want 1", n)
    		}
    	})
    }

    func TestConcurrentMissesShareOneLoad(t *testing.T) {
    	requireParallelLoads(t)
    	synctest.Test(t, func(t *testing.T) {
    		a := newAPI()
    		c := NewMenuCache(10, a.load)
    		var chs []<-chan result
    		for range 20 {
    			chs = append(chs, get(c, t.Context(), "Taco Bout It"))
    		}
    		for i, ch := range chs {
    			if got := wait(t, ch); got.menu != (Menu{"Taco Bout It", 1}) || got.err != nil || got.at != time.Second {
    				t.Errorf("concurrent Get #%d = %v, %v after %v; want {Taco Bout It 1}, nil after 1s", i+1, got.menu, got.err, got.at)
    			}
    		}
    		if n := a.count()["Taco Bout It"]; n != 1 {
    			t.Errorf("20 concurrent Gets for one restaurant called load %d times, want 1", n)
    		}
    	})
    }

    func TestDifferentRestaurantsLoadInParallel(t *testing.T) {
    	requireParallelLoads(t)
    	synctest.Test(t, func(t *testing.T) {
    		a := newAPI()
    		c := NewMenuCache(10, a.load)
    		var chs []<-chan result
    		for _, r := range []string{"A", "B", "C", "D"} {
    			chs = append(chs, get(c, t.Context(), r))
    		}
    		for i, ch := range chs {
    			if got := wait(t, ch); got.at != time.Second || got.err != nil {
    				t.Errorf("Get of restaurant %d of 4, all at once: returned %v after %v, want nil after 1s (don't hold the lock while loading)", i+1, got.err, got.at)
    			}
    		}
    	})
    }

    func TestErrorsAreNotCached(t *testing.T) {
    	requireParallelLoads(t)
    	synctest.Test(t, func(t *testing.T) {
    		errDown := errors.New("restaurant API down")
    		a := newAPI()
    		a.fails["Curry Up"] = errDown
    		c := NewMenuCache(10, a.load)
    		first, second := get(c, t.Context(), "Curry Up"), get(c, t.Context(), "Curry Up")
    		for _, ch := range []<-chan result{first, second} {
    			if got := wait(t, ch); !errors.Is(got.err, errDown) {
    				t.Errorf("Get while the load fails = %v, %v; want the load's error for every waiter", got.menu, got.err)
    			}
    		}
    		if n := c.Len(); n != 0 {
    			t.Errorf("Len() after a failed load = %d, want 0", n)
    		}
    		if got := wait(t, get(c, t.Context(), "Curry Up")); got.err != nil || got.menu != (Menu{"Curry Up", 2}) {
    			t.Errorf("Get after a failed load = %v, %v; want a fresh load's {Curry Up 2}, nil", got.menu, got.err)
    		}
    	})
    }

    func TestEvictsLeastRecentlyUsed(t *testing.T) {
    	requireParallelLoads(t)
    	synctest.Test(t, func(t *testing.T) {
    		a := newAPI()
    		c := NewMenuCache(2, a.load)
    		for _, r := range []string{"A", "B", "A", "C"} { // A is used after B, so B is the oldest when C arrives
    			wait(t, get(c, t.Context(), r))
    		}
    		if n := c.Len(); n != 2 {
    			t.Errorf("capacity 2: Len() = %d, want 2", n)
    		}
    		before := a.count()
    		wait(t, get(c, t.Context(), "A"))
    		wait(t, get(c, t.Context(), "C"))
    		after := a.count()
    		if after["A"] != before["A"] || after["C"] != before["C"] {
    			t.Errorf("capacity 2, Gets A B A C: A and C should still be cached, but getting them loaded A %d and C %d more times", after["A"]-before["A"], after["C"]-before["C"])
    		}
    		wait(t, get(c, t.Context(), "B"))
    		if after := a.count(); after["B"] != 2 {
    			t.Errorf("capacity 2, Gets A B A C: B should have been evicted (least recently used), but Get(B) didn't load it again")
    		}
    	})
    }

    func TestCallerGivingUpDoesNotCancelOthers(t *testing.T) {
    	requireParallelLoads(t)
    	synctest.Test(t, func(t *testing.T) {
    		errLeft := errors.New("customer closed the app")
    		a := newAPI()
    		c := NewMenuCache(10, a.load)
    		impatient, cancel := context.WithCancelCause(t.Context())
    		time.AfterFunc(500*time.Millisecond, func() { cancel(errLeft) })

    		quitter := get(c, impatient, "Wok This Way") // starts the load...
    		synctest.Wait()
    		patient := get(c, t.Context(), "Wok This Way") // ...and this one joins it

    		if got := wait(t, quitter); !errors.Is(got.err, errLeft) || got.at != 500*time.Millisecond {
    			t.Errorf("caller cancelled at 0.5s: Get = %v after %v, want the cause %q after 0.5s", got.err, got.at, errLeft)
    		}
    		if got := wait(t, patient); got.err != nil || got.menu != (Menu{"Wok This Way", 1}) || got.at != time.Second {
    			t.Errorf("other caller of the same load: Get = %v, %v after %v; want {Wok This Way 1}, nil after 1s", got.menu, got.err, got.at)
    		}
    		if len(a.cancelled) != 0 {
    			t.Errorf("load was cancelled (%v) although one caller was still waiting for it", a.cancelled)
    		}
    		if n := a.count()["Wok This Way"]; n != 1 {
    			t.Errorf("load called %d times, want 1", n)
    		}
    	})
    }

    func TestLoadCancelledWhenEveryoneGivesUp(t *testing.T) {
    	requireParallelLoads(t)
    	synctest.Test(t, func(t *testing.T) {
    		a := newAPI()
    		c := NewMenuCache(10, a.load)
    		ctx1, cancel1 := context.WithTimeout(t.Context(), 300*time.Millisecond)
    		defer cancel1()
    		ctx2, cancel2 := context.WithTimeout(t.Context(), 600*time.Millisecond)
    		defer cancel2()
    		first, second := get(c, ctx1, "Brew Bros"), get(c, ctx2, "Brew Bros")
    		wait(t, first)
    		if got := wait(t, second); !errors.Is(got.err, context.DeadlineExceeded) {
    			t.Errorf("Get with a 600ms timeout on a 1s load = %v, want context.DeadlineExceeded", got.err)
    		}
    		synctest.Wait()
    		a.mu.Lock()
    		cancelled := fmt.Sprint(a.cancelled)
    		a.mu.Unlock()
    		if cancelled != "[600ms Brew Bros]" {
    			t.Errorf("both callers gave up (at 300ms and 600ms): load cancellations %s, want [600ms Brew Bros] (cancel the load once nobody is waiting)", cancelled)
    		}

    		if got := wait(t, get(c, t.Context(), "Brew Bros")); got.err != nil || got.menu != (Menu{"Brew Bros", 2}) {
    			t.Errorf("Get after an abandoned load = %v, %v; want a fresh load's {Brew Bros 2}, nil", got.menu, got.err)
    		}
    	})
    }

    func TestManyGoroutines(t *testing.T) {
    	requireParallelLoads(t)
    	synctest.Test(t, func(t *testing.T) {
    		a := newAPI()
    		c := NewMenuCache(3, a.load)
    		var wg sync.WaitGroup
    		for g := range 20 {
    			wg.Go(func() {
    				for i := range 30 {
    					r := fmt.Sprint("R", (g+i)%6)
    					m, err := c.Get(t.Context(), r)
    					if err != nil || m.Restaurant != r {
    						t.Errorf("Get(%s) = %v, %v", r, m, err)
    						return
    					}
    					if n := c.Len(); n > 3 {
    						t.Errorf("capacity 3, Len() = %d", n)
    						return
    					}
    				}
    			})
    		}
    		wg.Wait()
    	})
    }
---

Opening a restaurant on Dispatchly shows its menu, and fetching a menu from a
restaurant's API takes a while. At lunchtime hundreds of customers open the
same few menus at once, so Dispatchly keeps a small cache in front of the
API.

Implement `MenuCache`:

- `NewMenuCache(capacity, load)` creates an empty cache that holds up to
  `capacity` menus and fetches missing ones with `load`.
- `Get(ctx, restaurant)` returns the cached menu straight away if there is
  one. Otherwise it loads it and caches it.
- **One load per restaurant:** while a load for a restaurant is in progress,
  other `Get`s for it wait for that same load instead of starting their own.
  Loads for *different* restaurants run in parallel.
- **Errors aren't cached:** every caller waiting on a failed load gets its
  error, and the next `Get` tries again.
- **LRU eviction:** when a new menu would make more than `capacity`, drop the
  least recently used one (a cache hit counts as a use).
- **Cancellation:** a caller whose `ctx` is done stops waiting and returns
  `context.Cause(ctx)`, but the load carries on for everyone else. Only when
  **every** caller waiting for a load has given up is the load's context
  cancelled; a later `Get` then starts a fresh load.
- `Len()` returns the number of cached menus.

## Example

```go
cache := NewMenuCache(2, load)  // load takes 1s
// 20 goroutines call cache.Get(ctx, "Taco Bout It") at once:
//   load runs once, and all 20 get the same menu after 1s
cache.Get(ctx, "Taco Bout It")  // instant: cached
```

## Constraints

- `capacity` ≥ 1. Many goroutines call `Get` and `Len` concurrently.
- The load's context should keep the values of the caller that started it
  (`context.WithoutCancel` helps) but not its cancellation.
- The tests use `synctest`: they count `load` calls, check that parallel
  loads finish at 1s rather than 4s, and check the exact moment every caller
  and every load gives up.
