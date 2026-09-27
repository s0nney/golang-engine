---
title: Type Switches
quiz:
  - question: |
      What does this print?

      ```go
      func describe(b Block) string {
          switch b := b.(type) {
          case Heading:
              return fmt.Sprintf("h%d", b.Level)
          case Paragraph:
              return "p"
          default:
              return "?"
          }
      }

      func main() {
          fmt.Println(describe(Heading{Level: 3}), describe(CodeBlock{}))
      }
      ```
    options:
      - text: '`h3 p`'
      - text: It doesn't compile, because `CodeBlock` isn't handled
      - text: '`h0 ?`'
      - text: '`h3 ?`'
        correct: true
    explanation: |
      The `Heading` case matches and uses `b.Level`. `CodeBlock` has no case, so it
      falls through to `default`. Go doesn't force you to handle every case, which
      is exactly why a `default` matters.
  - question: 'Inside `case Heading:` in `switch b := b.(type)`, what is the type of `b`?'
    options:
      - text: '`Block`'
      - text: '`any`'
      - text: '`Heading`'
        correct: true
      - text: '`*Heading`'
    explanation: |
      In a single-type case, the switch variable has that concrete type, so you can
      read `b.Level` directly. In a `default` case, or a case listing several
      types, it keeps the interface type.
---

A sealed interface gives you the "one of" data. A **type switch** is how you work
with it: look at which alternative you have and handle each one. It's Go's closest
thing to pattern matching.

## Rendering blocks

Here's Doc2Doc turning blocks into HTML and back into plain text, two separate
operations over the same sum type:

```go
package main

import (
	"fmt"
	"html"
	"strings"
)

// Block is one of: Heading, Paragraph, CodeBlock.
type Block interface{ isBlock() }

type Heading struct {
	Level int
	Text  string
}
type Paragraph struct{ Text string }
type CodeBlock struct{ Lang, Code string }

func (Heading) isBlock()   {}
func (Paragraph) isBlock() {}
func (CodeBlock) isBlock() {}

func toHTML(b Block) string {
	switch b := b.(type) {
	case Heading:
		return fmt.Sprintf("<h%d>%s</h%d>", b.Level, html.EscapeString(b.Text), b.Level)
	case Paragraph:
		return "<p>" + html.EscapeString(b.Text) + "</p>"
	case CodeBlock:
		return fmt.Sprintf(`<pre><code class="language-%s">%s</code></pre>`,
			b.Lang, html.EscapeString(b.Code))
	default:
		panic(fmt.Sprintf("toHTML: unknown block %T", b))
	}
}

func toText(b Block) string {
	switch b := b.(type) {
	case Heading:
		return strings.ToUpper(b.Text)
	case Paragraph:
		return b.Text
	case CodeBlock:
		return "    " + strings.ReplaceAll(strings.TrimSpace(b.Code), "\n", "\n    ")
	default:
		panic(fmt.Sprintf("toText: unknown block %T", b))
	}
}

func main() {
	doc := []Block{
		Heading{Level: 1, Text: "Doc2Doc & you"},
		Paragraph{Text: "Converts <anything>."},
		CodeBlock{Lang: "go", Code: "fmt.Println(1 < 2)\n"},
	}
	for _, b := range doc {
		fmt.Println(toHTML(b))
	}
	fmt.Println()
	for _, b := range doc {
		fmt.Println(toText(b))
	}
}
```

```text
<h1>Doc2Doc &amp; you</h1>
<p>Converts &lt;anything&gt;.</p>
<pre><code class="language-go">fmt.Println(1 &lt; 2)
</code></pre>

DOC2DOC & YOU
Converts <anything>.
    fmt.Println(1 < 2)
```

## How the switch works

- `switch b := b.(type)` checks the dynamic type of the value inside the interface.
- In each single-type `case`, the new `b` **has that concrete type**, so
  `b.Level` is available in the `Heading` case and `b.Lang` in the `CodeBlock` case.
- Cases can list several types (`case Heading, Paragraph:`), but then `b` stays a
  `Block`, since Go can't know which one it is.

## Always add a default

Go **won't** tell you if you forget a case. If someone adds an `Image` block to the
package and forgets to update `toText`, the switch silently matches nothing. Without
a `default`, `toText` would fall off the end of the switch... except that a function
with a result must return, so the compiler would make you write *something* after the
switch anyway.

A `default` that **panics** with the unexpected type (`%T` prints it) turns a silent
bug into a loud one the first time a test hits it. Some teams return an error instead
of panicking. Either way, don't just ignore unknown cases.

## Pure functions over data

Notice that `toHTML` and `toText` are pure functions, and adding a third operation
(`wordCount`, `toMarkdown`, `tableOfContents`) never touches the block types. You add
one new function with one new type switch. This is the functional way of organising
code: **data** is a closed set of shapes, and **behaviour** lives in functions that
take that data apart.

The OOP way is the mirror image: behaviour lives in methods, so adding an operation
means touching every type, but adding a type touches nothing else. Neither is always
better. Pick the one that matches what changes more often in your program.
