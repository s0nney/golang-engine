---
title: Writing Middleware
quiz:
  - question: What's the usual signature of middleware in Go?
    options:
      - text: '`func(w http.ResponseWriter, r *http.Request) error`'
      - text: '`func(next http.Handler) http.Handler`'
        correct: true
      - text: '`func(mux *http.ServeMux)`'
      - text: '`func(r *http.Request) *http.Request`'
    explanation: |
      Middleware takes the handler it wraps and returns a new handler. Because input and
      output have the same type, middleware can wrap other middleware in any order.
  - question: |
      This logging middleware always logs status `200`, even for 404s. Why?

      ```go
      func logRequests(next http.Handler) http.Handler {
      	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      		next.ServeHTTP(w, r)
      		log.Printf("%s %s %d", r.Method, r.URL.Path, http.StatusOK)
      	})
      }
      ```
    options:
      - text: '`next.ServeHTTP` resets the status to 200 when it returns'
      - text: The status is hard-coded; there's no way to read it back from an `http.ResponseWriter`, so you have to wrap the writer to record it
        correct: true
      - text: The log line runs before the handler
      - text: 404s never reach middleware
    explanation: |
      `http.ResponseWriter` has no "what status did I send?" method. To log it, the middleware
      passes the handler a wrapper whose `WriteHeader` remembers the code before
      forwarding it.
---

Squeak's handlers are starting to need the same things: log every request, attach a
request ID, recover from panics, count hits. Copying that into every handler would be a
mess. **Middleware** is code that runs *around* a handler, written once and applied to
many. You built the idea from scratch in the functional programming course's
[decorators and middleware](/courses/learn-functional-programming/decorators-and-middleware/http-middleware)
chapter. Now it's time to use it on a real server.

## The shape

Middleware is a function that takes a handler and returns a handler:

```go
func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// before: runs on the way in
		next.ServeHTTP(w, r)
		// after: runs on the way out
	})
}
```

Think of it as a layer of an onion. The request passes through the outer layers
before it reaches your handler, and the response passes back out through them.
Middleware can:

- **inspect or modify** the request before passing it on (read headers, add context),
- **short-circuit**, answering itself and *not* calling `next` (for example a 401 for a
  missing token),
- **inspect or decorate** the response by wrapping `w`.

## A logging middleware

The obvious first middleware logs each request. The method and path are easy. The
**status code** is the tricky bit: the handler writes it into `w`, and
`http.ResponseWriter` has no getter. The fix is to wrap the writer in your own type
that remembers what passed through:

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

// statusRecorder wraps a ResponseWriter and remembers the status and size.
type statusRecorder struct {
	http.ResponseWriter // embedded: every other method passes straight through
	status              int
	bytes               int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK // a Write without WriteHeader means 200
	}
	n, err := s.ResponseWriter.Write(b)
	s.bytes += n
	return n, err
}

// Unwrap lets http.ResponseController reach the real writer underneath.
func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }

func logRequests(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(rec, r)

		if rec.status == 0 {
			rec.status = http.StatusOK // the handler wrote nothing at all
		}
		logger.Printf("%s %s -> %d (%d bytes) in %v",
			r.Method, r.URL.Path, rec.status, rec.bytes, time.Since(start).Round(time.Second))
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/squeaks", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"body":"first squeak!"}]`)
	})

	logger := log.New(os.Stdout, "squeak: ", 0)
	handler := logRequests(logger, mux) // wrap the whole mux

	for _, path := range []string{"/api/squeaks", "/api/cheese"} {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", path, nil))
	}
}
```

```
squeak: GET /api/squeaks -> 200 (26 bytes) in 0s
squeak: GET /api/cheese -> 404 (19 bytes) in 0s
```

(Real code would round to the millisecond or log the raw duration. Rounding to seconds
just keeps the output above predictable.)

A few details worth copying:

- **Embedding** `http.ResponseWriter` means the wrapper already has `Header()`, `Write`
  and `WriteHeader`. You only override the ones you care about.
- A handler that never calls `WriteHeader` implicitly sends 200, so the recorder
  defaults to 200.
- **`Unwrap`** matters. Embedding hides optional interfaces such as `http.Flusher`.
  `http.NewResponseController(w)` looks for an `Unwrap` method to find the original
  writer, so flushing and per-request deadlines keep working through your wrapper.

Because `logRequests` also takes a `*log.Logger`, it isn't quite the plain
`func(http.Handler) http.Handler` shape. That's fine: when middleware needs
configuration, a common trick is a function that returns middleware:

```go
func logRequests(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler { /* ... */ }
}
```

## Wrap the mux, or wrap a route

You can wrap the entire mux, as above, so every request gets logged, including 404s and
405s the mux generates itself. Or you can wrap single routes:

```go
mux.Handle("GET /admin/metrics", requireAdmin(http.HandlerFunc(cfg.handleMetrics)))
```

Logging, request IDs and panic recovery usually go around everything. Auth checks
usually go around specific routes.

## Middleware gotcha: calling next twice, or not at all

Forget to call `next.ServeHTTP` and the request silently stops there: the client gets
an empty 200. Call it twice and the handler runs twice, probably writing two bodies.
When middleware short-circuits on purpose, it should write a full response (status and
body) and `return` straight away.
