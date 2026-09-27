---
title: Retry a POST Safely
difficulty: hard
after: http-clients-in-practice
hints:
  - 'Encode the JSON **once** into a `[]byte`, then build a brand-new request for every attempt with `bytes.NewReader(body)`. A request''s body is a reader: after the first attempt it''s used up, so re-sending the same `*http.Request` sends nothing.'
  - 'Write `attempt(ctx) (Issue, wait time.Duration, retry bool, err error)` for one try. Retryable means a network error, 429 or 5xx. For those, `wait` is `Retry-After` if it parses as whole seconds, else `BaseDelay << (n-1)` for retry n. Then the loop only decides whether to stop, wait or return.'
  - 'Before waiting, check `ctx.Deadline()`: if `time.Until(deadline) < wait`, return an error wrapping `context.DeadlineExceeded` straight away. Then wait with `select { case <-time.After(wait): case <-ctx.Done(): return Issue{}, ctx.Err() }`, never a bare `time.Sleep`.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"sync/atomic"
    	"time"
    )

    // NewIssue is the body of POST /issues.
    type NewIssue struct {
    	Title  string   `json:"title"`
    	Labels []string `json:"labels,omitempty"`
    }

    // Issue is what the API sends back.
    type Issue struct {
    	ID    int    `json:"id"`
    	Title string `json:"title"`
    }

    // Creator creates issues, retrying safely.
    type Creator struct {
    	Client      *http.Client
    	BaseURL     string
    	MaxAttempts int           // attempts in total, at least 1
    	BaseDelay   time.Duration // wait before retry n (without Retry-After) is BaseDelay << (n-1)
    }

    func (c *Creator) Create(ctx context.Context, key string, in NewIssue) (Issue, error) {
    	return Issue{}, nil
    }

    func main() {
    	var calls atomic.Int32
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Println("server got:", r.Method, r.URL.Path, "key", r.Header.Get("Idempotency-Key"))
    		if calls.Add(1) == 1 {
    			w.WriteHeader(http.StatusServiceUnavailable)
    			return
    		}
    		w.WriteHeader(http.StatusCreated)
    		fmt.Fprint(w, `{"id": 101, "title": "Dark mode"}`)
    	}))
    	defer srv.Close()

    	c := &Creator{Client: srv.Client(), BaseURL: srv.URL, MaxAttempts: 3, BaseDelay: 100 * time.Millisecond}
    	fmt.Println(c.Create(context.Background(), "k-7f3a", NewIssue{Title: "Dark mode"}))
    	// want:
    	// server got: POST /issues key k-7f3a
    	// server got: POST /issues key k-7f3a
    	// {101 Dark mode} <nil>
    }
  solution: |
    package main

    import (
    	"bytes"
    	"context"
    	"encoding/json/v2"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"net/url"
    	"strconv"
    	"sync/atomic"
    	"time"
    )

    // NewIssue is the body of POST /issues.
    type NewIssue struct {
    	Title  string   `json:"title"`
    	Labels []string `json:"labels,omitempty"`
    }

    // Issue is what the API sends back.
    type Issue struct {
    	ID    int    `json:"id"`
    	Title string `json:"title"`
    }

    // Creator creates issues, retrying safely.
    type Creator struct {
    	Client      *http.Client
    	BaseURL     string
    	MaxAttempts int           // attempts in total, at least 1
    	BaseDelay   time.Duration // wait before retry n (without Retry-After) is BaseDelay << (n-1)
    }

    func (c *Creator) Create(ctx context.Context, key string, in NewIssue) (Issue, error) {
    	u, err := url.JoinPath(c.BaseURL, "issues")
    	if err != nil {
    		return Issue{}, err
    	}
    	body, err := json.Marshal(in)
    	if err != nil {
    		return Issue{}, err
    	}
    	for n := 1; ; n++ {
    		iss, wait, retry, err := c.attempt(ctx, u, key, body)
    		if err == nil {
    			return iss, nil
    		}
    		if !retry || n >= c.MaxAttempts {
    			return Issue{}, err
    		}
    		if wait < 0 {
    			wait = c.BaseDelay << (n - 1)
    		}
    		if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) < wait {
    			return Issue{}, fmt.Errorf("creating issue: next retry in %v is past the deadline: %w (last error: %v)", wait, context.DeadlineExceeded, err)
    		}
    		select {
    		case <-time.After(wait):
    		case <-ctx.Done():
    			return Issue{}, ctx.Err()
    		}
    	}
    }

    // attempt sends one POST. wait is the server's Retry-After, or -1 if it sent none.
    func (c *Creator) attempt(ctx context.Context, u, key string, body []byte) (iss Issue, wait time.Duration, retry bool, err error) {
    	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
    	if err != nil {
    		return Issue{}, 0, false, err
    	}
    	req.Header.Set("Content-Type", "application/json")
    	req.Header.Set("Idempotency-Key", key)
    	resp, err := c.Client.Do(req)
    	if err != nil {
    		// The idempotency key makes a lost response safe to retry, unless we gave up.
    		return Issue{}, -1, ctx.Err() == nil, err
    	}
    	defer resp.Body.Close()

    	switch {
    	case resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated:
    		if err := json.UnmarshalRead(resp.Body, &iss); err != nil {
    			return Issue{}, 0, false, fmt.Errorf("creating issue: %w", err)
    		}
    		return iss, 0, false, nil
    	case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
    		wait = -1
    		if s, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && s >= 0 {
    			wait = time.Duration(s) * time.Second
    		}
    		return Issue{}, wait, true, fmt.Errorf("creating issue: %s", resp.Status)
    	default:
    		return Issue{}, 0, false, fmt.Errorf("creating issue: %s", resp.Status)
    	}
    }

    func main() {
    	var calls atomic.Int32
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Println("server got:", r.Method, r.URL.Path, "key", r.Header.Get("Idempotency-Key"))
    		if calls.Add(1) == 1 {
    			w.WriteHeader(http.StatusServiceUnavailable)
    			return
    		}
    		w.WriteHeader(http.StatusCreated)
    		fmt.Fprint(w, `{"id": 101, "title": "Dark mode"}`)
    	}))
    	defer srv.Close()

    	c := &Creator{Client: srv.Client(), BaseURL: srv.URL, MaxAttempts: 3, BaseDelay: 100 * time.Millisecond}
    	fmt.Println(c.Create(context.Background(), "k-7f3a", NewIssue{Title: "Dark mode"}))
    }
  tests: |
    package main

    import (
    	"context"
    	"errors"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"sync"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // testStep scripts one answer of the fake API.
    type testStep struct {
    	status     int    // 0 means: drop the connection without answering
    	retryAfter string // Retry-After header, if not empty
    }

    type testSeen struct {
    	at            time.Duration // since the test started (fake time)
    	method, path  string
    	key, ctype    string
    	body          string
    	contentLength int64
    }

    // testScript is a fake API that answers with steps in order, then 201s.
    type testScript struct {
    	start time.Time
    	steps []testStep

    	mu   sync.Mutex
    	seen []testSeen
    }

    func (s *testScript) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    	b, _ := io.ReadAll(r.Body)
    	s.mu.Lock()
    	n := len(s.seen)
    	s.seen = append(s.seen, testSeen{time.Since(s.start), r.Method, r.URL.Path, r.Header.Get("Idempotency-Key"),
    		r.Header.Get("Content-Type"), string(b), r.ContentLength})
    	s.mu.Unlock()
    	if n < len(s.steps) {
    		st := s.steps[n]
    		if st.status == 0 {
    			panic(http.ErrAbortHandler) // the connection drops mid-request
    		}
    		if st.retryAfter != "" {
    			w.Header().Set("Retry-After", st.retryAfter)
    		}
    		w.WriteHeader(st.status)
    		io.WriteString(w, `{"error": "not now"}`)
    		return
    	}
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(http.StatusCreated)
    	io.WriteString(w, `{"id": 101, "title": "Dark mode"}`)
    }

    func (s *testScript) requests() []testSeen {
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	return append([]testSeen(nil), s.seen...)
    }

    func (s *testScript) times() []time.Duration {
    	var ts []time.Duration
    	for _, r := range s.requests() {
    		ts = append(ts, r.at)
    	}
    	return ts
    }

    // testCreate runs Create against steps inside a synctest bubble and calls check.
    func testCreate(t *testing.T, steps []testStep, attempts int, ctxFn func(context.Context) (context.Context, context.CancelFunc),
    	check func(t *testing.T, s *testScript, iss Issue, err error, elapsed time.Duration)) {
    	synctest.Test(t, func(t *testing.T) {
    		s := &testScript{start: time.Now(), steps: steps}
    		srv := httptest.NewTestServer(t, s)
    		c := &Creator{Client: srv.Client(), BaseURL: "https://api.trackr.dev/v1/", MaxAttempts: attempts, BaseDelay: time.Second}
    		ctx, cancel := context.WithCancel(t.Context())
    		if ctxFn != nil {
    			ctx, cancel = ctxFn(t.Context())
    		}
    		defer cancel()
    		iss, err := c.Create(ctx, "key-8d1f", NewIssue{Title: "Dark mode", Labels: []string{"ui"}})
    		check(t, s, iss, err, time.Since(s.start))
    	})
    }

    func testWantTimes(t *testing.T, s *testScript, want ...time.Duration) {
    	t.Helper()
    	got := s.times()
    	if len(got) != len(want) {
    		t.Errorf("server received %d requests at %v, want %d at %v", len(got), got, len(want), want)
    		return
    	}
    	for i := range got {
    		if got[i] != want[i] {
    			t.Errorf("server received requests at %v, want %v", got, want)
    			return
    		}
    	}
    }

    func TestCreateFirstTry(t *testing.T) {
    	testCreate(t, nil, 3, nil, func(t *testing.T, s *testScript, iss Issue, err error, _ time.Duration) {
    		if err != nil || iss != (Issue{101, "Dark mode"}) {
    			t.Errorf("Create = %+v, %v, want {101 Dark mode}, nil", iss, err)
    		}
    		reqs := s.requests()
    		if len(reqs) != 1 {
    			t.Fatalf("server received %d requests, want 1", len(reqs))
    		}
    		r := reqs[0]
    		if r.method != "POST" || r.path != "/v1/issues" {
    			t.Errorf("request was %s %s, want POST /v1/issues", r.method, r.path)
    		}
    		if r.key != "key-8d1f" || r.ctype != "application/json" {
    			t.Errorf("Idempotency-Key = %q, Content-Type = %q, want \"key-8d1f\" and \"application/json\"", r.key, r.ctype)
    		}
    		if want := `{"title":"Dark mode","labels":["ui"]}`; r.body != want {
    			t.Errorf("request body = %s, want %s", r.body, want)
    		}
    	})
    }

    func TestCreateRetriesWithSameBodyAndKey(t *testing.T) {
    	steps := []testStep{{status: 503}, {status: 502}, {status: 500}}
    	testCreate(t, steps, 4, nil, func(t *testing.T, s *testScript, iss Issue, err error, elapsed time.Duration) {
    		if err != nil || iss.ID != 101 {
    			t.Errorf("503, 502, 500, then 201: Create = %+v, %v, want issue 101, nil", iss, err)
    		}
    		testWantTimes(t, s, 0, time.Second, 3*time.Second, 7*time.Second)
    		for i, r := range s.requests() {
    			if r.key != "key-8d1f" || r.body != `{"title":"Dark mode","labels":["ui"]}` || r.contentLength != int64(len(r.body)) {
    				t.Errorf("attempt %d sent key %q, Content-Length %d, body %q; every attempt must send the same key and the full body (build a new request each time)", i+1, r.key, r.contentLength, r.body)
    			}
    		}
    	})
    }

    func TestCreateHonoursRetryAfter(t *testing.T) {
    	steps := []testStep{{status: 429, retryAfter: "7"}, {status: 503, retryAfter: "0"}, {status: 503, retryAfter: "soon"}}
    	testCreate(t, steps, 5, nil, func(t *testing.T, s *testScript, iss Issue, err error, elapsed time.Duration) {
    		if err != nil || iss.ID != 101 {
    			t.Errorf("Create = %+v, %v, want issue 101, nil", iss, err)
    		}
    		// Retry-After 7 -> 7s; Retry-After 0 -> right away; "soon" is invalid -> backoff for retry 3 (4s).
    		testWantTimes(t, s, 0, 7*time.Second, 7*time.Second, 11*time.Second)
    	})
    }

    func TestCreateRetriesDroppedConnection(t *testing.T) {
    	testCreate(t, []testStep{{status: 0}}, 3, nil, func(t *testing.T, s *testScript, iss Issue, err error, _ time.Duration) {
    		if err != nil || iss.ID != 101 {
    			t.Errorf("connection dropped, then 201: Create = %+v, %v, want issue 101, nil (the Idempotency-Key makes it safe to retry)", iss, err)
    		}
    		testWantTimes(t, s, 0, time.Second)
    	})
    }

    func TestCreateDoesNotRetryClientErrors(t *testing.T) {
    	for _, status := range []int{400, 401, 409, 422} {
    		testCreate(t, []testStep{{status: status, retryAfter: "1"}}, 3, nil, func(t *testing.T, s *testScript, iss Issue, err error, elapsed time.Duration) {
    			if err == nil || !strings.Contains(err.Error(), http.StatusText(status)) {
    				t.Errorf("status %d: Create error = %v, want an error mentioning %q", status, err, http.StatusText(status))
    			}
    			if iss != (Issue{}) {
    				t.Errorf("status %d: Create returned %+v with its error, want Issue{}", status, iss)
    			}
    			testWantTimes(t, s, 0)
    		})
    	}
    }

    func TestCreateGivesUp(t *testing.T) {
    	steps := []testStep{{status: 503}, {status: 503}, {status: 503}, {status: 503}, {status: 503}}
    	testCreate(t, steps, 4, nil, func(t *testing.T, s *testScript, iss Issue, err error, elapsed time.Duration) {
    		if err == nil || !strings.Contains(err.Error(), "503") {
    			t.Errorf("always 503: Create error = %v, want an error mentioning 503", err)
    		}
    		testWantTimes(t, s, 0, time.Second, 3*time.Second, 7*time.Second)
    		if elapsed != 7*time.Second {
    			t.Errorf("Create returned after %v, want 7s (no wait after the last attempt)", elapsed)
    		}
    	})
    }

    func TestCreateCancelledWhileWaiting(t *testing.T) {
    	steps := []testStep{{status: 503}, {status: 503}, {status: 503}}
    	cancelAt := func(parent context.Context) (context.Context, context.CancelFunc) {
    		ctx, cancel := context.WithCancel(parent)
    		time.AfterFunc(2500*time.Millisecond, cancel)
    		return ctx, cancel
    	}
    	testCreate(t, steps, 5, cancelAt, func(t *testing.T, s *testScript, iss Issue, err error, elapsed time.Duration) {
    		if !errors.Is(err, context.Canceled) {
    			t.Errorf("cancelled during the second wait: Create error = %v, want context.Canceled", err)
    		}
    		if elapsed != 2500*time.Millisecond {
    			t.Errorf("cancelled at 2.5s, but Create returned at %v (wait in a select on ctx.Done())", elapsed)
    		}
    		testWantTimes(t, s, 0, time.Second)
    	})
    }

    func TestCreateRetryAfterPastDeadline(t *testing.T) {
    	deadline := func(parent context.Context) (context.Context, context.CancelFunc) {
    		return context.WithTimeout(parent, 10*time.Second)
    	}
    	testCreate(t, []testStep{{status: 429, retryAfter: "30"}}, 3, deadline, func(t *testing.T, s *testScript, iss Issue, err error, elapsed time.Duration) {
    		if !errors.Is(err, context.DeadlineExceeded) {
    			t.Errorf("Retry-After 30 with 10s left: Create error = %v, want one wrapping context.DeadlineExceeded", err)
    		}
    		if elapsed != 0 {
    			t.Errorf("Retry-After 30 with 10s left: Create returned after %v, want 0s (don't wait for a retry you can't make)", elapsed)
    		}
    		testWantTimes(t, s, 0)
    	})
    }
