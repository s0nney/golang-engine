---
title: Once, OnceFunc and OnceValue
quiz:
  - question: |
      Ten goroutines call `getConfig()` at the same moment. What's true?

      ```go
      var getConfig = sync.OnceValue(func() Config {
      	return loadConfig() // takes 2 seconds
      })
      ```
    options:
      - text: '`loadConfig` may run up to ten times, but only one result is kept'
      - text: The first caller gets the config and the other nine get a zero `Config` straight away
      - text: '`loadConfig` runs once, and all ten calls wait for it and return the same value'
        correct: true
      - text: It deadlocks, because all ten goroutines wait on each other
    explanation: |
      No call to a `Once` function returns until the single run of `f` has
      finished. Everyone who arrives during those 2 seconds waits, then gets
      the cached result. That's exactly what makes it safe for lazy
      initialization.
  - question: A function wrapped in `sync.OnceValues` returns an error the first time it runs. What happens on the next call?
    options:
      - text: It runs again, hoping to succeed this time
      - text: It returns the same cached result and error without running again
        correct: true
      - text: It panics
      - text: It returns a zero value and a `nil` error
    explanation: |
      The `Once` family caches whatever `f` returned, errors included, and
      never runs it again. That's right for things that can't recover (a
      broken config file) but wrong for flaky network calls, where you'd
      want a retry. Pick the tool for the failure mode.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"sync"
    	"sync/atomic"
    	"time"
    )

    type Menu struct {
    	Restaurant string
    	Dishes     int
    }

    // newMenuLoader returns a function that loads the menu on first use.
    // However many goroutines call it, and however concurrently, load must
    // run at most once, and every caller gets the same Menu and error.
    func newMenuLoader(load func() (Menu, error)) func() (Menu, error) {
    	var (
    		loaded bool
    		menu   Menu
    		err    error
    	)
    	return func() (Menu, error) {
    		if !loaded {
    			menu, err = load()
    			loaded = true
    		}
    		return menu, err
    	}
    }

    func main() {
    	var loads atomic.Int64
    	getMenu := newMenuLoader(func() (Menu, error) {
    		loads.Add(1)
    		time.Sleep(20 * time.Millisecond) // a slow restaurant API
    		return Menu{Restaurant: "pho-king", Dishes: 42}, nil
    	})

    	var wg sync.WaitGroup
    	for range 10 {
    		wg.Go(func() { getMenu() })
    	}
    	wg.Wait()

    	m, err := getMenu()
    	fmt.Println(m, err)
    	fmt.Println("restaurant API called", loads.Load(), "time(s)")
    }
  solution: |
    package main

    import (
    	"fmt"
    	"sync"
    	"sync/atomic"
    	"time"
    )

    type Menu struct {
    	Restaurant string
    	Dishes     int
    }

    // newMenuLoader returns a function that loads the menu on first use.
    // However many goroutines call it, and however concurrently, load must
    // run at most once, and every caller gets the same Menu and error.
    func newMenuLoader(load func() (Menu, error)) func() (Menu, error) {
    	return sync.OnceValues(load)
    }

    func main() {
    	var loads atomic.Int64
    	getMenu := newMenuLoader(func() (Menu, error) {
    		loads.Add(1)
    		time.Sleep(20 * time.Millisecond) // a slow restaurant API
    		return Menu{Restaurant: "pho-king", Dishes: 42}, nil
    	})

    	var wg sync.WaitGroup
    	for range 10 {
    		wg.Go(func() { getMenu() })
    	}
    	wg.Wait()

    	m, err := getMenu()
    	fmt.Println(m, err)
    	fmt.Println("restaurant API called", loads.Load(), "time(s)")
    }
  tests: |
    package main

    import (
    	"errors"
    	"sync"
    	"sync/atomic"
    	"testing"
    	"time"
    )

    func TestLoadsOnceUnderConcurrency(t *testing.T) {
    	var loads atomic.Int64
    	release := make(chan struct{})
    	getMenu := newMenuLoader(func() (Menu, error) {
    		loads.Add(1)
    		<-release // hold every load open until all callers have arrived
    		return Menu{Restaurant: "thai-tanic", Dishes: 17}, nil
    	})
    	if n := loads.Load(); n != 0 {
    		t.Fatalf("newMenuLoader called load %d time(s) before anyone asked for the menu, want 0 (load lazily)", n)
    	}

    	const callers = 20
    	menus := make([]Menu, callers)
    	errs := make([]error, callers)
    	var wg sync.WaitGroup
    	for i := range callers {
    		wg.Go(func() { menus[i], errs[i] = getMenu() })
    	}
    	time.Sleep(50 * time.Millisecond) // let every caller arrive
    	close(release)
    	wg.Wait()

    	if n := loads.Load(); n != 1 {
    		t.Errorf("%d concurrent callers caused %d loads, want exactly 1", callers, n)
    	}
    	for i := range callers {
    		if menus[i] != (Menu{"thai-tanic", 17}) || errs[i] != nil {
    			t.Errorf("caller %d got (%v, %v), want ({thai-tanic 17}, <nil>)", i, menus[i], errs[i])
    		}
    	}
    }

    func TestLaterCallsReuseResult(t *testing.T) {
    	var loads atomic.Int64
    	getMenu := newMenuLoader(func() (Menu, error) {
    		loads.Add(1)
    		return Menu{Restaurant: "wok-this-way", Dishes: int(loads.Load())}, nil
    	})
    	for range 5 {
    		if m, _ := getMenu(); m.Dishes != 1 {
    			t.Errorf("getMenu() = %v, want the menu from the first load, {wok-this-way 1}", m)
    		}
    	}
    	if n := loads.Load(); n != 1 {
    		t.Errorf("5 sequential calls caused %d loads, want 1", n)
    	}
    }

    func TestErrorIsCachedToo(t *testing.T) {
    	var loads atomic.Int64
    	errDown := errors.New("menu service down")
    	getMenu := newMenuLoader(func() (Menu, error) {
    		loads.Add(1)
    		return Menu{}, errDown
    	})
    	for range 3 {
    		if _, err := getMenu(); err != errDown {
    			t.Errorf("getMenu() error = %v, want %v", err, errDown)
    		}
    	}
    	if n := loads.Load(); n != 1 {
    		t.Errorf("a failing load ran %d times, want 1 (the error is cached along with the result)", n)
    	}
    }
