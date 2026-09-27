---
title: Compiling and Running
quiz:
  - question: What is the main difference between `go run` and `go build`?
    options:
      - text: '`go run` only checks your code for mistakes and never runs it'
      - text: '`go build` compiles your program into an executable file you can run later, while `go run` compiles and runs it straight away'
        correct: true
      - text: '`go build` is for Windows and `go run` is for Linux and macOS'
      - text: There is no difference; they are two names for the same command
    explanation: |
      Both commands compile your code. `go run` compiles into a temporary
      location and runs the result immediately, which is handy while
      developing. `go build` writes an executable file you can run (or ship)
      whenever you like.
  - question: Why can a compiler catch some mistakes before your program ever runs?
    options:
      - text: Because it translates the whole program ahead of time and checks it as it goes
        correct: true
      - text: Because it runs the program secretly first to see if it crashes
      - text: Because it asks the Go team to review your code
    explanation: |
      A compiler reads and translates your *entire* program before it runs.
      Along the way it checks things like types and names, so many bugs
      (such as a typo in a function name) are reported up front instead of
      in the middle of sending a customer's text messages.
exercise:
  starter: |
    package main

    import "fmt"

    func main() {
    	fmt.Println("Compiling Textio...")
    }
  solution: |
    package main

    import "fmt"

    func main() {
    	fmt.Println("Compiling Textio...")
    	fmt.Println("Build complete!")
    	fmt.Println("Textio server starting...")
    }
  expected_output: |
    Compiling Textio...
    Build complete!
    Textio server starting...
---

Computers don't understand Go. Deep down, a processor only understands **machine code**: long sequences of numbers that mean things like "add these two values" or "jump to this instruction". Somebody has to translate your Go into machine code. There are two main ways languages do this.

## Interpreted languages

Languages like Python and JavaScript are usually **interpreted**. Another program, the **interpreter**, reads your code and carries it out line by line, *while* the program runs. You need the interpreter installed wherever your code runs.

## Compiled languages

Go is a **compiled** language. Before your program runs, a program called the **compiler** translates the *whole thing* into machine code and saves it as an **executable** file. After that, the executable runs on its own. No Go installation needed.

Compiling gives Go some big advantages:

- **Speed.** Machine code runs directly on the processor, so compiled programs are usually much faster than interpreted ones.
- **Early error checking.** The compiler checks your entire program before it runs. Misspelled a name? Used text where a number was expected? You find out immediately, not when a customer hits that line of code at 3 a.m.
- **Easy deployment.** A Go program compiles into a single file. Copy it to a server and run it.

Go's compiler is also famously *fast*, so you rarely wait long.

## `go run`

While you're writing code, the quickest way to try it is `go run`. Save this as `main.go`:

```go
package main

import "fmt"

func main() {
	fmt.Println("Textio server starting...")
}
```

Then, in a terminal in the same folder:

```text
$ go run main.go
Textio server starting...
```

`go run` compiles your program to a temporary location and runs it straight away. It's perfect for experimenting.

## `go build`

When you want a program you can keep, share or deploy, use `go build`:

```text
$ go build main.go
$ ./main
Textio server starting...
```

`go build` writes an executable (called `main` here, or `main.exe` on Windows). You can run it again and again without compiling each time, and you can copy it to a computer that doesn't have Go installed.

## When the compiler says no

If your code has a mistake, compiling fails and nothing runs. Here `Println` is misspelled:

```go
package main

import "fmt"

func main() {
	fmt.Printn("Textio server starting...")
}
```

```text
$ go run main.go
# command-line-arguments
./main.go:6:6: undefined: fmt.Printn
```

The error tells you the file, the line (6) and the column (6), and what went wrong. Get used to reading these messages. They're the compiler being helpful, not rude.

Go's compiler is strict in other ways too. It refuses to compile a program that imports a package it doesn't use, or declares a variable it never uses. That sounds fussy, but it keeps Go codebases tidy.

## Your turn

When you press **Run** or **Submit**, your program goes through exactly the steps in
this lesson: it's compiled first, and only runs if compiling succeeds.

Change the program so it prints these three lines, in this order:

```text
Compiling Textio...
Build complete!
Textio server starting...
```

Add one `fmt.Println` call per line. Capital letters and punctuation must match
exactly. While you're here, try misspelling `Println` and pressing **Run** to see
a real compile error, then fix it again.
