---
title: Conversion Middleware
difficulty: hard
after: decorators-and-middleware
hints:
  - 'Every middleware has the same shape: `return func(next Handler) Handler { return func(req Request) Response { ... } }`. To **short-circuit**, return a `Response` without calling `next`. `Trace` calls `logf` before `next(req)` and again, with the status, after it returns.'
  - '`Recover` needs a deferred function that can change the response after a panic. Give the inner function a **named result**, `func(req Request) (resp Response)`, and in `defer func() { if recover() != nil { resp = Response{500, "internal error"} } }()`.'
  - '`OnlyFor(format, mw)` returns a middleware that, given `next`, builds `wrapped := mw(next)` **once**, then returns a handler that sends matching requests to `wrapped` and everything else straight to `next`.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"slices"
    	"strings"
    )

    type Request struct {
    	User   string
    	Format string
    	Body   string
    }

    type Response struct {
    	Status int
    	Body   string
    }

    type Handler func(Request) Response

    type Middleware func(Handler) Handler

    // Chain wraps h so that the FIRST middleware is the outermost layer.
    // It's already written for you.
    func Chain(h Handler, mws ...Middleware) Handler {
    	for _, mw := range slices.Backward(mws) {
    		h = mw(h)
    	}
    	return h
    }

    func Trace(name string, logf func(string)) Middleware {
    	return func(next Handler) Handler { return next }
    }

    func RequireUser(allowed ...string) Middleware {
    	return func(next Handler) Handler { return next }
    }

    func MaxBody(n int) Middleware {
    	return func(next Handler) Handler { return next }
    }

    func Recover() Middleware {
    	return func(next Handler) Handler { return next }
    }

    func OnlyFor(format string, mw Middleware) Middleware {
    	return func(next Handler) Handler { return next }
    }

    func convert(req Request) Response {
    	if req.Format == "pdf" && req.Body == "" {
    		panic("empty PDF")
    	}
    	return Response{200, strings.ToUpper(req.Body)}
    }

    func main() {
    	logf := func(s string) { fmt.Println("  log:", s) }
    	h := Chain(convert,
    		Recover(),
    		Trace("outer", logf),
    		RequireUser("ada", "linus"),
    		OnlyFor("pdf", MaxBody(10)),
    		Trace("inner", logf),
    	)
    	for _, req := range []Request{
    		{"ada", "html", "hello"},                // want 200 HELLO, logs outer> inner> inner<200 outer<200
    		{"eve", "html", "hello"},                // want 403 forbidden, logs outer> outer<403
    		{"ada", "pdf", "this body is too long"}, // want 413 too large, logs outer> outer<413
    		{"linus", "pdf", ""},                    // want 500 internal error, logs outer> inner>
    	} {
    		fmt.Printf("%+v\n", req)
    		fmt.Printf("  -> %+v\n", h(req))
    	}
    }
  solution: |
    package main

    import (
    	"fmt"
    	"slices"
    	"strings"
    )

    type Request struct {
    	User   string
    	Format string
    	Body   string
    }

    type Response struct {
    	Status int
    	Body   string
    }

    type Handler func(Request) Response

    type Middleware func(Handler) Handler

    // Chain wraps h so that the FIRST middleware is the outermost layer.
    // It's already written for you.
    func Chain(h Handler, mws ...Middleware) Handler {
    	for _, mw := range slices.Backward(mws) {
    		h = mw(h)
    	}
    	return h
    }

    func Trace(name string, logf func(string)) Middleware {
    	return func(next Handler) Handler {
    		return func(req Request) Response {
    			logf(name + ">")
    			resp := next(req)
    			logf(fmt.Sprintf("%s<%d", name, resp.Status))
    			return resp
    		}
    	}
    }

    func RequireUser(allowed ...string) Middleware {
    	return func(next Handler) Handler {
    		return func(req Request) Response {
    			if !slices.Contains(allowed, req.User) {
    				return Response{403, "forbidden"}
    			}
    			return next(req)
    		}
    	}
    }

    func MaxBody(n int) Middleware {
    	return func(next Handler) Handler {
    		return func(req Request) Response {
    			if len(req.Body) > n {
    				return Response{413, "too large"}
    			}
    			return next(req)
    		}
    	}
    }

    func Recover() Middleware {
    	return func(next Handler) Handler {
    		return func(req Request) (resp Response) {
    			defer func() {
    				if recover() != nil {
    					resp = Response{500, "internal error"}
    				}
    			}()
    			return next(req)
    		}
    	}
    }

    func OnlyFor(format string, mw Middleware) Middleware {
    	return func(next Handler) Handler {
    		wrapped := mw(next)
    		return func(req Request) Response {
    			if req.Format == format {
    				return wrapped(req)
    			}
    			return next(req)
    		}
    	}
    }

    func convert(req Request) Response {
    	if req.Format == "pdf" && req.Body == "" {
    		panic("empty PDF")
    	}
    	return Response{200, strings.ToUpper(req.Body)}
    }

    func main() {
    	logf := func(s string) { fmt.Println("  log:", s) }
    	h := Chain(convert,
    		Recover(),
    		Trace("outer", logf),
    		RequireUser("ada", "linus"),
    		OnlyFor("pdf", MaxBody(10)),
    		Trace("inner", logf),
    	)
    	for _, req := range []Request{
    		{"ada", "html", "hello"},
    		{"eve", "html", "hello"},
    		{"ada", "pdf", "this body is too long"},
    		{"linus", "pdf", ""},
    	} {
    		fmt.Printf("%+v\n", req)
    		fmt.Printf("  -> %+v\n", h(req))
    	}
    }
  tests: |
    package main

    import (
    	"fmt"
    	"slices"
    	"testing"
    )

    // recorder is a handler that records every request it receives.
    type recorder struct{ got []Request }

    func (r *recorder) handle(req Request) Response {
    	r.got = append(r.got, req)
    	return Response{200, "ok:" + req.Body}
    }

    func logger() (*[]string, func(string)) {
    	var lines []string
    	return &lines, func(s string) { lines = append(lines, s) }
    }

    func TestTraceOrder(t *testing.T) {
    	lines, logf := logger()
    	rec := &recorder{}
    	h := Chain(rec.handle, Trace("a", logf), Trace("b", logf), Trace("c", logf))
    	resp := h(Request{"ada", "html", "x"})
    	if resp != (Response{200, "ok:x"}) {
    		t.Errorf("response = %+v, want {Status:200 Body:ok:x}", resp)
    	}
    	want := []string{"a>", "b>", "c>", "c<200", "b<200", "a<200"}
    	if !slices.Equal(*lines, want) {
    		t.Errorf("Trace a, b, c logged %q, want %q", *lines, want)
    	}
    }

    func TestRequireUser(t *testing.T) {
    	tests := []struct {
    		user    string
    		allowed []string
    		want    Response
    		called  bool
    	}{
    		{"ada", []string{"ada", "linus"}, Response{200, "ok:doc"}, true},
    		{"linus", []string{"ada", "linus"}, Response{200, "ok:doc"}, true},
    		{"eve", []string{"ada", "linus"}, Response{403, "forbidden"}, false},
    		{"", []string{"ada"}, Response{403, "forbidden"}, false},
    		{"ada", nil, Response{403, "forbidden"}, false},
    	}
    	for _, tt := range tests {
    		rec := &recorder{}
    		resp := RequireUser(tt.allowed...)(rec.handle)(Request{tt.user, "html", "doc"})
    		if resp != tt.want {
    			t.Errorf("RequireUser(%q) with user %q: response = %+v, want %+v", tt.allowed, tt.user, resp, tt.want)
    		}
    		if called := len(rec.got) > 0; called != tt.called {
    			t.Errorf("RequireUser(%q) with user %q: handler called = %v, want %v", tt.allowed, tt.user, called, tt.called)
    		}
    	}
    }

    func TestMaxBody(t *testing.T) {
    	tests := []struct {
    		body string
    		max  int
    		want Response
    	}{
    		{"12345", 5, Response{200, "ok:12345"}},
    		{"123456", 5, Response{413, "too large"}},
    		{"", 0, Response{200, "ok:"}},
    		{"x", 0, Response{413, "too large"}},
    	}
    	for _, tt := range tests {
    		rec := &recorder{}
    		resp := MaxBody(tt.max)(rec.handle)(Request{"ada", "html", tt.body})
    		if resp != tt.want {
    			t.Errorf("MaxBody(%d) with a %d-byte body: response = %+v, want %+v", tt.max, len(tt.body), resp, tt.want)
    		}
    		if tt.want.Status != 200 && len(rec.got) > 0 {
    			t.Errorf("MaxBody(%d) rejected a %d-byte body but still called the handler", tt.max, len(tt.body))
    		}
    	}
    }

    // safeCall calls h(req) and turns a panic that escapes h into a test failure.
    func safeCall(t *testing.T, name string, h Handler, req Request) (resp Response) {
    	t.Helper()
    	defer func() {
    		if r := recover(); r != nil {
    			t.Errorf("%s: the panic %q escaped: Recover must catch it and return a 500 response", name, r)
    		}
    	}()
    	return h(req)
    }

    func TestRecover(t *testing.T) {
    	boom := func(req Request) Response { panic("renderer exploded") }
    	resp := safeCall(t, "Recover()(boom)", Recover()(boom), Request{"ada", "pdf", ""})
    	if resp != (Response{500, "internal error"}) {
    		t.Errorf("Recover around a panicking handler returned %+v, want {Status:500 Body:internal error}", resp)
    	}
    	rec := &recorder{}
    	resp = Recover()(rec.handle)(Request{"ada", "pdf", "fine"})
    	if resp != (Response{200, "ok:fine"}) {
    		t.Errorf("Recover around a working handler returned %+v, want {Status:200 Body:ok:fine}", resp)
    	}
    }

    func TestOnlyFor(t *testing.T) {
    	lines, logf := logger()
    	rec := &recorder{}
    	h := Chain(rec.handle, OnlyFor("pdf", Trace("pdf", logf)), OnlyFor("pdf", MaxBody(3)))
    	if resp := h(Request{"ada", "html", "long html body"}); resp != (Response{200, "ok:long html body"}) {
    		t.Errorf("OnlyFor(\"pdf\", MaxBody(3)) on an html request: response = %+v, want {Status:200 Body:ok:long html body}", resp)
    	}
    	if len(*lines) != 0 {
    		t.Errorf("OnlyFor(\"pdf\", Trace) logged %q for an html request, want nothing", *lines)
    	}
    	if resp := h(Request{"ada", "pdf", "long"}); resp != (Response{413, "too large"}) {
    		t.Errorf("OnlyFor(\"pdf\", MaxBody(3)) on a 4-byte pdf request: response = %+v, want {Status:413 Body:too large}", resp)
    	}
    	if want := []string{"pdf>", "pdf<413"}; !slices.Equal(*lines, want) {
    		t.Errorf("OnlyFor(\"pdf\", Trace) logged %q for a pdf request, want %q", *lines, want)
    	}
    	if len(rec.got) != 1 {
    		t.Errorf("the handler received %d requests, want 1 (the html one)", len(rec.got))
    	}
    }

    func TestMiddlewareIsBuiltOnce(t *testing.T) {
    	builds := 0
    	counting := func(next Handler) Handler {
    		builds++
    		return next
    	}
    	rec := &recorder{}
    	h := Chain(rec.handle, OnlyFor("pdf", counting), Recover(), OnlyFor("html", counting))
    	for range 5 {
    		h(Request{"ada", "pdf", "a"})
    		h(Request{"ada", "html", "b"})
    	}
    	if builds != 2 {
    		t.Errorf("building a chain with 2 OnlyFor(counting) layers and sending 10 requests built the inner middleware %d times, want 2: call mw(next) once when OnlyFor wraps next, not on every request", builds)
    	}
    }

    func TestFullChainShortCircuits(t *testing.T) {
    	tests := []struct {
    		req     Request
    		want    Response
    		wantLog []string
    		called  bool
    	}{
    		{Request{"ada", "html", "hello"}, Response{200, "ok:hello"}, []string{"outer>", "inner>", "inner<200", "outer<200"}, true},
    		{Request{"eve", "html", "hello"}, Response{403, "forbidden"}, []string{"outer>", "outer<403"}, false},
    		{Request{"ada", "pdf", "this body is too long"}, Response{413, "too large"}, []string{"outer>", "outer<413"}, false},
    		{Request{"ada", "html", "this body is too long"}, Response{200, "ok:this body is too long"}, []string{"outer>", "inner>", "inner<200", "outer<200"}, true},
    	}
    	for _, tt := range tests {
    		lines, logf := logger()
    		rec := &recorder{}
    		h := Chain(rec.handle,
    			Recover(),
    			Trace("outer", logf),
    			RequireUser("ada", "linus"),
    			OnlyFor("pdf", MaxBody(10)),
    			Trace("inner", logf),
    		)
    		resp := h(tt.req)
    		if resp != tt.want {
    			t.Errorf("full chain, request %+v: response = %+v, want %+v", tt.req, resp, tt.want)
    		}
    		if !slices.Equal(*lines, tt.wantLog) {
    			t.Errorf("full chain, request %+v: log = %q, want %q", tt.req, *lines, tt.wantLog)
    		}
    		if called := len(rec.got) > 0; called != tt.called {
    			t.Errorf("full chain, request %+v: handler called = %v, want %v", tt.req, called, tt.called)
    		}
    	}
    }

    func TestFullChainRecoversPanics(t *testing.T) {
    	lines, logf := logger()
    	panicky := func(req Request) Response { panic(fmt.Sprint("cannot render ", req.Format)) }
    	h := Chain(panicky, Recover(), Trace("outer", logf), RequireUser("ada"), Trace("inner", logf))
    	resp := safeCall(t, "full chain", h, Request{"ada", "pdf", ""})
    	if resp != (Response{500, "internal error"}) {
    		t.Errorf("full chain around a panicking handler: response = %+v, want {Status:500 Body:internal error}", resp)
    	}
    	if want := []string{"outer>", "inner>"}; !slices.Equal(*lines, want) {
    		t.Errorf("full chain around a panicking handler: log = %q, want %q (a panic skips the \"<\" lines)", *lines, want)
    	}
    }
