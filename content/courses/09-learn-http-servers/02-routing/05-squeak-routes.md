---
title: Squeak's Routes
quiz:
  - question: |
      You register `"GET /api/squeaks/{id}"` and `"DELETE /api/squeaks/{id}"`. What
      is the `Allow` header on the 405 response to `PATCH /api/squeaks/7`?
    options:
      - text: '`GET, DELETE`'
      - text: '`DELETE, GET, HEAD`'
        correct: true
      - text: '`GET, HEAD, DELETE, PATCH`'
      - text: There is no `Allow` header; you have to set it yourself
    explanation: |
      The mux collects every method that has a pattern matching the path, adds `HEAD`
      because of the `GET`, sorts them and sets `Allow` for you.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    )

    func handleHealthz(w http.ResponseWriter, r *http.Request) {
    	fmt.Fprint(w, "OK")
    }

    func handleHome(w http.ResponseWriter, r *http.Request) {
    	fmt.Fprint(w, "Welcome to Squeak")
    }

    func handleListSqueaks(w http.ResponseWriter, r *http.Request) {
    	fmt.Fprint(w, "all squeaks")
    }

    func handleCreateSqueak(w http.ResponseWriter, r *http.Request) {
    	w.WriteHeader(http.StatusCreated)
    	fmt.Fprint(w, "squeak created")
    }

    // handleGetSqueak answers "squeak <id>", e.g. "squeak 42".
    // If id isn't a positive whole number, it answers 400 with
    // http.Error(w, "invalid squeak id", http.StatusBadRequest).
    func handleGetSqueak(w http.ResponseWriter, r *http.Request) {
    	// ?
    	fmt.Fprint(w, "squeak ???")
    }

    // handleDeleteSqueak answers 204 No Content with no body.
    func handleDeleteSqueak(w http.ResponseWriter, r *http.Request) {
    	// ?
    }

    // handleUserSqueaks answers "squeaks by @<handle>", e.g. "squeaks by @pip".
    func handleUserSqueaks(w http.ResponseWriter, r *http.Request) {
    	// ?
    }

    func newRouter() *http.ServeMux {
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/healthz", handleHealthz)
    	// ?
    	return mux
    }

    func main() {
    	mux := newRouter()
    	for _, t := range []string{
    		"GET /",
    		"GET /api/squeaks",
    		"POST /api/squeaks",
    		"GET /api/squeaks/42",
    		"GET /api/squeaks/nibble",
    		"DELETE /api/squeaks/42",
    		"PUT /api/squeaks/42",
    		"GET /api/users/pip/squeaks",
    		"GET /about",
    	} {
    		method, path, _ := strings.Cut(t, " ")
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
    		fmt.Printf("%-27s -> %d %q\n", t, rec.Code, strings.TrimSpace(rec.Body.String()))
    	}
    }
  solution: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"strconv"
    	"strings"
    )

    func handleHealthz(w http.ResponseWriter, r *http.Request) {
    	fmt.Fprint(w, "OK")
    }

    func handleHome(w http.ResponseWriter, r *http.Request) {
    	fmt.Fprint(w, "Welcome to Squeak")
    }

    func handleListSqueaks(w http.ResponseWriter, r *http.Request) {
    	fmt.Fprint(w, "all squeaks")
    }

    func handleCreateSqueak(w http.ResponseWriter, r *http.Request) {
    	w.WriteHeader(http.StatusCreated)
    	fmt.Fprint(w, "squeak created")
    }

    func handleGetSqueak(w http.ResponseWriter, r *http.Request) {
    	id, err := strconv.Atoi(r.PathValue("id"))
    	if err != nil || id <= 0 {
    		http.Error(w, "invalid squeak id", http.StatusBadRequest)
    		return
    	}
    	fmt.Fprintf(w, "squeak %d", id)
    }

    func handleDeleteSqueak(w http.ResponseWriter, r *http.Request) {
    	w.WriteHeader(http.StatusNoContent)
    }

    func handleUserSqueaks(w http.ResponseWriter, r *http.Request) {
    	fmt.Fprintf(w, "squeaks by @%s", r.PathValue("handle"))
    }

    func newRouter() *http.ServeMux {
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/healthz", handleHealthz)
    	mux.HandleFunc("GET /{$}", handleHome)
    	mux.HandleFunc("GET /api/squeaks", handleListSqueaks)
    	mux.HandleFunc("POST /api/squeaks", handleCreateSqueak)
    	mux.HandleFunc("GET /api/squeaks/{id}", handleGetSqueak)
    	mux.HandleFunc("DELETE /api/squeaks/{id}", handleDeleteSqueak)
    	mux.HandleFunc("GET /api/users/{handle}/squeaks", handleUserSqueaks)
    	return mux
    }

    func main() {
    	mux := newRouter()
    	for _, t := range []string{
    		"GET /",
    		"GET /api/squeaks",
    		"POST /api/squeaks",
    		"GET /api/squeaks/42",
    		"GET /api/squeaks/nibble",
    		"DELETE /api/squeaks/42",
    		"PUT /api/squeaks/42",
    		"GET /api/users/pip/squeaks",
    		"GET /about",
    	} {
    		method, path, _ := strings.Cut(t, " ")
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
    		fmt.Printf("%-27s -> %d %q\n", t, rec.Code, strings.TrimSpace(rec.Body.String()))
    	}
    }
  tests: |
    package main

    import (
    	"net/http/httptest"
    	"strings"
    	"testing"
    )

    func TestRoutes(t *testing.T) {
    	mux := newRouter()
    	for _, tt := range []struct {
    		method, path string
    		wantCode     int
    		wantBody     string
    	}{
    		{"GET", "/", 200, "Welcome to Squeak"},
    		{"GET", "/api/healthz", 200, "OK"},
    		{"GET", "/api/squeaks", 200, "all squeaks"},
    		{"POST", "/api/squeaks", 201, "squeak created"},
    		{"GET", "/api/squeaks/42", 200, "squeak 42"},
    		{"GET", "/api/squeaks/7", 200, "squeak 7"},
    		{"HEAD", "/api/squeaks/7", 200, ""},
    		{"GET", "/api/squeaks/nibble", 400, "invalid squeak id"},
    		{"GET", "/api/squeaks/-3", 400, "invalid squeak id"},
    		{"GET", "/api/squeaks/0", 400, "invalid squeak id"},
    		{"DELETE", "/api/squeaks/42", 204, ""},
    		{"GET", "/api/users/pip/squeaks", 200, "squeaks by @pip"},
    		{"GET", "/api/users/whiskers/squeaks", 200, "squeaks by @whiskers"},
    		{"GET", "/about", 404, ""},
    		{"GET", "/api/squeaks/", 404, ""},
    		{"GET", "/api/squeaks/42/likes", 404, ""},
    		{"PUT", "/api/squeaks/42", 405, ""},
    		{"DELETE", "/api/squeaks", 405, ""},
    		{"POST", "/api/users/pip/squeaks", 405, ""},
    	} {
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))
    		if rec.Code != tt.wantCode {
    			t.Errorf("%s %s: status = %d, want %d", tt.method, tt.path, rec.Code, tt.wantCode)
    			continue
    		}
    		if tt.wantBody != "" {
    			if got := strings.TrimSpace(rec.Body.String()); got != tt.wantBody {
    				t.Errorf("%s %s: body = %q, want %q", tt.method, tt.path, got, tt.wantBody)
    			}
    		}
    		if tt.wantCode == 204 && rec.Body.Len() != 0 {
    			t.Errorf("%s %s: a 204 must have no body, got %q", tt.method, tt.path, rec.Body.String())
    		}
    	}
    }

    func TestAllowHeader(t *testing.T) {
    	mux := newRouter()
    	rec := httptest.NewRecorder()
    	mux.ServeHTTP(rec, httptest.NewRequest("PUT", "/api/squeaks/42", nil))
    	allow := rec.Header().Get("Allow")
    	for _, m := range []string{"GET", "DELETE"} {
    		if !strings.Contains(allow, m) {
    			t.Errorf("PUT /api/squeaks/42: Allow = %q, want it to include %s", allow, m)
    		}
    	}
    	if strings.Contains(allow, "POST") {
    		t.Errorf("PUT /api/squeaks/42: Allow = %q, POST should only be registered on /api/squeaks", allow)
    	}
    }
