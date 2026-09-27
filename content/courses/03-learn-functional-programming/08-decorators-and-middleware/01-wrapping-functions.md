---
title: Wrapping Functions
quiz:
  - question: Go has no `@decorator` syntax. How do you add behaviour around an existing function?
    options:
      - text: Use reflection to patch the function at runtime
      - text: Embed the function in a struct
      - text: Write a function that takes the original and returns a new function with the same signature that calls it
        correct: true
      - text: It can't be done in Go
    explanation: |
      A wrapper takes a function and returns a new one of the same type. The new
      one does extra work before or after calling the original. That's all a
      decorator really is.
  - question: |
      What does this print?

      ```go
      type Transform func(string) string

      func skipEmpty(next Transform) Transform {
          return func(s string) string {
              if s == "" {
                  return ""
              }
              return next(s)
          }
      }

      func main() {
          wrap := skipEmpty(func(s string) string { return "[" + s + "]" })
          fmt.Printf("%q %q\n", wrap(""), wrap("x"))
      }
      ```
    options:
      - text: '`"" "[x]"`'
        correct: true
      - text: '`"[]" "[x]"`'
      - text: '`"" ""`'
    explanation: |
      For the empty string the wrapper returns early and never calls `next`. For
      `"x"` it passes through to the original function, giving `"[x]"`.
---

In Python, a **decorator** is a function that takes a function and returns an
enhanced version of it, applied with the `@` syntax:

```python
@log_calls
def convert(doc): ...
```

Go has no `@` syntax, no annotations and no way to modify a function after it's
declared. But the *idea* behind decorators doesn't need special syntax at all. It's
just a higher-order function. You write the wrapping explicitly:

```go
convert := logCalls(convertMarkdown)
```

## The shape of a wrapper

A wrapper takes a function and returns a new function **with the same signature**:

```go
func wrapper(next Transform) Transform {
	return func(s string) string {
		// before
		out := next(s)
		// after
		return out
	}
}
```

Because the result has the same type as the input, callers can't tell the wrapped
function from the original. They just call it. And because the type matches, you can
wrap it again and again.

## Doc2Doc: guarding a converter

Doc2Doc's converters sometimes get empty input, or input that's far too big. Rather
than add the same checks to every converter, wrap them:

```go
package main

import (
	"fmt"
	"strings"
)

type Transform func(string) string

func skipEmpty(next Transform) Transform {
	return func(s string) string {
		if strings.TrimSpace(s) == "" {
			return ""
		}
		return next(s)
	}
}

func truncate(limit int, next Transform) Transform {
	return func(s string) string {
		out := next(s)
		if len(out) > limit {
			return out[:limit] + "..."
		}
		return out
	}
}

func toHeading(s string) string {
	return "# " + strings.ToUpper(s)
}

func main() {
	safe := truncate(12, skipEmpty(toHeading))

	fmt.Printf("%q\n", safe("   "))
	fmt.Printf("%q\n", safe("intro"))
	fmt.Printf("%q\n", safe("a very long heading"))
}
```

```text
""
"# INTRO"
"# A VERY LON..."
```

`toHeading` stays tiny and focused. The guards live in reusable wrappers that you can
apply to *any* `Transform`.

(`out[:limit]` slices bytes, so it could cut a multi-byte UTF-8 character in half.
Fine for this ASCII demo, but a real tool would count runes.)

## Before, after or instead

A wrapper can:

- run code **before** the call (validate input, start a timer, check permissions),
- run code **after** it (log the result, record the duration, post-process output),
- or **skip** the call entirely (cache hits, invalid input, rate limits).

You've seen one already: `memoize` from the Pure Functions chapter is a wrapper that
sometimes skips the call.

## Order matters

`truncate(12, skipEmpty(toHeading))` and `skipEmpty(truncate(12, toHeading))` behave
slightly differently, because the outer wrapper runs first on the way in and last on
the way out. Picture each wrapper as a layer of an onion around the original function.
You'll see this again with middleware later in this chapter.
