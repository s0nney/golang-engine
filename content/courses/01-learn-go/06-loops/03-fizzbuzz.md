---
title: 'Practice: FizzBuzz'
quiz:
  - question: |
      In this FizzBuzz solution, what does it print for `15`?

      ```go
      for i := range 20 {
      	n := i + 1
      	switch {
      	case n%3 == 0:
      		fmt.Println("Fizz")
      	case n%5 == 0:
      		fmt.Println("Buzz")
      	case n%15 == 0:
      		fmt.Println("FizzBuzz")
      	default:
      		fmt.Println(n)
      	}
      }
      ```
    options:
      - text: '`FizzBuzz`'
      - text: '`Fizz`'
        correct: true
      - text: '`Buzz`'
      - text: '`Fizz` and then `FizzBuzz`'
    explanation: |
      This is the classic FizzBuzz bug. A `switch` runs the *first* case
      that matches. 15 is divisible by 3, so `Fizz` wins and the `n%15`
      case is never reached. The most specific check must come first.
  - question: What is `17 % 5`?
    options:
      - text: '`3`'
      - text: '`3.4`'
      - text: '`2`'
        correct: true
      - text: '`0`'
    explanation: |
      `%` gives the remainder after division. 5 goes into 17 three times
      (15), leaving 2.
exercise:
  starter: |
    package main

    import "fmt"

    // flagged returns how many of the messages numbered 1 to total
    // are sent through the backup carrier (every 3rd message) or
    // logged (every 5th message). Count each message at most once.
    func flagged(total int) int {
    	count := 0
    	// ?
    	return count
    }

    func main() {
    	fmt.Println(flagged(15)) // want 7
    }
  solution: |
    package main

    import "fmt"

    func flagged(total int) int {
    	count := 0
    	for n := 1; n <= total; n++ {
    		if n%3 == 0 || n%5 == 0 {
    			count++
    		}
    	}
    	return count
    }

    func main() {
    	fmt.Println(flagged(15))
    }
  tests: |
    package main

    import "testing"

    func TestFlagged(t *testing.T) {
    	tests := []struct {
    		total int
    		want  int
    	}{
    		{15, 7},
    		{0, 0},
    		{2, 0},
    		{3, 1},
    		{5, 2},
    		{30, 14},
    		{100, 47},
    	}
    	for _, tt := range tests {
    		if got := flagged(tt.total); got != tt.want {
    			t.Errorf("flagged(%d) = %d, want %d", tt.total, got, tt.want)
    		}
    	}
    }
---

Time to put conditionals and loops together. FizzBuzz is a famous little programming puzzle, and it's often used in job interviews to check that someone can actually write a loop.

## The rules

Print the numbers from 1 to 15, but:

- For multiples of 3, print `Fizz` instead of the number.
- For multiples of 5, print `Buzz` instead.
- For multiples of **both** 3 and 5, print `FizzBuzz`.

Try working it out in your head before reading on. What tools do you need?

## Checking "is a multiple of"

A number is a multiple of 3 if dividing by 3 leaves **no remainder**. That's exactly what the remainder operator `%` tells you:

```go
fmt.Println(9 % 3)  // 0, so 9 is a multiple of 3
fmt.Println(10 % 3) // 1, so 10 isn't
```

So "`n` is a multiple of 3" is written `n%3 == 0`.

## A solution

```go
package main

import "fmt"

func main() {
	for i := range 15 {
		n := i + 1
		switch {
		case n%15 == 0:
			fmt.Println("FizzBuzz")
		case n%3 == 0:
			fmt.Println("Fizz")
		case n%5 == 0:
			fmt.Println("Buzz")
		default:
			fmt.Println(n)
		}
	}
}
```

```text
1
2
Fizz
4
Buzz
Fizz
7
8
Fizz
Buzz
11
Fizz
13
14
FizzBuzz
```

Notes:

- `range 15` counts 0 to 14, so we add 1 to get 1 to 15. (A three-part loop, `for n := 1; n <= 15; n++`, works just as well.)
- A value-less `switch` checks each condition in order and runs the first match.
- **Order matters.** A multiple of 15 is also a multiple of 3 and of 5. If the `n%3` case came first, 15 would print `Fizz`. Always check the most specific condition first.
- `n%15 == 0` works because a number divisible by both 3 and 5 is divisible by 15. You could also write `n%3 == 0 && n%5 == 0`.

## The Textio version

Textio's product team has a similar rule for batching messages. Every 3rd message is sent through a backup carrier, every 5th message is logged for quality checks, and messages that are both do both. Here it is, written with `if` statements that build up a string instead of a `switch`:

```go
package main

import "fmt"

func main() {
	for n := 1; n <= 15; n++ {
		label := ""
		if n%3 == 0 {
			label += " backup"
		}
		if n%5 == 0 {
			label += " logged"
		}
		if label == "" {
			continue
		}
		fmt.Printf("message %d:%s\n", n, label)
	}
}
```

```text
message 3: backup
message 5: logged
message 6: backup
message 9: backup
message 10: logged
message 12: backup
message 15: backup logged
```

This version avoids the ordering problem completely: each rule is checked independently and their effects add up. Messages with no label hit `continue` and are skipped.

## Why this matters

FizzBuzz looks trivial, but it tests three core skills at once:

1. Writing a loop with the right start and end points (off-by-one errors are *very* common).
2. Choosing and ordering conditions correctly.
3. Using `%` to spot patterns in numbers.

If you can write it from memory without looking, you're in great shape for the rest of the course. Try it!

## Your turn

Using the same rules as the Textio version above (every 3rd message goes through the backup carrier, every 5th is logged), complete `flagged` so it returns **how many** of the messages numbered `1` to `total` are backed up, logged or both. A message that's both still counts once.

For example, from 1 to 15 the flagged messages are 3, 5, 6, 9, 10, 12 and 15, so `flagged(15)` returns `7`. Watch out for off-by-one errors: message `total` itself counts!
