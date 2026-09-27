---
title: What Is a Server?
quiz:
  - question: In the HTTP request/response model, who speaks first?
    options:
      - text: The server, which pushes data to every connected client
      - text: The client, which sends a request that the server answers
        correct: true
      - text: Whoever opened the TCP connection last
      - text: Neither; both send at the same time
    explanation: |
      HTTP is a request/response protocol. A client (a browser, `curl`, another program)
      sends a request, and the server sends back exactly one response to it. A plain
      HTTP server never starts a conversation on its own.
  - question: |
      What does this program print?

      ```go
      h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      	fmt.Fprintf(w, "hi from %s", r.URL.Path)
      })
      rec := httptest.NewRecorder()
      h.ServeHTTP(rec, httptest.NewRequest("GET", "/squeaks", nil))
      fmt.Println(rec.Code, rec.Body.String())
      ```
    options:
      - text: '`0 hi from /squeaks`'
      - text: '`200 hi from /squeaks`'
        correct: true
      - text: '`200 hi from GET /squeaks`'
      - text: Nothing, because no server is listening on a port
    explanation: |
      The handler never calls `WriteHeader`, so the first write sends an implicit
      `200 OK`. The recorder captures it. No port is involved: `httptest.NewRecorder`
      lets you call a handler directly, like any other function.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    )

    // hello answers every request as plain text. See the lesson for the rules.
    func hello(w http.ResponseWriter, r *http.Request) {
    	fmt.Fprint(w, "Squeak!")
    }

    func main() {
    	for _, target := range []string{"/api/squeaks", "/api/users/pip", "/api/secret-cheese"} {
    		rec := httptest.NewRecorder()
    		http.HandlerFunc(hello).ServeHTTP(rec, httptest.NewRequest("GET", target, nil))
    		fmt.Printf("%d mood=%q %q\n", rec.Code, rec.Result().Header.Get("X-Squeak-Mood"), rec.Body.String())
    	}
    }
  solution: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    )

    func hello(w http.ResponseWriter, r *http.Request) {
    	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    	w.Header().Set("X-Squeak-Mood", "cheerful")
    	if r.URL.Path == "/api/secret-cheese" {
    		w.WriteHeader(http.StatusForbidden)
    		fmt.Fprintln(w, "no peeking")
    		return
    	}
    	fmt.Fprintf(w, "Squeak! You asked for %s %s\n", r.Method, r.URL.Path)
    }

    func main() {
    	for _, target := range []string{"/api/squeaks", "/api/users/pip", "/api/secret-cheese"} {
    		rec := httptest.NewRecorder()
    		http.HandlerFunc(hello).ServeHTTP(rec, httptest.NewRequest("GET", target, nil))
    		fmt.Printf("%d mood=%q %q\n", rec.Code, rec.Result().Header.Get("X-Squeak-Mood"), rec.Body.String())
    	}
    }
  tests: |
    package main

    import (
    	"net/http"
    	"net/http/httptest"
    	"testing"
    )

    func TestHello(t *testing.T) {
    	for _, tt := range []struct {
    		method, path string
    		wantCode     int
    		wantBody     string
    	}{
    		{"GET", "/api/squeaks", 200, "Squeak! You asked for GET /api/squeaks\n"},
    		{"POST", "/api/squeaks", 200, "Squeak! You asked for POST /api/squeaks\n"},
    		{"DELETE", "/api/users/pip", 200, "Squeak! You asked for DELETE /api/users/pip\n"},
    		{"GET", "/api/secret-cheese", 403, "no peeking\n"},
    		{"POST", "/api/secret-cheese", 403, "no peeking\n"},
    	} {
    		rec := httptest.NewRecorder()
    		http.HandlerFunc(hello).ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))
    		res := rec.Result() // what a client would actually receive
    		if res.StatusCode != tt.wantCode {
    			t.Errorf("%s %s: status = %d, want %d", tt.method, tt.path, res.StatusCode, tt.wantCode)
    		}
    		if got := rec.Body.String(); got != tt.wantBody {
    			t.Errorf("%s %s: body = %q, want %q", tt.method, tt.path, got, tt.wantBody)
    		}
    		if got := res.Header.Get("Content-Type"); got != "text/plain; charset=utf-8" {
    			t.Errorf("%s %s: Content-Type = %q, want %q", tt.method, tt.path, got, "text/plain; charset=utf-8")
    		}
    		if got := res.Header.Get("X-Squeak-Mood"); got != "cheerful" {
    			t.Errorf("%s %s: X-Squeak-Mood = %q, want %q (headers must be set before WriteHeader)", tt.method, tt.path, got, "cheerful")
    		}
    	}
    }
