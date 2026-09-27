---
title: Handler vs HandlerFunc
quiz:
  - question: What is `http.HandlerFunc`?
    options:
      - text: An interface that every handler function implements automatically
      - text: A function type with a `ServeHTTP` method that calls the function itself
        correct: true
      - text: A built-in keyword for declaring handlers
      - text: A struct that wraps a handler and adds logging
    explanation: |
      `type HandlerFunc func(ResponseWriter, *Request)`, plus a `ServeHTTP` method that
      just calls `f(w, r)`. Converting a function to `HandlerFunc` makes it satisfy the
      `http.Handler` interface.
  - question: |
      Why won't this compile?

      ```go
      func home(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "hi") }

      mux.Handle("GET /{$}", home)
      ```
    options:
      - text: '`home` must return an `error`'
      - text: '`Handle` wants an `http.Handler`, and a plain function has no `ServeHTTP` method'
        correct: true
      - text: The pattern `"GET /{$}"` is invalid
      - text: Handlers must be methods, not functions
    explanation: |
      `mux.Handle` takes an `http.Handler`. Use `mux.HandleFunc("GET /{$}", home)`, or
      convert it with `mux.Handle("GET /{$}", http.HandlerFunc(home))`. Both do the same.
---

You've been writing handlers since the first lesson. Now it's time to see what they
really are, because middleware (the rest of this chapter) is built entirely on it.

## The interface

Everything in `net/http` revolves around one tiny interface:

```go
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
```

The server calls `ServeHTTP` for every request. `*http.ServeMux` is a `Handler` (its
`ServeHTTP` picks a route and calls *that* handler's `ServeHTTP`). `http.FileServerFS`
returns a `Handler`. `http.StripPrefix` takes a `Handler` and returns another one.

Any type can be a handler. Here's a struct that counts down squeaks until a launch
party:

```go
type countdown struct {
	left int
}

func (c *countdown) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%d squeaks until the cheese party\n", c.left)
}

mux.Handle("GET /api/party", &countdown{left: 99})
```

## Functions as handlers

Writing a struct for every endpoint would be tedious, so `net/http` defines a
function type with a method:

```go
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
	f(w, r)
}
```

That's the whole trick. `http.HandlerFunc(home)` doesn't *call* `home`. It
**converts** it to a type that has a `ServeHTTP` method, so it satisfies `Handler`.
`mux.HandleFunc(pattern, fn)` is a shortcut that does the conversion for you:

```go
mux.HandleFunc("GET /{$}", home)                // shortcut
mux.Handle("GET /{$}", http.HandlerFunc(home)) // same thing, spelled out
```

This "function type with a method" pattern is a Go classic: a function type gets a
method, so any function of the right shape can satisfy an interface.

## Handlers that need dependencies

Real handlers need things: a database, a config, a logger. Global variables work, but
they make testing painful. The idiomatic way is to hang handlers off a struct as
**methods**:

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

type apiConfig struct {
	appName  string
	maxChars int
}

func (cfg *apiConfig) handleInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s: squeaks can be %d characters long\n", cfg.appName, cfg.maxChars)
}

func main() {
	cfg := &apiConfig{appName: "Squeak", maxChars: 140}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/info", cfg.handleInfo) // a method value

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/info", nil))
	fmt.Print(rec.Body.String())
}
```

```
Squeak: squeaks can be 140 characters long
```

`cfg.handleInfo` is a **method value**: a function with `cfg` already bound to it, so
its signature is exactly `func(http.ResponseWriter, *http.Request)`. From here on,
Squeak's handlers are methods on an `apiConfig` that holds the store, secrets and
counters. In tests you build an `apiConfig` with fakes and call the same methods.

## Closures work too

Another option is a function that *returns* a handler and closes over what it needs:

```go
func handleInfo(appName string, maxChars int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s: squeaks can be %d characters long\n", appName, maxChars)
	}
}

mux.Handle("GET /api/info", handleInfo("Squeak", 140))
```

Both styles are common. Methods keep related handlers together, and closures make
each handler's dependencies explicit. Pick one per project and stay consistent.

## Handlers run concurrently

A final reminder. The same handler value serves every request, on many goroutines at
once. The `countdown` above only *reads* `c.left`, which is fine. The moment a handler
*writes* to shared fields, you need a mutex or an atomic. You'll hit exactly that at the
end of this chapter.
