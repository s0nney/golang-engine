---
title: Mutex vs RWMutex
quiz:
  - question: |
      What happens when `Rename` is called?

      ```go
      func (r *Registry) Rename(old, new string) {
      	r.mu.RLock()
      	defer r.mu.RUnlock()
      	if loc, ok := r.locs[old]; ok {
      		r.mu.Lock()
      		r.locs[new] = loc
      		r.mu.Unlock()
      	}
      }
      ```
    options:
      - text: It works, because the goroutine already holds the lock
      - text: It deadlocks, because `Lock` waits for all readers to leave, including this goroutine itself
        correct: true
      - text: It panics with "lock upgrade not supported"
      - text: It works, but is slower than using `Lock` from the start
    explanation: |
      Go's locks aren't reentrant and there's no lock upgrading. `Lock`
      waits until every read lock is released, but the goroutine holding one
      is the one waiting, so it waits forever. Take the write lock from the
      start when you might write.
  - question: When is an `RWMutex` likely to beat a plain `Mutex`?
    options:
      - text: Always, since it allows more concurrency
      - text: When writes are much more common than reads
      - text: When reads greatly outnumber writes and many goroutines read at the same time
        correct: true
      - text: When the critical section is a single integer increment
    explanation: |
      Read locks only pay off when lots of readers really do overlap. For
      tiny critical sections or write-heavy workloads, the extra bookkeeping
      of an `RWMutex` makes it slower than a `Mutex`. Measure before
      switching.
---

Channels are Go's headline feature, but they're not always the right tool. When several goroutines need to read and update **shared state** (a cache, a registry, a counter), a lock is usually simpler. As the Go proverb goes, *share memory by communicating*, but the standard library ships `sync` for a reason.

## Mutex recap

A `sync.Mutex` lets one goroutine at a time into a **critical section**:

```go
type Stats struct {
	mu        sync.Mutex
	delivered int // guarded by mu
}

func (s *Stats) Delivered() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delivered++
}
```

Conventions worth copying: put the mutex right above the fields it protects, say so in a comment, and `defer` the unlock straight after locking. The zero value is an unlocked mutex, ready to use.

## RWMutex: many readers or one writer

Dispatchly's courier registry is read constantly (every order lookup wants a courier's location) but written far less often. A plain mutex makes readers queue up behind each other even though reads can't interfere with each other. `sync.RWMutex` has two modes:

- `RLock` / `RUnlock`: **shared** read lock. Any number of goroutines can hold it at once.
- `Lock` / `Unlock`: **exclusive** write lock. It waits for all readers to leave, and blocks new ones while held.

```go
package main

import (
	"fmt"
	"sync"
)

// Registry tracks where every courier is.
// Many goroutines read locations; a few update them.
type Registry struct {
	mu   sync.RWMutex
	locs map[string]string // guarded by mu
}

func NewRegistry() *Registry {
	return &Registry{locs: make(map[string]string)}
}

func (r *Registry) Update(courier, loc string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.locs[courier] = loc
}

func (r *Registry) Location(courier string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	loc, ok := r.locs[courier]
	return loc, ok
}

func main() {
	reg := NewRegistry()

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Go(func() { reg.Update(fmt.Sprint("courier-", i%10), fmt.Sprint("block ", i)) })
		wg.Go(func() { reg.Location("courier-3") })
	}
	wg.Wait()

	fmt.Println(len(reg.locs), "couriers tracked")
}
```

```text
10 couriers tracked
```

## Which one?

Start with `Mutex`. It's simpler and, for short critical sections, usually faster: an `RWMutex` does more bookkeeping on every call. Switch to `RWMutex` only when reads vastly outnumber writes, the critical section does real work (not one map lookup), and profiling shows readers waiting on each other.

A reader holding `RLock` must never write. Nothing stops you: the compiler doesn't know which fields are "read" and which are "written". You'll only find out through the race detector or corrupted data.

## Gotchas

**No upgrading.** You can't turn a read lock into a write lock. Calling `Lock` while holding `RLock` deadlocks, because `Lock` waits for all readers to leave, including you. Go's mutexes aren't reentrant either: calling `Lock` twice in the same goroutine deadlocks. If a locked method needs to call another locked method, split out an unexported helper that assumes the lock is already held.

**Don't copy a mutex.** A copied mutex is a separate lock, so the copy protects nothing. It usually happens by accident with a value receiver or by passing a struct by value:

```go
func report(c Counter) { // Counter contains a sync.Mutex
	fmt.Println(c.n)
}
```

`go vet` catches this:

```text
report passes lock by value: Counter contains sync.Mutex
```

Use pointer receivers on types with a mutex, and pass them as pointers.

**Keep critical sections small.** Hold a lock only while touching shared state. Don't call a restaurant API or send on a channel while holding one: every other goroutine that needs the lock waits for that slow call, and a blocked channel send while locked is a classic deadlock. Copy what you need, unlock, then do the slow part.

**Protect every access.** A lock only works if *every* read and write of the data goes through it. One unguarded read in a logging function is still a data race.

## Further reading

- [`sync.RWMutex` documentation](https://pkg.go.dev/sync#RWMutex)
