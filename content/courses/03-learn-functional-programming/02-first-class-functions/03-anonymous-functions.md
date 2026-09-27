---
title: Anonymous Functions
quiz:
  - question: |
      What does this print?

      ```go
      result := func(a, b int) int {
          return a * b
      }(3, 4)
      fmt.Println(result)
      ```
    options:
      - text: '`7`'
      - text: 'A function value such as `0x47b2e0`'
      - text: It doesn't compile
      - text: '`12`'
        correct: true
    explanation: |
      The `(3, 4)` right after the closing brace calls the anonymous function
      immediately, so `result` is the returned `int`, 12.
  - question: Which call sorts `docs` (a `[]Doc`) by `Words`, smallest first?
    options:
      - text: '`slices.SortFunc(docs, func(a, b Doc) int { return cmp.Compare(a.Words, b.Words) })`'
        correct: true
      - text: '`slices.SortFunc(docs, func(a, b Doc) bool { return a.Words < b.Words })`'
      - text: '`slices.Sort(docs, Words)`'
      - text: '`slices.SortFunc(docs, func(d Doc) int { return d.Words })`'
    explanation: |
      `slices.SortFunc` wants a comparison function returning a negative number,
      zero or a positive number. `cmp.Compare` returns exactly that.
---

Not every function deserves a name. When you need a small function exactly once,
usually to pass it to another function, write it inline as an **anonymous function**
(also called a *function literal*):

```go
func(s string) bool {
	return strings.HasPrefix(s, "#")
}
```

It looks like a normal function declaration with the name removed. You can store it
in a variable, pass it straight to another function or call it immediately.

## Passing one inline

This is the most common use by far. Doc2Doc needs to find the first empty line in a
document (where the front matter ends):

```go
package main

import (
	"cmp"
	"fmt"
	"slices"
)

type Doc struct {
	Name  string
	Words int
}

func main() {
	lines := []string{"title: Report", "author: Sam", "", "Body text"}
	i := slices.IndexFunc(lines, func(l string) bool { return l == "" })
	fmt.Println("front matter ends at line", i)

	docs := []Doc{{"notes.md", 420}, {"spec.md", 1800}, {"todo.md", 35}}
	slices.SortFunc(docs, func(a, b Doc) int {
		return cmp.Compare(a.Words, b.Words)
	})
	fmt.Println(docs)
}
```

```text
front matter ends at line 2
[{todo.md 35} {notes.md 420} {spec.md 1800}]
```

The comparison logic sits right where it's used, so the reader doesn't have to scroll
off to find a `byWords` helper.

## Calling one immediately

Add parentheses after the closing brace and the function runs on the spot:

```go
func() {
	fmt.Println("Doc2Doc starting up")
}()
```

On its own this isn't very useful, but you'll see the pattern with `defer` and
`go` statements all the time:

```go
defer func() {
	fmt.Println("finished converting")
}()
```

## Storing one in a variable

```go
isBlank := func(s string) bool { return strings.TrimSpace(s) == "" }
fmt.Println(isBlank("   ")) // true
```

A variable holding an anonymous function behaves just like a named function, with
one difference: it can't call *itself* by name unless you declare the variable
first (`var fact func(int) int` and then assign it). That's rarely a problem.

## Keep them short

Go has no short "arrow" syntax, so anonymous functions are wordy. A good rule is
that if an anonymous function grows past a handful of lines, or you need it twice,
give it a name. `slices.IndexFunc(lines, isFrontMatterEnd)` often reads better than
five lines of inline logic.
