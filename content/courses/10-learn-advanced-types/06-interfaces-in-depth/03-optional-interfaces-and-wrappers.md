---
title: Optional Interfaces and Wrappers
quiz:
  - question: |
      `w` is a `*strings.Builder`, which has a `WriteString` method. What does this print?

      ```go
      type countingWriter struct {
      	io.Writer
      	n int
      }

      cw := &countingWriter{Writer: w}
      _, ok := any(cw).(io.StringWriter)
      fmt.Println(ok)
      ```
    options:
      - text: '`true`, because the embedded value has `WriteString`'
      - text: '`false`: embedding the `io.Writer` interface promotes only `Write`, whatever the dynamic value can do'
        correct: true
      - text: It doesn't compile
      - text: It panics
    explanation: |
      Promotion is decided at compile time from the *static* type of the embedded field.
      `io.Writer` has one method, so `countingWriter` gets exactly one promoted method.
      The builder's extra methods are hidden behind the wrapper.
  - question: Why does `http.ResponseController` look for an `Unwrap() http.ResponseWriter` method?
    options:
      - text: To decode the response body
      - text: So it can dig through middleware wrappers and find the original writer's optional methods, like flushing
        correct: true
      - text: To unwrap errors
      - text: Because every `ResponseWriter` must implement it
    explanation: |
      Middleware often wraps the `ResponseWriter`, hiding optional interfaces such as
      `http.Flusher`. An `Unwrap` method lets code that needs those features walk down the
      chain of wrappers to the value that really has them.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"io"
    	"strings"
    )

    // CountingWriter counts the bytes written through it to w.
    type CountingWriter struct {
    	w io.Writer
    	n int64
    }

    func NewCountingWriter(w io.Writer) *CountingWriter { return &CountingWriter{w: w} }

    // Write writes p to the underlying writer and counts the bytes written.
    func (c *CountingWriter) Write(p []byte) (int, error) {
    	// ? count the bytes
    	return c.w.Write(p)
    }

    // WriteString writes s, using the underlying writer's own WriteString
    // if it has one (no []byte conversion), and counts the bytes written.
    func (c *CountingWriter) WriteString(s string) (int, error) {
    	// ?
    	return 0, nil
    }

    // Count reports the total bytes written so far.
    func (c *CountingWriter) Count() int64 {
    	// ?
    	return 0
    }

    // Unwrap returns the underlying writer.
    func (c *CountingWriter) Unwrap() io.Writer {
    	// ?
    	return nil
    }

    func main() {
    	var sb strings.Builder
    	cw := NewCountingWriter(&sb)
    	fmt.Fprintf(cw, "entries=%d\n", 3)
    	io.WriteString(cw, "done\n")
    	fmt.Print(sb.String())
    	fmt.Println(cw.Count(), cw.Unwrap() == &sb)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"io"
    	"strings"
    )

    // CountingWriter counts the bytes written through it to w.
    type CountingWriter struct {
    	w io.Writer
    	n int64
    }

    func NewCountingWriter(w io.Writer) *CountingWriter { return &CountingWriter{w: w} }

    // Write writes p to the underlying writer and counts the bytes written.
    func (c *CountingWriter) Write(p []byte) (int, error) {
    	n, err := c.w.Write(p)
    	c.n += int64(n)
    	return n, err
    }

    // WriteString writes s, using the underlying writer's own WriteString
    // if it has one (no []byte conversion), and counts the bytes written.
    func (c *CountingWriter) WriteString(s string) (int, error) {
    	if sw, ok := c.w.(io.StringWriter); ok {
    		n, err := sw.WriteString(s)
    		c.n += int64(n)
    		return n, err
    	}
    	return c.Write([]byte(s))
    }

    // Count reports the total bytes written so far.
    func (c *CountingWriter) Count() int64 { return c.n }

    // Unwrap returns the underlying writer.
    func (c *CountingWriter) Unwrap() io.Writer { return c.w }

    func main() {
    	var sb strings.Builder
    	cw := NewCountingWriter(&sb)
    	fmt.Fprintf(cw, "entries=%d\n", 3)
    	io.WriteString(cw, "done\n")
    	fmt.Print(sb.String())
    	fmt.Println(cw.Count(), cw.Unwrap() == &sb)
    }
  tests: |
    package main

    import (
    	"io"
    	"strings"
    	"testing"
    )

    // recorder remembers which method was used.
    type recorder struct {
    	strings.Builder
    	writes, stringWrites int
    }

    func (r *recorder) Write(p []byte) (int, error) {
    	r.writes++
    	return r.Builder.Write(p)
    }

    func (r *recorder) WriteString(s string) (int, error) {
    	r.stringWrites++
    	return r.Builder.WriteString(s)
    }

    // plain has only Write.
    type plain struct{ data []byte }

    func (p *plain) Write(b []byte) (int, error) {
    	p.data = append(p.data, b...)
    	return len(b), nil
    }

    func TestCountsBytes(t *testing.T) {
    	var sb strings.Builder
    	cw := NewCountingWriter(&sb)
    	cw.Write([]byte("hello "))
    	cw.WriteString("world")
    	if got := cw.Count(); got != 11 {
    		t.Errorf("Count() after writing 11 bytes = %d, want 11", got)
    	}
    	if sb.String() != "hello world" {
    		t.Errorf("underlying writer got %q, want %q", sb.String(), "hello world")
    	}
    }

    func TestForwardsWriteString(t *testing.T) {
    	var r recorder
    	cw := NewCountingWriter(&r)
    	io.WriteString(cw, "abc")
    	if r.stringWrites != 1 || r.writes != 0 {
    		t.Errorf("io.WriteString through CountingWriter: underlying WriteString called %d times, Write %d times; want WriteString once and Write never", r.stringWrites, r.writes)
    	}
    	if cw.Count() != 3 {
    		t.Errorf("Count() = %d, want 3", cw.Count())
    	}
    }

    func TestFallsBackToWrite(t *testing.T) {
    	var p plain
    	cw := NewCountingWriter(&p)
    	if _, err := cw.WriteString("xyz"); err != nil {
    		t.Fatalf("WriteString returned %v", err)
    	}
    	if string(p.data) != "xyz" || cw.Count() != 3 {
    		t.Errorf("WriteString on a Write-only writer: wrote %q, Count() = %d; want \"xyz\", 3", p.data, cw.Count())
    	}
    }

    func TestUnwrap(t *testing.T) {
    	var p plain
    	cw := NewCountingWriter(&p)
    	if cw.Unwrap() != io.Writer(&p) {
    		t.Errorf("Unwrap() should return the writer passed to NewCountingWriter")
    	}
    }
