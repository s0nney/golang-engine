---
title: For Loops
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	for i := 1; i < 10; i += 3 {
      		fmt.Print(i, " ")
      	}
      	fmt.Println()
      }
      ```
    options:
      - text: '`1 4 7 10`'
      - text: '`1 4 7`'
        correct: true
      - text: '`0 3 6 9`'
      - text: '`1 2 3 4 5 6 7 8 9`'
    explanation: |
      `i` starts at 1 and goes up by 3 each time: 1, 4, 7. Next it would be
      10, but the condition `i < 10` is false, so the loop stops before
      printing it.
  - question: |
      How many times does `send()` run?

      ```go
      for range 3 {
      	send()
      }
      ```
    options:
      - text: '2'
      - text: '3'
        correct: true
      - text: '4'
      - text: It doesn't compile, because there's no loop variable
    explanation: |
      `for range n` runs the loop body `n` times. You can leave out the loop
      variable when you don't need to know which iteration you're on.
---

Textio needs to send the same reminder to 500 customers. You're not going to write `send()` 500 times. You need a **loop**: code that repeats.

Go has exactly **one** loop keyword: `for`. It's flexible enough to cover every kind of loop that other languages need three or four keywords for.

## The three-part `for` loop

The classic loop has three parts separated by semicolons:

```go
package main

import "fmt"

func main() {
	for i := 0; i < 3; i++ {
		fmt.Println("sending message", i)
	}
}
```

```text
sending message 0
sending message 1
sending message 2
```

1. **Init** (`i := 0`) runs once, before the loop starts.
2. **Condition** (`i < 3`) is checked before every iteration. If it's `false`, the loop ends.
3. **Post** (`i++`) runs after every iteration.

So the loop goes: set `i` to 0 → check `0 < 3` → run body → `i++` → check `1 < 3` → run body → ... → check `3 < 3` is false → stop.

Programmers usually count from 0, and `i` (for "index") is the traditional name for a loop counter.

Like `if`, a `for` loop has no parentheses around its parts, and the braces are required.

## Counting in other ways

The post statement can do anything, not just `i++`:

```go
package main

import "fmt"

func main() {
	// Count down
	for i := 3; i > 0; i-- {
		fmt.Print(i, "... ")
	}
	fmt.Println("sent!")

	// Count in steps of 50
	for batch := 0; batch < 200; batch += 50 {
		fmt.Println("batch starting at", batch)
	}
}
```

```text
3... 2... 1... sent!
batch starting at 0
batch starting at 50
batch starting at 100
batch starting at 150
```

`fmt.Print` is like `Println` but without the newline at the end.

## Ranging over an integer

Counting from 0 up to some number is so common that Go 1.22 added a shorter way to write it. `for i := range n` loops with `i` taking the values `0, 1, ..., n-1`:

```go
package main

import "fmt"

func main() {
	for i := range 3 {
		fmt.Println("retry attempt", i+1)
	}
}
```

```text
retry attempt 1
retry attempt 2
retry attempt 3
```

This does exactly the same as `for i := 0; i < 3; i++`, with less to read and less to get wrong. Modern Go code uses it whenever it just needs to count up from zero.

If you don't need the counter at all, drop it:

```go
for range 5 {
	fmt.Println("ping")
}
```

## Adding things up

A common loop pattern is building up a total. Say the first message costs 1 cent, the second 2 cents, and so on (a terrible pricing model, but a good example):

```go
package main

import "fmt"

func main() {
	total := 0
	for i := range 5 {
		total += i + 1
	}
	fmt.Println("total cents:", total)
}
```

```text
total cents: 15
```

## Each iteration gets a fresh variable

Since Go 1.22, the loop variable (`i` above) is a **new variable on every iteration**, not one variable that's reused and updated. For simple loops you won't notice any difference. It matters when code captures the variable to use later (you'll see this with goroutines), and it fixed one of Go's most famous gotchas. If you see old code doing `i := i` inside a loop, that was a workaround you no longer need.

## Further reading

- [Go by Example: For](https://gobyexample.com/for)
- [A Tour of Go: For](https://go.dev/tour/flowcontrol/1)
