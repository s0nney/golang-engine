---
title: 'Review: The Basics'
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      const (
      	Free = iota
      	Pro
      	Enterprise
      )

      func price(plan int) float64 {
      	switch plan {
      	case Pro:
      		return 9.99
      	case Enterprise:
      		return 99
      	}
      	return 0
      }

      func main() {
      	total := 0.0
      	for i := range 3 {
      		total += price(i)
      	}
      	fmt.Printf("%.2f\n", total)
      }
      ```
    options:
      - text: '`9.99`'
      - text: '`108.99`'
        correct: true
      - text: '`108.990000`'
      - text: '`99.00`'
    explanation: |
      `range 3` gives `i` the values 0, 1 and 2, which are `Free`, `Pro` and
      `Enterprise`. The prices are 0, 9.99 and 99, adding up to 108.99, and
      `%.2f` prints exactly two decimal places.
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	count := 1
      	if count > 0 {
      		count := count * 10
      		count++
      	}
      	fmt.Println(count)
      }
      ```
    options:
      - text: '`11`'
      - text: '`10`'
      - text: '`1`'
        correct: true
      - text: It doesn't compile
    explanation: |
      `count := count * 10` inside the `if` declares a new variable that
      shadows the outer one. The new variable becomes 11 and disappears at
      the `}`. The outer `count` was never touched.
  - question: |
      Which line fails to compile?

      ```go
      a := 10         // line 1
      b := 2.5        // line 2
      c := float64(a) // line 3
      d := a * b      // line 4
      ```
    options:
      - text: Line 2
      - text: Line 3
      - text: Line 4
        correct: true
    explanation: |
      `a` is an `int` and `b` is a `float64`. Go never converts between types
      automatically, so `a * b` is a mismatched-types error. Line 3 shows the
      fix: convert explicitly with `float64(a)`.
exercise:
  starter: |
    package main

    import "fmt"

    const (
    	freeMessages  = 100
    	standardLimit = 1000
    )

    // monthlyBill returns what a customer pays, in cents, for sending
    // messagesSent messages in a month.
    func monthlyBill(messagesSent int) int {
    	// ?
    	return 0
    }

    func main() {
    	for _, n := range []int{50, 100, 101, 1000, 1500} {
    		fmt.Println(n, "messages:", monthlyBill(n), "cents")
    	}
    }
  solution: |
    package main

    import "fmt"

    const (
    	freeMessages  = 100
    	standardLimit = 1000
    )

    func monthlyBill(messagesSent int) int {
    	if messagesSent <= freeMessages {
    		return 0
    	}
    	if messagesSent <= standardLimit {
    		return (messagesSent - freeMessages) * 2
    	}
    	return (standardLimit-freeMessages)*2 + (messagesSent - standardLimit)
    }

    func main() {
    	for _, n := range []int{50, 100, 101, 1000, 1500} {
    		fmt.Println(n, "messages:", monthlyBill(n), "cents")
    	}
    }
  tests: |
    package main

    import "testing"

    func TestMonthlyBill(t *testing.T) {
    	for _, tc := range []struct {
    		sent, want int
    	}{
    		{0, 0}, {50, 0}, {100, 0},
    		{101, 2}, {150, 100}, {1000, 1800},
    		{1001, 1801}, {1500, 2300},
    	} {
    		if got := monthlyBill(tc.sent); got != tc.want {
    			t.Errorf("monthlyBill(%d) = %d, want %d", tc.sent, got, tc.want)
    		}
    	}
    }
---

Congratulations, you've made it through the whole course! These last three lessons are a review. Each one mixes questions from several chapters. If you get one wrong, go back and re-read the lesson it came from: that's the fastest way to make it stick.

This first review covers the foundations: variables and types, constants, functions, scope, conditionals and loops.

## Quick recap

Here's a small Textio program that uses most of what you learned in the first half of the course. Before you run it in your head, try to name every concept it uses.

```go
package main

import "fmt"

const maxSMSLength = 160

func segments(body string) (int, error) {
	if len(body) == 0 {
		return 0, fmt.Errorf("empty message")
	}
	return (len(body) + maxSMSLength - 1) / maxSMSLength, nil
}

func main() {
	bodies := []string{"hi", "", "Your code is 4821"}
	for i, b := range bodies {
		n, err := segments(b)
		if err != nil {
			fmt.Printf("message %d: %v\n", i, err)
			continue
		}
		fmt.Printf("message %d: %d segment(s)\n", i, n)
	}
}
```

```text
message 0: 1 segment(s)
message 1: empty message
message 2: 1 segment(s)
```

You should spot: a package-level constant, a function with multiple return values, a guard clause with an early return, a slice literal, a `for ... range` loop, `if` with `err != nil`, `continue`, and `Printf` verbs.

Remember the big gotchas from these chapters:

- `:=` declares, `=` assigns. A stray `:=` in an inner block creates a shadow.
- Integer division throws away the remainder.
- Go never converts types for you.
- `switch` cases don't fall through.
- Since Go 1.22, `for i := range n` counts from 0 to n-1, and every iteration gets a fresh loop variable.

Now take the quiz.

## Your turn

Time to use the basics together. Textio's new pricing plan works like this:

- The first **100** messages each month are free.
- Messages 101 to 1000 cost **2 cents** each.
- Every message after the 1000th costs **1 cent**.

Complete `monthlyBill` so it returns the total in cents. For example, 150 messages
cost 100 cents (50 paid messages at 2 cents) and 1500 messages cost 2300 cents
(900 × 2 + 500 × 1). Use the two constants rather than typing `100` and `1000`
again, and check the boundaries: exactly 100 messages is still free.
