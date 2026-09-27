---
title: Goroutines Recap
quiz:
  - question: |
      What does this print?

      ```go
      func main() {
      	order := "A1"
      	go notify(order)
      	order = "B2"
      	time.Sleep(10 * time.Millisecond)
      }

      func notify(o string) {
      	fmt.Println("notifying", o)
      }
      ```
    options:
      - text: '`notifying A1`'
        correct: true
      - text: '`notifying B2`'
      - text: Either one, depending on scheduling
      - text: Nothing, because `main` returns first
    explanation: |
      The arguments of a `go` statement are evaluated immediately, in the
      calling goroutine, just like a normal function call. `notify` receives
      a copy of `"A1"` before `order` is reassigned. The 10ms sleep gives it
      time to print (a sleep is a poor way to wait, but it works here).
  - question: A goroutine started by an HTTP handler panics, and nothing in that goroutine calls `recover`. What happens?
    options:
      - text: Only that goroutine dies; the rest of the server carries on
      - text: The panic travels to the goroutine that started it, where it can be recovered
      - text: The whole program crashes
        correct: true
      - text: The runtime restarts the goroutine
    explanation: |
      An unrecovered panic in *any* goroutine terminates the entire program.
      `recover` only works inside a deferred function in the *same*
      goroutine that panicked, so a parent can't catch a child's panic.
---

Course 01 introduced goroutines. Here's a quick refresher, plus three details that trip people up in production.

## The basics

Put `go` in front of a function call and that call runs in a new **goroutine**, concurrently with the code that started it:

```go
go assignCourier(order)       // named function
go func() { track(order) }()  // function literal
```

The `go` statement returns immediately. It doesn't wait for the function, and it gives you no way to get a return value back. To wait, you use a `sync.WaitGroup` (next lesson) or a channel. To get results back, you use channels or write into shared memory safely.

## Detail 1: when `main` returns, everything stops

A Go program ends when `main` returns. Any goroutines still running are killed on the spot, with no cleanup and no warning:

```go
func main() {
	go fmt.Println("courier assigned")
	fmt.Println("order placed")
}
```

This usually prints only `order placed`. The new goroutine never got a chance to run before the program exited. Every goroutine you start needs someone to wait for it, or a good reason why it doesn't matter.

## Detail 2: arguments are evaluated immediately

The function value and its arguments are evaluated in the *calling* goroutine, at the moment the `go` statement runs. Only the call itself happens later:

```go
package main

import (
	"fmt"
	"slices"
)

func notify(order string, ready <-chan struct{}, out chan<- string) {
	<-ready
	out <- "argument saw " + order
}

func main() {
	ready := make(chan struct{})
	out := make(chan string)

	order := "A1"
	go notify(order, ready, out) // order is evaluated right now
	go func() {
		<-ready
		out <- "closure saw " + order // order is read later, when this line runs
	}()
	order = "B2"
	close(ready) // let both goroutines continue

	got := []string{<-out, <-out}
	slices.Sort(got) // the two goroutines may finish in either order
	fmt.Println(got[0])
	fmt.Println(got[1])
}
```

```text
argument saw A1
closure saw B2
```

Passing a value as an argument takes a snapshot. A closure that *reads* a variable sees whatever is in it when it runs. You'll see why that difference matters in the lesson on closures.

## Detail 3: a panic in any goroutine kills the program

If a goroutine panics and doesn't recover, the **whole program** crashes, not just that goroutine. And `recover` only works in a deferred call inside the goroutine that panicked. The goroutine that started it can't catch the panic.

So if a goroutine runs code that might panic (a plugin, a user-supplied callback), it must protect itself:

```go
package main

import (
	"fmt"
	"sync"
)

func safely(name string, f func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%s crashed: %v\n", name, r)
		}
	}()
	f()
}

func main() {
	var wg sync.WaitGroup
	wg.Go(func() {
		safely("surge-pricing", func() {
			var prices map[string]int
			prices["downtown"] = 3 // panics: assignment to entry in nil map
		})
	})
	wg.Wait()
	fmt.Println("dispatcher still running")
}
```

```text
surge-pricing crashed: assignment to entry in nil map
dispatcher still running
```

Don't sprinkle `recover` everywhere, though. Most panics are bugs that *should* crash loudly. Recover only at boundaries where one bad task mustn't take down unrelated work.

## Goroutines have no identity

Unlike threads in some languages, goroutines have no ID you can access, no name, and no handle to kill them. That's deliberate. The only way to stop a goroutine is to **ask it to stop**, by closing a channel or cancelling a `context`, and to write the goroutine so it listens. You'll get very good at this over the next few chapters.

## Further reading

- [Go by Example: Goroutines](https://gobyexample.com/goroutines)
- [Effective Go: Goroutines](https://go.dev/doc/effective_go#goroutines)
