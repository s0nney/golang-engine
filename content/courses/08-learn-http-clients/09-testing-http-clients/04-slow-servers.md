---
title: Slow Servers and Timeouts
quiz:
  - question: |
      This test hangs until the test binary's timeout kills it. Why?

      ```go
      block := make(chan struct{})
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          <-block
      }))
      defer srv.Close()
      // ... client call that times out after 50ms ...
      ```
    options:
      - text: The client timeout doesn't work against httptest servers
      - text: '`srv.Close()` waits for running handlers, and this one waits forever on `block`, which nobody closes'
        correct: true
      - text: Channels can't be used in handlers
    explanation: |
      The client gives up, but the handler is still blocked, and `Close` waits for it.
      Wait on `r.Context().Done()` instead (it fires when the client disconnects), or
      `defer close(block)` *after* `defer srv.Close()` so it runs first.
  - question: Inside a `synctest.Test` bubble, a handler calls `time.Sleep(time.Minute)` and the client has a 30-second timeout. How long does the test take in real time?
    options:
      - text: About 30 seconds
      - text: About a minute
      - text: Almost no time, because the bubble's fake clock jumps ahead whenever every goroutine in it is blocked
        correct: true
    explanation: |
      In a bubble, time only advances when every goroutine is durably blocked, and then
      it jumps straight to the next timer. The test sees exactly 30 seconds pass, in a
      few microseconds of real time.
