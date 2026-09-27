---
title: Function Composition
quiz:
  - question: |
      Using `compose` from this lesson, where `compose(f, g)(x)` means `g(f(x))`, what does this print?

      ```go
      addBang := func(s string) string { return s + "!" }
      pipeline := compose(addBang, strings.ToUpper)
      fmt.Println(pipeline("done"))
      ```
    options:
      - text: '`done!`'
      - text: '`DONE`'
      - text: '`!DONE`'
      - text: '`DONE!`'
        correct: true
    explanation: |
      `addBang` runs first, giving `"done!"`, then `strings.ToUpper` gives
      `"DONE!"`. Since `!` has no upper-case form, the order doesn't change the
      result here, but in general it matters.
  - question: |
      What does this print?

      ```go
      trim := func(s string) string { return strings.Trim(s, "*") }
      bold := func(s string) string { return "*" + s + "*" }
      fmt.Println(pipe(bold, trim)("hi"), pipe(trim, bold)("hi"))
      ```
      (`pipe` runs its functions left to right.)
    options:
      - text: '`*hi* *hi*`'
      - text: '`hi hi`'
      - text: '`*hi* hi`'
      - text: '`hi *hi*`'
        correct: true
    explanation: |
      `pipe(bold, trim)` adds the stars and then removes them, giving `hi`.
      `pipe(trim, bold)` trims nothing and then adds stars, giving `*hi*`. Order
      matters in composition.
---

**Composition** means combining two functions into one: the output of the first
becomes the input of the second. In maths it's written `g ∘ f`, meaning "apply `f`,
then `g`".

You've been composing all along by nesting calls:

```go
out := addFooter(strings.ToUpper(strings.TrimSpace(doc)))
```

That works, but you have to read it inside out, and you can't store the combination
to reuse it. Composition gives you a *new function* that does all the steps.

## A compose function

```go
package main

import (
	"fmt"
	"strings"
)

type Transform func(string) string

// compose returns a function that runs f, then g.
func compose(f, g Transform) Transform {
	return func(s string) string {
		return g(f(s))
	}
}

// pipe runs any number of transforms, left to right.
func pipe(steps ...Transform) Transform {
	return func(s string) string {
		for _, step := range steps {
			s = step(s)
		}
		return s
	}
}

func addFooter(s string) string {
	return s + "\n--\nDoc2Doc"
}

func main() {
	shoutTrimmed := compose(strings.TrimSpace, strings.ToUpper)
	fmt.Printf("%q\n", shoutTrimmed("  hello  "))

	publish := pipe(strings.TrimSpace, strings.ToUpper, addFooter)
	fmt.Println(publish("   release notes   "))
}
```

```text
"HELLO"
RELEASE NOTES
--
Doc2Doc
```

`publish` is a brand new function built entirely from existing ones. You can store
it, pass it around, or compose it further: `pipe(publish, wrapInHTML)`.

## Order matters

Composition isn't commutative. "Trim then add a footer" is not "add a footer then
trim":

```go
a := pipe(addFooter, strings.TrimSpace) // footer, then trim
b := pipe(strings.TrimSpace, addFooter) // trim, then footer
fmt.Printf("%q\n", a(" hi "))          // "hi \n--\nDoc2Doc"
fmt.Printf("%q\n", b(" hi "))          // "hi\n--\nDoc2Doc"
```

In `a`, the trailing space after `hi` is no longer at the end of the string when
`TrimSpace` runs, so it survives. Always think about which way data flows.

## Composition as configuration

Because pipelines are values, Doc2Doc can build them from user settings:

```go
steps := []Transform{strings.TrimSpace}
if cfg.Uppercase {
	steps = append(steps, strings.ToUpper)
}
if cfg.Footer {
	steps = append(steps, addFooter)
}
convert := pipe(steps...)
```

The file-processing loop just calls `convert(doc)` and never needs to know which
options were chosen. Adding a new option means adding a new small function, not
editing a big `if` chain in the middle of your processing code.

## compose vs pipe

`compose` and `pipe` only differ in argument order. Maths-y libraries tend to call
right-to-left composition `compose`, and left-to-right `pipe` or `then`. Since Go
code usually reads top to bottom, left to right, `pipe` tends to feel more natural.
