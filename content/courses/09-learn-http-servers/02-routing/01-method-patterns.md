---
title: Method Patterns
quiz:
  - question: |
      Only these routes are registered. What status does `PUT /api/squeaks` get?

      ```go
      mux.HandleFunc("GET /api/squeaks", listSqueaks)
      mux.HandleFunc("POST /api/squeaks", createSqueak)
      ```
    options:
      - text: '404 Not Found'
      - text: '405 Method Not Allowed, with an `Allow: GET, HEAD, POST` header'
        correct: true
      - text: '200, handled by `listSqueaks`'
      - text: '400 Bad Request'
    explanation: |
      The path matches registered patterns, just not with that method, so the mux answers
      `405 Method Not Allowed` and lists the methods that *would* work in the `Allow`
      header. `HEAD` is in there because a `GET` pattern also matches `HEAD`.
  - question: |
      A typo sneaks into Squeak's routes. What happens?

      ```go
      mux.HandleFunc("GET/api/squeaks", listSqueaks)
      ```
    options:
      - text: '`HandleFunc` panics because the method has no space after it'
      - text: It works exactly like `"GET /api/squeaks"`
      - text: 'It registers fine, but `GET` is read as a *host* name, so normal requests to `/api/squeaks` get a 404'
        correct: true
      - text: It matches `GET /api/squeaks` and every other method too
    explanation: |
      The method must be followed by a space. Without one, the mux sees no method at all
      and parses `GET/api/squeaks` as host `GET` plus path `/api/squeaks`. That's a valid
      pattern, so nothing panics. It just only matches requests whose `Host` header is
      `GET`, which is none of them.
---

Before Go 1.22, `http.ServeMux` could only match paths. If you wanted `GET /api/squeaks`
to list squeaks and `POST /api/squeaks` to create one, you wrote this:

```go
mux.HandleFunc("/api/squeaks", func(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listSqueaks(w, r)
	case http.MethodPost:
		createSqueak(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
})
```

It works, but every route repeats the same boilerplate, and it was one of the main reasons
people reached for third-party routers. Modern Go puts the method straight into the
pattern:

```go
mux.HandleFunc("GET /api/squeaks", listSqueaks)
mux.HandleFunc("POST /api/squeaks", createSqueak)
```

## The pattern syntax

A pattern has up to three parts:

```
[METHOD ][HOST]/[PATH]
```

- **METHOD** is optional. If it's present it must be followed by a space (or tab).
  Without a method, the pattern matches every method. Watch for typos: in
  `"GET/api/squeaks"` there's no space, so the mux reads `GET` as a *host* and the
  route silently never matches.
- **HOST** is optional and rare in APIs, e.g. `"api.squeak.dev/"`. Skip it unless you
  serve several domains from one process.
- **PATH** always starts with `/`.

Literal parts match **case-sensitively**, so `/API/squeaks` doesn't match
`/api/squeaks`.

## GET also means HEAD

A `HEAD` request asks for the headers of a `GET` without the body. The mux treats a
`GET` pattern as matching `HEAD` as well, and the server throws away the body for you.
Every other method must match exactly.

## Free 405s

Here's where method patterns really pay off. When a path matches but the method doesn't,
the mux answers `405 Method Not Allowed` and sets the `Allow` header on its own:

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/squeaks", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "all the squeaks")
	})
	mux.HandleFunc("POST /api/squeaks", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, "squeak posted")
	})

	for _, method := range []string{"GET", "HEAD", "POST", "DELETE"} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(method, "/api/squeaks", nil))
		fmt.Printf("%-6s -> %d Allow=%q body=%q\n",
			method, rec.Code, rec.Header().Get("Allow"), rec.Body.String())
	}
}
```

```
GET    -> 200 Allow="" body="all the squeaks\n"
HEAD   -> 200 Allow="" body="all the squeaks\n"
POST   -> 201 Allow="" body="squeak posted\n"
DELETE -> 405 Allow="GET, HEAD, POST" body="Method Not Allowed\n"
```

The `HEAD` line still shows a body because a recorder keeps everything the handler
writes. A real server wouldn't send those bytes over the wire.

A 405 is more helpful than a 404: it tells the client "right place, wrong verb". An
unknown path still gets `404 page not found`.

## Use the method constants

`net/http` has constants such as `http.MethodGet` and `http.MethodPost`. In patterns
the plain string is clearer, and it's what the docs use. When you compare `r.Method`
in code, prefer the constants, which the compiler checks for typos.

## Squeak's routing table

Here's the API you'll build over the course. Reading it top to bottom, you can already
see the whole shape of the app, which is a big part of why method patterns are nice:

```go
mux.HandleFunc("GET /api/healthz", handleHealthz)
mux.HandleFunc("POST /api/users", handleCreateUser)
mux.HandleFunc("POST /api/login", handleLogin)
mux.HandleFunc("GET /api/squeaks", handleListSqueaks)
mux.HandleFunc("POST /api/squeaks", handleCreateSqueak)
mux.HandleFunc("GET /api/squeaks/{id}", handleGetSqueak)
mux.HandleFunc("DELETE /api/squeaks/{id}", handleDeleteSqueak)
mux.HandleFunc("GET /admin/metrics", handleMetrics)
```

That `{id}` is a wildcard, and it's the subject of the next lesson.
