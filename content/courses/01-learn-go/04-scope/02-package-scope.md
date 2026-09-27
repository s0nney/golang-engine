---
title: Package Scope
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      var sent = 0

      func send() {
      	sent++
      }

      func main() {
      	send()
      	send()
      	fmt.Println(sent)
      }
      ```
    options:
      - text: '`0`'
      - text: '`1`'
      - text: '`2`'
        correct: true
      - text: It doesn't compile, because `send` can't see `sent`
    explanation: |
      `sent` is declared at the package level, so every function in the
      package shares the same variable. Each call to `send` adds one, so it
      ends at `2`.
  - question: Why can't you write `limit := 160` outside of any function?
    options:
      - text: Because package-level declarations must start with a keyword such as `var`, `const` or `func`
        correct: true
      - text: Because package-level variables must be constants
      - text: Because numbers can't be stored at the package level
    explanation: |
      The short declaration `:=` only works inside functions. At the package
      level, write `var limit = 160` (or `const limit = 160` if it never
      changes).
---

Variables declared inside a function are local to it. But you can also declare things **outside** of every function, at the top level of a file. Those names have **package scope**: every function in the package can use them.

## Package-level declarations

```go
package main

import "fmt"

const company = "Textio"

var messagesSent = 0

func send(to string) {
	messagesSent++
	fmt.Printf("%s -> %s\n", company, to)
}

func main() {
	send("Alice")
	send("Bob")
	fmt.Println("total sent:", messagesSent)
}
```

```text
Textio -> Alice
Textio -> Bob
total sent: 2
```

`messagesSent++` is shorthand for `messagesSent = messagesSent + 1`. (There's also `--` for subtracting one.)

Both `send` and `main` use the same `messagesSent` variable, because it's declared at the package level.

## Rules for package scope

- Package-level declarations must start with a keyword: `var`, `const`, `func`, `type` or `import`. **`:=` isn't allowed** outside a function.
- Order doesn't matter at the package level. A function can use a package variable declared further down the file.
- Package scope covers the **whole package**, even across multiple files. If `main.go` and `billing.go` are both `package main` in the same folder, they share package-level names. You'll learn more about this in the packages chapter.

Functions are package-level declarations too. That's why `main` can call `send` no matter where `send` appears in the file.

## Unused package variables are allowed

Unused *local* variables are compile errors, but unused package-level variables are not. Go only enforces the rule inside functions.

## Universe scope

There's one scope even bigger than the package: the **universe** block. It holds everything that's built into Go and available everywhere without an import: types like `int` and `string`, constants like `true`, `false` and `nil`, and functions like `len`, `min`, `max` and `append`.

## Use package variables sparingly

Package-level *constants* are great: they're read-only, so sharing them is safe. Package-level *variables* are riskier. Any function in the package can change them at any time, which makes code harder to follow and test. When you read a function that uses a package variable, you have to hunt around to see who else touches it.

Compare these two versions:

```go
// Hidden dependency: what does it read? Who changes it?
var costPerMessage = 0.01

func bill(messages int) float64 {
	return float64(messages) * costPerMessage
}
```

```go
// Everything it needs is in the signature.
func bill(messages int, costPerMessage float64) float64 {
	return float64(messages) * costPerMessage
}
```

The second one is easier to understand, reuse and test. A good rule of thumb: prefer constants and parameters, and reach for package variables only when you truly need shared state.
