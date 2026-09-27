---
title: 'Practice: Logging Middleware'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    )

    type Middleware func(http.Handler) http.Handler

    // statusRecorder wraps a ResponseWriter and remembers the status code.
    // Handlers that never call WriteHeader send 200, so start it at 200.
    type statusRecorder struct {
    	http.ResponseWriter
    	status int
    }

    func (r *statusRecorder) WriteHeader(code int) {
    	r.status = code
    	r.ResponseWriter.WriteHeader(code)
    }

    // logRequests returns a middleware that, after the wrapped handler has
    // run, calls logf with "METHOD PATH -> STATUS", e.g. "GET /convert -> 200".
    func logRequests(logf func(string)) Middleware {
    	return func(next http.Handler) http.Handler {
    		// ?
    		return next
    	}
    }

    // Chain wraps h in every middleware so that the FIRST one in mws is the
    // outermost layer (it sees the request first).
    func Chain(h http.Handler, mws ...Middleware) http.Handler {
    	// ?
    	return h
    }

    func convert(w http.ResponseWriter, r *http.Request) {
    	if r.URL.Query().Get("fmt") == "pdf" {
    		http.Error(w, "pdf not supported", http.StatusNotImplemented)
    		return
    	}
    	fmt.Fprint(w, "<h1>converted</h1>")
    }

    func main() {
    	logf := func(s string) { fmt.Println("log:", s) }
    	h := Chain(http.HandlerFunc(convert), logRequests(logf))

    	for _, url := range []string{"/convert?fmt=html", "/convert?fmt=pdf"} {
    		rec := httptest.NewRecorder()
    		h.ServeHTTP(rec, httptest.NewRequest("POST", url, nil))
    		fmt.Println("status:", rec.Code)
    	}
    }
  solution: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"slices"
    )

    type Middleware func(http.Handler) http.Handler

    // statusRecorder wraps a ResponseWriter and remembers the status code.
    // Handlers that never call WriteHeader send 200, so start it at 200.
    type statusRecorder struct {
    	http.ResponseWriter
    	status int
    }

    func (r *statusRecorder) WriteHeader(code int) {
    	r.status = code
    	r.ResponseWriter.WriteHeader(code)
    }

    // logRequests returns a middleware that, after the wrapped handler has
    // run, calls logf with "METHOD PATH -> STATUS", e.g. "GET /convert -> 200".
    func logRequests(logf func(string)) Middleware {
    	return func(next http.Handler) http.Handler {
    		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
    			next.ServeHTTP(rec, r)
    			logf(fmt.Sprintf("%s %s -> %d", r.Method, r.URL.Path, rec.status))
    		})
    	}
    }

    // Chain wraps h in every middleware so that the FIRST one in mws is the
    // outermost layer (it sees the request first).
    func Chain(h http.Handler, mws ...Middleware) http.Handler {
    	for _, mw := range slices.Backward(mws) {
    		h = mw(h)
    	}
    	return h
    }

    func convert(w http.ResponseWriter, r *http.Request) {
    	if r.URL.Query().Get("fmt") == "pdf" {
    		http.Error(w, "pdf not supported", http.StatusNotImplemented)
    		return
    	}
    	fmt.Fprint(w, "<h1>converted</h1>")
    }

    func main() {
    	logf := func(s string) { fmt.Println("log:", s) }
    	h := Chain(http.HandlerFunc(convert), logRequests(logf))

    	for _, url := range []string{"/convert?fmt=html", "/convert?fmt=pdf"} {
    		rec := httptest.NewRecorder()
    		h.ServeHTTP(rec, httptest.NewRequest("POST", url, nil))
    		fmt.Println("status:", rec.Code)
    	}
    }
  tests: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"slices"
    	"testing"
    )

    func TestLogRequests(t *testing.T) {
    	var logs []string
    	h := logRequests(func(s string) { logs = append(logs, s) })(http.HandlerFunc(convert))

    	rec := httptest.NewRecorder()
    	h.ServeHTTP(rec, httptest.NewRequest("GET", "/convert?fmt=html", nil))
    	if rec.Code != 200 || rec.Body.String() != "<h1>converted</h1>" {
    		t.Errorf("wrapped handler responded %d %q, want 200 %q (call next!)", rec.Code, rec.Body.String(), "<h1>converted</h1>")
    	}

    	rec = httptest.NewRecorder()
    	h.ServeHTTP(rec, httptest.NewRequest("POST", "/convert?fmt=pdf", nil))
    	if rec.Code != http.StatusNotImplemented {
    		t.Errorf("wrapped handler responded %d, want %d", rec.Code, http.StatusNotImplemented)
    	}

    	want := []string{"GET /convert -> 200", "POST /convert -> 501"}
    	if !slices.Equal(logs, want) {
    		t.Errorf("logged %q, want %q", logs, want)
    	}
    }

    func TestLogRequestsLogsAfterHandler(t *testing.T) {
    	var events []string
    	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		events = append(events, "handler")
    		w.WriteHeader(http.StatusAccepted)
    	})
    	h := logRequests(func(s string) { events = append(events, s) })(inner)
    	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("PUT", "/docs/1", nil))
    	want := []string{"handler", "PUT /docs/1 -> 202"}
    	if !slices.Equal(events, want) {
    		t.Errorf("events = %q, want %q (log after the handler runs)", events, want)
    	}
    }

    func tag(name string, events *[]string) Middleware {
    	return func(next http.Handler) http.Handler {
    		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    			*events = append(*events, "enter "+name)
    			next.ServeHTTP(w, r)
    			*events = append(*events, "leave "+name)
    		})
    	}
    }

    func TestChain(t *testing.T) {
    	var events []string
    	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		events = append(events, "handler")
    	})
    	h := Chain(final, tag("A", &events), tag("B", &events), tag("C", &events))
    	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
    	want := []string{"enter A", "enter B", "enter C", "handler", "leave C", "leave B", "leave A"}
    	if !slices.Equal(events, want) {
    		t.Errorf("Chain order:\n got %q\nwant %q", events, want)
    	}

    	events = nil
    	Chain(final).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
    	if fmt.Sprint(events) != "[handler]" {
    		t.Errorf("Chain with no middleware ran %q, want just the handler", events)
    	}
    }
---

Doc2Doc's web API needs an access log. Every request should produce one line like
`POST /convert -> 200` **after** the handler has finished, so the status is known.

## Your task

1. Finish `logRequests(logf)`. It returns a `Middleware` that:
   - wraps the `ResponseWriter` in the provided `statusRecorder`, starting its `status`
     at `http.StatusOK`,
   - calls the next handler with the recorder,
   - then calls `logf` with `"METHOD PATH -> STATUS"`, using `r.URL.Path` (no query
     string).
2. Finish `Chain(h, mws...)` so the **first** middleware in the list is the outermost
   layer and sees each request first.

```go
h := Chain(convertHandler, recoverPanics, logRequests(logf), requireToken(tok))
// request -> recoverPanics -> logRequests -> requireToken -> convertHandler
```

## Tips

- A handler can't tell you what status it sent. `statusRecorder` embeds the real
  `ResponseWriter` and overrides `WriteHeader` so it can remember the code. Handlers
  that never call `WriteHeader` send 200, which is why you start there.
- Wrap your function literal in `http.HandlerFunc(...)` to turn it into an
  `http.Handler`.
- `slices.Backward` walks a slice from the end, which makes `Chain` short.
