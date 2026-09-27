---
title: Goroutines and WaitGroups
quiz:
  - question: |
      What does this program most likely print?

      ```go
      package main

      import "fmt"

      func main() {
      	go fmt.Println("sending...")
      	fmt.Println("done")
      }
      ```
    options:
      - text: '`sending...` and then `done`'
      - text: Usually just `done`, because `main` returns before the goroutine gets to run
        correct: true
      - text: '`done` and then `sending...`, always'
      - text: It doesn't compile, because `go` needs a `func` literal
    explanation: |
      `go` starts the call in a new goroutine and moves on immediately.
      `main` prints `done` and returns, and when `main` returns the whole
      program exits, taking any unfinished goroutines with it. Nothing waits
      for `sending...`.
  - question: What does `wg.Wait()` do?
    options:
      - text: Starts all the goroutines in the group
      - text: Blocks until every function started with `wg.Go` has returned
        correct: true
      - text: Pauses for one second
      - text: Cancels the goroutines that are still running
    explanation: |
      A `sync.WaitGroup` counts running tasks. `wg.Go(f)` adds one and runs
      `f` in a new goroutine; when `f` returns the count drops. `Wait`
      blocks until the count reaches zero.
---

Textio has 10,000 appointment reminders to send, and each carrier request takes about a second, mostly spent waiting for the network. Sending them one after another would take almost three hours. Instead, you want to send lots of them **at the same time**. That's **concurrency**, and it's where Go really shines.

## The `go` keyword

A **goroutine** is a function running independently, at the same time as the rest of your program. To start one, put `go` in front of a function call:

```go
go send("+1-555-0100", "Your appointment is tomorrow")
```

That's it. The `go` statement starts `send` running in the background and **immediately** moves on to the next line, without waiting for `send` to finish.

Goroutines are extremely cheap. Each starts with just a few kilobytes of memory, and a single program can easily run hundreds of thousands of them. Go's runtime spreads them across all your CPU cores for you.

## The problem: `main` doesn't wait

```go
package main

import "fmt"

func main() {
	go fmt.Println("sending...")
	fmt.Println("done")
}
```

Run this and you'll almost always see only:

```text
done
```

When `main` returns, the program exits, and any goroutines still running are killed on the spot. Something has to **wait** for them.

## `sync.WaitGroup`

A `sync.WaitGroup` waits for a group of goroutines to finish. Since Go 1.25, its `Go` method starts a goroutine *and* tracks it in one step:

```go
package main

import (
	"fmt"
	"sync"
)

func send(to string) {
	fmt.Println("sent to", to)
}

func main() {
	recipients := []string{"alice", "bob", "carol"}

	var wg sync.WaitGroup
	for _, r := range recipients {
		wg.Go(func() {
			send(r)
		})
	}
	wg.Wait()
	fmt.Println("all reminders sent")
}
```

One possible output:

```text
sent to carol
sent to alice
sent to bob
all reminders sent
```

- `var wg sync.WaitGroup` is ready to use: its zero value is an empty group.
- `wg.Go(f)` runs `f` in a new goroutine and counts it.
- `wg.Wait()` blocks until every counted function has returned.

`func() { send(r) }` is a **function literal**, a function without a name, written inline. It can use variables from around it, like `r`.

## The order is unpredictable

Notice the names came out in a different order from the slice. Goroutines run independently, and the Go scheduler decides who runs when. Run the program again and you may get another order. **Never assume goroutines run in any particular order.** If order matters, you need to coordinate them, which is what channels (next lesson) are for.

## Loop variables and goroutines

Each goroutine above uses `r`, the loop variable. Before Go 1.22, all iterations shared **one** `r` variable, so goroutines often all saw the *last* value and sent three reminders to carol. You'd see workarounds like `r := r` in old code. Since Go 1.22 each iteration gets a fresh variable, so the code above just works.

## The old way

Before Go 1.25, you'd write the counting by hand:

```go
wg.Add(1)
go func() {
	defer wg.Done()
	send(r)
}()
```

Forgetting `Add` or `Done` caused hangs or crashes. `wg.Go` does both for you, and `go fix` can rewrite the old pattern.

## Concurrency is not magic

Starting goroutines is easy. The hard part is when they **share data**. Two goroutines changing the same variable at the same time is a **data race**, and it causes bugs that appear randomly. The rest of this chapter covers the two tools Go gives you to stay safe: channels and mutexes.

## Further reading

- [Go by Example: Goroutines](https://gobyexample.com/goroutines)
- [Go by Example: WaitGroups](https://gobyexample.com/waitgroups)
- [A Tour of Go: Goroutines](https://go.dev/tour/concurrency/1)
