---
title: HTTP Middleware
quiz:
  - question: What is the usual signature of a middleware in Go's `net/http` world?
    options:
      - text: '`func(http.Handler) http.Handler`'
        correct: true
      - text: '`func(http.ResponseWriter, *http.Request)`'
      - text: '`func(*http.Request) error`'
      - text: '`func(http.Handler)`'
    explanation: |
      Middleware takes the next handler and returns a new handler that wraps it.
      Because the input and output types match, middleware can be stacked.
  - question: |
      How can a plain function like `func(w http.ResponseWriter, r *http.Request)` be used where an `http.Handler` interface is required?
    options:
      - text: Functions automatically implement every interface
      - text: Wrap it in a struct with a `Handler` field
      - text: Convert it with `http.HandlerFunc(f)`, a function type whose `ServeHTTP` method calls the function
        correct: true
    explanation: |
      `http.HandlerFunc` is a named function type with a `ServeHTTP` method that
      just calls itself. Converting your function to it makes it satisfy
      `http.Handler`. It's the adapter pattern from the Function Types lesson.
  - question: 'In `requireToken` from this lesson, what happens when the token is wrong?'
    options:
      - text: The middleware writes a 401 response and returns without calling `next`
        correct: true
      - text: The request still reaches the wrapped handler, which is expected to check again
      - text: The server crashes
    explanation: |
      Middleware decides whether to call `next.ServeHTTP` at all. Returning early
      short-circuits the chain, so the protected handler never runs.
---

Wrapping functions is nice, but the place you'll use it most in real Go code is
**HTTP middleware**. It's the canonical Go "decorator".

Doc2Doc is growing a web API: `POST /convert` takes a document and returns HTML.
Every endpoint needs logging, authentication and panic recovery. You don't want that
code copied into every handler.

## Handlers are functions in disguise

Everything in `net/http` revolves around one interface:

```go
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
```

And one function type that adapts plain functions to it:

```go
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) { f(w, r) }
```

That's the function-type-with-a-method trick from earlier in the course, used by the
standard library itself.

## Middleware is a handler wrapper

A middleware takes a handler and returns a new handler:

```go
func(next http.Handler) http.Handler
```

Here are two for Doc2Doc, plus a test run using `httptest`, so no real server or
network is needed:

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("->", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		fmt.Println("<-", r.Method, r.URL.Path)
	})
}

func requireToken(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer "+token {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return // don't call next
			}
			next.ServeHTTP(w, r)
		})
	}
}

func convert(w http.ResponseWriter, r *http.Request) {
	fmt.Println("   converting")
	fmt.Fprint(w, "<h1>Hello</h1>")
}

func main() {
	var h http.Handler = http.HandlerFunc(convert)
	h = requireToken("s3cret")(h)
	h = logRequests(h)

	for _, auth := range []string{"Bearer s3cret", "Bearer nope"} {
		req := httptest.NewRequest("POST", "/convert", strings.NewReader("# Hello"))
		req.Header.Set("Authorization", auth)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)
		fmt.Println(rec.Code, strings.TrimSpace(rec.Body.String()))
	}
}
```

```text
-> POST /convert
   converting
<- POST /convert
200 <h1>Hello</h1>
-> POST /convert
<- POST /convert
401 unauthorized
```

Notice what happened on the second request. `logRequests` ran (it's the outer
layer), but `requireToken` stopped the request and `convert` never ran.

## Two kinds of middleware

- **Simple middleware** like `logRequests` has the signature
  `func(http.Handler) http.Handler` directly.
- **Configurable middleware** like `requireToken` needs settings, so it's a function
  that *returns* a middleware. That's currying from the last chapter:
  `requireToken("s3cret")(h)`.

## Hooking it into a server

In a real program you'd wrap the handler and register it with a `ServeMux`:

```go
mux := http.NewServeMux()
mux.Handle("POST /convert", logRequests(requireToken(token)(http.HandlerFunc(convert))))
log.Fatal(http.ListenAndServe(":8080", mux))
```

That nesting gets ugly fast. The next lesson fixes it with a `Chain` helper.
