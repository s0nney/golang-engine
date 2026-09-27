---
title: Chaining Middleware
quiz:
  - question: |
      With `Chain(h, A, B, C)` as defined in this lesson, in what order do the middlewares see an incoming request?
    options:
      - text: C, B, A
      - text: The order is random
      - text: A, B, C
        correct: true
      - text: Only A runs
    explanation: |
      `Chain` applies the list in reverse, so `A` ends up as the outermost layer.
      Requests enter through `A`, then `B`, then `C`, then reach the handler.
      Responses unwind in the opposite order.
  - question: Why should a panic-recovery middleware usually be the *outermost* layer?
    options:
      - text: Because `recover` only works in `main`
      - text: So it can catch panics from every other middleware and the handler inside it
        correct: true
      - text: Because it's the slowest middleware
    explanation: |
      A deferred `recover` only catches panics from code it wraps. Put it on the
      outside and it protects everything else in the chain.
---

Nesting middleware by hand reads inside-out and is easy to get wrong:

```go
h := recoverPanics(logRequests(requireToken(token)(convertHandler)))
```

Since every middleware has the same type, you can store them in a slice and apply
them in a loop. That's composition again, just for handlers.

## A Chain helper

```go
type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for _, mw := range slices.Backward(mws) {
		h = mw(h)
	}
	return h
}
```

`slices.Backward` walks the slice from the end, so the *first* middleware in the list
becomes the *outermost* wrapper. That way the list reads in the order requests flow
through it.

## The onion

Each middleware wraps the next, like layers of an onion. A request travels in
through every layer, reaches the handler, and the response travels back out:

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
)

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for _, mw := range slices.Backward(mws) {
		h = mw(h)
	}
	return h
}

func trace(name string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("enter", name)
			next.ServeHTTP(w, r)
			fmt.Println("leave", name)
		})
	}
}

func recoverPanics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("recovered:", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func convert(w http.ResponseWriter, r *http.Request) {
	fmt.Println("  handler runs")
	if r.URL.Query().Get("fmt") == "pdf" {
		panic("PDF output not implemented")
	}
	fmt.Fprint(w, "ok")
}

func main() {
	h := Chain(http.HandlerFunc(convert),
		recoverPanics,
		trace("logging"),
		trace("auth"),
	)

	for _, url := range []string{"/convert?fmt=html", "/convert?fmt=pdf"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", url, nil))
		fmt.Println("status", rec.Code)
	}
}
```

```text
enter logging
enter auth
  handler runs
leave auth
leave logging
status 200
enter logging
enter auth
  handler runs
recovered: PDF output not implemented
status 500
```

In the second request the handler panicked. The panic skipped the "leave" lines
entirely (the stack unwound straight past them) until `recoverPanics`, the outermost
layer, caught it and turned it into a clean 500 response instead of a dropped
connection.

## Ordering tips

- **Recovery** goes outermost, so it protects everything.
- **Logging** goes near the outside, so it sees every request, including rejected ones.
- **Authentication** goes before anything expensive, so bad requests stop early.
- **Per-route** middleware (like "only admins may delete") wraps just that route's
  handler, not the whole mux.

## The decorator pattern, Go style

Python gives decorators special syntax. Go gives you something arguably clearer:
ordinary functions with matching types, combined with ordinary loops. There's no
magic. You can read `Chain` in ten seconds and see exactly what order things happen
in. That's the Go philosophy applied to a classic functional idea.
