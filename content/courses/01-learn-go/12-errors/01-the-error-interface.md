---
title: The error Interface
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import (
      	"errors"
      	"fmt"
      )

      func charge(credits, cost int) (int, error) {
      	if cost > credits {
      		return credits, errors.New("insufficient credits")
      	}
      	return credits - cost, nil
      }

      func main() {
      	left, err := charge(5, 8)
      	if err != nil {
      		fmt.Println("error:", err, left)
      		return
      	}
      	fmt.Println("left:", left)
      }
      ```
    options:
      - text: '`left: -3`'
      - text: '`error: insufficient credits 5`'
        correct: true
      - text: '`error: insufficient credits -3`'
      - text: It panics, because the cost is too high
    explanation: |
      The cost (8) is more than the credits (5), so `charge` returns the
      original credits and a non-nil error. `main` checks `err`, prints it
      along with `left` (5), and returns early.
  - question: What must a type have to satisfy the `error` interface?
    options:
      - text: A field called `Message`
      - text: A method `Error() string`
        correct: true
      - text: It must be created with `errors.New`
      - text: It must be declared with the `implements error` keyword
    explanation: |
      `error` is an interface with one method, `Error() string`. Any type
      with that method is an error. Go has no `implements` keyword: types
      satisfy interfaces automatically.
---

Things go wrong. A phone number is invalid, a carrier is down, a customer runs out of credits. Many languages handle failure with **exceptions** that jump out of your code to some handler far away. Go does something simpler: **errors are ordinary values** that functions return.

## Returning an error

By convention, a function that can fail returns an `error` as its **last** return value. `nil` means "no error":

```go
package main

import (
	"errors"
	"fmt"
)

func validatePhone(phone string) error {
	if len(phone) < 8 {
		return errors.New("phone number too short")
	}
	if phone[0] != '+' {
		return errors.New("phone number must start with +")
	}
	return nil
}

func main() {
	for _, p := range []string{"+1-555-0100", "555", "15550100000"} {
		if err := validatePhone(p); err != nil {
			fmt.Println(p, "->", err)
			continue
		}
		fmt.Println(p, "-> ok")
	}
}
```

```text
+1-555-0100 -> ok
555 -> phone number too short
15550100000 -> phone number must start with +
```

`errors.New` creates a simple error from a message. (`phone[0]` is the first byte of the string, and `'+'` is a character literal.)

## Checking errors

The caller checks the error straight away, usually with `if err != nil`, and handles it before doing anything else:

```go
cost, err := quote(message)
if err != nil {
	return err
}
// use cost
```

You'll write this pattern constantly. It can feel repetitive, but it makes every possible failure **visible** right where it happens. There's no hidden control flow: if a function can fail, you can see it in its signature, and you can see exactly where it's handled.

When a function returns an error, the other return values are usually meaningless. Don't use them unless the documentation says you can.

Error messages are conventionally **lowercase** and **don't end with punctuation**, because they're often combined into longer messages, as you'll see in the next lesson.

## What is `error`, exactly?

`error` is an **interface**, a new kind of type. An interface lists methods, and **any type that has those methods satisfies the interface**. The `error` interface is built into Go and has just one method:

```go
type error interface {
	Error() string
}
```

So any type with an `Error() string` method *is* an error. There's no `implements` keyword. If the method is there, the type fits.

## Custom error types

Because `error` is an interface, you can create your own error types that carry extra information:

```go
package main

import "fmt"

type carrierError struct {
	carrier string
	code    int
}

func (e *carrierError) Error() string {
	return fmt.Sprintf("carrier %s rejected the message (code %d)", e.carrier, e.code)
}

func send(to string) error {
	return &carrierError{carrier: "Alpha Mobile", code: 503}
}

func main() {
	err := send("+1-555-0100")
	if err != nil {
		fmt.Println("send failed:", err)
	}
}
```

```text
send failed: carrier Alpha Mobile rejected the message (code 503)
```

`send` returns a `*carrierError`, but its return type is just `error`. The caller only needs to know "something went wrong, and here's the message". When you *do* need the extra fields, you can dig them out, which is the topic of the lesson after next.

`fmt.Println` calls the `Error()` method for you when it prints an error.

## Interfaces in general

`error` is your first interface, but it won't be your last. The idea is powerful: a function can accept "anything that can do X" without caring what type it really is. `fmt.Println` accepting any value at all is another example: its parameters have type `any`, the interface with no methods, which every type satisfies.

## Further reading

- [Go by Example: Errors](https://gobyexample.com/errors)
- [Go by Example: Custom Errors](https://gobyexample.com/custom-errors)
- [A Tour of Go: Errors](https://go.dev/tour/methods/19)
