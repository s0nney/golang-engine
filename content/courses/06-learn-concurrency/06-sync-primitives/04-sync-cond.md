---
title: Condition Variables with sync.Cond
quiz:
  - question: Why must `cond.Wait()` be called inside a `for` loop that re-checks the condition, rather than an `if`?
    options:
      - text: Because `Wait` only waits for a fixed amount of time
      - text: 'Because when `Wait` returns, the condition may no longer be true: another goroutine can grab the resource between the wake-up and relocking'
        correct: true
      - text: Because `Wait` panics if called only once
      - text: It doesn't matter, `if` works just as well
    explanation: |
      Being woken up means "something changed, go and look", not "the
      thing you want is ready". By the time the waiter has relocked the
      mutex, another goroutine may have taken the courier. The loop checks
      again and goes back to waiting if needed.
  - question: What's the channel equivalent of `cond.Broadcast()`?
    options:
      - text: Sending one value on a channel
      - text: Closing a channel
        correct: true
      - text: Setting a channel to `nil`
      - text: Reading from a buffered channel
    explanation: |
      Closing a channel wakes every receiver, just as `Broadcast` wakes
      every waiter. `Signal`, which wakes one waiter, corresponds to sending
      one value. Unlike a closed channel, though, a `Cond` can broadcast
      again and again.
---

Sometimes a goroutine needs to wait until some **shared state** reaches a condition: "wait until a courier is idle", "wait until the queue has room". The state is guarded by a mutex, so you can't just loop checking it (that's busy-waiting, and you'd need to hold the lock to check). `sync.Cond` lets a goroutine sleep until another goroutine says "the state changed, have another look".

## How it works

A `Cond` is paired with a lock, usually the mutex that guards the state:

```go
p.cond = sync.NewCond(&p.mu)
```

It has three methods:

- **`Wait()`**: must be called with the lock held. It atomically **unlocks** and goes to sleep. When woken, it **relocks** before returning.
- **`Signal()`**: wakes one waiting goroutine, if any.
- **`Broadcast()`**: wakes all waiting goroutines.

## A courier pool

Dispatchers call `Take` to get an idle courier, waiting if there are none. When a courier finishes a delivery, `Release` puts them back and wakes a waiting dispatcher:

```go
package main

import (
	"fmt"
	"sync"
)

// Pool is a set of idle couriers that dispatchers can wait on.
type Pool struct {
	mu   sync.Mutex
	cond *sync.Cond
	idle []string // guarded by mu
}

func NewPool() *Pool {
	p := &Pool{}
	p.cond = sync.NewCond(&p.mu)
	return p
}

// Take blocks until a courier is idle, then removes and returns it.
func (p *Pool) Take() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	for len(p.idle) == 0 { // always a loop, never an if
		p.cond.Wait() // unlocks mu while waiting, relocks before returning
	}
	c := p.idle[0]
	p.idle = p.idle[1:]
	return c
}

// Release makes a courier idle again and wakes one waiting dispatcher.
func (p *Pool) Release(courier string) {
	p.mu.Lock()
	p.idle = append(p.idle, courier)
	p.mu.Unlock()
	p.cond.Signal()
}

func main() {
	pool := NewPool()

	got := make([]string, 3)
	var wg sync.WaitGroup
	for i := range 3 {
		wg.Go(func() { got[i] = pool.Take() })
	}

	for _, c := range []string{"ana", "ben", "cy"} {
		pool.Release(c)
	}
	wg.Wait()

	fmt.Println("every dispatcher got a courier:", got[0] != "" && got[1] != "" && got[2] != "")
}
```

```text
every dispatcher got a courier: true
```

## Always wait in a loop

`for len(p.idle) == 0 { p.cond.Wait() }` is the only correct shape. A wake-up means "something changed", not "your condition is now true". Between `Signal` and the waiter relocking the mutex, another goroutine might call `Take` and grab the courier. The loop checks again, and if the pool is empty it waits some more.

## Why you'll rarely use it

The `sync.Cond` docs say it plainly: *for many simple use cases, users will be better off using channels than a Cond.* `Broadcast` corresponds to closing a channel and `Signal` to sending on one. The pool above is just a buffered channel of couriers:

```go
idle := make(chan string, 100)
idle <- "ana"   // Release
c := <-idle     // Take: blocks until a courier is available
```

And the channel version gets something `Cond` can't do: it works in a `select`, so `Take` can also watch `ctx.Done()` and give up. `Cond.Wait` has no timeout and no cancellation (you'd need the `context.AfterFunc` + `Broadcast` trick from the context chapter).

Reach for `sync.Cond` when many goroutines wait on a complex condition over state that already lives behind a mutex, and when you need to broadcast **repeatedly** (a closed channel can only be closed once). Otherwise, use a channel.
