---
title: Data Races vs Race Conditions
quiz:
  - question: |
      Every access to `balance` is protected by `mu`. Is this code correct when many goroutines call `Withdraw` at once?

      ```go
      func (w *Wallet) Withdraw(cents int) bool {
      	w.mu.Lock()
      	enough := w.balance >= cents
      	w.mu.Unlock()
      	if !enough {
      		return false
      	}
      	w.mu.Lock()
      	w.balance -= cents
      	w.mu.Unlock()
      	return true
      }
      ```
    options:
      - text: Yes, there's no data race, so it's correct
      - text: No, it's a race condition, since two goroutines can both pass the check before either subtracts, overdrawing the wallet
        correct: true
      - text: No, it's a data race on `balance`
      - text: No, it deadlocks because it locks twice
    explanation: |
      Every read and write is inside the lock, so the race detector stays
      quiet. But the check and the update happen in *separate* critical
      sections, and another goroutine can run in between. The fix is one
      critical section around both.
  - question: What is a data race, precisely?
    options:
      - text: Any bug that depends on the order goroutines run in
      - text: Two goroutines accessing the same memory at the same time, at least one of them writing, with no synchronization between them
        correct: true
      - text: Two goroutines reading the same variable
      - text: A goroutine that runs faster than expected
    explanation: |
      That's the definition in the Go memory model. Concurrent reads are
      fine. A write concurrent with any other access, with no mutex, channel
      or atomic ordering them, is a data race, and the behaviour is
      undefined.
---

"Race" gets used for two different bugs. They have different causes, different symptoms and different tools, so it's worth keeping them apart.

## Data races

A **data race** happens when two goroutines access the same memory **concurrently**, at least one access is a **write**, and **nothing synchronizes** them: no mutex, no channel operation, no atomic.

```go
delivered := 0
var wg sync.WaitGroup
for range 2 {
	wg.Go(func() {
		delivered++ // read, add, write: unsynchronized
	})
}
wg.Wait()
```

A data race is always a bug, even if it seems harmless. `delivered++` is really three steps (load, add, store), so updates get lost. Worse, the compiler and CPU are allowed to reorder and cache unsynchronized memory accesses, so a racy program can do things that no ordering of its statements could explain: see a half-written string, a slice whose length doesn't match its contents, a flag that's set but data that isn't. A racy map access can crash the program.

Data races are **mechanical**. There's a precise definition, and a tool, the race detector (next lesson), that finds them when they happen. The fix is always synchronization: a mutex, a channel, or an atomic.

## Race conditions

A **race condition** is a *logic* bug where the result depends on the timing of goroutines. It can happen in code with perfect synchronization and no data races at all.

Dispatchly's classic: two dispatchers try to give the same courier an order at the same moment:

```go
func (d *Dispatcher) Assign(courier, order string) bool {
	d.mu.Lock()
	_, busy := d.assigned[courier]
	d.mu.Unlock()
	if busy {
		return false
	}
	// ← another Assign can run right here
	d.mu.Lock()
	d.assigned[courier] = order
	d.mu.Unlock()
	return true
}
```

Every access to the map is locked, so there's no data race and the race detector is silent. But both calls can check "is ana busy?" before either marks her busy. Both return `true`, and ana gets two orders. This is **check-then-act**, the most common race condition. Its cousins:

- **Read-modify-write** split across lock sections: read a balance, unlock, compute, lock, write it back. Updates in between are lost.
- **Using `len(ch)` to decide** whether a send will block. By the time you send, it may have changed.
- **Assuming an order** between goroutines: "the logger goroutine will have started by now".

The fix is to make the check and the act **one atomic step**: one critical section, a `CompareAndSwap`, or a single goroutine that owns the state and handles requests one at a time.

## Side by side

| | Data race | Race condition |
| --- | --- | --- |
| What it is | Unsynchronized concurrent access to memory | Wrong result depending on timing |
| Can happen with locks everywhere? | No | Yes |
| Found by the race detector? | Yes, when it occurs during the run | No |
| Fix | Add synchronization | Make the logical operation atomic |

You can have either without the other. Fixing a data race by wrapping each access in its own lock often just turns it into a race condition, as in `Assign` above. Think about which *operations* must be indivisible, not just which *variables* are shared.

## Further reading

- [The Go Memory Model](https://go.dev/ref/mem)
