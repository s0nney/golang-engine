---
title: Testing Time with synctest
quiz:
  - question: |
      Inside `synctest.Test`, how long does this test take in real time?

      ```go
      synctest.Test(t, func(t *testing.T) {
      	start := time.Now()
      	time.Sleep(10 * time.Minute)
      	t.Log(time.Since(start))
      })
      ```
    options:
      - text: About 10 minutes
      - text: A tiny fraction of a second, and it logs `10m0s`
        correct: true
      - text: It never finishes because time is frozen
      - text: It panics because sleeping isn't allowed in a bubble
    explanation: |
      Inside the bubble, time is fake. When every goroutine in the bubble is blocked (here,
      the only one is sleeping), the clock jumps forward to the next timer. The test
      finishes instantly and `time.Since` reports exactly 10 minutes.
  - question: Why does Go 1.27's `httptest.NewTestServer` work inside a synctest bubble when `httptest.NewServer` doesn't fit well?
    options:
      - text: '`NewServer` is deprecated'
      - text: '`NewTestServer` uses an in-memory network, so its goroutines only block on things inside the bubble; real sockets involve the OS, which the bubble can''t see'
        correct: true
      - text: '`NewTestServer` disables time entirely'
      - text: '`NewServer` can''t be used in tests'
    explanation: |
      Time only advances when every goroutine in the bubble is *durably* blocked, meaning
      blocked on something only another bubble goroutine can unblock. A goroutine waiting on
      a real network socket doesn't count, because the OS could wake it at any time. In-memory
      pipes do count.
