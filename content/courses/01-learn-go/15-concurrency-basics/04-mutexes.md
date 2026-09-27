---
title: Mutexes
quiz:
  - question: |
      What's wrong with this code, where many goroutines call `record` at
      the same time?

      ```go
      var sent = map[string]int{}

      func record(user string) {
      	sent[user]++
      }
      ```
    options:
      - text: Nothing, `++` is safe to use from many goroutines
      - text: It's a data race, because several goroutines can read and write the map at the same moment
        correct: true
      - text: The map should be created with `make` instead of a literal
    explanation: |
      Concurrent writes to a map are a data race. Go's runtime may even crash
      with `concurrent map writes`. Protect the map with a `sync.Mutex`, or
      give it to one goroutine and talk to that goroutine over a channel.
  - question: Why does the lesson write `defer c.mu.Unlock()` straight after `c.mu.Lock()`?
    options:
      - text: Because `Unlock` must be called before `Lock`
      - text: So the mutex is always unlocked when the method returns, whichever path it returns through
        correct: true
      - text: Because deferred calls run faster
    explanation: |
      `defer` guarantees the unlock happens on every return path, including
      early returns and panics. Forgetting to unlock leaves every other
      goroutine waiting forever.
---

Channels are great for passing data between goroutines. But sometimes several goroutines genuinely need to update the **same** value, such as a shared counter of messages sent. For that, you need a lock.

## A data race

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	sent := 0
	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(func() {
			sent++ // DATA RACE
		})
	}
	wg.Wait()
	fmt.Println(sent)
}
```

You'd expect `1000`. Run it a few times and you may see `1000`, `987`, `952`... `sent++` is really three steps: read the value, add one, write it back. When two goroutines interleave those steps, both read `41`, both write `42`, and a message goes uncounted.

This is a **data race**: two goroutines access the same variable at the same time, and at least one of them writes. Races cause random, hard-to-reproduce bugs. Go has a built-in **race detector** to catch them. Run your program or tests with `-race`:

```text
$ go run -race main.go
==================
WARNING: DATA RACE
...
```

## `sync.Mutex`

A **mutex** ("mutual exclusion") is a lock that only one goroutine can hold at a time:

- `mu.Lock()` takes the lock. If another goroutine already holds it, `Lock` waits until it's free.
- `mu.Unlock()` releases it.

Code between `Lock` and `Unlock` is called the **critical section**. Only one goroutine can be inside it at once.

The usual pattern is to bundle the mutex with the data it protects, in a struct:

```go
package main

import (
	"fmt"
	"sync"
)

type sendCounter struct {
	mu     sync.Mutex
	counts map[string]int
}

func (c *sendCounter) record(user string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[user]++
}

func (c *sendCounter) get(user string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counts[user]
}

func main() {
	c := sendCounter{counts: make(map[string]int)}

	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(func() {
			c.record("alice")
		})
	}
	wg.Wait()
	fmt.Println(c.get("alice"))
}
```

```text
1000
```

Every time, guaranteed.

## Things to notice

- **`defer` the unlock.** Writing `defer c.mu.Unlock()` right after `Lock` means you can't forget it, even with early returns.
- **Reads need the lock too.** `get` locks as well. Reading while another goroutine writes is still a race.
- **Pointer receivers.** A mutex must never be copied (a copy is a separate lock that protects nothing), so the methods use `*sendCounter`. `go vet` warns if you copy a mutex by accident.
- **The zero value works.** A `sync.Mutex` needs no setup.
- **Keep critical sections short.** While one goroutine holds the lock, everyone else waits. Don't do slow work, like network calls, while holding it.

## Maps and concurrency

Go's maps are **not** safe for concurrent use. If goroutines write to a map at the same time, the runtime may stop the program with `fatal error: concurrent map writes`. Always protect shared maps with a mutex, like `sendCounter` does.

## Channels or mutexes?

- Use **channels** to pass data or ownership between goroutines, or to signal events like "done".
- Use a **mutex** to protect a small piece of shared state, like a counter or a cache.

Both are idiomatic. Pick whichever makes the code simpler to understand.

## Further reading

- [Go by Example: Mutexes](https://gobyexample.com/mutexes)
- [A Tour of Go: sync.Mutex](https://go.dev/tour/concurrency/9)
- [Learn Go with Tests: Sync](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/sync)
