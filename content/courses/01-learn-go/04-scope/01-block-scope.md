---
title: Block Scope
quiz:
  - question: |
      Why doesn't this program compile?

      ```go
      package main

      import "fmt"

      func main() {
      	credits := 5
      	if credits > 0 {
      		status := "active"
      	}
      	fmt.Println(status)
      }
      ```
    options:
      - text: '`credits` isn''t visible inside the `if` block'
      - text: '`status` only exists inside the `if` block, so it''s undefined where it''s printed'
        correct: true
      - text: '`fmt.Println` can''t print strings declared with `:=`'
    explanation: |
      A variable declared inside a block lives only until that block's closing
      `}`. `credits` is visible inside the `if` because inner blocks can see
      outer ones, but not the other way around. The compiler reports
      `undefined: status`.
  - question: 'True or false: a function can use a variable that was declared inside a *different* function.'
    options:
      - text: True, as long as the other function was called first
      - text: False, because each function's body is its own block
        correct: true
    explanation: |
      Variables declared inside a function are local to that function. To
      share a value between functions, pass it as an argument or return it.
exercise:
  starter: |
    package main

    import "fmt"

    // printReceipt can't see main's variables, so it always prints 0.
    func printReceipt() {
    	fmt.Println("receipt: 0 messages")
    }

    func main() {
    	sent := 42
    	printReceipt()

    	sent += 8
    	printReceipt()
    }
  solution: |
    package main

    import "fmt"

    func printReceipt(sent int) {
    	fmt.Println("receipt:", sent, "messages")
    }

    func main() {
    	sent := 42
    	printReceipt(sent)

    	sent += 8
    	printReceipt(sent)
    }
  expected_output: |
    receipt: 42 messages
    receipt: 50 messages
---

A variable's **scope** is the region of code where its name can be used. Outside its scope, the name simply doesn't exist, and the compiler will tell you so.

## Blocks

In Go, scope is based on **blocks**. A block is (usually) a pair of curly braces `{ }` and everything between them. Function bodies are blocks, `if` bodies are blocks, loop bodies are blocks.

The rule is simple:

> A variable declared inside a block can be used from where it's declared until the end of that block, including inside any blocks nested within it.

```go
package main

import "fmt"

func main() {
	sender := "Textio" // visible for the rest of main

	if len(sender) > 0 {
		greeting := "Hello from " + sender // sender is visible here
		fmt.Println(greeting)
	} // greeting's scope ends here

	fmt.Println(sender)
	// fmt.Println(greeting) would be: undefined: greeting
}
```

```text
Hello from Textio
Textio
```

Think of blocks as rooms with one-way windows. From an inner room you can see everything in the rooms around you. From outside, you can't see in.

## Declared before use

Inside a function, a variable only exists from the line where it's declared. You can't use it earlier, even in the same block:

```go
func main() {
	fmt.Println(count) // error: undefined: count
	count := 3
}
```

## Functions don't share local variables

Each function body is its own block. A variable declared in one function is invisible to every other function:

```go
package main

import "fmt"

func printBalance() {
	// fmt.Println(balance) would be: undefined: balance
	fmt.Println("no access to main's variables here")
}

func main() {
	balance := 20
	printBalance()
	fmt.Println(balance)
}
```

```text
no access to main's variables here
20
```

If `printBalance` needs the balance, pass it in as a parameter: `printBalance(balance)`. This is a feature, not a limitation. When you read a function, you know exactly where each of its values comes from: its parameters, its own local variables, or something declared at the package level (next lesson).

## Why small scopes are good

The smaller a variable's scope, the fewer lines of code can change it, and the less you have to keep in your head. Go programmers declare variables as close as possible to where they're used, often inside the `if` or loop that needs them.

```go
func send(message string) {
	// Only needed inside the if, so declare it there.
	if length := len(message); length > 160 {
		fmt.Println("splitting", length, "characters")
	}
}
```

That `if` with a declaration before the condition is a Go feature you'll meet properly in the conditionals chapter. The variable `length` exists only inside that `if` statement.

## Parameters are local variables too

A function's parameters are scoped to its body, just like variables declared inside it. Two different functions can both have a parameter called `message` without any conflict, because each lives in its own block.

## Your turn

`printReceipt` is supposed to show how many messages were sent, but `sent` lives in
`main`'s block, so `printReceipt` can't see it. Instead of guessing, it just prints `0`.

Give `printReceipt` an `int` parameter, use it in the message, and pass `sent` in
both calls. The program should print:

```text
receipt: 42 messages
receipt: 50 messages
```
