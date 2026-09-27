---
title: Closures and Loop Variables
quiz:
  - question: |
      In Go 1.22 or later (with `go 1.22` or higher in `go.mod`), what does this print?

      ```go
      var fs []func()
      for _, ext := range []string{".md", ".txt", ".html"} {
          fs = append(fs, func() { fmt.Print(ext, " ") })
      }
      for _, f := range fs {
          f()
      }
      ```
    options:
      - text: '`.html .html .html `'
      - text: '`.md .md .md `'
      - text: '`.md .txt .html `'
        correct: true
      - text: It doesn't compile
    explanation: |
      Since Go 1.22, each iteration gets its own `ext` variable, so each closure
      captures a different one. Before 1.22 there was a single shared `ext`, and all
      three closures printed `.html`.
  - question: |
      What does this print, even in Go 1.27?

      ```go
      ext := ""
      var fs []func()
      for _, e := range []string{".md", ".txt"} {
          ext = e
          fs = append(fs, func() { fmt.Print(ext, " ") })
      }
      for _, f := range fs {
          f()
      }
      ```
    options:
      - text: '`.md .txt `'
      - text: '`.md .md `'
      - text: '`.txt .txt `'
        correct: true
    explanation: |
      The Go 1.22 change only affects variables *declared by the loop*. Here `ext`
      is declared once, outside the loop, so both closures share it and both see
      its final value, `.txt`.
---

For over a decade, one bug bit almost every Go programmer. It involved closures
and loops, and Go 1.22 finally fixed it. You'll still meet the old behaviour in
older code and blog posts, so it's worth understanding.

## The old bug

Before Go 1.22, a `for` loop declared its variables **once**, and each iteration just
assigned a new value to that same variable. Since closures capture *variables*, every
closure created in the loop captured the very same one:

```go
// Before Go 1.22
var printers []func()
for _, name := range []string{"intro.md", "setup.md", "faq.md"} {
	printers = append(printers, func() { fmt.Println(name) })
}
for _, p := range printers {
	p()
}
// faq.md
// faq.md
// faq.md
```

By the time the closures ran, the loop had finished, and `name` held its last value.
The same bug struck with goroutines started inside loops, where it was even harder to
spot. The standard fix was a strange-looking copy:

```go
for _, name := range names {
	name := name // make a fresh variable for this iteration
	printers = append(printers, func() { fmt.Println(name) })
}
```

## The Go 1.22 fix

Since Go 1.22, **each iteration of a `for` loop has its own fresh copy of the loop
variables**. This applies to `range` loops and to three-clause loops like
`for i := 0; i < n; i++`. Closures now capture what you'd expect:

```go
package main

import "fmt"

func main() {
	var printers []func()
	for _, name := range []string{"intro.md", "setup.md", "faq.md"} {
		printers = append(printers, func() { fmt.Println("rendering", name) })
	}
	for _, p := range printers {
		p()
	}

	var squares []func() int
	for i := 0; i < 3; i++ {
		squares = append(squares, func() int { return i * i })
	}
	for _, sq := range squares {
		fmt.Print(sq(), " ")
	}
	fmt.Println()
}
```

```text
rendering intro.md
rendering setup.md
rendering faq.md
0 1 4 
```

Delete any `name := name` lines you find in modern code. They're harmless but no longer
needed, and `go fix` removes them for you.

## It depends on go.mod

The new behaviour is controlled by the `go` line in your module's `go.mod`, not by
which compiler you install. A module that says `go 1.21` keeps the **old** semantics
even when built with Go 1.27, so upgrading the compiler never silently changes an old
program. Bump the `go` line to get the new behaviour.

## What didn't change

Only variables **declared by the loop statement** are per-iteration. A variable
declared *outside* the loop is still one shared variable:

```go
current := ""
for _, name := range names {
	current = name
	handlers = append(handlers, func() { fmt.Println(current) }) // all share current
}
```

Every closure here prints the last name. If you want per-iteration values, use the
loop variable itself or declare a new variable *inside* the loop body.

## Further reading

- [The Go Blog: Fixing For Loops in Go 1.22](https://go.dev/blog/loopvar-preview)
