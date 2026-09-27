---
title: Chaining Middleware
quiz:
  - question: |
      With `Chain(c, A, B, C)` as defined in this lesson, in what order do the middlewares see an incoming document?
    options:
      - text: C, B, A
      - text: The order is random
      - text: A, B, C
        correct: true
      - text: Only A runs
    explanation: |
      `Chain` applies the list in reverse, so `A` ends up as the outermost layer.
      Documents enter through `A`, then `B`, then `C`, then reach the converter.
      Results unwind in the opposite order.
  - question: Why should a panic-recovery middleware usually be the *outermost* layer?
    options:
      - text: Because `recover` only works in `main`
      - text: So it can catch panics from every other middleware and the converter inside it
        correct: true
      - text: Because it's the slowest middleware
    explanation: |
      A deferred `recover` only catches panics from code it wraps. Put it on the
      outside and it protects everything else in the chain.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    type Converter func(string) string

    type Middleware func(Converter) Converter

    // Chain wraps c in every middleware so that the FIRST one in mws is the
    // outermost layer (it sees the document first).
    func Chain(c Converter, mws ...Middleware) Converter {
    	// ?
    	return c
    }

    // trimSpace trims leading and trailing whitespace from the document
    // before calling next.
    func trimSpace(next Converter) Converter {
    	// ?
    	return next
    }

    func shout(next Converter) Converter {
    	return func(doc string) string {
    		return next(strings.ToUpper(doc))
    	}
    }

    func paragraph(doc string) string {
    	return "<p>" + doc + "</p>"
    }

    func main() {
    	c := Chain(paragraph, trimSpace, shout)
    	fmt.Println(c("   hello, doc2doc  \n"))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"slices"
    	"strings"
    )

    type Converter func(string) string

    type Middleware func(Converter) Converter

    func Chain(c Converter, mws ...Middleware) Converter {
    	for _, mw := range slices.Backward(mws) {
    		c = mw(c)
    	}
    	return c
    }

    func trimSpace(next Converter) Converter {
    	return func(doc string) string {
    		return next(strings.TrimSpace(doc))
    	}
    }

    func shout(next Converter) Converter {
    	return func(doc string) string {
    		return next(strings.ToUpper(doc))
    	}
    }

    func paragraph(doc string) string {
    	return "<p>" + doc + "</p>"
    }

    func main() {
    	c := Chain(paragraph, trimSpace, shout)
    	fmt.Println(c("   hello, doc2doc  \n"))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func tag(name string, events *[]string) Middleware {
    	return func(next Converter) Converter {
    		return func(doc string) string {
    			*events = append(*events, "enter "+name)
    			out := next(doc)
    			*events = append(*events, "leave "+name)
    			return out
    		}
    	}
    }

    func TestChainOrder(t *testing.T) {
    	var events []string
    	final := func(doc string) string {
    		events = append(events, "convert "+doc)
    		return doc
    	}
    	c := Chain(final, tag("A", &events), tag("B", &events), tag("C", &events))
    	c("x")
    	want := []string{"enter A", "enter B", "enter C", "convert x", "leave C", "leave B", "leave A"}
    	if !slices.Equal(events, want) {
    		t.Errorf("Chain(c, A, B, C) order:\n got %q\nwant %q", events, want)
    	}
    }

    func TestChainNoMiddleware(t *testing.T) {
    	if got := Chain(paragraph)("hi"); got != "<p>hi</p>" {
    		t.Errorf("Chain(paragraph)(%q) = %q, want %q", "hi", got, "<p>hi</p>")
    	}
    }

    func TestTrimSpace(t *testing.T) {
    	var seen string
    	spy := func(doc string) string { seen = doc; return "[" + doc + "]" }
    	for _, tt := range []struct{ in, want string }{
    		{"  hello  ", "hello"},
    		{"\n\tdoc\n", "doc"},
    		{"already tidy", "already tidy"},
    		{"   ", ""},
    	} {
    		got := trimSpace(spy)(tt.in)
    		if seen != tt.want {
    			t.Errorf("trimSpace passed %q to the next converter, want %q", seen, tt.want)
    		}
    		if got != "["+tt.want+"]" {
    			t.Errorf("trimSpace(next)(%q) = %q, want next's result %q", tt.in, got, "["+tt.want+"]")
    		}
    	}
    }

    func TestChainWithTrimSpace(t *testing.T) {
    	c := Chain(paragraph, trimSpace, shout)
    	if got, want := c("  hi there \n"), "<p>HI THERE</p>"; got != want {
    		t.Errorf("Chain(paragraph, trimSpace, shout)(%q) = %q, want %q", "  hi there \n", got, want)
    	}
    }
---

Nesting middleware by hand reads inside-out and is easy to get wrong:

```go
c := recoverPanics(logCalls(maxSize(1 << 20)(markdownToHTML)))
```

Since every middleware has the same type, you can store them in a slice and apply
them in a loop. That's function composition again, specialised to converters.

## A Chain helper

```go
func Chain(c Converter, mws ...Middleware) Converter {
	for _, mw := range slices.Backward(mws) {
		c = mw(c)
	}
	return c
}
```

`slices.Backward` walks the slice from the end, so the *first* middleware in the list
becomes the *outermost* wrapper. That way the list reads in the order documents flow
through it.

## The onion

Each middleware wraps the next, like layers of an onion. A document travels in
through every layer, reaches the converter, and the result travels back out:

```go
package main

import (
	"fmt"
	"slices"
	"strings"
)

type Converter func(string) string

type Middleware func(Converter) Converter

func Chain(c Converter, mws ...Middleware) Converter {
	for _, mw := range slices.Backward(mws) {
		c = mw(c)
	}
	return c
}

func trace(name string) Middleware {
	return func(next Converter) Converter {
		return func(doc string) string {
			fmt.Println("enter", name)
			out := next(doc)
			fmt.Println("leave", name)
			return out
		}
	}
}

func recoverPanics(next Converter) Converter {
	return func(doc string) (out string) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("recovered:", err)
				out = "<!-- conversion failed -->"
			}
		}()
		return next(doc)
	}
}

func convert(doc string) string {
	fmt.Println("  converter runs")
	if strings.HasPrefix(doc, "%PDF") {
		panic("PDF input not implemented")
	}
	return "<p>" + doc + "</p>"
}

func main() {
	c := Chain(convert,
		recoverPanics,
		trace("logging"),
		trace("limits"),
	)

	for _, doc := range []string{"hello", "%PDF-1.7 ..."} {
		fmt.Println("result:", c(doc))
	}
}
```

```text
enter logging
enter limits
  converter runs
leave limits
leave logging
result: <p>hello</p>
enter logging
enter limits
  converter runs
recovered: PDF input not implemented
result: <!-- conversion failed -->
```

For the second document the converter panicked. The panic skipped the "leave" lines
entirely (the stack unwound straight past them) until `recoverPanics`, the outermost
layer, caught it. Its deferred function sets the **named result** `out`, so the chain
returns a placeholder instead of crashing the whole batch.

## Ordering tips

- **Recovery** goes outermost, so it protects everything.
- **Logging** goes near the outside, so it sees every document, including rejected ones.
- **Validation** (like `maxSize`) goes before anything expensive, so bad input stops early.
- **Per-format** middleware (like "strip smart quotes from Word documents") wraps just
  that format's converter, not every converter.

## The decorator pattern, Go style

Python gives decorators special syntax. Go gives you something arguably clearer:
ordinary functions with matching types, combined with ordinary loops. There's no
magic. You can read `Chain` in ten seconds and see exactly what order things happen
in.

## Your turn

1. Finish `Chain(c, mws...)` so the **first** middleware in the list is the
   outermost layer. With no middleware, it returns `c` unchanged.
2. Finish `trimSpace`, a middleware that trims leading and trailing whitespace from
   the document **before** passing it to the next converter.