exercise:
  starter: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"iter"
    	"net/http"
    	"net/http/httptest"
    	"net/url"
    	"strconv"
    	"time"
    )

    type Issue struct {
    	ID    int    `json:"id"`
    	Title string `json:"title"`
    	State string `json:"state"`
    }

    type issuePage struct {
    	Issues     []Issue `json:"issues"`
    	NextCursor string  `json:"next_cursor"`
    }

    // APIError is returned for any response outside the 2xx range.
    type APIError struct {
    	StatusCode int
    	Message    string
    }

    func (e *APIError) Error() string {
    	return fmt.Sprintf("trackr API: %d %s", e.StatusCode, e.Message)
    }

    // checkResponse returns nil for a 2xx response, or an *APIError (chapter 5).
    func checkResponse(resp *http.Response) error {
    	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
    		return nil
    	}
    	var body struct {
    		Error string `json:"error"`
    	}
    	msg := http.StatusText(resp.StatusCode)
    	if err := json.UnmarshalRead(resp.Body, &body); err == nil && body.Error != "" {
    		msg = body.Error
    	}
    	return &APIError{StatusCode: resp.StatusCode, Message: msg}
    }

    // retryable reports whether a status is worth another attempt (chapter 6).
    func retryable(code int) bool {
    	switch code {
    	case http.StatusTooManyRequests, http.StatusBadGateway,
    		http.StatusServiceUnavailable, http.StatusGatewayTimeout:
    		return true
    	}
    	return false
    }

    // retryAfter reads a Retry-After header in seconds (the date form is left out here).
    func retryAfter(h http.Header) (time.Duration, bool) {
    	secs, err := strconv.Atoi(h.Get("Retry-After"))
    	if err != nil || secs < 0 {
    		return 0, false
    	}
    	return time.Duration(secs) * time.Second, true
    }

    // Client talks to the Trackr API.
    type Client struct {
    	BaseURL     string        // e.g. "https://api.trackr.dev"
    	Token       string        // sent as "Authorization: Bearer <Token>"
    	HTTPClient  *http.Client  // used for every request
    	MaxAttempts int           // total attempts per request, including the first
    	BaseDelay   time.Duration // wait before retry n: BaseDelay << (n-1)
    }

    // newRequest builds a GET request for c.BaseURL + path with the given query.
    // It sets "Authorization: Bearer <c.Token>" and "Accept: application/json".
    func (c *Client) newRequest(ctx context.Context, path string, query url.Values) (*http.Request, error) {
    	// ?
    	return nil, errors.New("newRequest: not implemented")
    }

    // getPage fetches one page of GET /issues. It sends state (if not "") and
    // cursor (if not "") as query parameters. It retries dropped connections and
    // retryable statuses, up to c.MaxAttempts attempts in total, waiting
    // Retry-After if the server sent it and BaseDelay << (n-1) otherwise.
    func (c *Client) getPage(ctx context.Context, state, cursor string) (issuePage, error) {
    	// ?
    	return issuePage{}, errors.New("getPage: not implemented")
    }

    // Issues yields every issue with the given state, one page at a time.
    // On an error it yields (Issue{}, err) once and stops.
    func (c *Client) Issues(ctx context.Context, state string) iter.Seq2[Issue, error] {
    	return func(yield func(Issue, error) bool) {
    		// ?
    	}
    }

    func main() {
    	calls := 0
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		calls++
    		fmt.Printf("  (server: %s, attempt %d)\n", r.URL.RequestURI(), calls)
    		w.Header().Set("Content-Type", "application/json")
    		switch {
    		case calls == 2:
    			w.WriteHeader(http.StatusServiceUnavailable)
    			fmt.Fprint(w, `{"error": "deploying, try again"}`)
    		case r.URL.Query().Get("cursor") == "":
    			fmt.Fprint(w, `{"issues": [{"id": 1, "title": "Login broken", "state": "open"}, {"id": 2, "title": "Add dark mode", "state": "open"}], "next_cursor": "c-2"}`)
    		default:
    			fmt.Fprint(w, `{"issues": [{"id": 3, "title": "Export to CSV", "state": "open"}], "next_cursor": ""}`)
    		}
    	}))
    	defer srv.Close()

    	c := &Client{BaseURL: srv.URL, Token: "tk_demo", HTTPClient: srv.Client(), MaxAttempts: 3, BaseDelay: 10 * time.Millisecond}
    	for iss, err := range c.Issues(context.Background(), "open") {
    		if apiErr, ok := errors.AsType[*APIError](err); ok {
    			fmt.Println("API error:", apiErr.StatusCode, apiErr.Message)
    			return
    		}
    		if err != nil {
    			fmt.Println("error:", err)
    			return
    		}
    		fmt.Printf("#%d %s\n", iss.ID, iss.Title)
    	}
    	fmt.Println("done")
    }
  solution: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"iter"
    	"net/http"
    	"net/http/httptest"
    	"net/url"
    	"strconv"
    	"time"
    )

    type Issue struct {
    	ID    int    `json:"id"`
    	Title string `json:"title"`
    	State string `json:"state"`
    }

    type issuePage struct {
    	Issues     []Issue `json:"issues"`
    	NextCursor string  `json:"next_cursor"`
    }

    // APIError is returned for any response outside the 2xx range.
    type APIError struct {
    	StatusCode int
    	Message    string
    }

    func (e *APIError) Error() string {
    	return fmt.Sprintf("trackr API: %d %s", e.StatusCode, e.Message)
    }

    // checkResponse returns nil for a 2xx response, or an *APIError (chapter 5).
    func checkResponse(resp *http.Response) error {
    	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
    		return nil
    	}
    	var body struct {
    		Error string `json:"error"`
    	}
    	msg := http.StatusText(resp.StatusCode)
    	if err := json.UnmarshalRead(resp.Body, &body); err == nil && body.Error != "" {
    		msg = body.Error
    	}
    	return &APIError{StatusCode: resp.StatusCode, Message: msg}
    }

    // retryable reports whether a status is worth another attempt (chapter 6).
    func retryable(code int) bool {
    	switch code {
    	case http.StatusTooManyRequests, http.StatusBadGateway,
    		http.StatusServiceUnavailable, http.StatusGatewayTimeout:
    		return true
    	}
    	return false
    }

    // retryAfter reads a Retry-After header in seconds (the date form is left out here).
    func retryAfter(h http.Header) (time.Duration, bool) {
    	secs, err := strconv.Atoi(h.Get("Retry-After"))
    	if err != nil || secs < 0 {
    		return 0, false
    	}
    	return time.Duration(secs) * time.Second, true
    }

    // Client talks to the Trackr API.
    type Client struct {
    	BaseURL     string        // e.g. "https://api.trackr.dev"
    	Token       string        // sent as "Authorization: Bearer <Token>"
    	HTTPClient  *http.Client  // used for every request
    	MaxAttempts int           // total attempts per request, including the first
    	BaseDelay   time.Duration // wait before retry n: BaseDelay << (n-1)
    }

    func (c *Client) newRequest(ctx context.Context, path string, query url.Values) (*http.Request, error) {
    	u, err := url.Parse(c.BaseURL)
    	if err != nil {
    		return nil, err
    	}
    	u = u.JoinPath(path)
    	u.RawQuery = query.Encode()
    	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
    	if err != nil {
    		return nil, err
    	}
    	req.Header.Set("Authorization", "Bearer "+c.Token)
    	req.Header.Set("Accept", "application/json")
    	return req, nil
    }

    // try makes one attempt. retry reports whether another attempt might help,
    // and wait is the server's Retry-After hint (0 if none).
    func (c *Client) try(ctx context.Context, query url.Values) (page issuePage, retry bool, wait time.Duration, err error) {
    	req, err := c.newRequest(ctx, "/issues", query)
    	if err != nil {
    		return issuePage{}, false, 0, err
    	}
    	resp, err := c.HTTPClient.Do(req)
    	if err != nil {
    		return issuePage{}, ctx.Err() == nil, 0, err
    	}
    	defer resp.Body.Close()
    	if err := checkResponse(resp); err != nil {
    		wait, _ := retryAfter(resp.Header)
    		return issuePage{}, retryable(resp.StatusCode), wait, err
    	}
    	if err := json.UnmarshalRead(resp.Body, &page); err != nil {
    		return issuePage{}, false, 0, fmt.Errorf("decoding page: %w", err)
    	}
    	return page, false, 0, nil
    }

    func (c *Client) getPage(ctx context.Context, state, cursor string) (issuePage, error) {
    	query := url.Values{}
    	if state != "" {
    		query.Set("state", state)
    	}
    	if cursor != "" {
    		query.Set("cursor", cursor)
    	}
    	for attempt := 1; ; attempt++ {
    		page, retry, wait, err := c.try(ctx, query)
    		if err == nil {
    			return page, nil
    		}
    		if !retry || attempt >= c.MaxAttempts {
    			return issuePage{}, err
    		}
    		if wait == 0 {
    			wait = c.BaseDelay << (attempt - 1)
    		}
    		timer := time.NewTimer(wait)
    		select {
    		case <-ctx.Done():
    			timer.Stop()
    			return issuePage{}, ctx.Err()
    		case <-timer.C:
    		}
    	}
    }

    func (c *Client) Issues(ctx context.Context, state string) iter.Seq2[Issue, error] {
    	return func(yield func(Issue, error) bool) {
    		cursor := ""
    		for {
    			page, err := c.getPage(ctx, state, cursor)
    			if err != nil {
    				yield(Issue{}, err)
    				return
    			}
    			for _, iss := range page.Issues {
    				if !yield(iss, nil) {
    					return
    				}
    			}
    			if page.NextCursor == "" {
    				return
    			}
    			cursor = page.NextCursor
    		}
    	}
    }

    func main() {
    	calls := 0
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		calls++
    		fmt.Printf("  (server: %s, attempt %d)\n", r.URL.RequestURI(), calls)
    		w.Header().Set("Content-Type", "application/json")
    		switch {
    		case calls == 2:
    			w.WriteHeader(http.StatusServiceUnavailable)
    			fmt.Fprint(w, `{"error": "deploying, try again"}`)
    		case r.URL.Query().Get("cursor") == "":
    			fmt.Fprint(w, `{"issues": [{"id": 1, "title": "Login broken", "state": "open"}, {"id": 2, "title": "Add dark mode", "state": "open"}], "next_cursor": "c-2"}`)
    		default:
    			fmt.Fprint(w, `{"issues": [{"id": 3, "title": "Export to CSV", "state": "open"}], "next_cursor": ""}`)
    		}
    	}))
    	defer srv.Close()

    	c := &Client{BaseURL: srv.URL, Token: "tk_demo", HTTPClient: srv.Client(), MaxAttempts: 3, BaseDelay: 10 * time.Millisecond}
    	for iss, err := range c.Issues(context.Background(), "open") {
    		if apiErr, ok := errors.AsType[*APIError](err); ok {
    			fmt.Println("API error:", apiErr.StatusCode, apiErr.Message)
    			return
    		}
    		if err != nil {
    			fmt.Println("error:", err)
    			return
    		}
    		fmt.Printf("#%d %s\n", iss.ID, iss.Title)
    	}
    	fmt.Println("done")
    }
  tests: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"slices"
    	"strings"
    	"sync"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // fakeAPI serves /issues from pages: the first page for an empty cursor,
    // then "p2", "p3", ... Each entry in fail is used up by one request before
    // any page is served: 0 means drop the connection, anything else is a status.
    type fakeAPI struct {
    	pages      [][]int
    	fail       []int
    	retryAfter string

    	mu   sync.Mutex
    	reqs []*http.Request
    }

    func (f *fakeAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    	f.mu.Lock()
    	f.reqs = append(f.reqs, r)
    	var failWith int
    	failing := len(f.fail) > 0
    	if failing {
    		failWith, f.fail = f.fail[0], f.fail[1:]
    	}
    	f.mu.Unlock()

    	w.Header().Set("Content-Type", "application/json")
    	if failing {
    		if failWith == 0 {
    			panic(http.ErrAbortHandler)
    		}
    		if f.retryAfter != "" {
    			w.Header().Set("Retry-After", f.retryAfter)
    		}
    		w.WriteHeader(failWith)
    		fmt.Fprintf(w, `{"error": "fake failure %d"}`, failWith)
    		return
    	}
    	n := 0
    	if c := r.URL.Query().Get("cursor"); c != "" {
    		fmt.Sscanf(c, "p%d", &n)
    		n--
    	}
    	if n < 0 || n >= len(f.pages) {
    		w.WriteHeader(http.StatusBadRequest)
    		fmt.Fprintf(w, `{"error": "unknown cursor %q"}`, r.URL.Query().Get("cursor"))
    		return
    	}
    	var items []string
    	for _, id := range f.pages[n] {
    		items = append(items, fmt.Sprintf(`{"id": %d, "title": "Issue %d", "state": "open"}`, id, id))
    	}
    	next := ""
    	if n+1 < len(f.pages) {
    		next = fmt.Sprintf("p%d", n+2)
    	}
    	fmt.Fprintf(w, `{"issues": [%s], "next_cursor": %q}`, strings.Join(items, ", "), next)
    }

    func (f *fakeAPI) requests() []*http.Request {
    	f.mu.Lock()
    	defer f.mu.Unlock()
    	return slices.Clone(f.reqs)
    }

    // newClient returns a Client wired to f. Tests outside a synctest bubble use
    // real time, so they get a tiny BaseDelay.
    func newClient(t *testing.T, f *fakeAPI, baseDelay time.Duration) *Client {
    	srv := httptest.NewTestServer(t, f)
    	return &Client{
    		// srv.Client() sends every request to the fake, whatever the host.
    		// Nothing listens on this port, so a request that escapes it fails fast.
    		BaseURL:     "https://127.0.0.1:1",
    		Token:       "tk_test",
    		HTTPClient:  srv.Client(),
    		MaxAttempts: 3,
    		BaseDelay:   baseDelay,
    	}
    }

    func collect(c *Client, ctx context.Context, state string) ([]int, []error) {
    	var ids []int
    	var errs []error
    	for iss, err := range c.Issues(ctx, state) {
    		if err != nil {
    			errs = append(errs, err)
    			continue
    		}
    		ids = append(ids, iss.ID)
    	}
    	return ids, errs
    }

    func TestRequestShape(t *testing.T) {
    	f := &fakeAPI{pages: [][]int{{1, 2}}}
    	c := newClient(t, f, time.Millisecond)
    	ids, errs := collect(c, t.Context(), "open")
    	if len(errs) > 0 {
    		t.Fatalf("Issues yielded error %v (are you sending every request with c.HTTPClient to c.BaseURL?)", errs[0])
    	}
    	if !slices.Equal(ids, []int{1, 2}) {
    		t.Errorf("Issues yielded IDs %v, want [1 2]", ids)
    	}
    	reqs := f.requests()
    	if len(reqs) != 1 {
    		t.Fatalf("server saw %d requests, want 1", len(reqs))
    	}
    	r := reqs[0]
    	if r.Method != http.MethodGet || r.URL.Path != "/issues" {
    		t.Errorf("server saw %s %s, want GET /issues", r.Method, r.URL.Path)
    	}
    	if r.Host != "127.0.0.1:1" {
    		t.Errorf("request went to host %q, want 127.0.0.1:1 (build URLs from c.BaseURL)", r.Host)
    	}
    	q := r.URL.Query()
    	if q.Get("state") != "open" {
    		t.Errorf("query %q: want state=open", r.URL.RawQuery)
    	}
    	if q.Has("cursor") {
    		t.Errorf("query %q: the first request shouldn't send a cursor", r.URL.RawQuery)
    	}
    	if got := r.Header.Get("Authorization"); got != "Bearer tk_test" {
    		t.Errorf("Authorization = %q, want %q", got, "Bearer tk_test")
    	}
    	if got := r.Header.Get("Accept"); got != "application/json" {
    		t.Errorf("Accept = %q, want application/json", got)
    	}
    }

    func TestPaging(t *testing.T) {
    	f := &fakeAPI{pages: [][]int{{1, 2, 3}, {4, 5, 6}, {7}}}
    	c := newClient(t, f, time.Millisecond)
    	ids, errs := collect(c, t.Context(), "")
    	if len(errs) > 0 {
    		t.Fatalf("Issues yielded error %v", errs[0])
    	}
    	if !slices.Equal(ids, []int{1, 2, 3, 4, 5, 6, 7}) {
    		t.Errorf("Issues yielded IDs %v, want [1 2 3 4 5 6 7]", ids)
    	}
    	var cursors []string
    	for _, r := range f.requests() {
    		cursors = append(cursors, r.URL.Query().Get("cursor"))
    		if r.URL.Query().Has("state") {
    			t.Errorf("query %q: with state \"\", leave the state parameter out", r.URL.RawQuery)
    		}
    	}
    	if !slices.Equal(cursors, []string{"", "p2", "p3"}) {
    		t.Errorf("server saw cursors %q, want [\"\" \"p2\" \"p3\"]", cursors)
    	}
    }

    func TestStopsEarly(t *testing.T) {
    	f := &fakeAPI{pages: [][]int{{1, 2, 3}, {4, 5, 6}, {7}}}
    	c := newClient(t, f, time.Millisecond)
    	var ids []int
    	for iss, err := range c.Issues(t.Context(), "open") {
    		if err != nil {
    			t.Fatalf("Issues yielded error %v", err)
    		}
    		ids = append(ids, iss.ID)
    		if len(ids) == 4 {
    			break
    		}
    	}
    	if n := len(f.requests()); n != 2 {
    		t.Errorf("caller stopped after 4 issues, but the server saw %d requests, want 2 (fetch pages lazily)", n)
    	}
    }

    func TestTypedErrorNotRetried(t *testing.T) {
    	f := &fakeAPI{pages: [][]int{{1}}, fail: []int{http.StatusNotFound}}
    	c := newClient(t, f, time.Millisecond)
    	ids, errs := collect(c, t.Context(), "open")
    	if len(ids) != 0 || len(errs) != 1 {
    		t.Fatalf("got IDs %v and %d errors, want no IDs and exactly one error", ids, len(errs))
    	}
    	apiErr, ok := errors.AsType[*APIError](errs[0])
    	if !ok {
    		t.Fatalf("error %v (%T) isn't an *APIError; return what checkResponse gives you", errs[0], errs[0])
    	}
    	if apiErr.StatusCode != http.StatusNotFound || apiErr.Message != "fake failure 404" {
    		t.Errorf("APIError = %+v, want StatusCode 404 and Message \"fake failure 404\"", *apiErr)
    	}
    	if n := len(f.requests()); n != 1 {
    		t.Errorf("server saw %d requests for a 404, want 1 (don't retry non-retryable statuses)", n)
    	}
    }

    func TestRetriesWithBackoff(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		f := &fakeAPI{pages: [][]int{{1, 2}}, fail: []int{http.StatusServiceUnavailable, http.StatusBadGateway}}
    		c := newClient(t, f, time.Second)
    		start := time.Now()
    		ids, errs := collect(c, t.Context(), "open")
    		elapsed := time.Since(start)
    		if len(errs) > 0 {
    			t.Fatalf("after 503, 502, then 200: Issues yielded error %v, want success", errs[0])
    		}
    		if !slices.Equal(ids, []int{1, 2}) {
    			t.Errorf("Issues yielded IDs %v, want [1 2]", ids)
    		}
    		if n := len(f.requests()); n != 3 {
    			t.Errorf("server saw %d requests, want 3", n)
    		}
    		if elapsed != 3*time.Second {
    			t.Errorf("retries took %v of (fake) time, want 3s: wait BaseDelay (1s), then 2s", elapsed)
    		}
    	})
    }

    func TestRetryAfter(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		f := &fakeAPI{pages: [][]int{{1}}, fail: []int{http.StatusTooManyRequests}, retryAfter: "7"}
    		c := newClient(t, f, time.Second)
    		start := time.Now()
    		_, errs := collect(c, t.Context(), "open")
    		elapsed := time.Since(start)
    		if len(errs) > 0 {
    			t.Fatalf("after 429 then 200: Issues yielded error %v, want success", errs[0])
    		}
    		if elapsed != 7*time.Second {
    			t.Errorf("waited %v after a 429 with Retry-After: 7, want exactly 7s", elapsed)
    		}
    	})
    }

    func TestDroppedConnectionRetried(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		f := &fakeAPI{pages: [][]int{{1}}, fail: []int{0}}
    		c := newClient(t, f, time.Second)
    		ids, errs := collect(c, t.Context(), "open")
    		if len(errs) > 0 {
    			t.Fatalf("after one dropped connection: Issues yielded error %v, want a retry to succeed", errs[0])
    		}
    		if !slices.Equal(ids, []int{1}) || len(f.requests()) != 2 {
    			t.Errorf("got IDs %v after %d requests, want [1] after 2", ids, len(f.requests()))
    		}
    	})
    }

    func TestGivesUp(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		f := &fakeAPI{pages: [][]int{{1}}, fail: []int{503, 503, 503, 503, 503}}
    		c := newClient(t, f, time.Second)
    		start := time.Now()
    		_, errs := collect(c, t.Context(), "open")
    		elapsed := time.Since(start)
    		if len(errs) != 1 {
    			t.Fatalf("server always says 503: got %d errors, want exactly 1", len(errs))
    		}
    		if apiErr, ok := errors.AsType[*APIError](errs[0]); !ok || apiErr.StatusCode != 503 {
    			t.Errorf("error = %v, want the last attempt's *APIError (503)", errs[0])
    		}
    		if n := len(f.requests()); n != 3 {
    			t.Errorf("server saw %d requests, want MaxAttempts = 3", n)
    		}
    		if elapsed != 3*time.Second {
    			t.Errorf("giving up took %v of (fake) time, want 3s (1s + 2s, and no wait after the last attempt)", elapsed)
    		}
    	})
    }

    func TestContextStopsBackoff(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		f := &fakeAPI{pages: [][]int{{1}}, fail: []int{503, 503, 503}}
    		c := newClient(t, f, time.Second)
    		c.BaseDelay = time.Hour
    		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
    		defer cancel()
    		start := time.Now()
    		_, errs := collect(c, ctx, "open")
    		elapsed := time.Since(start)
    		if len(errs) != 1 || !errors.Is(errs[0], context.DeadlineExceeded) {
    			t.Fatalf("got errors %v, want one error matching context.DeadlineExceeded", errs)
    		}
    		if elapsed != 10*time.Second {
    			t.Errorf("returned after %v, want 10s: wait in a select on ctx.Done(), not time.Sleep", elapsed)
    		}
    	})
    }
