---
title: Channels as Queues
quiz:
  - question: |
      What does this print?

      ```go
      jobs := make(chan string, 3)
      jobs <- "ava"
      jobs <- "bo"
      fmt.Println(<-jobs, len(jobs), cap(jobs))
      ```
    options:
      - text: '`bo 1 3`'
      - text: '`ava 1 3`'
        correct: true
      - text: '`ava 2 3`'
    explanation: |
      A buffered channel is FIFO, so the first value received is `"ava"`. The
      receive happens before `len` is evaluated, leaving one item in a buffer of
      capacity 3.
  - question: A goroutine sends to a buffered channel whose buffer is full, and no one is receiving. What happens?
    options:
      - text: The oldest item is dropped
      - text: The channel grows, like append
      - text: The send blocks until some other goroutine receives
        correct: true
      - text: It panics
    explanation: |
      Channels never grow and never drop values. A send on a full channel waits.
      If no other goroutine ever receives, the program deadlocks, and if every
      goroutine is stuck the runtime reports `all goroutines are asleep`.
---

Go has a queue built right into the language: the **buffered channel**. A
channel made with `make(chan T, n)` holds up to `n` values in FIFO order, and
under the hood it's a ring buffer, just like the one you wrote.

## A channel is a FIFO

```go
package main

import "fmt"

func main() {
	jobs := make(chan string, 3)
	jobs <- "refresh:ava"
	jobs <- "refresh:bo"
	jobs <- "refresh:cy"
	close(jobs)

	for j := range jobs {
		fmt.Println("processing", j)
	}
}
```

```
processing refresh:ava
processing refresh:bo
processing refresh:cy
```

Send (`jobs <- v`) is enqueue, receive (`<-jobs`) is dequeue, and `len(jobs)`
is how many are waiting. `range` keeps receiving until the channel is closed
*and* empty.

## What channels add

Channels are made for passing data **between goroutines**, so they come with
things our `Ring` doesn't have:

- **They're safe for concurrent use.** Many goroutines can send and receive at
  once with no extra locking. Our `Ring[T]` would need a `sync.Mutex` for that.
- **They block.** Receiving from an empty channel waits for a value. Sending to
  a full one waits for space. That makes them perfect for producer/consumer
  pipelines, where the buffer smooths out bursts.

Here's Clout's rate-limited refresher: one goroutine queues jobs, and three
workers process them.

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	jobs := make(chan string, 100)
	results := make(chan string, 100)

	var wg sync.WaitGroup
	for w := range 3 {
		wg.Go(func() {
			for handle := range jobs {
				results <- fmt.Sprintf("worker %d refreshed %s", w, handle)
			}
		})
	}

	for _, h := range []string{"ava", "bo", "cy", "dee", "eve"} {
		jobs <- h
	}
	close(jobs)
	wg.Wait()
	close(results)

	count := 0
	for range results {
		count++
	}
	fmt.Println("refreshed", count, "accounts")
}
```

This prints `refreshed 5 accounts`. Which worker handles which account varies
from run to run, which is why we only print the count. `sync.WaitGroup.Go`
(Go 1.25+) starts a goroutine and tracks it in one call.

## What channels can't do

Channels are a great queue *between goroutines*, but they're a poor
general-purpose data structure:

- **Fixed capacity.** A channel never grows. If you don't know how many items
  you'll queue, a full channel means a blocked sender, and in a single goroutine
  that's a deadlock: `fatal error: all goroutines are asleep - deadlock!`
- **No peeking.** You can't look at the front item without removing it.
- **No iteration without consuming.** Ranging over a channel empties it.
- **Overhead.** Every operation involves synchronisation, which costs more than
  a plain slice write when only one goroutine is involved.

The rule of thumb: use a **channel** when goroutines are handing work to each
other, and a **ring buffer** or **slice-backed queue** when one goroutine just
needs a FIFO. The takeaway: a buffered channel is a concurrent, blocking, fixed-size ring-buffer
queue.

## Further reading

- [Go by Example: Channel Buffering](https://gobyexample.com/channel-buffering)
- [Go by Example: Worker Pools](https://gobyexample.com/worker-pools)
