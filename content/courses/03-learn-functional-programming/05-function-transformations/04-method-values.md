---
title: Method Values and Expressions
quiz:
  - question: |
      What does this print?

      ```go
      type Doc struct{ Title string }

      func (d Doc) Upper() string { return strings.ToUpper(d.Title) }

      func main() {
          d := Doc{Title: "draft"}
          f := d.Upper
          d.Title = "final"
          fmt.Println(f())
      }
      ```
    options:
      - text: '`FINAL`'
      - text: '`draft`'
      - text: It doesn't compile
      - text: '`DRAFT`'
        correct: true
    explanation: |
      `d.Upper` is a method value. With a value receiver, Go copies `d` at the moment
      the method value is created. Changing `d.Title` afterwards doesn't affect the
      copy that `f` holds.
  - question: What is the type of the method expression `Doc.Upper`?
    options:
      - text: '`func() string`'
      - text: '`func(*Doc) string`'
      - text: '`func(Doc) string`'
        correct: true
      - text: '`Doc`'
    explanation: |
      A method expression turns the receiver into the first ordinary parameter.
      `Doc.Upper` is a plain function that takes a `Doc` and returns a `string`.
---

Methods are functions too, and Go gives you two ways to turn a method into a function
value you can pass around.

## Method values: bind the receiver

Writing `x.Method` *without* calling it gives you a **method value**, a function that
has `x` already baked in:

```go
package main

import (
	"fmt"
	"strings"
)

func mapLines(doc string, f func(string) string) string {
	lines := strings.Split(doc, "\n")
	for i, l := range lines {
		lines[i] = f(l)
	}
	return strings.Join(lines, "\n")
}

func main() {
	smartQuotes := strings.NewReplacer(`"`, "”", "'", "’", "--", "—")
	doc := "It's \"done\"\nWait -- really?"

	fmt.Println(mapLines(doc, smartQuotes.Replace))
}
```

```text
It’s ”done”
Wait — really?
```

`smartQuotes.Replace` has type `func(string) string`, exactly what `mapLines`
wants. The `*strings.Replacer` rides along inside the function value. No wrapper
function needed.

(Yes, a real smart-quote converter would pick opening and closing quotes. Doc2Doc
v2 problems!)

## When the receiver is copied

A method value captures its receiver **when it's created**:

- With a **value receiver**, `d.Method` copies `d` right then. Later changes to `d`
  aren't seen.
- With a **pointer receiver**, it captures the pointer, so later changes to the
  pointed-to value *are* seen.

This is the same rule as for `defer d.Method()` and `go d.Method()`, and it causes
the same surprises.

## Method expressions: the receiver becomes a parameter

Writing `Type.Method` gives you a **method expression**, a function whose first
parameter is the receiver:

```go
type Doc struct {
	Title string
	Body  string
}

func (d Doc) Words() int { return len(strings.Fields(d.Body)) }

// Doc.Words has type func(Doc) int
```

That's perfect for generic helpers like `Map`:

```go
docs := []Doc{
	{Title: "a", Body: "one two"},
	{Title: "b", Body: "three"},
}
counts := Map(docs, Doc.Words) // [2 1]
```

For a pointer receiver, write `(*Doc).Method`, which has type `func(*Doc) ...`.

## Why this counts as a transformation

Method values and expressions let you move smoothly between the OOP and FP halves of
Go. A method written for objects becomes a plain function that plugs into `Map`,
`slices.SortFunc`, `Compose` or `http.HandleFunc`. You didn't write any glue code;
the language transformed the method for you.