exercise:
  starter: |
    package main

    import (
    	"encoding/json/v2"
    	"fmt"
    	"net"
    	"net/http"
    	"net/http/httptest"
    	"strconv"
    	"sync"
    	"time"
    )

    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	data, _ := json.Marshal(map[string]string{"error": msg})
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(code)
    	w.Write(data)
    }

    // clientIP returns the IP address part of r.RemoteAddr.
    func clientIP(r *http.Request) string {
    	host, _, err := net.SplitHostPort(r.RemoteAddr)
    	if err != nil {
    		return r.RemoteAddr
    	}
    	return host
    }

    // rateLimit allows each client IP at most limit requests per window.
    // The window starts at a client's first request (or first request after
    // the previous window ended). Extra requests get 429 Too Many Requests,
    // a JSON error body, and a Retry-After header with the whole seconds
    // (rounded up) until the window ends.
    func rateLimit(limit int, window time.Duration, next http.Handler) http.Handler {
    	// ?
    	return next
    }

    func main() {
    	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprint(w, "squeak posted")
    	})
    	h := rateLimit(3, time.Minute, ok)

    	for i := range 5 {
    		rec := httptest.NewRecorder()
    		h.ServeHTTP(rec, httptest.NewRequest("POST", "/api/squeaks", nil))
    		fmt.Printf("request %d: %d Retry-After=%q %s\n", i+1, rec.Code, rec.Header().Get("Retry-After"), rec.Body.String())
    	}
    	_, _ = strconv.Itoa, new(sync.Mutex)
    }
  solution: |
    package main

    import (
    	"encoding/json/v2"
    	"fmt"
    	"math"
    	"net"
    	"net/http"
    	"net/http/httptest"
    	"strconv"
    	"sync"
    	"time"
    )

    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	data, _ := json.Marshal(map[string]string{"error": msg})
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(code)
    	w.Write(data)
    }

    func clientIP(r *http.Request) string {
    	host, _, err := net.SplitHostPort(r.RemoteAddr)
    	if err != nil {
    		return r.RemoteAddr
    	}
    	return host
    }

    type windowCount struct {
    	start time.Time
    	count int
    }

    func rateLimit(limit int, window time.Duration, next http.Handler) http.Handler {
    	var mu sync.Mutex
    	clients := make(map[string]*windowCount)

    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		ip := clientIP(r)
    		now := time.Now()

    		mu.Lock()
    		wc, ok := clients[ip]
    		if !ok || now.Sub(wc.start) >= window {
    			wc = &windowCount{start: now}
    			clients[ip] = wc
    		}
    		allowed := wc.count < limit
    		if allowed {
    			wc.count++
    		}
    		retryAfter := wc.start.Add(window).Sub(now)
    		mu.Unlock()

    		if !allowed {
    			secs := int(math.Ceil(retryAfter.Seconds()))
    			w.Header().Set("Retry-After", strconv.Itoa(secs))
    			respondWithError(w, http.StatusTooManyRequests, "too many requests")
    			return
    		}
    		next.ServeHTTP(w, r)
    	})
    }

    func main() {
    	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprint(w, "squeak posted")
    	})
    	h := rateLimit(3, time.Minute, ok)

    	for i := range 5 {
    		rec := httptest.NewRecorder()
    		h.ServeHTTP(rec, httptest.NewRequest("POST", "/api/squeaks", nil))
    		fmt.Printf("request %d: %d Retry-After=%q %s\n", i+1, rec.Code, rec.Header().Get("Retry-After"), rec.Body.String())
    	}
    }
  tests: |
    package main

    import (
    	"encoding/json/v2"
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

    type result struct {
    	code       int
    	retryAfter string
    	body       string
    }

    func send(t *testing.T, c *http.Client) result {
    	t.Helper()
    	resp, err := c.Post("http://squeak.test/api/squeaks", "application/json", nil)
    	if err != nil {
    		t.Fatalf("request failed: %v", err)
    	}
    	defer resp.Body.Close()
    	b, _ := io.ReadAll(resp.Body)
    	return result{resp.StatusCode, resp.Header.Get("Retry-After"), string(b)}
    }

    func expect(t *testing.T, when string, got result, wantCode int, wantRetry string) {
    	t.Helper()
    	if got.code != wantCode {
    		t.Fatalf("%s: status = %d, want %d", when, got.code, wantCode)
    	}
    	if wantCode == http.StatusTooManyRequests {
    		if got.retryAfter != wantRetry {
    			t.Errorf("%s: Retry-After = %q, want %q", when, got.retryAfter, wantRetry)
    		}
    		var e map[string]string
    		if err := json.Unmarshal([]byte(got.body), &e); err != nil || e["error"] == "" {
    			t.Errorf("%s: 429 body = %q, want JSON like {\"error\":\"...\"}", when, got.body)
    		}
    	}
    }

    func TestWindowResets(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		srv := httptest.NewTestServer(t, rateLimit(3, time.Minute, okHandler))
    		c := srv.Client()

    		for i := range 3 {
    			expect(t, "request "+string(rune('1'+i)), send(t, c), 200, "")
    		}
    		expect(t, "4th request in the same minute", send(t, c), 429, "60")

    		time.Sleep(20 * time.Second)
    		expect(t, "20s later", send(t, c), 429, "40")

    		time.Sleep(39*time.Second + 500*time.Millisecond)
    		expect(t, "59.5s after the window started", send(t, c), 429, "1")

    		time.Sleep(500 * time.Millisecond)
    		expect(t, "exactly one minute later", send(t, c), 200, "")
    		expect(t, "2nd request of new window", send(t, c), 200, "")
    		expect(t, "3rd request of new window", send(t, c), 200, "")
    		expect(t, "4th request of new window", send(t, c), 429, "60")
    	})
    }

    func TestLongWindow(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		srv := httptest.NewTestServer(t, rateLimit(1, time.Hour, okHandler))
    		c := srv.Client()
    		expect(t, "first request", send(t, c), 200, "")
    		time.Sleep(59 * time.Minute)
    		expect(t, "59 minutes later", send(t, c), 429, "60")
    		time.Sleep(time.Minute)
    		expect(t, "an hour later", send(t, c), 200, "")
    	})
    }

    func TestPerClient(t *testing.T) {
    	h := rateLimit(2, time.Hour, okHandler)
    	do := func(addr string) int {
    		req := httptest.NewRequest("GET", "/", nil)
    		req.RemoteAddr = addr
    		rec := httptest.NewRecorder()
    		h.ServeHTTP(rec, req)
    		return rec.Code
    	}
    	for _, addr := range []string{"203.0.113.5:1111", "203.0.113.5:2222"} {
    		if code := do(addr); code != 200 {
    			t.Fatalf("request from %s: status = %d, want 200", addr, code)
    		}
    	}
    	if code := do("203.0.113.5:3333"); code != 429 {
    		t.Errorf("3rd request from 203.0.113.5 (different port) = %d, want 429: limit by IP, not IP:port", code)
    	}
    	if code := do("198.51.100.7:4444"); code != 200 {
    		t.Errorf("first request from another IP = %d, want 200: each IP has its own limit", code)
    	}
    	other := rateLimit(2, time.Hour, okHandler)
    	rec := httptest.NewRecorder()
    	req := httptest.NewRequest("GET", "/", nil)
    	req.RemoteAddr = "203.0.113.5:5555"
    	other.ServeHTTP(rec, req)
    	if rec.Code != 200 {
    		t.Errorf("a second rateLimit(...) handler shares state with the first; each call needs its own counters")
    	}
    }

    func TestConcurrentLimit(t *testing.T) {
    	h := rateLimit(10, time.Hour, okHandler)
    	var ok atomic.Int64
    	var wg sync.WaitGroup
    	for range 100 {
    		wg.Go(func() {
    			rec := httptest.NewRecorder()
    			h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
    			if rec.Code == 200 {
    				ok.Add(1)
    			}
    		})
    	}
    	wg.Wait()
    	if got := ok.Load(); got != 10 {
    		t.Errorf("100 simultaneous requests with limit 10: %d got through, want exactly 10", got)
    	}
    }
