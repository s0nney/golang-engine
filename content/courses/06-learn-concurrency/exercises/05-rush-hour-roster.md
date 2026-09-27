---
title: Rush Hour Roster
difficulty: easy
after: sync-primitives
hints:
  - '`WaitFor` locks `r.mu`, then calls `r.cond.Wait()` **in a loop** while the condition is false. `Wait` unlocks the mutex while it sleeps and locks it again before returning, so you can check `r.online` safely each time round.'
  - 'Why a loop and not an `if`? Being woken up only means "something changed". It doesn''t mean enough couriers are online for *this* waiter.'
  - 'Different waiters wait for different numbers, so `CheckIn` must wake **all** of them with `r.cond.Broadcast()`. `Signal` wakes only one, which may be a waiter that still has to keep waiting while the one that could go stays asleep.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"sync"
    	"time"
    )

    // Roster counts the couriers who are online.
    type Roster struct {
    	mu     sync.Mutex
    	cond   *sync.Cond // uses mu as its lock
    	online int
    }

    func NewRoster() *Roster {
    	r := &Roster{}
    	r.cond = sync.NewCond(&r.mu)
    	return r
    }

    // CheckIn marks one more courier as online and wakes any waiters.
    func (r *Roster) CheckIn() {
    	r.mu.Lock()
    	defer r.mu.Unlock()
    	r.online++
    	// TODO: wake the waiters.
    }

    // CheckOut marks one courier as offline.
    func (r *Roster) CheckOut() {
    	r.mu.Lock()
    	defer r.mu.Unlock()
    	r.online--
    }

    // WaitFor blocks until at least n couriers are online.
    func (r *Roster) WaitFor(n int) {
    	// TODO: lock r.mu, then r.cond.Wait() for as long as fewer than n
    	// couriers are online.
    }

    func main() {
    	r := NewRoster()
    	go func() {
    		for range 3 {
    			time.Sleep(20 * time.Millisecond)
    			fmt.Println("a courier checks in")
    			r.CheckIn()
    		}
    	}()
    	r.WaitFor(3)
    	fmt.Println("3 couriers online: open the lunch rush!")
    	// want: three check-ins, then the lunch rush opens
    }
  solution: |
    package main

    import (
    	"fmt"
    	"sync"
    	"time"
    )

    // Roster counts the couriers who are online.
    type Roster struct {
    	mu     sync.Mutex
    	cond   *sync.Cond // uses mu as its lock
    	online int
    }

    func NewRoster() *Roster {
    	r := &Roster{}
    	r.cond = sync.NewCond(&r.mu)
    	return r
    }

    // CheckIn marks one more courier as online and wakes any waiters.
    func (r *Roster) CheckIn() {
    	r.mu.Lock()
    	defer r.mu.Unlock()
    	r.online++
    	r.cond.Broadcast()
    }

    // CheckOut marks one courier as offline.
    func (r *Roster) CheckOut() {
    	r.mu.Lock()
    	defer r.mu.Unlock()
    	r.online--
    }

    // WaitFor blocks until at least n couriers are online.
    func (r *Roster) WaitFor(n int) {
    	r.mu.Lock()
    	defer r.mu.Unlock()
    	for r.online < n {
    		r.cond.Wait()
    	}
    }

    func main() {
    	r := NewRoster()
    	go func() {
    		for range 3 {
    			time.Sleep(20 * time.Millisecond)
    			fmt.Println("a courier checks in")
    			r.CheckIn()
    		}
    	}()
    	r.WaitFor(3)
    	fmt.Println("3 couriers online: open the lunch rush!")
    }
  tests: |
    package main

    import (
    	"sync/atomic"
    	"testing"
    	"testing/synctest"
    )

    // waiter calls r.WaitFor(n) in a goroutine and records when it returns.
    func waiter(r *Roster, n int) *atomic.Bool {
    	var done atomic.Bool
    	go func() {
    		r.WaitFor(n)
    		done.Store(true)
    	}()
    	synctest.Wait() // let it start waiting
    	return &done
    }

    func TestWaitForBlocksUntilEnough(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		r := NewRoster()
    		done := waiter(r, 3)
    		for i := range 2 {
    			r.CheckIn()
    			synctest.Wait()
    			if done.Load() {
    				t.Fatalf("WaitFor(3) returned after only %d of 3 check-ins", i+1)
    			}
    		}
    		r.CheckIn()
    		synctest.Wait()
    		if !done.Load() {
    			t.Fatalf("WaitFor(3) still blocked after 3 check-ins: wake waiters in CheckIn")
    		}
    	})
    }

    func TestWaitForAlreadyMet(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		r := NewRoster()
    		r.CheckIn()
    		r.CheckIn()
    		for _, n := range []int{0, 1, 2} {
    			if done := waiter(r, n); !done.Load() {
    				t.Errorf("with 2 couriers online, WaitFor(%d) blocked, want it to return straight away", n)
    			}
    		}
    	})
    }

    func TestEveryWaiterIsWoken(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		r := NewRoster()
    		needTwo := waiter(r, 2) // started first, so it's first in line
    		needOne := waiter(r, 1)

    		r.CheckIn()
    		synctest.Wait()
    		if needTwo.Load() {
    			t.Errorf("WaitFor(2) returned with only 1 courier online: re-check the condition in a loop after Wait")
    		}
    		if !needOne.Load() {
    			t.Errorf("WaitFor(1) still blocked with 1 courier online: CheckIn must wake every waiter (Broadcast), not just one")
    		}

    		r.CheckIn()
    		synctest.Wait()
    		if !needTwo.Load() {
    			t.Errorf("WaitFor(2) still blocked with 2 couriers online")
    		}
    		r.CheckIn() // wake anything still stuck so the test can end
    		r.CheckIn()
    	})
    }

    func TestCheckOut(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		r := NewRoster()
    		r.CheckIn()
    		r.CheckIn()
    		r.CheckOut()
    		done := waiter(r, 2)
    		if done.Load() {
    			t.Fatalf("2 check-ins and 1 check-out, but WaitFor(2) returned")
    		}
    		r.CheckIn()
    		synctest.Wait()
    		if !done.Load() {
    			t.Fatalf("WaitFor(2) still blocked with 2 couriers online")
    		}
    	})
    }
---

Dispatchly doesn't open a city's **lunch rush** (with its surge discounts)
until enough couriers are online to handle it. Several parts of the system
wait for this, each for a different number of couriers.

`Roster` counts online couriers. `CheckOut` is done. Finish the other two
methods using the `sync.Cond` in the struct:

- `WaitFor(n)` blocks until at least `n` couriers are online. If there already
  are, it returns straight away.
- `CheckIn()` adds a courier and wakes up the waiters, so each one can check
  whether it can go now.

## Example

```go
r := NewRoster()
go r.WaitFor(2) // blocks...
go r.WaitFor(1) // blocks...
r.CheckIn()     // WaitFor(1) returns; WaitFor(2) keeps waiting
r.CheckIn()     // WaitFor(2) returns
```

## Constraints

- No polling (`time.Sleep` loops): a waiter must wake up as soon as a
  check-in lets it go. The tests use `synctest.Wait`, which returns once every
  goroutine is blocked, and then check who has returned.
- Many goroutines may call every method at the same time.
