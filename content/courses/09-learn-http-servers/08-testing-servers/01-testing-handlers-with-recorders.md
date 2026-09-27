---
title: Testing Handlers with Recorders
quiz:
  - question: |
      A handler under test reads `r.PathValue("id")`, and the test calls it directly:

      ```go
      req := httptest.NewRequest("GET", "/api/squeaks/42", nil)
      cfg.handleGetSqueak(rec, req)
      ```

      What does `PathValue("id")` return inside the handler?
    options:
      - text: '`"42"`, parsed from the URL'
      - text: '`""`, because no mux matched a pattern, so nothing set the path values'
        correct: true
      - text: It panics
      - text: '`"/api/squeaks/42"`'
    explanation: |
      Path values are filled in by `ServeMux` when it matches a pattern. Calling the handler
      directly skips the mux, so either call `req.SetPathValue("id", "42")` first, or send
      the request through the router.
  - question: What's the difference between `httptest.NewRequest` and `http.NewRequest`?
    options:
      - text: They're identical
      - text: '`httptest.NewRequest` builds a *server-side* request for passing straight to a handler, and panics on bad input instead of returning an error'
        correct: true
      - text: '`httptest.NewRequest` actually sends the request over the network'
      - text: '`http.NewRequest` can only be used in tests'
    explanation: |
      `http.NewRequest` makes a *client* request for `http.Client.Do`. `httptest.NewRequest`
      makes a request as a server would receive it (with `RemoteAddr`, `Host` and so on
      filled in), and it panics on errors, which keeps test code short.
---

You've been using `httptest.NewRecorder` since the first lesson to *show* handlers
working. In this chapter you'll use it for what it was built for: automated tests. You
know `go test`, table tests and subtests from the testing course. Here's how they apply
to servers.

## The three ingredients

```go
req := httptest.NewRequest("POST", "/api/squeaks", strings.NewReader(`{"body":"hi"}`))
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)
```

- **`httptest.NewRequest`** builds an incoming request, exactly as the server would hand
  it to a handler. It fills in `RemoteAddr` (`192.0.2.1:1234`) and `Host`
  (`example.com`), and it panics on invalid input, so tests don't need error checks.
- **`httptest.NewRecorder`** is an `http.ResponseWriter` that records everything:
  `rec.Code`, `rec.Header()`, `rec.Body` (a `*bytes.Buffer`).
- **`ServeHTTP`** calls the handler. No network, no goroutines, no ports. It's just a
  function call, so tests are fast and deterministic.

## A complete handler test

Here's a test file for Squeak's create endpoint. Note the helpers: they keep each test
focused on *what* it checks.

```go
package main

import (
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestConfig(t *testing.T) *apiConfig {
	t.Helper()
	return &apiConfig{squeaks: NewMemoryStore()}
}

// decodeBody fails the test if the response isn't the JSON we expect.
func decodeBody[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("response body %q is not valid JSON: %v", rec.Body.String(), err)
	}
	return v
}

func TestCreateSqueak(t *testing.T) {
	cfg := newTestConfig(t)
	req := httptest.NewRequest("POST", "/api/squeaks", strings.NewReader(`{"body":"beware the cat"}`))
	rec := httptest.NewRecorder()

	cfg.handleCreateSqueak(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", rec.Code, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	got := decodeBody[Squeak](t, rec)
	if got.Body != "beware the ****" {
		t.Errorf("body = %q, want the cleaned %q", got.Body, "beware the ****")
	}
}

func TestCreateSqueakRejectsEmptyBody(t *testing.T) {
	cfg := newTestConfig(t)
	req := httptest.NewRequest("POST", "/api/squeaks", strings.NewReader(`{"body":"   "}`))
	rec := httptest.NewRecorder()

	cfg.handleCreateSqueak(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if e := decodeBody[map[string]string](t, rec); e["error"] == "" {
		t.Errorf("error response %q has no \"error\" field", rec.Body)
	}
}
```

Some habits worth copying:

- **Include the body in failure messages.** "status = 500, want 201" is a mystery.
  "status = 500, want 201; body: {"error":"couldn't create squeak"}" points at the cause.
- **`t.Fatalf` when later checks make no sense** (no point decoding a squeak from a 400),
  and `t.Errorf` when they're independent.
- **Generic helpers** like `decodeBody[T]` cut the noise, and `t.Helper()` makes
  failures point at the caller's line.
- **Test the failure paths.** Most bugs live in the 4xx branches nobody clicked through.

## Path values and context

Calling a handler directly skips the mux and the middleware, so anything *they* would
have set up is missing:

```go
req := httptest.NewRequest("DELETE", "/api/squeaks/"+sq.ID.String(), nil)
req.SetPathValue("id", sq.ID.String()) // normally the mux does this

ctx := context.WithValue(req.Context(), userIDKey{}, pip) // normally requireAuth does this
req = req.WithContext(ctx)
```

Both are fine for focused unit tests. The alternative is to send the request through
your full router, which you'll do with table tests shortly.

## Testing middleware

Middleware is a function from handler to handler, so test it by wrapping a tiny stub
handler and checking what reaches it and what comes out:

```go
func TestRequestIDMiddleware(t *testing.T) {
	var seen string
	stub := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = requestIDFrom(r.Context())
	})
	rec := httptest.NewRecorder()
	withRequestID(stub).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	if seen == "" {
		t.Fatal("handler saw no request ID in its context")
	}
	if got := rec.Header().Get("X-Request-ID"); got != seen {
		t.Errorf("X-Request-ID header = %q, want the context's ID %q", got, seen)
	}
}
```

You've already met this style: the tests behind this course's middleware exercises
look just like it.

## rec.Result() for client-style checks

`rec.Result()` returns an `*http.Response`, as a client would see it. It's handy when
you already have helpers that take responses, or want `resp.Cookies()`. For most tests,
`rec.Code`, `rec.Header()` and `rec.Body` are all you need.
