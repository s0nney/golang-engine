---
title: A Metrics Middleware
quiz:
  - question: |
      Two requests run this middleware at the same moment. What can go wrong?

      ```go
      type apiConfig struct{ hits int }

      func (cfg *apiConfig) count(next http.Handler) http.Handler {
      	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      		cfg.hits++
      		next.ServeHTTP(w, r)
      	})
      }
      ```
    options:
      - text: Nothing; `++` on an int is atomic in Go
      - text: It's a data race; both goroutines can read the same old value and one increment is lost
        correct: true
      - text: It deadlocks
      - text: The second request waits for the first to finish
    explanation: |
      `cfg.hits++` is a read, an add and a write. Two goroutines can interleave those
      steps and lose updates, and the race detector (`go test -race`) flags it. Use
      `atomic.Int64` or a mutex.
  - question: Why must `apiConfig` (which contains an `atomic.Int64`) be passed around as a pointer?
    options:
      - text: Atomics are only allowed inside pointers
      - text: Copying the struct copies the counter; each copy would count separately and `go vet` warns about it
        correct: true
      - text: Pointers make the counter faster
      - text: It doesn't matter, either way works
    explanation: |
      A copy of the struct is a separate counter. Increments to one copy are invisible to
      the others. `atomic.Int64` embeds a `noCopy` marker precisely so `go vet` catches
      accidental copies.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"sync/atomic"
    )

    type apiConfig struct {
    	hits atomic.Int64
    }

    // middlewareMetricsInc adds 1 to cfg.hits for every request, then calls next.
    func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
    	// ?
    	return next
    }

    // handleMetrics writes "Hits: <n>" as text/plain; charset=utf-8.
    func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
    	// ?
    }

    // handleReset sets the counter back to 0 and writes "Hits reset to 0".
    func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
    	// ?
    }

    func handleListSqueaks(w http.ResponseWriter, r *http.Request) {
    	fmt.Fprint(w, "all squeaks")
    }

    func handleGetSqueak(w http.ResponseWriter, r *http.Request) {
    	fmt.Fprintf(w, "squeak %s", r.PathValue("id"))
    }

    // newRouter registers Squeak's routes. The two /api/ routes must be counted
    // by the metrics middleware; the /admin/ routes must NOT be counted.
    func newRouter(cfg *apiConfig) *http.ServeMux {
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/squeaks", handleListSqueaks)
    	mux.HandleFunc("GET /api/squeaks/{id}", handleGetSqueak)
    	// ?
    	return mux
    }

    func main() {
    	cfg := &apiConfig{}
    	mux := newRouter(cfg)
    	for _, target := range []string{"/api/squeaks", "/api/squeaks/1", "/api/squeaks/2", "/admin/metrics"} {
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, httptest.NewRequest("GET", target, nil))
    		fmt.Printf("GET %-15s -> %d %q\n", target, rec.Code, rec.Body.String())
    	}
    }
  solution: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"sync/atomic"
    )

    type apiConfig struct {
    	hits atomic.Int64
    }

    func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		cfg.hits.Add(1)
    		next.ServeHTTP(w, r)
    	})
    }

    func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
    	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    	fmt.Fprintf(w, "Hits: %d", cfg.hits.Load())
    }

    func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
    	cfg.hits.Store(0)
    	fmt.Fprint(w, "Hits reset to 0")
    }

    func handleListSqueaks(w http.ResponseWriter, r *http.Request) {
    	fmt.Fprint(w, "all squeaks")
    }

    func handleGetSqueak(w http.ResponseWriter, r *http.Request) {
    	fmt.Fprintf(w, "squeak %s", r.PathValue("id"))
    }

    func newRouter(cfg *apiConfig) *http.ServeMux {
    	mux := http.NewServeMux()
    	mux.Handle("GET /api/squeaks", cfg.middlewareMetricsInc(http.HandlerFunc(handleListSqueaks)))
    	mux.Handle("GET /api/squeaks/{id}", cfg.middlewareMetricsInc(http.HandlerFunc(handleGetSqueak)))
    	mux.HandleFunc("GET /admin/metrics", cfg.handleMetrics)
    	mux.HandleFunc("POST /admin/reset", cfg.handleReset)
    	return mux
    }

    func main() {
    	cfg := &apiConfig{}
    	mux := newRouter(cfg)
    	for _, target := range []string{"/api/squeaks", "/api/squeaks/1", "/api/squeaks/2", "/admin/metrics"} {
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, httptest.NewRequest("GET", target, nil))
    		fmt.Printf("GET %-15s -> %d %q\n", target, rec.Code, rec.Body.String())
    	}
    }
  tests: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"sync"
    	"testing"
    )

    func get(h http.Handler, method, target string) *httptest.ResponseRecorder {
    	rec := httptest.NewRecorder()
    	h.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
    	return rec
    }

    func TestMiddlewareCounts(t *testing.T) {
    	cfg := &apiConfig{}
    	calls := 0
    	h := cfg.middlewareMetricsInc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		calls++
    		w.Write([]byte("squeak"))
    	}))
    	for range 3 {
    		if rec := get(h, "GET", "/api/squeaks"); rec.Body.String() != "squeak" {
    			t.Fatalf("the middleware must call next; body = %q, want %q", rec.Body.String(), "squeak")
    		}
    	}
    	if calls != 3 {
    		t.Errorf("next was called %d times, want 3", calls)
    	}
    	if got := cfg.hits.Load(); got != 3 {
    		t.Errorf("after 3 requests hits = %d, want 3", got)
    	}
    }

    func TestMetricsEndpoint(t *testing.T) {
    	cfg := &apiConfig{}
    	mux := newRouter(cfg)
    	get(mux, "GET", "/api/squeaks")
    	get(mux, "GET", "/api/squeaks/7")
    	get(mux, "GET", "/admin/metrics") // must not count itself

    	rec := get(mux, "GET", "/admin/metrics")
    	if rec.Code != 200 {
    		t.Fatalf("GET /admin/metrics status = %d, want 200", rec.Code)
    	}
    	if got := rec.Body.String(); got != "Hits: 2" {
    		t.Errorf("GET /admin/metrics body = %q, want %q (only /api/ requests count)", got, "Hits: 2")
    	}
    	if ct := rec.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
    		t.Errorf("Content-Type = %q, want %q", ct, "text/plain; charset=utf-8")
    	}
    }

    func TestReset(t *testing.T) {
    	cfg := &apiConfig{}
    	mux := newRouter(cfg)
    	for range 5 {
    		get(mux, "GET", "/api/squeaks")
    	}
    	if rec := get(mux, "GET", "/admin/reset"); rec.Code != http.StatusMethodNotAllowed {
    		t.Errorf("GET /admin/reset status = %d, want 405 (register it for POST only)", rec.Code)
    	}
    	rec := get(mux, "POST", "/admin/reset")
    	if rec.Code != 200 || rec.Body.String() != "Hits reset to 0" {
    		t.Errorf("POST /admin/reset = %d %q, want 200 %q", rec.Code, rec.Body.String(), "Hits reset to 0")
    	}
    	if got := get(mux, "GET", "/admin/metrics").Body.String(); got != "Hits: 0" {
    		t.Errorf("after reset, metrics = %q, want %q", got, "Hits: 0")
    	}
    }

    func TestConcurrentHits(t *testing.T) {
    	cfg := &apiConfig{}
    	mux := newRouter(cfg)
    	var wg sync.WaitGroup
    	for g := range 50 {
    		wg.Go(func() {
    			for i := range 40 {
    				get(mux, "GET", fmt.Sprintf("/api/squeaks/%d", g*100+i))
    			}
    		})
    	}
    	wg.Wait()
    	if got := get(mux, "GET", "/admin/metrics").Body.String(); got != "Hits: 2000" {
    		t.Errorf("after 2000 concurrent requests metrics = %q, want %q", got, "Hits: 2000")
    	}
    }
