---
title: Idempotency Keys
difficulty: hard
after: authorization-and-webhooks
hints:
  - 'You can''t replay what you didn''t keep, so don''t hand `next` the real `w`. Give it your own `http.ResponseWriter` that collects the status, a header map and a `bytes.Buffer`. When `next` returns, copy all three to `w`, and keep them in a map keyed by `user + "\x00" + key` (the `\x00` stops `"ab"+"c"` from colliding with `"a"+"bc"`).'
  - 'The race to worry about is two retries arriving together. Under one mutex, look the key up and, if it''s free, store an **in-flight** entry *before* calling `next`. A second request that finds an in-flight entry answers 409 straight away. After `next` returns, lock again and either complete the entry (status < 500) or delete it (5xx) so a retry can run.'
  - 'Read the body with `io.ReadAll(http.MaxBytesReader(w, r.Body, maxBody))`, fingerprint it with `sha256.Sum256` together with the method and path, then give `next` a fresh reader: `r.Body = io.NopCloser(bytes.NewReader(body))`. Compare fingerprints before anything else when a key is already known.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"time"
    )

    type ctxKey struct{}

    // withUser returns r as seen by a handler behind the auth middleware.
    func withUser(r *http.Request, userID string) *http.Request {
    	return r.WithContext(context.WithValue(r.Context(), ctxKey{}, userID))
    }

    func userIDFrom(ctx context.Context) (string, bool) {
    	id, ok := ctx.Value(ctxKey{}).(string)
    	return id, ok
    }

    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	data, _ := json.Marshal(map[string]string{"error": msg})
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(code)
    	w.Write(data)
    }

    const (
    	maxBody   = 1 << 20
    	maxKeyLen = 255
    )

    func idempotency(ttl time.Duration, next http.Handler) http.Handler {
    	return next
    }

    func main() {
    	n := 0
    	create := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		n++
    		w.Header().Set("Content-Type", "application/json")
    		w.WriteHeader(http.StatusCreated)
    		fmt.Fprintf(w, `{"id":%d}`, n)
    	})
    	h := idempotency(24*time.Hour, create)
    	for range 2 {
    		req := httptest.NewRequest("POST", "/api/squeaks", strings.NewReader(`{"body":"on the train"}`))
    		req.Header.Set("Idempotency-Key", "5f0c3e0a")
    		rec := httptest.NewRecorder()
    		h.ServeHTTP(rec, withUser(req, "pip"))
    		fmt.Println(rec.Code, rec.Body.String(), "replayed:", rec.Header().Get("Idempotent-Replayed"))
    	}
    	fmt.Println("squeaks created:", n) // want: 1
    }
  solution: |
    package main

    import (
    	"bytes"
    	"context"
    	"crypto/sha256"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"sync"
    	"time"
    )

    type ctxKey struct{}

    func withUser(r *http.Request, userID string) *http.Request {
    	return r.WithContext(context.WithValue(r.Context(), ctxKey{}, userID))
    }

    func userIDFrom(ctx context.Context) (string, bool) {
    	id, ok := ctx.Value(ctxKey{}).(string)
    	return id, ok
    }

    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	data, _ := json.Marshal(map[string]string{"error": msg})
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(code)
    	w.Write(data)
    }

    const (
    	maxBody   = 1 << 20
    	maxKeyLen = 255
    )

    // capture is a ResponseWriter that keeps the response instead of sending it.
    type capture struct {
    	header http.Header
    	status int
    	body   bytes.Buffer
    }

    func (c *capture) Header() http.Header { return c.header }

    func (c *capture) WriteHeader(code int) {
    	if c.status == 0 {
    		c.status = code
    	}
    }

    func (c *capture) Write(b []byte) (int, error) {
    	c.WriteHeader(http.StatusOK)
    	return c.body.Write(b)
    }

    type entry struct {
    	fingerprint [32]byte
    	done        bool
    	status      int
    	header      http.Header
    	body        []byte
    	expires     time.Time
    }

    func idempotency(ttl time.Duration, next http.Handler) http.Handler {
    	var mu sync.Mutex
    	entries := make(map[string]*entry)

    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		key := r.Header.Get("Idempotency-Key")
    		user, ok := userIDFrom(r.Context())
    		if r.Method != http.MethodPost || key == "" || !ok {
    			next.ServeHTTP(w, r)
    			return
    		}
    		if len(key) > maxKeyLen {
    			respondWithError(w, http.StatusBadRequest, "Idempotency-Key is too long")
    			return
    		}
    		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBody))
    		if err != nil {
    			if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
    				respondWithError(w, http.StatusRequestEntityTooLarge, "request body too large")
    				return
    			}
    			respondWithError(w, http.StatusBadRequest, "couldn't read request body")
    			return
    		}
    		r.Body = io.NopCloser(bytes.NewReader(body))
    		fp := sha256.Sum256(fmt.Appendf(nil, "%s %s\n%s", r.Method, r.URL.Path, body))
    		scope := user + "\x00" + key

    		mu.Lock()
    		if e, found := entries[scope]; found && e.done && !time.Now().Before(e.expires) {
    			delete(entries, scope)
    		}
    		if e, found := entries[scope]; found {
    			var (
    				done   = e.done
    				status = e.status
    				header = e.header
    				stored = e.body
    			)
    			mu.Unlock()
    			switch {
    			case e.fingerprint != fp:
    				respondWithError(w, http.StatusUnprocessableEntity, "Idempotency-Key was already used for a different request")
    			case !done:
    				respondWithError(w, http.StatusConflict, "a request with this Idempotency-Key is still in progress")
    			default:
    				for k, v := range header {
    					w.Header()[k] = v
    				}
    				w.Header().Set("Idempotent-Replayed", "true")
    				w.WriteHeader(status)
    				w.Write(stored)
    			}
    			return
    		}
    		e := &entry{fingerprint: fp}
    		entries[scope] = e
    		mu.Unlock()

    		finished := false
    		defer func() {
    			if !finished { // next panicked: free the key for a retry
    				mu.Lock()
    				delete(entries, scope)
    				mu.Unlock()
    			}
    		}()
    		c := &capture{header: make(http.Header)}
    		next.ServeHTTP(c, r)
    		finished = true
    		if c.status == 0 {
    			c.status = http.StatusOK
    		}

    		mu.Lock()
    		if c.status >= 500 {
    			delete(entries, scope)
    		} else {
    			e.done, e.status, e.header, e.body = true, c.status, c.header.Clone(), bytes.Clone(c.body.Bytes())
    			e.expires = time.Now().Add(ttl)
    		}
    		mu.Unlock()

    		for k, v := range c.header {
    			w.Header()[k] = v
    		}
    		w.WriteHeader(c.status)
    		w.Write(c.body.Bytes())
    	})
    }

    func main() {
    	n := 0
    	create := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		n++
    		w.Header().Set("Content-Type", "application/json")
    		w.WriteHeader(http.StatusCreated)
    		fmt.Fprintf(w, `{"id":%d}`, n)
    	})
    	h := idempotency(24*time.Hour, create)
    	for range 2 {
    		req := httptest.NewRequest("POST", "/api/squeaks", strings.NewReader(`{"body":"on the train"}`))
    		req.Header.Set("Idempotency-Key", "5f0c3e0a")
    		rec := httptest.NewRecorder()
    		h.ServeHTTP(rec, withUser(req, "pip"))
    		fmt.Println(rec.Code, rec.Body.String(), "replayed:", rec.Header().Get("Idempotent-Replayed"))
    	}
    	fmt.Println("squeaks created:", n)
    }
  tests: |
    package main

    import (
    	"encoding/json/v2"
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"sync"
    	"sync/atomic"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // squeakAPI creates a squeak per call. It echoes the body it received.
    type squeakAPI struct {
    	calls atomic.Int64
    	gate  chan struct{} // if non-nil, each call waits for it to close
    }

    func (a *squeakAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    	n := a.calls.Add(1)
    	if a.gate != nil {
    		<-a.gate
    	}
    	body, _ := io.ReadAll(r.Body)
    	if strings.Contains(string(body), "boom") {
    		http.Error(w, "database unavailable", http.StatusServiceUnavailable)
    		return
    	}
    	w.Header().Set("Content-Type", "application/json")
    	w.Header().Set("Location", fmt.Sprintf("/api/squeaks/%d", n))
    	w.WriteHeader(http.StatusCreated)
    	fmt.Fprintf(w, `{"id":%d,"got":%q}`, n, body)
    }

    func post(h http.Handler, user, key, path, body string) *httptest.ResponseRecorder {
    	req := httptest.NewRequest("POST", path, strings.NewReader(body))
    	if key != "" {
    		req.Header.Set("Idempotency-Key", key)
    	}
    	if user != "" {
    		req = withUser(req, user)
    	}
    	rec := httptest.NewRecorder()
    	h.ServeHTTP(rec, req)
    	return rec
    }

    func wantError(t *testing.T, what string, rec *httptest.ResponseRecorder, code int) {
    	t.Helper()
    	if rec.Code != code {
    		t.Errorf("%s: status %d %s, want %d", what, rec.Code, rec.Body.String(), code)
    		return
    	}
    	var e map[string]string
    	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil || e["error"] == "" {
    		t.Errorf("%s: body %s, want a JSON {\"error\":\"...\"}", what, rec.Body.String())
    	}
    }

    const squeak = `{"body":"on the train"}`

    func TestReplay(t *testing.T) {
    	api := &squeakAPI{}
    	h := idempotency(24*time.Hour, api)
    	first := post(h, "pip", "k1", "/api/squeaks", squeak)
    	wantBody := `{"id":1,"got":"{\"body\":\"on the train\"}"}`
    	if first.Code != 201 || first.Body.String() != wantBody {
    		t.Fatalf("first request: %d %s, want 201 %s (the handler must still see the request body)", first.Code, first.Body.String(), wantBody)
    	}
    	if first.Header().Get("Location") != "/api/squeaks/1" || first.Header().Get("Content-Type") != "application/json" {
    		t.Errorf("first request: headers %v, want the handler's Location and Content-Type", first.Header())
    	}
    	if v := first.Header().Get("Idempotent-Replayed"); v != "" {
    		t.Errorf("first request: Idempotent-Replayed = %q, want it absent", v)
    	}
    	for i := range 3 {
    		again := post(h, "pip", "k1", "/api/squeaks", squeak)
    		if again.Code != 201 || again.Body.String() != wantBody {
    			t.Errorf("retry %d: %d %s, want the stored 201 %s", i+1, again.Code, again.Body.String(), wantBody)
    		}
    		if again.Header().Get("Location") != "/api/squeaks/1" || again.Header().Get("Content-Type") != "application/json" {
    			t.Errorf("retry %d: headers %v, want the stored Location and Content-Type", i+1, again.Header())
    		}
    		if v := again.Header().Get("Idempotent-Replayed"); v != "true" {
    			t.Errorf("retry %d: Idempotent-Replayed = %q, want \"true\"", i+1, v)
    		}
    	}
    	if n := api.calls.Load(); n != 1 {
    		t.Errorf("handler ran %d times for one key, want 1", n)
    	}
    	if rec := post(h, "pip", "k2", "/api/squeaks", squeak); rec.Code != 201 || api.calls.Load() != 2 {
    		t.Errorf("a new key must run the handler again: got %d, handler calls %d", rec.Code, api.calls.Load())
    	}
    }

    func TestPassThrough(t *testing.T) {
    	api := &squeakAPI{}
    	h := idempotency(time.Hour, api)
    	post(h, "pip", "", "/api/squeaks", squeak)
    	post(h, "pip", "", "/api/squeaks", squeak)
    	if n := api.calls.Load(); n != 2 {
    		t.Errorf("two POSTs without Idempotency-Key: handler ran %d times, want 2", n)
    	}
    	post(h, "", "k1", "/api/squeaks", squeak)
    	post(h, "", "k1", "/api/squeaks", squeak)
    	if n := api.calls.Load(); n != 4 {
    		t.Errorf("two POSTs with a key but no user: handler ran %d times in total, want 4 (pass them through)", n)
    	}
    	for range 2 {
    		req := withUser(httptest.NewRequest("GET", "/api/squeaks", nil), "pip")
    		req.Header.Set("Idempotency-Key", "k1")
    		h.ServeHTTP(httptest.NewRecorder(), req)
    	}
    	if n := api.calls.Load(); n != 6 {
    		t.Errorf("GET requests with a key: handler ran %d times in total, want 6 (only POST is deduplicated)", n)
    	}
    }

    func TestScopedPerUser(t *testing.T) {
    	api := &squeakAPI{}
    	h := idempotency(time.Hour, api)
    	a := post(h, "pip", "shared-key", "/api/squeaks", squeak)
    	b := post(h, "whiskers", "shared-key", "/api/squeaks", squeak)
    	if b.Code != 201 || b.Body.String() == a.Body.String() || b.Header().Get("Idempotent-Replayed") != "" {
    		t.Errorf("whiskers reusing pip's key got %d %s: keys are per user, so whiskers must get a fresh result, never pip's", b.Code, b.Body.String())
    	}
    	c := post(h, "pi", "pshared-key", "/api/squeaks", squeak)
    	if c.Header().Get("Idempotent-Replayed") != "" {
    		t.Errorf(`user "pi" with key "pshared-key" got pip's replay: join user and key with a separator`)
    	}
    }

    func TestKeyReusedForDifferentRequest(t *testing.T) {
    	api := &squeakAPI{}
    	h := idempotency(time.Hour, api)
    	post(h, "pip", "k1", "/api/squeaks", squeak)
    	wantError(t, "same key, different body", post(h, "pip", "k1", "/api/squeaks", `{"body":"different"}`), 422)
    	wantError(t, "same key, different path", post(h, "pip", "k1", "/api/reports", squeak), 422)
    	if n := api.calls.Load(); n != 1 {
    		t.Errorf("handler ran %d times, want 1: a mismatched request must not run", n)
    	}
    	if rec := post(h, "pip", "k1", "/api/squeaks", squeak); rec.Code != 201 || rec.Header().Get("Idempotent-Replayed") != "true" {
    		t.Errorf("the original request still replays after a mismatch: got %d", rec.Code)
    	}
    }

    func TestServerErrorsAreNotStored(t *testing.T) {
    	api := &squeakAPI{}
    	h := idempotency(time.Hour, api)
    	if rec := post(h, "pip", "k1", "/api/squeaks", `{"body":"boom"}`); rec.Code != 503 {
    		t.Fatalf("handler failure: status %d, want the handler's 503 passed on", rec.Code)
    	}
    	if rec := post(h, "pip", "k1", "/api/squeaks", `{"body":"boom"}`); rec.Code != 503 || rec.Header().Get("Idempotent-Replayed") != "" {
    		t.Errorf("retry after a 503 got a replay: don't store 5xx responses, so a retry can succeed")
    	}
    	if n := api.calls.Load(); n != 2 {
    		t.Errorf("handler ran %d times for two tries after 5xx, want 2", n)
    	}
    }

    func TestLimits(t *testing.T) {
    	api := &squeakAPI{}
    	h := idempotency(time.Hour, api)
    	wantError(t, "256-byte key", post(h, "pip", strings.Repeat("k", 256), "/api/squeaks", squeak), 400)
    	if rec := post(h, "pip", strings.Repeat("k", 255), "/api/squeaks", squeak); rec.Code != 201 {
    		t.Errorf("255-byte key: status %d, want 201", rec.Code)
    	}
    	wantError(t, "body over 1 MiB", post(h, "pip", "big", "/api/squeaks", strings.Repeat("x", maxBody+1)), 413)
    	if n := api.calls.Load(); n != 1 {
    		t.Errorf("handler ran %d times, want 1 (rejected requests must not reach it)", n)
    	}
    }

    func TestInFlightDuplicate(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		api := &squeakAPI{gate: make(chan struct{})}
    		h := idempotency(time.Hour, api)
    		results := make(chan *httptest.ResponseRecorder, 20)
    		for range 20 {
    			go func() { results <- post(h, "pip", "k1", "/api/squeaks", squeak) }()
    		}
    		synctest.Wait() // one request is inside the handler; the rest have answered
    		if n := api.calls.Load(); n != 1 {
    			close(api.gate) // let the extra handler calls finish before failing
    			for range 20 {
    				<-results
    			}
    			t.Fatalf("20 simultaneous requests with one key: handler entered %d times, want 1 (reserve the key before calling next)", n)
    		}
    		for i := range 19 {
    			wantError(t, fmt.Sprintf("duplicate %d while the first is running", i+1), <-results, 409)
    		}
    		wantError(t, "different body while the first is running", post(h, "pip", "k1", "/api/squeaks", `{"body":"other"}`), 422)
    		close(api.gate)
    		if first := <-results; first.Code != 201 {
    			t.Errorf("the first request: status %d, want 201", first.Code)
    		}
    		if rec := post(h, "pip", "k1", "/api/squeaks", squeak); rec.Code != 201 || rec.Header().Get("Idempotent-Replayed") != "true" {
    			t.Errorf("retry after the first finished: %d, replayed %q; want the stored 201", rec.Code, rec.Header().Get("Idempotent-Replayed"))
    		}
    	})
    }

    func TestExpiry(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		api := &squeakAPI{}
    		h := idempotency(24*time.Hour, api)
    		post(h, "pip", "k1", "/api/squeaks", squeak)
    		time.Sleep(24*time.Hour - time.Second)
    		if rec := post(h, "pip", "k1", "/api/squeaks", squeak); rec.Header().Get("Idempotent-Replayed") != "true" {
    			t.Errorf("23h59m59s later: status %d without replay, want the stored response", rec.Code)
    		}
    		time.Sleep(time.Second)
    		rec := post(h, "pip", "k1", "/api/squeaks", squeak)
    		if rec.Code != 201 || rec.Header().Get("Idempotent-Replayed") != "" || api.calls.Load() != 2 {
    			t.Errorf("24h later: status %d, replayed %q, handler calls %d; want a fresh run (the key has expired)", rec.Code, rec.Header().Get("Idempotent-Replayed"), api.calls.Load())
    		}
    	})
    }

    func TestConcurrentKeys(t *testing.T) {
    	api := &squeakAPI{}
    	h := idempotency(time.Hour, api)
    	var wg sync.WaitGroup
    	for i := range 50 {
    		wg.Go(func() { post(h, fmt.Sprint("user", i%5), fmt.Sprint("key", i%10), "/api/squeaks", squeak) })
    	}
    	wg.Wait()
    	if n := api.calls.Load(); n != 10 {
    		t.Errorf("50 requests over 10 distinct (user, key) pairs, in no particular order: handler ran %d times, want 10", n)
    	}
    }
