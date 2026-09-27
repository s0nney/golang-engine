---
title: 'Practice: Logging Middleware'
exercise:
  starter: |
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

    // logSizes returns a middleware that, AFTER the wrapped converter has
    // run, calls logf with "NAME: IN bytes -> OUT bytes",
    // e.g. "md->html: 7 bytes -> 14 bytes".
    func logSizes(name string, logf func(string)) Middleware {
    	return func(next Converter) Converter {
    		// ?
    		return next
    	}
    }

    // recoverPanics returns a middleware that stops a panic in the wrapped
    // converter from escaping. When one happens it calls logf with
    // "panic: <value>" and the converter returns "<!-- conversion failed -->".
    func recoverPanics(logf func(string)) Middleware {
    	return func(next Converter) Converter {
    		// ?
    		return next
    	}
    }

    func markdownToHTML(doc string) string {
    	if strings.HasPrefix(doc, "%PDF") {
    		panic("PDF input not supported")
    	}
    	return "<h1>" + strings.TrimPrefix(doc, "# ") + "</h1>"
    }

    func main() {
    	logf := func(s string) { fmt.Println("log:", s) }
    	convert := Chain(markdownToHTML, recoverPanics(logf), logSizes("md->html", logf))

    	// The PDF will panic until recoverPanics works.
    	for _, doc := range []string{"# Hello", "%PDF-1.7"} {
    		fmt.Println("result:", convert(doc))
    	}
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

    func logSizes(name string, logf func(string)) Middleware {
    	return func(next Converter) Converter {
    		return func(doc string) string {
    			out := next(doc)
    			logf(fmt.Sprintf("%s: %d bytes -> %d bytes", name, len(doc), len(out)))
    			return out
    		}
    	}
    }

    func recoverPanics(logf func(string)) Middleware {
    	return func(next Converter) Converter {
    		return func(doc string) (out string) {
    			defer func() {
    				if r := recover(); r != nil {
    					logf(fmt.Sprint("panic: ", r))
    					out = "<!-- conversion failed -->"
    				}
    			}()
    			return next(doc)
    		}
    	}
    }

    func markdownToHTML(doc string) string {
    	if strings.HasPrefix(doc, "%PDF") {
    		panic("PDF input not supported")
    	}
    	return "<h1>" + strings.TrimPrefix(doc, "# ") + "</h1>"
    }

    func main() {
    	logf := func(s string) { fmt.Println("log:", s) }
    	convert := Chain(markdownToHTML, recoverPanics(logf), logSizes("md->html", logf))

    	for _, doc := range []string{"# Hello", "%PDF-1.7"} {
    		fmt.Println("result:", convert(doc))
    	}
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func TestLogSizes(t *testing.T) {
    	var events []string
    	logf := func(s string) { events = append(events, s) }
    	inner := func(doc string) string {
    		events = append(events, "convert")
    		return "<p>" + doc + "</p>"
    	}
    	c := logSizes("txt->html", logf)(inner)
    	if got := c("hello"); got != "<p>hello</p>" {
    		t.Errorf("wrapped converter returned %q, want %q (return next's result)", got, "<p>hello</p>")
    	}
    	want := []string{"convert", "txt->html: 5 bytes -> 12 bytes"}
    	if !slices.Equal(events, want) {
    		t.Errorf("events = %q, want %q (log once, after the converter runs)", events, want)
    	}
    }

    func TestRecoverPanics(t *testing.T) {
    	var logs []string
    	logf := func(s string) { logs = append(logs, s) }
    	c := recoverPanics(logf)(markdownToHTML)

    	if got := c("# Hi"); got != "<h1>Hi</h1>" {
    		t.Errorf("recoverPanics(...)(markdownToHTML)(%q) = %q, want %q", "# Hi", got, "<h1>Hi</h1>")
    	}
    	if len(logs) != 0 {
    		t.Errorf("logged %q for a conversion that didn't panic, want nothing", logs)
    	}

    	func() {
    		defer func() {
    			if r := recover(); r != nil {
    				t.Fatalf("the panic escaped recoverPanics: %v", r)
    			}
    		}()
    		if got := c("%PDF-1.7"); got != "<!-- conversion failed -->" {
    			t.Errorf("after a panic the converter returned %q, want %q", got, "<!-- conversion failed -->")
    		}
    	}()
    	want := []string{"panic: PDF input not supported"}
    	if !slices.Equal(logs, want) {
    		t.Errorf("logged %q, want %q", logs, want)
    	}
    }

    func TestFullChain(t *testing.T) {
    	var logs []string
    	logf := func(s string) { logs = append(logs, s) }
    	c := Chain(markdownToHTML, recoverPanics(logf), logSizes("md->html", logf))
    	c("# Hello")
    	c("%PDF-1.7")
    	want := []string{"md->html: 7 bytes -> 14 bytes", "panic: PDF input not supported"}
    	if !slices.Equal(logs, want) {
    		t.Errorf("logs = %q, want %q", logs, want)
    	}
    }
---

Doc2Doc converts whole folders of documents in one batch. Two things keep going
wrong: nobody can tell which converters are slow or bloated, and one bad file (a PDF
sneaking in among the Markdown) panics and kills the entire batch.

Both fixes belong in middleware, not in the converters.

## Your task

`Converter`, `Middleware` and `Chain` are already written. Finish two configurable
middlewares, each taking a `logf func(string)` so the caller decides where log lines
go (the tests collect them in a slice):

1. **`logSizes(name, logf)`** returns a `Middleware` that calls the next converter,
   then calls `logf` with `"NAME: IN bytes -> OUT bytes"`, for example
   `md->html: 7 bytes -> 14 bytes`. It returns the converter's result unchanged.
2. **`recoverPanics(logf)`** returns a `Middleware` that calls the next converter but
   stops a panic from escaping. If one happens, it calls `logf` with
   `"panic: "` followed by the panic value, and the converter returns
   `<!-- conversion failed -->`. Normal conversions pass through untouched and log
   nothing.

```go
convert := Chain(markdownToHTML, recoverPanics(logf), logSizes("md->html", logf))
```

## Tips

- Use `len(doc)` and `len(out)` for the byte counts.
- A deferred function can change what the surrounding function returns only through
  a **named result**: `return func(doc string) (out string) { defer ...; return next(doc) }`.
- `fmt.Sprint("panic: ", r)` turns any panic value into a string.
- `recoverPanics` is first in the `Chain`, so it's the outermost layer. When the PDF
  panics, `logSizes` never gets to log, and that's expected.
