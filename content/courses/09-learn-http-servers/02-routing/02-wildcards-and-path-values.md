---
title: Wildcards and Path Values
quiz:
  - question: |
      What does `GET /files/avatars/pip.png` print?

      ```go
      mux.HandleFunc("GET /files/{path...}", func(w http.ResponseWriter, r *http.Request) {
      	fmt.Fprint(w, r.PathValue("path"))
      })
      ```
    options:
      - text: '`avatars`'
      - text: '`/avatars/pip.png`'
      - text: '`avatars/pip.png`'
        correct: true
      - text: Nothing, because `{path...}` only matches one segment
    explanation: |
      A `...` wildcard matches the rest of the path, slashes included. The value doesn't
      include the slash that comes before the wildcard, so it's `avatars/pip.png`.
  - question: |
      The only route is `"GET /api/squeaks/{id}"`. Which request does it match?
    options:
      - text: '`GET /api/squeaks/`'
      - text: '`GET /api/squeaks/42/likes`'
      - text: '`GET /api/squeaks/42`'
        correct: true
      - text: '`GET /api/squeaks`'
    explanation: |
      A plain `{name}` wildcard matches exactly one non-empty path segment.
      `/api/squeaks/42/likes` has an extra segment, and the other two have no
      segment at all where `{id}` sits.
  - question: What does `r.PathValue("idd")` return when the pattern was `"GET /api/squeaks/{id}"`?
    options:
      - text: It panics because there's no wildcard called `idd`
      - text: The empty string
        correct: true
      - text: The same as `r.PathValue("id")`
      - text: A compile error
    explanation: |
      `PathValue` returns `""` for names that aren't in the matched pattern. The compiler
      can't check wildcard names, so a typo quietly gives you an empty ID. Validate what you
      get back.
---

Every squeak has an ID, and the URL for one squeak looks like `/api/squeaks/42`. You
can't register a pattern per ID, so patterns support **wildcards**: named placeholders
for a path segment.

```go
mux.HandleFunc("GET /api/squeaks/{id}", handleGetSqueak)
```

Inside the handler, `r.PathValue` hands you whatever the wildcard matched:

```go
func handleGetSqueak(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fmt.Fprintf(w, "you asked for squeak %s\n", id)
}
```

## Single-segment wildcards

`{name}` matches exactly **one** path segment: the text between two slashes, or after
the last slash. It must be a whole segment, so `"/squeak-{id}"` is invalid and panics at
registration. The name has to be a valid Go identifier.

You can use several in one pattern:

```go
mux.HandleFunc("GET /api/users/{handle}/squeaks/{id}", handleUserSqueak)
// GET /api/users/pip/squeaks/7  ->  handle="pip", id="7"
```

## Rest-of-path wildcards

Add `...` to the **last** wildcard and it swallows the rest of the path, slashes and
all. It's handy for proxies and file-like URLs:

```go
mux.HandleFunc("GET /media/{key...}", handleMedia)
// GET /media/avatars/pip.png  ->  key="avatars/pip.png"
// GET /media/                 ->  key=""
```

A `...` wildcard also matches an *empty* rest. It can only appear at the end, since
nothing could come after it.

## Matching only the end with {$}

A pattern ending in `/` matches a whole subtree: `"GET /"` matches *every* GET request
that nothing else claims. That's rarely what you want for a home page. The special
wildcard `{$}` means "and the path ends here":

```go
mux.HandleFunc("GET /{$}", handleHome) // only exactly "/"
```

## Seeing it all at once

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "home")
	})
	mux.HandleFunc("GET /api/squeaks/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "squeak id=%q", r.PathValue("id"))
	})
	mux.HandleFunc("GET /media/{key...}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "media key=%q", r.PathValue("key"))
	})

	for _, path := range []string{
		"/",
		"/about",
		"/api/squeaks/42",
		"/api/squeaks/42/likes",
		"/api/squeaks/hello%20world",
		"/media/avatars/pip.png",
	} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		fmt.Printf("%-28s %d %s\n", path, rec.Code, strings.TrimSpace(rec.Body.String()))
	}
}
```

```
/                            200 home
/about                       404 404 page not found
/api/squeaks/42              200 squeak id="42"
/api/squeaks/42/likes        404 404 page not found
/api/squeaks/hello%20world   200 squeak id="hello world"
/media/avatars/pip.png       200 media key="avatars/pip.png"
```

Notice that path values are **unescaped** for you: `%20` became a space.

## Path values are untrusted input

`PathValue` always returns a string, and the client chose it. If your IDs are numbers,
parse them and reply `400 Bad Request` when that fails:

```go
id, err := strconv.Atoi(r.PathValue("id"))
if err != nil {
	http.Error(w, "squeak id must be a number", http.StatusBadRequest)
	return
}
```

Squeak's IDs will turn out to be UUIDs rather than numbers, but the rule is the same:
parse, validate, and never paste a path value straight into a SQL query, a file path
or a shell command.

## Testing with SetPathValue

If you call a handler directly in a test, without a mux in front of it, nothing fills
in the path values. `r.SetPathValue("id", "42")` sets one by hand, so you can test the
handler on its own. You'll use it in the testing chapter.
