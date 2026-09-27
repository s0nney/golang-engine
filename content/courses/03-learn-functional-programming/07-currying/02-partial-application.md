---
title: Partial Application
quiz:
  - question: What's the difference between currying and partial application?
    options:
      - text: Currying turns an n-argument function into a chain of one-argument functions. Partial application fixes some arguments now and returns a function taking the rest
        correct: true
      - text: They're two names for the same thing
      - text: Partial application only works with methods
      - text: Currying fixes the last argument and partial application fixes the first
    explanation: |
      Currying always reshapes the function into single-argument steps. Partial
      application just pre-fills some arguments, however many, and leaves a
      function waiting for the others.
  - question: |
      What does this print?

      ```go
      func Partial[A, B, R any](f func(A, B) R, a A) func(B) R {
          return func(b B) R { return f(a, b) }
      }

      func main() {
          hasMd := Partial(strings.HasSuffix, ".md")
          fmt.Println(hasMd("notes.md"))
      }
      ```
    options:
      - text: '`false`'
        correct: true
      - text: '`true`'
      - text: It doesn't compile
    explanation: |
      `strings.HasSuffix(s, suffix)` takes the string first and the suffix second.
      `Partial` fixed the *first* argument, so this calls
      `strings.HasSuffix(".md", "notes.md")`, asking whether `".md"` ends in
      `"notes.md"`. It doesn't. Argument order matters!
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    func Partial[A, B, R any](f func(A, B) R, a A) func(B) R {
    	return func(b B) R { return f(a, b) }
    }

    // Flip returns a function that takes f's two arguments in the opposite order.
    func Flip[A, B, R any](f func(A, B) R) func(B, A) R {
    	return func(b B, a A) R {
    		var zero R
    		return zero // ?
    	}
    }

    // Uncurry turns a curried function back into a two-argument one:
    // Uncurry(f)(a, b) == f(a)(b).
    func Uncurry[A, B, R any](f func(A) func(B) R) func(A, B) R {
    	return func(a A, b B) R {
    		var zero R
    		return zero // ?
    	}
    }

    func wrapTag(tag string) func(string) string {
    	return func(s string) string { return "<" + tag + ">" + s + "</" + tag + ">" }
    }

    func main() {
    	hasTodo := Partial(Flip(strings.Contains), "TODO")
    	fmt.Println(hasTodo("TODO: add the date"), hasTodo("all done")) // true false

    	wrap := Uncurry(wrapTag)
    	fmt.Println(wrap("em", "really")) // <em>really</em>
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    func Partial[A, B, R any](f func(A, B) R, a A) func(B) R {
    	return func(b B) R { return f(a, b) }
    }

    func Flip[A, B, R any](f func(A, B) R) func(B, A) R {
    	return func(b B, a A) R { return f(a, b) }
    }

    func Uncurry[A, B, R any](f func(A) func(B) R) func(A, B) R {
    	return func(a A, b B) R { return f(a)(b) }
    }

    func wrapTag(tag string) func(string) string {
    	return func(s string) string { return "<" + tag + ">" + s + "</" + tag + ">" }
    }

    func main() {
    	hasTodo := Partial(Flip(strings.Contains), "TODO")
    	fmt.Println(hasTodo("TODO: add the date"), hasTodo("all done"))

    	wrap := Uncurry(wrapTag)
    	fmt.Println(wrap("em", "really"))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"strings"
    	"testing"
    )

    func TestFlip(t *testing.T) {
    	flipped := Flip(strings.Repeat)
    	if flipped == nil {
    		t.Fatal("Flip(strings.Repeat) returned nil")
    	}
    	if got := flipped(3, "ab"); got != "ababab" {
    		t.Errorf("Flip(strings.Repeat)(3, %q) = %q, want %q", "ab", got, "ababab")
    	}
    	sub := Flip(func(a, b int) int { return a - b })
    	if got := sub(10, 3); got != -7 {
    		t.Errorf("Flip(a - b)(10, 3) = %d, want -7", got)
    	}
    	contains := Flip(strings.Contains)
    	if !contains("TODO", "x TODO y") || contains("TODO", "done") {
    		t.Errorf("Flip(strings.Contains)(substr, s) gives the wrong answers")
    	}
    }

    func TestUncurry(t *testing.T) {
    	join := Uncurry(func(sep string) func(parts []string) string {
    		return func(parts []string) string { return strings.Join(parts, sep) }
    	})
    	if join == nil {
    		t.Fatal("Uncurry(...) returned nil")
    	}
    	if got := join(", ", []string{"a", "b"}); got != "a, b" {
    		t.Errorf("Uncurry(join)(%q, [a b]) = %q, want %q", ", ", got, "a, b")
    	}
    	w := Uncurry(wrapTag)
    	if got := w("code", "go vet"); got != "<code>go vet</code>" {
    		t.Errorf("Uncurry(wrapTag)(%q, %q) = %q, want %q", "code", "go vet", got, "<code>go vet</code>")
    	}
    	calls := 0
    	count := Uncurry(func(n int) func(s string) string {
    		calls++
    		return func(s string) string { return fmt.Sprint(n, s) }
    	})
    	count(1, "x")
    	count(2, "y")
    	if calls != 2 {
    		t.Errorf("Uncurry should call the outer function once per call, got %d calls for 2 uses", calls)
    	}
    }
---

**Partial application** means fixing some of a function's arguments now and getting
back a function that takes the rest later. It's the practical cousin of currying, and
it's something Go programmers do all the time, often without naming it.

- **Currying**: `f(a, b, c)` becomes `f(a)(b)(c)`. Always one argument at a time.
- **Partial application**: `f(a, b, c)` with `a` fixed becomes `g(b, c)`. Fix as many
  arguments as you like, and take the rest in one go.

## Partial application with a closure

The simplest way in Go is a closure:

```go
package main

import (
	"fmt"
	"strings"
)

func replaceIn(old, new, doc string) string {
	return strings.ReplaceAll(doc, old, new)
}

func main() {
	// Fix the first two arguments, leave doc open.
	fixBrand := func(doc string) string {
		return replaceIn("doc-2-doc", "Doc2Doc", doc)
	}

	fmt.Println(fixBrand("Try doc-2-doc today!"))
	fmt.Println(fixBrand("doc-2-doc v2 is out"))
}
```

```text
Try Doc2Doc today!
Doc2Doc v2 is out
```

`fixBrand` is `replaceIn` with two of its three arguments already decided.

## A generic helper

```go
func Partial[A, B, R any](f func(A, B) R, a A) func(B) R {
	return func(b B) R { return f(a, b) }
}
```

It fixes the **first** argument. That's where argument order bites in Go: most
standard library functions put the "subject" first (`strings.HasPrefix(s, prefix)`,
`strings.Contains(s, substr)`), so the argument you'd want to fix, the prefix or
substring, is usually the *second* one. You'd need a mirror-image helper:

```go
func PartialRight[A, B, R any](f func(A, B) R, b B) func(A) R {
	return func(a A) R { return f(a, b) }
}

isMarkdown := PartialRight(strings.HasSuffix, ".md")
isMarkdown("notes.md") // true
```

Or just write the closure, which is often clearer:
`func(s string) bool { return strings.HasSuffix(s, ".md") }`.

## You already do this

Several everyday Go idioms are partial application in disguise:

- **Method values.** `replacer.Replace` is `(*Replacer).Replace` with the receiver
  fixed.
- **Constructors that capture config.** `strings.NewReplacer("--", "—")` fixes the
  replacement pairs and gives you something to call on many documents.
- **Handler factories.** `handleConvert(store)` fixes a dependency and returns an
  `http.HandlerFunc`, which the router calls with the remaining arguments.

Whenever you catch yourself passing the same first few arguments over and over,
partial application lets you fix them once and name the result.

## Your turn

`Partial` fixes the *first* argument. But `strings.Contains(s, substr)` takes the document first, and Doc2Doc wants to fix the *substring*: "does this document contain TODO?". Two more generic helpers make any function fit:

1. `Flip(f)` returns a function taking `f`'s two arguments in the opposite order: `Flip(f)(b, a) == f(a, b)`.
2. `Uncurry(f)` turns a curried function back into a two-argument one: `Uncurry(f)(a, b) == f(a)(b)`.

With them, `Partial(Flip(strings.Contains), "TODO")` is a ready-made `func(string) bool`.
