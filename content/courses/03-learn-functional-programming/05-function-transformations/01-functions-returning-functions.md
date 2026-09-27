---
title: Functions That Return Functions
quiz:
  - question: |
      What does this print?

      ```go
      func surround(left, right string) func(string) string {
          return func(s string) string {
              return left + s + right
          }
      }

      func main() {
          bold := surround("**", "**")
          code := surround("`", "`")
          fmt.Println(bold(code("go")))
      }
      ```
    options:
      - text: '``**`go`**``'
        correct: true
      - text: '`**go**`'
      - text: '`` `**go**` ``'
      - text: It doesn't compile
    explanation: |
      `code("go")` runs first and returns `` `go` ``. Then `bold` wraps that in
      `**`, giving ``**`go`**``. Each returned function remembers its own `left`
      and `right`.
  - question: What is the type of `surround("<b>", "</b>")`?
    options:
      - text: '`string`'
      - text: '`func(string) string`'
        correct: true
      - text: '`func(string, string) string`'
      - text: '`func() string`'
    explanation: |
      `surround` itself has type `func(string, string) func(string) string`. Calling
      it returns its result type, a `func(string) string`.
---

You've passed functions *into* functions. Now let's go the other way: functions that
**build and return** new functions. This is sometimes called a *function factory*.

## Why build functions?

Doc2Doc supports many output formats, and each one wraps text differently: Markdown
bold uses `**`, HTML uses `<b>` and `</b>`, and so on. You could write a separate
function for each:

```go
func mdBold(s string) string   { return "**" + s + "**" }
func htmlBold(s string) string { return "<b>" + s + "</b>" }
func mdCode(s string) string   { return "`" + s + "`" }
// ...and twenty more
```

They're all the same shape with different strings. Instead, write one function that
*makes* them:

```go
package main

import "fmt"

func surround(left, right string) func(string) string {
	return func(s string) string {
		return left + s + right
	}
}

func main() {
	mdBold := surround("**", "**")
	htmlBold := surround("<b>", "</b>")
	quoted := surround(`"`, `"`)

	fmt.Println(mdBold("important"))
	fmt.Println(htmlBold("important"))
	fmt.Println(quoted("important"))
}
```

```text
**important**
<b>important</b>
"important"
```

The return type `func(string) string` tells you that `surround` hands back a
function. Each one it returns remembers the `left` and `right` it was built with.
That memory is a closure, which gets its own chapter soon.

## Choosing a function at runtime

Factories also pick behaviour based on input. Doc2Doc chooses a line-ending
converter based on a flag:

```go
func lineEndings(style string) (func(string) string, error) {
	switch style {
	case "unix":
		return func(s string) string {
			return strings.ReplaceAll(s, "\r\n", "\n")
		}, nil
	case "windows":
		return func(s string) string {
			s = strings.ReplaceAll(s, "\r\n", "\n")
			return strings.ReplaceAll(s, "\n", "\r\n")
		}, nil
	default:
		return nil, fmt.Errorf("unknown line ending style %q", style)
	}
}
```

The caller asks for a converter once, checks the error once, then applies the
function to every file. The decision is made in one place instead of inside the
file loop.

## Reading the types

Types of higher-order functions can look scary. Read them left to right:

```go
func(string, string) func(string) string
```

"A function taking two strings and returning (a function taking a string and
returning a string)." If it gets confusing, name the inner type:

```go
type Transform func(string) string

func surround(left, right string) Transform
```

Much easier on the eyes.
