---
title: Serving Static Files
quiz:
  - question: |
      `assets` contains a single file, `logo.txt`. What status does `GET /app/logo.txt` get?

      ```go
      mux.Handle("GET /app/", http.FileServerFS(assets))
      ```
    options:
      - text: '200, because the pattern matches'
      - text: '404, because the file server looks for `app/logo.txt`'
        correct: true
      - text: '301, a redirect to `/logo.txt`'
      - text: '405, because file servers only accept `HEAD`'
    explanation: |
      The file server maps the *whole* URL path onto the file system, so it looks for
      `app/logo.txt`, which doesn't exist. Wrap it in `http.StripPrefix("/app", ...)` so it
      sees `/logo.txt`.
  - question: What's the difference between `http.FileServer` and `http.FileServerFS`?
    options:
      - text: '`FileServerFS` takes any `fs.FS`, such as `os.DirFS` or an `embed.FS`; `FileServer` takes an `http.FileSystem`'
        correct: true
      - text: '`FileServerFS` caches files in memory; `FileServer` reads from disk every time'
      - text: '`FileServerFS` can only serve embedded files'
      - text: There is no difference; one is an alias of the other
    explanation: |
      `FileServer(http.Dir("static"))` predates the `io/fs` package. `FileServerFS`
      (Go 1.22) accepts the standard `fs.FS` interface directly, so embedded files,
      directories and in-memory test file systems all plug straight in.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"io/fs"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"testing/fstest"
    )

    // newServer builds Squeak's server. It should:
    //   - answer GET /api/healthz with 200, body "OK" and
    //     Content-Type "text/plain; charset=utf-8"
    //   - serve the files in assets under GET /app/ (so /app/logo.txt serves logo.txt)
    //   - set ReadHeaderTimeout 5s, WriteTimeout 10s and IdleTimeout 60s
    func newServer(addr string, assets fs.FS) *http.Server {
    	mux := http.NewServeMux()
    	// ?

    	return &http.Server{
    		Addr:    addr,
    		Handler: mux,
    		// ?
    	}
    }

    func main() {
    	assets := fstest.MapFS{
    		"index.html": {Data: []byte("<h1>Welcome to Squeak</h1>")},
    		"logo.txt":   {Data: []byte("(\\_/)\n(o.o)")},
    	}
    	srv := newServer(":8080", assets)
    	fmt.Println("timeouts:", srv.ReadHeaderTimeout, srv.WriteTimeout, srv.IdleTimeout)

    	for _, target := range []string{"/api/healthz", "/app/", "/app/logo.txt", "/app/missing.txt"} {
    		rec := httptest.NewRecorder()
    		srv.Handler.ServeHTTP(rec, httptest.NewRequest("GET", target, nil))
    		fmt.Printf("GET %-17s -> %d %q\n", target, rec.Code, strings.TrimSpace(rec.Body.String()))
    	}
    }
  solution: |
    package main

    import (
    	"fmt"
    	"io/fs"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"testing/fstest"
    	"time"
    )

    func newServer(addr string, assets fs.FS) *http.Server {
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
    		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    		w.Write([]byte("OK"))
    	})
    	mux.Handle("GET /app/", http.StripPrefix("/app", http.FileServerFS(assets)))

    	return &http.Server{
    		Addr:              addr,
    		Handler:           mux,
    		ReadHeaderTimeout: 5 * time.Second,
    		WriteTimeout:      10 * time.Second,
    		IdleTimeout:       60 * time.Second,
    	}
    }

    func main() {
    	assets := fstest.MapFS{
    		"index.html": {Data: []byte("<h1>Welcome to Squeak</h1>")},
    		"logo.txt":   {Data: []byte("(\\_/)\n(o.o)")},
    	}
    	srv := newServer(":8080", assets)
    	fmt.Println("timeouts:", srv.ReadHeaderTimeout, srv.WriteTimeout, srv.IdleTimeout)

    	for _, target := range []string{"/api/healthz", "/app/", "/app/logo.txt", "/app/missing.txt"} {
    		rec := httptest.NewRecorder()
    		srv.Handler.ServeHTTP(rec, httptest.NewRequest("GET", target, nil))
    		fmt.Printf("GET %-17s -> %d %q\n", target, rec.Code, strings.TrimSpace(rec.Body.String()))
    	}
    }
  tests: |
    package main

    import (
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"testing"
    	"testing/fstest"
    	"time"
    )

    var testAssets = fstest.MapFS{
    	"index.html":    {Data: []byte("<h1>Squeak home</h1>")},
    	"logo.txt":      {Data: []byte("mouse logo")},
    	"css/style.css": {Data: []byte("body { color: grey; }")},
    }

    func do(t *testing.T, srv *http.Server, method, target string) *httptest.ResponseRecorder {
    	t.Helper()
    	if srv == nil || srv.Handler == nil {
    		t.Fatal("newServer returned a server with no Handler")
    	}
    	rec := httptest.NewRecorder()
    	srv.Handler.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
    	return rec
    }

    func TestTimeouts(t *testing.T) {
    	srv := newServer(":9999", testAssets)
    	if srv.Addr != ":9999" {
    		t.Errorf("Addr = %q, want %q", srv.Addr, ":9999")
    	}
    	for _, tt := range []struct {
    		name      string
    		got, want time.Duration
    	}{
    		{"ReadHeaderTimeout", srv.ReadHeaderTimeout, 5 * time.Second},
    		{"WriteTimeout", srv.WriteTimeout, 10 * time.Second},
    		{"IdleTimeout", srv.IdleTimeout, 60 * time.Second},
    	} {
    		if tt.got != tt.want {
    			t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
    		}
    	}
    }

    func TestHealthz(t *testing.T) {
    	srv := newServer(":8080", testAssets)
    	rec := do(t, srv, "GET", "/api/healthz")
    	if rec.Code != 200 {
    		t.Fatalf("GET /api/healthz status = %d, want 200", rec.Code)
    	}
    	if got := rec.Body.String(); got != "OK" {
    		t.Errorf("GET /api/healthz body = %q, want %q", got, "OK")
    	}
    	if got := rec.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
    		t.Errorf("GET /api/healthz Content-Type = %q, want %q", got, "text/plain; charset=utf-8")
    	}
    	if rec := do(t, srv, "POST", "/api/healthz"); rec.Code != http.StatusMethodNotAllowed {
    		t.Errorf("POST /api/healthz status = %d, want 405 (register it for GET only)", rec.Code)
    	}
    }

    func TestStaticFiles(t *testing.T) {
    	srv := newServer(":8080", testAssets)
    	for _, tt := range []struct {
    		target   string
    		wantCode int
    		wantBody string
    	}{
    		{"/app/", 200, "<h1>Squeak home</h1>"},
    		{"/app/logo.txt", 200, "mouse logo"},
    		{"/app/css/style.css", 200, "body { color: grey; }"},
    		{"/app/nope.txt", 404, ""},
    	} {
    		rec := do(t, srv, "GET", tt.target)
    		if rec.Code != tt.wantCode {
    			t.Errorf("GET %s status = %d, want %d", tt.target, rec.Code, tt.wantCode)
    			continue
    		}
    		if tt.wantBody != "" && !strings.Contains(rec.Body.String(), tt.wantBody) {
    			t.Errorf("GET %s body = %q, want it to contain %q", tt.target, rec.Body.String(), tt.wantBody)
    		}
    	}
    	if rec := do(t, srv, "GET", "/logo.txt"); rec.Code != 404 {
    		t.Errorf("GET /logo.txt status = %d, want 404 (files live under /app/ only)", rec.Code)
    	}
    }
