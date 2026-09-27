---
title: sync.Map and When Not to Use It
quiz:
  - question: 'Dispatchly keeps a per-restaurant order count that every request increments. Which is the better fit?'
    options:
      - text: '`sync.Map`, because it''s designed for concurrent use'
      - text: A `map[string]int` guarded by a `sync.Mutex`
        correct: true
      - text: A plain `map[string]int` with no locking, since increments are fast
      - text: '`sync.Map` with `LoadOrStore` for each increment'
    explanation: |
      Every request writes, and keys are overwritten constantly, which is
      neither of the two cases `sync.Map` is optimized for. A normal map
      behind a mutex is type-safe, simple, and lets you update the count in
      one locked step. With no lock at all, the runtime can crash with
      "concurrent map writes".
  - question: 'What does `v, ok := m.Load("ana")` give you if `m` is a `sync.Map`?'
    options:
      - text: '`v` is whatever type you stored, checked by the compiler'
      - text: '`v` has type `any`, so you need a type assertion to use it'
        correct: true
      - text: '`v` is always a `string`'
      - text: It doesn't compile without a type parameter
    explanation: |
      `sync.Map` predates generics. Keys and values are `any`, so every
      read needs a type assertion, and the compiler won't stop you storing
      the wrong type.
---

Go's built-in maps are not safe for concurrent use. If one goroutine writes while another reads or writes, the runtime may crash the whole program with `fatal error: concurrent map writes` (it deliberately checks for this). So when you see a map shared between goroutines, you might reach for `sync.Map`, "a map that's safe for concurrent use". Usually, that's the wrong move.

## What sync.Map looks like

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var sessions sync.Map // courier ID -> device token

	sessions.Store("ana", "tok-81f")
	sessions.Store("ben", "tok-2c9")

	if tok, ok := sessions.Load("ana"); ok {
		fmt.Println("ana:", tok.(string)) // values come back as any
	}

	actual, loaded := sessions.LoadOrStore("ana", "tok-NEW")
	fmt.Println("ana:", actual, "already there:", loaded)

	sessions.Delete("ben")
	sessions.Range(func(k, v any) bool {
		fmt.Println(k, "=>", v)
		return true // keep going
	})
}
```

```text
ana: tok-81f
ana: tok-81f already there: true
ana => tok-81f
```

It works, but notice the costs:

- **No type safety.** Keys and values are `any`. Every `Load` needs a type assertion, and nothing stops you storing an `int` where a `string` belongs.
- **No `len`.** There's no way to ask how many entries there are without ranging over all of them.
- **No compound operations.** "Increment this count" or "update two entries together" can't be done in one step. You'd need `CompareAndSwap` retry loops or a separate lock anyway.

## What it's actually for

The documentation is unusually direct: *"The Map type is specialized. Most code should use a plain Go map instead, with separate locking or coordination."* It's optimized for two situations:

1. **Write once, read many**: a cache that only grows, where each key is stored once and then read a huge number of times, such as compiled templates keyed by name.
2. **Disjoint keys**: many goroutines each reading and writing their *own* separate sets of keys.

In those two cases, on machines with many cores, it can reduce lock contention a lot compared to one mutex. In every other case it's usually slower *and* harder to use.

## The default: a map and a mutex

```go
type OrderCounts struct {
	mu     sync.Mutex
	counts map[string]int // restaurant -> orders today; guarded by mu
}

func (c *OrderCounts) Inc(restaurant string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[restaurant]++
}

func (c *OrderCounts) Get(restaurant string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counts[restaurant]
}
```

It's type-safe, `c.counts[restaurant]++` is one step under the lock, and adding a `Len` method or a second map that must stay consistent is trivial. If profiling ever shows readers contending, switch to an `RWMutex` first.

## Choosing, in one table

| Situation | Use |
| --- | --- |
| Shared map, any normal workload | `map` + `sync.Mutex` |
| Read-heavy, with real contention | `map` + `sync.RWMutex` |
| Grow-only cache or disjoint keys per goroutine, measured contention | `sync.Map` |
| One goroutine owns the map, others send it requests | plain `map`, no lock, fed by a channel |

That last row is worth remembering: a map owned by a single goroutine needs no locking at all. Other goroutines talk to the owner through channels. That's *share memory by communicating* in its purest form.
