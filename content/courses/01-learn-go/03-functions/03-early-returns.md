---
title: Early Returns
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func status(credits int, blocked bool) string {
      	if blocked {
      		return "blocked"
      	}
      	if credits == 0 {
      		return "no credits"
      	}
      	return "ok"
      }

      func main() {
      	fmt.Println(status(0, true))
      }
      ```
    options:
      - text: '`no credits`'
      - text: '`blocked`'
        correct: true
      - text: '`ok`'
      - text: '`blocked` and then `no credits`'
    explanation: |
      `blocked` is `true`, so the first `return` runs. `return` exits the
      function immediately, so the credits check never happens.
  - question: What is a "guard clause"?
    options:
      - text: A comment that protects code from being changed
      - text: A check at the top of a function that returns early when something is wrong
        correct: true
      - text: A special keyword that stops a program from crashing
    explanation: |
      Guard clauses handle the unusual or invalid cases first and `return`
      straight away. The rest of the function can then focus on the normal
      case without deep nesting.
---

`return` doesn't just hand back a value. It **ends the function right away**. Go programmers use that on purpose to keep code flat and readable.

## Nesting gets messy

Say Textio should only send a message if the user has credits, the message isn't empty, and it isn't too long. Here's one way to write it:

```go
func canSend(credits int, message string) string {
	if credits > 0 {
		if len(message) > 0 {
			if len(message) <= 160 {
				return "ok"
			} else {
				return "message too long"
			}
		} else {
			return "message is empty"
		}
	} else {
		return "no credits"
	}
}
```

It works, but the "happy path" (the normal case) is buried three levels deep, and each error is far from the check that caused it.

## Guard clauses

Flip each check around. Handle the problem and **return early**. Whatever's left at the bottom is the happy path:

```go
package main

import "fmt"

func canSend(credits int, message string) string {
	if credits <= 0 {
		return "no credits"
	}
	if len(message) == 0 {
		return "message is empty"
	}
	if len(message) > 160 {
		return "message too long"
	}
	return "ok"
}

func main() {
	fmt.Println(canSend(0, "hi"))
	fmt.Println(canSend(5, ""))
	fmt.Println(canSend(5, "Your code is 4821"))
}
```

```text
no credits
message is empty
ok
```

These early checks are called **guard clauses**. Each one guards the rest of the function against a bad case. Read top to bottom, the function is a checklist: "no credits? stop. Empty? stop. Too long? stop. Otherwise, ok."

Same behaviour, no nesting, no `else`. This style is so common in Go that you'll hear it described as "keep the happy path on the left": the normal flow of the code runs down the left edge, and indented blocks are for exceptions.

## Early returns in functions without results

A function that returns nothing can still use a bare `return` to exit early:

```go
package main

import "fmt"

func sendReminder(name string, optedOut bool) {
	if optedOut {
		return
	}
	fmt.Println("Reminder sent to", name)
}

func main() {
	sendReminder("Alice", false)
	sendReminder("Bob", true)
}
```

```text
Reminder sent to Alice
```

Bob opted out, so the function returns before printing anything.

## Every path must return

If a function declares a return type, the compiler checks that *every* path through it ends in a `return`. This won't compile:

```go
func label(credits int) string {
	if credits > 0 {
		return "active"
	}
} // error: missing return
```

What should `label(0)` return? Go won't let you leave that question unanswered.

## Why this matters

When you reach the errors chapter, you'll write this pattern hundreds of times:

```go
result, err := doSomething()
if err != nil {
	return err
}
// carry on with result
```

That's a guard clause too: bail out early if something went wrong, otherwise keep going.
