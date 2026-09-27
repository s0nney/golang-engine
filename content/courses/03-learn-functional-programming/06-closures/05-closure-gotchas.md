---
title: Closure Gotchas
quiz:
  - question: |
      What does this print?

      ```go
      func main() {
          status := "starting"
          defer fmt.Println("A:", status)
          defer func() { fmt.Println("B:", status) }()
          status = "done"
      }
      ```
    options:
      - text: |
          `A: starting` then `B: done`
      - text: |
          `B: done` then `A: done`
      - text: |
          `B: starting` then `A: starting`
      - text: |
          `B: done` then `A: starting`
        correct: true
    explanation: |
      Deferred calls run last-in, first-out, so B prints first. A deferred call's
      *arguments* are evaluated immediately, so A captured `"starting"`. The
      deferred closure reads `status` when it runs, and sees `"done"`.
  - question: A long-lived closure captures a 50 MB `[]byte` but only uses its length. What's the problem?
    options:
      - text: The closure can't read the slice's length
      - text: The whole 50 MB backing array stays in memory as long as the closure is alive
        correct: true
      - text: Go copies the 50 MB slice every time the closure is called
    explanation: |
      Captured variables stay alive as long as the closure does. Compute `n := len(data)`
      first and capture only `n`, so the garbage collector can free the big array.
---

Closures are powerful, and that power comes with a few traps. Here are the ones that
bite Go programmers most often.

## 1. Captured variables change

A closure reads the *current* value of a captured variable when it runs, not the
value at the time it was created. If the variable changes in between, so does the
closure's behaviour:

```go
prefix := "> "
quote := func(s string) string { return prefix + s }
prefix = "| " // every later call to quote uses "| "
```

If you want a snapshot, copy the value into a new variable that nothing else touches,
or pass it as a parameter to a factory function, which gives the closure its own
private copy:

```go
func quoter(prefix string) func(string) string {
	return func(s string) string { return prefix + s }
}
```

## 2. defer arguments vs deferred closures

`defer f(x)` evaluates `x` **right away** and runs `f` later.
`defer func() { f(x) }()` evaluates `x` **later**, when the closure runs. Mixing them
up leads to confusing logs:

```go
package main

import "fmt"

func convert(name string) (err error) {
	status := "started"
	defer fmt.Println("[arg]    ", name, status)
	defer func() { fmt.Println("[closure]", name, status, err) }()

	status = "failed"
	return fmt.Errorf("bad header in %s", name)
}

func main() {
	convert("report.md")
}
```

```text
[closure] report.md failed bad header in report.md
[arg]     report.md started
```

The closure form is the one that can see the final `status`, and even the named
result `err`. That's why deferred error handling always uses a closure.

## 3. Goroutines sharing a captured variable

Closures started with `go` run concurrently. If they write to a shared captured
variable without synchronisation, that's a data race:

```go
total := 0
var wg sync.WaitGroup
for _, doc := range docs {
	wg.Go(func() {
		total += len(strings.Fields(doc)) // DATA RACE on total
	})
}
wg.Wait()
```

The loop variable `doc` is fine (it's per-iteration since Go 1.22), but `total` is
shared. Use a `sync.Mutex`, `atomic.Int64`, or have each goroutine send its result
on a channel. `go test -race` finds these bugs for you.

## 4. Closures keep things alive

Everything a closure captures stays in memory as long as the closure is reachable.
Capture a huge slice to use just one number from it, then store that closure in a
long-lived map of handlers, and the huge slice can never be freed:

```go
data, _ := os.ReadFile("huge.log")
n := len(data)
sizeOf := func() int { return n } // captures only n, so data can be freed
```

## 5. Recursive closures need a declaration first

A closure can't refer to itself by the name it's being assigned to with `:=`,
because the variable doesn't exist yet. Declare it first:

```go
var walk func(n *Node) int
walk = func(n *Node) int {
	total := 1
	for _, c := range n.Children {
		total += walk(c)
	}
	return total
}
```

With these five in mind, closures become one of the most useful tools in your Go
toolbox.
