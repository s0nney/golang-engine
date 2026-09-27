---
title: Trailing Slashes and Redirects
quiz:
  - question: |
      The only route is `"GET /app/"`. What does `GET /app` get?
    options:
      - text: '404, because `/app` and `/app/` are different paths'
      - text: '200, because trailing slashes are ignored'
      - text: 'A `307 Temporary Redirect` to `/app/`'
        correct: true
      - text: 'A `308 Permanent Redirect` to `/app/`'
    explanation: |
      Registering a subtree (`/app/`) also makes the mux redirect its root without the
      slash (`/app`) to it. Current Go versions use `307`, and older ones used `301`.
      Registering `/app` separately turns the redirect off.
  - question: Why is `307` a better choice than `301` for these automatic redirects?
    options:
      - text: '`307` is cached forever by browsers, so it''s faster'
      - text: 'Clients must repeat the request with the same method and body, so a `POST` stays a `POST`'
        correct: true
      - text: '`301` is not allowed for API responses'
      - text: '`307` makes the client drop its cookies, which is more secure'
    explanation: |
      After a `301`, most clients historically switch a `POST` to a `GET` and drop the body.
      A `307` requires the same method and body, and it isn't cached permanently the way a
      `301` is.
  - question: |
      The only route is `"GET /api/users"`. What does `GET /api/users/` get?
    options:
      - text: A redirect to `/api/users`
      - text: '404 Not Found'
        correct: true
      - text: '200 from the `/api/users` handler'
      - text: '405 Method Not Allowed'
    explanation: |
      The mux only adds slashes, it never removes them. `/api/users` without a trailing
      slash is an exact match, so `/api/users/` matches nothing and gets a 404.
---

To you, `/app` and `/app/` look the same. To `ServeMux`, they're two different paths,
and a trailing slash in a *pattern* changes its meaning completely.

## A trailing slash means "subtree"

- `"GET /api/users"` matches exactly `/api/users`. Nothing else.
- `"GET /app/"` matches `/app/` **and everything below it**: `/app/logo.txt`,
  `/app/css/style.css`, and so on. A trailing slash works like an anonymous `{...}`
  wildcard.

That's why the file server in chapter 1 was registered as `"GET /app/"`: it needs every
file under the prefix.

## The automatic redirect

If you register a subtree like `/app/` and a request arrives for `/app`, the mux
assumes the client forgot the slash and **redirects** it:

```
GET /app
HTTP/1.1 307 Temporary Redirect
Location: /app/
```

It only goes one way. The mux adds a missing slash for subtrees, but never removes an
extra one, so with `"GET /api/users"` registered, `/api/users/` is a plain 404.

To stop the redirect, register the slash-less path yourself. An explicit pattern beats
the automatic behaviour:

```go
mux.HandleFunc("GET /app", handleAppRoot) // no redirect for /app any more
mux.Handle("GET /app/", appFiles)
```

## Path cleaning redirects too

The mux also tidies messy paths before routing. Repeated slashes and `.` or `..`
segments get redirected to their clean form:

```
GET /api//users          -> 307, Location: /api/users
GET /app/../api/users    -> 307, Location: /api/users
```

So your handlers never see `..` in `r.URL.Path`, which is one less path-traversal
trick to worry about.

## 301 vs 307: a change in Go 1.26

For years these automatic redirects used **`301 Moved Permanently`**. That caused two
problems:

1. **Methods changed.** After a `301`, browsers and many clients re-send a `POST` as a
   `GET` and drop the body. A client that posted to `/api/squeaks` when the route was
   `/api/squeaks/` would find its squeak silently turned into a listing request.
2. **Browsers cache 301s forever.** Fix your routes later and some users are still
   stuck with the old redirect.

Since Go 1.26, `ServeMux` answers with **`307 Temporary Redirect`**, which requires the
client to repeat the *same* method with the *same* body and isn't cached permanently.
Go 1.27 behaves the same way, so here's the round trip with a real client:

```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/squeaks/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Fprintf(w, "%s %s got %s", r.Method, r.URL.Path, body)
	})

	srv := httptest.NewServer(mux) // a real server on a loopback port
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/squeaks", "application/json",
		strings.NewReader(`{"body":"cheese?"}`))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	reply, _ := io.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode, string(reply))
}
```

```
200 POST /api/squeaks/ got {"body":"cheese?"}
```

The client got a 307, followed it, and re-sent the `POST` with its body. With a 301 it
would have arrived as a `GET` with no body. (`httptest.NewServer` starts a real server
on a random loopback port. It has its own lesson in the testing chapter.)

If you're writing to someone else's server and see an unexpected 307 in your logs, that's
usually a missing trailing slash.

## Advice for APIs

Redirects are friendly for browsers and wasteful for APIs, since every one costs an
extra round trip. For Squeak:

- Register collections **without** a trailing slash: `"GET /api/squeaks"`, not
  `"GET /api/squeaks/"`.
- Use wildcards for items: `"GET /api/squeaks/{id}"`.
- Keep trailing-slash subtrees for things that really are trees, like static files.
- Use `{$}` when you want a slash-terminated path to match exactly, such as
  `"GET /{$}"` for the home page.