---

The last piece of Squeak's admin tooling for this chapter: an endpoint that says how
many API requests the server has handled. It's a **stateful** middleware. It doesn't
just pass requests along, it remembers something across all of them.

## Shared state, many goroutines

Every request runs on its own goroutine, and they all run the same middleware. A plain
`int` counter is a data race: `hits++` is really "read, add one, write back", and two
goroutines can both read 41 and both write 42. You learned the cures in the
concurrency course: a mutex, or an **atomic**.

For a single number, `sync/atomic`'s typed atomics are the neatest tool:

```go
type apiConfig struct {
	hits atomic.Int64
}

cfg.hits.Add(1)     // increment, safely
cfg.hits.Load()     // read, safely
cfg.hits.Store(0)   // overwrite, safely
```

The zero value is ready to use, so `&apiConfig{}` just works.

## Keep the config behind a pointer

Handlers and middleware are methods on `*apiConfig`, and you always pass `cfg` as a
pointer. If you copied the struct, you'd copy the counter too, and each copy would count
on its own. `go vet` warns you ("copies lock value") if you try.

## Counting only some routes

Middleware doesn't have to wrap the whole mux. Here you only want to count real API
traffic, not people refreshing the metrics page, so wrap just the API handlers:

```go
mux.Handle("GET /api/squeaks", cfg.middlewareMetricsInc(http.HandlerFunc(handleListSqueaks)))
```

Notice `mux.Handle` (not `HandleFunc`), because the middleware returns an
`http.Handler`, and the `http.HandlerFunc(...)` conversion that turns the plain
function into a handler the middleware can wrap.

## Admin endpoints

The two admin routes live under `/admin/`:

- `GET /admin/metrics` shows the count.
- `POST /admin/reset` sets it back to zero. It's `POST` because it *changes* state, and a
  `GET` should never do that (browsers prefetch links and crawlers follow them).

In the auth chapters these will get locked down. For now, anyone can reset your
metrics.

## Your task

1. `middlewareMetricsInc` adds 1 to `cfg.hits` and then calls `next`.
2. `handleMetrics` sets `Content-Type` to `text/plain; charset=utf-8` and writes
   `Hits: <n>`, for example `Hits: 2`.
3. `handleReset` stores 0 and writes `Hits reset to 0`.
4. In `newRouter`, wrap the two `/api/` routes with the middleware, and register
   `GET /admin/metrics` and `POST /admin/reset` **without** it.

The tests fire 2,000 requests from 50 goroutines at once and expect exactly
`Hits: 2000`, so the counter must be concurrency-safe.
