---
title: Issue States
difficulty: medium
after: http-clients-in-practice
hints:
  - 'Put one request in its own function, `fetchState(ctx, client, u, perRequest)`. Inside, `ctx, cancel := context.WithTimeout(ctx, perRequest)` then `defer cancel()`. The deferred cancel runs when that function returns, **after** you''ve read the body. A single timeout around the whole loop would be a budget for all the requests together.'
  - 'Every response you get back must be closed, including 404s you ignore. An unclosed body keeps its connection busy, so the next request has to open a new one, and the tests count connections.'
  - 'Don''t unwrap or replace the error from `client.Do`: it already wraps `context.DeadlineExceeded`, so `fmt.Errorf("issue %d: %w", id, err)` keeps `errors.Is` working. Skip ids you''ve already fetched with a `map[int]bool` (or check the result map).'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"time"
    )

    func issueStates(ctx context.Context, client *http.Client, baseURL string, ids []int, perRequest time.Duration) (map[int]string, error) {
    	return nil, nil
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		switch r.URL.Path {
    		case "/issues/1":
    			fmt.Fprint(w, `{"id": 1, "state": "open"}`)
    		case "/issues/2":
    			fmt.Fprint(w, `{"id": 2, "state": "closed"}`)
    		default:
    			http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
    		}
    	}))
    	defer srv.Close()

    	fmt.Println(issueStates(context.Background(), srv.Client(), srv.URL, []int{1, 2, 3}, time.Second))
    	// want: map[1:open 2:closed] <nil>
    }
  solution: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"net/url"
    	"strconv"
    	"time"
    )

    func issueStates(ctx context.Context, client *http.Client, baseURL string, ids []int, perRequest time.Duration) (map[int]string, error) {
    	states := map[int]string{}
    	seen := map[int]bool{}
    	for _, id := range ids {
    		if seen[id] {
    			continue
    		}
    		seen[id] = true
    		u, err := url.JoinPath(baseURL, "issues", strconv.Itoa(id))
    		if err != nil {
    			return nil, err
    		}
    		state, found, err := fetchState(ctx, client, u, perRequest)
    		if err != nil {
    			return nil, fmt.Errorf("issue %d: %w", id, err)
    		}
    		if found {
    			states[id] = state
    		}
    	}
    	return states, nil
    }

    // fetchState GETs one issue. found is false for a 404.
    func fetchState(ctx context.Context, client *http.Client, u string, perRequest time.Duration) (state string, found bool, err error) {
    	ctx, cancel := context.WithTimeout(ctx, perRequest)
    	defer cancel() // runs after the body has been read
    	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
    	if err != nil {
    		return "", false, err
    	}
    	resp, err := client.Do(req)
    	if err != nil {
    		return "", false, err
    	}
    	defer resp.Body.Close()
    	switch resp.StatusCode {
    	case http.StatusOK:
    		var iss struct {
    			State string `json:"state"`
    		}
    		if err := json.UnmarshalRead(resp.Body, &iss); err != nil {
    			return "", false, err
    		}
    		return iss.State, true, nil
    	case http.StatusNotFound:
    		return "", false, nil
    	default:
    		return "", false, fmt.Errorf("unexpected status %s", resp.Status)
    	}
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		switch r.URL.Path {
    		case "/issues/1":
    			fmt.Fprint(w, `{"id": 1, "state": "open"}`)
    		case "/issues/2":
    			fmt.Fprint(w, `{"id": 2, "state": "closed"}`)
    		default:
    			http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
    		}
    	}))
    	defer srv.Close()

    	fmt.Println(issueStates(context.Background(), srv.Client(), srv.URL, []int{1, 2, 3}, time.Second))
    }
  tests: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"io"
    	"maps"
    	"net"
    	"net/http"
    	"net/http/httptest"
    	"slices"
    	"strconv"
    	"strings"
    	"sync"
    	"sync/atomic"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    type testBodyTracker struct {
    	next           http.RoundTripper
    	opened, closed atomic.Int32
    }

    func (tb *testBodyTracker) RoundTrip(req *http.Request) (*http.Response, error) {
    	resp, err := tb.next.RoundTrip(req)
    	if err == nil {
    		tb.opened.Add(1)
    		resp.Body = &testCloseCounter{resp.Body, &tb.closed}
    	}
    	return resp, err
    }

    type testCloseCounter struct {
    	io.ReadCloser
    	n *atomic.Int32
    }

    func (c *testCloseCounter) Close() error {
    	c.n.Add(1)
    	return c.ReadCloser.Close()
    }

    // testAPI serves issues whose id is in states (404 otherwise), sleeps for
    // delay[id] first, and records every path it was asked for.
    type testAPI struct {
    	states map[int]string
    	delay  map[int]time.Duration
    	status map[int]int // non-zero: answer this status instead
    	body   map[int]string

    	mu    sync.Mutex
    	paths []string
    }

    func (a *testAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    	a.mu.Lock()
    	a.paths = append(a.paths, r.Method+" "+r.URL.Path)
    	a.mu.Unlock()
    	id, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/issues/"))
    	if d := a.delay[id]; d > 0 {
    		select {
    		case <-time.After(d):
    		case <-r.Context().Done():
    			return
    		}
    	}
    	w.Header().Set("Content-Type", "application/json")
    	if s := a.status[id]; s != 0 {
    		w.WriteHeader(s)
    		io.WriteString(w, `{"error": "something went wrong"}`)
    		return
    	}
    	if b, ok := a.body[id]; ok {
    		io.WriteString(w, b)
    		return
    	}
    	state, ok := a.states[id]
    	if !ok {
    		w.WriteHeader(http.StatusNotFound)
    		fmt.Fprintf(w, `{"error": "issue %d not found", "detail": "%s"}`, id, strings.Repeat("x", 2000))
    		return
    	}
    	fmt.Fprintf(w, `{"id": %d, "title": "Issue %d", "state": %q}`, id, id, state)
    }

    func (a *testAPI) requests() []string {
    	a.mu.Lock()
    	defer a.mu.Unlock()
    	return slices.Clone(a.paths)
    }

    func TestIssueStates(t *testing.T) {
    	api := &testAPI{states: map[int]string{1: "open", 2: "closed", 5: "open", 8: "in review ✓"}}
    	srv := httptest.NewServer(api)
    	defer srv.Close()
    	got, err := issueStates(t.Context(), srv.Client(), srv.URL+"/api/", []int{1, 2, 3, 5, 8, 2, 1}, time.Second)
    	want := map[int]string{1: "open", 2: "closed", 5: "open", 8: "in review ✓"}
    	if err != nil || !maps.Equal(got, want) {
    		t.Errorf("issueStates = %v, %v, want %v, nil (404s are left out)", got, err, want)
    	}
    	wantReqs := []string{"GET /api/issues/1", "GET /api/issues/2", "GET /api/issues/3", "GET /api/issues/5", "GET /api/issues/8"}
    	if reqs := api.requests(); !slices.Equal(reqs, wantReqs) {
    		t.Errorf("server received %v\n\twant %v (in order, each id once)", reqs, wantReqs)
    	}
    }

    func TestIssueStatesEmpty(t *testing.T) {
    	api := &testAPI{}
    	srv := httptest.NewServer(api)
    	defer srv.Close()
    	for _, ids := range [][]int{nil, {404, 405}} {
    		got, err := issueStates(t.Context(), srv.Client(), srv.URL+"/api", ids, time.Second)
    		if err != nil || got == nil || len(got) != 0 {
    			t.Errorf("issueStates(%v) where nothing exists = %#v, %v, want an empty, non-nil map", ids, got, err)
    		}
    	}
    }

    func TestIssueStatesErrors(t *testing.T) {
    	tests := []struct {
    		why string
    		api *testAPI
    	}{
    		{"500 on issue 2", &testAPI{states: map[int]string{1: "open", 2: "open", 3: "open"}, status: map[int]int{2: 500}}},
    		{"401 on issue 2", &testAPI{states: map[int]string{1: "open", 2: "open", 3: "open"}, status: map[int]int{2: 401}}},
    		{"malformed JSON for issue 2", &testAPI{states: map[int]string{1: "open", 3: "open"}, body: map[int]string{2: `{"id": 2, "state": `}}},
    	}
    	for _, tt := range tests {
    		srv := httptest.NewServer(tt.api)
    		got, err := issueStates(t.Context(), srv.Client(), srv.URL+"/api", []int{1, 2, 3}, time.Second)
    		srv.Close()
    		if err == nil || got != nil {
    			t.Errorf("%s: issueStates = %v, %v, want nil and an error", tt.why, got, err)
    			continue
    		}
    		if !strings.Contains(err.Error(), "2") {
    			t.Errorf("%s: error %q should say which issue failed (2)", tt.why, err)
    		}
    		if reqs := tt.api.requests(); len(reqs) != 2 {
    			t.Errorf("%s: server received %v, want to stop after issue 2", tt.why, reqs)
    		}
    	}
    }

    func TestIssueStatesReusesConnection(t *testing.T) {
    	api := &testAPI{states: map[int]string{}}
    	var ids []int
    	for id := 1; id <= 30; id++ {
    		ids = append(ids, id)
    		if id%3 == 0 {
    			api.states[id] = "open"
    		}
    	}
    	srv := httptest.NewUnstartedServer(api)
    	var conns atomic.Int32
    	srv.Config.ConnState = func(_ net.Conn, s http.ConnState) {
    		if s == http.StateNew {
    			conns.Add(1)
    		}
    	}
    	srv.Start()
    	defer srv.Close()
    	tb := &testBodyTracker{next: srv.Client().Transport}
    	got, err := issueStates(t.Context(), &http.Client{Transport: tb}, srv.URL+"/api", ids, time.Second)
    	if err != nil || len(got) != 10 {
    		t.Fatalf("issueStates(1..30) = %d states, %v, want 10 states, nil", len(got), err)
    	}
    	if tb.opened.Load() != 30 {
    		t.Errorf("%d responses went through the client you were given, want 30", tb.opened.Load())
    	}
    	if tb.closed.Load() != tb.opened.Load() {
    		t.Errorf("%d responses but %d body closes: close every body, including 404s", tb.opened.Load(), tb.closed.Load())
    	}
    	// Closing an unread body lets Go drain it in the background, so now and
    	// then a second connection is opened while the first is still draining.
    	if n := conns.Load(); n > 3 {
    		t.Errorf("30 sequential requests opened %d connections; closing every body needs only one or two (a body you don't close can't give its connection back)", n)
    	}
    }

    func TestIssueStatesPerRequestTimeout(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		api := &testAPI{
    			states: map[int]string{1: "open", 2: "open", 3: "open", 4: "open"},
    			delay:  map[int]time.Duration{1: 1500 * time.Millisecond, 2: 1500 * time.Millisecond, 3: time.Hour},
    		}
    		srv := httptest.NewTestServer(t, api)
    		start := time.Now()
    		got, err := issueStates(t.Context(), srv.Client(), "https://api.trackr.dev/api", []int{1, 2, 3, 4}, 2*time.Second)
    		elapsed := time.Since(start)
    		if !errors.Is(err, context.DeadlineExceeded) || got != nil {
    			t.Errorf("issue 3 never answers: issueStates = %v, %v, want nil and an error wrapping context.DeadlineExceeded", got, err)
    		}
    		if elapsed != 5*time.Second {
    			t.Errorf("issueStates gave up after %v, want 5s (1.5s + 1.5s + a 2s timeout for issue 3)", elapsed)
    		}
    		if reqs := api.requests(); len(reqs) != 3 {
    			t.Errorf("server received %v, want issues 1-3 only", reqs)
    		}
    	})
    }

    func TestIssueStatesTimeoutIsPerRequest(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		api := &testAPI{
    			states: map[int]string{1: "open", 2: "closed", 3: "open", 4: "open"},
    			delay:  map[int]time.Duration{1: 1500 * time.Millisecond, 2: 1500 * time.Millisecond, 3: 1500 * time.Millisecond, 4: 1500 * time.Millisecond},
    		}
    		srv := httptest.NewTestServer(t, api)
    		start := time.Now()
    		got, err := issueStates(t.Context(), srv.Client(), "https://api.trackr.dev/api", []int{1, 2, 3, 4}, 2*time.Second)
    		if err != nil || len(got) != 4 {
    			t.Errorf("four 1.5s responses with a 2s per-request timeout: issueStates = %v, %v, want 4 states, nil (the timeout is per request, not for the whole batch)", got, err)
    		}
    		if elapsed := time.Since(start); elapsed != 6*time.Second {
    			t.Errorf("issueStates took %v, want 6s", elapsed)
    		}
    	})
    }

    func TestIssueStatesParentContext(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		api := &testAPI{states: map[int]string{1: "open", 2: "open"}, delay: map[int]time.Duration{1: time.Hour, 2: time.Hour}}
    		srv := httptest.NewTestServer(t, api)
    		ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
    		defer cancel()
    		start := time.Now()
    		_, err := issueStates(ctx, srv.Client(), "https://api.trackr.dev/api", []int{1, 2}, time.Minute)
    		if !errors.Is(err, context.DeadlineExceeded) {
    			t.Errorf("caller's context expires after 3s: issueStates error = %v, want context.DeadlineExceeded", err)
    		}
    		if elapsed := time.Since(start); elapsed != 3*time.Second {
    			t.Errorf("caller's context expires after 3s, but issueStates returned after %v (derive each timeout from ctx)", elapsed)
    		}
    	})
    }
