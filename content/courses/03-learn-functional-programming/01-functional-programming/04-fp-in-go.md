---
title: Functional Programming in Go
quiz:
  - question: Which of these does Go **not** have?
    options:
      - text: Functions that can be stored in variables
      - text: Closures
      - text: Built-in sum types with exhaustive pattern matching
        correct: true
      - text: Generics
    explanation: |
      Go has first-class functions, closures and generics. It has no built-in sum
      types or pattern matching. You can emulate them with interfaces and type
      switches, but the compiler won't check that you handled every case.
  - question: Why might a Go programmer choose a `for` loop over a deeply recursive function?
    options:
      - text: Go doesn't do tail-call optimisation, so each recursive call uses more stack
        correct: true
      - text: Go doesn't allow a function to call itself
      - text: Loops are declarative and recursion is imperative
      - text: Recursive functions can't return values in Go
    explanation: |
      Recursion works fine in Go, but the compiler never turns a tail call into a
      jump. Each call adds a stack frame, so very deep recursion costs memory and
      time that a loop would not.
---

Before diving in, let's be honest about what kind of FP language Go is. It's not
Haskell, and it doesn't try to be. Some functional ideas fit Go beautifully. Others
fight the language.

## Where Go helps

- **First-class functions.** Functions are values with types like
  `func(string) string`. You can store, pass and return them.
- **Closures.** Anonymous functions capture the variables around them.
- **Generics.** You can write one `Map` that works for every element type.
- **Iterators.** Since Go 1.23, `iter.Seq` and range-over-func let you build lazy
  pipelines.
- **Value semantics.** Structs and arrays are copied when assigned or passed, which
  gives you cheap immutability for small data.
- **Generic methods** (Go 1.27). A method can now declare its own type parameters,
  so a `List[T]` can have a `Map[U]` method that returns a `List[U]`.

Here's that last one in action:

```go
package main

import (
	"fmt"
	"strings"
)

type List[T any] []T

func (l List[T]) Map[U any](f func(T) U) List[U] {
	out := make(List[U], 0, len(l))
	for _, v := range l {
		out = append(out, f(v))
	}
	return out
}

func main() {
	words := List[string]{"go", "makes", "docs"}
	fmt.Println(words.Map(strings.ToUpper))
	fmt.Println(words.Map(func(s string) int { return len(s) }))
}
```

```text
[GO MAKES DOCS]
[2 5 4]
```

## Where Go pushes back

- **No immutability keyword.** `const` only works for numbers, strings and
  booleans. Slices and maps can always be changed by anyone holding them.
- **No tail-call optimisation.** Deep recursion grows the stack, so loops are often
  the better tool.
- **Verbose anonymous functions.** There's no short arrow syntax. You write
  `func(s string) int { return len(s) }` in full.
- **No sum types or pattern matching.** You'll emulate them with sealed interfaces.
- **Interface methods can't be generic.** A generic method can't satisfy an
  interface, so you can't write a `Mapper` interface with a generic `Map` method.
- **No standard `Map`/`Filter`/`Reduce` for slices.** You'll write your own (it
  takes a few lines) or use loops.

## The Go way

Go's culture values code that is obvious to the next reader. A five-stage pipeline of
nested anonymous functions can be harder to read than an eight-line loop. The goal
of this course isn't to make Go look like Haskell. It's to give you FP *ideas*
(pure functions, immutability, composition, laziness) that make your Go code clearer
and safer, and the judgement to know when to use them.
