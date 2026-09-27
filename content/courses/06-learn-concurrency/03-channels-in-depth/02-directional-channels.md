---
title: Directional Channel Types
quiz:
  - question: |
      Which line fails to compile?

      ```go
      func courier(jobs <-chan string, done chan<- string) {
      	j := <-jobs     // line 1
      	done <- j       // line 2
      	jobs <- "retry" // line 3
      	close(done)     // line 4
      }
      ```
    options:
      - text: Line 1
      - text: Line 2
      - text: Line 3
        correct: true
      - text: Line 4
    explanation: |
      `jobs` is receive-only (`<-chan string`), so sending on it is a compile
      error. Closing `done` is allowed: closing is a sender's job, and `done`
      is send-only.
  - question: You have a `chan Order` called `ch`. Which assignment does **not** compile?
    options:
      - text: '`var r <-chan Order = ch`'
      - text: '`var s chan<- Order = ch`'
      - text: '`var b chan Order = r`, where `r` is the `<-chan Order` from the first option'
        correct: true
      - text: All three compile
    explanation: |
      A bidirectional channel converts implicitly to either direction, but
      there's no way back. Once a value has type `<-chan Order`, the compiler
      will never let anyone send on it through that variable.
---

A plain `chan Order` can be used for both sending and receiving. Most functions only need one of the two, and Go lets you say so in the type:

| Type | Meaning | Allowed |
| --- | --- | --- |
| `chan T` | bidirectional | send, receive, close |
| `chan<- T` | send-only | send, close |
| `<-chan T` | receive-only | receive |

The arrow shows the direction of data relative to the `chan` keyword: in `chan<-` the data flows *into* the channel, in `<-chan` it flows *out*.

## Directions in function signatures

In Dispatchly, the kitchen produces ready orders and the pickup counter consumes them:

```go
package main

import "fmt"

// kitchen only sends finished orders.
func kitchen(orders []string, ready chan<- string) {
	for _, o := range orders {
		ready <- o + " (ready)"
	}
	close(ready)
}

// counter only receives them.
func counter(ready <-chan string) {
	for o := range ready {
		fmt.Println("pick up:", o)
	}
}

func main() {
	ready := make(chan string) // bidirectional
	go kitchen([]string{"A1", "B2", "C3"}, ready)
	counter(ready)
}
```

```text
pick up: A1 (ready)
pick up: B2 (ready)
pick up: C3 (ready)
```

`main` creates an ordinary bidirectional channel and passes it to both functions. Go converts it to the narrower type automatically at each call. Inside `kitchen` and `counter`, the compiler now enforces each role:

```go
func counter(ready <-chan string) {
	ready <- "fake order" // compile error
	close(ready)          // compile error
}
```

```text
invalid operation: cannot send to receive-only channel <-chan string ready (variable of type <-chan string)
invalid operation: cannot close receive-only channel ready (variable of type <-chan string)
```

Notice that you can't close a receive-only channel. Closing tells receivers "no more values are coming", and only a sender can promise that. You'll see why that matters in the next lesson.

## Returning a receive-only channel

A very common shape is a function that starts a goroutine and hands back a channel for the caller to read:

```go
func trackCourier(id string) <-chan Location {
	out := make(chan Location)
	go func() {
		defer close(out)
		for _, loc := range gpsPings(id) {
			out <- loc
		}
	}()
	return out
}
```

Inside, `out` is bidirectional, because the goroutine needs to send and close. The return type narrows it, so callers can only receive. They can't accidentally send fake locations or close a channel they don't own.

The conversion only goes one way. A bidirectional channel converts to either direction, but there's no way to turn a `<-chan T` back into a `chan T`.

## Why bother?

- **Documentation you can't ignore.** `func dispatch(jobs <-chan Job, results chan<- Result)` tells you at a glance which side of each channel the function is on.
- **Bugs become compile errors.** Sending on the wrong channel, or a consumer closing a channel it doesn't own, won't build.
- **Ownership is clear.** Whoever holds the bidirectional channel (usually the function that created it) is its owner and decides when to close it.

Make directional types your default for parameters and return values. Only the code that creates the channel needs the bidirectional type.
