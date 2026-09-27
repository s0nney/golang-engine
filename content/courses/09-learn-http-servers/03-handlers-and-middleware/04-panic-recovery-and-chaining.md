---
title: Panic Recovery and Chaining
quiz:
  - question: |
      What does this print?

      ```go
      h := chain(final, tag("A"), tag("B"), tag("C"))
      // chain makes the first middleware the outermost.
      // tag(x) prints x before calling next, and x again after.
      // final prints "handler".
      ```
    options:
      - text: '`A B C handler C B A`'
        correct: true
      - text: '`C B A handler A B C`'
      - text: '`A B C handler A B C`'
      - text: '`handler A B C`'
    explanation: |
      The outermost middleware runs first on the way in and last on the way out, like
      nested function calls: A calls B, B calls C, C calls the handler, and then each
      one returns in reverse order.
  - question: A handler panics *after* it has already written `200 OK` and half of the body. What can your recovery middleware do?
    options:
      - text: Replace the response with a clean `500 Internal Server Error`
      - text: Nothing useful for the client; the status line is already sent, so it can only log the panic (and the connection should be dropped)
        correct: true
      - text: Undo the partial write and retry the handler
      - text: Send a second status line with 500
    explanation: |
      Once the status and headers are on the wire, they can't be taken back. A later
      `WriteHeader` call is ignored with a "superfluous WriteHeader" warning. The best you
      can do is log it. Re-panicking with `http.ErrAbortHandler` makes the server abort the
      response so the client sees it's broken instead of trusting a truncated body.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"log"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    )

    // recoverPanics catches a panic in next. It logs it with
    // log.Printf("panic: %v", value) and answers
    // http.Error(w, "internal server error", http.StatusInternalServerError).
    // A panic with http.ErrAbortHandler must be re-panicked, not recovered.
    func recoverPanics(next http.Handler) http.Handler {
    	// ?
    	return next
    }

    // chain wraps h in the given middleware. The FIRST middleware in the
    // list must be the OUTERMOST, so chain(h, a, b) behaves like a(b(h)).
    func chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
    	// ?
    	return h
    }

    // addHeader returns middleware that appends value to the X-Trace
    // response header before calling next.
    func addHeader(value string) func(http.Handler) http.Handler {
    	return func(next http.Handler) http.Handler {
    		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    			w.Header().Add("X-Trace", value)
    			next.ServeHTTP(w, r)
    		})
    	}
    }

    func main() {
    	log.SetFlags(0) // no timestamps, so the output is easy to read
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/squeaks", func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprint(w, "squeaks!")
    	})
    	mux.HandleFunc("GET /api/cheese", func(w http.ResponseWriter, r *http.Request) {
    		var stock map[string]int
    		stock["gouda"]++ // oops: nil map
    	})

    	h := chain(mux, recoverPanics, addHeader("outer"), addHeader("inner"))

    	for _, path := range []string{"/api/squeaks", "/api/cheese"} {
    		rec := httptest.NewRecorder()
    		func() {
    			defer func() {
    				if v := recover(); v != nil {
    					fmt.Println("the panic escaped all the way to main:", v)
    				}
    			}()
    			h.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
    		}()
    		fmt.Printf("%s -> %d %q trace=%v\n", path, rec.Code,
    			strings.TrimSpace(rec.Body.String()), rec.Header().Values("X-Trace"))
    	}
    }
  solution: |
    package main

    import (
    	"fmt"
    	"log"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    )

    func recoverPanics(next http.Handler) http.Handler {
    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		defer func() {
    			v := recover()
    			if v == nil {
    				return
    			}
    			if v == http.ErrAbortHandler {
    				panic(v)
    			}
    			log.Printf("panic: %v", v)
    			http.Error(w, "internal server error", http.StatusInternalServerError)
    		}()
    		next.ServeHTTP(w, r)
    	})
    }

    func chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
    	for i := len(mws) - 1; i >= 0; i-- {
    		h = mws[i](h)
    	}
    	return h
    }

    func addHeader(value string) func(http.Handler) http.Handler {
    	return func(next http.Handler) http.Handler {
    		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    			w.Header().Add("X-Trace", value)
    			next.ServeHTTP(w, r)
    		})
    	}
    }

    func main() {
    	log.SetFlags(0) // no timestamps, so the output is easy to read
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/squeaks", func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprint(w, "squeaks!")
    	})
    	mux.HandleFunc("GET /api/cheese", func(w http.ResponseWriter, r *http.Request) {
    		var stock map[string]int
    		stock["gouda"]++ // oops: nil map
    	})

    	h := chain(mux, recoverPanics, addHeader("outer"), addHeader("inner"))

    	for _, path := range []string{"/api/squeaks", "/api/cheese"} {
    		rec := httptest.NewRecorder()
    		func() {
    			defer func() {
    				if v := recover(); v != nil {
    					fmt.Println("the panic escaped all the way to main:", v)
    				}
    			}()
    			h.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
    		}()
    		fmt.Printf("%s -> %d %q trace=%v\n", path, rec.Code,
    			strings.TrimSpace(rec.Body.String()), rec.Header().Values("X-Trace"))
    	}
    }
  tests: |
    package main

    import (
    	"bytes"
    	"log"
    	"net/http"
    	"net/http/httptest"
    	"os"
    	"slices"
    	"strings"
    	"testing"
    )

    func captureLog(t *testing.T) *bytes.Buffer {
    	t.Helper()
    	var buf bytes.Buffer
    	log.SetOutput(&buf)
    	t.Cleanup(func() { log.SetOutput(os.Stderr) })
    	return &buf
    }

    func serve(t *testing.T, h http.Handler) (rec *httptest.ResponseRecorder, escaped any) {
    	t.Helper()
    	rec = httptest.NewRecorder()
    	func() {
    		defer func() { escaped = recover() }()
    		h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/cheese", nil))
    	}()
    	return rec, escaped
    }

    func TestRecoverPanics(t *testing.T) {
    	logs := captureLog(t)
    	h := recoverPanics(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		panic("out of gouda")
    	}))
    	rec, escaped := serve(t, h)
    	if escaped != nil {
    		t.Fatalf("recoverPanics let the panic %v escape", escaped)
    	}
    	if rec.Code != http.StatusInternalServerError {
    		t.Errorf("status = %d, want 500", rec.Code)
    	}
    	if got := strings.TrimSpace(rec.Body.String()); got != "internal server error" {
    		t.Errorf("body = %q, want %q", got, "internal server error")
    	}
    	if !strings.Contains(logs.String(), "panic: out of gouda") {
    		t.Errorf("log output = %q, want it to contain %q", logs.String(), "panic: out of gouda")
    	}
    }

    func TestRecoverPanicsPassesThrough(t *testing.T) {
    	h := recoverPanics(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		w.WriteHeader(http.StatusTeapot)
    		w.Write([]byte("fine"))
    	}))
    	rec, escaped := serve(t, h)
    	if escaped != nil {
    		t.Fatalf("unexpected panic %v", escaped)
    	}
    	if rec.Code != http.StatusTeapot || rec.Body.String() != "fine" {
    		t.Errorf("got %d %q, want 418 %q (don't change responses when nothing panics)", rec.Code, rec.Body.String(), "fine")
    	}
    }

    func TestRecoverPanicsRepanicsAbort(t *testing.T) {
    	captureLog(t)
    	h := recoverPanics(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		panic(http.ErrAbortHandler)
    	}))
    	_, escaped := serve(t, h)
    	if escaped != http.ErrAbortHandler {
    		t.Errorf("panic(http.ErrAbortHandler) should be re-panicked, got escaped = %v", escaped)
    	}
    }

    func TestChainOrder(t *testing.T) {
    	var order []string
    	mw := func(name string) func(http.Handler) http.Handler {
    		return func(next http.Handler) http.Handler {
    			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    				order = append(order, name+" in")
    				next.ServeHTTP(w, r)
    				order = append(order, name+" out")
    			})
    		}
    	}
    	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		order = append(order, "handler")
    	})
    	chain(final, mw("A"), mw("B"), mw("C")).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
    	want := []string{"A in", "B in", "C in", "handler", "C out", "B out", "A out"}
    	if !slices.Equal(order, want) {
    		t.Errorf("call order = %v, want %v", order, want)
    	}

    	order = nil
    	chain(final).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
    	if !slices.Equal(order, []string{"handler"}) {
    		t.Errorf("chain with no middleware: call order = %v, want [handler]", order)
    	}
    }

    func TestChainWithRecovery(t *testing.T) {
    	captureLog(t)
    	boom := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		var m map[string]int
    		m["x"]++
    	})
    	rec, escaped := serve(t, chain(boom, recoverPanics, addHeader("outer"), addHeader("inner")))
    	if escaped != nil {
    		t.Fatalf("panic escaped the chain: %v", escaped)
    	}
    	if rec.Code != 500 {
    		t.Errorf("status = %d, want 500", rec.Code)
    	}
    	if got := rec.Header().Values("X-Trace"); !slices.Equal(got, []string{"outer", "inner"}) {
    		t.Errorf("X-Trace = %v, want [outer inner]", got)
    	}
    }
