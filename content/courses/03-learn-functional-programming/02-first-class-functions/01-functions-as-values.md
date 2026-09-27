---
title: Functions as Values
quiz:
  - question: |
      What does this print?

      ```go
      f := strings.ToUpper
      g := f
      fmt.Println(g("md"))
      ```
    options:
      - text: '`md`'
      - text: '`MD`'
        correct: true
      - text: 'A memory address such as `0x4a1b20`'
      - text: It doesn't compile
    explanation: |
      `strings.ToUpper` without parentheses is the function itself. Assigning it to
      `f` and then `g` just copies the function value. Calling `g("md")` runs
      `strings.ToUpper`, giving `MD`.
  - question: What is the zero value of a variable declared as `var fmtr func(string) string`?
    options:
      - text: '`nil`, and calling it panics'
        correct: true
      - text: A function that returns its input unchanged
      - text: A function that returns `""`
    explanation: |
      Function variables default to `nil`. Calling a nil function causes a runtime
      panic, so check `if fmtr != nil` before calling one that might be unset.
---

In Go, functions are **first-class values**. That means a function can go anywhere
a value can go: into a variable, a struct field, a map, a slice, a function argument
or a return value.

## Storing a function in a variable

Leave off the parentheses and you get the function itself rather than the result of
calling it:

```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	format := strings.ToUpper // no (), so we store the function
	fmt.Println(format("doc2doc"))

	format = strings.ToLower // any func(string) string fits
	fmt.Println(format("DOC2DOC"))
}
```

```text
DOC2DOC
doc2doc
```

`format` has type `func(string) string`. Any function with that exact signature can
be stored in it, whether it's from the standard library or your own code.

## Functions in a map

This is where things get useful. Doc2Doc supports several output styles, and the
user picks one by name on the command line. Instead of a long `switch`, store the
formatters in a map:

```go
package main

import (
	"fmt"
	"strings"
)

func shout(s string) string   { return strings.ToUpper(s) + "!" }
func whisper(s string) string { return strings.ToLower(s) + "..." }

func main() {
	formatters := map[string]func(string) string{
		"shout":   shout,
		"whisper": whisper,
		"title":   strings.ToTitle,
	}

	for _, style := range []string{"shout", "whisper", "bold"} {
		f, ok := formatters[style]
		if !ok {
			fmt.Println("unknown style:", style)
			continue
		}
		fmt.Println(f("Hello Docs"))
	}
}
```

```text
HELLO DOCS!
hello docs...
unknown style: bold
```

Adding a new style is now one line in the map, and no control flow changes.

## Nil functions

The zero value of a function type is `nil`. Calling a nil function panics:

```go
var f func(string) string
fmt.Println(f == nil) // true
f("boom")             // panic: runtime error: invalid memory address or nil pointer dereference
```

That's why the map lookup above uses the comma-ok form. A missing key would give you
a nil function, and calling it would crash Doc2Doc.

## Comparing functions

You can compare a function value to `nil`, but you **can't** compare two functions
to each other. `f == g` doesn't compile. For the same reason, functions can't be map
keys. Go doesn't define what "equal functions" means, so it doesn't let you ask.

## Further reading

- [A Tour of Go: Function values](https://go.dev/tour/moretypes/24)
