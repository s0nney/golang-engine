---
title: Retries with Backoff and Jitter
quiz:
  - question: Why add random **jitter** to the backoff delay?
    options:
      - text: To make the code harder to predict for attackers
      - text: So that many clients that failed at the same moment don't all retry at the same moment too
        correct: true
      - text: Because `time.Sleep` isn't precise
    explanation: |
      Without jitter, a thousand clients knocked over by the same outage retry in
      lockstep after 1s, 2s, 4s, and hit the recovering server in synchronized waves.
      Randomness spreads them out.
  - question: 'With `base := 100 * time.Millisecond`, what is `base << 3`?'
    options:
      - text: 300ms
      - text: 800ms
        correct: true
      - text: 1.6s
    explanation: |
      Shifting left by 3 multiplies by 2³ = 8. Retry *n* waits `base << (n-1)`: 100ms,
      200ms, 400ms, 800ms, and so on.
  - question: Why wait with `select` on `ctx.Done()` and a timer instead of `time.Sleep`?
    options:
      - text: '`time.Sleep` isn''t allowed in goroutines'
      - text: So a cancelled or expired context stops the wait immediately instead of sleeping it out
        correct: true
      - text: Timers are more accurate than `time.Sleep`
    explanation: |
      A backoff can reach many seconds. If the user presses Ctrl+C, `trackr` should stop
      now, not after the sleep finishes.
