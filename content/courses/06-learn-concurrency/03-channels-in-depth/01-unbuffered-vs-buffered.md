---
title: Unbuffered vs Buffered Channels
quiz:
  - question: |
      What happens when this runs?

      ```go
      func main() {
      	orders := make(chan string, 2)
      	orders <- "A1"
      	orders <- "B2"
      	orders <- "C3"
      	fmt.Println(<-orders)
      }
      ```
    options:
      - text: It prints `A1`
      - text: It prints `C3`
      - text: 'It crashes with `fatal error: all goroutines are asleep - deadlock!`'
        correct: true
      - text: The third send overwrites `A1`, then it prints `B2`
    explanation: |
      The buffer holds two values. The third send finds it full and blocks
      until someone receives, but the only goroutine that could receive is
      `main` itself, which is stuck on the send. The runtime notices that no
      goroutine can ever make progress and crashes.
  - question: Goroutine K sends an order on an **unbuffered** channel and the send statement completes. What does K now know?
    options:
      - text: Nothing, the value might still be waiting in the channel
      - text: Another goroutine has received the value
        correct: true
      - text: Another goroutine has finished processing the value
      - text: The channel has been closed
    explanation: |
      An unbuffered channel has nowhere to store a value, so a send can only
      complete by handing it directly to a receiver. That makes unbuffered
      channels a *synchronization* point, not just a pipe. It says nothing
      about what the receiver did next, though.
---

Every channel has a **capacity**, fixed when you create it. That one number completely changes how the channel behaves.

```go
handoff := make(chan Order)      // unbuffered: capacity 0
queue   := make(chan Order, 100) // buffered: capacity 100
```

## Unbuffered: a handoff

An unbuffered channel has no storage. A send blocks until a receiver takes the value, and a receive blocks until a sender offers one. Both goroutines have to show up at the same time, like a cook handing a bag directly to a courier at the counter. Neither can leave until the handoff is done.

That makes an unbuffered channel a **synchronization** tool as well as a way to move data. When a send on an unbuffered channel completes, the sender knows the receiver has the value. Go's memory model guarantees that everything the sender wrote before sending is visible to the receiver after receiving.

The flip side: with no receiver, a send blocks forever. If that happens in `main` with nothing else running, the runtime catches it:

```go
package main

func main() {
	handoff := make(chan string)
	handoff <- "A1" // no one will ever receive this
}
```

```text
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan send]:
main.main()
	/tmp/main.go:5 +0x31
exit status 2
```

## Buffered: a queue

A buffered channel has a queue of slots. A send only blocks if the buffer is **full**, and a receive only blocks if it's **empty**. Values come out in the order they went in (FIFO), like tickets on the kitchen's order rail.

`len(ch)` reports how many values are waiting and `cap(ch)` the capacity:

```go
package main

import "fmt"

func main() {
	rail := make(chan string, 3)
	rail <- "A1"
	rail <- "B2"
	fmt.Println("waiting:", len(rail), "of", cap(rail))

	fmt.Println("cooking", <-rail)
	fmt.Println("cooking", <-rail)
	fmt.Println("waiting:", len(rail))
}
```

```text
waiting: 2 of 3
cooking A1
cooking B2
waiting: 0
```

Note that `main` sent to and received from the same channel without any other goroutine. That's only possible because the buffer had room. An unbuffered version would deadlock on the first send.

Don't make decisions based on `len(ch)` in concurrent code. By the time you act on the number, another goroutine may have changed it. It's fine for logging and metrics.

## Choosing a capacity

Buffering changes *when* goroutines block, never *whether* your program is correct. A program that deadlocks with an unbuffered channel usually just deadlocks a bit later with a buffered one. Pick a capacity for a reason:

- **0 (unbuffered)**: the default. Use it when you want the sender to know the value was received, or when you have no good reason for anything else.
- **1**: a common choice for a goroutine that sends exactly one result, so it can finish and exit even if nobody is listening any more.
- **N, where you know N exactly**: when N goroutines each send one value, a buffer of N means no sender ever blocks. You used this to fix the leak in `fastestETA`.
- **A larger buffer** to absorb bursts: say, the kitchen sometimes gets 20 orders in a second but handles them steadily. Pick the size from measurements, and remember a buffer only smooths bursts. If producers are *always* faster than consumers, any buffer eventually fills up.

A huge buffer "just to be safe" often hides a bug, such as a consumer that stopped consuming, until it fills up in production.

## Further reading

- [Go by Example: Channel Buffering](https://gobyexample.com/channel-buffering)
- [The Go Memory Model](https://go.dev/ref/mem), the section on channel communication