---

Squeak is mostly a JSON API, but it also ships a tiny web front end: an `index.html`, a
stylesheet and a logo. Serving files is so common that `net/http` has it built in.

## FileServer and FileServerFS

```go
// The classic way: serve the ./static directory from disk.
mux.Handle("/", http.FileServer(http.Dir("./static")))

// The io/fs way (Go 1.22+): any fs.FS works.
mux.Handle("/", http.FileServerFS(os.DirFS("static")))
```

Both give you a handler that:

- maps the request path to a file (`/logo.txt` becomes `logo.txt`),
- serves `index.html` for a directory, or a listing if there isn't one,
- sets `Content-Type` from the file extension,
- supports `Range` requests and `If-Modified-Since` caching, and
- refuses to escape the root with `..` tricks.

`FileServerFS` is the one to prefer, because `fs.FS` is everywhere. You can hand it an
`embed.FS` to bake your assets into the binary:

```go
//go:embed static
var static embed.FS

// Sub strips the "static/" directory so index.html sits at the root.
assets, err := fs.Sub(static, "static")
```

And in tests you can hand it a `fstest.MapFS`, an in-memory file system made from a map.

## Serving under a prefix with StripPrefix

You usually don't want files at the root, because `/api/...` lives there. Squeak serves
its front end under `/app/`. The catch: a file server maps the *entire* path, so
`/app/logo.txt` makes it look for a file called `app/logo.txt`. `http.StripPrefix`
removes the prefix before the file server sees the request:

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing/fstest"
)

func main() {
	assets := fstest.MapFS{
		"logo.txt": {Data: []byte("(\\_/)")},
	}

	broken := http.NewServeMux()
	broken.Handle("GET /app/", http.FileServerFS(assets))

	fixed := http.NewServeMux()
	fixed.Handle("GET /app/", http.StripPrefix("/app", http.FileServerFS(assets)))

	for name, mux := range map[string]*http.ServeMux{"broken": broken, "fixed": fixed} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/app/logo.txt", nil))
		fmt.Printf("%s: %d %q\n", name, rec.Code, rec.Body.String())
	}
}
```

The map order is random, but you'll see these two lines:

```
broken: 404 "404 page not found\n"
fixed: 200 "(\\_/)"
```

Strip `"/app"`, not `"/app/"`. The file server wants a path that still starts with
`/`, so `/app/logo.txt` must become `/logo.txt`.

## Gotchas

- **Directory listings.** If a directory has no `index.html`, the file server lists its
  contents. That's rarely what you want in production, so make sure every directory you
  serve has an index, or wrap the handler to reject paths ending in `/`.
- **`/index.html` redirects.** Requesting `/app/index.html` redirects to `/app/`. That's
  the file server tidying URLs, not a bug in your router.
- **Don't serve your project root.** `http.Dir(".")` happily serves your source code
  and `.env` file. Point it at a dedicated directory.

## Your turn

Build Squeak's first real server. Complete `newServer` so that it:

- answers `GET /api/healthz` with status 200, the body `OK` and a
  `Content-Type` of `text/plain; charset=utf-8` (load balancers will poll this to
  check the server is up),
- serves the files in `assets` under `GET /app/`, so `/app/logo.txt` serves `logo.txt`
  and `/app/` serves `index.html`,
- sets `ReadHeaderTimeout` to 5 seconds, `WriteTimeout` to 10 seconds and
  `IdleTimeout` to 60 seconds.

`main` drives a few requests through the server's handler with `httptest`, so **Run**
shows you what each path returns without opening a port. Register both routes for
`GET` only, as `"GET /api/healthz"`. You'll learn exactly how those method patterns work
in the next chapter, and the tests check that a `POST` to the health check gets a 405.
