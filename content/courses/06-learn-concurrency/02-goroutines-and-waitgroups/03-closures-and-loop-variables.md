---
title: Closures and Loop Variables
quiz:
  - question: |
      In a module whose `go.mod` says `go 1.27`, what does this print?

      ```go
      results := make([]int, 3)
      var wg sync.WaitGroup
      for i := 0; i < 3; i++ {
      	wg.Go(func() { results[i] = i * 10 })
      }
      wg.Wait()
      fmt.Println(results)
      ```
    options:
      - text: '`[0 10 20]`'
        correct: true
      - text: '`[0 0 30]`'
      - text: It panics with an index out of range
      - text: The output changes from run to run
    explanation: |
      Since Go 1.22, every iteration of a `for` loop gets a fresh `i`. Each
      closure captures its own copy, so goroutine 0 writes `results[0] = 0`,
      and so on. Before Go 1.22 all three closures shared one `i`, which could
      already be 3 when they ran, causing exactly the panic in option three.
  - question: |
      Which line causes a data race?

      ```go
      var current string                      // line 1
      for _, c := range couriers {            // line 2
      	current = c                           // line 3
      	wg.Go(func() { ping(current) })       // line 4
      }
      ```
    options:
      - text: Line 2, because `c` is shared between iterations
      - text: 'Lines 3 and 4: the loop writes `current` while earlier goroutines read it'
        correct: true
      - text: None. Go 1.22 made loop closures safe
    explanation: |
      Go 1.22 only changed variables *declared by the `for` statement* itself.
      `current` is declared outside the loop, so every closure shares the same
      variable. The loop keeps writing it while goroutines read it: a data race,
      and couriers get pinged twice or not at all. Use `c` directly instead.
---

Goroutines are usually started with a closure: `wg.Go(func() { ... })`. Closures capture **variables**, not values. Combine that with loops and goroutines, and you get one of Go's most famous bugs. The good news is that Go 1.22 fixed the most common form of it. The bad news is that the rest are still there.

## The bug that used to bite everyone

Before Go 1.22, a `for` loop declared its variables **once**, and each iteration updated the same variable. So this code:

```go
for _, c := range []string{"ana", "ben", "cy"} {
	go func() { fmt.Println("pinging", c) }()
}
```

would often print `pinging cy` three times: by the time the goroutines ran, the loop had finished and `c` held its last value. Everyone learned to write the defensive copy `c := c` at the top of the loop body.

## Go 1.22: one variable per iteration

Since Go 1.22, each iteration of a `for` loop gets its **own** copy of the loop variables. That applies to `range` loops and three-clause loops alike. Each closure captures a different variable, so this is correct:

```go
package main

import (
	"fmt"
	"slices"
	"sync"
)

func main() {
	couriers := []string{"ana", "ben", "cy"}

	var mu sync.Mutex
	var pinged []string
	var wg sync.WaitGroup
	for _, c := range couriers {
		wg.Go(func() {
			mu.Lock()
			defer mu.Unlock()
			pinged = append(pinged, c)
		})
	}
	wg.Wait()

	slices.Sort(pinged) // goroutines finish in any order
	fmt.Println(pinged)
}
```

```text
[ana ben cy]
```

The `c := c` copy is now unnecessary, and `go fix` removes it for you.

The new behaviour depends on the **language version of the module**, the `go` line in `go.mod`. A module still declaring `go 1.21` keeps the old semantics even when built with Go 1.27. If you inherit an old codebase, check before you delete those copies!

## What Go 1.22 did *not* fix

Only variables declared **by the `for` statement** get per-iteration copies. Anything declared outside the loop is still shared by every closure:

```go
var current string
for _, c := range couriers {
	current = c
	wg.Go(func() { ping(current) }) // BUG: all goroutines share current
}
```

The loop writes `current` while goroutines from earlier iterations read it. That's a **data race**: some couriers get pinged twice, others never. The fix is to use the loop variable, or declare the variable *inside* the loop body so each iteration gets its own:

```go
for _, c := range couriers {
	msg := "you have a new order, " + c // declared per iteration
	wg.Go(func() { send(msg) })
}
```

The same trap appears with accumulators:

```go
total := 0
for _, o := range orders {
	wg.Go(func() { total += o.Price }) // BUG: data race on total
}
```

Every goroutine reads and writes the same `total`. Protect it with a mutex, use an atomic counter, or (often simplest) have each goroutine write into its own slot and add the slots up after `Wait`.

## Arguments vs captures

If you want a goroutine to work on a **snapshot** of a value, you can pass it as an argument, since arguments are evaluated when the `go` statement runs:

```go
for {
	o := nextOrder()
	go func(o Order) { dispatch(o) }(o) // explicit snapshot
}
```

With per-iteration variables this is rarely needed any more, but it's handy when the value comes from a variable you're about to modify.

## A mental checklist

When you write `go func() { ... }()` or `wg.Go(func() { ... })`, look at every variable the closure mentions and ask:

1. Was it declared by the loop header, or inside the loop body? Then it's private to this iteration. Good.
2. Was it declared outside the loop? Then it's shared. Is anyone writing to it while the goroutine runs? If so, you have a race.

## Further reading

- [Fixing For Loops in Go 1.22](https://go.dev/blog/loopvar-preview) on the Go blog
