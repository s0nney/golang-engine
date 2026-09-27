---
title: Routing Precedence
quiz:
  - question: |
      Which handler serves `GET /api/users/me`?

      ```go
      mux.HandleFunc("GET /api/users/{handle}", handleUser)
      mux.HandleFunc("GET /api/users/me", handleMe)
      mux.HandleFunc("GET /api/", handleAPIFallback)
      ```
    options:
      - text: '`handleUser`, because it was registered first'
      - text: '`handleMe`, because it matches a strict subset of what the others match'
        correct: true
      - text: '`handleAPIFallback`, because it has the shortest pattern'
      - text: None; registering these three panics because they overlap
    explanation: |
      Registration order never matters. The most *specific* pattern wins, and
      `/api/users/me` matches exactly one path, which the other two also match. Overlapping
      is fine as long as one pattern is clearly more specific than the other.
  - question: |
      What happens when this program starts?

      ```go
      mux.HandleFunc("GET /api/squeaks/{id}", handleGetSqueak)
      mux.HandleFunc("GET /api/{kind}/latest", handleLatest)
      ```
    options:
      - text: It runs; `/api/squeaks/latest` goes to `handleLatest`
      - text: It runs; `/api/squeaks/latest` goes to `handleGetSqueak`
      - text: The second `HandleFunc` panics because the patterns conflict
        correct: true
      - text: It compiles but every request gets a 500
    explanation: |
      Both patterns match `/api/squeaks/latest`, but neither is more specific: one has a
      literal where the other has a wildcard, *and* the other way round. The mux refuses to
      guess and panics at registration, with a message showing a path that both match.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    )

    // reply returns a handler that writes name, so you can see which one ran.
    func reply(name string) http.HandlerFunc {
    	return func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprint(w, name)
    	}
    }

    // newRouter panics as soon as it runs. Read the panic message, then fix
    // the patterns so that every route in the lesson's table works.
    func newRouter() *http.ServeMux {
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /", reply("home"))
    	mux.HandleFunc("/api/healthz", reply("healthz"))
    	mux.HandleFunc("GET /api/users/{handle}", reply("user"))
    	mux.HandleFunc("GET /api/users/me", reply("me"))
    	mux.HandleFunc("GET /api/squeaks/{id}", reply("squeak"))
    	mux.HandleFunc("GET /api/{kind}/latest", reply("latest"))
    	return mux
    }

    func main() {
    	var mux *http.ServeMux
    	func() {
    		defer func() {
    			if v := recover(); v != nil {
    				fmt.Println("newRouter panicked:", v)
    			}
    		}()
    		mux = newRouter()
    	}()
    	if mux == nil {
    		return
    	}
    	for _, t := range []string{
    		"GET /", "GET /about", "POST /about", "GET /api/healthz", "POST /api/healthz",
    		"GET /api/users/me", "GET /api/users/pip", "GET /api/squeaks/7", "GET /api/squeaks/latest",
    	} {
    		method, path, _ := strings.Cut(t, " ")
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
    		fmt.Printf("%-24s -> %d %s\n", t, rec.Code, strings.TrimSpace(rec.Body.String()))
    	}
    }
  solution: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    )

    func reply(name string) http.HandlerFunc {
    	return func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprint(w, name)
    	}
    }

    func newRouter() *http.ServeMux {
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /{$}", reply("home"))
    	mux.HandleFunc("GET /api/healthz", reply("healthz"))
    	mux.HandleFunc("GET /api/users/{handle}", reply("user"))
    	mux.HandleFunc("GET /api/users/me", reply("me"))
    	mux.HandleFunc("GET /api/squeaks/{id}", reply("squeak"))
    	mux.HandleFunc("GET /api/squeaks/latest", reply("latest"))
    	return mux
    }

    func main() {
    	var mux *http.ServeMux
    	func() {
    		defer func() {
    			if v := recover(); v != nil {
    				fmt.Println("newRouter panicked:", v)
    			}
    		}()
    		mux = newRouter()
    	}()
    	if mux == nil {
    		return
    	}
    	for _, t := range []string{
    		"GET /", "GET /about", "POST /about", "GET /api/healthz", "POST /api/healthz",
    		"GET /api/users/me", "GET /api/users/pip", "GET /api/squeaks/7", "GET /api/squeaks/latest",
    	} {
    		method, path, _ := strings.Cut(t, " ")
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
    		fmt.Printf("%-24s -> %d %s\n", t, rec.Code, strings.TrimSpace(rec.Body.String()))
    	}
    }
  tests: |
    package main

    import (
    	"net/http"
    	"net/http/httptest"
    	"testing"
    )

    func buildRouter(t *testing.T) (mux *http.ServeMux) {
    	t.Helper()
    	defer func() {
    		if v := recover(); v != nil {
    			t.Fatalf("newRouter panicked: %v", v)
    		}
    	}()
    	return newRouter()
    }

    func TestRoutes(t *testing.T) {
    	mux := buildRouter(t)
    	for _, tt := range []struct {
    		method, path string
    		wantCode     int
    		wantBody     string
    	}{
    		{"GET", "/", 200, "home"},
    		{"GET", "/about", 404, ""},
    		{"POST", "/about", 404, ""},
    		{"POST", "/", 405, ""},
    		{"GET", "/api/healthz", 200, "healthz"},
    		{"POST", "/api/healthz", 405, ""},
    		{"GET", "/api/users/me", 200, "me"},
    		{"GET", "/api/users/pip", 200, "user"},
    		{"GET", "/api/users/latest", 200, "user"},
    		{"GET", "/api/squeaks/7", 200, "squeak"},
    		{"GET", "/api/squeaks/latest", 200, "latest"},
    		{"GET", "/api/squeaks/7/latest", 404, ""},
    	} {
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))
    		if rec.Code != tt.wantCode {
    			t.Errorf("%s %s: status = %d, want %d", tt.method, tt.path, rec.Code, tt.wantCode)
    			continue
    		}
    		if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
    			t.Errorf("%s %s: handled by %q, want %q", tt.method, tt.path, rec.Body.String(), tt.wantBody)
    		}
    	}
    }