exercise:
  starter: |
    package main

    import (
    	"context"
    	"fmt"
    	"math/rand/v2"
    	"net/http"
    	"net/http/httptest"
    	"sync/atomic"
    	"time"
    )

    // retryable reports whether a response with this status is worth retrying.
    func retryable(status int) bool {
    	switch status {
    	case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusBadGateway,
    		http.StatusServiceUnavailable, http.StatusGatewayTimeout:
    		return true
    	}
    	return false
    }

    // backoff returns how long to wait before retry number attempt (1, 2, 3, ...):
    // exponential growth from base, with full jitter.
    func backoff(attempt int, base time.Duration) time.Duration {
    	d := base << (attempt - 1)
    	return rand.N(d) + 1
    }

    // getWithRetry GETs url and returns the body of the first 2xx response.
    //
    //   - A network error or a retryable status means: wait backoff(attempt, baseDelay)
    //     and try again, up to maxAttempts attempts in total.
    //   - Any other non-2xx status fails immediately, without retrying.
    //   - While waiting, give up straight away if ctx is done, returning ctx.Err().
    //   - After the last failed attempt, return an error.
    func getWithRetry(ctx context.Context, client *http.Client, url string, maxAttempts int, baseDelay time.Duration) ([]byte, error) {
    	// ?
    	return nil, nil
    }

    func main() {
    	var calls atomic.Int32
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		n := calls.Add(1)
    		if n < 3 {
    			fmt.Printf("server: attempt %d -> 503\n", n)
    			http.Error(w, "busy", http.StatusServiceUnavailable)
    			return
    		}
    		fmt.Printf("server: attempt %d -> 200\n", n)
    		fmt.Fprint(w, `{"status":"ok"}`)
    	}))
    	defer srv.Close()

    	body, err := getWithRetry(context.Background(), srv.Client(), srv.URL+"/health", 5, 10*time.Millisecond)
    	fmt.Printf("body=%s err=%v calls=%d\n", body, err, calls.Load())
    }
  solution: |
    package main

    import (
    	"context"
    	"fmt"
    	"io"
    	"math/rand/v2"
    	"net/http"
    	"net/http/httptest"
    	"sync/atomic"
    	"time"
    )

    // retryable reports whether a response with this status is worth retrying.
    func retryable(status int) bool {
    	switch status {
    	case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusBadGateway,
    		http.StatusServiceUnavailable, http.StatusGatewayTimeout:
    		return true
    	}
    	return false
    }

    // backoff returns how long to wait before retry number attempt (1, 2, 3, ...):
    // exponential growth from base, with full jitter.
    func backoff(attempt int, base time.Duration) time.Duration {
    	d := base << (attempt - 1)
    	return rand.N(d) + 1
    }

    func getWithRetry(ctx context.Context, client *http.Client, url string, maxAttempts int, baseDelay time.Duration) ([]byte, error) {
    	var lastErr error
    	for attempt := 1; attempt <= maxAttempts; attempt++ {
    		if attempt > 1 {
    			timer := time.NewTimer(backoff(attempt-1, baseDelay))
    			select {
    			case <-ctx.Done():
    				timer.Stop()
    				return nil, ctx.Err()
    			case <-timer.C:
    			}
    		}

    		body, retry, err := try(ctx, client, url)
    		if err == nil {
    			return body, nil
    		}
    		if !retry {
    			return nil, err
    		}
    		lastErr = err
    	}
    	return nil, fmt.Errorf("giving up after %d attempts: %w", maxAttempts, lastErr)
    }

    // try makes one attempt. retry reports whether a failure is worth retrying.
    func try(ctx context.Context, client *http.Client, url string) (body []byte, retry bool, err error) {
    	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    	if err != nil {
    		return nil, false, err
    	}
    	resp, err := client.Do(req)
    	if err != nil {
    		return nil, ctx.Err() == nil, err
    	}
    	defer resp.Body.Close()

    	if resp.StatusCode < 200 || resp.StatusCode > 299 {
    		return nil, retryable(resp.StatusCode), fmt.Errorf("GET %s: %s", url, resp.Status)
    	}
    	body, err = io.ReadAll(resp.Body)
    	return body, err != nil, err
    }

    func main() {
    	var calls atomic.Int32
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		n := calls.Add(1)
    		if n < 3 {
    			fmt.Printf("server: attempt %d -> 503\n", n)
    			http.Error(w, "busy", http.StatusServiceUnavailable)
    			return
    		}
    		fmt.Printf("server: attempt %d -> 200\n", n)
    		fmt.Fprint(w, `{"status":"ok"}`)
    	}))
    	defer srv.Close()

    	body, err := getWithRetry(context.Background(), srv.Client(), srv.URL+"/health", 5, 10*time.Millisecond)
    	fmt.Printf("body=%s err=%v calls=%d\n", body, err, calls.Load())
    }
  tests: |
    package main

    import (
    	"context"
    	"errors"
    	"net/http"
    	"net/http/httptest"
    	"sync/atomic"
    	"testing"
    	"time"
    )

    // flaky answers with the given statuses in order, then 200 "ok" forever.
    func flaky(statuses ...int) (*httptest.Server, *atomic.Int32) {
    	var calls atomic.Int32
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		n := int(calls.Add(1))
    		if n <= len(statuses) {
    			http.Error(w, "nope", statuses[n-1])
    			return
    		}
    		w.Write([]byte("ok"))
    	}))
    	return srv, &calls
    }

    func TestSucceedsFirstTime(t *testing.T) {
    	srv, calls := flaky()
    	defer srv.Close()
    	body, err := getWithRetry(context.Background(), srv.Client(), srv.URL, 3, time.Millisecond)
    	if err != nil || string(body) != "ok" {
    		t.Fatalf("getWithRetry = (%q, %v), want (\"ok\", nil)", body, err)
    	}
    	if n := calls.Load(); n != 1 {
    		t.Errorf("server saw %d requests, want 1 (don't retry a success)", n)
    	}
    }

    func TestRetriesRetryableStatuses(t *testing.T) {
    	srv, calls := flaky(503, 429, 502, 500, 504)
    	defer srv.Close()
    	body, err := getWithRetry(context.Background(), srv.Client(), srv.URL, 6, time.Millisecond)
    	if err != nil || string(body) != "ok" {
    		t.Fatalf("after 503, 429, 502, 500, 504 then 200: getWithRetry = (%q, %v), want (\"ok\", nil)", body, err)
    	}
    	if n := calls.Load(); n != 6 {
    		t.Errorf("server saw %d requests, want 6", n)
    	}
    }

    func TestGivesUpAfterMaxAttempts(t *testing.T) {
    	srv, calls := flaky(503, 503, 503, 503, 503, 503, 503, 503, 503, 503)
    	defer srv.Close()
    	body, err := getWithRetry(context.Background(), srv.Client(), srv.URL, 4, time.Millisecond)
    	if err == nil {
    		t.Fatalf("server always says 503: getWithRetry = (%q, nil), want an error", body)
    	}
    	if n := calls.Load(); n != 4 {
    		t.Errorf("maxAttempts=4: server saw %d requests, want exactly 4", n)
    	}
    }

    func TestDoesNotRetryClientErrors(t *testing.T) {
    	for _, status := range []int{400, 401, 404, 422} {
    		srv, calls := flaky(status)
    		_, err := getWithRetry(context.Background(), srv.Client(), srv.URL, 5, time.Millisecond)
    		srv.Close()
    		if err == nil {
    			t.Errorf("status %d: getWithRetry returned nil error", status)
    		}
    		if n := calls.Load(); n != 1 {
    			t.Errorf("status %d: server saw %d requests, want 1 (a %d won't fix itself, so don't retry)", status, n, status)
    		}
    	}
    }

    func TestRetriesNetworkErrors(t *testing.T) {
    	srv := httptest.NewServer(http.NotFoundHandler())
    	url := srv.URL
    	srv.Close() // connection refused from now on

    	start := time.Now()
    	_, err := getWithRetry(context.Background(), http.DefaultClient, url, 3, time.Millisecond)
    	if err == nil {
    		t.Fatal("getWithRetry against a closed server returned nil error")
    	}
    	if time.Since(start) > 2*time.Second {
    		t.Errorf("took %v; with a 1ms base delay it should be quick", time.Since(start))
    	}
    }

    func TestStopsWhenContextDone(t *testing.T) {
    	srv, calls := flaky(503, 503, 503)
    	defer srv.Close()

    	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
    	defer cancel()
    	start := time.Now()
    	_, err := getWithRetry(ctx, srv.Client(), srv.URL, 5, time.Hour)
    	if el := time.Since(start); el > 2*time.Second {
    		t.Fatalf("getWithRetry kept waiting for %v after the context expired; select on ctx.Done() while waiting", el)
    	}
    	if !errors.Is(err, context.DeadlineExceeded) {
    		t.Errorf("getWithRetry returned %v, want an error matching context.DeadlineExceeded", err)
    	}
    	if n := calls.Load(); n != 1 {
    		t.Errorf("server saw %d requests, want 1", n)
    	}
    }
