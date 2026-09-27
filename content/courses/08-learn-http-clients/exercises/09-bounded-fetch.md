---
title: Bounded Fetch
difficulty: hard
after: http-clients-in-practice
hints:
  - 'Order is easy if every goroutine writes to **its own index**: `issues := make([]Issue, len(ids))`, and the goroutine for `ids[i]` sets `issues[i]`. No sorting, no mutex for the slice.'
  - 'A buffered channel `sem := make(chan struct{}, maxInFlight)` is a semaphore: send before starting a request (it blocks when `maxInFlight` are running), receive when it''s done. Acquire it in the loop **before** `go`, and stop the loop when `ctx` is done.'
  - 'Derive `ctx, cancel := context.WithCancel(ctx)`. The first goroutine to fail stores its error (a `sync.Once` or a mutex) and calls `cancel()`, which aborts the requests still running. After `wg.Wait()`, return that first error: the others are just `context canceled` echoes of it.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"time"
    )

    // Issue is one Trackr issue.
    type Issue struct {
    	ID    int    `json:"id"`
    	Title string `json:"title"`
    }

    func fetchIssues(ctx context.Context, client *http.Client, baseURL string, ids []int, maxInFlight int) ([]Issue, error) {
    	return nil, nil
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		id := strings.TrimPrefix(r.URL.Path, "/issues/")
    		time.Sleep(50 * time.Millisecond)
    		fmt.Fprintf(w, `{"id": %s, "title": "Issue %s"}`, id, id)
    	}))
    	defer srv.Close()

    	start := time.Now()
    	issues, err := fetchIssues(context.Background(), srv.Client(), srv.URL, []int{5, 3, 8, 1, 9, 2}, 3)
    	fmt.Println(issues, err)
    	fmt.Println("about 100ms?", time.Since(start).Round(50*time.Millisecond))
    	// want:
    	// [{5 Issue 5} {3 Issue 3} {8 Issue 8} {1 Issue 1} {9 Issue 9} {2 Issue 2}] <nil>
    	// about 100ms? 100ms
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
    	"strings"
    	"sync"
    	"time"
    )

    // Issue is one Trackr issue.
    type Issue struct {
    	ID    int    `json:"id"`
    	Title string `json:"title"`
    }

    func fetchIssues(ctx context.Context, client *http.Client, baseURL string, ids []int, maxInFlight int) ([]Issue, error) {
    	ctx, cancel := context.WithCancel(ctx)
    	defer cancel()

    	issues := make([]Issue, len(ids))
    	sem := make(chan struct{}, max(1, maxInFlight))
    	var (
    		wg       sync.WaitGroup
    		once     sync.Once
    		firstErr error
    	)
    	fail := func(err error) {
    		once.Do(func() {
    			firstErr = err
    			cancel()
    		})
    	}

    loop:
    	for i, id := range ids {
    		select {
    		case sem <- struct{}{}:
    		case <-ctx.Done():
    			break loop
    		}
    		wg.Go(func() {
    			defer func() { <-sem }()
    			iss, err := fetchOne(ctx, client, baseURL, id)
    			if err != nil {
    				fail(fmt.Errorf("issue %d: %w", id, err))
    				return
    			}
    			issues[i] = iss
    		})
    	}
    	wg.Wait()
    	if firstErr == nil && ctx.Err() != nil {
    		firstErr = ctx.Err() // the caller cancelled
    	}
    	if firstErr != nil {
    		return nil, firstErr
    	}
    	return issues, nil
    }

    func fetchOne(ctx context.Context, client *http.Client, baseURL string, id int) (Issue, error) {
    	u, err := url.JoinPath(baseURL, "issues", strconv.Itoa(id))
    	if err != nil {
    		return Issue{}, err
    	}
    	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
    	if err != nil {
    		return Issue{}, err
    	}
    	resp, err := client.Do(req)
    	if err != nil {
    		return Issue{}, err
    	}
    	defer resp.Body.Close()
    	if resp.StatusCode != http.StatusOK {
    		return Issue{}, fmt.Errorf("unexpected status %s", resp.Status)
    	}
    	var iss Issue
    	if err := json.UnmarshalRead(resp.Body, &iss); err != nil {
    		return Issue{}, err
    	}
    	return iss, nil
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		id := strings.TrimPrefix(r.URL.Path, "/issues/")
    		time.Sleep(50 * time.Millisecond)
    		fmt.Fprintf(w, `{"id": %s, "title": "Issue %s"}`, id, id)
    	}))
    	defer srv.Close()

    	start := time.Now()
    	issues, err := fetchIssues(context.Background(), srv.Client(), srv.URL, []int{5, 3, 8, 1, 9, 2}, 3)
    	fmt.Println(issues, err)
    	fmt.Println("about 100ms?", time.Since(start).Round(50*time.Millisecond))
    }
  tests: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"strconv"
    	"strings"
    	"sync"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // testAPI serves /v1/issues/{id} after delay(id), tracking how many requests
    // are in flight at once.
    type testAPI struct {
    	delay  func(id int) time.Duration
    	status map[int]int
    	body   map[int]string

    	mu                      sync.Mutex
    	inFlight, peak, started int
    	cancelled               int
    }

    func (a *testAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/v1/issues/"))
    	if err != nil || r.Method != http.MethodGet {
    		http.Error(w, "bad request "+r.Method+" "+r.URL.Path, http.StatusBadRequest)
    		return
    	}
    	a.mu.Lock()
    	a.started++
    	a.inFlight++
    	a.peak = max(a.peak, a.inFlight)
    	a.mu.Unlock()
    	defer func() {
    		a.mu.Lock()
    		a.inFlight--
    		a.mu.Unlock()
    	}()
    	if a.delay != nil {
    		select {
    		case <-time.After(a.delay(id)):
    		case <-r.Context().Done():
    			a.mu.Lock()
    			a.cancelled++
    			a.mu.Unlock()
    			return
    		}
    	}
    	if s := a.status[id]; s != 0 {
    		w.WriteHeader(s)
    		io.WriteString(w, `{"error": "boom"}`)
    		return
    	}
    	if b, ok := a.body[id]; ok {
    		io.WriteString(w, b)
    		return
    	}
    	fmt.Fprintf(w, `{"id": %d, "title": "Issue %d"}`, id, id)
    }

    func (a *testAPI) stats() (peak, started, cancelled int) {
    	a.mu.Lock()
    	defer a.mu.Unlock()
    	return a.peak, a.started, a.cancelled
    }

    func testIDs(n int) []int {
    	ids := make([]int, n)
    	for i := range ids {
    		ids[i] = 100 + (i*37)%n // shuffled, distinct
    	}
    	return ids
    }

    func testCheckOrder(t *testing.T, ids []int, got []Issue) {
    	t.Helper()
    	if len(got) != len(ids) {
    		t.Fatalf("got %d issues for %d ids", len(got), len(ids))
    	}
    	for i, id := range ids {
    		if want := (Issue{id, fmt.Sprintf("Issue %d", id)}); got[i] != want {
    			t.Fatalf("issues[%d] = %+v, want %+v (results must be in the same order as ids)", i, got[i], want)
    		}
    	}
    }

    func TestFetchLimitAndOrder(t *testing.T) {
    	for _, tt := range []struct{ n, k int }{{10, 3}, {12, 4}, {5, 1}, {4, 10}, {9, 9}} {
    		t.Run(fmt.Sprintf("%d ids, max %d", tt.n, tt.k), func(t *testing.T) {
    			synctest.Test(t, func(t *testing.T) {
    				api := &testAPI{delay: func(int) time.Duration { return time.Second }}
    				srv := httptest.NewTestServer(t, api)
    				ids := testIDs(tt.n)
    				start := time.Now()
    				got, err := fetchIssues(t.Context(), srv.Client(), "https://api.trackr.dev/v1/", ids, tt.k)
    				elapsed := time.Since(start)
    				if err != nil {
    					t.Fatalf("fetchIssues = %v", err)
    				}
    				testCheckOrder(t, ids, got)
    				peak, started, _ := api.stats()
    				if want := min(tt.n, tt.k); peak != want {
    					t.Errorf("at most %d requests were in flight at once, want exactly %d (maxInFlight %d)", peak, want, tt.k)
    				}
    				if started != tt.n {
    					t.Errorf("server saw %d requests, want %d", started, tt.n)
    				}
    				if want := time.Duration((tt.n+tt.k-1)/tt.k) * time.Second; elapsed != want {
    					t.Errorf("fetching %d one-second issues with maxInFlight %d took %v, want %v", tt.n, tt.k, elapsed, want)
    				}
    			})
    		})
    	}
    }

    func TestFetchStartsNextAsSoonAsOneFinishes(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		// id 100 is slow; the rest are quick. A pool keeps the other slot busy
    		// while 100 is still running, rather than waiting for a whole batch.
    		api := &testAPI{delay: func(id int) time.Duration {
    			if id == 100 {
    				return 10 * time.Second
    			}
    			return time.Second
    		}}
    		srv := httptest.NewTestServer(t, api)
    		ids := []int{100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110}
    		start := time.Now()
    		got, err := fetchIssues(t.Context(), srv.Client(), "https://api.trackr.dev/v1", ids, 2)
    		if err != nil {
    			t.Fatalf("fetchIssues = %v", err)
    		}
    		testCheckOrder(t, ids, got)
    		if elapsed := time.Since(start); elapsed != 10*time.Second {
    			t.Errorf("one 10s issue plus ten 1s issues with maxInFlight 2 took %v, want 10s (start the next request as soon as a slot frees up)", elapsed)
    		}
    	})
    }

    func TestFetchReverseCompletion(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		api := &testAPI{delay: func(id int) time.Duration { return time.Duration(200-id) * time.Millisecond }}
    		srv := httptest.NewTestServer(t, api)
    		ids := testIDs(15)
    		got, err := fetchIssues(t.Context(), srv.Client(), "https://api.trackr.dev/v1", ids, 15)
    		if err != nil {
    			t.Fatalf("fetchIssues = %v", err)
    		}
    		testCheckOrder(t, ids, got)
    	})
    }

    func TestFetchEmpty(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		api := &testAPI{}
    		srv := httptest.NewTestServer(t, api)
    		got, err := fetchIssues(t.Context(), srv.Client(), "https://api.trackr.dev/v1", nil, 3)
    		if err != nil || len(got) != 0 {
    			t.Errorf("fetchIssues(no ids) = %v, %v, want no issues and no error", got, err)
    		}
    		if _, started, _ := api.stats(); started != 0 {
    			t.Errorf("fetchIssues(no ids) sent %d requests", started)
    		}
    	})
    }

    func TestFetchFirstErrorCancelsTheRest(t *testing.T) {
    	tests := []struct {
    		why    string
    		api    *testAPI
    		inErr  string
    		failAt time.Duration
    	}{
    		{"issue 105 answers 500 after 2s", &testAPI{status: map[int]int{105: 500}}, "105", 2 * time.Second},
    		{"issue 102 sends malformed JSON after 2s", &testAPI{body: map[int]string{102: `{"id": 102, "title": `}}, "102", 2 * time.Second},
    	}
    	for _, tt := range tests {
    		t.Run(tt.why, func(t *testing.T) {
    			synctest.Test(t, func(t *testing.T) {
    				tt.api.delay = func(id int) time.Duration {
    					if id == 105 || id == 102 {
    						return 2 * time.Second
    					}
    					return time.Hour
    				}
    				srv := httptest.NewTestServer(t, tt.api)
    				ids := []int{100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115}
    				start := time.Now()
    				got, err := fetchIssues(t.Context(), srv.Client(), "https://api.trackr.dev/v1", ids, 8)
    				elapsed := time.Since(start)
    				if err == nil || got != nil {
    					t.Fatalf("fetchIssues = %v, %v, want nil and an error", got, err)
    				}
    				if errors.Is(err, context.Canceled) || !strings.Contains(err.Error(), tt.inErr) {
    					t.Errorf("fetchIssues error = %q, want the first real failure (mentioning issue %s), not a cancellation", err, tt.inErr)
    				}
    				if elapsed != tt.failAt {
    					t.Errorf("fetchIssues returned after %v, want %v: cancel the requests still in flight when one fails", elapsed, tt.failAt)
    				}
    				if _, started, _ := tt.api.stats(); started > 9 {
    					t.Errorf("server saw %d requests, want no new ones after the failure (at most 8 were running)", started)
    				}
    			})
    		})
    	}
    }

    func TestFetchCallerCancels(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		api := &testAPI{delay: func(int) time.Duration { return time.Second }}
    		srv := httptest.NewTestServer(t, api)
    		ctx, cancel := context.WithTimeout(t.Context(), 2500*time.Millisecond)
    		defer cancel()
    		start := time.Now()
    		got, err := fetchIssues(ctx, srv.Client(), "https://api.trackr.dev/v1", testIDs(20), 2)
    		if !errors.Is(err, context.DeadlineExceeded) || got != nil {
    			t.Errorf("caller's deadline passes: fetchIssues = %v, %v, want nil and context.DeadlineExceeded", got, err)
    		}
    		if elapsed := time.Since(start); elapsed != 2500*time.Millisecond {
    			t.Errorf("fetchIssues returned after %v, want 2.5s", elapsed)
    		}
    		if _, started, _ := api.stats(); started > 6 {
    			t.Errorf("server saw %d requests, want no new ones after the deadline", started)
    		}
    	})
    }