---

`trackr`'s timeouts are some of its most important code, and the easiest to leave
untested, because a test that waits 30 seconds for a timeout is a test nobody runs.
There are two good ways to make these tests fast.

## Option 1: tiny real timeouts

Make the fake server hang, and give the client a very short deadline:

```go
func TestGetIssueRespectsDeadline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // hang until the client gives up
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, HTTPClient: srv.Client()}
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := c.GetIssue(ctx, 1)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("GetIssue error = %v, want context.DeadlineExceeded", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("GetIssue took %v; it should give up after about 50ms", elapsed)
	}
}
```

Two details make it robust:

- **The handler waits on `r.Context().Done()`**, which fires when the client
  disconnects. Don't use `time.Sleep(10 * time.Second)` in a handler. `srv.Close()`
  waits for handlers to finish, so the test would take 10 seconds anyway.
- **The upper bound is generous.** Checking `elapsed < 60ms` would fail on a busy CI
  machine. Check that it gave up "reasonably soon" (well under the server's hang), not
  that it hit an exact time.

This works, but every such test still costs real milliseconds, and the upper bound is
fuzzy by necessity.

## Option 2: fake time with synctest

You met `testing/synctest` (Go 1.25) in
[Learn Concurrency](/courses/learn-concurrency/testing-concurrent-code/synctest). A
quick recap: `synctest.Test` runs a function in a **bubble** with its own fake clock.
Inside the bubble, time stands still while any goroutine can make progress. When
**every** goroutine in the bubble is blocked (on a sleep, a timer, a channel), the
clock jumps straight to the next timer that would fire.

Go 1.27's `httptest.NewTestServer` runs its in-memory network inside the bubble too,
so a whole HTTP exchange can run on fake time:

```go
func TestClientTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-time.After(time.Minute): // a very slow server, in fake time
			case <-r.Context().Done(): // the client gave up
			}
		}))
		client := srv.Client()
		client.Timeout = 30 * time.Second
		c := &Client{BaseURL: "https://api.trackr.dev", HTTPClient: client}

		start := time.Now()
		_, err := c.GetIssue(t.Context(), 1)
		if err == nil {
			t.Fatal("GetIssue succeeded against a server that never answers")
		}
		if got := time.Since(start); got != 30*time.Second {
			t.Errorf("gave up after %v, want exactly 30s", got)
		}
	})
}
```

This test checks for **exactly** 30 seconds, and it finishes in a few milliseconds of
real time. A minute-long wait in the handler costs nothing here, because the fake clock
skips through it once everything else is blocked. It still watches `r.Context()`, for
the same reason as in option 1: the server's cleanup waits for running handlers.

A few rules of the bubble:

- Use `synctest.Test(t, func(t *testing.T) { ... })` and the inner `t`.
- Create the server, the client and any contexts **inside** the bubble. A
  `NewServer` listening on a real port involves real network I/O, which the bubble
  can't wait on, so use `NewTestServer`.
- `synctest.Wait()` blocks until every other goroutine in the bubble is blocked.
  That's handy for asserting "nothing happened yet" before advancing time with
  `time.Sleep`, or with `synctest.Sleep` (Go 1.27), which does both in one call.

## Testing retries on fake time

Retries are where synctest really pays off. Backoff delays of 1s, 2s and 4s would make
a real-time test crawl, and jitter would make exact assertions impossible. In a bubble
you can use production-sized delays and check the total exactly:

```go
synctest.Test(t, func(t *testing.T) {
	flaky := &Flaky{Failures: 2, Body: `{"issues": [], "next_cursor": ""}`} // lesson 3
	srv := httptest.NewTestServer(t, flaky)
	c := &Client{BaseURL: "https://api.trackr.dev", HTTPClient: srv.Client(),
		MaxAttempts: 3, BaseDelay: time.Second}

	start := time.Now()
	for _, err := range c.Issues(t.Context(), "open") {
		if err != nil {
			t.Fatal(err)
		}
	}
	if got := time.Since(start); got != 3*time.Second { // 1s, then 2s
		t.Errorf("retries took %v, want 3s", got)
	}
	if n := flaky.Requests(); n != 3 {
		t.Errorf("server saw %d requests, want 3", n)
	}
})
```

If your backoff uses jitter, either assert a range or make the random source a field
that tests can replace, the same "inject what varies" idea as lesson 2.

## Which to use?

Use real short timeouts when the code under test uses a real network or other things
outside Go's control. Use synctest whenever you can. It's exact, deterministic and
instant, and it makes testing retries with backoff almost boring.

## Your turn: assemble Trackr

Time to put the course together. The starter has the pieces you've already built:
`APIError` and `checkResponse` (chapter 5), `retryable` and a seconds-only `retryAfter`
(chapter 6). Complete the three methods that tie them together:

1. **`newRequest(ctx, path, query)`** builds a `GET` for `c.BaseURL` + `path` with the
   encoded `query` (`url.URL.JoinPath` and `RawQuery` from chapter 2 work well), and
   sets `Authorization: Bearer <c.Token>` and `Accept: application/json`.
2. **`getPage(ctx, state, cursor)`** fetches one page of `/issues`. Send `state` and
   `cursor` as query parameters, each only when it isn't empty. Build a fresh request
   for every attempt and send it with `c.HTTPClient`. Retry a dropped connection or a
   `retryable` status, making at most `c.MaxAttempts` attempts in total. Before retry
   *n*, wait `Retry-After` if the server sent one, otherwise `c.BaseDelay << (n-1)`,
   in a `select` that also watches `ctx.Done()`. Return other `*APIError`s (like a
   404) straight away, and the last error once you're out of attempts.
3. **`Issues(ctx, state)`** returns an `iter.Seq2[Issue, error]` that pages through
   every issue lazily, following `next_cursor` until it's empty, and yields an error
   once and stops (chapter 7).

(Real clients add jitter to the backoff. It's left out here so the tests can check
exact timings.) The hidden tests use `NewTestServer` with synctest, so they check
your delays to the second without waiting for them. If a helper like
`try(ctx, query) (page, retry, wait, err)` for a single attempt keeps `getPage` tidy,
write one.

With that, the core of `trackr list` is a few lines:

```go
state := flag.String("state", "open", "only list issues in this state")
flag.Parse()

c := &Client{
	BaseURL:     cmp.Or(os.Getenv("TRACKR_API_URL"), "https://api.trackr.dev"),
	Token:       os.Getenv("TRACKR_TOKEN"),
	HTTPClient:  &http.Client{Timeout: 30 * time.Second},
	MaxAttempts: 4,
	BaseDelay:   500 * time.Millisecond,
}
for iss, err := range c.Issues(ctx, *state) {
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("#%d %s\n", iss.ID, iss.Title)
}
```

## Wrapping up

That's the course. You've taken `trackr` from a single `http.Get` to a client that
builds URLs and requests carefully, speaks JSON with `encoding/json/v2`, turns bad
statuses into typed errors, times out, retries politely, respects rate limits, streams
pages through an iterator, keeps TLS verification on and its secrets out of logs, and
is tested end to end without touching the network.

You've spent nine chapters on the client side of the conversation, faking servers with
`httptest`. Next, build the real thing:
[Learn HTTP Servers in Go](/courses/learn-http-servers) has you build Squeak, a
small JSON API, with routing, middleware, JSON handlers, storage, authentication and
graceful shutdown.

## Further reading

- The Go blog, "Testing concurrent code with testing/synctest": https://go.dev/blog/synctest
