---
title: Panic, Defer and Recover
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	defer fmt.Println("one")
      	defer fmt.Println("two")
      	fmt.Println("three")
      }
      ```
    options:
      - text: '`one`, `two`, `three`'
      - text: '`three`, `one`, `two`'
      - text: '`three`, `two`, `one`'
        correct: true
      - text: '`two`, `one`, `three`'
    explanation: |
      Deferred calls run when the function returns, in reverse order (last
      in, first out). `three` prints first, then the deferred calls run:
      `two`, then `one`.
  - question: A user types an invalid phone number into Textio's signup form. Should your code `panic` or return an `error`?
    options:
      - text: Panic, because the input is wrong
      - text: Return an error, because bad input is an expected problem the caller can handle
        correct: true
      - text: Neither; ignore it
    explanation: |
      Bad input, network failures and missing files are normal, expected
      problems. Return an error. Panics are for bugs and truly impossible
      situations, where the program can't sensibly continue.
---

You've already seen a **panic**: dereferencing a nil pointer, indexing past the end of a slice, or writing to a nil map. A panic is Go's emergency stop. So when should *you* panic, and how is it different from returning an error?

## `panic`

You can trigger a panic yourself with the built-in `panic` function:

```go
func mustPositive(n int) int {
	if n <= 0 {
		panic("n must be positive")
	}
	return n
}
```

When a panic happens, the current function stops, deferred calls run (more on those in a moment), then its caller stops, and so on up the chain. If nothing stops it, the program crashes with the panic message and a stack trace.

## Errors versus panics

The rule of thumb in Go is simple:

- **Return an error** for anything that can reasonably go wrong: bad user input, a network timeout, a missing file, a carrier rejecting a message. These are *expected*, and the caller should decide what to do.
- **Panic** only for **bugs** and situations that should be impossible, where continuing would be dangerous or meaningless: a corrupt internal state, a programmer calling a function in a way it was never meant to be called.

If you're unsure, return an error. Library code in particular should almost never panic, because it takes the decision away from the caller.

A common exception is code that runs at startup. Functions named `MustSomething` panic on failure and are used when failure means the program can't start at all, like a hard-coded configuration being invalid:

```go
var phonePattern = regexp.MustCompile(`^\+[0-9-]+$`) // panics if the pattern is broken
```

## `defer`

Before looking at recovering from panics, you need `defer`. A **deferred** function call is postponed until the surrounding function returns, however it returns: normally, early, or by panicking.

```go
package main

import "fmt"

func sendBatch() {
	fmt.Println("open carrier connection")
	defer fmt.Println("close carrier connection")

	fmt.Println("sending 3 messages")
}

func main() {
	sendBatch()
}
```

```text
open carrier connection
sending 3 messages
close carrier connection
```

`defer` keeps setup and cleanup next to each other, so you can't forget the cleanup no matter how many `return` statements the function has. You'll use it for closing files, closing network connections and unlocking mutexes. When a function defers several calls, they run in **reverse order**, last in first out.

## `recover`

`recover` stops a panic in its tracks. It only works inside a deferred function. If the function is panicking, `recover` returns the panic value and the function returns normally instead of crashing:

```go
package main

import "fmt"

func safeSend(to string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("send to %s panicked: %v", to, r)
		}
	}()

	var carriers map[string]int
	carriers[to]++ // bug: nil map write panics
	return nil
}

func main() {
	err := safeSend("+1-555-0100")
	fmt.Println("recovered:", err)
	fmt.Println("the server keeps running")
}
```

```text
recovered: send to +1-555-0100 panicked: assignment to entry in nil map
the server keeps running
```

The deferred function uses a **named return value** (`err`) so it can change what `safeSend` returns after the panic.

Recovering is rare in everyday code. Its main job is at the edges of a program: a web server recovers from a panic in one request handler so one bug doesn't take down every customer's connection. Go's standard `net/http` server does exactly that. Don't use `panic` and `recover` as a substitute for returning errors. That's what errors are for.

## Further reading

- [Go by Example: Panic](https://gobyexample.com/panic)
- [Go by Example: Defer](https://gobyexample.com/defer)
- [Go by Example: Recover](https://gobyexample.com/recover)
