---
title: Declaring Variables
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	messagesSent := 10
      	messagesSent = messagesSent + 5
      	fmt.Println(messagesSent)
      }
      ```
    options:
      - text: '`10`'
      - text: '`5`'
      - text: '`15`'
        correct: true
      - text: It doesn't compile, because `messagesSent` is assigned twice
    explanation: |
      `:=` creates the variable with the value `10`. The next line uses plain
      `=` to *reassign* it: it reads the current value (10), adds 5 and stores
      the result, 15, back into the same variable.
  - question: |
      Why won't this code compile?

      ```go
      x := 1
      x := 2
      ```
    options:
      - text: Variable names must be longer than one letter
      - text: You can't store `2` in a variable that held `1`
      - text: '`:=` declares a *new* variable, and `x` already exists in this scope'
        correct: true
    explanation: |
      `:=` means "declare and assign". Declaring the same name twice in the
      same scope is an error. To change an existing variable, use plain `=`:
      `x = 2`.
---

Programs need to remember things: a customer's name, how many messages they've sent, whether their account is active. A **variable** is a named box that holds a value.

## Declaring with `var`

The most explicit way to create a variable uses the `var` keyword, a name, and a **type** (what kind of value the box holds):

```go
package main

import "fmt"

func main() {
	var customerName string = "Alice"
	var messagesSent int = 42
	fmt.Println(customerName, messagesSent)
}
```

```text
Alice 42
```

`fmt.Println` can print several values at once. It puts a space between them.

## The short declaration `:=`

Inside a function, Go programmers almost always use the **short variable declaration** operator `:=` instead. Go looks at the value on the right and works out the type for you:

```go
customerName := "Alice" // Go infers string
messagesSent := 42      // Go infers int
```

This is exactly the same as the `var` version above, just shorter. Read `:=` as "is declared as".

You still need `var` in a couple of places:

- **Outside of functions**, where `:=` isn't allowed.
- When you want to declare a variable **without** a value yet: `var total int`.

## Changing a variable

Once a variable exists, use a plain `=` to give it a new value:

```go
package main

import "fmt"

func main() {
	balance := 100
	fmt.Println(balance)

	balance = 75
	fmt.Println(balance)
}
```

```text
100
75
```

Remember the difference:

- `:=` **declares** a new variable (and sets its first value).
- `=` **assigns** a new value to a variable that already exists.

## Declaring several at once

You can declare multiple variables on one line:

```go
sender, recipient := "Textio", "Bob"
```

## Types are fixed

Every variable in Go has a type, and it can never change. Once `messagesSent` is an `int`, it holds whole numbers forever. This won't compile:

```go
messagesSent := 42
messagesSent = "lots" // error: cannot use "lots" (untyped string constant) as int value in assignment
```

This is called **static typing**, and it's one of the ways the compiler catches mistakes before your code runs.

## Unused variables are errors

Go refuses to compile a program with a variable that's declared but never used:

```go
func main() {
	cost := 2 // error: declared and not used: cost
}
```

It feels strict at first, but it catches real bugs, like calculating something and then forgetting to use it.

## Naming variables

Go style uses **camelCase**: start lowercase and capitalise each new word, like `messagesSent` or `customerName`. Names can contain letters, digits and underscores, but can't start with a digit. Pick names that say what the value *means*: `costPerMessage` beats `c`.

## Further reading

- [Go by Example: Variables](https://gobyexample.com/variables)
- [A Tour of Go: Short variable declarations](https://go.dev/tour/basics/10)
