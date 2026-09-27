---
title: Token Buckets
difficulty: hard
after: testing-servers
hints:
  - 'You don''t need a goroutine that adds tokens every tick. Store, per client, how full the bucket was and **when** you last looked. On each request, add the time that has passed since then, cap it at full, and move "last looked" to now.'
  - 'Floats make `Retry-After` wobble (`0.9999999` tokens). Measure the bucket in **time** instead: `credit time.Duration`, where one token is `refill` worth of credit and a full bucket is `capacity * refill`. Refill is `credit = min(credit + now.Sub(last), full)`, a request costs `refill`, remaining tokens are `credit / refill`, and the wait for the next token is `refill - credit`.'
  - 'The client key is `"user:" + id` when `userIDFrom` finds a user, else `"ip:" + clientIP(r)`. The prefixes keep a user called `203.0.113.5` apart from that IP address. Keep the map and the arithmetic under one mutex, created inside `rateLimit`.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"fmt"
    	"net"
    	"net/http"
    	"net/http/httptest"
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

    func clientIP(r *http.Request) string {
    	host, _, err := net.SplitHostPort(r.RemoteAddr)
    	if err != nil {
    		return r.RemoteAddr
    	}
    	return host
    }

    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	data, _ := json.Marshal(map[string]string{"error": msg})
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(code)
    	w.Write(data)
    }

    func rateLimit(capacity int, refill time.Duration, next http.Handler) http.Handler {
    	return next
    }

    func main() {
    	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprint(w, "squeaked")
    	})
    	h := rateLimit(3, 10*time.Second, ok)
    	for i := range 5 {
    		rec := httptest.NewRecorder()
    		h.ServeHTTP(rec, withUser(httptest.NewRequest("POST", "/api/squeaks", nil), "pip"))
    		fmt.Printf("request %d: %d remaining=%q retry-after=%q\n", i+1, rec.Code,
    			rec.Header().Get("X-RateLimit-Remaining"), rec.Header().Get("Retry-After"))
    	}
    }
  solution: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"fmt"
    	"net"
    	"net/http"
    	"net/http/httptest"
    	"strconv"
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

    func clientIP(r *http.Request) string {
    	host, _, err := net.SplitHostPort(r.RemoteAddr)
    	if err != nil {
    		return r.RemoteAddr
    	}
    	return host
    }

    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	data, _ := json.Marshal(map[string]string{"error": msg})
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(code)
    	w.Write(data)
    }

    // bucket measures its tokens in time: one token is worth refill.
    type bucket struct {
    	credit time.Duration
    	last   time.Time
    }

    func clientKey(r *http.Request) string {
    	if id, ok := userIDFrom(r.Context()); ok {
    		return "user:" + id
    	}
    	return "ip:" + clientIP(r)
    }

    func rateLimit(capacity int, refill time.Duration, next http.Handler) http.Handler {
    	full := time.Duration(capacity) * refill
    	var mu sync.Mutex
    	buckets := make(map[string]*bucket)

    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		key := clientKey(r)
    		now := time.Now()

    		mu.Lock()
    		b, ok := buckets[key]
    		if !ok {
    			b = &bucket{credit: full, last: now}
    			buckets[key] = b
    		}
    		b.credit = min(b.credit+now.Sub(b.last), full)
    		b.last = now
    		allowed := b.credit >= refill
    		if allowed {
    			b.credit -= refill
    		}
    		remaining := int(b.credit / refill)
    		wait := refill - b.credit
    		mu.Unlock()

    		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
    		if !allowed {
    			secs := (wait + time.Second - 1) / time.Second
    			w.Header().Set("Retry-After", strconv.Itoa(int(secs)))
    			respondWithError(w, http.StatusTooManyRequests, "slow down")
    			return
    		}
    		next.ServeHTTP(w, r)
    	})
    }

    func main() {
    	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprint(w, "squeaked")
    	})
    	h := rateLimit(3, 10*time.Second, ok)
    	for i := range 5 {
    		rec := httptest.NewRecorder()
    		h.ServeHTTP(rec, withUser(httptest.NewRequest("POST", "/api/squeaks", nil), "pip"))
    		fmt.Printf("request %d: %d remaining=%q retry-after=%q\n", i+1, rec.Code,
    			rec.Header().Get("X-RateLimit-Remaining"), rec.Header().Get("Retry-After"))
    	}
    }
  tests: |
    package main

    import (
    	"encoding/json/v2"
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"sync"
    	"sync/atomic"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    	io.WriteString(w, "ok")
    })

    // testAuth plays the auth middleware: X-Test-User becomes the logged-in user.
    func testAuth(next http.Handler) http.Handler {
    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		if u := r.Header.Get("X-Test-User"); u != "" {
    			r = withUser(r, u)
    		}
    		next.ServeHTTP(w, r)
    	})
    }

    type client struct {
    	t    *testing.T
    	http *http.Client
    	user string
    }

    func (c client) expect(when string, wantCode int, wantRemaining, wantRetry string) {
    	c.t.Helper()
    	req, _ := http.NewRequest("POST", "http://squeak.test/api/squeaks", nil)
    	req.Header.Set("X-Test-User", c.user)
    	resp, err := c.http.Do(req)
    	if err != nil {
    		c.t.Fatalf("%s: request failed: %v", when, err)
    	}
    	defer resp.Body.Close()
    	body, _ := io.ReadAll(resp.Body)
    	if resp.StatusCode != wantCode {
    		c.t.Fatalf("%s (%s): status %d, want %d", when, c.user, resp.StatusCode, wantCode)
    	}
    	if got := resp.Header.Get("X-RateLimit-Remaining"); got != wantRemaining {
    		c.t.Errorf("%s (%s): X-RateLimit-Remaining = %q, want %q", when, c.user, got, wantRemaining)
    	}
    	if got := resp.Header.Get("Retry-After"); got != wantRetry {
    		c.t.Errorf("%s (%s): Retry-After = %q, want %q", when, c.user, got, wantRetry)
    	}
    	if wantCode == 429 {
    		var e map[string]string
    		if err := json.Unmarshal(body, &e); err != nil || e["error"] == "" {
    			c.t.Errorf("%s: 429 body = %q, want a JSON {\"error\":\"...\"}", when, body)
    		}
    	}
    }

    func TestBucketDrainsAndRefills(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		srv := httptest.NewTestServer(t, testAuth(rateLimit(3, 10*time.Second, okHandler)))
    		pip := client{t, srv.Client(), "pip"}

    		pip.expect("1st request", 200, "2", "")
    		pip.expect("2nd request", 200, "1", "")
    		pip.expect("3rd request", 200, "0", "")
    		pip.expect("4th request, bucket empty", 429, "0", "10")

    		time.Sleep(4 * time.Second)
    		pip.expect("4s later (0.4 tokens)", 429, "0", "6")

    		time.Sleep(6 * time.Second)
    		pip.expect("10s after emptying (1 token)", 200, "0", "")
    		pip.expect("right after", 429, "0", "10")

    		time.Sleep(25 * time.Second)
    		pip.expect("25s later (2.5 tokens)", 200, "1", "")
    		pip.expect("then (1.5 tokens)", 200, "0", "")
    		pip.expect("then (0.5 tokens)", 429, "0", "5")
    	})
    }

    func TestBucketIsCapped(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		srv := httptest.NewTestServer(t, testAuth(rateLimit(3, 10*time.Second, okHandler)))
    		pip := client{t, srv.Client(), "pip"}
    		pip.expect("first request", 200, "2", "")
    		time.Sleep(time.Hour)
    		pip.expect("an hour later", 200, "2", "")
    		pip.expect("then", 200, "1", "")
    		pip.expect("then", 200, "0", "")
    		pip.expect("then", 429, "0", "10")
    	})
    }

    func TestRetryAfterRoundsUp(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		srv := httptest.NewTestServer(t, testAuth(rateLimit(1, 1500*time.Millisecond, okHandler)))
    		pip := client{t, srv.Client(), "pip"}
    		pip.expect("first", 200, "0", "")
    		pip.expect("second (1.5s to wait)", 429, "0", "2")
    		time.Sleep(1400 * time.Millisecond)
    		pip.expect("1.4s later (0.1s to wait)", 429, "0", "1")
    		time.Sleep(100 * time.Millisecond)
    		pip.expect("1.5s later", 200, "0", "")
    	})
    }

    func TestPerClient(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		srv := httptest.NewTestServer(t, testAuth(rateLimit(2, time.Minute, okHandler)))
    		pip := client{t, srv.Client(), "pip"}
    		whiskers := client{t, srv.Client(), "whiskers"}
    		pip.expect("pip 1", 200, "1", "")
    		pip.expect("pip 2", 200, "0", "")
    		pip.expect("pip 3", 429, "0", "60")
    		whiskers.expect("whiskers 1, same connection as pip", 200, "1", "")
    	})

    	h := testAuth(rateLimit(1, time.Hour, okHandler))
    	send := func(user, addr string) int {
    		req := httptest.NewRequest("GET", "/", nil)
    		req.RemoteAddr = addr
    		if user != "" {
    			req.Header.Set("X-Test-User", user)
    		}
    		rec := httptest.NewRecorder()
    		h.ServeHTTP(rec, req)
    		return rec.Code
    	}
    	if code := send("", "203.0.113.5:1111"); code != 200 {
    		t.Fatalf("first anonymous request from 203.0.113.5: %d, want 200", code)
    	}
    	if code := send("", "203.0.113.5:2222"); code != 429 {
    		t.Errorf("second anonymous request from 203.0.113.5 (another port): %d, want 429 (anonymous clients are limited by IP)", code)
    	}
    	if code := send("203.0.113.5", "198.51.100.1:1"); code != 200 {
    		t.Errorf("a user whose ID is \"203.0.113.5\": %d, want 200 (users and IPs are separate keys)", code)
    	}
    	if code := send("pip", "203.0.113.5:3333"); code != 200 {
    		t.Errorf("logged-in pip from the busy IP: %d, want 200 (users are limited by user ID, not IP)", code)
    	}
    	if code := send("pip", "198.51.100.9:4444"); code != 429 {
    		t.Errorf("pip again from another IP: %d, want 429 (pip's bucket follows pip)", code)
    	}
    	other := testAuth(rateLimit(1, time.Hour, okHandler))
    	rec := httptest.NewRecorder()
    	req := httptest.NewRequest("GET", "/", nil)
    	req.Header.Set("X-Test-User", "pip")
    	other.ServeHTTP(rec, req)
    	if rec.Code != 200 {
    		t.Errorf("a second rateLimit(...) shares buckets with the first; each needs its own")
    	}
    }

    func TestConcurrentRequests(t *testing.T) {
    	h := testAuth(rateLimit(10, time.Hour, okHandler))
    	var passed atomic.Int64
    	var wg sync.WaitGroup
    	for i := range 100 {
    		wg.Go(func() {
    			req := httptest.NewRequest("POST", "/api/squeaks", nil)
    			req.Header.Set("X-Test-User", fmt.Sprint("user", i%2))
    			rec := httptest.NewRecorder()
    			h.ServeHTTP(rec, req)
    			if rec.Code == 200 {
    				passed.Add(1)
    			}
    		})
    	}
    	wg.Wait()
    	if got := passed.Load(); got != 20 {
    		t.Errorf("100 simultaneous requests from 2 users, capacity 10 each: %d passed, want exactly 20", got)
    	}
    }