---

Welcome to the last course of the roadmap. You already know how to *call* web APIs with
an HTTP client. Now you'll build the other side: a **server**.

Across this course you'll grow one project, **Squeak**, a tiny microblogging API where
mice post short messages called *squeaks*. By the end it will have users, logins,
signed tokens, owner-only deletes, webhooks, an admin metrics endpoint, tests and a
graceful shutdown. You'll use nothing but Go's standard library.

## What a server does

A web server is just a program that:

1. **Listens** on a network port (like `8080`) for incoming connections.
2. **Reads** an HTTP request from each connection: a method, a path, headers and maybe a body.
3. **Decides** what to do based on that request (this is your code).
4. **Writes** an HTTP response: a status code, headers and a body.

Then it goes back to listening. A busy server does this for thousands of connections at
once, and Go's `net/http` runs each request in its own goroutine, so your code is
concurrent from the very first line whether you like it or not. Remember that when we get
to storage.

Here's the shape of a single exchange, as raw text on the wire:

```
GET /api/squeaks HTTP/1.1
Host: localhost:8080
Accept: application/json

HTTP/1.1 200 OK
Content-Type: application/json

[{"body":"cheese is back in stock"}]
```

## Handlers

In Go, the "decide and respond" part is a **handler**: anything with a
`ServeHTTP(http.ResponseWriter, *http.Request)` method. The `*http.Request` is what the
client sent. The `http.ResponseWriter` is how you answer.

You don't even need a network to try one. The `net/http/httptest` package has a
*recorder* that plays the role of the client's connection and captures whatever the
handler writes:

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func hello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Squeak! You asked for %s %s\n", r.Method, r.URL.Path)
}

func main() {
	req := httptest.NewRequest("GET", "/api/squeaks", nil)
	rec := httptest.NewRecorder()

	http.HandlerFunc(hello).ServeHTTP(rec, req)

	fmt.Println("status:", rec.Code)
	fmt.Println("type:  ", rec.Header().Get("Content-Type"))
	fmt.Print("body:   ", rec.Body.String())
}
```

It prints:

```
status: 200
type:   text/plain; charset=utf-8
body:   Squeak! You asked for GET /api/squeaks
```

Notice two things:

- We never set the status. If a handler writes a body without calling
  `w.WriteHeader`, Go sends `200 OK` for you.
- `http.HandlerFunc(hello)` is a type conversion, not a function call. It turns a plain
  function into something with a `ServeHTTP` method. You'll dig into that in the
  handlers chapter.

## Why this course runs handlers without a port

A real server blocks forever waiting for connections, which is exactly what you want in
production and exactly what you *don't* want in an exercise that must finish in a few
seconds. So most programs in this course drive handlers with `httptest`, the same way
you'll test them professionally in the testing chapter. Whenever a lesson shows a
server listening on a port, try it on your own machine with `go run` and `curl`.

## Order matters: headers, status, body

A response goes out in a fixed order: status line, then headers, then body. A
`ResponseWriter` follows the same order:

1. Set headers with `w.Header().Set(...)`.
2. Optionally pick a status with `w.WriteHeader(code)`.
3. Write the body with `w.Write` or `fmt.Fprint(w, ...)`.

Once the status is sent (by `WriteHeader`, or implicitly by the first write), the headers
are on their way, and later `Header().Set` calls are silently ignored.

## Your turn

Complete `hello` so that every response:

- has the headers `Content-Type: text/plain; charset=utf-8` and `X-Squeak-Mood: cheerful`,
- for the path `/api/secret-cheese`, has status **403** and the body `no peeking` plus
  a newline (use `http.StatusForbidden`),
- for any other path, has status 200 and the body
  `Squeak! You asked for <METHOD> <PATH>` plus a newline.

The tests use `rec.Result()`, which shows what a real client would receive, so headers
set after `WriteHeader` won't count.
