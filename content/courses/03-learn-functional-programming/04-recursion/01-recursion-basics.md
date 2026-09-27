---
title: Recursion Basics
quiz:
  - question: What is the *base case* of a recursive function?
    options:
      - text: The first line of the function
      - text: The case with the largest input
      - text: The call the function makes to itself
      - text: The case that returns an answer directly without calling the function again
        correct: true
    explanation: |
      The base case is where the recursion stops. Without one, the function
      calls itself forever, or at least until the stack runs out.
  - question: |
      What does this print?

      ```go
      func reverse(s string) string {
          if s == "" {
              return ""
          }
          return reverse(s[1:]) + s[:1]
      }

      func main() {
          fmt.Println(reverse("doc"))
      }
      ```
    options:
      - text: '`doc`'
      - text: '`cod`'
        correct: true
      - text: '`odc`'
      - text: It recurses forever
    explanation: |
      `reverse("doc")` is `reverse("oc") + "d"`, which is `(reverse("c") + "o") + "d"`,
      which is `((reverse("") + "c") + "o") + "d"`, giving `"cod"`. The empty string
      is the base case. (Slicing a string works on bytes, so this simple version is
      only correct for ASCII text.)
---

A **recursive** function is one that calls itself. It sounds like a trick, but it's
a natural way to solve problems that contain smaller copies of themselves: a folder
contains folders, a document contains sections that contain sections, and a list is
"the first item, plus the rest of the list".

## Two ingredients

Every recursive function needs:

1. A **base case**: an input small enough to answer directly, with no recursion.
2. A **recursive case**: the function calls itself on a *smaller* input and uses
   that answer to build its own.

Each recursive call must move closer to the base case, or it never ends.

## Doc2Doc: counting words across documents

Doc2Doc has a slice of document bodies and needs the total word count. Think of it
recursively: the total for a list is the words in the first document plus the total
for the rest. The total for an empty list is zero.

```go
package main

import (
	"fmt"
	"strings"
)

func totalWords(docs []string) int {
	if len(docs) == 0 { // base case
		return 0
	}
	first := len(strings.Fields(docs[0]))
	return first + totalWords(docs[1:]) // recursive case: a shorter slice
}

func main() {
	docs := []string{
		"Doc2Doc converts files",
		"It also formats them",
		"and counts words",
	}
	fmt.Println(totalWords(docs))
}
```

```text
10
```

Let's trace it:

```text
totalWords([3 docs]) = 3 + totalWords([2 docs])
totalWords([2 docs]) = 4 + totalWords([1 doc])
totalWords([1 doc])  = 3 + totalWords([])
totalWords([])       = 0                       <- base case
```

Then the answers flow back up: `0`, `3`, `7`, `10`.

`docs[1:]` doesn't copy anything. It's a new slice header pointing into the same
array, so each call is cheap.

## Thinking recursively

The trick is to *trust the recursion*. When writing `totalWords`, don't try to
picture every call. Assume `totalWords(docs[1:])` already works for the smaller slice,
and ask one question: "How do I use that answer to handle one more document?"

## Honest advice for Go

For a flat list like this, a `for` loop is simpler and faster, and most Go
programmers would write one:

```go
total := 0
for _, d := range docs {
	total += len(strings.Fields(d))
}
```

Recursion earns its keep with data that is *itself* recursive, such as trees,
nested sections and folders inside folders. You'll get to those in a couple of
lessons. First, let's look at what actually happens in memory when a function calls
itself.

## Further reading

- [Go by Example: Recursion](https://gobyexample.com/recursion)
