---
title: Hello, World
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	// fmt.Println("Hello")
      	fmt.Println("Textio")
      }
      ```
    options:
      - text: '`Hello` and then `Textio`, on two lines'
      - text: '`Textio`'
        correct: true
      - text: Nothing, because the program has a comment in it
      - text: '`// fmt.Println("Hello")` and then `Textio`'
    explanation: |
      The first line inside `main` starts with `//`, which makes it a comment.
      Go ignores comments completely, so only `Textio` is printed.
  - question: Where does a Go program start running?
    options:
      - text: At the first line of the file
      - text: At the `import` statement
      - text: At the `main` function in the `main` package
        correct: true
      - text: At whichever function is defined last
    explanation: |
      Every Go program begins by running the function called `main` inside
      the package called `main`. When `main` finishes, the program ends.
exercise:
  starter: |
    package main

    import "fmt"

    func main() {
    	fmt.Println("Hello, World!")
    }
  solution: |
    package main

    import "fmt"

    func main() {
    	fmt.Println("Welcome to Textio!")
    	fmt.Println("Your first message is on its way.")
    }
  expected_output: |
    Welcome to Textio!
    Your first message is on its way.
---

It's tradition: the first program you write in any language prints "Hello, World". Here it is in Go:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
```

It prints:

```text
Hello, World!
```

Five lines of code, and every one of them matters. Let's go through them.

## `package main`

Go code is organised into **packages**, which are groups of related code. The package called `main` is special: it's the one that becomes a program you can run.

## `import "fmt"`

This line brings in the `fmt` package (short for "format") from Go's **standard library**, the big collection of packages that ships with Go. `fmt` knows how to print text.

## `func main() { ... }`

This defines a **function** named `main`. A function is a named block of instructions. The instructions live between the curly braces `{` and `}`.

`main` is special too: it's where your program **starts**. Go runs the code inside `main` from top to bottom, and when it reaches the closing `}`, the program ends.

## `fmt.Println("Hello, World!")`

This is the instruction that does the work. It calls the `Println` function from the `fmt` package ("print line") and hands it some text. The text in double quotes is called a **string**. `Println` prints it and then moves to a new line.

You can print as many lines as you like:

```go
package main

import "fmt"

func main() {
	fmt.Println("Welcome to Textio!")
	fmt.Println("Your message has been queued.")
}
```

```text
Welcome to Textio!
Your message has been queued.
```

## Comments

Sometimes you want to leave a note for humans in your code. That's what **comments** are for. Go ignores them completely.

```go
package main

import "fmt"

// main is where the program starts.
func main() {
	// Greet the new customer.
	fmt.Println("Welcome to Textio!")

	/*
		Block comments can span
		several lines.
	*/
}
```

- `//` starts a **line comment**. Everything after it on that line is ignored.
- `/* ... */` is a **block comment** and can span many lines. Go programmers mostly use `//`, even for long notes.

A good comment explains *why* the code does something, not *what* it does. The code already says what it does.

```go
// Bad: says what the code already says.
// Print hello.
fmt.Println("Hello")

// Good: explains the reason.
// Carriers reject empty messages, so always send a greeting.
fmt.Println("Hello")
```

A handy trick: putting `//` in front of a line of code "comments it out" so it doesn't run. That's useful while you experiment.

## Capital letters matter

Go is **case-sensitive**. `fmt.Println` works, but `fmt.println` or `FMT.Println` won't compile. Copy names exactly.

## Your turn

Textio's onboarding flow needs a welcome screen. Change the program so it prints exactly these two lines, in this order:

```text
Welcome to Textio!
Your first message is on its way.
```

Use one `fmt.Println` call per line. Capital letters and punctuation matter!

## Further reading

- [Go by Example: Hello World](https://gobyexample.com/hello-world)
