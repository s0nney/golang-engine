---
title: The Limits of Sum Types in Go
quiz:
  - question: |
      What does this print?

      ```go
      func kind(b Block) string {
          switch b.(type) {
          case Heading:
              return "heading"
          case Paragraph:
              return "paragraph"
          default:
              return "other"
          }
      }

      func main() {
          var b Block
          fmt.Println(kind(b))
      }
      ```
    options:
      - text: '`heading`'
      - text: '`other`'
        correct: true
      - text: It panics with a nil pointer dereference
      - text: It doesn't compile
    explanation: |
      A nil interface holds no type at all, so it matches none of the concrete cases
      and lands in `default`. You can catch it explicitly with `case nil:`. Every
      sealed interface quietly has this extra "nothing" case.
  - question: What happens at compile time if you add an `Image` block type but forget to handle it in an existing type switch?
    options:
      - text: The compiler reports a non-exhaustive switch
      - text: '`go vet` always reports it'
      - text: Nothing. The code compiles and `Image` values fall into `default` (or match no case) at runtime
        correct: true
    explanation: |
      Go has no exhaustiveness checking for type switches. Only a third-party linter
      or a test that covers every block type will catch the gap.
---

Sealed interfaces and type switches get you most of the way to sum types. Here's
the honest list of where they fall short, and what Go programmers do about it.

## 1. No exhaustiveness checking

This is the big one. In Rust or Haskell, adding a new alternative makes every
incomplete `match` a **compile error**. In Go, add `Image` to Doc2Doc's blocks and
every type switch still compiles. The missing case only shows up at runtime,
usually as the `default` branch's panic.

Mitigations:

- Always write a `default` that panics or returns an error naming the type (`%T`).
- Write a test that runs every operation on one value of *every* block type. When
  someone adds a type, they add it to the test's list, and the panics show up.
- Use a third-party linter. For example, `go-sumtype` checks type switches over
  interfaces you mark with a special comment. None of these ship with Go.

## 2. The hidden nil case

Every interface type has a zero value, `nil`, which holds no alternative at all. So a
"sum" of three types really has four states. A nil `Block` falls through to
`default` in a type switch, and calling a method on it panics. Handle it with
`case nil:` where it can happen, and don't return nil blocks from your parser.

## 3. The seal can be broken

As you saw, another package can embed one of your types in its own struct and inherit
the marker method. It's rare, but it means the set of types isn't *guaranteed* closed.

## 4. Allocation and indirection

Storing a struct in an interface usually means allocating it on the heap and
following a pointer to read it. For a slice of a million blocks, that's slower than a
slice of plain structs. Usually it doesn't matter, but in hot loops it can.

## 5. No compact syntax

Three alternatives cost you an interface, three structs and three marker methods,
where Rust needs four lines. It's boilerplate, but it's simple boilerplate.

## Sum types you already use: errors

Go's `error` is a kind of open sum type, and you inspect it the same way. Since Go
1.26, `errors.AsType` gives you a generic, type-safe way to check for a specific
alternative, even when it's wrapped:

```go
package main

import (
	"errors"
	"fmt"
	"io/fs"
)

type ParseError struct {
	Line int
	Msg  string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Line, e.Msg)
}

func convert(name string) error {
	switch name {
	case "broken.md":
		return fmt.Errorf("convert %s: %w", name, &ParseError{Line: 7, Msg: "unclosed code block"})
	case "missing.md":
		return fmt.Errorf("convert %s: %w", name, fs.ErrNotExist)
	}
	return nil
}

func main() {
	for _, name := range []string{"ok.md", "broken.md", "missing.md"} {
		err := convert(name)
		if pe, ok := errors.AsType[*ParseError](err); ok {
			fmt.Printf("%s: fix line %d\n", name, pe.Line)
		} else if errors.Is(err, fs.ErrNotExist) {
			fmt.Printf("%s: no such file\n", name)
		} else if err == nil {
			fmt.Printf("%s: converted\n", name)
		}
	}
}
```

```text
ok.md: converted
broken.md: fix line 7
missing.md: no such file
```

## So should you use them?

Yes, when the data really is "one of a fixed set": parsed document blocks, tokens in
a lexer, commands in a CLI, events in a system. A sealed interface documents that
intent clearly, and type switches keep each operation in one readable function.

Just remember what Go does and doesn't check for you, and back it up with a `default`
branch and tests.

## Course wrap-up

You've now used every major functional idea in Go: first-class and higher-order
functions, pure functions and immutability, recursion, composition, closures,
currying, decorators, lazy iterators and sum types. More importantly, you've seen
where each one fits Go and where a plain loop or a struct is the better choice.
Doc2Doc is ready to ship. Go build something!
