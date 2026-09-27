---
title: Stateful Closures
quiz:
  - question: |
      Using `counter` from this lesson, what does this print?

      ```go
      a := counter()
      b := counter()
      a()
      a()
      fmt.Println(a(), b())
      ```
    options:
      - text: '`3 4`'
      - text: '`3 1`'
        correct: true
      - text: '`1 1`'
      - text: '`2 1`'
    explanation: |
      Each call to `counter()` creates a fresh `n`. `a` has been called three times
      when it's printed, so it returns 3. `b` has its own `n` and returns 1.
  - question: Why is a closure-based counter not safe to call from several goroutines at once?
    options:
      - text: Closures can't be passed to goroutines
      - text: The goroutines would all get separate copies of `n`
      - text: Two goroutines could run `n++` on the same captured variable at the same time, which is a data race
        correct: true
    explanation: |
      All goroutines share the one captured `n`. Unsynchronised reads and writes
      from several goroutines are a data race. Protect it with a `sync.Mutex` or use
      `sync/atomic`.
---

Because a closure can *change* the variables it captures, and those variables live
as long as the closure does, a closure can carry **state** between calls. It's like
a tiny object with one method and private fields that nobody else can reach.

## A counter

```go
func counter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}
```

Each call to `counter` creates a new `n` and a new closure that owns it. No one else
can read or change that `n`, not even code in the same package. That's stronger
privacy than an unexported struct field!

## Doc2Doc: numbering footnotes

Doc2Doc turns `[^]` markers into numbered footnotes. Each document gets its own
numbering, starting from 1:

```go
package main

import (
	"fmt"
	"strings"
)

func counter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}

func numberFootnotes(doc string) string {
	next := counter() // fresh numbering for each document
	var b strings.Builder
	for {
		before, after, found := strings.Cut(doc, "[^]")
		b.WriteString(before)
		if !found {
			break
		}
		fmt.Fprintf(&b, "[^%d]", next())
		doc = after
	}
	return b.String()
}

func main() {
	fmt.Println(numberFootnotes("Go[^] is fast[^] and fun[^]."))
	fmt.Println(numberFootnotes("A second doc[^]."))
}
```

```text
Go[^1] is fast[^2] and fun[^3].
A second doc[^1].
```

The second document starts at 1 again because it got a brand new counter.

## Closures vs structs

You could write the same thing as a struct:

```go
type Counter struct{ n int }

func (c *Counter) Next() int {
	c.n++
	return c.n
}
```

Which is better? It depends:

- A **closure** is lighter when there's a single operation. It can be passed anywhere
  a `func() int` is expected, with no interface needed.
- A **struct** is better when you need several operations (`Next`, `Reset`, `Peek`),
  want to inspect the state in a debugger, or need a zero value that works without a
  constructor.

## State is a side effect

A stateful closure is **not pure**: calling `next()` twice gives different answers.
That's fine, but be honest about it. Keep stateful closures local and short-lived,
like `next` inside `numberFootnotes`. From the outside, `numberFootnotes` is still
pure: same document in, same numbered document out, every time.

## Concurrency warning

If several goroutines call the same stateful closure, they race on the captured
variable. Guard it:

```go
func safeCounter() func() int {
	var mu sync.Mutex
	n := 0
	return func() int {
		mu.Lock()
		defer mu.Unlock()
		n++
		return n
	}
}
```

The mutex is captured too, so every call shares the same lock.

## Further reading

- [Go by Example: Closures](https://gobyexample.com/closures)