---

Many standard library functions accept a small interface but check, at runtime, whether the value can do *more*. Those extra capabilities are called **optional interfaces**, and they're easy to lose by accident when you wrap a value.

## Asking for more

`io.WriteString(w, s)` accepts any `io.Writer`, but first checks for a faster path:

```go
if sw, ok := w.(io.StringWriter); ok {
	return sw.WriteString(s) // no []byte conversion needed
}
return w.Write([]byte(s))
```

The pattern is everywhere: `io.Copy` checks for `io.WriterTo` and `io.ReaderFrom`, `fmt` checks for `Stringer`, `Formatter` and `error`, `net/http` checks for `http.Flusher`. The function's signature stays small, and capable values get the fast path.

## Wrappers hide capabilities

Now wrap a writer to count bytes. Embedding the `io.Writer` promotes `Write`, which seems like all we need:

```go
package main

import (
	"fmt"
	"io"
	"strings"
)

type ReadWriteCloser interface {
	io.ReadCloser
	io.WriteCloser // both have Close: fine since Go 1.14
}

// countingWriter embeds an io.Writer to promote Write.
type countingWriter struct {
	io.Writer
	n int
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.Writer.Write(p)
	c.n += n
	return n, err
}

func main() {
	var sb strings.Builder
	var w io.Writer = &sb
	_, direct := w.(io.StringWriter)

	cw := &countingWriter{Writer: &sb}
	w = cw
	_, wrapped := w.(io.StringWriter)
	fmt.Println(direct, wrapped)

	io.WriteString(cw, "hello") // falls back to Write([]byte("hello"))
	fmt.Println(cw.n, sb.String())
}
```

```
true false
5 hello
```

The `*strings.Builder` has `WriteString`, but `countingWriter` doesn't. Promotion is decided at **compile time**, from the embedded field's *static* type, `io.Writer`, which has exactly one method. So `io.WriteString` falls back to `Write([]byte(s))`, converting (and possibly allocating) for nothing. Nothing is broken, just slower.

With some optional interfaces, losing them *is* a bug. Wrap an `http.ResponseWriter` in logging middleware and the wrapper no longer has `Flush`, so streaming responses silently stop streaming.

## Overlapping embedded interfaces

A side note on embedding interfaces *in interfaces*: since Go 1.14, the embedded interfaces may share methods, as long as the signatures match. That's why the `ReadWriteCloser` above can embed both `io.ReadCloser` and `io.WriteCloser`, even though both have `Close() error`. Before 1.14 that was a "duplicate method" error.

## Keeping the capabilities

Two techniques, both used in the standard library:

1. **Implement the optional method and forward conditionally.** The wrapper gets `WriteString`, which uses the inner writer's `WriteString` if it has one and falls back to `Write` if not. The wrapper now satisfies `io.StringWriter` either way, and the fast path is preserved when possible.
2. **Offer `Unwrap`.** An `Unwrap()` method returning the inner value lets code that needs a capability dig for it. `http.ResponseController` does exactly this with `Unwrap() http.ResponseWriter`, and `errors.Is`/`errors.As` do it with `Unwrap() error`.

What you *can't* do in Go is create a wrapper whose method set depends on the wrapped value's dynamic type. Method sets are fixed at compile time. (Some libraries generate one wrapper type per combination of optional interfaces for this reason.)

## Your turn

Build Stash's `CountingWriter`:

- `Write` forwards to the inner writer and adds the bytes actually written to the count.
- `WriteString` uses the inner writer's `WriteString` if it's an `io.StringWriter`, and otherwise falls back to `Write`. Either way, count the bytes.
- `Count` returns the total; `Unwrap` returns the inner writer.