---

Retrying a `GET` is harmless. Retrying a `POST /issues` is how you end up with
three copies of "Dark mode". Unless the API supports an **idempotency key**:
the client sends a unique `Idempotency-Key` header, and if the server sees the
same key twice it returns the issue it already created instead of making another.
With that in place, `trackr new` can retry creates as boldly as reads.

Complete `(*Creator).Create(ctx, key, in)`.

**Each attempt**

- `POST <BaseURL>/issues` (`BaseURL` may end in `/`) with `c.Client`, the JSON
  encoding of `in` as the body, `Content-Type: application/json` and
  `Idempotency-Key: <key>`. Every attempt sends the same key and the same,
  **complete** body.
- `200` or `201`: decode and return the `Issue`.

**Retrying**

- Retry after a failed request (dropped connection and so on), a `429`, or any
  `5xx`. Anything else, such as `400`, `409` or `422`, fails at once with an
  error that includes the status text.
- Make at most `MaxAttempts` attempts in total. When they're all used up,
  return an error that describes the last failure (e.g. includes `503`). Don't
  wait after the last attempt.
- Before retry *n* (n = 1, 2, 3, ...), wait for the response's `Retry-After`
  if it's a whole number of seconds (`0` means "right away"). Otherwise wait
  `BaseDelay << (n-1)`: 1s, 2s, 4s, ... for a `BaseDelay` of 1s.
- If `ctx` is cancelled during a wait, return straight away with an error
  that `errors.Is` matches against `ctx.Err()`.
- If `ctx` has a deadline that comes **before** the wait would end, don't wait
  at all: return immediately with an error wrapping `context.DeadlineExceeded`.

Return `Issue{}` with every error.

## Example

`BaseDelay` = 1s, `MaxAttempts` = 5:

```
t=0s   POST /issues -> 429, Retry-After: 7
t=7s   POST /issues -> 503, Retry-After: 0
t=7s   POST /issues -> 503, Retry-After: soon   (invalid, so back off: retry 3 -> 4s)
t=11s  POST /issues -> 201 {"id": 101, ...}      => Issue{101, "Dark mode"}, nil
```

## Constraints

- The tests run in a `testing/synctest` bubble with `httptest.NewTestServer`,
  so they check the time of every request to the exact second without
  actually waiting. Use `time.After` or timers (not a busy loop), and wait in a
  `select` that also watches `ctx.Done()`.
- The fake API records the method, path, headers and body of every attempt.
- No jitter here: the tests need exact times.
