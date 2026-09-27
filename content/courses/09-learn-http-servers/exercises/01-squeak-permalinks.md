---
title: Squeak Permalinks
difficulty: easy
after: routing
hints:
  - 'Put the method in the pattern (`"GET /api/squeaks/{id}"`) and read the wildcard with `r.PathValue("id")`. You don''t need to check `r.Method` yourself: once a path has method patterns, `ServeMux` answers other methods with `405` and an `Allow` header on its own.'
  - 'A wildcard normally matches one path segment. To match the rest of the path, slashes and all, end the pattern with `{path...}`. To match `/` and *only* `/`, use `/{$}`, because a plain `/` matches everything.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    )

    // newRouter returns Squeak's public routes. Each handler writes plain text:
    //
    //	GET    /                             -> welcome to squeak
    //	GET    /api/squeaks/{id}             -> squeak <id>
    //	DELETE /api/squeaks/{id}             -> deleted <id>
    //	GET    /api/users/{handle}/squeaks   -> squeaks by @<handle>
    //	GET    /media/<any path>             -> media <path>
    //
    // Everything else is a 404, and a known path with the wrong method is a 405.
    func newRouter() *http.ServeMux {
    	mux := http.NewServeMux()
    	// Register one pattern per route, with the method in the pattern.
    	// Use r.PathValue to read wildcards and fmt.Fprintf to write the reply.
    	return mux
    }

    func main() {
    	mux := newRouter()
    	for _, target := range []string{"/api/squeaks/42", "/api/users/pip/squeaks", "/media/avatars/pip.png"} {
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, httptest.NewRequest("GET", target, nil))
    		fmt.Printf("GET %s -> %d %q\n", target, rec.Code, rec.Body.String())
    	}
    }
  solution: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    )

    func newRouter() *http.ServeMux {
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprint(w, "welcome to squeak")
    	})
    	mux.HandleFunc("GET /api/squeaks/{id}", func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprintf(w, "squeak %s", r.PathValue("id"))
    	})
    	mux.HandleFunc("DELETE /api/squeaks/{id}", func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprintf(w, "deleted %s", r.PathValue("id"))
    	})
    	mux.HandleFunc("GET /api/users/{handle}/squeaks", func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprintf(w, "squeaks by @%s", r.PathValue("handle"))
    	})
    	mux.HandleFunc("GET /media/{path...}", func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprintf(w, "media %s", r.PathValue("path"))
    	})
    	return mux
    }

    func main() {
    	mux := newRouter()
    	for _, target := range []string{"/api/squeaks/42", "/api/users/pip/squeaks", "/media/avatars/pip.png"} {
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, httptest.NewRequest("GET", target, nil))
    		fmt.Printf("GET %s -> %d %q\n", target, rec.Code, rec.Body.String())
    	}
    }
  tests: |
    package main

    import (
    	"net/http/httptest"
    	"testing"
    )

    func serve(method, target string) *httptest.ResponseRecorder {
    	rec := httptest.NewRecorder()
    	newRouter().ServeHTTP(rec, httptest.NewRequest(method, target, nil))
    	return rec
    }

    func TestRoutes(t *testing.T) {
    	tests := []struct {
    		method, target string
    		want           string
    	}{
    		{"GET", "/", "welcome to squeak"},
    		{"GET", "/api/squeaks/42", "squeak 42"},
    		{"GET", "/api/squeaks/0192f1e2-8c3a", "squeak 0192f1e2-8c3a"},
    		{"DELETE", "/api/squeaks/7", "deleted 7"},
    		{"GET", "/api/users/pip/squeaks", "squeaks by @pip"},
    		{"GET", "/api/users/Big_Cheese/squeaks", "squeaks by @Big_Cheese"},
    		{"GET", "/media/pip.png", "media pip.png"},
    		{"GET", "/media/avatars/2026/pip.png", "media avatars/2026/pip.png"},
    	}
    	for _, tt := range tests {
    		rec := serve(tt.method, tt.target)
    		if rec.Code != 200 || rec.Body.String() != tt.want {
    			t.Errorf("%s %s = %d %q, want 200 %q", tt.method, tt.target, rec.Code, rec.Body.String(), tt.want)
    		}
    	}
    }

    func TestHeadFollowsGet(t *testing.T) {
    	if rec := serve("HEAD", "/api/squeaks/42"); rec.Code != 200 {
    		t.Errorf("HEAD /api/squeaks/42 = %d, want 200 (a GET pattern also serves HEAD)", rec.Code)
    	}
    }

    func TestNotFound(t *testing.T) {
    	for _, target := range []string{"/nope", "/api/squeaks", "/api/squeaks/42/likes", "/api/users/pip", "/favicon.ico"} {
    		if rec := serve("GET", target); rec.Code != 404 {
    			t.Errorf("GET %s = %d %q, want 404", target, rec.Code, rec.Body.String())
    		}
    	}
    }

    func TestMethodNotAllowed(t *testing.T) {
    	tests := []struct {
    		method, target, allow string
    	}{
    		{"POST", "/api/squeaks/42", "DELETE, GET, HEAD"},
    		{"PUT", "/api/squeaks/42", "DELETE, GET, HEAD"},
    		{"DELETE", "/api/users/pip/squeaks", "GET, HEAD"},
    		{"POST", "/media/pip.png", "GET, HEAD"},
    	}
    	for _, tt := range tests {
    		rec := serve(tt.method, tt.target)
    		if rec.Code != 405 {
    			t.Errorf("%s %s = %d, want 405 Method Not Allowed", tt.method, tt.target, rec.Code)
    			continue
    		}
    		if got := rec.Header().Get("Allow"); got != tt.allow {
    			t.Errorf("%s %s: Allow header = %q, want %q", tt.method, tt.target, got, tt.allow)
    		}
    	}
    }
---

Squeak's public links need a router. Every link is a `GET` except deleting a
squeak, and each reply is a line of plain text for now.

Complete `newRouter` so it registers these routes on a fresh `http.ServeMux`:

| Method | Path | Response body |
|---|---|---|
| `GET` | `/` (exactly) | `welcome to squeak` |
| `GET` | `/api/squeaks/{id}` | `squeak <id>` |
| `DELETE` | `/api/squeaks/{id}` | `deleted <id>` |
| `GET` | `/api/users/{handle}/squeaks` | `squeaks by @<handle>` |
| `GET` | `/media/` followed by any path | `media <path>` |

Anything else must be a `404`. A known path with the wrong method must be a
`405 Method Not Allowed` with an `Allow` header listing the methods that work.

## Examples

```
GET    /api/squeaks/42              200 squeak 42
GET    /media/avatars/2026/pip.png  200 media avatars/2026/pip.png
GET    /favicon.ico                 404
POST   /api/squeaks/42              405 Allow: DELETE, GET, HEAD
```

## Constraints

- Use method patterns and wildcards. `ServeMux` produces the 404s, the 405s and
  the `Allow` header by itself, so you don't have to.
- `/` must match only the root path, not every path.
- A `GET` pattern also answers `HEAD` requests.
