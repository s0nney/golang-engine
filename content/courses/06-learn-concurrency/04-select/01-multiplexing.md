---
title: Multiplexing with Select
quiz:
  - question: |
      Both channels already hold a value. What does this print?

      ```go
      a := make(chan string, 1)
      b := make(chan string, 1)
      a <- "kitchen"
      b <- "courier"

      select {
      case m := <-a:
      	fmt.Println(m)
      case m := <-b:
      	fmt.Println(m)
      }
      ```
    options:
      - text: Always `kitchen`, because its case comes first
      - text: Always `courier`, because the last send wins
      - text: '`kitchen` or `courier`, chosen at random'
        correct: true
      - text: Both lines
    explanation: |
      When several cases are ready, `select` picks one uniformly at random.
      Case order means nothing. That's deliberate: it stops one busy channel
      from starving the others. Only one case runs per `select`.
  - question: A `select` has no `default` case and none of its channels are ready. What happens?
    options:
      - text: It returns immediately and runs nothing
      - text: It blocks until at least one case can proceed
        correct: true
      - text: It panics
      - text: It runs the first case with a zero value
    explanation: |
      Without `default`, `select` waits, like a single send or receive
      would, until one of its cases can go. With zero cases at all,
      `select {}` blocks forever.
---

A single channel operation waits for exactly one thing. Real goroutines usually need to wait for *whichever of several things happens first*: an order is ready, a courier arrives, the customer cancels, a timer fires. That's what `select` does. It's Go's tool for **multiplexing** channels.

## Quick recap

```go
select {
case msg := <-kitchen:
	// a value arrived from kitchen
case courier <- job:
	// our send to courier went through
}
```

- Each `case` is a single send or receive.
- `select` blocks until at least one case can proceed, then runs **exactly one**.
- If several are ready, it picks one **at random**. Source order doesn't give priority.

## Serving several sources in a loop

The most common shape is `for` + `select`: a goroutine that keeps reacting to whatever arrives next. Here the dispatcher hears from the kitchen and a courier, in whichever order they report:

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	kitchen := make(chan string)
	courier := make(chan string)

	go func() {
		time.Sleep(30 * time.Millisecond)
		kitchen <- "order A1 is ready"
	}()
	go func() {
		time.Sleep(10 * time.Millisecond)
		courier <- "ana has arrived"
	}()

	for range 2 {
		select {
		case msg := <-kitchen:
			fmt.Println("kitchen:", msg)
		case msg := <-courier:
			fmt.Println("courier:", msg)
		}
	}
}
```

```text
courier: ana has arrived
kitchen: order A1 is ready
```

If you instead wrote `fmt.Println(<-kitchen)` followed by `fmt.Println(<-courier)`, `main` would sit waiting on the kitchen for 30ms while the courier's message was already available, and the courier goroutine would be stuck for that time too. With `select`, each message is handled the moment it arrives.

## How a select runs

It's worth knowing the exact steps, because they explain some surprising behaviour:

1. **Every channel expression and every value to send is evaluated once**, top to bottom, when the `select` starts. In `case out <- compute():`, `compute()` runs even if a different case ends up being chosen.
2. If one or more cases are ready, one is chosen at random and runs.
3. Otherwise, if there's a `default`, it runs (next lesson).
4. Otherwise the goroutine blocks until some case becomes ready.

Point 1 is a classic gotcha. Don't put expensive or side-effecting calls in case expressions. Compute the value first.

## Random choice and priority

Because the choice is random, you can't say "prefer this channel". If you need priority, say *cancellation beats new work*, check the important channel first with a separate non-blocking `select`, or check again inside the case. You'll see this in the done-channel lesson.

The randomness is a feature. If the choice were top-to-bottom, a constantly busy first channel would starve every case below it.

## Breaking out of a for-select loop

`break` inside a `select` only exits the **select**, not the surrounding `for`. To leave the loop, `return` from the function (usually the cleanest), or use a labelled break:

```go
dispatch:
	for {
		select {
		case o, ok := <-orders:
			if !ok {
				break dispatch // leaves the for loop
			}
			assign(o)
		case c := <-couriers:
			register(c)
		}
	}
```

Plain `break` here would just start the next loop iteration, receiving from a closed channel again and again. Many Go programmers wrap the loop in its own function and use `return` to avoid labels altogether.

## Further reading

- [Go by Example: Select](https://gobyexample.com/select)
- [The Go spec: Select statements](https://go.dev/ref/spec#Select_statements)