---

The fixed-window limiter from the testing chapter lets a client burst twice
its limit around a window boundary, and then locks it out for a whole minute.
Squeak's new limiter is a **token bucket**: every client has a bucket of up to
`capacity` tokens, each request takes one, and tokens drip back in steadily,
one every `refill`. Short bursts are fine; sustained floods aren't.

Write `rateLimit(capacity, refill, next)`:

- **Who is the client?** The logged-in user (`userIDFrom`), or, for anonymous
  requests, the IP from `clientIP(r)`. A user and an IP never share a bucket,
  even if the user ID looks like an IP.
- A client's first request finds a **full** bucket.
- Tokens refill continuously: after half a `refill`, a bucket has gained half
  a token. A bucket never holds more than `capacity`.
- If the bucket holds at least one whole token, take one and call `next`.
- Otherwise respond `429` with `respondWithError`, and a `Retry-After` header:
  the whole seconds (rounded **up**) until one token is available.
- Every response from the limiter carries `X-RateLimit-Remaining`: the whole
  tokens left after this request (rounded down).

## Example

With `capacity` 3 and `refill` 10 seconds, for one user:

```
t=0s    200  X-RateLimit-Remaining: 2
t=0s    200  X-RateLimit-Remaining: 1
t=0s    200  X-RateLimit-Remaining: 0
t=0s    429  X-RateLimit-Remaining: 0  Retry-After: 10
t=4s    429  X-RateLimit-Remaining: 0  Retry-After: 6
t=10s   200  X-RateLimit-Remaining: 0
t=35s   200  X-RateLimit-Remaining: 1      (2.5 tokens had built up)
```

## Constraints

- Use `time.Now()`. The tests run your limiter behind `httptest.NewTestServer`
  inside `testing/synctest`, so hours pass instantly.
- 100 simultaneous requests from two users with `capacity` 10 must let exactly
  20 through.
- Each call to `rateLimit` has its own buckets.
