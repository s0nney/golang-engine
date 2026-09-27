---
title: While Loops, Break and Continue
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	for i := range 6 {
      		if i == 4 {
      			break
      		}
      		if i%2 == 0 {
      			continue
      		}
      		fmt.Print(i, " ")
      	}
      	fmt.Println()
      }
      ```
    options:
      - text: '`1 3 5`'
      - text: '`0 2`'
      - text: '`1 3`'
        correct: true
      - text: '`0 1 2 3`'
    explanation: |
      Even numbers (0 and 2) hit `continue`, which skips to the next
      iteration without printing. Odd numbers 1 and 3 are printed. When `i`
      reaches 4, `break` ends the loop entirely, so 5 is never reached.
  - question: How do you write a loop that runs forever (until something inside it breaks out)?
    options:
      - text: '`while true { ... }`'
      - text: '`for { ... }`'
        correct: true
      - text: '`loop { ... }`'
      - text: '`for true; true; true { ... }`'
    explanation: |
      A `for` with nothing after it loops forever. Go has no `while` or
      `loop` keyword. Use `break` or `return` to get out.
---

Sometimes you don't know in advance how many times to loop. You just want to keep going **while** something is true. Other languages have a `while` keyword for that. Go just uses `for` again.

## `for` with only a condition

Drop the init and post parts and keep only the condition, and `for` behaves like a `while` loop:

```go
package main

import "fmt"

func main() {
	credits := 10
	cost := 3

	for credits >= cost {
		credits -= cost
		fmt.Println("sent! credits left:", credits)
	}
	fmt.Println("out of credits")
}
```

```text
sent! credits left: 7
sent! credits left: 4
sent! credits left: 1
out of credits
```

The loop keeps going as long as `credits >= cost` is true, however many iterations that takes.

## Infinite loops

Drop the condition too, and you get a loop that runs forever:

```go
for {
	// runs until something inside stops it
}
```

That sounds like a bug, but it's common in servers: "forever: wait for a request, handle it". To get out, use `break` (or `return` from the function).

## `break`: stop the loop

`break` immediately exits the innermost loop:

```go
package main

import "fmt"

func main() {
	attempt := 0
	for {
		attempt++
		fmt.Println("attempt", attempt)
		delivered := attempt == 3 // pretend the 3rd try works
		if delivered {
			break
		}
	}
	fmt.Println("delivered after", attempt, "attempts")
}
```

```text
attempt 1
attempt 2
attempt 3
delivered after 3 attempts
```

`break` works in every kind of `for` loop, not just infinite ones.

## `continue`: skip to the next iteration

`continue` skips the rest of the current iteration and jumps straight to the next one:

```go
package main

import "fmt"

func main() {
	for i := range 5 {
		if i == 2 {
			fmt.Println("customer", i, "opted out, skipping")
			continue
		}
		fmt.Println("reminder sent to customer", i)
	}
}
```

```text
reminder sent to customer 0
reminder sent to customer 1
customer 2 opted out, skipping
reminder sent to customer 3
reminder sent to customer 4
```

`continue` works like a guard clause for loops: handle the special case, skip ahead, and keep the main body un-nested.

## Nested loops and labels

Loops can live inside loops. `break` and `continue` only affect the **innermost** loop. To break out of an outer loop, give it a **label**:

```go
package main

import "fmt"

func main() {
outer:
	for batch := range 3 {
		for msg := range 3 {
			if batch == 1 && msg == 1 {
				fmt.Println("rate limited, stopping everything")
				break outer
			}
			fmt.Println("batch", batch, "message", msg)
		}
	}
}
```

```text
batch 0 message 0
batch 0 message 1
batch 0 message 2
batch 1 message 0
rate limited, stopping everything
```

Labels are rare, but when you need them, they're much cleaner than juggling extra `bool` variables.

## Summary of `for`'s shapes

```go
for i := 0; i < n; i++ { } // classic three-part loop
for i := range n { }       // count from 0 to n-1
for condition { }          // "while" loop
for { }                    // infinite loop
```

And later you'll use `for ... range` to loop over lists and maps. One keyword, many shapes.

## Further reading

- [A Tour of Go: For is Go's "while"](https://go.dev/tour/flowcontrol/3)
