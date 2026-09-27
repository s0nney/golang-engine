---
title: Request IDs and Context
quiz:
  - question: Why should a context key be an unexported custom type, like `type ctxKey struct{}`, rather than a string?
    options:
      - text: Strings can't be used as context keys
      - text: Custom types are faster to look up
      - text: A string key like `"requestID"` could collide with another package using the same string; a private type can't
        correct: true
      - text: The context package requires keys to be structs
    explanation: |
      Context values are matched by `==` on the key, including its type. Two packages that
      both use the string `"requestID"` would overwrite each other. A key of an unexported
      type from your package can't be created (or clobbered) by anyone else.
  - question: |
      What's wrong with this middleware?

      ```go
      func withRequestID(next http.Handler) http.Handler {
      	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      		ctx := context.WithValue(r.Context(), requestIDKey{}, rand.Text())
      		r.WithContext(ctx)
      		next.ServeHTTP(w, r)
      	})
      }
      ```
    options:
      - text: '`rand.Text` is not safe for concurrent use'
      - text: '`r.WithContext` returns a new request; the result is thrown away, so `next` never sees the ID'
        correct: true
      - text: Middleware can't change the context
      - text: '`context.WithValue` must be given a string key'
    explanation: |
      Requests are treated as immutable. `WithContext` returns a *shallow copy* with the new
      context, and the original `r` is unchanged. It should be
      `next.ServeHTTP(w, r.WithContext(ctx))`.
---

When a mouse reports "my squeak failed at 14:02", you want to find the exact log lines for
that request among thousands. The standard trick is a **request ID**: a unique string
generated for every request, attached to every log line, and sent back to the client in
a response header so they can quote it in bug reports.

A request ID needs to travel from the middleware that creates it to the handlers and
helpers deep down the call stack. That's what the request's **context** is for.

## Every request carries a context

`r.Context()` returns a `context.Context` that the server creates for each request. You
know it from the concurrency course: it carries cancellation (it's cancelled when the
client disconnects or the handler returns) and it can carry **request-scoped values**.

To add a value you create a derived context and a derived request:

```go
ctx := context.WithValue(r.Context(), requestIDKey{}, id)
r = r.WithContext(ctx) // a shallow copy of r with the new context
```

## A private key type

Context values are looked up by key. Use an **unexported** type so no other package can
collide with your key, and wrap access in small helpers so the rest of your code never
touches the key directly:

```go
type requestIDKey struct{}

func requestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id // "" if there isn't one
}
```

The comma-ok type assertion means a missing value gives `""` rather than a panic.

## The middleware

```go
package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"net/http/httptest"
)

type requestIDKey struct{}

func requestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" || len(id) > 64 {
			id = rand.Text() // 26 random base32 characters
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func handleSqueak(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "handled request %s", requestIDFrom(r.Context()))
}

func main() {
	h := withRequestID(http.HandlerFunc(handleSqueak))

	// A client (or a load balancer in front of us) that already sent an ID.
	req := httptest.NewRequest("GET", "/api/squeaks", nil)
	req.Header.Set("X-Request-ID", "pip-debug-7")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	fmt.Println(rec.Header().Get("X-Request-ID"), "|", rec.Body.String())

	// No ID sent: the middleware makes one up.
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/squeaks", nil))
	fmt.Println(len(rec.Header().Get("X-Request-ID")), "random characters")
}
```

```
pip-debug-7 | handled request pip-debug-7
26 random characters
```

Some choices in there:

- **Reuse an incoming ID.** If a load balancer or another service already assigned one,
  keeping it lets you follow a request across several services. But the header is
  client input, so cap its length (and in stricter setups, only trust it from your
  own proxies).
- **`crypto/rand.Text()`** returns a random string with at least 128 bits of
  randomness, which is plenty for IDs. (In the storage chapter you'll meet `uuid.NewV7`,
  which works too.)
- **Set the response header before calling `next`.** Headers must be set before the
  handler writes the status. Afterwards it's too late.

## What belongs in context, and what doesn't

Context values are for **request-scoped** data that crosses API boundaries: request IDs,
the authenticated user, trace IDs. They are *not* a way to smuggle a database handle or
config into your handlers. Those belong in your `apiConfig` struct, where the compiler
can see them. A good rule: if a function can't do its job without a value, pass it as a
parameter.

## Using the ID in logs

Now the logging middleware from the last lesson can print `requestIDFrom(r.Context())`
on every line, as long as it runs *inside* `withRequestID`. Order matters when you stack
middleware, and that's the next lesson.
