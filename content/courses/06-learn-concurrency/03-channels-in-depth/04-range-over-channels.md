---
title: Ranging Over Channels
quiz:
  - question: |
      `pings` returns early when the GPS feed fails. What happens to a caller doing `for p := range pings(...)` when that happens?

      ```go
      func pings(courier string) <-chan string {
      	out := make(chan string)
      	go func() {
      		for {
      			p, err := readGPS(courier)
      			if err != nil {
      				return
      			}
      			out <- p
      		}
      		close(out)
      	}()
      	return out
      }
      ```
    options:
      - text: The loop ends normally, because the goroutine returned
      - text: The loop receives an empty string and then ends
      - text: The loop waits forever, because the early `return` skips `close(out)`
        correct: true
      - text: It panics with "receive from closed channel"
    explanation: |
      Returning from the goroutine doesn't close anything. The `close(out)`
      line is never reached, so the caller's `range` never ends. Writing
      `defer close(out)` as the goroutine's first line closes the channel on
      every exit path, including early returns added later.
  - question: A consumer ranges over a producer's unbuffered channel and hits `break` after 2 of 10 values. What happens to the producer goroutine?
    options:
      - text: It's stopped automatically when the loop breaks
      - text: It panics with "send on closed channel"
      - text: It blocks forever on its next send, leaking, unless it has another way to learn that it should stop
        correct: true
      - text: It keeps sending into the channel's hidden overflow buffer
    explanation: |
      `break` just stops receiving. It doesn't close anything or signal
      anyone. The producer's next send has no receiver, so it blocks forever.
      Producers that might be abandoned need a stop signal, which is exactly
      what done channels and contexts are for.
---

`for v := range ch` receives until the channel is **closed and drained**. You used it in Learn Go. Here we'll look at its contract from both sides: the producer must promise to close, and the consumer must promise to keep reading. Break either promise and a goroutine gets stuck.

## The producer closes, the consumer ranges

The most common shape in Go is a function that starts a goroutine and returns a receive-only channel, which the caller ranges over:

```go
package main

import "fmt"

func pings(courier string, n int) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for i := range n {
			out <- fmt.Sprintf("%s ping #%d", courier, i+1)
		}
	}()
	return out
}

func main() {
	for p := range pings("ana", 3) {
		fmt.Println(p)
	}
	fmt.Println("stream ended")
}
```

```text
ana ping #1
ana ping #2
ana ping #3
stream ended
```

`defer close(out)` as the goroutine's first line is the idiom. It closes the channel however the goroutine exits, including early returns you (or a colleague) add later.

## Broken promise 1: the producer never closes

In Learn Go, forgetting `close` crashed the program with "all goroutines are asleep - deadlock!". That only happens because *every* goroutine was stuck. In a real server, with HTTP listeners, tickers and other goroutines still running, the runtime can't tell anything is wrong. There's no crash, just a consumer goroutine waiting forever for a value that will never come: a leak.

## Broken promise 2: the consumer stops early

The reverse problem: the consumer stops before the producer is done.

```go
for p := range pings("ana", 1000) {
	if isAtRestaurant(p) {
		break // found what we wanted
	}
}
```

`break` doesn't tell the producer anything. Its next `out <- ...` has no receiver and blocks forever. The consumer can't fix this by closing the channel either, because then the producer would panic on its next send.

The answer is a second signal going the *other* way, from consumer to producer, meaning "stop now". Chapter 4 builds that with `select`, and chapter 5 swaps it for `context.Context`. Until then, remember: **a `range` over a channel is only safe if the producer is guaranteed to close it, and the consumer is guaranteed to read until it does.**

## Channels or iterators?

You may notice that `pings` looks a lot like an `iter.Seq[string]`. If all you want is to produce a sequence of values lazily, a range-over-func iterator is simpler and faster: no goroutine, no channel, and `break` just works, because the iterator's `yield` returns `false` and the iterator stops. Reach for a channel when the values are produced *concurrently*: by a goroutine doing real work in parallel with the consumer, or by several producers at once.