---

As Squeak grows, patterns start to overlap. `GET /api/users/me` (the logged-in mouse)
and `GET /api/users/{handle}` (any mouse) both match the path `/api/users/me`. Which one
runs?

Older routers answered "whichever you registered first" or "the longest pattern".
`ServeMux` has one rule that's easy to state:

> **The most specific pattern wins.** Pattern A is more specific than B if A matches a
> strict subset of the requests B matches.

Registration order is irrelevant. You can shuffle your `HandleFunc` calls and routing
doesn't change.

## Applying the rule

- `/api/users/me` vs `/api/users/{handle}`: every request the first matches, the second
  matches too, but not the other way round. The literal is more specific, so `me` wins.
- `GET /api/squeaks/{id}` vs `/api/squeaks/{id}` (no method): the `GET` version matches
  fewer requests, so it wins for GET and HEAD. Other methods fall through to the one
  without a method.
- `/api/squeaks/{id}` vs `/api/` (a subtree): the wildcard pattern only matches
  three-segment paths, while `/api/` matches everything under `/api/`. The wildcard wins.

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	mux := http.NewServeMux()
	route := func(name string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "%-10s (pattern %q)", name, r.Pattern)
		}
	}
	mux.HandleFunc("GET /", route("catch-all"))
	mux.HandleFunc("GET /api/", route("api"))
	mux.HandleFunc("GET /api/users/{handle}", route("user"))
	mux.HandleFunc("GET /api/users/me", route("me"))

	for _, path := range []string{"/api/users/me", "/api/users/pip", "/api/teapot", "/about"} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		fmt.Printf("%-15s -> %s\n", path, rec.Body.String())
	}
}
```

```
/api/users/me   -> me         (pattern "GET /api/users/me")
/api/users/pip  -> user       (pattern "GET /api/users/{handle}")
/api/teapot     -> api        (pattern "GET /api/")
/about          -> catch-all  (pattern "GET /")
```

`r.Pattern` holds the pattern that matched. It's great in logs and metrics, because
`GET /api/squeaks/{id}` groups nicely where a thousand different IDs wouldn't.

## Conflicts panic at startup

Sometimes neither pattern is more specific. Take these two:

```go
mux.HandleFunc("GET /api/squeaks/{id}", handleGetSqueak)
mux.HandleFunc("GET /api/{kind}/latest", handleLatest)
```

Both match `/api/squeaks/latest`. But the first also matches `/api/squeaks/7`, which the
second doesn't, and the second matches `/api/users/latest`, which the first doesn't.
Neither is a subset of the other, so the second `HandleFunc` **panics**:

```
pattern "GET /api/{kind}/latest" ... conflicts with pattern "GET /api/squeaks/{id}" ...:
GET /api/{kind}/latest and GET /api/squeaks/{id} both match some paths, like "/api/squeaks/latest".
But neither is more specific than the other.
```

That's a feature. A routing ambiguity blows up the moment the program starts rather than
surprising a user in production. Registering the same pattern twice panics too, even
if the wildcard names differ (`{id}` vs `{squeakID}`).

A sneakier conflict mixes methods and paths:

```go
mux.HandleFunc("GET /", handleHome)
mux.HandleFunc("/index.html", handleIndex) // panics
```

`GET /` matches fewer *methods* but more *paths*. Neither is a subset, so they conflict.
Adding `GET` to the second pattern fixes it.

## The catch-all and 405s

One more consequence: a catch-all `GET /` changes what unknown paths return. With it
registered, `POST /anything` gets **405** rather than 404, because some pattern matches
the path, just not for POST. If you want clean 404s for other methods, register the
catch-all without a method, or use `GET /{$}` so it only covers the home page.

## Quick checklist

- Literal segments beat wildcards in the same position.
- A pattern with a method beats the same pattern without one.
- A single-segment `{x}` beats a trailing `/` or `{x...}` that covers more.
- If you can find one path each pattern matches and the other doesn't, they conflict, and
  you'll hear about it at startup.

## Your task

A teammate wrote `newRouter`, and Squeak crashes on startup. **Run** it to read the
panic, then fix the patterns (you won't need to touch the handlers) so that:

| Request | Result |
|---|---|
| `GET /` | `home`, and *only* the exact path `/` |
| `GET /about`, `POST /about` | 404 |
| `POST /` | 405 |
| `GET /api/healthz` | `healthz` |
| `POST /api/healthz` | 405 |
| `GET /api/users/me` | `me` |
| `GET /api/users/pip`, `GET /api/users/latest` | `user` |
| `GET /api/squeaks/7` | `squeak` |
| `GET /api/squeaks/latest` | `latest` |

There are two conflicts to untangle. Once the first is fixed, **Run** again to see the
second. Squeak only needs "latest" for squeaks, so the `{kind}` wildcard can go.
