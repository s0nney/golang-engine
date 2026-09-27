---
title: Side Effects
quiz:
  - question: Which of these is **not** a side effect?
    options:
      - text: Returning a newly built string
        correct: true
      - text: Writing a converted file to disk
      - text: Appending to a slice owned by the caller
      - text: Printing a progress message
    explanation: |
      Returning a value is the whole point of a function, not a side effect. The
      others change the world outside the function: the disk, the caller's data or
      the terminal.
  - question: |
      What does this print?

      ```go
      func normalize(lines []string) []string {
          for i := range lines {
              lines[i] = strings.TrimSpace(lines[i])
          }
          return lines
      }

      func main() {
          doc := []string{" a ", " b "}
          clean := normalize(doc)
          fmt.Printf("%q %q\n", doc[0], clean[0])
      }
      ```
    options:
      - text: '`" a " "a"`'
      - text: '`" a " " a "`'
      - text: '`"a" "a"`'
        correct: true
    explanation: |
      `normalize` writes into the slice it was given, and `doc` and `clean` share the
      same backing array. Modifying an argument like this is a side effect: the
      caller's `doc` changed too.
---

A **side effect** is anything a function does besides computing its return value.
Common side effects are:

- printing to the terminal or writing logs
- reading or writing files, databases or the network
- changing a package-level variable
- modifying data the caller passed in (a slice, a map, or something behind a pointer)
- reading the clock or generating random numbers (strictly, those are
  *non-deterministic* rather than side effects, but they break purity the same way)

Side effects aren't evil. They're how programs do anything useful. The trouble is
*hidden* side effects: a function whose name says "format" that also quietly
changes your data.

## A sneaky side effect

Here's a Doc2Doc helper that strips trailing whitespace. Spot the problem:

```go
package main

import (
	"fmt"
	"strings"
)

func stripTrailing(lines []string) []string {
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	return lines
}

func main() {
	original := []string{"# Title   ", "body\t"}
	cleaned := stripTrailing(original)

	fmt.Printf("%q\n", cleaned)
	fmt.Printf("%q\n", original) // oops
}
```

```text
["# Title" "body"]
["# Title" "body"]
```

The function *looks* like it returns a new slice, but it rewrites the caller's
slice in place. If Doc2Doc wanted to show a diff between the original and the cleaned
version, there's no original left!

The pure version builds a new slice and leaves its input alone:

```go
func stripTrailing(lines []string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = strings.TrimRight(l, " \t")
	}
	return out
}
```

## Functional core, imperative shell

A popular way to organise programs is to split them in two:

- The **functional core** is pure functions that hold all the logic.
- The **imperative shell** is a thin layer that does the I/O and calls the core.

```go
// Imperative shell: small, does the I/O.
func run(inPath, outPath string) error {
	data, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}
	out := convert(string(data)) // pure core
	return os.WriteFile(outPath, []byte(out), 0o644)
}

// Functional core: all the logic, no I/O.
func convert(doc string) string {
	// trim, wrap, number headings...
	return doc
}
```

`convert` can be tested with plain strings. `run` is so simple that there's barely
anything to test.

## When side effects are fine

Go is pragmatic, and so should you be. Mutating a slice or map *that your function
created itself* is invisible to the caller, so the function is still pure from the
outside. A `strings.Builder` that you fill and then return as a string is a perfectly
pure use of mutation. The rule of thumb is simple: **don't change what you don't
own**, and make any effect obvious from the function's name.