---

Some behaviour only shows up over time: tokens expire after an hour, refresh tokens
after a month, and rate limits reset every minute. Testing them honestly used to mean
either slow tests (`time.Sleep(time.Minute)`) or bending your code around a fake clock.
`testing/synctest` makes it painless.

## A quick refresher

You learned `testing/synctest` in the concurrency course's
[synctest lesson](/courses/learn-concurrency/testing-concurrent-code/synctest). The one
fact you need here: inside `synctest.Test(t, func(t *testing.T) { ... })`, time is fake,
and it jumps forward whenever every goroutine in the bubble is durably blocked. So
`time.Sleep(time.Hour)` returns instantly, and `time.Since` reports exactly one hour.

## Servers in the bubble

The trouble with HTTP servers used to be the network. A goroutine blocked reading a
real socket isn't "durably blocked", since the operating system could wake it at any
moment, so the bubble's clock would never advance.

Go 1.27's **`httptest.NewTestServer`** fixes that with its in-memory network. Create the
server *inside* the bubble, and the server goroutines, the client and the connection all
live there too:

```go
func TestTokenExpires(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		app := newTestApp(t)
		srv := httptest.NewTestServer(t, app.handler)
		token := app.tokenFor(t, app.pip) // issued "now", valid for one hour

		if got := getMe(t, srv.Client(), token); got != 200 {
			t.Fatalf("fresh token: status %d, want 200", got)
		}
		time.Sleep(time.Hour) // instant
		if got := getMe(t, srv.Client(), token); got != 401 {
			t.Fatalf("token after 1h: status %d, want 401", got)
		}
	})
}
```

Your production code calls plain `time.Now()`. There's no clock interface, no injected
`now` function. The bubble fakes time for everything inside it.

## Rate limiting

Squeak needs one more defence before going live: a **rate limiter**, so one client can't
flood `POST /api/squeaks` or brute-force `/api/login`. The simplest kind is a **fixed
window** counter per client:

- The first request from an IP starts a window (say one minute) with a count of 1.
- Each further request in the window adds 1. Once the count reaches the limit, requests
  get **`429 Too Many Requests`**.
- When the window has passed, the next request starts a fresh window.

A 429 should include a **`Retry-After`** header saying how many seconds to wait, so
polite clients back off exactly as long as needed.

Fixed windows are simple but allow bursts at window edges (the limit at 0:59 *and*
again at 1:00). Token buckets and sliding windows smooth that out, and
`golang.org/x/time/rate` implements a token bucket. The testing technique is the same
for all of them.

## Your task

Implement `rateLimit(limit, window, next)`:

1. Identify the client with the provided `clientIP(r)`, so different ports on the same
   IP share a limit.
2. Keep a map from IP to its current window (start time and count), protected by a
   **mutex**. Create the map *inside* `rateLimit`, so each limiter has its own state.
3. If the IP has no window, or `now.Sub(start) >= window`, start a new one at `now`.
4. If the count is below `limit`, add one and call `next`.
5. Otherwise respond with `respondWithError(w, http.StatusTooManyRequests, ...)` after
   setting `Retry-After` to the seconds left in the window, **rounded up**
   (`math.Ceil`). 20 seconds into a one-minute window, that's `40`. With 0.5 seconds left,
   it's `1`.

Use `time.Now()`. The tests run your limiter behind `httptest.NewTestServer` inside
`synctest.Test` and sleep through minutes and hours in a few milliseconds. They also fire
100 simultaneous requests and expect exactly `limit` of them to get through.

Delete the `_, _ = strconv.Itoa, new(sync.Mutex)` line once you use those packages.
