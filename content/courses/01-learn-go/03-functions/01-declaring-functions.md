---
title: Declaring Functions
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func messageCost(length int) float64 {
      	if length > 160 {
      		return 0.02
      	}
      	return 0.01
      }

      func main() {
      	fmt.Println(messageCost(200) + messageCost(20))
      }
      ```
    options:
      - text: '`0.02`'
      - text: '`0.04`'
      - text: '`0.03`'
        correct: true
      - text: '`220`'
    explanation: |
      `messageCost(200)` returns `0.02` because 200 is more than 160, and
      `messageCost(20)` returns `0.01`. The two results are added: `0.03`.
  - question: In `func greet(name string) string`, what does the second `string` mean?
    options:
      - text: The type of the value the function returns
        correct: true
      - text: The type of the `name` parameter
      - text: The name of the function's package
      - text: That the function can be called with two strings
    explanation: |
      After the parameter list comes the return type. `greet` takes one
      parameter, `name`, of type `string`, and hands back a `string`.
---

So far, all of your code has lived inside `main`. That works for tiny programs, but Textio's code base has thousands of lines. **Functions** let you give a chunk of code a name, then run it whenever you need it.

## Your first function

```go
package main

import "fmt"

func printWelcome() {
	fmt.Println("Welcome to Textio!")
	fmt.Println("Reply STOP to unsubscribe.")
}

func main() {
	printWelcome()
	printWelcome()
}
```

```text
Welcome to Textio!
Reply STOP to unsubscribe.
Welcome to Textio!
Reply STOP to unsubscribe.
```

- `func printWelcome() { ... }` **declares** (defines) a function called `printWelcome`.
- `printWelcome()` **calls** it: Go jumps into the function, runs its body, then comes back and carries on where it left off.

Declaring a function doesn't run it. Only calling it does. And the order you declare functions in a file doesn't matter: `main` can call functions written above or below it.

## Parameters: giving a function input

Functions become much more useful when you can hand them data. **Parameters** are variables listed in the parentheses, each with a name and a type:

```go
package main

import "fmt"

func greet(name string, code int) {
	fmt.Printf("Hi %s, your code is %d\n", name, code)
}

func main() {
	greet("Alice", 4821)
	greet("Bob", 1234)
}
```

```text
Hi Alice, your code is 4821
Hi Bob, your code is 1234
```

The values you pass in a call (`"Alice"`, `4821`) are called **arguments**. They're matched to parameters in order, and their types must match.

When several parameters in a row share a type, you can write the type once:

```go
func totalMessages(sent, failed int) { ... } // both are ints
```

## Return values: getting output back

A function can also **return** a value to whoever called it. Put the return type after the parameters, and use the `return` keyword:

```go
package main

import "fmt"

func messageCost(length int) float64 {
	if length > 160 {
		return 0.02
	}
	return 0.01
}

func main() {
	cost := messageCost(42)
	fmt.Println(cost)
	fmt.Println(messageCost(300))
}
```

```text
0.01
0.02
```

(You'll learn exactly how `if` works in the conditionals chapter. For now: if the message is longer than 160 characters, it costs more.)

`return` does two things: it hands back the value, and it **immediately exits** the function. Nothing after it runs.

## Reading a function signature

The first line of a function is its **signature**. It tells you everything you need to use the function:

```go
func messageCost(length int) float64
//   ^ name      ^ inputs    ^ output
```

Go reads types left to right, and puts the type *after* the name. That's the opposite of languages like C or Java, but it reads naturally: "`length` is an `int`".

## Why use functions?

- **Reuse.** Write it once, call it everywhere.
- **Naming.** `messageCost(length)` explains itself far better than the raw `if` statement inside it.
- **Testing.** Small functions with clear inputs and outputs are easy to test, as you'll see later.

## Further reading

- [Go by Example: Functions](https://gobyexample.com/functions)
- [A Tour of Go: Functions](https://go.dev/tour/basics/4)
