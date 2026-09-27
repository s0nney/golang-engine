---
title: Imperative vs Declarative
quiz:
  - question: Which snippet is the most *declarative*?
    options:
      - text: 'A `for` loop with an index that sets a `found` flag and calls `break`'
      - text: 'A `goto` that jumps back to the top of a loop'
      - text: 'A `switch` on a loop counter'
      - text: '`slices.Contains(tags, "draft")`'
        correct: true
    explanation: |
      Declarative code says *what* you want ("does `tags` contain `"draft"`?") and
      leaves the *how* to someone else. The loop with a flag spells out every step,
      which makes it imperative.
  - question: |
      What does this print?

      ```go
      lines := []string{"# Intro", "hello", "## Setup", "run it"}
      n := 0
      for _, l := range lines {
          if strings.HasPrefix(l, "#") {
              n++
          }
      }
      fmt.Println(n)
      ```
    options:
      - text: '`2`'
        correct: true
      - text: '`1`'
      - text: '`4`'
      - text: '`0`'
    explanation: |
      Both `"# Intro"` and `"## Setup"` start with `#`, so the counter ends at 2.
      This is imperative code: you manage the counter yourself.
---

There are two broad ways to tell a computer what to do.

- **Imperative** code describes *how* to do something, step by step: "start a counter
  at zero, look at each line, if it starts with `#` add one, then return the counter."
- **Declarative** code describes *what* you want: "the number of lines that are
  headings."

Functional programming leans declarative. You describe the result, and small
reusable functions handle the steps.

## Doc2Doc: does a document have a heading?

Here's the imperative version. You manage a flag and a loop by hand:

```go
func hasHeading(lines []string) bool {
	found := false
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			found = true
			break
		}
	}
	return found
}
```

And here's a declarative version using the `slices` package:

```go
func hasHeading(lines []string) bool {
	return slices.ContainsFunc(lines, isHeading)
}

func isHeading(line string) bool {
	return strings.HasPrefix(line, "# ")
}
```

The second version reads almost like the English question. The loop still exists,
but it lives inside `slices.ContainsFunc`, written once and tested by the Go team.

## A full example

```go
package main

import (
	"fmt"
	"slices"
	"strings"
)

func isHeading(line string) bool {
	return strings.HasPrefix(line, "# ")
}

func main() {
	doc := "# Report\nSales went up.\nCosts went down."
	lines := strings.Split(doc, "\n")

	fmt.Println(slices.ContainsFunc(lines, isHeading))
	fmt.Println(slices.IndexFunc(lines, isHeading))
	fmt.Println(len(strings.Fields(doc)))
}
```

```text
true
0
8
```

Notice that `isHeading` is passed to `slices.ContainsFunc` *without* parentheses. You
aren't calling it, you're handing the function itself over. That's a first-class
function, and a whole chapter is coming up on it.

## You already know declarative code

SQL is declarative: `SELECT title FROM docs WHERE draft = false` never says how
to scan the table. HTML and CSS are declarative too. `strings.Fields(doc)` is
declarative: you ask for the words and don't care how it finds the spaces.

## Is imperative bad?

No! Go is a very imperative language, and a clear `for` loop is often the most
readable thing you can write. Declarative style shines when a well-named helper
(`slices.ContainsFunc`, `strings.Fields`, `slices.SortFunc`) already expresses your
intent. Reach for it when it makes the code *say what it means*, not just to look
clever.
