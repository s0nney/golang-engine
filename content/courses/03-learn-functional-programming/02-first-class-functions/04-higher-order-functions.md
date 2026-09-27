---
title: Higher-Order Functions
quiz:
  - question: What makes a function *higher-order*?
    options:
      - text: It's declared at the top level of a package
      - text: It takes a function as an argument, returns a function, or both
        correct: true
      - text: It uses generics
      - text: It calls itself
    explanation: |
      A higher-order function works with other functions as data. `slices.SortFunc`
      takes one, and a function that builds and returns a formatter returns one.
  - question: |
      What does this print?

      ```go
      func applyToLines(doc string, f func(string) string) string {
          var out []string
          for line := range strings.Lines(doc) {
              out = append(out, f(strings.TrimSuffix(line, "\n")))
          }
          return strings.Join(out, "|")
      }

      func main() {
          fmt.Println(applyToLines("a\nb\n", strings.ToUpper))
      }
      ```
    options:
      - text: '`A|B`'
        correct: true
      - text: '`A|B|`'
      - text: '`a|b`'
      - text: '`A\nB\n`'
    explanation: |
      `strings.Lines` yields `"a\n"` and `"b\n"`. There's no empty third line,
      because the input ends right after the last newline. Each line is trimmed,
      upper-cased and joined with `|`.
---

A **higher-order function** is a function that takes another function as a
parameter, returns a function, or both. You've already used a few:
`slices.ContainsFunc`, `slices.IndexFunc` and `slices.SortFunc` all take a function
that tells them what to do.

Higher-order functions let you separate the *shape* of an algorithm ("go through
every line") from the *details* ("upper-case it", "add a line number").

## Doc2Doc: a line processor

Doc2Doc applies lots of per-line operations: indenting, numbering, stripping
trailing spaces. Without higher-order functions you'd write the same loop again and
again. With one, you write the loop once:

```go
package main

import (
	"fmt"
	"strings"
)

func mapLines(doc string, f func(string) string) string {
	lines := strings.Split(doc, "\n")
	for i, line := range lines {
		lines[i] = f(line)
	}
	return strings.Join(lines, "\n")
}

func indent(line string) string {
	return "    " + line
}

func quote(line string) string {
	return "> " + line
}

func main() {
	doc := "Roses are red\nGo is too"
	fmt.Println(mapLines(doc, indent))
	fmt.Println(mapLines(doc, quote))
	fmt.Println(mapLines(doc, strings.ToUpper))
}
```

```text
    Roses are red
    Go is too
> Roses are red
> Go is too
ROSES ARE RED
GO IS TOO
```

`mapLines` doesn't know or care what happens to each line. The caller decides by
passing in a function.

## Functions that return functions

The other half of "higher-order" is returning a function. Say you want indents of
different widths. Instead of writing `indent2`, `indent4` and `indent8`, write a
function that *builds* indenters:

```go
func indentBy(n int) func(string) string {
	pad := strings.Repeat(" ", n)
	return func(line string) string {
		return pad + line
	}
}
```

Now `mapLines(doc, indentBy(2))` indents by two spaces. The returned function
remembers `pad` even after `indentBy` has returned. That's a *closure*, and you'll
spend a whole chapter on those.

## Where you'll see them in Go

- `slices.SortFunc`, `slices.IndexFunc`, `slices.DeleteFunc`
- `strings.Map` (applies a `func(rune) rune` to every character) and
  `strings.FieldsFunc`
- `sort.Search`, `sync.OnceValue`, `http.HandlerFunc`
- `time.AfterFunc`, `filepath.WalkDir`

Once you notice the pattern, you'll see it everywhere. Any time part of an
algorithm varies, pass that part in as a function.
