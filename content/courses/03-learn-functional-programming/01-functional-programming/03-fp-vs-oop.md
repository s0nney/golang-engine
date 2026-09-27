---
title: FP vs OOP
quiz:
  - question: |
      What does this print?

      ```go
      type Doc struct{ Title string }

      func Renamed(d Doc, title string) Doc {
          d.Title = title
          return d
      }

      func main() {
          a := Doc{Title: "draft"}
          b := Renamed(a, "final")
          fmt.Println(a.Title, b.Title)
      }
      ```
    options:
      - text: '`draft final`'
        correct: true
      - text: '`final final`'
      - text: '`draft draft`'
      - text: It doesn't compile, because you can't assign to a parameter
    explanation: |
      Structs are passed by value, so `Renamed` gets its own copy of `a`. It changes
      the copy and returns it. The original `a` still has the title `"draft"`.
  - question: Which statement about FP and OOP in Go is most accurate?
    options:
      - text: Go code usually mixes both, using types with methods and plain functions wherever each one reads best
        correct: true
      - text: You must pick one style per program
      - text: Go forbids methods in functional code
      - text: FP is always faster than OOP
    explanation: |
      FP and OOP are tools, not religions. Idiomatic Go mixes structs, methods,
      interfaces and function values freely.
---

You've just finished a course on object-oriented programming in Go, so let's compare
the two styles directly.

## Two ways to rename a document

**Object-oriented style** bundles data with the methods that change it. The object
owns its state, and methods *mutate* that state:

```go
type Doc struct {
	Title string
	Lines []string
}

func (d *Doc) Rename(title string) {
	d.Title = title
}
```

**Functional style** keeps data and functions separate. Functions don't change
their input. They return a *new* value instead:

```go
func Renamed(d Doc, title string) Doc {
	d.Title = title // d is a copy, so the caller's Doc is untouched
	return d
}
```

Let's see the difference in a program:

```go
package main

import "fmt"

type Doc struct {
	Title string
}

func (d *Doc) Rename(title string) {
	d.Title = title
}

func Renamed(d Doc, title string) Doc {
	d.Title = title
	return d
}

func main() {
	a := Doc{Title: "draft"}
	a.Rename("v1")
	fmt.Println(a.Title)

	b := Renamed(a, "v2")
	fmt.Println(a.Title, b.Title)
}
```

```text
v1
v1 v2
```

The method changed `a` in place. The function left `a` alone and gave you `b`.

## The trade-offs

| | OOP | FP |
|---|---|---|
| Focus | Objects that own state | Data flowing through functions |
| Changing data | Mutate in place | Return a new value |
| Reuse through | Interfaces, embedding | Passing and composing functions |
| Risk | Hidden changes to shared state | Extra copying |

Mutating in place is cheap and sometimes exactly what you want, like a buffer that
grows as Doc2Doc reads a file. Returning new values is safer, because nobody else's
copy changes under their feet. That matters a lot once goroutines are involved.

## A Go-specific catch

Copying a struct copies its fields, but a slice field is only a small header
pointing at a shared backing array. If `Doc` had a `Lines []string` field, the copy
inside `Renamed` would share the same lines as the original. Changing `d.Lines[0]`
would change the caller's document too. You'll learn how to deal with that in the
Pure Functions chapter.

## Not a war

In Go you don't have to choose sides. A typical Doc2Doc might have a `Document` type
with a few methods, a `Formatter` interface, and a pile of small functions like
`Trim`, `Wrap` and `CountWords` that transform text. Use methods when a type
naturally owns its behaviour, and plain functions when you're transforming data.