---

`trackr status 12 15 99 104` prints the state of several issues. The Trackr API
has no batch endpoint, so it's one `GET` per issue, sent one after another over
**one** reused connection, each with its **own** timeout so a single stuck
request can't hang the command.

Complete `issueStates(ctx, client, baseURL, ids, perRequest)`:

- For each id, in order, `GET <baseURL>/issues/<id>` with `client`. The
  response is JSON like `{"id": 12, "title": "...", "state": "open"}`.
- `200 OK`: record the issue's `state` in the result map under its id.
- `404 Not Found`: the issue doesn't exist. Leave it out of the map and carry on.
- Any other status, malformed JSON or a failed request: stop and return `nil`
  and an error that mentions the id. Don't send any more requests.
- Fetch each distinct id **once**, even if it appears twice in `ids`.
- Give each request at most `perRequest` to finish, **including** reading its
  body, derived from `ctx` so the caller can still cancel everything. A timeout
  must be detectable with `errors.Is(err, context.DeadlineExceeded)`.
- Return an empty, non-nil map when nothing was found.

## Example

```
issueStates(ctx, c, "https://api.trackr.dev/api", []int{1, 2, 3, 2}, time.Second)
// GET /api/issues/1 -> 200 {"state": "open"}
// GET /api/issues/2 -> 200 {"state": "closed"}
// GET /api/issues/3 -> 404
// => map[1:open 2:closed], nil
```

## Constraints

- One test sends 30 requests, 20 of them 404s, and counts the TCP connections
  the server saw. Close every body, including the 404s you ignore.
- The timing tests run on synctest's fake clock and check exact durations:
  four 1.5-second responses with `perRequest` = 2s must all succeed (6s total),
  while a server that never answers must fail after exactly 2s.