---

Some work should happen exactly once: parsing the delivery-zone map, connecting to the pricing service, compiling a big regular expression. Doing it at startup slows every launch, even when nobody needs it. Doing it **lazily**, on first use, is faster, but then two goroutines might both see "not loaded yet" and load it twice, racing on the result. The `sync` package has a family of tools for exactly this.

## sync.Once

`once.Do(f)` runs `f` the first time it's called, and never again:

```go
type ZoneMap struct {
	once  sync.Once
	zones []Zone
}

func (m *ZoneMap) Zones() []Zone {
	m.once.Do(func() {
		m.zones = parseZonesFile()
	})
	return m.zones
}
```

The important guarantee: **no call to `Do` returns until `f` has finished**. If ten goroutines call `Zones()` at once, one runs `parseZonesFile` and the other nine wait for it. Then all ten see the loaded `zones`, with no data race.

## OnceFunc, OnceValue, OnceValues

Since Go 1.21 there are three wrappers that are usually nicer than a bare `Once`. Each takes a function and returns a new function that runs the original at most once:

| Wrapper | Wraps | Returns |
| --- | --- | --- |
| `sync.OnceFunc(f)` | `func()` | `func()` |
| `sync.OnceValue(f)` | `func() T` | `func() T`, cached result |
| `sync.OnceValues(f)` | `func() (T1, T2)` | `func() (T1, T2)`, both cached |

`OnceValue` turns lazy initialization into a one-liner, with no separate `once` and value fields to keep in sync:

```go
package main

import (
	"fmt"
	"sync"
)

type Zones struct {
	names []string
}

var zones = sync.OnceValue(func() *Zones {
	fmt.Println("loading delivery zones...") // expensive: reads a big file
	return &Zones{names: []string{"downtown", "harbour", "old town"}}
})

func main() {
	var wg sync.WaitGroup
	for range 5 {
		wg.Go(func() {
			_ = zones().names
		})
	}
	wg.Wait()
	fmt.Println("zones:", zones().names)
}
```

```text
loading delivery zones...
zones: [downtown harbour old town]
```

Six calls, one load. `OnceValues` is the same for the ever-popular `(T, error)` pair.

## Errors and panics are cached too

Whatever the function returned the first time, errors included, is what every later call gets:

```go
var loadPricing = sync.OnceValues(func() (map[string]int, error) {
	fmt.Println("fetching pricing table...")
	return nil, errors.New("pricing service unavailable")
})
```

Call it three times and you'll see `fetching pricing table...` once and the same error three times. That's what you want for a malformed config file that will never parse. It's *not* what you want for a network call that failed because of a blip: the service stays "unavailable" until the process restarts. For those, use a mutex and only store the result on success.

Panics differ between the two APIs. With `OnceFunc`, `OnceValue` and `OnceValues`, if `f` panics, **every** call re-panics with the same value. With a plain `Once`, `Do` treats a panic as "done", and later calls silently do nothing, leaving you with half-initialized state. That's another reason to prefer the wrappers.

## Don't call it recursively

Because `Do` waits for `f` to finish, calling the same `Once` from inside `f` deadlocks: the inner call waits for the outer one, which is waiting for the inner one.

## Your turn

Loading a restaurant's menu from its API is slow, so Dispatchly loads each menu lazily and caches it. `newMenuLoader` takes a `load` function and returns a `getMenu` function that callers use instead.

The starter caches with a plain `loaded` flag. It works when calls happen one after another, but when many goroutines ask for the menu at once they all see `loaded == false` and hammer the restaurant's API. Press **Run** to count the calls. (It's also a data race.)

Fix `newMenuLoader` so that `load` runs at most once no matter how many goroutines call `getMenu` concurrently, and every caller gets the same menu and error. `load` must not run until the first call.