---

Mice post squeaks from the underground train. The request reaches Squeak, the
squeak is created, and then the tunnel eats the response. The app retries, and
now there are two identical squeaks. The fix is an **`Idempotency-Key`** header:
the app picks a random key for each squeak it's trying to post and sends the
same key on every retry. The server remembers what it answered the first time.

Write `idempotency(ttl, next)`, a middleware that handles this for any `POST`
handler:

1. Requests that aren't `POST`, have no `Idempotency-Key` header, or have no
   user (`userIDFrom`) go straight to `next`.
2. Keys longer than `maxKeyLen` bytes get `400`. Bodies over `maxBody` bytes get
   `413`. Neither reaches `next`.
3. Keys belong to a user: pip's key never matches whiskers' key.
4. A key's **first** request runs `next`. Its response (status, headers and
   body) is stored and also sent to the client as usual.
5. A later request with the same user and key:
   - with a different method, path or body: `422`, without running `next`.
   - while the first request is **still running**: `409`.
   - after it finished: the stored status, headers and body, plus the header
     `Idempotent-Replayed: true`. `next` doesn't run again.
6. Responses with a status of `500` or more aren't stored, so a retry runs
   `next` again.
7. A stored response is replayed for less than `ttl`. From `ttl` after it was
   stored, the key is forgotten and a request with it runs `next` afresh.

All the error responses use `respondWithError`.

## Example

```
POST /api/squeaks   Idempotency-Key: 5f0c3e0a   {"body":"on the train"}
201 {"id":1}

POST /api/squeaks   Idempotency-Key: 5f0c3e0a   {"body":"on the train"}   (retry)
201 {"id":1}        Idempotent-Replayed: true
```

## Constraints

- Twenty identical requests arriving at once must run `next` exactly once: one
  gets the real response and the others get `409`.
- `next` must still be able to read the request body.
- The tests use `testing/synctest`, so plain `time.Now()` is fine for the
  expiry: they skip through 24 hours instantly.
