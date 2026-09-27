---
title: The Done Channel
quiz:
  - question: Why is a done channel stopped with `close(done)` rather than `done <- struct{}{}`?
    options:
      - text: Sending on a `chan struct{}` doesn't compile
      - text: A send wakes up exactly one receiver, but closing wakes up every current and future receiver
        correct: true
      - text: Closing is faster than sending
      - text: A send would panic if nobody is listening
    explanation: |
      Each sent value is received by one goroutine only. With five workers
      you'd need five sends, and you'd have to know how many are listening.
      A closed channel is ready for *every* receiver, forever, so one
      `close` is a broadcast.
  - question: |
      This worker is told to stop with `close(done)` while it's blocked on `out <- result`, and nobody is reading `out`. What happens?

      ```go
      for j := range jobs {
      	result := process(j)
      	out <- result
      }
      ```
    options:
      - text: The worker notices `done` and returns
      - text: The worker panics
      - text: The worker stays blocked forever, because it never looks at `done`
        correct: true
      - text: The runtime drops the send and moves on
    explanation: |
      Closing `done` only helps goroutines that are waiting on it. Every
      place a goroutine can block (here, the range receive and the send)
      needs to be a `select` that also watches `done`.
---

You've now seen several ways a goroutine can end up stuck: a producer nobody reads from any more, a consumer that stopped early, a worker waiting for jobs that will never come. What they're all missing is a way for *someone else* to say "stop". The classic answer, before `context` existed, is the **done channel**.

## A broadcast that says stop

A done channel carries no data. Its only job is to be closed:

```go
done := make(chan struct{})
// ... later, to stop everyone:
close(done)
```

`struct{}` is the empty struct, which takes zero bytes. It signals "this channel is for signalling, not data". Closing is the key: a receive on a closed channel is always ready, so **every** goroutine waiting on `<-done`, now or later, wakes up. A send would only wake one.

## Workers that listen

Each goroutine checks `done` wherever it might block, using `select`:

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

// courier delivers jobs until done is closed, then reports how many it did.
func courier(jobs <-chan string, done <-chan struct{}) int {
	delivered := 0
	for {
		select {
		case <-done:
			return delivered
		case <-jobs:
			time.Sleep(time.Millisecond) // deliver it
			delivered++
		}
	}
}

func main() {
	jobs := make(chan string)
	done := make(chan struct{})

	names := []string{"ana", "ben", "cy"}
	counts := make([]int, len(names))
	var wg sync.WaitGroup
	for i := range names {
		wg.Go(func() { counts[i] = courier(jobs, done) })
	}

	for i := range 9 {
		jobs <- fmt.Sprint("order-", i)
	}

	close(done) // one close tells every courier to go home
	wg.Wait()
	fmt.Println("all couriers stopped, delivered:", counts[0]+counts[1]+counts[2])
}
```

```text
all couriers stopped, delivered: 9
```

Notice that `jobs` is never closed here. The couriers stop because of `done`, not because the work ran out. That's the difference: closing a data channel says "no more work", while closing `done` says "stop, even if there's more work". And unlike closing `jobs`, it's safe no matter how many senders there are, because nobody ever sends on `done`.

`close(done)` followed by `wg.Wait()` is the standard shutdown: signal, then wait until everyone has actually stopped.

## Watch done at *every* blocking point

A goroutine can only react to `done` while it's waiting on it. A producer that ranges over its input and sends results has **two** places it can block, so both need a `select`:

```go
func prices(done <-chan struct{}, ids <-chan string) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for {
			var id string
			select {
			case <-done:
				return
			case v, ok := <-ids:
				if !ok {
					return
				}
				id = v
			}

			select {
			case <-done:
				return
			case out <- lookupPrice(id):
			}
		}
	}()
	return out
}
```

This is exactly the fix for the "consumer breaks out early" problem from the last chapter. The consumer closes `done` when it's had enough, and the producer, blocked on `out <- ...`, takes the `done` case and returns instead of leaking.

## Giving done priority

`select` picks randomly among ready cases. If jobs keep coming, a worker might pick a job even though `done` is already closed. It will get to `done` eventually, but if you want stopping to win, check it first without blocking:

```go
for {
	select {
	case <-done:
		return
	default:
	}

	select {
	case <-done:
		return
	case j := <-jobs:
		handle(j)
	}
}
```

## Asking to stop isn't the same as stopped

Closing `done` only *asks* the workers to stop. It doesn't tell you they have. If you need to know that a worker has finished its cleanup (flushed its last delivery update, say), you need a second signal going the other way: a **completion** channel that the *worker* closes as it exits.

```go
stop := make(chan struct{})     // closed by the caller: "please stop"
finished := make(chan struct{}) // closed by the worker: "I have stopped"
go func() {
	defer close(finished)
	<-stop
	// flush the last location update, then return
}()

close(stop) // ask
<-finished  // wait until it has really stopped
```

With several workers, a `WaitGroup` plays the completion role, which is exactly the `close(done)` then `wg.Wait()` shutdown above. Give the two channels different names so nobody confuses "stop requested" with "stopped".

## From done channels to context

The done channel is simple and it works, but real services need more. How do you stop a whole *tree* of goroutines, where each one started its own helpers? How do you combine "stop when the user cancels" with "stop after 2 seconds"? How does a goroutine find out *why* it was stopped?

The standard library's answer is `context.Context`. Under the hood it's built on exactly this idea: `ctx.Done()` returns a channel that gets closed. That's the next chapter.

## Further reading

- [Go blog: Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines)