---

A bug in one handler shouldn't take down Squeak, and it shouldn't leave the client
staring at a dropped connection either. That's the job of **recovery middleware**.

## What happens to a panic today

`net/http` already recovers panics in handlers, per connection: it logs
`http: panic serving 127.0.0.1:53712: ...` with a stack trace, then **closes the
connection**. The server stays up, but the client gets no response at all, just an
EOF, and your structured logs don't see it.

Recovery middleware does better: it catches the panic, logs it your way, and answers a
proper `500 Internal Server Error`.

```go
func recoverPanics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				log.Printf("panic: %v", v)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
```

`recover` only works in a deferred function **on the same goroutine** as the panic. If
a handler starts its own goroutine and that one panics, no middleware can save you, and
the whole program crashes. Goroutines you start need their own `recover`.

### Two edge cases

- **`http.ErrAbortHandler`** is a sentinel panic value that means "abort this
  response on purpose". The server knows not to log it. Recovery middleware should
  re-panic with it rather than turning it into a 500.
- **Headers already sent.** If the handler wrote a status before panicking, your
  `http.Error` can't change it. Go logs a "superfluous WriteHeader" warning and the
  client gets a half-written response. Logging is all you can do.

## Chaining middleware

With three or four middleware, nesting gets ugly and reads inside-out:

```go
handler := recoverPanics(withRequestID(logRequests(mux)))
```

A small `chain` helper lists them in the order a request passes through them:

```go
handler := chain(mux, recoverPanics, withRequestID, logRequests)
```

The **first** middleware in the list should be the **outermost** layer. To build that,
wrap from the *last* one backwards, so the first ends up on the outside.

### Order matters

- **Recovery goes outermost**, so it catches panics from every other middleware too.
- **Request IDs before logging**, so the logger can read the ID from the context.
- **Auth goes inside logging**, so rejected requests still get logged.

## Your task

1. Complete `recoverPanics`. When `next` panics it must log with
   `log.Printf("panic: %v", v)` and respond with
   `http.Error(w, "internal server error", http.StatusInternalServerError)`.
   If the panic value is `http.ErrAbortHandler`, panic again with it instead. When
   nothing panics, don't touch the response.
2. Complete `chain` so that `chain(h, a, b, c)` behaves exactly like `a(b(c(h)))`, and
   `chain(h)` with no middleware returns `h` as it is.

**Run** it first: the nil-map panic in `/api/cheese` escapes all the way to `main`,
and the trace header shows the middleware aren't applied yet.
