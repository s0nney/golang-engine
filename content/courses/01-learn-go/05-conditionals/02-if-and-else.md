---
title: If and Else
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	length := 160
      	if length > 160 {
      		fmt.Println("split")
      	} else if length == 160 {
      		fmt.Println("exactly full")
      	} else if length > 100 {
      		fmt.Println("long")
      	} else {
      		fmt.Println("short")
      	}
      }
      ```
    options:
      - text: '`split`'
      - text: '`exactly full`'
        correct: true
      - text: '`exactly full` and then `long`'
      - text: '`long`'
    explanation: |
      Go checks each condition from top to bottom and runs only the **first**
      branch whose condition is true. `160 > 160` is false, `160 == 160` is
      true, so it prints `exactly full` and skips the rest, even though
      `length > 100` is also true.
  - question: |
      In this code, where can `cost` be used?

      ```go
      if cost := price(msg); cost > 1 {
      	// A
      } else {
      	// B
      }
      // C
      ```
    options:
      - text: Only at A
      - text: At A and B
        correct: true
      - text: At A, B and C
      - text: Only at C
    explanation: |
      A variable declared in an `if`'s init statement is scoped to the whole
      `if` statement, including every `else if` and `else` branch. It
      disappears after the final `}`, so it isn't available at C.
---

Now that you can ask yes-or-no questions, you can make your program **do different things** depending on the answer.

## `if`

An `if` statement runs a block of code only when its condition is `true`:

```go
package main

import "fmt"

func main() {
	credits := 3
	if credits < 5 {
		fmt.Println("Low balance: please top up")
	}
	fmt.Println("done")
}
```

```text
Low balance: please top up
done
```

Some Go rules:

- No parentheses around the condition: `if credits < 5`, not `if (credits < 5)`.
- The curly braces are **always** required, even for one line.
- The opening `{` must be on the same line as the `if`.
- The condition must be a `bool`. `if credits { ... }` doesn't compile; write `if credits != 0`.

## `else`

`else` runs when the condition is false:

```go
if credits > 0 {
	fmt.Println("sending...")
} else {
	fmt.Println("out of credits")
}
```

The `else` goes on the same line as the closing `}` of the `if`.

## `else if`

To check several conditions in order, chain them with `else if`. Go tries each one from the top and runs **only the first** branch that matches:

```go
package main

import "fmt"

func tier(messagesPerMonth int) string {
	if messagesPerMonth > 10000 {
		return "enterprise"
	} else if messagesPerMonth > 1000 {
		return "business"
	} else if messagesPerMonth > 0 {
		return "starter"
	} else {
		return "inactive"
	}
}

func main() {
	fmt.Println(tier(50000))
	fmt.Println(tier(2500))
	fmt.Println(tier(0))
}
```

```text
enterprise
business
inactive
```

Order matters! If the `> 0` check came first, every active customer would be labelled "starter".

Also notice: each branch here ends in `return`, so the `else` keywords aren't really needed. Go style (remember guard clauses?) usually drops them:

```go
func tier(messagesPerMonth int) string {
	if messagesPerMonth > 10000 {
		return "enterprise"
	}
	if messagesPerMonth > 1000 {
		return "business"
	}
	if messagesPerMonth > 0 {
		return "starter"
	}
	return "inactive"
}
```

## `if` with an init statement

Go lets you run a short statement **before** the condition, separated by a semicolon. It's usually used to declare a variable that only the `if` needs:

```go
package main

import "fmt"

func cost(message string) float64 {
	if length := len(message); length > 160 {
		return 0.02
	}
	return 0.01
}

func main() {
	fmt.Println(cost("short and sweet"))
}
```

```text
0.01
```

The variable declared in the init statement (`length` here) is scoped to the `if` statement: it's available in the condition, the body, and any `else if` and `else` branches, and then it's gone. That keeps it from cluttering the rest of the function.

You'll see this pattern constantly with errors and map lookups:

```go
if err := send(msg); err != nil {
	// handle the error
}
```

Don't worry about those details yet. Just recognise the shape: *statement; condition*.

## Further reading

- [Go by Example: If/Else](https://gobyexample.com/if-else)
- [A Tour of Go: If with a short statement](https://go.dev/tour/flowcontrol/6)
