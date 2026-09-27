---
title: Currying
quiz:
  - question: |
      What does this print?

      ```go
      func tag(name string) func(string) string {
          return func(body string) string {
              return "<" + name + ">" + body + "</" + name + ">"
          }
      }

      func main() {
          fmt.Println(tag("em")("wow"))
      }
      ```
    options:
      - text: '`<wow>em</wow>`'
      - text: It doesn't compile, because you can't call a call
      - text: '`emwow`'
      - text: '`<em>wow</em>`'
        correct: true
    explanation: |
      `tag("em")` returns a function. The second pair of parentheses calls that
      function with `"wow"`, producing `<em>wow</em>`.
  - question: What is the curried form of `func(a int, b string) bool`?
    options:
      - text: '`func(int, string) func() bool`'
      - text: '`func(func(int) string) bool`'
      - text: '`func(int) func(string) bool`'
        correct: true
      - text: '`func(string) func(int) bool`'
    explanation: |
      Currying takes the parameters one at a time, in order. You pass the `int`
      and get back a function waiting for the `string`, which finally returns the
      `bool`. (`func(string) func(int) bool` is curried too, but it takes the parameters in the
      wrong order.)
---

**Currying** transforms a function that takes several arguments into a chain of
functions that each take **one** argument. It's named after the logician Haskell
Curry (yes, the language is named after him too).

Instead of:

```go
func wrap(tag, text string) string
wrap("b", "hello")
```

a curried version looks like:

```go
func wrap(tag string) func(string) string
wrap("b")("hello")
```

The first call takes the tag and returns a function that waits for the text.

## A three-level curry

Doc2Doc builds HTML elements with a tag, a CSS class and some content. Here it is
fully curried:

```go
package main

import "fmt"

func element(tag string) func(class string) func(content string) string {
	return func(class string) func(string) string {
		return func(content string) string {
			return fmt.Sprintf(`<%s class="%s">%s</%s>`, tag, class, content, tag)
		}
	}
}

func main() {
	fmt.Println(element("p")("intro")("Welcome to Doc2Doc"))

	span := element("span")
	warning := span("warning")
	note := span("note")

	fmt.Println(warning("Back up your files"))
	fmt.Println(note("Markdown is supported"))
}
```

```text
<p class="intro">Welcome to Doc2Doc</p>
<span class="warning">Back up your files</span>
<span class="note">Markdown is supported</span>
```

The magic is in the middle part. You can stop after any argument and keep the
partially filled-in function for later. `span` is "an element with the tag already
chosen", and `warning` is "a span with the class already chosen". Each one is an
ordinary Go function value that you can store, pass around or put in a map.

## A generic curry helper

Writing curried functions by hand is tedious. Generics can convert any
two-argument function:

```go
func Curry[A, B, R any](f func(A, B) R) func(A) func(B) R {
	return func(a A) func(B) R {
		return func(b B) R {
			return f(a, b)
		}
	}
}
```

Now `Curry(strings.Repeat)` has type `func(string) func(int) string`:

```go
dashes := Curry(strings.Repeat)("-")
fmt.Println(dashes(10)) // ----------
```

A `Curry3` for three arguments follows the same pattern with one more layer.

## Honest talk

In Haskell, *every* function is curried automatically, and `wrap "b" "hello"` is
just two function calls in a row. In Go, currying is manual, the types grow nested
`func`s quickly, and `f(a)(b)(c)` looks unusual to most Go readers. The next two
lessons cover the idea behind currying that *is* useful every day in Go, partial
application, and the situations where currying earns its keep.