---

Doc2Doc's conversion service handles requests with a `Handler`, and wraps it
in layers of **middleware** for logging, access control, size limits and crash
protection. Each layer can pass a request on to the next one, or answer it
itself and **short-circuit** everything inside it.

`Chain` is already written: `Chain(h, A, B, C)` makes `A` the outermost layer,
so a request passes through `A`, `B`, `C`, then `h`, and the response travels
back out through `C`, `B`, `A`.

Write five middleware constructors:

| Middleware | Behaviour |
|---|---|
| `Trace(name, logf)` | calls `logf(name + ">")`, then `next`, then `logf` again with the name, `<` and the response's status, e.g. `"auth<403"`; returns `next`'s response |
| `RequireUser(allowed...)` | if `req.User` isn't in `allowed`, returns `{403, "forbidden"}` without calling `next` |
| `MaxBody(n)` | if `len(req.Body) > n`, returns `{413, "too large"}` without calling `next` |
| `Recover()` | if `next` panics, returns `{500, "internal error"}` instead of crashing |
| `OnlyFor(format, mw)` | runs requests whose `Format` is `format` through `mw`; all others go straight to `next`, skipping `mw` |

`OnlyFor` must call `mw(next)` **once**, when it wraps `next`, not once per
request: building a middleware layer can be expensive.

## Example

```go
h := Chain(convert,
	Recover(),
	Trace("outer", logf),
	RequireUser("ada", "linus"),
	OnlyFor("pdf", MaxBody(10)),
	Trace("inner", logf),
)
h(Request{"ada", "html", "hello"})                 // {200 HELLO}; logs outer> inner> inner<200 outer<200
h(Request{"eve", "html", "hello"})                 // {403 forbidden}; logs outer> outer<403
h(Request{"ada", "pdf", "this body is too long"}) // {413 too large}; logs outer> outer<413
h(Request{"ada", "html", "this body is too long"}) // {200 ...}: the size limit is for PDFs only
h(Request{"linus", "pdf", ""})                    // convert panics: {500 internal error}; logs outer> inner>
```

The last request shows why `Recover` goes outermost: the panic unwinds
straight through both `Trace` layers (so neither logs its `<` line) and is
caught at the top.

## Constraints

- A short-circuiting layer must not call `next` at all, so the handler and every
  layer inside it never see the request.
- No package-level state: the tests build many chains and compare their logs.
