---
title: Sealed Interfaces
quiz:
  - question: Why does the `Block` interface in this lesson have an unexported method `isBlock()`?
    options:
      - text: To make blocks faster
      - text: Because interfaces must have at least one unexported method
      - text: So that `isBlock` can be called from other packages
      - text: So that only types in the same package can implement it, which closes the set of alternatives
        correct: true
    explanation: |
      Other packages can't declare a method named `isBlock`, because unexported
      names belong to their package. So the list of `Block` types is fixed to the
      ones you wrote. That's what "sealed" means.
  - question: |
      Which of these is a `Block` in the lesson's code?
    options:
      - text: '`"# Title"`'
      - text: '`struct{ Text string }{"Title"}`'
      - text: '`Heading{Level: 1, Text: "Title"}`'
        correct: true
      - text: '`nil` is the only `Block`'
    explanation: |
      `Heading` has an `isBlock()` method, so it satisfies `Block`. A string or an
      anonymous struct has no such method. (A nil `Block` variable is possible, but
      it holds no block at all, which the next lessons deal with.)
---

The idiomatic way to emulate a sum type in Go is a **sealed interface**: an interface
that only a fixed set of types, all in your package, can implement.

## The recipe

1. Declare an interface with an **unexported marker method**.
2. Declare one struct per alternative, each with its own fields.
3. Give each struct the marker method.

```go
package doc

// Block is one of: Heading, Paragraph, CodeBlock.
type Block interface {
	isBlock()
}

type Heading struct {
	Level int
	Text  string
}

type Paragraph struct {
	Text string
}

type CodeBlock struct {
	Lang string
	Code string
}

func (Heading) isBlock()   {}
func (Paragraph) isBlock() {}
func (CodeBlock) isBlock() {}
```

The marker method does nothing at all. Its only job is to say "I'm a Block." Because
`isBlock` is unexported, code in other packages *can't* write a method with that name,
so nobody outside package `doc` can add a fourth kind of block. The set is closed,
or **sealed**.

Now every alternative has exactly the fields that make sense for it. There's no
`Lang` on a `Paragraph` and no `Level` on a `CodeBlock`. Illegal states are much
harder to build.

## A parser that returns blocks

Here's a tiny Doc2Doc parser, all in one file so you can run it:

```go
package main

import (
	"fmt"
	"strings"
)

type Block interface {
	isBlock()
}

type Heading struct {
	Level int
	Text  string
}

type Paragraph struct {
	Text string
}

type CodeBlock struct {
	Lang string
	Code string
}

func (Heading) isBlock()   {}
func (Paragraph) isBlock() {}
func (CodeBlock) isBlock() {}

func parseBlock(chunk string) Block {
	if rest, ok := strings.CutPrefix(chunk, "```"); ok {
		lang, code, _ := strings.Cut(rest, "\n")
		return CodeBlock{Lang: lang, Code: strings.TrimSuffix(code, "```")}
	}
	if strings.HasPrefix(chunk, "#") {
		level := len(chunk) - len(strings.TrimLeft(chunk, "#"))
		return Heading{Level: level, Text: strings.TrimSpace(chunk[level:])}
	}
	return Paragraph{Text: chunk}
}

func main() {
	doc := "## Setup\n\nInstall Doc2Doc first.\n\n```sh\ngo install doc2doc\n```"
	for chunk := range strings.SplitSeq(doc, "\n\n") {
		fmt.Printf("%#v\n", parseBlock(chunk))
	}
}
```

```text
main.Heading{Level:2, Text:"Setup"}
main.Paragraph{Text:"Install Doc2Doc first."}
main.CodeBlock{Lang:"sh", Code:"go install doc2doc\n"}
```

`parseBlock` returns a `Block`, and each value inside it is exactly one of the three
alternatives. That's our sum type.

## Document the alternatives

Go can't list the members of a sealed interface in the type itself, so **write them in
the doc comment** (`// Block is one of: Heading, Paragraph, CodeBlock.`). Readers,
and linters, depend on it.

## A crack in the seal

The seal isn't airtight. Another package can **embed** one of your types in its own
struct, and the embedded `isBlock` method gets promoted:

```go
type FancyHeading struct {
	doc.Heading // promotes isBlock, so FancyHeading is a doc.Block!
	Color string
}
```

In practice this is rare and clearly deliberate. Treat the sealed interface as a strong
convention, not a security boundary.

## Why not put real methods on the interface?

You *could* give `Block` a `Render() string` method and have each type implement it.
That's the OOP approach, and it's great when the set of *operations* is fixed but new
*types* are added often. Sum types are the opposite: the set of *types* is fixed and
you add new *operations* (render to HTML, render to Markdown, count words, build a
table of contents) freely, each one as a function with a type switch. That's the
next lesson.