---

Time to give Squeak a proper routing table. You'll register every route with the right
method, pull values out of the path, and let the mux handle 404s and 405s for you.

## Your task

Complete the handlers and `newRouter` so that:

| Route | Handler | Response |
|---|---|---|
| `GET /` (exactly `/`, nothing else) | `handleHome` | 200 `Welcome to Squeak` |
| `GET /api/healthz` | `handleHealthz` | 200 `OK` (already registered) |
| `GET /api/squeaks` | `handleListSqueaks` | 200 `all squeaks` |
| `POST /api/squeaks` | `handleCreateSqueak` | 201 `squeak created` |
| `GET /api/squeaks/{id}` | `handleGetSqueak` | 200 `squeak 42` |
| `DELETE /api/squeaks/{id}` | `handleDeleteSqueak` | 204, no body |
| `GET /api/users/{handle}/squeaks` | `handleUserSqueaks` | 200 `squeaks by @pip` |

And the handlers:

1. `handleGetSqueak` reads the `id` path value. If it isn't a positive whole number
   (`strconv.Atoi` fails, or the result is zero or negative), respond with
   `http.Error(w, "invalid squeak id", http.StatusBadRequest)` and `return`. Otherwise
   write `squeak <id>`.
2. `handleDeleteSqueak` sends status `204 No Content` and nothing else.
3. `handleUserSqueaks` writes `squeaks by @<handle>`.

The tests also check what you *don't* handle yourself:

- `/about`, `/api/squeaks/` and `/api/squeaks/42/likes` must be 404s. Watch out: a plain
  `"GET /"` pattern would catch `/about`.
- `PUT /api/squeaks/42` must be a 405 whose `Allow` header lists `GET` and `DELETE`.

## Hints

- Don't forget the `return` after `http.Error`. Without it the handler keeps going and
  writes `squeak ...` after the error message.
- Calling `w.WriteHeader(http.StatusNoContent)` with no body is a complete response.
- Press **Run** to see the table of responses before you **Submit**.