---

`trackr export --ids 5,3,8,...` fetches hundreds of issues. One at a time is
slow, and all at once gets `trackr` rate-limited (or banned). The sweet spot is
a fixed number of requests **in flight**, with results that still come out in
the order the user asked for.

Complete `fetchIssues(ctx, client, baseURL, ids, maxInFlight)`:

- For each id, `GET <baseURL>/issues/<id>` with `client` (`baseURL` may end in
  `/`). A `200 OK` carries JSON like `{"id": 5, "title": "..."}`.
- Never have more than `maxInFlight` requests running at once, and start the
  next one **as soon as** any running request finishes (a pool, not batches).
  Treat `maxInFlight < 1` as 1.
- Return the issues in the **same order as `ids`**, whatever order the responses
  arrive in.
- If any request fails (network error, a status other than 200, malformed
  JSON), stop: cancel the requests still in flight, start no new ones, and
  return `nil` and **that** error. The error should mention the failing id, and
  must not be a `context.Canceled` from one of the requests you cancelled.
- If `ctx` is cancelled or its deadline passes, return `nil` and an error
  matching `ctx.Err()` promptly, without starting new requests.
- No ids: no requests, no error, no issues.

## Example

With `maxInFlight` = 2 and each response taking 1 second:

```
t=0s  start 5, 3
t=1s  5 and 3 done -> start 8, 1
t=2s  8 and 1 done -> start 9
t=3s  9 done       => [{5 ...} {3 ...} {8 ...} {1 ...} {9 ...}], nil
```

## Constraints

- The tests run in a `testing/synctest` bubble with `httptest.NewTestServer`.
  The fake API records the peak number of concurrent requests and the tests
  check exact durations, e.g. 10 one-second requests with `maxInFlight` 3 must
  take exactly 4s.
- Every goroutine you start must have finished by the time `fetchIssues`
  returns. The bubble reports leaked goroutines as a failure.
- Avoid data races: if each goroutine writes only its own slot of the result
  slice, the slice needs no mutex.
