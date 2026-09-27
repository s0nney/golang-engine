---
title: Concurrency vs Parallelism
quiz:
  - question: A single-core machine runs a Go program with 50 goroutines. Which statement is true?
    options:
      - text: The program is parallel but not concurrent
      - text: The program is concurrent, but at any instant only one goroutine is executing
        correct: true
      - text: The program can't run, because goroutines need at least one core each
      - text: The goroutines run one after another, each to completion, in the order they started
    explanation: |
      Concurrency is about *structure*: 50 independent tasks that make progress
      in overlapping periods of time. Parallelism is about *execution*: tasks
      running at the same instant. With one core, the scheduler interleaves the
      goroutines, so they're concurrent but never parallel.
  - question: |
      Each call to `callRestaurant` sleeps for 100ms, simulating a network call.
      Roughly how long does this take?

      ```go
      var wg sync.WaitGroup
      for _, r := range []string{"pho-king", "wok-this-way", "thai-tanic"} {
      	wg.Go(func() { callRestaurant(r) })
      }
      wg.Wait()
      ```
    options:
      - text: About 300ms
      - text: About 100ms
        correct: true
      - text: About 0ms, because `wg.Go` doesn't wait
      - text: It depends entirely on the number of CPU cores
    explanation: |
      The three calls wait at the same time, so the total is about as long as
      the slowest call. Sleeping (or waiting on the network) doesn't use a CPU,
      so this speed-up happens even on a single core.
---

Welcome to Dispatchly! We're a food-delivery platform: customers place orders, restaurants cook them, and couriers race across town to deliver them. Behind the app sits a dispatch engine that talks to hundreds of restaurants and thousands of couriers **at the same time**.

In the concurrency chapter of [Learn Go](/courses/learn-go/concurrency-basics/goroutines) you met goroutines, channels, `select` and mutexes. In this course you'll go much deeper, and by the end you'll be able to design, debug and *test* concurrent Go code with confidence.

First, some vocabulary. Two words get mixed up constantly: **concurrency** and **parallelism**.

## Concurrency is about structure

A program is **concurrent** when it's made of independent tasks whose lifetimes overlap. Task B can start before task A finishes.

Think of a single chef in a Dispatchly partner kitchen. She puts the rice on, then chops vegetables while the rice cooks, then flips a steak, then checks the rice. She only has two hands and does one thing at any instant, but she's juggling several dishes. That's concurrency.

## Parallelism is about execution

A program runs in **parallel** when several tasks execute at *literally* the same instant, which needs several CPU cores.

Now hire three chefs, each cooking a different dish at the same moment. That's parallelism.

Rob Pike, one of Go's creators, put it like this:

> Concurrency is about *dealing with* lots of things at once. Parallelism is about *doing* lots of things at once.

Concurrency is a way to *structure* a program. Parallelism is one way that structure can be *executed*. A concurrent program may or may not run in parallel: that depends on the hardware and the runtime, not on your code.

## Seeing the difference

Dispatchly asks three restaurants whether they can take an order. Each call takes 100ms, mostly waiting for the network (simulated here with `time.Sleep`):

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func callRestaurant(name string) {
	time.Sleep(100 * time.Millisecond) // waiting on the network
}

func main() {
	restaurants := []string{"pho-king", "wok-this-way", "thai-tanic"}

	start := time.Now()
	for _, r := range restaurants {
		callRestaurant(r)
	}
	fmt.Println("one by one:", time.Since(start).Round(100*time.Millisecond))

	start = time.Now()
	var wg sync.WaitGroup
	for _, r := range restaurants {
		wg.Go(func() { callRestaurant(r) })
	}
	wg.Wait()
	fmt.Println("concurrently:", time.Since(start).Round(100*time.Millisecond))
}
```

```text
one by one: 300ms
concurrently: 100ms
```

The concurrent version is three times faster, and it would be three times faster **even on a single-core machine**. While one goroutine waits for the network, it isn't using the CPU at all, so the others can run. The waits overlap.

This is the key insight of the whole course: most real programs spend their time *waiting* (for networks, disks, databases and people). Concurrency lets you wait for many things at once.

## Why Go is good at this

Many languages bolt concurrency on with threads, callbacks or `async`/`await`. Go builds it into the language:

- **Goroutines** are cheap. A new one starts with a stack of a few kilobytes, so a single program can run hundreds of thousands of them.
- **Channels** and `select` let goroutines communicate safely.
- The **runtime scheduler** decides which goroutine runs on which CPU core, so you describe the *structure* and Go handles the *execution*.

Go's motto is: *Don't communicate by sharing memory; share memory by communicating.* You'll see what that means (and when to ignore it) throughout this course.

## What's coming

Over the next chapters you'll learn how the scheduler works, how to avoid leaking goroutines, every corner of channels and `select`, cancellation with `context`, the `sync` and `sync/atomic` packages, the classic concurrency patterns, and how to hunt down races, deadlocks and leaks. You'll finish by testing all of it without flaky sleeps and assembling Dispatchly's dispatch engine.

## Further reading

- [Concurrency is not Parallelism](https://go.dev/blog/waza-talk) by Rob Pike
