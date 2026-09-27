---
title: Atomic Values
quiz:
  - question: |
      Is this safe to call from many goroutines at once?

      ```go
      var inFlight atomic.Int64

      func startDelivery() bool {
      	if inFlight.Load() >= 100 {
      		return false // at capacity
      	}
      	inFlight.Add(1)
      	return true
      }
      ```
    options:
      - text: Yes, both operations are atomic
      - text: No, two goroutines can both load 99, both pass the check, and both add, ending at 101
        correct: true
      - text: No, `atomic.Int64` must be passed by pointer
      - text: No, `Load` and `Add` can't be used on the same variable
    explanation: |
      Each operation is atomic on its own, but the check-then-act *sequence*
      isn't. Another goroutine can slip in between `Load` and `Add`. Use a
      `CompareAndSwap` loop, a mutex, or a semaphore when a decision depends
      on the current value.
  - question: 'Why prefer `atomic.Int64` over calling `atomic.AddInt64(&n, 1)` on a plain `int64`?'
    options:
      - text: It's much faster
      - text: The type guarantees every access is atomic, since you can't accidentally read or write it with plain `n++`, and it's always correctly aligned
        correct: true
      - text: '`atomic.AddInt64` was removed in Go 1.19'
      - text: '`atomic.Int64` can also store strings'
    explanation: |
      With the old functions, nothing stops someone writing `n++` somewhere
      else, which is a data race. The typed values only expose atomic
      methods. They also fix a subtle alignment problem with 64-bit values
      on 32-bit platforms.
---

A mutex is a big hammer for a tiny nail when all you want is to count orders. `sync/atomic` provides operations that the CPU performs **indivisibly**: no other goroutine can ever see them half done. They're the building blocks the `sync` package itself is made of.

## Typed atomic values

Since Go 1.19, `sync/atomic` has types that make every access atomic:

| Type | Holds |
| --- | --- |
| `atomic.Int32`, `atomic.Int64` | signed integers |
| `atomic.Uint32`, `atomic.Uint64`, `atomic.Uintptr` | unsigned integers |
| `atomic.Bool` | a bool |
| `atomic.Pointer[T]` | a `*T` |
| `atomic.Value` | any value (older, untyped; prefer `Pointer[T]`) |

The zero value is ready to use. The methods are `Load`, `Store`, `Swap` and `CompareAndSwap`, plus `Add` (and `And`/`Or` since Go 1.23) on the integer types.

```go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Metrics struct {
	ordersPlaced atomic.Int64
	surge        atomic.Bool
}

func main() {
	var m Metrics

	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(func() { m.ordersPlaced.Add(1) })
	}
	wg.Wait()

	m.surge.Store(m.ordersPlaced.Load() > 500)
	fmt.Println("orders:", m.ordersPlaced.Load())
	fmt.Println("surge pricing:", m.surge.Load())
}
```

```text
orders: 1000
surge pricing: true
```

With a plain `int64` and `n++`, a thousand concurrent increments would lose updates. `n++` is really *load, add, store*, and two goroutines can load the same old value. `Add` does all three as one step.

Prefer these types over the older functions like `atomic.AddInt64(&n, 1)`. With the functions, nothing stops someone else writing `n++` on the same variable. With the types, there's no non-atomic way in. Like mutexes, they must not be copied after first use, and `go vet` checks that.

## CompareAndSwap

`CompareAndSwap(old, new)` sets the value to `new` **only if** it's currently `old`, and reports whether it did. It's how you make a decision atomically. Here, four couriers race to accept the same order, and exactly one wins:

```go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var assignedTo atomic.Pointer[string] // nil: order not yet assigned

	couriers := []string{"ana", "ben", "cy", "dee"}
	wins := make([]bool, len(couriers))
	var wg sync.WaitGroup
	for i, c := range couriers {
		wg.Go(func() {
			// Claim the order only if nobody has claimed it yet.
			wins[i] = assignedTo.CompareAndSwap(nil, &c)
		})
	}
	wg.Wait()

	winners := 0
	for _, w := range wins {
		if w {
			winners++
		}
	}
	fmt.Println("winners:", winners)
	fmt.Println("order assigned:", assignedTo.Load() != nil)
}
```

```text
winners: 1
order assigned: true
```

Which courier wins varies from run to run, but there's always exactly one.

## Swapping whole snapshots with Pointer[T]

`atomic.Pointer[T]` is great for **read-mostly configuration**. Readers `Load` a pointer to an immutable snapshot, and an updater builds a new value and `Store`s it:

```go
package main

import (
	"fmt"
	"sync/atomic"
)

type Pricing struct {
	BaseFee   int
	PerKmCent int
}

var pricing atomic.Pointer[Pricing]

func quote(km int) int {
	p := pricing.Load() // one consistent snapshot
	return p.BaseFee + km*p.PerKmCent
}

func main() {
	pricing.Store(&Pricing{BaseFee: 299, PerKmCent: 80})
	fmt.Println("5km:", quote(5))

	// An admin updates prices: build a new value, then swap it in.
	pricing.Store(&Pricing{BaseFee: 349, PerKmCent: 90})
	fmt.Println("5km:", quote(5))
}
```

```text
5km: 699
5km: 799
```

A quote never sees the new base fee combined with the old per-km rate, because it reads both fields from the same snapshot. The rule that makes this work: **never modify a `Pricing` after storing it.** Always build a fresh one.

## Atomic operations don't compose

Each atomic operation is safe on its own, but a *sequence* of them isn't. "Load, check, then add" can interleave with another goroutine doing the same thing between your steps. When a decision depends on the current value, use `CompareAndSwap` in a retry loop, or just use a mutex.

## When to use atomics

Use them for independent counters, flags and snapshot pointers. The moment two values must change together (a count *and* a total, a map *and* its size), use a mutex. Mutex code is easier to get right, and code that's easy to get right beats code that's a few nanoseconds faster.