---

Networks hiccup. Servers restart, load balancers shuffle, and a request that failed a
moment ago often works if you just try again. A good client retries, **carefully**.

## What to retry

From the last two chapters:

- **Network errors** (connection refused or reset, timeouts): maybe. The server may
  have acted before the failure, so only retry idempotent requests.
- **429, 502, 503, 504**: yes, they're temporary by definition. 500 is often worth one
  more go.
- **Other 4xx**: never. The same request will fail the same way.
- **Context cancelled or expired**: never. The caller has given up.

## Don't hammer: back off

Retrying instantly in a tight loop turns one struggling server into a *very* struggling
server. **Exponential backoff** waits longer after each failure:

```
attempt 1 fails -> wait 100ms
attempt 2 fails -> wait 200ms
attempt 3 fails -> wait 400ms
attempt 4 fails -> wait 800ms
```

In Go, that doubling is a left shift: `base << (n-1)`. Real clients also cap it
(`min(base<<(n-1), 30*time.Second)`) so it never grows absurdly.

## Jitter: don't retry in lockstep

Picture 10,000 `trackr` users when the API blips for a second. With plain exponential
backoff, all of them retry at exactly +100ms, then +300ms, then +700ms: a synchronized
stampede. **Jitter** adds randomness. "Full jitter" picks a random delay between zero
and the backoff:

```go
func backoff(attempt int, base time.Duration) time.Duration {
	d := base << (attempt - 1)
	return rand.N(d) + 1 // random in [1ns, d]; math/rand/v2
}
```

`rand.N` from `math/rand/v2` is generic, so it works directly on `time.Duration`. It
panics on zero, so `base` must be positive.

## Waiting politely

`time.Sleep` can't be interrupted. Waiting in a `select` lets cancellation cut the wait
short:

```go
timer := time.NewTimer(backoff(attempt, base))
select {
case <-ctx.Done():
	timer.Stop()
	return nil, ctx.Err()
case <-timer.C:
}
```

## Every attempt is a fresh request

A `*http.Request` shouldn't be sent twice, and for a POST its body reader is used up
after the first attempt. Build a new request in each attempt, or use `req.GetBody` to
get a fresh body. And **close each failed response's body** before the next attempt,
or every retry leaks a connection.

## Putting a bound on it

Always limit retries, both by count (`maxAttempts`) and by time (the context
deadline). Without the second, 5 attempts with a 10s `Client.Timeout` and growing
backoff can take well over a minute.

## Your turn

Complete `getWithRetry(ctx, client, url, maxAttempts, baseDelay)`. The `retryable` and
`backoff` helpers are done. It should:

1. `GET` the URL with `client`, using a request built with `ctx`. Build a new request
   for each attempt.
2. On a 2xx, read the body and return it.
3. On a **non-retryable** status, return an error straight away, with no more attempts.
4. On a network error or a **retryable** status, close the body (if any) and, if
   attempts remain, wait `backoff(n, baseDelay)` before the next one, where `n` is the
   number of attempts that have failed so far. Stop waiting and return `ctx.Err()` if
   the context is done.
5. Make at most `maxAttempts` requests, then return an error.

A small helper that makes *one* attempt, say
`try(ctx, client, url) (body []byte, retry bool, err error)`, keeps the loop tidy and
makes `defer resp.Body.Close()` work per attempt.
